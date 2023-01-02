package frp

import "net/http"

type Auth interface {
	SetAuth(r *http.Request)
}

type BasicAuth struct {
	Username string
	Password string
}

func NewBasicAuth(username string, password string) Auth {
	return &BasicAuth{Username: username, Password: password}
}

func (b *BasicAuth) SetAuth(r *http.Request) {
	r.SetBasicAuth(b.Username, b.Password)
}
