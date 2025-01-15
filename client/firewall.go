package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *HetznerRobotClient) GetFirewall(ctx context.Context, ip string) (*HetznerRobotFirewall, error) {
	fullURL := fmt.Sprintf("%s/firewall/%s", c.Config.BaseURL, ip)

	bytes, err := c.makeAPICall(ctx, "GET", fullURL, nil, []int{http.StatusOK, http.StatusAccepted})
	if err != nil {
		return nil, err
	}

	var fwResp HetznerRobotFirewallResponse
	if err = json.Unmarshal(bytes, &fwResp); err != nil {
		return nil, err
	}
	return &fwResp.Firewall, nil
}

func (c *HetznerRobotClient) SetFirewall(ctx context.Context, firewall HetznerRobotFirewall) error {
	fullURL := fmt.Sprintf("%s/firewall/%s", c.Config.BaseURL, firewall.IP)

	whitelistHOS := "false"
	if firewall.WhitelistHetznerServices {
		whitelistHOS = "true"
	}

	data := url.Values{}
	data.Set("whitelist_hos", whitelistHOS)
	data.Set("status", firewall.Status)

	for idx, rule := range firewall.Rules.Input {
		data.Set(fmt.Sprintf("rules[input][%d][ip_version]", idx), "ipv4")
		if rule.Name != "" {
			data.Set(fmt.Sprintf("rules[input][%d][name]", idx), rule.Name)
		}
		if rule.DstIP != "" {
			data.Set(fmt.Sprintf("rules[input][%d][dst_ip]", idx), rule.DstIP)
		}
		if rule.DstPort != "" {
			data.Set(fmt.Sprintf("rules[input][%d][dst_port]", idx), rule.DstPort)
		}
		if rule.SrcIP != "" {
			data.Set(fmt.Sprintf("rules[input][%d][src_ip]", idx), rule.SrcIP)
		}
		if rule.SrcPort != "" {
			data.Set(fmt.Sprintf("rules[input][%d][src_port]", idx), rule.SrcPort)
		}
		if rule.Protocol != "" {
			data.Set(fmt.Sprintf("rules[input][%d][protocol]", idx), rule.Protocol)
		}
		if rule.TCPFlags != "" {
			data.Set(fmt.Sprintf("rules[input][%d][tcp_flags]", idx), rule.TCPFlags)
		}
		data.Set(fmt.Sprintf("rules[input][%d][action]", idx), rule.Action)
	}

	_, err := c.makeAPICall(ctx, "POST", fullURL, data, []int{http.StatusOK, http.StatusAccepted})
	return err
}

func (c *HetznerRobotClient) makeAPICall(
	ctx context.Context,
	method string,
	fullURL string,
	data url.Values,
	expectedStatus []int,
) ([]byte, error) {

	var body io.Reader
	if data != nil {
		body = strings.NewReader(data.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("makeAPICall: can't create request: %w", err)
	}

	req.SetBasicAuth(c.Config.Username, c.Config.Password)

	if data != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("makeAPICall: request failed: %w", err)
	}
	defer resp.Body.Close()

	isValid := false
	for _, code := range expectedStatus {
		if resp.StatusCode == code {
			isValid = true
			break
		}
	}
	if !isValid {
		out, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("makeAPICall: unexpected status %d, body: %s", resp.StatusCode, string(out))
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("makeAPICall: can't read body: %w", err)
	}
	return out, nil
}
