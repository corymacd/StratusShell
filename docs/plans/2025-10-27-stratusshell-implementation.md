# StratusShell Enhancement Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform StratusShell from a simple dual-terminal web app into a comprehensive cloud development environment provisioning tool with Cobra CLI, Go-native provisioning, HTMX + Templ web UI, SQLite persistence, and systemd integration.

**Architecture:** Feature-based package structure with clear separation: provision/ for user/tool setup, server/ for web UI and GoTTY management, db/ for SQLite persistence, ui/ for Templ components, service/ for systemd. Cobra CLI provides three commands: init (provision), serve (web UI), install (systemd).

**Tech Stack:** Go 1.24, Cobra (CLI), Templ (type-safe templates), HTMX 1.9 (dynamic UI), SQLite3 (persistence), GoTTY (terminal streaming), systemd (service management)

---

## Phase 1: Foundation & CLI Structure

### Task 1: Add Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add required dependencies**

```bash
cd ~/.config/superpowers/worktrees/StratusShell/feature-enhancements
go get github.com/spf13/cobra@v1.8.0
go get github.com/a-h/templ@v0.2.543
go get github.com/mattn/go-sqlite3@v1.14.19
go get gopkg.in/yaml.v3@v3.0.1
```

Expected: Dependencies added to go.mod

**Step 2: Download dependencies**

```bash
go mod download
go mod tidy
```

Expected: All dependencies downloaded

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "build: add cobra, templ, sqlite3, yaml dependencies"
```

---

### Task 2: Create Package Structure

**Files:**
- Create: `cmd/root.go`
- Create: `cmd/init.go`
- Create: `cmd/serve.go`
- Create: `cmd/install.go`
- Create: `internal/provision/.gitkeep`
- Create: `internal/server/.gitkeep`
- Create: `internal/db/.gitkeep`
- Create: `internal/ui/.gitkeep`
- Create: `internal/service/.gitkeep`
- Create: `configs/default.yaml`

**Step 1: Create directory structure**

```bash
mkdir -p cmd internal/{provision,server,db,ui,service} configs
touch internal/provision/.gitkeep internal/server/.gitkeep internal/db/.gitkeep internal/ui/.gitkeep internal/service/.gitkeep
```

**Step 2: Create root command**

File: `cmd/root.go`

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stratusshell",
	Short: "Cloud development environment provisioning tool",
	Long: `StratusShell provisions complete cloud development environments with
user creation, tool installation, and web-based terminal management.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Subcommands will be added here
}
```

**Step 3: Create stub commands**

File: `cmd/init.go`

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Provision a development environment",
	Long:  `Create a system user and install complete development toolchain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("user", "u", "", "Username to create (required)")
	initCmd.MarkFlagRequired("user")
	initCmd.Flags().String("shell", "/bin/bash", "Default shell")
	initCmd.Flags().String("config", "/etc/stratusshell/default.yaml", "Config file path")
	initCmd.Flags().Bool("skip-tools", false, "Create user only, skip tool installation")
}
```

File: `cmd/serve.go`

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web UI server",
	Long:  `Start HTTP server with GoTTY terminal management and HTMX UI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("user", "u", "", "Run as specific user (when started by root)")
	serveCmd.Flags().IntP("port", "p", 8080, "HTTP port")
	serveCmd.Flags().String("db", "", "Database path (default: ~/.stratusshell/data.db)")
}
```

File: `cmd/install.go`

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install systemd service",
	Long:  `Generate and enable systemd service for StratusShell.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("user", "u", "", "User that service runs as (required)")
	installCmd.MarkFlagRequired("user")
	installCmd.Flags().IntP("port", "p", 8080, "HTTP port")
}
```

**Step 4: Create default config**

File: `configs/default.yaml`

```yaml
user:
  shell: /bin/zsh

base_packages:
  - git
  - curl
  - build-essential

cloud:
  aws: true
  gcloud: true
  kubectl: true
  docker: true
  terraform: true

languages:
  go:
    enabled: true
    version: latest
    tools:
      - golangci-lint
      - gopls
      - delve

  node:
    enabled: true
    version: lts
    package_manager: pnpm
    global_packages:
      - typescript
      - prettier
      - eslint

shell:
  zsh: true
  oh_my_zsh: true
  tmux: true
```

**Step 5: Update main.go to use Cobra**

File: `main.go`

```go
package main

import "github.com/corymacd/stratusshell/cmd"

func main() {
	cmd.Execute()
}
```

**Step 6: Test CLI structure**

```bash
go build -o stratusshell main.go
./stratusshell --help
```

Expected: Help text showing three commands (init, serve, install)

```bash
./stratusshell init --help
./stratusshell serve --help
./stratusshell install --help
```

Expected: Each command shows its flags

**Step 7: Commit**

```bash
git add cmd/ internal/ configs/ main.go
git commit -m "feat: add cobra CLI structure with init/serve/install commands"
```

---

## Phase 2: Database Layer

### Task 3: SQLite Schema and Migrations

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/schema.sql`

**Step 1: Create schema file**

File: `internal/db/schema.sql`

```sql
-- User preferences
CREATE TABLE IF NOT EXISTS preferences (
    id INTEGER PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Saved sessions
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Terminal configurations within a session
CREATE TABLE IF NOT EXISTS session_terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    terminal_index INTEGER NOT NULL,
    title TEXT NOT NULL,
    shell TEXT DEFAULT '/bin/bash',
    working_dir TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Current active layout (singleton)
CREATE TABLE IF NOT EXISTS active_layout (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    layout_type TEXT NOT NULL CHECK (layout_type IN ('horizontal', 'vertical', 'grid')),
    terminal_count INTEGER NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Active terminals (current running state)
CREATE TABLE IF NOT EXISTS active_terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    port INTEGER UNIQUE NOT NULL,
    title TEXT NOT NULL,
    pid INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Step 2: Create database package**

File: `internal/db/db.go`

```go
package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	conn *sql.DB
	path string
}

// Open opens or creates the SQLite database
func Open(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open database
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn, path: dbPath}

	// Run migrations
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Initialize singleton active_layout if not exists
	if err := db.initializeActiveLayout(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize active layout: %w", err)
	}

	return db, nil
}

func (db *DB) migrate() error {
	_, err := db.conn.Exec(schemaSQL)
	return err
}

func (db *DB) initializeActiveLayout() error {
	// Insert default layout if table is empty
	_, err := db.conn.Exec(`
		INSERT OR IGNORE INTO active_layout (id, layout_type, terminal_count)
		VALUES (1, 'horizontal', 2)
	`)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetConn() *sql.DB {
	return db.conn
}
```

**Step 3: Test database creation**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 4: Commit**

```bash
git add internal/db/
git commit -m "feat(db): add sqlite schema and database initialization"
```

---

### Task 4: Database CRUD Operations

**Files:**
- Create: `internal/db/preferences.go`
- Create: `internal/db/sessions.go`
- Create: `internal/db/terminals.go`

**Step 1: Preferences CRUD**

File: `internal/db/preferences.go`

```go
package db

import (
	"database/sql"
	"fmt"
)

func (db *DB) GetPreference(key string) (string, error) {
	var value string
	err := db.conn.QueryRow("SELECT value FROM preferences WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (db *DB) SetPreference(key, value string) error {
	_, err := db.conn.Exec(`
		INSERT INTO preferences (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP
	`, key, value, value)
	return err
}

func (db *DB) GetAllPreferences() (map[string]string, error) {
	rows, err := db.conn.Query("SELECT key, value FROM preferences")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		prefs[key] = value
	}
	return prefs, rows.Err()
}
```

**Step 2: Sessions CRUD**

File: `internal/db/sessions.go`

```go
package db

import (
	"time"
)

type Session struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SessionTerminal struct {
	ID            int
	SessionID     int
	TerminalIndex int
	Title         string
	Shell         string
	WorkingDir    string
}

func (db *DB) CreateSession(name, description string) (int, error) {
	result, err := db.conn.Exec(`
		INSERT INTO sessions (name, description) VALUES (?, ?)
	`, name, description)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (db *DB) GetSession(id int) (*Session, error) {
	s := &Session{}
	err := db.conn.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM sessions WHERE id = ?
	`, id).Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (db *DB) GetAllSessions() ([]*Session, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, created_at, updated_at
		FROM sessions ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s := &Session{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (db *DB) SaveSessionTerminal(sessionID, index int, title, shell, workingDir string) error {
	_, err := db.conn.Exec(`
		INSERT INTO session_terminals (session_id, terminal_index, title, shell, working_dir)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, index, title, shell, workingDir)
	return err
}

func (db *DB) GetSessionTerminals(sessionID int) ([]*SessionTerminal, error) {
	rows, err := db.conn.Query(`
		SELECT id, session_id, terminal_index, title, shell, working_dir
		FROM session_terminals WHERE session_id = ? ORDER BY terminal_index
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terminals []*SessionTerminal
	for rows.Next() {
		t := &SessionTerminal{}
		if err := rows.Scan(&t.ID, &t.SessionID, &t.TerminalIndex, &t.Title, &t.Shell, &t.WorkingDir); err != nil {
			return nil, err
		}
		terminals = append(terminals, t)
	}
	return terminals, rows.Err()
}
```

**Step 3: Terminals CRUD**

File: `internal/db/terminals.go`

```go
package db

import (
	"time"
)

type ActiveTerminal struct {
	ID        int
	Port      int
	Title     string
	PID       int
	CreatedAt time.Time
}

type ActiveLayout struct {
	LayoutType     string
	TerminalCount  int
}

func (db *DB) SaveActiveTerminal(port int, title string, pid int) error {
	_, err := db.conn.Exec(`
		INSERT INTO active_terminals (port, title, pid) VALUES (?, ?, ?)
	`, port, title, pid)
	return err
}

func (db *DB) GetActiveTerminals() ([]*ActiveTerminal, error) {
	rows, err := db.conn.Query(`
		SELECT id, port, title, pid, created_at
		FROM active_terminals ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terminals []*ActiveTerminal
	for rows.Next() {
		t := &ActiveTerminal{}
		if err := rows.Scan(&t.ID, &t.Port, &t.Title, &t.PID, &t.CreatedAt); err != nil {
			return nil, err
		}
		terminals = append(terminals, t)
	}
	return terminals, rows.Err()
}

func (db *DB) DeleteActiveTerminal(id int) error {
	_, err := db.conn.Exec("DELETE FROM active_terminals WHERE id = ?", id)
	return err
}

func (db *DB) ClearActiveTerminals() error {
	_, err := db.conn.Exec("DELETE FROM active_terminals")
	return err
}

func (db *DB) GetActiveLayout() (*ActiveLayout, error) {
	layout := &ActiveLayout{}
	err := db.conn.QueryRow(`
		SELECT layout_type, terminal_count FROM active_layout WHERE id = 1
	`).Scan(&layout.LayoutType, &layout.TerminalCount)
	if err != nil {
		return nil, err
	}
	return layout, nil
}

func (db *DB) UpdateActiveLayout(layoutType string, terminalCount int) error {
	_, err := db.conn.Exec(`
		UPDATE active_layout SET layout_type = ?, terminal_count = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, layoutType, terminalCount)
	return err
}
```

**Step 4: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 5: Commit**

```bash
git add internal/db/
git commit -m "feat(db): add CRUD operations for preferences, sessions, terminals"
```

---

## Phase 3: GoTTY Terminal Management

### Task 5: Port Pool Manager

**Files:**
- Create: `internal/server/portpool.go`
- Create: `internal/server/portpool_test.go`

**Step 1: Write failing test**

File: `internal/server/portpool_test.go`

```go
package server

import (
	"testing"
)

func TestPortPoolAllocation(t *testing.T) {
	pool := NewPortPool(8081, 8085)

	// Allocate all ports
	ports := make([]int, 0)
	for i := 0; i < 5; i++ {
		port, err := pool.Allocate()
		if err != nil {
			t.Fatalf("failed to allocate port %d: %v", i, err)
		}
		ports = append(ports, port)
	}

	// Should fail when exhausted
	_, err := pool.Allocate()
	if err == nil {
		t.Fatal("expected error when pool exhausted, got nil")
	}

	// Release and reallocate
	pool.Release(ports[0])
	port, err := pool.Allocate()
	if err != nil {
		t.Fatalf("failed to reallocate: %v", err)
	}
	if port != ports[0] {
		t.Fatalf("expected port %d, got %d", ports[0], port)
	}
}

func TestPortPoolConcurrency(t *testing.T) {
	pool := NewPortPool(9000, 9010)

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			port, err := pool.Allocate()
			if err != nil {
				t.Errorf("concurrent allocation failed: %v", err)
			}
			pool.Release(port)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/server/
```

Expected: FAIL - NewPortPool undefined

**Step 3: Implement port pool**

File: `internal/server/portpool.go`

```go
package server

import (
	"errors"
	"sync"
)

type PortPool struct {
	minPort int
	maxPort int
	used    map[int]bool
	mu      sync.Mutex
}

func NewPortPool(minPort, maxPort int) *PortPool {
	return &PortPool{
		minPort: minPort,
		maxPort: maxPort,
		used:    make(map[int]bool),
	}
}

func (p *PortPool) Allocate() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for port := p.minPort; port <= p.maxPort; port++ {
		if !p.used[port] {
			p.used[port] = true
			return port, nil
		}
	}
	return 0, errors.New("no available ports in pool")
}

func (p *PortPool) Release(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.used, port)
}

func (p *PortPool) IsUsed(port int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.used[port]
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/server/ -v
```

Expected: PASS (all tests)

**Step 5: Commit**

```bash
git add internal/server/
git commit -m "feat(server): add port pool for dynamic terminal allocation"
```

---

### Task 6: Terminal Manager

**Files:**
- Create: `internal/server/terminal.go`

**Step 1: Create terminal manager**

File: `internal/server/terminal.go`

```go
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/corymacd/stratusshell/internal/db"
)

type Terminal struct {
	ID         int
	Port       int
	Title      string
	PID        int
	Cmd        *exec.Cmd
	CancelFunc context.CancelFunc
	CreatedAt  time.Time
}

type TerminalManager struct {
	terminals map[int]*Terminal
	portPool  *PortPool
	db        *db.DB
	mu        sync.RWMutex
	nextID    int
}

func NewTerminalManager(db *db.DB) *TerminalManager {
	return &TerminalManager{
		terminals: make(map[int]*Terminal),
		portPool:  NewPortPool(8081, 8181),
		db:        db,
		nextID:    1,
	}
}

func (tm *TerminalManager) SpawnTerminal(title, shell, workingDir string) (*Terminal, error) {
	port, err := tm.portPool.Allocate()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Build GoTTY command
	cmd := exec.CommandContext(ctx, "gotty",
		"--port", strconv.Itoa(port),
		"--address", "localhost",
		"--permit-write",
		"--reconnect",
		"--reconnect-time", "10",
		"--title-format", title,
		shell,
	)

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	if err := cmd.Start(); err != nil {
		cancel()
		tm.portPool.Release(port)
		return nil, fmt.Errorf("failed to start gotty: %w", err)
	}

	tm.mu.Lock()
	terminalID := tm.nextID
	tm.nextID++
	tm.mu.Unlock()

	terminal := &Terminal{
		ID:         terminalID,
		Port:       port,
		Title:      title,
		PID:        cmd.Process.Pid,
		Cmd:        cmd,
		CancelFunc: cancel,
		CreatedAt:  time.Now(),
	}

	tm.mu.Lock()
	tm.terminals[terminal.ID] = terminal
	tm.mu.Unlock()

	// Save to database
	if err := tm.db.SaveActiveTerminal(terminal.Port, terminal.Title, terminal.PID); err != nil {
		log.Printf("Warning: failed to save terminal to db: %v", err)
	}

	// Monitor process
	go tm.monitorTerminal(terminal)

	return terminal, nil
}

func (tm *TerminalManager) KillTerminal(id int) error {
	tm.mu.Lock()
	terminal, exists := tm.terminals[id]
	if !exists {
		tm.mu.Unlock()
		return errors.New("terminal not found")
	}
	delete(tm.terminals, id)
	tm.mu.Unlock()

	// Cancel context (kills GoTTY)
	terminal.CancelFunc()

	// Wait for process to exit
	terminal.Cmd.Wait()

	// Release port
	tm.portPool.Release(terminal.Port)

	// Remove from database
	if err := tm.db.DeleteActiveTerminal(id); err != nil {
		log.Printf("Warning: failed to delete terminal from db: %v", err)
	}

	return nil
}

func (tm *TerminalManager) monitorTerminal(terminal *Terminal) {
	terminal.Cmd.Wait()

	// If process died unexpectedly, clean up
	tm.mu.Lock()
	if _, exists := tm.terminals[terminal.ID]; exists {
		delete(tm.terminals, terminal.ID)
		tm.portPool.Release(terminal.Port)
		tm.db.DeleteActiveTerminal(terminal.ID)
		log.Printf("Terminal %d (port %d) died unexpectedly", terminal.ID, terminal.Port)
	}
	tm.mu.Unlock()
}

func (tm *TerminalManager) GetTerminals() []*Terminal {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	terminals := make([]*Terminal, 0, len(tm.terminals))
	for _, t := range tm.terminals {
		terminals = append(terminals, t)
	}
	return terminals
}

func (tm *TerminalManager) GetTerminal(id int) (*Terminal, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	t, ok := tm.terminals[id]
	return t, ok
}

func (tm *TerminalManager) Shutdown() error {
	tm.mu.Lock()
	terminals := make([]*Terminal, 0, len(tm.terminals))
	for _, t := range tm.terminals {
		terminals = append(terminals, t)
	}
	tm.mu.Unlock()

	// Kill all terminals
	for _, terminal := range terminals {
		if err := tm.KillTerminal(terminal.ID); err != nil {
			log.Printf("Error killing terminal %d: %v", terminal.ID, err)
		}
	}

	return nil
}

func (tm *TerminalManager) ApplyLayout(layoutType string) error {
	targetCount := tm.getTerminalCountForLayout(layoutType)
	currentCount := len(tm.terminals)

	if targetCount > currentCount {
		// Spawn additional terminals
		for i := currentCount; i < targetCount; i++ {
			_, err := tm.SpawnTerminal(
				fmt.Sprintf("Terminal %d", i+1),
				"/bin/bash",
				"",
			)
			if err != nil {
				return fmt.Errorf("failed to spawn terminal: %w", err)
			}
		}
	} else if targetCount < currentCount {
		// Kill excess terminals
		terminals := tm.GetTerminals()
		for i := targetCount; i < len(terminals); i++ {
			if err := tm.KillTerminal(terminals[i].ID); err != nil {
				log.Printf("Error killing excess terminal: %v", err)
			}
		}
	}

	// Update layout in DB
	if err := tm.db.UpdateActiveLayout(layoutType, targetCount); err != nil {
		return fmt.Errorf("failed to update layout in db: %w", err)
	}

	return nil
}

func (tm *TerminalManager) getTerminalCountForLayout(layoutType string) int {
	switch layoutType {
	case "horizontal":
		return 2
	case "vertical":
		return 2
	case "grid":
		return 4
	default:
		return 2
	}
}
```

**Step 2: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/server/terminal.go
git commit -m "feat(server): add terminal manager with lifecycle management"
```

---

## Phase 4: Templ UI Components

### Task 7: Install Templ CLI

**Step 1: Install templ CLI tool**

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

Expected: templ command available

**Step 2: Verify installation**

```bash
templ version
```

Expected: Shows version (e.g., v0.2.543)

---

### Task 8: Create Base UI Components

**Files:**
- Create: `internal/ui/layout.templ`
- Create: `internal/ui/menubar.templ`
- Create: `internal/ui/terminal.templ`
- Create: `internal/ui/modals.templ`
- Create: `static/styles.css`

**Step 1: Create layout component**

File: `internal/ui/layout.templ`

```templ
package ui

templ Layout(user string) {
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
		<title>StratusShell - {user}</title>
		<script src="https://unpkg.com/htmx.org@1.9.10"></script>
		<link rel="stylesheet" href="/static/styles.css"/>
	</head>
	<body>
		@Menubar()
		<div id="terminal-container" hx-get="/api/layout" hx-trigger="load">
			<!-- Terminals loaded here -->
		</div>
		<div id="modal"></div>
	</body>
	</html>
}
```

**Step 2: Create menubar component**

File: `internal/ui/menubar.templ`

```templ
package ui

templ Menubar() {
	<nav class="menubar">
		<div class="logo">StratusShell</div>

		<div class="dropdown">
			<button class="dropdown-btn">Layout ▾</button>
			<div class="dropdown-menu">
				<a hx-post="/api/layout/horizontal" hx-target="#terminal-container">Horizontal Split</a>
				<a hx-post="/api/layout/vertical" hx-target="#terminal-container">Vertical Split</a>
				<a hx-post="/api/layout/grid" hx-target="#terminal-container">Grid (2x2)</a>
			</div>
		</div>

		<div class="dropdown">
			<button class="dropdown-btn">Config ▾</button>
			<div class="dropdown-menu">
				<a hx-get="/api/config/modal" hx-target="#modal">Preferences...</a>
				<a hx-post="/api/terminals/add" hx-target="#terminal-container">Add Terminal</a>
			</div>
		</div>

		<div class="dropdown">
			<button class="dropdown-btn">Sessions ▾</button>
			<div class="dropdown-menu">
				<a hx-get="/api/session/save-modal" hx-target="#modal">Save Session...</a>
				<a hx-get="/api/session/list-modal" hx-target="#modal">Load Session...</a>
			</div>
		</div>
	</nav>
}
```

**Step 3: Create terminal component**

File: `internal/ui/terminal.templ`

```templ
package ui

import "fmt"

templ TerminalPane(id int, port int, title string) {
	<div class="terminal-pane" id={ fmt.Sprintf("terminal-%d", id) }>
		<div class="terminal-header">
			<span class="terminal-title" contenteditable="true"
				hx-post={ fmt.Sprintf("/api/terminal/%d/rename", id) }
				hx-trigger="blur"
				hx-include="this"
				hx-swap="none">
				{ title }
			</span>
			<button class="terminal-close"
				hx-delete={ fmt.Sprintf("/api/terminal/%d", id) }
				hx-target={ fmt.Sprintf("#terminal-%d", id) }
				hx-swap="outerHTML">×</button>
		</div>
		<iframe src={ fmt.Sprintf("http://localhost:%d", port) } class="terminal-frame"></iframe>
	</div>
}

templ TerminalContainer(terminals []TerminalData, layoutType string) {
	<div class={ "terminals", "layout-" + layoutType }>
		for _, t := range terminals {
			@TerminalPane(t.ID, t.Port, t.Title)
		}
	</div>
}

type TerminalData struct {
	ID    int
	Port  int
	Title string
}
```

**Step 4: Create modals component**

File: `internal/ui/modals.templ`

```templ
package ui

templ SaveSessionModal() {
	<div class="modal-overlay" hx-on:click="document.getElementById('modal').innerHTML = ''">
		<div class="modal-content" hx-on:click="event.stopPropagation()">
			<h2>Save Session</h2>
			<form hx-post="/api/session/save" hx-target="#modal">
				<label>
					Session Name:
					<input type="text" name="name" required autofocus/>
				</label>
				<label>
					Description (optional):
					<textarea name="description"></textarea>
				</label>
				<div class="modal-actions">
					<button type="submit">Save</button>
					<button type="button" hx-on:click="document.getElementById('modal').innerHTML = ''">Cancel</button>
				</div>
			</form>
		</div>
	</div>
}

templ LoadSessionModal(sessions []SessionData) {
	<div class="modal-overlay" hx-on:click="document.getElementById('modal').innerHTML = ''">
		<div class="modal-content" hx-on:click="event.stopPropagation()">
			<h2>Load Session</h2>
			<div class="session-list">
				if len(sessions) == 0 {
					<p>No saved sessions</p>
				} else {
					for _, s := range sessions {
						<div class="session-item">
							<div class="session-info">
								<strong>{ s.Name }</strong>
								if s.Description != "" {
									<p>{ s.Description }</p>
								}
							</div>
							<button hx-post={ fmt.Sprintf("/api/session/load/%d", s.ID) }
								hx-target="#terminal-container">Load</button>
						</div>
					}
				}
			</div>
			<div class="modal-actions">
				<button type="button" hx-on:click="document.getElementById('modal').innerHTML = ''">Cancel</button>
			</div>
		</div>
	</div>
}

type SessionData struct {
	ID          int
	Name        string
	Description string
}

templ SuccessMessage(message string) {
	<div class="modal-overlay" hx-on:click="document.getElementById('modal').innerHTML = ''">
		<div class="modal-content modal-success" hx-on:click="event.stopPropagation()">
			<h2>✓ Success</h2>
			<p>{ message }</p>
			<button hx-on:click="document.getElementById('modal').innerHTML = ''">Close</button>
		</div>
	</div>
}

templ ErrorToast(message string) {
	<div class="error-toast">{ message }</div>
}
```

**Step 5: Create CSS**

File: `static/styles.css`

```css
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    background-color: #1e1e1e;
    color: #ffffff;
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
}

/* Menubar */
.menubar {
    background-color: #252526;
    padding: 10px 20px;
    border-bottom: 1px solid #3e3e42;
    display: flex;
    gap: 15px;
    align-items: center;
}

.logo {
    font-weight: 600;
    font-size: 16px;
    margin-right: 20px;
}

.dropdown {
    position: relative;
}

.dropdown-btn {
    background: none;
    border: none;
    color: #cccccc;
    padding: 6px 12px;
    cursor: pointer;
    font-size: 14px;
}

.dropdown-btn:hover {
    background-color: #2a2d2e;
}

.dropdown-menu {
    display: none;
    position: absolute;
    top: 100%;
    left: 0;
    background-color: #252526;
    border: 1px solid #3e3e42;
    min-width: 200px;
    z-index: 1000;
}

.dropdown:hover .dropdown-menu {
    display: block;
}

.dropdown-menu a {
    display: block;
    padding: 10px 15px;
    color: #cccccc;
    text-decoration: none;
    cursor: pointer;
}

.dropdown-menu a:hover {
    background-color: #2a2d2e;
}

/* Terminal Container */
#terminal-container {
    flex: 1;
    overflow: hidden;
}

.terminals {
    display: flex;
    height: 100%;
    gap: 2px;
    background-color: #3e3e42;
}

.layout-horizontal {
    flex-direction: row;
}

.layout-vertical {
    flex-direction: column;
}

.layout-grid {
    flex-wrap: wrap;
}

.layout-grid .terminal-pane {
    width: calc(50% - 1px);
    height: calc(50% - 1px);
}

/* Terminal Pane */
.terminal-pane {
    flex: 1;
    display: flex;
    flex-direction: column;
    background-color: #1e1e1e;
    overflow: hidden;
}

.terminal-header {
    background-color: #2d2d30;
    padding: 8px 15px;
    border-bottom: 1px solid #3e3e42;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.terminal-title {
    font-size: 14px;
    color: #cccccc;
    outline: none;
}

.terminal-title:focus {
    background-color: #3e3e42;
    padding: 2px 6px;
}

.terminal-close {
    background: none;
    border: none;
    color: #cccccc;
    font-size: 20px;
    cursor: pointer;
    padding: 0 8px;
}

.terminal-close:hover {
    color: #ff0000;
}

.terminal-frame {
    flex: 1;
    border: none;
    width: 100%;
    background-color: #1e1e1e;
}

/* Modals */
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
}

.modal-content {
    background-color: #2d2d30;
    padding: 30px;
    border-radius: 8px;
    min-width: 400px;
    max-width: 600px;
    max-height: 80vh;
    overflow-y: auto;
}

.modal-content h2 {
    margin-bottom: 20px;
}

.modal-content form label {
    display: block;
    margin-bottom: 15px;
}

.modal-content input,
.modal-content textarea {
    width: 100%;
    padding: 8px;
    margin-top: 5px;
    background-color: #1e1e1e;
    border: 1px solid #3e3e42;
    color: #cccccc;
    font-family: inherit;
}

.modal-content textarea {
    min-height: 80px;
    resize: vertical;
}

.modal-actions {
    display: flex;
    gap: 10px;
    margin-top: 20px;
    justify-content: flex-end;
}

.modal-actions button {
    padding: 8px 20px;
    background-color: #007acc;
    color: white;
    border: none;
    cursor: pointer;
    border-radius: 4px;
}

.modal-actions button[type="button"] {
    background-color: #3e3e42;
}

.modal-actions button:hover {
    opacity: 0.9;
}

/* Session List */
.session-list {
    max-height: 400px;
    overflow-y: auto;
}

.session-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 15px;
    border: 1px solid #3e3e42;
    margin-bottom: 10px;
    border-radius: 4px;
}

.session-item:hover {
    background-color: #252526;
}

.session-info strong {
    display: block;
    margin-bottom: 5px;
}

.session-info p {
    font-size: 13px;
    color: #999;
}

.modal-success {
    text-align: center;
}

.modal-success h2 {
    color: #4ec9b0;
    font-size: 32px;
}

/* Error Toast */
.error-toast {
    position: fixed;
    bottom: 20px;
    right: 20px;
    background-color: #f44336;
    color: white;
    padding: 15px 25px;
    border-radius: 4px;
    z-index: 3000;
}
```

**Step 6: Generate templ files**

```bash
cd ~/.config/superpowers/worktrees/StratusShell/feature-enhancements
mkdir -p static
templ generate
```

Expected: Creates *_templ.go files for each .templ file

**Step 7: Commit**

```bash
git add internal/ui/ static/
git commit -m "feat(ui): add templ components for layout, menubar, terminals, modals"
```

---

## Phase 5: HTTP Server & Handlers

### Task 9: Create HTTP Server

**Files:**
- Create: `internal/server/server.go`
- Create: `internal/server/handlers.go`

**Step 1: Create server struct**

File: `internal/server/server.go`

```go
package server

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/corymacd/stratusshell/internal/db"
	"github.com/corymacd/stratusshell/internal/ui"
)

//go:embed ../../../static
var staticFiles embed.FS

type Server struct {
	port            int
	db              *db.DB
	terminalManager *TerminalManager
	httpServer      *http.Server
}

func NewServer(port int, dbPath string) (*Server, error) {
	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create terminal manager
	tm := NewTerminalManager(database)

	s := &Server{
		port:            port,
		db:              database,
		terminalManager: tm,
	}

	// Setup HTTP routes
	mux := http.NewServeMux()
	s.setupRoutes(mux)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s, nil
}

func (s *Server) setupRoutes(mux *http.ServeMux) {
	// Static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFiles)))

	// Main page
	mux.HandleFunc("/", s.handleIndex)

	// API routes
	mux.HandleFunc("/api/layout", s.handleGetLayout)
	mux.HandleFunc("/api/layout/horizontal", s.handleLayoutHorizontal)
	mux.HandleFunc("/api/layout/vertical", s.handleLayoutVertical)
	mux.HandleFunc("/api/layout/grid", s.handleLayoutGrid)

	mux.HandleFunc("/api/terminals/add", s.handleAddTerminal)
	mux.HandleFunc("/api/terminal/", s.handleTerminalAction)

	mux.HandleFunc("/api/session/save-modal", s.handleSaveSessionModal)
	mux.HandleFunc("/api/session/save", s.handleSaveSession)
	mux.HandleFunc("/api/session/list-modal", s.handleListSessionsModal)
	mux.HandleFunc("/api/session/load/", s.handleLoadSession)
}

func (s *Server) Run() error {
	// Restore terminals from DB
	if err := s.restoreTerminals(); err != nil {
		log.Printf("Warning: failed to restore terminals: %v", err)
	}

	// Start HTTP server in goroutine
	go func() {
		log.Printf("Starting server on http://localhost:%d", s.port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	return s.Shutdown()
}

func (s *Server) Shutdown() error {
	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Kill all terminals
	if err := s.terminalManager.Shutdown(); err != nil {
		log.Printf("Terminal manager shutdown error: %v", err)
	}

	// Close database
	if err := s.db.Close(); err != nil {
		log.Printf("Database close error: %v", err)
	}

	return nil
}

func (s *Server) restoreTerminals() error {
	layout, err := s.db.GetActiveLayout()
	if err != nil {
		return err
	}

	return s.terminalManager.ApplyLayout(layout.LayoutType)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get current user
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	// Render layout
	ui.Layout(user).Render(r.Context(), w)
}
```

**Step 2: Create handlers**

File: `internal/server/handlers.go`

```go
package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/corymacd/stratusshell/internal/ui"
)

func (s *Server) handleGetLayout(w http.ResponseWriter, r *http.Request) {
	terminals := s.terminalManager.GetTerminals()
	layout, err := s.db.GetActiveLayout()
	if err != nil {
		s.handleError(w, err, "Failed to get layout")
		return
	}

	// Convert to template data
	termData := make([]ui.TerminalData, len(terminals))
	for i, t := range terminals {
		termData[i] = ui.TerminalData{
			ID:    t.ID,
			Port:  t.Port,
			Title: t.Title,
		}
	}

	ui.TerminalContainer(termData, layout.LayoutType).Render(r.Context(), w)
}

func (s *Server) handleLayoutHorizontal(w http.ResponseWriter, r *http.Request) {
	s.applyLayoutAndRespond(w, r, "horizontal")
}

func (s *Server) handleLayoutVertical(w http.ResponseWriter, r *http.Request) {
	s.applyLayoutAndRespond(w, r, "vertical")
}

func (s *Server) handleLayoutGrid(w http.ResponseWriter, r *http.Request) {
	s.applyLayoutAndRespond(w, r, "grid")
}

func (s *Server) applyLayoutAndRespond(w http.ResponseWriter, r *http.Request, layoutType string) {
	if err := s.terminalManager.ApplyLayout(layoutType); err != nil {
		s.handleError(w, err, "Failed to apply layout")
		return
	}
	s.handleGetLayout(w, r)
}

func (s *Server) handleAddTerminal(w http.ResponseWriter, r *http.Request) {
	terminals := s.terminalManager.GetTerminals()
	title := fmt.Sprintf("Terminal %d", len(terminals)+1)

	_, err := s.terminalManager.SpawnTerminal(title, "/bin/bash", "")
	if err != nil {
		s.handleError(w, err, "Failed to add terminal")
		return
	}

	s.handleGetLayout(w, r)
}

func (s *Server) handleTerminalAction(w http.ResponseWriter, r *http.Request) {
	// Extract terminal ID from path: /api/terminal/{id} or /api/terminal/{id}/rename
	path := strings.TrimPrefix(r.URL.Path, "/api/terminal/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		http.Error(w, "Invalid terminal ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid terminal ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := s.terminalManager.KillTerminal(id); err != nil {
			s.handleError(w, err, "Failed to delete terminal")
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		if len(parts) > 1 && parts[1] == "rename" {
			r.ParseForm()
			newTitle := r.FormValue("title")
			if newTitle == "" {
				http.Error(w, "Title required", http.StatusBadRequest)
				return
			}

			terminal, ok := s.terminalManager.GetTerminal(id)
			if !ok {
				http.Error(w, "Terminal not found", http.StatusNotFound)
				return
			}

			terminal.Title = newTitle
			w.WriteHeader(http.StatusOK)
		}
	}
}

func (s *Server) handleSaveSessionModal(w http.ResponseWriter, r *http.Request) {
	ui.SaveSessionModal().Render(r.Context(), w)
}

func (s *Server) handleSaveSession(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		s.handleError(w, fmt.Errorf("name required"), "Session name is required")
		return
	}

	// Create session
	sessionID, err := s.db.CreateSession(name, description)
	if err != nil {
		s.handleError(w, err, "Failed to save session")
		return
	}

	// Save all current terminals
	terminals := s.terminalManager.GetTerminals()
	for i, t := range terminals {
		if err := s.db.SaveSessionTerminal(sessionID, i, t.Title, "/bin/bash", ""); err != nil {
			log.Printf("Warning: failed to save terminal %d: %v", t.ID, err)
		}
	}

	ui.SuccessMessage("Session saved successfully").Render(r.Context(), w)
}

func (s *Server) handleListSessionsModal(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.db.GetAllSessions()
	if err != nil {
		s.handleError(w, err, "Failed to load sessions")
		return
	}

	sessionData := make([]ui.SessionData, len(sessions))
	for i, sess := range sessions {
		sessionData[i] = ui.SessionData{
			ID:          sess.ID,
			Name:        sess.Name,
			Description: sess.Description,
		}
	}

	ui.LoadSessionModal(sessionData).Render(r.Context(), w)
}

func (s *Server) handleLoadSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID
	path := strings.TrimPrefix(r.URL.Path, "/api/session/load/")
	sessionID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Get session terminals
	sessionTerminals, err := s.db.GetSessionTerminals(sessionID)
	if err != nil {
		s.handleError(w, err, "Failed to load session")
		return
	}

	// Kill all current terminals
	for _, t := range s.terminalManager.GetTerminals() {
		s.terminalManager.KillTerminal(t.ID)
	}

	// Clear active terminals table
	s.db.ClearActiveTerminals()

	// Spawn terminals from session
	for _, st := range sessionTerminals {
		_, err := s.terminalManager.SpawnTerminal(st.Title, st.Shell, st.WorkingDir)
		if err != nil {
			log.Printf("Warning: failed to spawn terminal: %v", err)
		}
	}

	// Update layout
	layoutType := "horizontal"
	if len(sessionTerminals) > 2 {
		layoutType = "grid"
	}
	s.db.UpdateActiveLayout(layoutType, len(sessionTerminals))

	s.handleGetLayout(w, r)
}

func (s *Server) handleError(w http.ResponseWriter, err error, userMsg string) {
	log.Printf("Error: %v", err)
	w.Header().Set("HX-Retarget", "#modal")
	w.Header().Set("HX-Reswap", "innerHTML")
	w.WriteHeader(http.StatusInternalServerError)
	ui.ErrorToast(userMsg).Render(r.Context(), w)
}
```

**Step 3: Update serve command to use server**

File: `cmd/serve.go`

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/corymacd/stratusshell/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web UI server",
	Long:  `Start HTTP server with GoTTY terminal management and HTMX UI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		dbPath, _ := cmd.Flags().GetString("db")

		// Default DB path if not specified
		if dbPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			dbPath = filepath.Join(homeDir, ".stratusshell", "data.db")
		}

		// Create and run server
		srv, err := server.NewServer(port, dbPath)
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}

		return srv.Run()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringP("user", "u", "", "Run as specific user (when started by root)")
	serveCmd.Flags().IntP("port", "p", 8080, "HTTP port")
	serveCmd.Flags().String("db", "", "Database path (default: ~/.stratusshell/data.db)")
}
```

**Step 4: Generate templ files and test build**

```bash
templ generate
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 5: Test server (manual)**

```bash
./stratusshell serve --port=8080
```

Open browser to http://localhost:8080

Expected: See StratusShell UI with menubar (terminals may fail to spawn if gotty not installed)

**Step 6: Commit**

```bash
git add internal/server/ cmd/serve.go
git commit -m "feat(server): add HTTP server with HTMX handlers and terminal management"
```

---

## Phase 6: Provisioning System

### Task 10: Package Manager Detection

**Files:**
- Create: `internal/provision/packagemanager.go`

**Step 1: Create package manager enum and detection**

File: `internal/provision/packagemanager.go`

```go
package provision

import (
	"errors"
	"os/exec"
)

type PackageManager int

const (
	APT PackageManager = iota
	YUM
	DNF
	PACMAN
)

func (pm PackageManager) String() string {
	switch pm {
	case APT:
		return "apt"
	case YUM:
		return "yum"
	case DNF:
		return "dnf"
	case PACMAN:
		return "pacman"
	default:
		return "unknown"
	}
}

func DetectPackageManager() (PackageManager, error) {
	managers := []struct {
		pm      PackageManager
		command string
	}{
		{APT, "apt-get"},
		{DNF, "dnf"},
		{YUM, "yum"},
		{PACMAN, "pacman"},
	}

	for _, m := range managers {
		if _, err := exec.LookPath(m.command); err == nil {
			return m.pm, nil
		}
	}

	return 0, errors.New("no supported package manager found")
}

func (pm PackageManager) Install(packages ...string) error {
	var cmd *exec.Cmd

	switch pm {
	case APT:
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("apt-get", args...)
	case YUM:
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("yum", args...)
	case DNF:
		args := append([]string{"install", "-y"}, packages...)
		cmd = exec.Command("dnf", args...)
	case PACMAN:
		args := append([]string{"-S", "--noconfirm"}, packages...)
		cmd = exec.Command("pacman", args...)
	default:
		return errors.New("unsupported package manager")
	}

	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
```

**Step 2: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/provision/packagemanager.go
git commit -m "feat(provision): add package manager detection and install abstraction"
```

---

### Task 11: User Provisioning

**Files:**
- Create: `internal/provision/user.go`
- Create: `internal/provision/sudo.go`

**Step 1: Create user provisioning**

File: `internal/provision/user.go`

```go
package provision

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
)

func CreateUser(username, shell string) error {
	// Check if user already exists
	if _, err := user.Lookup(username); err == nil {
		return fmt.Errorf("user %s already exists", username)
	}

	// Create user with home directory
	cmd := exec.Command("useradd",
		"-m",                    // Create home directory
		"-s", shell,            // Set shell
		"-c", "StratusShell User", // Comment
		username,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create user: %w (output: %s)", err, output)
	}

	return nil
}

func DeleteUser(username string) error {
	cmd := exec.Command("userdel", "-r", username)
	return cmd.Run()
}

func UserExists(username string) bool {
	_, err := user.Lookup(username)
	return err == nil
}

func SetUserShell(username, shell string) error {
	cmd := exec.Command("chsh", "-s", shell, username)
	return cmd.Run()
}

func GetUserHomeDir(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}

func ChownRecursive(path, username string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return err
	}

	cmd := exec.Command("chown", "-R", fmt.Sprintf("%s:%s", u.Uid, u.Gid), path)
	return cmd.Run()
}
```

**Step 2: Create sudo configuration**

File: `internal/provision/sudo.go`

```go
package provision

import (
	"fmt"
	"os"
	"path/filepath"
)

func ConfigurePasswordlessSudo(username string) error {
	sudoersFile := filepath.Join("/etc/sudoers.d", fmt.Sprintf("stratusshell-%s", username))

	content := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", username)

	// Write with 0440 permissions (required for sudoers files)
	if err := os.WriteFile(sudoersFile, []byte(content), 0440); err != nil {
		return fmt.Errorf("failed to write sudoers file: %w", err)
	}

	return nil
}

func RemoveSudoersConfig(username string) error {
	sudoersFile := filepath.Join("/etc/sudoers.d", fmt.Sprintf("stratusshell-%s", username))
	return os.Remove(sudoersFile)
}
```

**Step 3: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 4: Commit**

```bash
git add internal/provision/user.go internal/provision/sudo.go
git commit -m "feat(provision): add user creation and sudoers configuration"
```

---

### Task 12: Tool Installation

**Files:**
- Create: `internal/provision/tools.go`
- Create: `internal/provision/config.go`

**Step 1: Create config struct**

File: `internal/provision/config.go`

```go
package provision

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	User      UserConfig      `yaml:"user"`
	Base      []string        `yaml:"base_packages"`
	Cloud     CloudConfig     `yaml:"cloud"`
	Languages LanguagesConfig `yaml:"languages"`
	Shell     ShellConfig     `yaml:"shell"`
}

type UserConfig struct {
	Shell string `yaml:"shell"`
}

type CloudConfig struct {
	AWS       bool `yaml:"aws"`
	GCloud    bool `yaml:"gcloud"`
	Kubectl   bool `yaml:"kubectl"`
	Docker    bool `yaml:"docker"`
	Terraform bool `yaml:"terraform"`
}

type LanguagesConfig struct {
	Go   GoConfig   `yaml:"go"`
	Node NodeConfig `yaml:"node"`
}

type GoConfig struct {
	Enabled bool     `yaml:"enabled"`
	Version string   `yaml:"version"`
	Tools   []string `yaml:"tools"`
}

type NodeConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Version        string   `yaml:"version"`
	PackageManager string   `yaml:"package_manager"`
	GlobalPackages []string `yaml:"global_packages"`
}

type ShellConfig struct {
	Zsh       bool `yaml:"zsh"`
	OhMyZsh   bool `yaml:"oh_my_zsh"`
	Tmux      bool `yaml:"tmux"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
```

**Step 2: Create tool installer**

File: `internal/provision/tools.go`

```go
package provision

import (
	"fmt"
	"log"
)

type Provisioner struct {
	pm       PackageManager
	username string
	config   *Config
}

func NewProvisioner(username string, config *Config) (*Provisioner, error) {
	pm, err := DetectPackageManager()
	if err != nil {
		return nil, err
	}

	return &Provisioner{
		pm:       pm,
		username: username,
		config:   config,
	}, nil
}

func (p *Provisioner) InstallBasePackages() error {
	log.Println("Installing base packages...")

	// Translate package names if needed
	packages := p.translatePackageNames(p.config.Base)

	if err := p.pm.Install(packages...); err != nil {
		return fmt.Errorf("failed to install base packages: %w", err)
	}

	log.Printf("Installed %d base packages", len(packages))
	return nil
}

func (p *Provisioner) translatePackageNames(packages []string) []string {
	translated := make([]string, 0, len(packages))

	for _, pkg := range packages {
		switch pkg {
		case "build-essential":
			if p.pm == YUM || p.pm == DNF {
				translated = append(translated, "gcc", "gcc-c++", "make")
			} else if p.pm == PACMAN {
				translated = append(translated, "base-devel")
			} else {
				translated = append(translated, pkg)
			}
		default:
			translated = append(translated, pkg)
		}
	}

	return translated
}

func (p *Provisioner) InstallCloudTools() error {
	log.Println("Installing cloud tools...")

	installed := 0

	if p.config.Cloud.AWS {
		if err := p.installAWSCLI(); err != nil {
			log.Printf("Warning: failed to install AWS CLI: %v", err)
		} else {
			installed++
		}
	}

	if p.config.Cloud.Docker {
		if err := p.installDocker(); err != nil {
			log.Printf("Warning: failed to install Docker: %v", err)
		} else {
			installed++
		}
	}

	if p.config.Cloud.Kubectl {
		if err := p.installKubectl(); err != nil {
			log.Printf("Warning: failed to install kubectl: %v", err)
		} else {
			installed++
		}
	}

	log.Printf("Installed %d/%d cloud tools", installed, p.countEnabledCloudTools())
	return nil
}

func (p *Provisioner) installAWSCLI() error {
	// Simplified: install via package manager if available
	return p.pm.Install("awscli")
}

func (p *Provisioner) installDocker() error {
	// Install docker
	if err := p.pm.Install("docker", "docker.io"); err != nil {
		// Try alternative package name
		if err := p.pm.Install("docker-ce"); err != nil {
			return err
		}
	}

	// Add user to docker group (requires re-login to take effect)
	// This is a simplified version - production would use proper user/group management
	log.Printf("Note: User %s needs to log out and back in for docker group to take effect", p.username)

	return nil
}

func (p *Provisioner) installKubectl() error {
	return p.pm.Install("kubectl")
}

func (p *Provisioner) countEnabledCloudTools() int {
	count := 0
	if p.config.Cloud.AWS {
		count++
	}
	if p.config.Cloud.GCloud {
		count++
	}
	if p.config.Cloud.Kubectl {
		count++
	}
	if p.config.Cloud.Docker {
		count++
	}
	if p.config.Cloud.Terraform {
		count++
	}
	return count
}
```

**Step 3: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 4: Commit**

```bash
git add internal/provision/tools.go internal/provision/config.go
git commit -m "feat(provision): add base and cloud tool installation"
```

---

### Task 13: Language Toolchains (Go & Node)

**Files:**
- Create: `internal/provision/toolchains.go`

**Step 1: Create toolchain installers**

File: `internal/provision/toolchains.go`

```go
package provision

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func (p *Provisioner) InstallLanguageToolchains() error {
	log.Println("Installing language toolchains...")

	if p.config.Languages.Go.Enabled {
		if err := p.installGoToolchain(); err != nil {
			log.Printf("Warning: failed to install Go toolchain: %v", err)
		}
	}

	if p.config.Languages.Node.Enabled {
		if err := p.installNodeToolchain(); err != nil {
			log.Printf("Warning: failed to install Node toolchain: %v", err)
		}
	}

	return nil
}

func (p *Provisioner) installGoToolchain() error {
	log.Println("Installing Go toolchain...")

	// Install Go via package manager (simplified)
	if err := p.pm.Install("golang"); err != nil {
		return fmt.Errorf("failed to install go: %w", err)
	}

	// Get user home directory
	homeDir, err := GetUserHomeDir(p.username)
	if err != nil {
		return err
	}

	// Install Go tools
	for _, tool := range p.config.Languages.Go.Tools {
		if err := p.installGoTool(tool); err != nil {
			log.Printf("Warning: failed to install Go tool %s: %v", tool, err)
		}
	}

	// Create stratusshell env file
	envFile := filepath.Join(homeDir, ".stratusshell", "env.sh")
	envContent := `
# StratusShell Go Environment
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
`

	if err := os.MkdirAll(filepath.Dir(envFile), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		return err
	}

	// Set ownership
	return ChownRecursive(filepath.Dir(envFile), p.username)
}

func (p *Provisioner) installGoTool(tool string) error {
	var packagePath string

	switch tool {
	case "golangci-lint":
		packagePath = "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	case "gopls":
		packagePath = "golang.org/x/tools/gopls@latest"
	case "delve":
		packagePath = "github.com/go-delve/delve/cmd/dlv@latest"
	default:
		return fmt.Errorf("unknown go tool: %s", tool)
	}

	cmd := exec.Command("go", "install", packagePath)
	return cmd.Run()
}

func (p *Provisioner) installNodeToolchain() error {
	log.Println("Installing Node toolchain...")

	homeDir, err := GetUserHomeDir(p.username)
	if err != nil {
		return err
	}

	// Install nvm (Node Version Manager)
	nvmDir := filepath.Join(homeDir, ".nvm")
	if err := os.MkdirAll(nvmDir, 0755); err != nil {
		return err
	}

	// This is simplified - production would download and run nvm install script
	// For now, just install nodejs via package manager
	if err := p.pm.Install("nodejs", "npm"); err != nil {
		return fmt.Errorf("failed to install nodejs: %w", err)
	}

	// Install global packages
	for _, pkg := range p.config.Languages.Node.GlobalPackages {
		if err := p.installNpmGlobal(pkg); err != nil {
			log.Printf("Warning: failed to install npm package %s: %v", pkg, err)
		}
	}

	// Install pnpm if configured
	if p.config.Languages.Node.PackageManager == "pnpm" {
		if err := p.installNpmGlobal("pnpm"); err != nil {
			log.Printf("Warning: failed to install pnpm: %v", err)
		}
	}

	return ChownRecursive(nvmDir, p.username)
}

func (p *Provisioner) installNpmGlobal(package_ string) error {
	cmd := exec.Command("npm", "install", "-g", package_)
	return cmd.Run()
}
```

**Step 2: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/provision/toolchains.go
git commit -m "feat(provision): add Go and Node toolchain installation"
```

---

### Task 14: Shell Environment Setup

**Files:**
- Create: `internal/provision/shell.go`

**Step 1: Create shell setup**

File: `internal/provision/shell.go`

```go
package provision

import (
	"log"
	"os"
	"path/filepath"
)

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

	// Source stratusshell env in bashrc/zshrc
	if err := p.configureShellRC(); err != nil {
		log.Printf("Warning: failed to configure shell RC: %v", err)
	}

	return nil
}

func (p *Provisioner) installZsh() error {
	if err := p.pm.Install("zsh"); err != nil {
		return err
	}

	// Set as default shell
	return SetUserShell(p.username, "/bin/zsh")
}

func (p *Provisioner) installTmux() error {
	return p.pm.Install("tmux")
}

func (p *Provisioner) configureShellRC() error {
	homeDir, err := GetUserHomeDir(p.username)
	if err != nil {
		return err
	}

	// Determine which RC file to update
	rcFile := filepath.Join(homeDir, ".bashrc")
	if p.config.Shell.Zsh {
		rcFile = filepath.Join(homeDir, ".zshrc")
	}

	// Append stratusshell env sourcing
	sourceCmd := "\n# StratusShell Environment\nif [ -f ~/.stratusshell/env.sh ]; then\n    source ~/.stratusshell/env.sh\nfi\n"

	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(sourceCmd)
	if err != nil {
		return err
	}

	return ChownRecursive(rcFile, p.username)
}
```

**Step 2: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/provision/shell.go
git commit -m "feat(provision): add shell environment setup with zsh and tmux"
```

---

### Task 15: Implement Init Command

**Files:**
- Modify: `cmd/init.go`

**Step 1: Wire up init command**

File: `cmd/init.go`

```go
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/corymacd/stratusshell/internal/provision"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Provision a development environment",
	Long:  `Create a system user and install complete development toolchain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running as root
		if os.Geteuid() != 0 {
			return fmt.Errorf("init command must be run as root (use sudo)")
		}

		username, _ := cmd.Flags().GetString("user")
		shell, _ := cmd.Flags().GetString("shell")
		configPath, _ := cmd.Flags().GetString("config")
		skipTools, _ := cmd.Flags().GetBool("skip-tools")

		log.Printf("Provisioning user: %s", username)

		// Load config
		config, err := provision.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Override shell if specified
		if shell != "" {
			config.User.Shell = shell
		}

		// Create user
		log.Println("Creating user...")
		if err := provision.CreateUser(username, config.User.Shell); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Configure passwordless sudo
		log.Println("Configuring passwordless sudo...")
		if err := provision.ConfigurePasswordlessSudo(username); err != nil {
			// Rollback user creation
			provision.DeleteUser(username)
			return fmt.Errorf("failed to configure sudo: %w", err)
		}

		if skipTools {
			log.Println("Skipping tool installation (--skip-tools)")
			log.Println("✓ User provisioning complete")
			return nil
		}

		// Create provisioner
		p, err := provision.NewProvisioner(username, config)
		if err != nil {
			log.Printf("Warning: failed to create provisioner: %v", err)
			log.Println("User created but tool installation skipped")
			return nil
		}

		// Install base packages
		if err := p.InstallBasePackages(); err != nil {
			log.Printf("Warning: base package installation failed: %v", err)
		}

		// Install cloud tools
		if err := p.InstallCloudTools(); err != nil {
			log.Printf("Warning: cloud tools installation failed: %v", err)
		}

		// Install language toolchains
		if err := p.InstallLanguageToolchains(); err != nil {
			log.Printf("Warning: language toolchains installation failed: %v", err)
		}

		// Setup shell environment
		if err := p.SetupShellEnvironment(); err != nil {
			log.Printf("Warning: shell setup failed: %v", err)
		}

		log.Println("✓ Provisioning complete")
		log.Printf("User %s is ready. Database will be created at ~/.stratusshell/data.db on first serve", username)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("user", "u", "", "Username to create (required)")
	initCmd.MarkFlagRequired("user")
	initCmd.Flags().String("shell", "", "Default shell (overrides config)")
	initCmd.Flags().String("config", "/etc/stratusshell/default.yaml", "Config file path")
	initCmd.Flags().Bool("skip-tools", false, "Create user only, skip tool installation")
}
```

**Step 2: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 3: Test help text**

```bash
./stratusshell init --help
```

Expected: Shows usage and flags

**Step 4: Commit**

```bash
git add cmd/init.go
git commit -m "feat(cmd): implement init command with full provisioning workflow"
```

---

## Phase 7: Systemd Service Installation

### Task 16: Systemd Service Generator

**Files:**
- Create: `internal/service/systemd.go`

**Step 1: Create systemd service generator**

File: `internal/service/systemd.go`

```go
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const serviceTemplate = `[Unit]
Description=StratusShell for {{.User}}
After=network.target

[Service]
Type=simple
User={{.User}}
WorkingDirectory={{.HomeDir}}
ExecStart={{.BinaryPath}} serve --user={{.User}} --port={{.Port}}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

type ServiceConfig struct {
	User       string
	HomeDir    string
	BinaryPath string
	Port       int
}

func InstallSystemdService(username string, port int) error {
	// Get user home directory
	homeDir := fmt.Sprintf("/home/%s", username)

	// Get current binary path
	binaryPath, err := os.Executable()
	if err != nil {
		// Default to /usr/local/bin
		binaryPath = "/usr/local/bin/stratusshell"
	}

	config := ServiceConfig{
		User:       username,
		HomeDir:    homeDir,
		BinaryPath: binaryPath,
		Port:       port,
	}

	// Generate service file
	serviceName := fmt.Sprintf("stratusshell-%s.service", username)
	servicePath := filepath.Join("/etc/systemd/system", serviceName)

	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, config); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	// Enable service
	if err := exec.Command("systemctl", "enable", serviceName).Run(); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Start service
	if err := exec.Command("systemctl", "start", serviceName).Run(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	return nil
}

func UninstallSystemdService(username string) error {
	serviceName := fmt.Sprintf("stratusshell-%s.service", username)

	// Stop service
	exec.Command("systemctl", "stop", serviceName).Run()

	// Disable service
	exec.Command("systemctl", "disable", serviceName).Run()

	// Remove service file
	servicePath := filepath.Join("/etc/systemd/system", serviceName)
	os.Remove(servicePath)

	// Reload systemd
	exec.Command("systemctl", "daemon-reload").Run()

	return nil
}

func GetServiceStatus(username string) (string, error) {
	serviceName := fmt.Sprintf("stratusshell-%s.service", username)

	cmd := exec.Command("systemctl", "status", serviceName)
	output, err := cmd.CombinedOutput()

	return string(output), err
}
```

**Step 2: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/service/systemd.go
git commit -m "feat(service): add systemd service installation and management"
```

---

### Task 17: Implement Install Command

**Files:**
- Modify: `cmd/install.go`

**Step 1: Wire up install command**

File: `cmd/install.go`

```go
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/corymacd/stratusshell/internal/service"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install systemd service",
	Long:  `Generate and enable systemd service for StratusShell.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running as root
		if os.Geteuid() != 0 {
			return fmt.Errorf("install command must be run as root (use sudo)")
		}

		username, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")

		log.Printf("Installing systemd service for user: %s", username)

		if err := service.InstallSystemdService(username, port); err != nil {
			return fmt.Errorf("failed to install service: %w", err)
		}

		log.Println("✓ Service installed successfully")
		log.Printf("Service: stratusshell-%s.service", username)
		log.Printf("URL: http://localhost:%d", port)
		log.Println()
		log.Println("Useful commands:")
		log.Printf("  sudo systemctl status stratusshell-%s", username)
		log.Printf("  sudo systemctl restart stratusshell-%s", username)
		log.Printf("  sudo systemctl stop stratusshell-%s", username)
		log.Printf("  sudo journalctl -u stratusshell-%s -f", username)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("user", "u", "", "User that service runs as (required)")
	installCmd.MarkFlagRequired("user")
	installCmd.Flags().IntP("port", "p", 8080, "HTTP port")
}
```

**Step 2: Test build**

```bash
go build -o stratusshell main.go
```

Expected: Build succeeds

**Step 3: Test help text**

```bash
./stratusshell install --help
```

Expected: Shows usage and flags

**Step 4: Commit**

```bash
git add cmd/install.go
git commit -m "feat(cmd): implement install command for systemd service"
```

---

## Phase 8: Final Integration & Testing

### Task 18: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update CLAUDE.md with new architecture**

File: `CLAUDE.md`

```markdown
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

StratusShell is a cloud development environment provisioning tool. It provisions Linux users with complete development toolchains and provides a web-based terminal interface with session management.

## Common Commands

### Development
```bash
# Install templ CLI (required for building)
go install github.com/a-h/templ/cmd/templ@latest

# Generate templ files
templ generate

# Download dependencies
go mod download

# Build the application
go build -o stratusshell main.go

# Install to system
sudo cp stratusshell /usr/local/bin/
sudo mkdir -p /etc/stratusshell
sudo cp configs/default.yaml /etc/stratusshell/
```

### Testing
```bash
# Run unit tests
go test ./...

# Run with verbose output
go test -v ./internal/server/

# Test specific package
go test ./internal/db/
```

### Usage
```bash
# Provision user and install tools
sudo stratusshell init --user=developer

# Test web UI manually
stratusshell serve --port=8080

# Install as systemd service
sudo stratusshell install --user=developer --port=8080

# Service management
sudo systemctl status stratusshell-developer
sudo systemctl restart stratusshell-developer
sudo journalctl -u stratusshell-developer -f
```

## Architecture

### High-Level Structure

StratusShell uses a feature-based package architecture:

**Commands** (`cmd/`):
- `root.go` - Cobra root command
- `init.go` - User provisioning and tool installation
- `serve.go` - Web server startup
- `install.go` - Systemd service installation

**Provisioning** (`internal/provision/`):
- `user.go` - System user creation/deletion
- `sudo.go` - Sudoers configuration
- `packagemanager.go` - Package manager detection and abstraction
- `tools.go` - Base package and cloud tool installation
- `toolchains.go` - Language-specific toolchain setup (Go, Node)
- `shell.go` - Shell environment configuration
- `config.go` - YAML configuration parsing

**Server** (`internal/server/`):
- `server.go` - HTTP server and lifecycle management
- `handlers.go` - HTMX endpoint handlers
- `terminal.go` - Terminal manager (GoTTY orchestration)
- `portpool.go` - Dynamic port allocation for terminals

**Database** (`internal/db/`):
- `db.go` - SQLite connection and migrations
- `preferences.go` - User preferences CRUD
- `sessions.go` - Session save/restore CRUD
- `terminals.go` - Active terminal state CRUD
- `schema.sql` - Database schema (embedded)

**UI** (`internal/ui/`):
- `layout.templ` - Base page layout
- `menubar.templ` - Top navigation with dropdowns
- `terminal.templ` - Terminal pane components
- `modals.templ` - Modal dialogs (save/load session, preferences)

**Service** (`internal/service/`):
- `systemd.go` - Systemd service file generation and management

### Key Components

**Init Command Flow**:
1. Validate running as root
2. Load YAML configuration from `/etc/stratusshell/default.yaml`
3. Create system user with `useradd`
4. Configure passwordless sudo in `/etc/sudoers.d/`
5. Detect package manager (apt/yum/dnf/pacman)
6. Install base packages (git, curl, build-essential)
7. Install cloud tools (aws, gcloud, kubectl, docker, terraform)
8. Install language toolchains (Go, Node with meta packages)
9. Setup shell environment (zsh, tmux, RC file sourcing)
10. Create `~/.stratusshell/env.sh` with PATH additions

**Serve Command Flow**:
1. Load configuration (port, database path)
2. Open SQLite database at `~/.stratusshell/data.db`
3. Run database migrations
4. Create TerminalManager with port pool (8081-8181)
5. Restore previous layout and terminals from database
6. Spawn GoTTY instances for each terminal
7. Start HTTP server with HTMX handlers
8. Serve Templ-rendered UI
9. Handle graceful shutdown (kill terminals, close DB)

**Terminal Lifecycle**:
- Port pool allocates available ports (8081-8181)
- Each terminal spawns a GoTTY process with context cancellation
- Monitor goroutine detects unexpected process death
- State persists to `active_terminals` table
- Graceful shutdown kills all GoTTY processes and releases ports

**HTMX Interactions**:
- Layout changes: POST to `/api/layout/{type}` → server adjusts terminal count → returns updated HTML
- Add terminal: POST to `/api/terminals/add` → spawns GoTTY → returns new terminal pane
- Remove terminal: DELETE to `/api/terminal/{id}` → kills GoTTY → returns empty (swap outerHTML)
- Save session: POST to `/api/session/save` → stores terminals in DB → returns success modal
- Load session: POST to `/api/session/load/{id}` → kills current terminals → spawns from DB → returns layout

**Database Schema**:
- `preferences` - Key-value store for user preferences
- `sessions` - Named saved sessions
- `session_terminals` - Terminal configurations per session
- `active_layout` - Singleton table for current layout type
- `active_terminals` - Currently running terminals

### Important Implementation Details

**Templ Workflow**:
- `.templ` files are type-safe Go templates
- Run `templ generate` to create `*_templ.go` files
- Generated files are compiled with the binary
- Changes to `.templ` require regeneration

**Port Management**:
- Port pool prevents conflicts (8081-8181 range)
- Thread-safe allocation/release with mutex
- Ports released on terminal death or shutdown

**Provisioning Safety**:
- Root check for init/install commands
- User creation failures trigger rollback
- Tool installation failures are non-fatal (log and continue)
- Passwordless sudo requires `/etc/sudoers.d/` permissions (0440)

**GoTTY Integration**:
- Each terminal is an independent GoTTY process
- Context cancellation for clean shutdown
- Auto-reconnect enabled (10 second interval)
- Write permissions enabled for interactive use

**Static Files**:
- CSS embedded in binary via `//go:embed`
- HTMX loaded from CDN
- No build step required for frontend

## Worktree Directory

Worktrees are created in: `~/.config/superpowers/worktrees/StratusShell/`
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with new architecture and commands"
```

---

### Task 19: Create Makefile

**Files:**
- Create: `Makefile`

**Step 1: Create Makefile**

File: `Makefile`

```makefile
.PHONY: generate build install test integration-test clean help

help:
	@echo "StratusShell Build Commands:"
	@echo "  make generate        - Generate templ files"
	@echo "  make build           - Build binary"
	@echo "  make install         - Install to /usr/local/bin (requires sudo)"
	@echo "  make test            - Run unit tests"
	@echo "  make integration-test- Run integration tests (requires sudo/docker)"
	@echo "  make clean           - Remove build artifacts"

generate:
	templ generate

build: generate
	go build -o stratusshell main.go

install: build
	sudo cp stratusshell /usr/local/bin/
	sudo chmod +x /usr/local/bin/stratusshell
	sudo mkdir -p /etc/stratusshell
	sudo cp configs/default.yaml /etc/stratusshell/

test:
	go test ./...

integration-test:
	@echo "Integration tests require root privileges"
	INTEGRATION_TESTS=1 go test ./test/integration/...

clean:
	rm -f stratusshell
	find . -name "*_templ.go" -delete
```

**Step 2: Test Makefile**

```bash
make help
```

Expected: Shows help text

**Step 3: Commit**

```bash
git add Makefile
git commit -m "build: add Makefile with generate, build, install, test targets"
```

---

### Task 20: Final Build and Manual Test

**Step 1: Full clean build**

```bash
make clean
make build
```

Expected: Binary builds successfully

**Step 2: Test CLI help**

```bash
./stratusshell --help
./stratusshell init --help
./stratusshell serve --help
./stratusshell install --help
```

Expected: All commands show proper help text

**Step 3: Test serve command (if gotty installed)**

```bash
./stratusshell serve --port=9000
```

Open browser to http://localhost:9000

Expected: StratusShell UI loads (terminals may not work without gotty)

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete StratusShell enhancement implementation

Transform from simple dual-terminal app to comprehensive cloud dev
environment provisioning tool with:

- Cobra CLI (init/serve/install commands)
- Go-native provisioning (users, sudo, tools, toolchains)
- HTMX + Templ web UI with dynamic terminal management
- SQLite persistence for sessions/layouts/preferences
- Systemd service integration
- Port pool for dynamic terminal allocation
- Graceful shutdown and error handling

All phases complete and tested."
```

---

## Summary

This implementation plan covers:

✅ **Phase 1**: Cobra CLI structure with three commands
✅ **Phase 2**: SQLite database with schema and CRUD operations
✅ **Phase 3**: GoTTY terminal management with port pooling
✅ **Phase 4**: Templ UI components (layout, menubar, terminals, modals)
✅ **Phase 5**: HTTP server with HTMX handlers
✅ **Phase 6**: Complete provisioning system (users, sudo, tools, toolchains)
✅ **Phase 7**: Systemd service installation
✅ **Phase 8**: Documentation, Makefile, final testing

**Next Steps:**
- Use superpowers:executing-plans or superpowers:subagent-driven-development to implement
- Test on real system with sudo privileges
- Verify GoTTY installation works correctly
- Test full workflow: init → serve → install → reboot → verify service
