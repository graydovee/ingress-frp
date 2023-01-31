package frp

import (
	"fmt"
	"strings"
)

type Config interface {
	fmt.Stringer
	ToMap() map[string]string
	EnableGroup() bool
}

func ConfigEquals(cfg1, cfg2 Config) bool {
	m1 := cfg1.ToMap()
	m2 := cfg2.ToMap()
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		v2, ok := m2[k]
		if !ok {
			return false
		}
		if v1 != v2 {
			return false
		}
	}
	return true
}

type MapConfig map[string]string

func (p MapConfig) String() string {
	var pair []string
	for k, v := range p {
		if strings.HasPrefix(k, "plugin_crt") || strings.HasPrefix(k, "plugin_key") {
			pair = append(pair, fmt.Sprintf("%s:%s", k, "******"))
		} else {
			pair = append(pair, fmt.Sprintf("%s:%s", k, v))
		}
	}

	return "{" + strings.Join(pair, ", ") + "}"
}

func (p MapConfig) ToMap() map[string]string {
	m := make(map[string]string, len(p))
	for k, v := range p {
		m[k] = v
	}
	return m
}

func (p MapConfig) EnableGroup() bool {
	if p == nil {
		return false
	}
	switch p["type"] {
	case "tcp", "http", "tcpmux":
		return p["group"] != ""
	}
	return false
}

// HttpConfig
// [web01]
// type = http
// local_port = 80
// local_ip = 80
// custom_domains = web.yourdomain.com
// locations = /
// group = web
// group_key = 123
type HttpConfig struct {
	LocalIp   string
	LocalPort string
	Host      string
	Locations string
	Group     string
	GroupKey  string
}

func (h *HttpConfig) EnableGroup() bool {
	return true
}

func (h *HttpConfig) ToMap() map[string]string {
	m := make(map[string]string)
	m["type"] = "http"
	if len(h.LocalIp) > 0 {
		m["local_ip"] = h.LocalIp
	}
	if len(h.LocalPort) > 0 {
		m["local_port"] = h.LocalPort
	}
	if len(h.Host) > 0 {
		m["custom_domains"] = h.Host
	}
	if len(h.Locations) > 0 {
		m["locations"] = h.Locations
	}
	if len(h.Group) > 0 {
		m["group"] = h.Group
	}
	if len(h.GroupKey) > 0 {
		m["group_key"] = h.GroupKey
	}
	return m
}

func (h *HttpConfig) String() string {
	return MapConfig(h.ToMap()).String()
}

func NewHttpConfig(m map[string]string) *HttpConfig {
	return &HttpConfig{
		LocalIp:   m["local_ip"],
		LocalPort: m["local_port"],
		Host:      m["custom_domains"],
		Locations: m["locations"],
		Group:     m["group"],
		GroupKey:  m["group_key"],
	}
}

// Https2HttpConfig
// [test_htts2http]
// type = https
// custom_domains = git.graydove.cn
//
// plugin = https2http
// plugin_local_addr = 127.0.0.1:3000
//
// # HTTPS 证书相关的配置
//
// plugin_crt_base64 = xxx
// plugin_key_base64 = xxx
type Https2HttpConfig struct {
	HttpConfig // todo location not supported
	CrtBase64  string
	KeyBase64  string
}

func (h *Https2HttpConfig) EnableGroup() bool {
	return false
}

func (h *Https2HttpConfig) ToMap() map[string]string {
	m := make(map[string]string)
	m["type"] = "https"
	m["plugin"] = "https2http"
	if len(h.LocalIp) > 0 {
		if len(h.LocalPort) > 0 {
			m["plugin_local_addr"] = h.LocalIp + ":" + h.LocalPort
		} else {
			m["plugin_local_addr"] = h.LocalIp
		}
	}
	if len(h.Host) > 0 {
		m["custom_domains"] = h.Host
	}
	if len(h.Locations) > 0 {
		m["locations"] = h.Locations
	}
	if len(h.Group) > 0 {
		m["group"] = h.Group
	}
	if len(h.GroupKey) > 0 {
		m["group_key"] = h.GroupKey
	}
	if len(h.CrtBase64) > 0 {
		m["plugin_crt_base64"] = h.CrtBase64
	}
	if len(h.KeyBase64) > 0 {
		m["plugin_key_base64"] = h.KeyBase64
	}
	return m
}

func NewHttpsConfig(m map[string]string) *Https2HttpConfig {
	return &Https2HttpConfig{
		HttpConfig: HttpConfig{
			LocalIp:   m["local_ip"],
			LocalPort: m["local_port"],
			Host:      m["custom_domains"],
			Locations: m["locations"],
			Group:     m["group"],
			GroupKey:  m["group_key"],
		},
		CrtBase64: m["plugin_crt_base64"],
		KeyBase64: m["plugin_key_base64"],
	}
}

func (h *Https2HttpConfig) String() string {
	return h.HttpConfig.String()
}
