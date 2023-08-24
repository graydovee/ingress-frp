package config

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
	LocalIp           string `cfg:"local_ip"`
	LocalPort         string `cfg:"local_port"`
	Host              string `cfg:"custom_domains"`
	Locations         string `cfg:"locations"`
	Group             string `cfg:"group"`
	GroupKey          string `cfg:"group_key"`
	Redirect          string `cfg:"redirect"`
	HostHeaderRewrite string `cfg:"host_header_rewrite"`
	HeaderXFromWhere  string `cfg:"header_X-From-Where"`
}

var _ Config = (*HttpConfig)(nil)

func (h *HttpConfig) EnableGroup() bool {
	return true
}

func (h *HttpConfig) ToMap() map[string]string {
	m := ToConfigMap(h)
	m["type"] = TypeHttp
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
		Redirect:  m["redirect"],
	}
}

// Https2HttpConfig
// [test_htts2http]
// type = https
// custom_domains = web.yourdomain.com
//
// plugin = https2http
// plugin_local_addr = 127.0.0.1:3000
//
// # HTTPS 证书相关的配置
//
// plugin_crt_base64 = xxx
// plugin_key_base64 = xxx
type Https2HttpConfig struct {
	HttpConfig        // todo location not supported
	CrtBase64  string `cfg:"plugin_crt_base64"`
	KeyBase64  string `cfg:"plugin_key_base64"`
}

var _ Config = (*Https2HttpConfig)(nil)

func (h *Https2HttpConfig) EnableGroup() bool {
	return false
}

func (h *Https2HttpConfig) ToMap() map[string]string {
	m := ToConfigMap(h)
	m["type"] = TypeHttps
	m["plugin"] = PluginTypeHttps2Http
	if len(h.LocalIp) > 0 {
		if len(h.LocalPort) > 0 {
			m["plugin_local_addr"] = h.LocalIp + ":" + h.LocalPort
		} else {
			m["plugin_local_addr"] = h.LocalIp
		}
	}
	delete(m, "local_ip")
	delete(m, "local_port")
	return m
}

func NewHttps2HttpConfig(m map[string]string) *Https2HttpConfig {
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

// ServerHttpsConfig
// [git]
// type = server_https
// local_port = 4000
// custom_domains = git.graydove.cn
// tls_crts=xxx
// tls_keys=xxx
// group=test
// group_key=test
type ServerHttpsConfig struct {
	HttpConfig
	TlsCrt string `cfg:"tls_crts"`
	TlsKey string `cfg:"tls_keys"`
}

var _ Config = (*ServerHttpsConfig)(nil)

func (h *ServerHttpsConfig) EnableGroup() bool {
	return true
}

func (h *ServerHttpsConfig) ToMap() map[string]string {
	m := ToConfigMap(h)
	m["type"] = TypeServerHttps
	return m
}

func NewServerHttpsConfig(m map[string]string) *ServerHttpsConfig {
	return &ServerHttpsConfig{
		HttpConfig: HttpConfig{
			LocalIp:   m["local_ip"],
			LocalPort: m["local_port"],
			Host:      m["custom_domains"],
			Locations: m["locations"],
			Group:     m["group"],
			GroupKey:  m["group_key"],
		},
		TlsCrt: m["tls_crts"],
		TlsKey: m["tls_keys"],
	}
}

func (h *ServerHttpsConfig) String() string {
	return h.HttpConfig.String()
}

// Https2HttpsConfig
// [test_htts2http]
// type = https
// custom_domains = web.yourdomain.com
//
// plugin = https2http
// plugin_local_addr = 127.0.0.1:3000
//
// # HTTPS 证书相关的配置
//
// plugin_crt_base64 = xxx
// plugin_key_base64 = xxx
type Https2HttpsConfig struct {
	HttpConfig        // todo location not supported
	CrtBase64  string `cfg:"plugin_crt_base64"`
	KeyBase64  string `cfg:"plugin_key_base64"`
}

var _ Config = (*Https2HttpsConfig)(nil)

func (h *Https2HttpsConfig) EnableGroup() bool {
	return false
}

func (h *Https2HttpsConfig) ToMap() map[string]string {
	m := ToConfigMap(h)
	m["type"] = TypeHttps
	m["plugin"] = PluginTypeHttps2Http
	if len(h.LocalIp) > 0 {
		if len(h.LocalPort) > 0 {
			m["plugin_local_addr"] = h.LocalIp + ":" + h.LocalPort
		} else {
			m["plugin_local_addr"] = h.LocalIp
		}
	}
	delete(m, "local_ip")
	delete(m, "local_port")
	return m
}

func NewHttps2HttpsConfig(m map[string]string) *Https2HttpsConfig {
	return &Https2HttpsConfig{
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

// ServerHttps2HttpsConfig
// [git]
// type = server_https
// custom_domains = git.graydove.cn
// tls_crts=xxx
// tls_keys=xxx
// group=test
// group_key=test
// plugin=http2https
// plugin_local_addr=127.0.0.1:8080
type ServerHttps2HttpsConfig struct {
	HttpConfig
	TlsCrt string `cfg:"tls_crts"`
	TlsKey string `cfg:"tls_keys"`
}

var _ Config = (*ServerHttps2HttpsConfig)(nil)

func (h *ServerHttps2HttpsConfig) EnableGroup() bool {
	return true
}

func (h *ServerHttps2HttpsConfig) ToMap() map[string]string {
	m := ToConfigMap(h)
	m["type"] = TypeServerHttps
	m["plugin"] = PluginTypeHttp2Https
	if len(h.LocalIp) > 0 {
		if len(h.LocalPort) > 0 {
			m["plugin_local_addr"] = h.LocalIp + ":" + h.LocalPort
		} else {
			m["plugin_local_addr"] = h.LocalIp
		}
	}
	return m
}

func NewServerHttps2HttpsConfig(m map[string]string) *ServerHttps2HttpsConfig {
	return &ServerHttps2HttpsConfig{
		HttpConfig: HttpConfig{
			LocalIp:   m["local_ip"],
			LocalPort: m["local_port"],
			Host:      m["custom_domains"],
			Locations: m["locations"],
			Group:     m["group"],
			GroupKey:  m["group_key"],
		},
		TlsCrt: m["tls_crts"],
		TlsKey: m["tls_keys"],
	}
}

func (h *ServerHttps2HttpsConfig) String() string {
	return h.HttpConfig.String()
}
