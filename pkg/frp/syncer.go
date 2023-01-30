package frp

import (
	"context"
	"fmt"
	"github.com/grydovee/ingress-frp/pkg/constants"
	"github.com/grydovee/ingress-frp/pkg/utils"
	"net"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sync"
	"time"
)

type Syncer interface {
	Start(ctx context.Context) error
	SetConfig(configs *Configs)
	SetProxies(configs map[string]Config)
	Sync()
}

type syncer struct {
	ctx context.Context

	domainWatcher *utils.DomainWatcher
	clients       map[string]Client

	configs *Configs
	ch      chan struct{}
	mu      sync.Mutex
}

func NewSyncer(addr string, port uint16, uname string, passwd string) Syncer {
	s := &syncer{
		domainWatcher: utils.NewDomainWatcher(addr),
		ch:            make(chan struct{}),
	}
	s.domainWatcher.OnClientChange = func(ips []net.IP) {
		s.mu.Lock()
		defer s.mu.Unlock()

		newClients := make(map[string]Client)
		for _, ip := range ips {
			if _, ok := s.clients[ip.String()]; ok {
				newClients[ip.String()] = s.clients[ip.String()]
			} else {
				newClients[ip.String()] = NewClient(ip, port, uname, passwd)
			}
		}
		s.clients = newClients
		s.Sync()
	}
	return s
}

func (s *syncer) Start(ctx context.Context) error {
	if s.domainWatcher != nil {
		go s.domainWatcher.Start(ctx)
	}
	s.ctx = ctx
	ticker := time.NewTicker(constants.FrpClientSyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			close(s.ch)
			s.mu.Unlock()
			return nil
		case <-s.ch:
			s.sync(ctx)
		case <-ticker.C:
			s.sync(ctx)
		}
	}
}

func (s *syncer) SetConfig(configs *Configs) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs = configs

	s.Sync()
}

func (s *syncer) SetProxies(configs map[string]Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.configs == nil {
		s.configs = &Configs{}
	}
	s.configs.Proxy = configs

	s.Sync()
}

func (s *syncer) Sync() {
	select {
	case <-s.ctx.Done():
		return
	case <-s.ch:
		s.ch <- struct{}{}
	default:
	}
}

func (s *syncer) sync(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l := log.FromContext(ctx)
	for _, cli := range s.clients {
		configs, err := cli.GetConfigs(ctx)
		if err != nil {
			l.Error(err, "get config error", "client", cli.Addr())
			continue
		}
		if reflect.DeepEqual(configs, s.configs) {
			continue
		}
		l.Info("sync config", "client", cli.Addr())

		newProxy := make(map[string]Config)
		for name, config := range s.configs.Proxy {
			newProxy[fmt.Sprintf("%s/%s", cli.Addr(), name)] = config
		}
		configs.Proxy = newProxy
		if err = cli.SetConfig(ctx, configs); err != nil {
			l.Error(err, "set config error", "client", cli.Addr())
			continue
		}

		if err := cli.Reload(ctx); err != nil {
			l.Error(err, "reload config error", "client", cli.Addr())
			continue
		}
	}
}
