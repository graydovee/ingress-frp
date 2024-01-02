package frp

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	TypeHttp        = "http"
	TypeHttps       = "https"
	TypeServerHttps = "server_https"

	PluginTypeHttps2Http  = "https2http"
	PluginTypeHttps2Https = "https2https"
	PluginTypeHttp2Https  = "http2https"
)

type Config interface {
	fmt.Stringer
	ToMap() map[string]string
	EnableGroup() bool
}

func Equals(cfg1, cfg2 Config) bool {
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

func ToConfigMap(o any) map[string]string {
	m := make(map[string]string)
	val := reflect.ValueOf(o)
	// Check if s is a pointer and get the underlying element if it is
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Make sure the provided value is a struct
	if val.Kind() != reflect.Struct {
		return m
	}

	processStructFields(val, m)
	return m
}

func processStructFields(val reflect.Value, m map[string]string) {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("cfg")

		// Check if the field is an embedded struct, if so, recurse into it
		if field.Kind() == reflect.Struct {
			processStructFields(field, m)
			continue
		}

		// Otherwise process the field normally
		if tag != "" && field.Kind() == reflect.String && field.String() != "" {
			m[tag] = field.String()
		}
	}
}
