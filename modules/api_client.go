package modules

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type Server struct {
	IP         string `json:"server_ip"`
	IPv6Net    string `json:"server_ipv6_net"`
	Number     int    `json:"server_number"`
	Name       string `json:"server_name"`
	Product    string `json:"product"`
	Datacenter string `json:"dc"`
	Traffic    string `json:"traffic"`
	Status     string `json:"status"`
	Cancelled  bool   `json:"cancelled"`
	PaidUntil  string `json:"paid_until"`
}

func fetchAllServers(cfg *providerConfig) ([]Server, error) {
	url := fmt.Sprintf("%s/server", cfg.url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cfg.username, cfg.password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch all servers: %s", resp.Status)
	}

	var rawServers []struct {
		Server Server `json:"server"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawServers); err != nil {
		return nil, err
	}

	servers := make([]Server, len(rawServers))
	for i, raw := range rawServers {
		servers[i] = raw.Server
	}

	return servers, nil
}

func fetchServerByID(cfg *providerConfig, id int) (Server, error) {
	url := fmt.Sprintf("%s/server/%d", cfg.url, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Server{}, err
	}
	req.SetBasicAuth(cfg.username, cfg.password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Server{}, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var result struct {
		Server Server `json:"server"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Server{}, err
	}
	return result.Server, nil
}

func renameServer(cfg *providerConfig, id int, newName string) error {
	url := fmt.Sprintf("%s/server/%d", cfg.url, id)
	payload := strings.NewReader("server_name=" + newName)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return err
	}
	req.SetBasicAuth(cfg.username, cfg.password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to rename server: %s", resp.Status)
	}
	return nil
}

func ExecuteSSHCommand(ip, username, password, command string) (string, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, "22"), config)
	if err != nil {
		return "", fmt.Errorf("failed to establish SSH connection to %s: %v", ip, err)
	}
	defer func(client *ssh.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil {

		}
	}(session)

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v, output: %s", err, output)
	}

	return string(output), nil
}

func WaitForHost(ip string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		pingCmd := exec.Command("ping", "-c", "1", ip)
		if err := pingCmd.Run(); err == nil {
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, "22"), 5*time.Second)
			if err == nil {
				err := conn.Close()
				if err != nil {
					return err
				}
				return nil // Host is available
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("host %s is not reachable within the timeout", ip)
		}

		time.Sleep(5 * time.Second)
	}
}

func IsTalosInstalled(ip string) (bool, error) {
	url := fmt.Sprintf("https://%s:50000/healthz", ip)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false, nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	return false, nil
}
