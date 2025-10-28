package provision

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/corymacd/StratusShell/internal/audit"
	"github.com/corymacd/StratusShell/internal/validation"
)

// ClaudeSettings represents the structure of Claude Code settings.json
type ClaudeSettings struct {
	Permissions ClaudePermissions `json:"permissions"`
}

// ClaudePermissions represents the permissions structure in Claude settings
type ClaudePermissions struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
	Ask   []string `json:"ask"`
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
