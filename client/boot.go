package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func (c *HetznerRobotClient) ResetServer(ctx context.Context, serverID int, resetType string) (*HetznerResetResponse, error) {
	endpoint := fmt.Sprintf("/reset/%d", serverID)
	data := url.Values{}
	data.Set("type", resetType)
	resp, err := c.DoRequest("POST", endpoint, strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return nil, fmt.Errorf("error resetting server %d with type %s: %w", serverID, resetType, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d, body: %s", resp.StatusCode, string(body))
	}
	var resetResp HetznerResetResponse
	if err := json.NewDecoder(resp.Body).Decode(&resetResp); err != nil {
		return nil, fmt.Errorf("error parsing reset response: %w", err)
	}
	return &resetResp, nil
}

func (c *HetznerRobotClient) EnableRescueMode(ctx context.Context, serverID int, os string, sshKeys []string) (*HetznerRescueResponse, error) {
	endpoint := fmt.Sprintf("/boot/%d/rescue", serverID)
	data := url.Values{}
	data.Set("os", os)
	for _, key := range sshKeys {
		data.Add("authorized_key[]", key)
	}
	resp, err := c.DoRequest("POST", endpoint, strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return nil, fmt.Errorf("error enabling rescue mode for server %d: %w", serverID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d, body: %s", resp.StatusCode, string(body))
	}
	var rescueResp HetznerRescueResponse
	if err := json.NewDecoder(resp.Body).Decode(&rescueResp); err != nil {
		return nil, fmt.Errorf("error parsing rescue response: %w", err)
	}
	return &rescueResp, nil
}

func (c *HetznerRobotClient) DisableRescueMode(ctx context.Context, serverID int) error {
	endpoint := fmt.Sprintf("/boot/%d/rescue", serverID)
	resp, err := c.DoRequest("DELETE", endpoint, nil, "")
	if err != nil {
		return fmt.Errorf("error disabling rescue mode for server %d: %w", serverID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[DEBUG] Rescue mode disabled for server %d\n", serverID)
	return nil
}

func (c *HetznerRobotClient) RenameServer(ctx context.Context, serverID int, newName string) (*HetznerRenameResponse, error) {
	endpoint := fmt.Sprintf("/server/%d", serverID)
	data := url.Values{}
	data.Set("server_name", newName)
	resp, err := c.DoRequest("POST", endpoint, strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return nil, fmt.Errorf("error renaming server %d to %s: %w", serverID, newName, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d, body: %s", resp.StatusCode, string(body))
	}
	var renameResp HetznerRenameResponse
	if err := json.NewDecoder(resp.Body).Decode(&renameResp); err != nil {
		return nil, fmt.Errorf("error parsing rename response: %w", err)
	}
	return &renameResp, nil
}

func (c *HetznerRobotClient) InstallTalosOS(ctx context.Context, serverIP, password string) error {
	sshConfig := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", serverIP), sshConfig)
	if err != nil {
		return fmt.Errorf("failed to SSH to %s: %w", serverIP, err)
	}
	defer conn.Close()
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func(session *ssh.Session) {
		err := session.Close()
		if err != nil {
		}
	}(session)
	script := `#!/usr/bin/env bash
set -eux

mdadm --stop /dev/md0 || true
mdadm --remove /dev/md0 || true
mdadm --zero-superblock /dev/nvme0n1 /dev/nvme1n1 || true

wipefs --all --force /dev/nvme0n1
wipefs --all --force /dev/nvme1n1

dd if=/dev/zero of=/dev/nvme0n1 bs=1M count=10 || true
dd if=/dev/zero of=/dev/nvme1n1 bs=1M count=10 || true

wget https://factory.talos.dev/image/3531bf15c8738b4bc46f2cdd7c5cd68fea388796b291117f0ee38b51a335fc47/v1.9.2/metal-amd64.raw.zst -O talos.raw.zst
zstd -d talos.raw.zst -o talos.raw

dd if=talos.raw of=/dev/nvme0n1 bs=4M status=progress
sync

reboot
`
	output, err := session.CombinedOutput(script)
	if err != nil {
		return fmt.Errorf("failed to run Talos install script on %s: %s\nerror: %w", serverIP, string(output), err)
	}
	return nil
}

func (c *HetznerRobotClient) RebootServer(ctx context.Context, serverID int, resetType string) error {
	endpoint := fmt.Sprintf("/reset/%d", serverID)
	data := url.Values{}
	data.Set("type", resetType)

	resp, err := c.DoRequest("POST", endpoint, strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
	if err != nil {
		return fmt.Errorf("error rebooting server %d with reset type %s: %w", serverID, resetType, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[DEBUG] Server %d reset with type: %s\n", serverID, resetType)

	if resetType == "power" || resetType == "power_long" {
		time.Sleep(30 * time.Second)
		fmt.Printf("Turning on server %d after %s reset\n", serverID, resetType)
		powerData := url.Values{}
		powerData.Set("action", "on")

		powerResp, err := c.DoRequest("POST", endpoint, strings.NewReader(data.Encode()), "application/x-www-form-urlencoded")
		if err != nil {
			return fmt.Errorf("error turning on server %d after %s reset: %w", serverID, resetType, err)
		}
		defer powerResp.Body.Close()

		if powerResp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(powerResp.Body)
			return fmt.Errorf("unexpected status code %d when turning on server, body: %s", powerResp.StatusCode, string(body))
		}

		fmt.Printf("[DEBUG] Server %d successfully powered on\n", serverID)
	}

	return nil
}

func (c *HetznerRobotClient) WipeAllDisks(ctx context.Context, serverIP, password string) error {
	sshConfig := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", serverIP), sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect via SSH to %s: %w", serverIP, err)
	}
	defer conn.Close()
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()
	script := `#!/usr/bin/env bash
set -eux

for disk in $(lsblk -dn -o NAME,TYPE | awk '$2 == "disk" {print $1}'); do
  echo "Wiping disk /dev/$disk..."
  mdadm --stop /dev/md0 || true
  mdadm --remove /dev/md0 || true
  wipefs --all --force "/dev/$disk"
  dd if=/dev/zero of="/dev/$disk" bs=1M count=10 status=progress || true
done
sync
`
	out, err := session.CombinedOutput(script)
	if err != nil {
		return fmt.Errorf("failed to wipe disks on %s: %s\nerror: %w", serverIP, string(out), err)
	}
	return nil
}
