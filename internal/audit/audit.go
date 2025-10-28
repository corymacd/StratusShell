package audit

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// ActionType represents the type of action being audited
type ActionType string

const (
	// Terminal actions
	ActionTerminalSpawn  ActionType = "terminal.spawn"
	ActionTerminalKill   ActionType = "terminal.kill"
	ActionTerminalRename ActionType = "terminal.rename"

	// Session actions
	ActionSessionCreate ActionType = "session.create"
	ActionSessionLoad   ActionType = "session.load"
	ActionSessionDelete ActionType = "session.delete"

	// Layout actions
	ActionLayoutChange ActionType = "layout.change"

	// Auth actions
	ActionAuthLogin  ActionType = "auth.login"
	ActionAuthLogout ActionType = "auth.logout"

	// Provisioning actions
	ActionUserCreate       ActionType = "provision.user.create"
	ActionUserDelete       ActionType = "provision.user.delete"
	ActionUserShellChange  ActionType = "provision.user.shell_change"
	ActionUserGroupAdd     ActionType = "provision.user.group_add"
	ActionSudoersConfig    ActionType = "provision.sudoers.config"
	ActionSudoersRemove    ActionType = "provision.sudoers.remove"
	ActionChownRecursive   ActionType = "provision.chown_recursive"
	ActionToolInstall      ActionType = "provision.tool.install"
)

// Outcome represents the result of an action
type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
)

// Entry represents a single audit log entry
type Entry struct {
	Timestamp time.Time              `json:"timestamp"`
	Action    ActionType             `json:"action"`
	Actor     string                 `json:"actor"`
	Target    string                 `json:"target,omitempty"`
	Outcome   Outcome                `json:"outcome"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Logger provides structured audit logging
type Logger struct {
	// In production, this would write to a dedicated audit log file or service
	// For now, we use the standard logger with structured JSON
}

// NewLogger creates a new audit logger
func NewLogger() *Logger {
	return &Logger{}
}

// Log writes an audit entry
func (l *Logger) Log(entry Entry) {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Serialize to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("AUDIT ERROR: Failed to serialize audit entry: %v", err)
		return
	}

	// Write to log
	log.Printf("AUDIT: %s", string(data))
}

// LogTerminalSpawn logs terminal creation
func (l *Logger) LogTerminalSpawn(actor string, terminalID int, title string, outcome Outcome, err error) {
	entry := Entry{
		Action:  ActionTerminalSpawn,
		Actor:   actor,
		Target:  fmt.Sprintf("terminal:%d", terminalID),
		Outcome: outcome,
		Details: map[string]interface{}{
			"title": title,
		},
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.Log(entry)
}

// LogTerminalKill logs terminal deletion
func (l *Logger) LogTerminalKill(actor string, terminalID int, outcome Outcome, err error) {
	entry := Entry{
		Action:  ActionTerminalKill,
		Actor:   actor,
		Target:  fmt.Sprintf("terminal:%d", terminalID),
		Outcome: outcome,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.Log(entry)
}

// LogTerminalRename logs terminal rename
func (l *Logger) LogTerminalRename(actor string, terminalID int, oldTitle, newTitle string, outcome Outcome, err error) {
	entry := Entry{
		Action:  ActionTerminalRename,
		Actor:   actor,
		Target:  fmt.Sprintf("terminal:%d", terminalID),
		Outcome: outcome,
		Details: map[string]interface{}{
			"old_title": oldTitle,
			"new_title": newTitle,
		},
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.Log(entry)
}

// LogSessionCreate logs session creation
func (l *Logger) LogSessionCreate(actor string, sessionID int, name string, outcome Outcome, err error) {
	entry := Entry{
		Action:  ActionSessionCreate,
		Actor:   actor,
		Target:  fmt.Sprintf("session:%d", sessionID),
		Outcome: outcome,
		Details: map[string]interface{}{
			"name": name,
		},
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.Log(entry)
}

// LogSessionLoad logs session loading
func (l *Logger) LogSessionLoad(actor string, sessionID int, outcome Outcome, err error) {
	entry := Entry{
		Action:  ActionSessionLoad,
		Actor:   actor,
		Target:  fmt.Sprintf("session:%d", sessionID),
		Outcome: outcome,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.Log(entry)
}

// LogLayoutChange logs layout changes
func (l *Logger) LogLayoutChange(actor string, layoutType string, outcome Outcome, err error) {
	entry := Entry{
		Action:  ActionLayoutChange,
		Actor:   actor,
		Target:  fmt.Sprintf("layout:%s", layoutType),
		Outcome: outcome,
		Details: map[string]interface{}{
			"layout_type": layoutType,
		},
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.Log(entry)
}

// LogAuthLogin logs authentication login
func (l *Logger) LogAuthLogin(actor string, outcome Outcome, err error) {
	entry := Entry{
		Action:  ActionAuthLogin,
		Actor:   actor,
		Outcome: outcome,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.Log(entry)
}

// LogAuthLogout logs authentication logout
func (l *Logger) LogAuthLogout(actor string, outcome Outcome) {
	entry := Entry{
		Action:  ActionAuthLogout,
		Actor:   actor,
		Outcome: outcome,
	}

	l.Log(entry)
}

// OutcomeFromError returns OutcomeSuccess if err is nil, otherwise OutcomeFailure
func OutcomeFromError(err error) Outcome {
	if err == nil {
		return OutcomeSuccess
	}
	return OutcomeFailure
}

// ErrorString returns the error message if err is not nil, otherwise empty string
func ErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
