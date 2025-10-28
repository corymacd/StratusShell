package provision

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeConfigMarshaling(t *testing.T) {
	settings := ClaudeSettings{
		Permissions: ClaudePermissions{
			Allow: []string{"gh", "npm"},
			Deny:  []string{"rm", "dd"},
			Ask:   []string{"git"},
		},
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("failed to marshal settings: %v", err)
	}

	var unmarshaled ClaudeSettings
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	if len(unmarshaled.Permissions.Allow) != 2 {
		t.Errorf("expected 2 allow entries, got %d", len(unmarshaled.Permissions.Allow))
	}
	if len(unmarshaled.Permissions.Deny) != 2 {
		t.Errorf("expected 2 deny entries, got %d", len(unmarshaled.Permissions.Deny))
	}
	if len(unmarshaled.Permissions.Ask) != 1 {
		t.Errorf("expected 1 ask entry, got %d", len(unmarshaled.Permissions.Ask))
	}
}

func TestClaudeConfigCreation(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create mock provisioner with minimal config
	config := &Config{
		Claude: ClaudeConfig{
			Enabled: true,
			Allow:   []string{"gh"},
			Deny:    []string{},
			Ask:     []string{},
		},
	}

	// Manually create the .claude directory and settings.json
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("failed to create .claude directory: %v", err)
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	settings := ClaudeSettings{
		Permissions: ClaudePermissions{
			Allow: config.Claude.Allow,
			Deny:  config.Claude.Deny,
			Ask:   config.Claude.Ask,
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal settings: %v", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		t.Fatalf("failed to write settings file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Fatal("settings.json was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("failed to read settings file: %v", err)
	}

	var readSettings ClaudeSettings
	if err := json.Unmarshal(content, &readSettings); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	if len(readSettings.Permissions.Allow) != 1 || readSettings.Permissions.Allow[0] != "gh" {
		t.Errorf("expected allow=['gh'], got %v", readSettings.Permissions.Allow)
	}
}

func TestClaudeConfigDefaults(t *testing.T) {
	settings := ClaudeSettings{
		Permissions: ClaudePermissions{
			Allow: []string{},
			Deny:  []string{},
			Ask:   []string{},
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal default settings: %v", err)
	}

	var unmarshaled ClaudeSettings
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	// Verify empty arrays are preserved
	if unmarshaled.Permissions.Allow == nil {
		t.Error("expected non-nil Allow array")
	}
	if unmarshaled.Permissions.Deny == nil {
		t.Error("expected non-nil Deny array")
	}
	if unmarshaled.Permissions.Ask == nil {
		t.Error("expected non-nil Ask array")
	}
}

func TestLoadConfigWithClaude(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `user:
  shell: /bin/bash

base_packages:
  - git

cloud:
  aws: false
  gcloud: false
  kubectl: false
  docker: false
  terraform: false

languages:
  go:
    enabled: false
    version: ""
    tools: []
  node:
    enabled: false
    version: ""
    package_manager: ""
    global_packages: []

shell:
  zsh: false
  oh_my_zsh: false
  tmux: false

claude:
  enabled: true
  allow:
    - gh
    - npm
  deny:
    - rm
  ask:
    - git
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify Claude config was loaded
	if !config.Claude.Enabled {
		t.Error("expected claude.enabled to be true")
	}
	if len(config.Claude.Allow) != 2 {
		t.Errorf("expected 2 allow entries, got %d", len(config.Claude.Allow))
	}
	if len(config.Claude.Deny) != 1 {
		t.Errorf("expected 1 deny entry, got %d", len(config.Claude.Deny))
	}
	if len(config.Claude.Ask) != 1 {
		t.Errorf("expected 1 ask entry, got %d", len(config.Claude.Ask))
	}
}

func TestClaudeConfigWithMCPServers(t *testing.T) {
	settings := ClaudeSettings{
		Permissions: ClaudePermissions{
			Allow: []string{"gh"},
			Deny:  []string{},
			Ask:   []string{},
		},
		MCPServers: map[string]MCPServerConfig{
			"playwright": {
				Command: "npx",
				Args:    []string{"-y", "@playwright/mcp"},
			},
			"github": {
				Command: "npx",
				Args:    []string{"-y", "github-mcp-server"},
			},
			"linear": {
				Command: "npx",
				Args:    []string{"-y", "@mseep/linear-mcp"},
			},
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal settings with MCP servers: %v", err)
	}

	var unmarshaled ClaudeSettings
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal settings: %v", err)
	}

	if len(unmarshaled.MCPServers) != 3 {
		t.Errorf("expected 3 MCP servers, got %d", len(unmarshaled.MCPServers))
	}

	// Verify playwright config
	playwright, ok := unmarshaled.MCPServers["playwright"]
	if !ok {
		t.Error("playwright MCP server not found")
	} else {
		if playwright.Command != "npx" {
			t.Errorf("expected playwright command to be 'npx', got '%s'", playwright.Command)
		}
		if len(playwright.Args) != 2 {
			t.Errorf("expected 2 args for playwright, got %d", len(playwright.Args))
		}
	}

	// Verify github config
	github, ok := unmarshaled.MCPServers["github"]
	if !ok {
		t.Error("github MCP server not found")
	} else {
		if github.Command != "npx" {
			t.Errorf("expected github command to be 'npx', got '%s'", github.Command)
		}
	}

	// Verify linear config
	linear, ok := unmarshaled.MCPServers["linear"]
	if !ok {
		t.Error("linear MCP server not found")
	} else {
		if linear.Command != "npx" {
			t.Errorf("expected linear command to be 'npx', got '%s'", linear.Command)
		}
	}
}

func TestLoadConfigWithMCPServers(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config-mcp.yaml")

	configContent := `user:
  shell: /bin/bash

base_packages:
  - git

cloud:
  aws: false
  gcloud: false
  kubectl: false
  docker: false
  terraform: false

languages:
  go:
    enabled: false
    version: ""
    tools: []
  node:
    enabled: false
    version: ""
    package_manager: ""
    global_packages: []

shell:
  zsh: false
  oh_my_zsh: false
  tmux: false

claude:
  enabled: true
  allow:
    - gh
  deny: []
  ask: []
  mcp_servers:
    - name: "playwright"
      package: "@playwright/mcp"
      command: "npx"
      args:
        - "-y"
        - "@playwright/mcp"
    - name: "github"
      package: "github-mcp-server"
      command: "npx"
      args:
        - "-y"
        - "github-mcp-server"
    - name: "linear"
      package: "@mseep/linear-mcp"
      command: "npx"
      args:
        - "-y"
        - "@mseep/linear-mcp"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Load config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify MCP servers were loaded
	if len(config.Claude.MCPServers) != 3 {
		t.Errorf("expected 3 MCP servers, got %d", len(config.Claude.MCPServers))
	}

	// Verify playwright config
	var playwrightFound bool
	for _, mcp := range config.Claude.MCPServers {
		if mcp.Name == "playwright" {
			playwrightFound = true
			if mcp.Package != "@playwright/mcp" {
				t.Errorf("expected playwright package to be '@playwright/mcp', got '%s'", mcp.Package)
			}
			if mcp.Command != "npx" {
				t.Errorf("expected playwright command to be 'npx', got '%s'", mcp.Command)
			}
			if len(mcp.Args) != 2 {
				t.Errorf("expected 2 args for playwright, got %d", len(mcp.Args))
			}
		}
	}
	if !playwrightFound {
		t.Error("playwright MCP server not found in config")
	}
}
