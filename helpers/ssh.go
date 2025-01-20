package helpers

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
)

// SSHConfig содержит параметры для подключения по SSH
type SSHConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// ExecuteSSHCommand выполняет команду по SSH
func ExecuteSSHCommand(cfg SSHConfig, command string) (string, error) {
	authMethod := ssh.Password(cfg.Password)

	clientConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return "", fmt.Errorf("failed to connect to SSH on %s: %w", addr, err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %s, error: %w", command, err)
	}

	return string(output), nil
}

// CreateDirectory создает директорию на удаленном сервере через SSH
func CreateDirectory(cfg SSHConfig, dirPath string) error {
	cmd := fmt.Sprintf("mkdir -p %s", dirPath)
	_, err := ExecuteSSHCommand(cfg, cmd)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}
	return nil
}
