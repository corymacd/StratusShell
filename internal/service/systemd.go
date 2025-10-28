package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"text/template"
)

const serviceTemplate = `[Unit]
Description=StratusShell for {{.User}}
After=network.target

[Service]
Type=simple
User={{.User}}
WorkingDirectory={{.HomeDir}}
ExecStart={{.BinaryPath}} serve --user={{.User}} --port={{.Port}}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

type ServiceConfig struct {
	User       string
	HomeDir    string
	BinaryPath string
	Port       int
}

// Helper functions for testability
func getServiceName(username string) string {
	return fmt.Sprintf("stratusshell-%s.service", username)
}

func getServicePath(serviceName string) string {
	return filepath.Join("/etc/systemd/system", serviceName)
}

func getBinaryPath() string {
	binaryPath, err := os.Executable()
	if err != nil {
		// Default to /usr/local/bin
		return "/usr/local/bin/stratusshell"
	}
	return binaryPath
}

func getUserHomeDir(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", fmt.Errorf("failed to lookup user %q: %w", username, err)
	}
	return u.HomeDir, nil
}

var validUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validateServiceConfig(config ServiceConfig) error {
	// Validate username - should not contain special characters that could cause issues
	if config.User == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if !validUsernameRegex.MatchString(config.User) {
		return fmt.Errorf("username contains invalid characters. Only alphanumeric, underscore, and dash are allowed")
	}
	
	// Validate port range
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", config.Port)
	}
	
	// Validate paths exist
	if config.HomeDir == "" {
		return fmt.Errorf("home directory cannot be empty")
	}
	
	if config.BinaryPath == "" {
		return fmt.Errorf("binary path cannot be empty")
	}
	
	return nil
}

func parseServiceTemplate() (*template.Template, error) {
	return template.New("service").Parse(serviceTemplate)
}

func generateServiceContent(tmpl *template.Template, config ServiceConfig) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func InstallSystemdService(username string, port int) error {
	// Get user home directory from the system
	homeDir, err := getUserHomeDir(username)
	if err != nil {
		return err
	}

	// Get current binary path
	binaryPath := getBinaryPath()

	config := ServiceConfig{
		User:       username,
		HomeDir:    homeDir,
		BinaryPath: binaryPath,
		Port:       port,
	}

	// Validate configuration
	if err := validateServiceConfig(config); err != nil {
		return fmt.Errorf("invalid service configuration: %w", err)
	}

	// Generate service file
	serviceName := getServiceName(username)
	servicePath := getServicePath(serviceName)

	tmpl, err := parseServiceTemplate()
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, config); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", serviceName).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Start service
	if err := exec.Command("systemctl", "start", serviceName).Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func UninstallSystemdService(username string) error {
	serviceName := getServiceName(username)
	servicePath := getServicePath(serviceName)

	// Stop service
	exec.Command("systemctl", "stop", serviceName).Run()

	// Disable service
	exec.Command("systemctl", "disable", serviceName).Run()

	// Remove service file
	os.Remove(servicePath)

	// Reload systemd
	exec.Command("systemctl", "daemon-reload").Run()

	return nil
}

func GetServiceStatus(username string) (string, error) {
	serviceName := getServiceName(username)

	cmd := exec.Command("systemctl", "status", serviceName)
	output, err := cmd.CombinedOutput()

	return string(output), err
}
