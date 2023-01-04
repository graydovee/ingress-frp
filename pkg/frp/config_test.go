package frp

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	client := NewFakeClient()
	cfgStr, err := client.GetConfig()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(cfgStr))
	cfg, err := Unmarshal(cfgStr)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(cfg)
	if cfg.Proxy == nil {
		cfg.Proxy = make(map[string]Proxy)
	}
	cfg.Proxy["ingress1"] = Proxy{
		"local_ip":    "0.0.0.0",
		"local_port":  "8080",
		"remote_port": "80",
	}
	fmt.Println(string(Marshal(cfg)))
}
