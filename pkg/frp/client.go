package frp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/grydovee/ingress-frp/pkg/frp/config"
	"io"
	"net"
	"net/http"
)

type Client interface {
	Addr() *net.TCPAddr
	GetConfigs(ctx context.Context) (*config.Configs, error)
	SetConfig(ctx context.Context, config *config.Configs) error
	Reload(ctx context.Context) error
}

func NewClient(addr net.IP, port uint16, uname, passwd string) Client {
	client := &http.Client{}
	return &frpClient{
		cli:  client,
		addr: &net.TCPAddr{IP: addr, Port: int(port)},
		auth: NewBasicAuth(uname, passwd),
	}
}

type frpClient struct {
	cli  *http.Client
	addr *net.TCPAddr
	auth Auth
}

func (c *frpClient) Addr() *net.TCPAddr {
	return c.addr
}

func (c *frpClient) Reload(ctx context.Context) error {
	request, err := http.NewRequest(ApiReload.Method(), c.buildPath(ApiReload.URI()), nil)
	request.WithContext(ctx)
	if err != nil {
		return err
	}
	if c.auth != nil {
		c.auth.SetAuth(request)
	}
	if err != nil {
		return err
	}
	response, err := c.cli.Do(request)

	if response.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(response.Body)
		if len(msg) > 0 {
			return fmt.Errorf("err code: %d, msg: %s", response.StatusCode, string(msg))
		} else {
			return fmt.Errorf("err code: %d", response.StatusCode)
		}
	}
	return nil
}

func (c *frpClient) GetConfigs(ctx context.Context) (*config.Configs, error) {
	request, err := http.NewRequest(ApiGetConfig.Method(), c.buildPath(ApiGetConfig.URI()), nil)
	request.WithContext(ctx)
	if err != nil {
		return nil, err
	}
	if c.auth != nil {
		c.auth.SetAuth(request)
	}
	if err != nil {
		return nil, err
	}
	response, err := c.cli.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(response.Body)
		if len(msg) > 0 {
			return nil, fmt.Errorf("err code: %d, msg: %s", response.StatusCode, string(msg))
		} else {
			return nil, fmt.Errorf("err code: %d", response.StatusCode)
		}
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	cfg, err := config.Unmarshal(body)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *frpClient) SetConfig(ctx context.Context, configs *config.Configs) error {
	data := config.Marshal(configs)
	request, err := http.NewRequest(ApiPutConfig.Method(), c.buildPath(ApiPutConfig.URI()), io.NopCloser(bytes.NewReader(data)))
	request.WithContext(ctx)
	if err != nil {
		return err
	}
	if c.auth != nil {
		c.auth.SetAuth(request)
	}
	if err != nil {
		return err
	}

	response, err := c.cli.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(response.Body)
		if len(msg) > 0 {
			return fmt.Errorf("err code: %d, msg: %s", response.StatusCode, string(msg))
		} else {
			return fmt.Errorf("err code: %d", response.StatusCode)
		}
	}
	return nil
}

func (c *frpClient) buildPath(api string) string {
	return fmt.Sprintf("http://%s%s", c.addr.String(), api)
}
