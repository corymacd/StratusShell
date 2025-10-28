package validation

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Input validation patterns
var (
	// Terminal title: alphanumeric, spaces, dashes, underscores (1-100 chars)
	terminalTitleRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]{1,100}$`)

	// Session name: alphanumeric, spaces, dashes, underscores (1-100 chars)
	sessionNameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]{1,100}$`)

	// Port number: 1024-65535 (unprivileged ports)
	portMin = 1024
	portMax = 65535
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateTerminalTitle validates terminal title input
func ValidateTerminalTitle(title string) error {
	title = strings.TrimSpace(title)

	if title == "" {
		return &ValidationError{Field: "title", Message: "title cannot be empty"}
	}

	if !terminalTitleRegex.MatchString(title) {
		return &ValidationError{
			Field:   "title",
			Message: "title can only contain letters, numbers, spaces, dashes, and underscores (max 100 characters)",
		}
	}

	return nil
}

// ValidateSessionName validates session name input
func ValidateSessionName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return &ValidationError{Field: "name", Message: "session name cannot be empty"}
	}

	if !sessionNameRegex.MatchString(name) {
		return &ValidationError{
			Field:   "name",
			Message: "session name can only contain letters, numbers, spaces, dashes, and underscores (max 100 characters)",
		}
	}

	return nil
}

// ValidateSessionDescription validates session description
func ValidateSessionDescription(description string) error {
	description = strings.TrimSpace(description)

	if len(description) > 500 {
		return &ValidationError{
			Field:   "description",
			Message: "description cannot exceed 500 characters",
		}
	}

	return nil
}

// ValidatePort validates port number
func ValidatePort(port int) error {
	if port < portMin || port > portMax {
		return &ValidationError{
			Field:   "port",
			Message: fmt.Sprintf("port must be between %d and %d", portMin, portMax),
		}
	}

	return nil
}

// ValidateWorkingDir validates working directory path
func ValidateWorkingDir(dir string) error {
	if dir == "" {
		return nil // Empty is allowed - will use default
	}

	// Clean the path
	dir = filepath.Clean(dir)

	// Prevent directory traversal attacks
	if strings.Contains(dir, "..") {
		return &ValidationError{
			Field:   "working_dir",
			Message: "working directory cannot contain '..'",
		}
	}

	// Must be absolute path
	if !filepath.IsAbs(dir) {
		return &ValidationError{
			Field:   "working_dir",
			Message: "working directory must be an absolute path",
		}
	}

	return nil
}

// ValidateShell validates shell path
func ValidateShell(shell string) error {
	if shell == "" {
		return nil // Empty is allowed - will use default
	}

	// Clean the path
	shell = filepath.Clean(shell)

	// Prevent directory traversal attacks
	if strings.Contains(shell, "..") {
		return &ValidationError{
			Field:   "shell",
			Message: "shell path cannot contain '..'",
		}
	}

	// Must be absolute path
	if !filepath.IsAbs(shell) {
		return &ValidationError{
			Field:   "shell",
			Message: "shell must be an absolute path",
		}
	}

	// Common safe shells
	allowedShells := []string{
		"/bin/bash",
		"/bin/sh",
		"/bin/zsh",
		"/bin/fish",
		"/usr/bin/bash",
		"/usr/bin/sh",
		"/usr/bin/zsh",
		"/usr/bin/fish",
	}

	for _, allowed := range allowedShells {
		if shell == allowed {
			return nil
		}
	}

	return &ValidationError{
		Field:   "shell",
		Message: fmt.Sprintf("shell must be one of: %s", strings.Join(allowedShells, ", ")),
	}
}

// ValidateTerminalID validates terminal ID
func ValidateTerminalID(id int) error {
	if id < 0 {
		return &ValidationError{
			Field:   "id",
			Message: "terminal ID must be non-negative",
		}
	}

	return nil
}

// ValidateSessionID validates session ID
func ValidateSessionID(id int) error {
	if id < 1 {
		return &ValidationError{
			Field:   "id",
			Message: "session ID must be positive",
		}
	}

	return nil
}

// ValidateLayoutType validates layout type
func ValidateLayoutType(layoutType string) error {
	validLayouts := []string{"horizontal", "vertical", "grid"}

	for _, valid := range validLayouts {
		if layoutType == valid {
			return nil
		}
	}

	return &ValidationError{
		Field:   "layout_type",
		Message: fmt.Sprintf("layout type must be one of: %s", strings.Join(validLayouts, ", ")),
	}
}

// SanitizeString removes potentially dangerous characters from strings
func SanitizeString(s string) string {
	// Remove control characters except newline and tab
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)

	// Trim whitespace
	return strings.TrimSpace(s)
}
