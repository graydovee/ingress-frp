package frp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type Client interface {
	GetConfigs() (*Configs, error)
	SetConfig(config *Configs) error
	Reload() error
}

func NewClient(addr string, port uint16, uname, passwd string) Client {
	client := &http.Client{}
	return &frpClient{
		cli:  client,
		addr: fmt.Sprintf("%s:%d", addr, port),
		auth: NewBasicAuth(uname, passwd),
	}
}

type frpClient struct {
	cli  *http.Client
	addr string
	auth Auth
}

func (c *frpClient) Reload() error {
	request, err := http.NewRequest(ApiReload.Method(), c.buildPath(ApiReload.URI()), nil)
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

func (c *frpClient) GetConfigs() (*Configs, error) {
	request, err := http.NewRequest(ApiGetConfig.Method(), c.buildPath(ApiGetConfig.URI()), nil)
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

	cfg, err := Unmarshal(body)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *frpClient) SetConfig(config *Configs) error {
	data := Marshal(config)
	request, err := http.NewRequest(ApiPutConfig.Method(), c.buildPath(ApiPutConfig.URI()), io.NopCloser(bytes.NewReader(data)))
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

func (c *frpClient) buildRequest(api API) (*http.Request, error) {
	request, err := http.NewRequest(api.Method(), c.buildPath(api.URI()), nil)
	if err != nil {
		return nil, err
	}
	if c.auth != nil {
		c.auth.SetAuth(request)
	}
	return request, nil
}

func (c *frpClient) buildPath(api string) string {
	return fmt.Sprintf("http://%s%s", c.addr, api)
}
