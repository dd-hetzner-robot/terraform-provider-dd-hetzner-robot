package client

import (
	"fmt"
	"hcloud-robot-provider/shared"
	"io"
	"net/http"
)

func NewHetznerRobotClient(config *shared.ProviderConfig) *HetznerRobotClient {
	return &HetznerRobotClient{
		Config: config,
		Client: &http.Client{},
	}
}

func (c *HetznerRobotClient) DoRequest(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.Config.BaseURL, path), body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	return resp, nil
}
