package frp

import (
	"gopkg.in/ini.v1"
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
