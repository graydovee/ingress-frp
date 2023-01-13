package frp

import (
	"bytes"
	"gopkg.in/ini.v1"
	"sort"
	"strings"
)

var iniOptions = ini.LoadOptions{
	Insensitive:         false,
	InsensitiveSections: false,
	InsensitiveKeys:     false,
	IgnoreInlineComment: true,
	AllowBooleanKeys:    true,
}

type Configs struct {
	Common MapConfig
	Proxy  map[string]Config
}

func Unmarshal(data []byte) (*Configs, error) {
	f, err := ini.LoadSources(iniOptions, data)
	if err != nil {
		return nil, err
	}

	c := &Configs{
		Proxy: make(map[string]Config),
	}
	for _, section := range f.Sections() {
		name := section.Name()

		if name == "common" {
			c.Common = section.KeysHash()
			continue
		}
		if name == ini.DefaultSection || strings.HasPrefix(name, "range:") {
			continue
		}

		c.Proxy[name] = MapConfig(section.KeysHash())
	}
	return c, nil
}

type Ini struct {
	b *bytes.Buffer
}

func NewIni() *Ini {
	return &Ini{b: bytes.NewBuffer(nil)}
}

func (i *Ini) Write(name string, config Config) {
	writeHead(i.b, name)
	foreach(config.ToMap(), func(k string, v string) bool {
		i.b.WriteString(k)
		i.b.WriteByte('=')
		i.b.WriteString(v)
		i.b.WriteByte('\n')
		return true
	})
	i.b.WriteByte('\n')
}

func (i *Ini) Bytes() []byte {
	return i.b.Bytes()
}

func Marshal(config *Configs) []byte {
	if config == nil {
		return nil
	}

	i := NewIni()
	if config.Common != nil {
		i.Write("common", config.Common)
	}
	if config.Proxy != nil {
		foreach(config.Proxy, func(name string, proxy Config) bool {
			i.Write(name, proxy)
			return true
		})
	}
	return i.Bytes()
}

func writeHead(buffer *bytes.Buffer, key string) {
	buffer.WriteByte('[')
	buffer.Write([]byte(key))
	buffer.WriteString("]\n")
}

func foreach[V any](m map[string]V, f func(k string, v V) bool) {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !f(k, m[k]) {
			break
		}
	}
}
