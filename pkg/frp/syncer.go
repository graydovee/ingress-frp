package frp

import (
	"context"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sync"
	"time"
)

const syncInterval = time.Minute

type Syncer struct {
	ctx context.Context

	clients []Client

	configs *Configs
	ch      chan struct{}
	mu      sync.Mutex
}

func NewSyncer(clients ...Client) *Syncer {
	return &Syncer{
		ch:      make(chan struct{}),
		clients: clients,
	}
}

func (s *Syncer) Start(ctx context.Context) error {
	s.ctx = ctx
	ticker := time.NewTicker(syncInterval)
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

func (s *Syncer) SetConfig(configs *Configs) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs = configs

	s.Sync()
}

func (s *Syncer) SetProxies(configs map[string]Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs.Proxy = configs

	s.Sync()
}

func (s *Syncer) Sync() {
	select {
	case <-s.ctx.Done():
		return
	case <-s.ch:
		s.ch <- struct{}{}
	default:
	}
}

func (s *Syncer) sync(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	l := log.FromContext(ctx)
	for _, cli := range s.clients {
		configs, err := cli.GetConfigs(ctx)
		if err != nil {
			l.Error(err, "get config error", "client", cli.Info())
			continue
		}
		if reflect.DeepEqual(configs, s.configs) {
			continue
		}
		l.Info("sync config", "client", cli.Info())

		configs.Proxy = s.configs.Proxy
		if err = cli.SetConfig(ctx, configs); err != nil {
			l.Error(err, "set config error", "client", cli.Info())
			continue
		}

		if err := cli.Reload(ctx); err != nil {
			l.Error(err, "reload config error", "client", cli.Info())
			continue
		}
	}
}
