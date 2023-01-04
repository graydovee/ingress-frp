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

type Config struct {
	Common map[string]string
	Proxy  map[string]Proxy
}

type Proxy map[string]string

func Unmarshal(data []byte) (*Config, error) {
	f, err := ini.LoadSources(iniOptions, data)
	if err != nil {
		return nil, err
	}

	c := &Config{
		Proxy: make(map[string]Proxy),
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

		c.Proxy[name] = section.KeysHash()
	}
	return c, nil
}

func Marshal(config *Config) []byte {
	if config == nil {
		return nil
	}

	buffer := bytes.NewBuffer(nil)
	if config.Common != nil {
		marshalMap(buffer, "common", config.Common)
	}
	if config.Proxy != nil {
		Foreach(config.Proxy, func(name string, proxy Proxy) bool {
			marshalMap(buffer, name, proxy)
			return true
		})
	}
	return buffer.Bytes()
}

func marshalMap(buffer *bytes.Buffer, headName string, m map[string]string) {
	buffer.WriteByte('[')
	buffer.Write([]byte(headName))
	buffer.WriteString("]\n")
	Foreach(m, func(k string, v string) bool {
		buffer.WriteString(k)
		buffer.WriteByte('=')
		buffer.WriteString(v)
		buffer.WriteByte('\n')
		return true
	})
	buffer.WriteByte('\n')
	return
}

func Foreach[V any](m map[string]V, f func(k string, v V) bool) {
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
