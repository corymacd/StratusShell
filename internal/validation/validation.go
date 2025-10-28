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

	// Username: lowercase letters, digits, dashes, underscores (1-32 chars)
	// Must start with lowercase letter
	usernameRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

	// Group name: lowercase letters, digits, dashes, underscores (1-32 chars)
	// Can start with lowercase letter or underscore
	groupnameRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

	// Reserved usernames that cannot be used
	reservedUsernames = map[string]bool{
		"root": true, "daemon": true, "bin": true, "sys": true, "sync": true, 
		"games": true, "man": true, "lp": true, "mail": true, "news": true, 
		"uucp": true, "proxy": true, "www-data": true, "backup": true, "list": true,
		"irc": true, "gnats": true, "nobody": true, "systemd-network": true, 
		"systemd-resolve": true, "systemd-timesync": true, "messagebus": true, 
		"systemd-coredump": true, "syslog": true,
	}

	// Reserved group names that cannot be used
	reservedGroupnames = map[string]bool{
		"root": true, "daemon": true, "bin": true, "sys": true, "adm": true, 
		"tty": true, "disk": true, "lp": true, "mail": true, "news": true, 
		"uucp": true, "man": true, "proxy": true, "kmem": true, "dialout": true,
		"fax": true, "voice": true, "cdrom": true, "floppy": true, "tape": true, 
		"sudo": true, "audio": true, "dip": true, "www-data": true, "backup": true, 
		"operator": true, "list": true, "irc": true, "src": true, "gnats": true, 
		"shadow": true, "utmp": true, "video": true, "sasl": true, "plugdev": true,
		"staff": true, "games": true, "users": true, "nogroup": true, 
		"systemd-journal": true, "systemd-network": true, "systemd-resolve": true, 
		"systemd-timesync": true, "messagebus": true, "systemd-coredump": true, 
		"syslog": true, "ssh": true,
	}

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

// ValidateUsername validates username for system user creation
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return &ValidationError{Field: "username", Message: "username cannot be empty"}
	}

	// Check against regex pattern
	if !usernameRegex.MatchString(username) {
		return &ValidationError{
			Field:   "username",
			Message: "username must start with lowercase letter and contain only lowercase letters, numbers, dashes, and underscores (max 32 characters)",
		}
	}

	// Check if username is reserved
	if reservedUsernames[username] {
		return &ValidationError{
			Field:   "username",
			Message: fmt.Sprintf("username '%s' is reserved and cannot be used", username),
		}
	}

	return nil
}

// ValidateGroupname validates group name for system group operations
func ValidateGroupname(groupname string) error {
	groupname = strings.TrimSpace(groupname)

	if groupname == "" {
		return &ValidationError{Field: "groupname", Message: "groupname cannot be empty"}
	}

	// Check against regex pattern
	if !groupnameRegex.MatchString(groupname) {
		return &ValidationError{
			Field:   "groupname",
			Message: "groupname must start with lowercase letter or underscore and contain only lowercase letters, numbers, dashes, and underscores (max 32 characters)",
		}
	}

	// Check if groupname is reserved
	if reservedGroupnames[groupname] {
		return &ValidationError{
			Field:   "groupname",
			Message: fmt.Sprintf("groupname '%s' is reserved and cannot be used", groupname),
		}
	}

	return nil
}
