package service

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	// Get user home directory
	homeDir := fmt.Sprintf("/home/%s", username)

	// Get current binary path
	binaryPath := getBinaryPath()

	config := ServiceConfig{
		User:       username,
		HomeDir:    homeDir,
		BinaryPath: binaryPath,
		Port:       port,
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
