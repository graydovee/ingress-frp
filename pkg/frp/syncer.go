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
	SetProxies(key string, configs map[string]Config)
	Sync()
}

type syncer struct {
	ctx context.Context

	domainWatcher *utils.DomainWatcher
	clients       []Client

	configsMap map[string]map[string]Config
	ch         chan struct{}
	mu         sync.Mutex
}

func NewSyncer(addr string, port uint16, uname string, passwd string) Syncer {
	s := &syncer{
		domainWatcher: utils.NewDomainWatcher(addr),
		ch:            make(chan struct{}),
		configsMap:    make(map[string]map[string]Config),
	}
	s.domainWatcher.OnClientChange = func(ips []net.IP) {
		s.mu.Lock()
		defer s.mu.Unlock()

		newClients := make([]Client, 0)
		for _, ip := range ips {
			var foundCli Client
			for i := range s.clients {
				if s.clients[i].Addr().IP.Equal(ip) {
					foundCli = s.clients[i]
					break
				}
			}
			if foundCli != nil {
				newClients = append(newClients, foundCli)
			} else {
				newClients = append(newClients, NewClient(ip, port, uname, passwd))
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

func (s *syncer) SetProxies(key string, configs map[string]Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configsMap[key] = configs

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
	if s.configsMap == nil {
		return
	}

	singletonProxies := make(map[string]Config)
	groupProxies := make(map[string]Config)
	for _, configs := range s.configsMap {
		for key, config := range configs {
			if config.EnableGroup() {
				groupProxies[key] = config
			} else {
				singletonProxies[key] = config
			}
		}
	}
	l := log.FromContext(ctx)

	for i, cli := range s.clients {
		configs, err := cli.GetConfigs(ctx)
		if err != nil {
			l.Error(err, "get config error", "client", cli.Addr())
			continue
		}
		newProxy := make(map[string]Config)
		for name, config := range groupProxies {
			newProxy[fmt.Sprintf("%s/%s", cli.Addr(), name)] = config
		}

		for name, config := range singletonProxies {
			if i == hashStr(name)%len(s.clients) {
				newProxy[fmt.Sprintf("%s/%s", cli.Addr(), name)] = config
			}
		}

		if reflect.DeepEqual(newProxy, configs.Proxy) {
			continue
		}
		l.Info("sync config", "client", cli.Addr())

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

func hashStr(str string) int {
	var hash int
	for i := 0; i < len(str); i++ {
		hash = hash*31 + int(str[i])
	}
	return hash
}
