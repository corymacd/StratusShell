package provision

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/corymacd/StratusShell/internal/audit"
	"github.com/corymacd/StratusShell/internal/validation"
)

var auditLogger = audit.NewLogger()

func ConfigurePasswordlessSudo(username string) error {
	// Validate username to prevent path traversal and injection
	if err := validation.ValidateUsername(username); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersConfig,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("validation failed: %v", err),
		})
		return fmt.Errorf("invalid username: %w", err)
	}

	sudoersFile := filepath.Join("/etc/sudoers.d", fmt.Sprintf("stratusshell-%s", username))
	content := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", username)

	// Create temporary file for validation
	tmpFile, err := os.CreateTemp("", "sudoers-validate-*")
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersConfig,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to create temp file: %v", err),
		})
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersConfig,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to write temp file: %v", err),
		})
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// Set correct permissions on temp file
	if err := os.Chmod(tmpPath, 0440); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersConfig,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to set permissions: %v", err),
		})
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Validate syntax with visudo
	cmd := exec.Command("visudo", "-c", "-f", tmpPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersConfig,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("sudoers syntax validation failed: %v (output: %s)", err, output),
		})
		return fmt.Errorf("sudoers syntax validation failed: %w (output: %s)", err, output)
	}

	// Validation passed, copy to final location
	input, err := os.ReadFile(tmpPath)
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersConfig,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to read validated file: %v", err),
		})
		return fmt.Errorf("failed to read validated file: %w", err)
	}

	// Write with 0440 permissions (required for sudoers files)
	if err := os.WriteFile(sudoersFile, input, 0440); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersConfig,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to write sudoers file: %v", err),
		})
		return fmt.Errorf("failed to write sudoers file: %w", err)
	}

	// Success - log audit entry
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionSudoersConfig,
		Actor:   "system",
		Target:  username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"file": sudoersFile,
		},
	})

	return nil
}

func RemoveSudoersConfig(username string) error {
	// Validate username to prevent path traversal
	if err := validation.ValidateUsername(username); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersRemove,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("validation failed: %v", err),
		})
		return fmt.Errorf("invalid username: %w", err)
	}

	sudoersFile := filepath.Join("/etc/sudoers.d", fmt.Sprintf("stratusshell-%s", username))

	if err := os.Remove(sudoersFile); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionSudoersRemove,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("failed to remove sudoers file: %v", err),
		})
		return err
	}

	// Success - log audit entry
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionSudoersRemove,
		Actor:   "system",
		Target:  username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"file": sudoersFile,
		},
	})

	return nil
}
