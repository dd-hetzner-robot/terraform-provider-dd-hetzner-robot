package client

import (
	"bytes"
	"fmt"
	"net/http"

	"hcloud-robot-provider/shared"
)

type HetznerRobotClient struct {
	Client *http.Client
	Config *shared.ProviderConfig
}

func NewHetznerRobotClient(config *shared.ProviderConfig) *HetznerRobotClient {
	return &HetznerRobotClient{
		Client: &http.Client{},
		Config: config,
	}
}

func (c *HetznerRobotClient) DoRequest(method, path string, body *bytes.Buffer, contentType string) (*http.Response, error) {
	if c.Client == nil || c.Config.URL == "" {
		return nil, fmt.Errorf("client not properly configured")
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.Config.URL, path), body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	return c.Client.Do(req)
}
