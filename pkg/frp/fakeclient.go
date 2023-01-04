package frp

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
	cfg *Config
}

func NewFakeClient() Client {
	return &fakeClient{}
}

func (f *fakeClient) GetConfig() ([]byte, error) {
	if f.cfg == nil {
		cfg, err := Unmarshal([]byte(defaultConfig))
		if err != nil {
			return nil, err
		}
		f.cfg = cfg
	}
	return Marshal(f.cfg), nil
}

func (f *fakeClient) SetConfig(config *Config) error {
	f.cfg = config
	return nil
}
