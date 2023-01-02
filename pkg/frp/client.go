package frp

import (
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	cli  *http.Client
	addr string
	auth Auth
}

func NewClient() *Client {
	client := &http.Client{}
	return &Client{
		cli:  client,
		addr: "127.0.0.1:7400",
		auth: NewBasicAuth("admin", "admin"),
	}
}

func (c *Client) GetConfig() ([]byte, error) {
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

func (c *Client) SetConfig() {
}

func (c *Client) buildRequest(api API) (*http.Request, error) {
	request, err := http.NewRequest(api.Method(), c.buildPath(api.URI()), nil)
	if err != nil {
		return nil, err
	}
	if c.auth != nil {
		c.auth.SetAuth(request)
	}
	return request, nil
}

func (c *Client) buildPath(api string) string {
	return fmt.Sprintf("http://%s%s", c.addr, api)
}
