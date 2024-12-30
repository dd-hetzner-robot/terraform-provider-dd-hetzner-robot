package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type RescueResponse struct {
	Rescue struct {
		ServerIP     string `json:"server_ip"`
		ServerIPv6   string `json:"server_ipv6_net"`
		ServerNumber int    `json:"server_number"`
		OS           string `json:"os"`
		Active       bool   `json:"active"`
		Password     string `json:"password"`
	} `json:"rescue"`
}

func EnableRescueMode(cfg *providerConfig, serverID int) (RescueResponse, error) {
	url := fmt.Sprintf("%s/boot/%d/rescue", cfg.url, serverID)
	payload := "os=linux"

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return RescueResponse{}, fmt.Errorf("failed to create rescue mode request for server %d: %v", serverID, err)
	}
	req.SetBasicAuth(cfg.username, cfg.password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return RescueResponse{}, fmt.Errorf("failed to enable rescue mode for server %d: %v", serverID, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return RescueResponse{}, fmt.Errorf("failed to enable rescue mode for server %d: %s", serverID, resp.Status)
	}

	var result RescueResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return RescueResponse{}, fmt.Errorf("failed to parse rescue mode response for server %d: %v", serverID, err)
	}

	return result, nil
}

func RebootServer(cfg *providerConfig, serverID int) error {
	if err := resetServer(cfg, serverID, "power"); err != nil {
		return fmt.Errorf("failed to power off server %d: %v", serverID, err)
	}

	time.Sleep(10 * time.Second)

	if err := resetServer(cfg, serverID, "power"); err != nil {
		return fmt.Errorf("failed to power on server %d: %v", serverID, err)
	}

	return nil
}

func resetServer(cfg *providerConfig, serverID int, resetType string) error {
	url := fmt.Sprintf("%s/reset/%d", cfg.url, serverID)
	payload := fmt.Sprintf("type=%s", resetType)

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create reset request for server %d: %v", serverID, err)
	}
	req.SetBasicAuth(cfg.username, cfg.password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send reset request for server %d: %v", serverID, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server reset failed for server %d: %s", serverID, resp.Status)
	}

	return nil
}
