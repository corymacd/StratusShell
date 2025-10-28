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
- `version.go` - Version information

**Provisioning** (`internal/provision/`):
- `user.go` - System user creation/deletion
- `sudo.go` - Sudoers configuration
- `packagemanager.go` - Package manager detection and abstraction
- `tools.go` - Base package and cloud tool installation
- `toolchains.go` - Language-specific toolchain setup (Go, Node)
- `shell.go` - Shell environment configuration
- `config.go` - YAML configuration parsing
- `claude.go` - Claude AI integration and command filtering

**Server** (`internal/server/`):
- `server.go` - HTTP server and lifecycle management
- `handlers.go` - HTMX endpoint handlers
- `terminal.go` - Terminal manager (GoTTY orchestration)
- `portpool.go` - Dynamic port allocation for terminals
- `auth.go` - Authentication middleware
- `health.go` - Health check endpoints
- `gotty_wrapper.go` - GoTTY process wrapper

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

**Security** (`internal/validation/`, `internal/audit/`, `internal/middleware/`):
- Input validation and sanitization
- Security audit logging
- Security middleware for HTTP handlers

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

**Claude AI Integration**:
- Configuration in `configs/default.yaml` under `claude` section
- Command filtering with allow/deny/ask lists
- Security validation for allowed commands
- Audit logging for all command executions
- MCP server support for Playwright, Linear, and GitHub
- Automatic MCP installation during user provisioning
- Settings file at `~/.claude/settings.json`

**MCP Servers**:
- **Playwright**: Browser automation and testing (`@playwright/mcp`)
- **Linear**: Project management integration (`@mseep/linear-mcp`)
- **GitHub**: Repository operations (`github-mcp-server`)
- Configured in `claude.mcp_servers` section of YAML
- Installed globally via npm during provisioning
- See `docs/CLAUDE_MCP.md` for detailed documentation

## Worktree Directory

Worktrees are created in: `~/.config/superpowers/worktrees/StratusShell/`
