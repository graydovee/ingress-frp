package frp

import (
	"fmt"
	"io"
	"net/http"
)

type Client interface {
	GetConfig() ([]byte, error)
	SetConfig(config *Config) error
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

func (c *frpClient) GetConfig() ([]byte, error) {
	request, err := c.buildRequest(ApiGetConfig)
	if err != nil {
		return nil, err
	}
	response, err := c.cli.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("err code: %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *frpClient) SetConfig(config *Config) error {
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
