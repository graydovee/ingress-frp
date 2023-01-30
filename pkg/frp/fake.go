package frp

import (
	"context"
	"net"
)

const defaultConfig = `
[common]
admin_port=7400
admin_user=admin
admin_pwd=admin
server_addr=8.8.8.8
server_port=7000
admin_addr=127.0.0.1

[ssh]
local_port=22
remote_port=22
type=tcp
local_ip=127.0.0.1

[kube-apiserver]
type=tcp
local_ip=127.0.0.1
local_port=9443
remote_port=9443
group=kube-apiserver

[web]
local_ip=0.0.0.0
remote_port=80
local_port=8080
`

type fakeClient struct {
	cfg *Configs
}

func NewFakeClient() Client {
	return &fakeClient{}
}

func (f *fakeClient) GetConfigs(ctx context.Context) (*Configs, error) {
	if f.cfg == nil {
		cfg, err := Unmarshal([]byte(defaultConfig))
		if err != nil {
			return nil, err
		}
		f.cfg = cfg
	}
	bytes := Marshal(f.cfg)
	return Unmarshal(bytes)
}

func (f *fakeClient) SetConfig(ctx context.Context, config *Configs) error {
	f.cfg = config
	return nil
}

func (f *fakeClient) Reload(ctx context.Context) error {
	return nil
}

func (f *fakeClient) Addr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 7000,
	}
}
func NewFakeSyncer() Syncer {
	return &syncer{
		clients: map[string]Client{
			"127.0.0.1": NewFakeClient(),
		},
	}
}
