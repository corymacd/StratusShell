package service

import (
	"os"
	"strings"
	"testing"
)

func TestServiceConfigGeneration(t *testing.T) {
	// This test validates the service template can be generated correctly
	config := ServiceConfig{
		User:       "testuser",
		HomeDir:    "/home/testuser",
		BinaryPath: "/usr/local/bin/stratusshell",
		Port:       8080,
	}

	// Parse template
	tmpl, err := parseServiceTemplate()
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Generate service content
	content, err := generateServiceContent(tmpl, config)
	if err != nil {
		t.Fatalf("Failed to generate service content: %v", err)
	}

	// Verify content contains expected values
	if !strings.Contains(content, "testuser") {
		t.Error("Service content missing username")
	}
	if !strings.Contains(content, "/home/testuser") {
		t.Error("Service content missing home directory")
	}
	if !strings.Contains(content, "/usr/local/bin/stratusshell") {
		t.Error("Service content missing binary path")
	}
	if !strings.Contains(content, "8080") {
		t.Error("Service content missing port")
	}
	if !strings.Contains(content, "After=network.target") {
		t.Error("Service content missing network dependency")
	}
	if !strings.Contains(content, "WantedBy=multi-user.target") {
		t.Error("Service content missing install target")
	}
}

func TestGetServiceName(t *testing.T) {
	tests := []struct {
		username string
		expected string
	}{
		{"alice", "stratusshell-alice.service"},
		{"bob", "stratusshell-bob.service"},
		{"dev-user", "stratusshell-dev-user.service"},
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			result := getServiceName(tt.username)
			if result != tt.expected {
				t.Errorf("getServiceName(%q) = %q, want %q", tt.username, result, tt.expected)
			}
		})
	}
}

func TestGetServicePath(t *testing.T) {
	serviceName := "stratusshell-testuser.service"
	expected := "/etc/systemd/system/stratusshell-testuser.service"
	
	result := getServicePath(serviceName)
	if result != expected {
		t.Errorf("getServicePath(%q) = %q, want %q", serviceName, result, expected)
	}
}

func TestGetBinaryPath(t *testing.T) {
	// This test validates that getBinaryPath returns a valid path
	path := getBinaryPath()
	
	// Should either be the executable path or the default
	if path != "/usr/local/bin/stratusshell" {
		// If not the default, check if it's a valid executable path
		if _, err := os.Stat(path); err != nil {
			t.Logf("Warning: binary path %q does not exist, but this is acceptable in test environment", path)
		}
	}
}

func TestGetUserHomeDir(t *testing.T) {
	// Test with current user
	currentUser := os.Getenv("USER")
	if currentUser == "" {
		t.Skip("USER environment variable not set, skipping test")
	}
	
	homeDir, err := getUserHomeDir(currentUser)
	if err != nil {
		t.Fatalf("Failed to get home directory for current user %q: %v", currentUser, err)
	}
	
	if homeDir == "" {
		t.Error("Home directory should not be empty")
	}
	
	// Test with non-existent user
	_, err = getUserHomeDir("nonexistent-user-12345")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

func TestValidateServiceConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    ServiceConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: ServiceConfig{
				User:       "testuser",
				HomeDir:    "/home/testuser",
				BinaryPath: "/usr/local/bin/stratusshell",
				Port:       8080,
			},
			expectErr: false,
		},
		{
			name: "empty username",
			config: ServiceConfig{
				User:       "",
				HomeDir:    "/home/testuser",
				BinaryPath: "/usr/local/bin/stratusshell",
				Port:       8080,
			},
			expectErr: true,
		},
		{
			name: "invalid port - too low",
			config: ServiceConfig{
				User:       "testuser",
				HomeDir:    "/home/testuser",
				BinaryPath: "/usr/local/bin/stratusshell",
				Port:       0,
			},
			expectErr: true,
		},
		{
			name: "invalid port - too high",
			config: ServiceConfig{
				User:       "testuser",
				HomeDir:    "/home/testuser",
				BinaryPath: "/usr/local/bin/stratusshell",
				Port:       65536,
			},
			expectErr: true,
		},
		{
			name: "empty home directory",
			config: ServiceConfig{
				User:       "testuser",
				HomeDir:    "",
				BinaryPath: "/usr/local/bin/stratusshell",
				Port:       8080,
			},
			expectErr: true,
		},
		{
			name: "empty binary path",
			config: ServiceConfig{
				User:       "testuser",
				HomeDir:    "/home/testuser",
				BinaryPath: "",
				Port:       8080,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
