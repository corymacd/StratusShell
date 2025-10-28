package provision

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"

	"github.com/corymacd/StratusShell/internal/audit"
	"github.com/corymacd/StratusShell/internal/validation"
)

func CreateUser(username, shell string) error {
	// Validate username to prevent injection and path traversal
	if err := validation.ValidateUsername(username); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserCreate,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("username validation failed: %v", err),
		})
		return fmt.Errorf("invalid username: %w", err)
	}

	// Validate shell path
	if err := validation.ValidateShell(shell); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserCreate,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("shell validation failed: %v", err),
		})
		return fmt.Errorf("invalid shell: %w", err)
	}

	// Check if user already exists
	if _, err := user.Lookup(username); err == nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserCreate,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   "user already exists",
		})
		return fmt.Errorf("user %s already exists", username)
	}

	// Create user with home directory
	// Use "--" to prevent username from being interpreted as a flag (prevents injection)
	cmd := exec.Command("useradd",
		"-m",        // Create home directory
		"-s", shell, // Set shell
		"-c", "StratusShell User", // Comment
		"--", // End of options
		username,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserCreate,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("useradd command failed: %v", err),
			Details: map[string]interface{}{
				"output": string(output),
			},
		})
		return fmt.Errorf("failed to create user: %w (output: %s)", err, output)
	}

	// Success - log audit entry
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionUserCreate,
		Actor:   "system",
		Target:  username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"shell": shell,
		},
	})

	return nil
}

func DeleteUser(username string) error {
	// Validate username
	if err := validation.ValidateUsername(username); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserDelete,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("validation failed: %v", err),
		})
		return fmt.Errorf("invalid username: %w", err)
	}

	// Check if user exists before attempting deletion
	if _, err := user.Lookup(username); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserDelete,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   "user does not exist",
		})
		return fmt.Errorf("user %s does not exist", username)
	}

	// Use "--" to prevent username from being interpreted as a flag
	cmd := exec.Command("userdel", "-r", "--", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserDelete,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("userdel command failed: %v", err),
			Details: map[string]interface{}{
				"output": string(output),
			},
		})
		return fmt.Errorf("failed to delete user: %w (output: %s)", err, output)
	}

	// Success - log audit entry
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionUserDelete,
		Actor:   "system",
		Target:  username,
		Outcome: audit.OutcomeSuccess,
	})

	return nil
}

func UserExists(username string) bool {
	_, err := user.Lookup(username)
	return err == nil
}

func SetUserShell(username, shell string) error {
	// Validate inputs
	if err := validation.ValidateUsername(username); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserShellChange,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("username validation failed: %v", err),
		})
		return fmt.Errorf("invalid username: %w", err)
	}

	if err := validation.ValidateShell(shell); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserShellChange,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("shell validation failed: %v", err),
		})
		return fmt.Errorf("invalid shell: %w", err)
	}

	// Use "--" to prevent arguments from being interpreted as flags
	cmd := exec.Command("chsh", "-s", shell, "--", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionUserShellChange,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("chsh command failed: %v", err),
			Details: map[string]interface{}{
				"output": string(output),
				"shell":  shell,
			},
		})
		return fmt.Errorf("failed to set user shell: %w (output: %s)", err, output)
	}

	// Success - log audit entry
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionUserShellChange,
		Actor:   "system",
		Target:  username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"shell": shell,
		},
	})

	return nil
}

func GetUserHomeDir(username string) (string, error) {
	// Validate username
	if err := validation.ValidateUsername(username); err != nil {
		return "", fmt.Errorf("invalid username: %w", err)
	}

	u, err := user.Lookup(username)
	if err != nil {
		return "", fmt.Errorf("user lookup failed: %w", err)
	}
	return u.HomeDir, nil
}

func ChownRecursive(path, username string) error {
	// Validate username
	if err := validation.ValidateUsername(username); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionChownRecursive,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("username validation failed: %v", err),
		})
		return fmt.Errorf("invalid username: %w", err)
	}

	// Validate path to prevent traversal attacks
	if err := validation.ValidateWorkingDir(path); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionChownRecursive,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("path validation failed: %v", err),
			Details: map[string]interface{}{
				"path": path,
			},
		})
		return fmt.Errorf("invalid path: %w", err)
	}

	// Verify path exists before attempting chown
	if _, err := os.Stat(path); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionChownRecursive,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("path does not exist: %v", err),
			Details: map[string]interface{}{
				"path": path,
			},
		})
		return fmt.Errorf("path does not exist: %w", err)
	}

	u, err := user.Lookup(username)
	if err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionChownRecursive,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("user lookup failed: %v", err),
		})
		return fmt.Errorf("user lookup failed: %w", err)
	}

	// Use "--" to prevent arguments from being interpreted as flags
	ownerSpec := fmt.Sprintf("%s:%s", u.Uid, u.Gid)
	cmd := exec.Command("chown", "-R", "--", ownerSpec, path)
	if output, err := cmd.CombinedOutput(); err != nil {
		auditLogger.Log(audit.Entry{
			Action:  audit.ActionChownRecursive,
			Actor:   "system",
			Target:  username,
			Outcome: audit.OutcomeFailure,
			Error:   fmt.Sprintf("chown command failed: %v", err),
			Details: map[string]interface{}{
				"output": string(output),
				"path":   path,
			},
		})
		return fmt.Errorf("failed to change ownership: %w (output: %s)", err, output)
	}

	// Success - log audit entry
	auditLogger.Log(audit.Entry{
		Action:  audit.ActionChownRecursive,
		Actor:   "system",
		Target:  username,
		Outcome: audit.OutcomeSuccess,
		Details: map[string]interface{}{
			"path": path,
		},
	})

	return nil
}
