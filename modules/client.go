package modules

import (
	"fmt"
	"net/http"
)

type HetznerRobotClient struct {
	username string
	password string
	url      string
	client   *http.Client
}

func NewHetznerRobotClient(cfg *providerConfig) *HetznerRobotClient {
	return &HetznerRobotClient{
		username: cfg.username,
		password: cfg.password,
		url:      cfg.url,
		client:   &http.Client{},
	}
}

func (c *HetznerRobotClient) DoRequest(method, endpoint string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.url, endpoint), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.username, c.password)
	return c.client.Do(req)
}
