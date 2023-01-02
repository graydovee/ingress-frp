package frp

import (
	"fmt"
	"testing"
)

func TestClient_GetConfig(t *testing.T) {
	client := NewClient()
	cfg, err := client.GetConfig()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(cfg))
	parse, err := Unmarshal(cfg)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(parse)
}
