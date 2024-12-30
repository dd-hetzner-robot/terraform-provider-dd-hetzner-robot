package modules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HetznerRobotFirewallResponse struct {
	Firewall HetznerRobotFirewall `json:"firewall"`
}

type HetznerRobotFirewall struct {
	IP                       string                    `json:"server_ip"`
	WhitelistHetznerServices bool                      `json:"whitelist_hos"`
	Status                   string                    `json:"status"`
	Rules                    HetznerRobotFirewallRules `json:"rules"`
}

type HetznerRobotFirewallRules struct {
	Input []HetznerRobotFirewallRule `json:"input"`
}

type HetznerRobotFirewallRule struct {
	Name     string `json:"name,omitempty"`
	DstIP    string `json:"dst_ip,omitempty"`
	DstPort  string `json:"dst_port,omitempty"`
	SrcIP    string `json:"src_ip,omitempty"`
	SrcPort  string `json:"src_port,omitempty"`
	Protocol string `json:"protocol"`
	Action   string `json:"action"`
}

func (c *HetznerRobotClient) getFirewall(ctx context.Context, ip string) (*HetznerRobotFirewall, error) {
	reqURL := fmt.Sprintf("%s/firewall/%s", c.url, ip)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response HetznerRobotFirewallResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.Firewall, nil
}

func (c *HetznerRobotClient) setFirewall(ctx context.Context, firewall HetznerRobotFirewall) error {
	reqURL := fmt.Sprintf("%s/firewall/%s", c.url, firewall.IP)

	formData := url.Values{}
	formData.Set("whitelist_hos", fmt.Sprintf("%t", firewall.WhitelistHetznerServices))
	formData.Set("status", firewall.Status)

	for idx, rule := range firewall.Rules.Input {
		formData.Set(fmt.Sprintf("rules[input][%d][protocol]", idx), rule.Protocol)
		formData.Set(fmt.Sprintf("rules[input][%d][action]", idx), rule.Action)

		if rule.Name != "" {
			formData.Set(fmt.Sprintf("rules[input][%d][name]", idx), rule.Name)
		}
		if rule.DstPort != "" {
			formData.Set(fmt.Sprintf("rules[input][%d][dst_port]", idx), rule.DstPort)
		}
		if rule.DstIP != "" {
			formData.Set(fmt.Sprintf("rules[input][%d][dst_ip]", idx), rule.DstIP)
		}
		if rule.SrcPort != "" {
			formData.Set(fmt.Sprintf("rules[input][%d][src_port]", idx), rule.SrcPort)
		}
		if rule.SrcIP != "" {
			formData.Set(fmt.Sprintf("rules[input][%d][src_ip]", idx), rule.SrcIP)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
