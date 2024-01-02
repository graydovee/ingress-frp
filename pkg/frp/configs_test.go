package frp

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	l := log.FromContext(context.Background())
	client := NewFakeClient()
	cfg, err := client.GetConfigs(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	l.Info("print cfg", "cfg", cfg)
	if cfg.Proxy == nil {
		cfg.Proxy = make(map[string]Config)
	}
	cfg.Proxy["ingress1"] = MapConfig{
		"local_ip":    "0.0.0.0",
		"local_port":  "8080",
		"remote_port": "80",
	}
	cfg.Proxy["ingress2"] = &Https2HttpConfig{
		HttpConfig: HttpConfig{
			Host:      "example.com",
			Locations: "/",
			LocalIp:   "127.0.0.1",
			LocalPort: "3000",
			HttpUser:  "user",
			HttpPwd:   "pwd",
		},
		CrtBase64: "123321123",
		KeyBase64: "321123321",
	}

	cfg.Proxy["ingress3"] = &ServerHttpsConfig{
		HttpConfig: HttpConfig{
			Host:      "example.com",
			Locations: "/",
			LocalIp:   "127.0.0.1",
			LocalPort: "3000",
		},
		TlsCrt: "123321123",
		TlsKey: "321123321",
	}
	cfg.Proxy["ingress4"] = &HttpConfig{
		Redirect: "https://baidu.com",
	}

	l.Info(string(Marshal(cfg)))

	l.Info("print cfg", "cfg", cfg)

	s := NewFakeSyncer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.Start(ctx)
	time.Sleep(time.Second)

	s.SetProxies("", cfg.Proxy)
	s.Sync()

	time.Sleep(time.Second)
	s.Sync()

	time.Sleep(time.Second)
	s.Sync()
}
