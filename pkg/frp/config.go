package frp

type Config interface {
	ToMap() map[string]string
}

type MapConfig map[string]string

func (p MapConfig) ToMap() map[string]string {
	m := make(map[string]string, len(p))
	for k, v := range p {
		m[k] = v
	}
	return m
}

// HttpConfig
// [web01]
// type = http
// local_port = 80
// local_ip = 80
// custom_domains = web.yourdomain.com
// locations = /
type HttpConfig struct {
	LocalIp   string
	LocalPort string
	Host      string
	Locations string
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
	return m
}

func NewHttpConfig(m map[string]string) *HttpConfig {
	return &HttpConfig{
		LocalIp:   m["local_ip"],
		LocalPort: m["local_port"],
		Host:      m["custom_domains"],
		Locations: m["locations"],
	}
}
