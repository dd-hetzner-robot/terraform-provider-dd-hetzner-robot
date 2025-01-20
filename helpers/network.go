package helpers

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

// PingHost проверяет доступность сервера по ICMP (ping)
func PingHost(host string, timeout time.Duration) error {
	cmd := exec.Command("ping", "-c", "1", "-W", fmt.Sprintf("%d", int(timeout.Seconds())), host)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ping failed for host %s: %v", host, err)
	}
	return nil
}

// CheckPortAvailability проверяет доступность указанного порта на хосте (например, SSH)
func CheckPortAvailability(host string, port int, timeout time.Duration) error {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("port %d on host %s is not available: %v", port, host, err)
	}
	defer conn.Close()
	return nil
}

// WaitForServer доступность сервера по ICMP и SSH
func WaitForServer(host string, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if err := PingHost(host, 5*time.Second); err == nil {
			if err := CheckPortAvailability(host, port, 5*time.Second); err == nil {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout waiting for server %s to become available", host)
}
