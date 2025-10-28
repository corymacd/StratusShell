package provision

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/corymacd/StratusShell/internal/audit"
	"github.com/corymacd/StratusShell/internal/validation"
)

// ClaudeSettings represents the structure of Claude Code settings.json
type ClaudeSettings struct {
	Permissions ClaudePermissions           `json:"permissions"`
	MCPServers  map[string]MCPServerConfig  `json:"mcpServers,omitempty"`
}

// ClaudePermissions represents the permissions structure in Claude settings
type ClaudePermissions struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
	Ask   []string `json:"ask"`
}

// MCPServerConfig represents the configuration for an MCP server
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// SetupClaudeConfig configures Claude Code settings for the user
func (p *Provisioner) SetupClaudeConfig() error {
	if !p.config.Claude.Enabled {
		log.Println("Claude Code configuration is disabled, skipping...")
		return nil
	}

	log.Println("Setting up Claude Code configuration...")

	homeDir, err := GetUserHomeDir(p.username)
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to get home directory: %v", err),
		})
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Validate home directory path
	if err := validation.ValidateWorkingDir(homeDir); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("home directory validation failed: %v", err),
		})
		return fmt.Errorf("invalid home directory: %w", err)
	}

	// Create .claude directory
	claudeDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to create .claude directory: %v", err),
			Details: map[string]interface{}{
				"directory": claudeDir,
			},
		})
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Create settings.json
	settingsPath := filepath.Join(claudeDir, "settings.json")
	settings := ClaudeSettings{
		Permissions: ClaudePermissions{
			Allow: p.config.Claude.Allow,
			Deny:  p.config.Claude.Deny,
			Ask:   p.config.Claude.Ask,
		},
	}

	// Add MCP server configurations if any are defined
	if len(p.config.Claude.MCPServers) > 0 {
		settings.MCPServers = make(map[string]MCPServerConfig)
		for _, mcpServer := range p.config.Claude.MCPServers {
			settings.MCPServers[mcpServer.Name] = MCPServerConfig{
				Command: mcpServer.Command,
				Args:    mcpServer.Args,
				Env:     mcpServer.Env,
			}
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to marshal Claude settings: %v", err),
		})
		return fmt.Errorf("failed to marshal Claude settings: %w", err)
	}

	// Write settings.json
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to write Claude settings: %v", err),
			Details: map[string]interface{}{
				"file": settingsPath,
			},
		})
		return fmt.Errorf("failed to write Claude settings: %w", err)
	}

	// Set ownership of .claude directory and contents
	if err := ChownRecursive(claudeDir, p.username); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to set ownership of .claude directory: %v", err),
			Details: map[string]interface{}{
				"directory": claudeDir,
			},
		})
		return fmt.Errorf("failed to set ownership: %w", err)
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"directory": claudeDir,
			"file":      settingsPath,
			"action":    "setup_claude_config",
		},
	})

	log.Printf("Claude Code configuration created at: %s", settingsPath)
	return nil
}

// InstallMCPServers installs configured MCP servers as npm global packages
func (p *Provisioner) InstallMCPServers() error {
	if !p.config.Claude.Enabled {
		log.Println("Claude Code is disabled, skipping MCP server installation...")
		return nil
	}

	if len(p.config.Claude.MCPServers) == 0 {
		log.Println("No MCP servers configured, skipping installation...")
		return nil
	}

	log.Println("Installing MCP servers...")

	installed := 0
	failed := 0

	for _, mcpServer := range p.config.Claude.MCPServers {
		if mcpServer.Package == "" {
			log.Printf("Skipping %s: no package specified", mcpServer.Name)
			continue
		}

		log.Printf("Installing MCP server: %s (%s)", mcpServer.Name, mcpServer.Package)
		if err := p.installMCPServer(mcpServer); err != nil {
			log.Printf("Warning: failed to install %s: %v", mcpServer.Name, err)
			failed++
		} else {
			installed++
		}
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"stage":     "mcp_servers",
			"installed": installed,
			"failed":    failed,
			"total":     len(p.config.Claude.MCPServers),
		},
	})

	log.Printf("Installed %d/%d MCP servers", installed, len(p.config.Claude.MCPServers))
	return nil
}

// installMCPServer installs a single MCP server using npm
func (p *Provisioner) installMCPServer(server MCPServerInstall) error {
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Details: map[string]interface{}{
			"mcp_server": server.Name,
			"package":    server.Package,
		},
	})

	// Execute npm install -g <package>
	cmd := exec.Command("npm", "install", "-g", server.Package)
	err := cmd.Run()

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeFromError(err),
		Details: map[string]interface{}{
			"mcp_server": server.Name,
			"package":    server.Package,
		},
		Error: audit.ErrorString(err),
	})

	return err
}
