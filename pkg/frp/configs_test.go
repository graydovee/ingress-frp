package frp

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
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
		},
		CrtBase64: "123321123",
		KeyBase64: "321123321",
	}

	l.Info(string(Marshal(cfg)))

	l.Info("print cfg", "cfg", cfg)
}
