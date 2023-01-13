package frp

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	client := NewFakeClient()
	cfg, err := client.GetConfigs()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(cfg)
	if cfg.Proxy == nil {
		cfg.Proxy = make(map[string]Config)
	}
	cfg.Proxy["ingress1"] = MapConfig{
		"local_ip":    "0.0.0.0",
		"local_port":  "8080",
		"remote_port": "80",
	}
	fmt.Println(string(Marshal(cfg)))
}
