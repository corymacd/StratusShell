package provision

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/corymacd/StratusShell/internal/audit"
	"github.com/corymacd/StratusShell/internal/validation"
)

// SetupShellEnvironment configures the shell environment for the user
func (p *Provisioner) SetupShellEnvironment() error {
	log.Println("Setting up shell environment...")

	if p.config.Shell.Zsh {
		if err := p.installZsh(); err != nil {
			log.Printf("Warning: failed to install zsh: %v", err)
		}
	}

	if p.config.Shell.Tmux {
		if err := p.installTmux(); err != nil {
			log.Printf("Warning: failed to install tmux: %v", err)
		}
	}

	// Configure shell RC files to source stratusshell env
	if err := p.configureShellRC(); err != nil {
		log.Printf("Warning: failed to configure shell RC: %v", err)
	}

	return nil
}

// installZsh installs zsh and sets it as the default shell
func (p *Provisioner) installZsh() error {
	log.Println("Installing zsh...")

	if err := p.pm.Install("zsh"); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("zsh installation failed: %v", err),
			Details: map[string]interface{}{
				"tool": "zsh",
			},
		})
		return err
	}

	// Set zsh as default shell
	if err := SetUserShell(p.username, "/bin/zsh"); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to set zsh as default shell: %v", err),
			Details: map[string]interface{}{
				"tool": "zsh",
			},
		})
		return err
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"tool":  "zsh",
			"shell": "/bin/zsh",
		},
	})

	return nil
}

// installTmux installs tmux terminal multiplexer
func (p *Provisioner) installTmux() error {
	log.Println("Installing tmux...")

	err := p.pm.Install("tmux")
	
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeFromError(err),
		Details: map[string]interface{}{
			"tool": "tmux",
		},
		Error: audit.ErrorString(err),
	})

	return err
}

// configureShellRC updates shell RC files to source StratusShell environment
func (p *Provisioner) configureShellRC() error {
	homeDir, err := GetUserHomeDir(p.username)
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to get home directory: %v", err),
		})
		return err
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

	// Determine which RC file to update
	rcFile := filepath.Join(homeDir, ".bashrc")
	if p.config.Shell.Zsh {
		rcFile = filepath.Join(homeDir, ".zshrc")
	}

	// Validate RC file path
	if err := validation.ValidateWorkingDir(filepath.Dir(rcFile)); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("RC file directory validation failed: %v", err),
		})
		return fmt.Errorf("invalid RC file directory: %w", err)
	}

	// Content to append
	sourceCmd := "\n# StratusShell Environment\nif [ -f ~/.stratusshell/env.sh ]; then\n    source ~/.stratusshell/env.sh\nfi\n"

	// Check if the source command already exists
	if _, err := os.Stat(rcFile); err == nil {
		content, err := os.ReadFile(rcFile)
		if err != nil {
			auditLogger.Log(audit.Entry{
				Action:  audit.ActionToolInstall,
				Actor:   "system",
				Target:  p.username,
				Outcome: audit.OutcomeFailure,
				Error:   fmt.Sprintf("failed to read RC file: %v", err),
				Details: map[string]interface{}{
					"rc_file": rcFile,
				},
			})
			return err
		}

		// Check if already configured
		if contains(string(content), "StratusShell Environment") {
			log.Printf("Shell RC already configured: %s", rcFile)
			return nil
		}
	}

	// Open file for appending
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to open RC file: %v", err),
			Details: map[string]interface{}{
				"rc_file": rcFile,
			},
		})
		return err
	}
	defer f.Close()

	// Append source command
	if _, err = f.WriteString(sourceCmd); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionToolInstall,
			Actor:   "system",
			Target:  p.username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to write to RC file: %v", err),
			Details: map[string]interface{}{
				"rc_file": rcFile,
			},
		})
		return err
	}

	// Set ownership
	if err := ChownRecursive(rcFile, p.username); err != nil {
		return err
	}

	auditLogger.Log(audit.Entry{
		Action:  audit.ActionToolInstall,
		Actor:   "system",
		Target:  p.username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"rc_file": rcFile,
			"action":  "configured_shell_rc",
		},
	})

	log.Printf("Configured shell RC file: %s", rcFile)
	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		findSubstring(s, substr)))
}

// findSubstring searches for a substring within a string
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
