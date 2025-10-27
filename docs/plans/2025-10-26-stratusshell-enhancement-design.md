# StratusShell Enhancement Design

**Date:** 2025-10-26
**Status:** Design Approved
**Author:** Design Session with Claude Code

## Executive Summary

Transform StratusShell from a simple dual-terminal web app into a comprehensive cloud development environment provisioning tool. The enhanced system will provision Linux users with complete development toolchains and provide a dynamic web-based terminal interface with session management.

**Core Value Proposition:**
- One command to provision a fully configured development environment
- Persistent terminal sessions with layout management
- Production-ready systemd service integration
- Type-safe, maintainable UI with HTMX + Templ

## Goals & Constraints

### Primary Goals
1. Add Cobra CLI with three commands: `init` (provision), `serve` (web UI), `install` (systemd)
2. Provision system users with passwordless sudo and complete dev toolchains
3. Enhance web UI with dynamic terminal management (HTMX + Templ)
4. Persist sessions, layouts, and preferences (SQLite)
5. Support systemd service installation for always-on access

### Constraints
- Keep it simple: single binary, minimal dependencies
- Preserve GoTTY features: auto-reconnect, write permissions, PTY support
- Avoid heavy frontend: no npm/webpack, server-side rendering with HTMX
- Go-native provisioning: no shell scripts, testable Go code

### Success Criteria
- `init` command successfully provisions user + tools
- Web UI supports dynamic layout changes without page reload
- Sessions persist across server restarts
- Systemd service runs reliably and survives reboots

## Architecture Overview

### Package Structure (Feature-Based)

```
stratusshell/
├── main.go                    # Entry point, Cobra setup
├── cmd/
│   ├── root.go               # Root command
│   ├── init.go               # Provisioning command
│   ├── serve.go              # Web server command
│   └── install.go            # Systemd service installation
├── internal/
│   ├── provision/            # User & tool provisioning
│   │   ├── user.go           # useradd, home dir, shell
│   │   ├── sudo.go           # /etc/sudoers.d/ passwordless config
│   │   ├── tools.go          # Package manager detection + install
│   │   └── toolchains.go     # Go (gvm), Node (nvm), meta packages
│   ├── service/              # Systemd management
│   │   └── systemd.go        # Generate .service, enable, start
│   ├── server/               # Web server (used by serve + service)
│   │   ├── server.go         # HTTP + GoTTY orchestration
│   │   ├── handlers.go       # HTMX endpoint handlers
│   │   └── gotty.go          # Terminal lifecycle management
│   ├── db/                   # SQLite persistence
│   │   ├── db.go             # Schema, migrations, connection
│   │   ├── sessions.go       # Layout persistence
│   │   └── preferences.go    # Theme, font, keybindings
│   └── ui/                   # Templ components
│       ├── layout.templ      # Base page structure
│       ├── menubar.templ     # Layout/config controls
│       ├── terminal.templ    # Terminal pane component
│       └── modals.templ      # Config dialogs
└── configs/
    └── default.yaml          # Tool installation manifest
```

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Cobra CLI** | Industry-standard Go CLI framework, excellent UX |
| **Separate init/install** | Test provisioning before committing to service |
| **Go-native provisioning** | Testable, better error handling than shell scripts |
| **Keep GoTTY** | Proven terminal emulation, WebSocket support |
| **HTMX for UI** | Server-driven, no heavy frontend build |
| **Templ templates** | Type-safe, compile-time checking, IDE support |
| **SQLite persistence** | Embedded, zero-config, sufficient for single-user |
| **Systemd service** | Production-ready, auto-restart, survives reboots |
| **Feature-based packages** | Clear boundaries, testable, maintainable |

## Command Interfaces

### 1. Init Command (Provisioning)

**Usage:**
```bash
sudo stratusshell init --user=developer [options]
```

**Options:**
- `--user` (required): Username to create
- `--shell`: Shell path (default: /bin/bash)
- `--config`: Path to custom tools.yaml
- `--skip-tools`: Create user only, skip tool installation

**What it does:**
1. Validates running as root
2. Creates system user with home directory
3. Configures passwordless sudo in `/etc/sudoers.d/stratusshell-<user>`
4. Installs base tools (git, curl, build-essential)
5. Installs cloud CLIs (aws, gcloud, kubectl, docker, terraform)
6. Sets up shell environment (zsh, oh-my-zsh if configured)
7. Installs language toolchains:
   - **Go:** Install Go binary, gvm, golangci-lint, gopls, delve
   - **Node:** Install via nvm, pnpm, yarn, typescript, prettier, eslint
8. Creates SQLite database: `~/.stratusshell/data.db`
9. Sets ownership (user:user) for all created files

**Error Handling:**
- User creation failures: Rollback (userdel -r)
- Sudo config failures: Rollback
- Tool installation failures: Log and continue (non-fatal)
- Summary report shows success/failure per tool

### 2. Serve Command (Web UI)

**Usage:**
```bash
# Manual testing
stratusshell serve --user=developer --port=8080

# As current user
stratusshell serve --port=8080

# Called by systemd
ExecStart=/usr/local/bin/stratusshell serve --user=developer --port=8080
```

**What it does:**
1. Loads SQLite database from `~/.stratusshell/data.db`
2. Restores previous session (terminals, layout) from DB
3. Starts HTTP server on specified port
4. Spawns GoTTY instances for each terminal
5. Serves HTMX-powered web UI

**Options:**
- `--user`: Run as specific user (when started by root)
- `--port`: HTTP port (default: 8080)
- `--db`: Database path (default: ~/.stratusshell/data.db)

### 3. Install Command (Systemd Service)

**Usage:**
```bash
sudo stratusshell install --user=developer [options]
```

**Options:**
- `--user` (required): User that service runs as
- `--port`: HTTP port (default: 8080)

**What it does:**
1. Generates systemd service file: `/etc/systemd/system/stratusshell-<user>.service`
2. Runs `systemctl daemon-reload`
3. Runs `systemctl enable stratusshell-<user>`
4. Runs `systemctl start stratusshell-<user>`

**Service file template:**
```ini
[Unit]
Description=StratusShell for %i
After=network.target

[Service]
Type=simple
User=developer
ExecStart=/usr/local/bin/stratusshell serve --user=developer --port=8080
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

## Web UI Architecture

### Layout Structure

```
┌─────────────────────────────────────────────────┐
│  Menubar [Layout▾] [Config▾] [Sessions▾]       │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────┐  ┌──────────────┐            │
│  │ Terminal 1   │  │ Terminal 2   │            │
│  │ (GoTTY)      │  │ (GoTTY)      │            │
│  │              │  │              │            │
│  └──────────────┘  └──────────────┘            │
│                                                  │
│  ┌──────────────────────────────────┐          │
│  │ Terminal 3 (GoTTY)               │          │
│  │                                  │          │
│  └──────────────────────────────────┘          │
└─────────────────────────────────────────────────┘
```

### Templ Components

**1. layout.templ** - Base page:
```go
templ Layout(user string) {
  <!DOCTYPE html>
  <html>
    <head>
      <title>StratusShell - {user}</title>
      <script src="https://unpkg.com/htmx.org@1.9.10"></script>
      <link rel="stylesheet" href="/static/styles.css"/>
    </head>
    <body>
      @Menubar()
      <div id="terminal-container" hx-get="/api/layout" hx-trigger="load">
        <!-- Terminals loaded here -->
      </div>
    </body>
  </html>
}
```

**2. menubar.templ** - Top controls:
```go
templ Menubar() {
  <nav class="menubar">
    <div class="dropdown">
      <button>Layout ▾</button>
      <div class="menu">
        <a hx-post="/api/layout/horizontal" hx-target="#terminal-container">Horizontal Split</a>
        <a hx-post="/api/layout/vertical" hx-target="#terminal-container">Vertical Split</a>
        <a hx-post="/api/layout/grid" hx-target="#terminal-container">Grid (2x2)</a>
      </div>
    </div>

    <div class="dropdown">
      <button>Config ▾</button>
      <div class="menu">
        <a hx-get="/api/config/modal" hx-target="#modal">Preferences...</a>
        <a hx-post="/api/terminals/add" hx-target="#terminal-container">Add Terminal</a>
      </div>
    </div>

    <div class="dropdown">
      <button>Sessions ▾</button>
      <div class="menu">
        <a hx-post="/api/session/save" hx-target="#modal">Save Session...</a>
        <a hx-get="/api/session/list" hx-target="#modal">Load Session...</a>
      </div>
    </div>
  </nav>
}
```

**3. terminal.templ** - Terminal pane:
```go
templ TerminalPane(id int, port int, title string) {
  <div class="terminal-pane" id={"terminal-" + fmt.Sprint(id)}>
    <div class="terminal-header">
      <span contenteditable hx-post={"/api/terminal/" + fmt.Sprint(id) + "/rename"}>
        {title}
      </span>
      <button hx-delete={"/api/terminal/" + fmt.Sprint(id)}
              hx-target={"#terminal-" + fmt.Sprint(id)}
              hx-swap="outerHTML">×</button>
    </div>
    <iframe src={"http://localhost:" + fmt.Sprint(port)}
            class="terminal-frame"></iframe>
  </div>
}
```

### HTMX Interaction Patterns

| User Action | HTMX Request | Server Response | Result |
|-------------|--------------|-----------------|--------|
| Change layout to grid | `POST /api/layout/grid` | Updated terminal container HTML | 4 terminals in 2x2 grid |
| Add terminal | `POST /api/terminals/add` | New terminal pane HTML | Additional terminal appears |
| Remove terminal | `DELETE /api/terminal/{id}` | Empty content | Terminal removed, GoTTY killed |
| Rename terminal | `POST /api/terminal/{id}/rename` | Updated header HTML | Terminal title changes |
| Save session | `POST /api/session/save` | Confirmation modal | Session saved to DB |
| Load session | `POST /api/session/load/{id}` | Full terminal container | All terminals restored |
| Open preferences | `GET /api/config/modal` | Modal HTML | Preferences dialog opens |

## Data Model (SQLite)

**Database Location:** `~/.stratusshell/data.db` (owned by provisioned user)

### Schema

```sql
-- User preferences (themes, fonts, keybindings)
CREATE TABLE preferences (
    id INTEGER PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Saved sessions (named layouts)
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Terminal configurations within a session
CREATE TABLE session_terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    terminal_index INTEGER NOT NULL,
    title TEXT NOT NULL,
    shell TEXT DEFAULT '/bin/bash',
    working_dir TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Current active layout (singleton, always 1 row)
CREATE TABLE active_layout (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    layout_type TEXT NOT NULL CHECK (layout_type IN ('horizontal', 'vertical', 'grid')),
    terminal_count INTEGER NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Currently running terminals
CREATE TABLE active_terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    port INTEGER UNIQUE NOT NULL,
    title TEXT NOT NULL,
    pid INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Data Flow Examples

**Server Startup:**
1. Load `active_layout` → restore previous layout type
2. Load `active_terminals` → restore terminal states
3. Spawn GoTTY instances for each active terminal
4. Render UI with restored state

**Layout Change (horizontal → grid):**
1. User clicks "Grid (2x2)"
2. HTMX: `POST /api/layout/grid`
3. Server: Update `active_layout.layout_type = 'grid'`, `terminal_count = 4`
4. Server: Spawn 2 additional GoTTY instances (had 2, need 4)
5. Server: Update `active_terminals` table
6. Server: Render terminal container with 4 terminals in grid
7. Response: HTMX swaps `#terminal-container`

**Save Session:**
1. User clicks "Save Session..." → modal appears
2. User enters name: "Python Development"
3. HTMX: `POST /api/session/save` with name
4. Server: Create `sessions` row
5. Server: Copy all `active_terminals` → `session_terminals`
6. Response: Success modal

**Load Session:**
1. User clicks "Load Session..." → list appears
2. User selects "Python Development"
3. HTMX: `POST /api/session/load/{id}`
4. Server: Kill all current GoTTY instances
5. Server: Read `session_terminals` for session
6. Server: Spawn new GoTTY instances with saved config
7. Server: Update `active_layout` and `active_terminals`
8. Response: Rebuilt terminal container

## GoTTY Terminal Management

### Dynamic Port Allocation

```go
type PortPool struct {
    minPort int  // 8081
    maxPort int  // 8181 (100 terminals max)
    used    map[int]bool
    mu      sync.Mutex
}
```

**Port range:** 8081-8181 (100 terminal capacity)

### Terminal Lifecycle

**Spawn:**
1. Allocate port from pool
2. Create context with cancel function
3. Execute `gotty --port {port} --permit-write --reconnect /bin/bash`
4. Store Terminal struct with PID, port, cancel func
5. Save to `active_terminals` table
6. Monitor process in goroutine

**Kill:**
1. Cancel context (stops GoTTY process)
2. Wait for process exit
3. Release port back to pool
4. Delete from `active_terminals` table
5. Clean up Terminal struct

**Monitor:**
- Goroutine per terminal watches process
- If process dies unexpectedly: clean up state automatically
- Prevents zombie processes and port leaks

**Restore on startup:**
1. Read `active_terminals` from DB
2. Spawn GoTTY for each terminal
3. If spawn fails: log error, continue (don't block startup)

### Layout Management

```go
func (tm *TerminalManager) ApplyLayout(layoutType string) error {
    targetCount := map[string]int{
        "horizontal": 2,
        "vertical":   2,
        "grid":       4,
    }[layoutType]

    currentCount := len(tm.terminals)

    if targetCount > currentCount {
        // Spawn additional terminals
    } else if targetCount < currentCount {
        // Kill excess terminals
    }

    tm.db.UpdateActiveLayout(layoutType, targetCount)
}
```

**Graceful Shutdown:**
1. Iterate all terminals
2. Cancel each context
3. Wait for processes to exit
4. Release all ports
5. State remains in DB for next startup

## Tool Installation Strategy

### Package Manager Detection

**Supported:**
- APT (Debian/Ubuntu)
- YUM (RHEL/CentOS)
- DNF (Fedora)
- Pacman (Arch)

**Detection logic:** Check for package manager binaries, return appropriate enum

### Installation Phases

**Phase 1: Base System Tools**
```yaml
base_tools:
  - git
  - curl
  - wget
  - build-essential  # or 'Development Tools' for yum
  - ca-certificates
  - gnupg
```

**Phase 2: Cloud CLIs**
- **AWS CLI:** Download from AWS, install to /usr/local/bin
- **gcloud:** Add Google Cloud apt/yum repo, install via package manager
- **kubectl:** Download from dl.k8s.io
- **Docker:** Add Docker repo, install, add user to docker group
- **Terraform:** Add HashiCorp repo, install

**Phase 3: Language Toolchains**

**Go:**
1. Download latest Go binary from golang.org
2. Extract to `/usr/local/go`
3. Add to PATH in user's shell RC file
4. Install gvm (Go Version Manager) to user's home
5. Install meta packages:
   - `golangci-lint` (linting)
   - `gopls` (LSP server)
   - `delve` (debugger)
   - `air` (hot reload)

**Node:**
1. Install nvm (Node Version Manager) to user's home
2. Use nvm to install latest LTS Node
3. Install global meta packages:
   - `pnpm` (fast package manager)
   - `yarn` (alternative package manager)
   - `typescript` (TypeScript compiler)
   - `ts-node` (TypeScript REPL)
   - `prettier` (code formatter)
   - `eslint` (linter)

**Phase 4: Shell Environment**
1. Install zsh (if configured)
2. Install oh-my-zsh with plugins (git, docker, kubectl)
3. Set default shell: `chsh -s /bin/zsh {user}`
4. Install tmux with basic config
5. Create `~/.stratusshell/env.sh` with PATH additions

### Configuration File

```yaml
# configs/default.yaml
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

### Error Handling

**Critical failures (rollback):**
- User creation fails
- Sudo configuration fails

**Non-critical failures (log and continue):**
- Individual tool installation fails
- Language toolchain setup fails

**Rollback logic:**
```go
type ProvisionState struct {
    UserCreated    bool
    InstalledTools []string
}

func (p *Provisioner) rollback(state *ProvisionState) {
    if state.UserCreated {
        exec.Command("userdel", "-r", state.Username).Run()
    }
}
```

**Summary report:**
```
Provisioning Summary for user 'developer':
✓ User created
✓ Passwordless sudo configured
✓ Base tools installed (5/5)
✓ Cloud CLIs installed (4/5) - gcloud failed (network timeout)
✓ Go toolchain installed
✓ Node toolchain installed
✗ Shell environment setup failed (oh-my-zsh download failed)

Status: Provisioning completed with warnings
```

## Error Handling & Testing

### Error Handling Strategy

**Server Errors:**
- Return HTMX-friendly error responses
- Use `HX-Retarget` header to show error toasts
- Log detailed errors server-side, show user-friendly messages

**Terminal Spawning:**
- Retry logic (3 attempts with exponential backoff)
- If all retries fail: return error to user
- Monitor goroutine detects unexpected crashes

**Database Errors:**
- Retry on "database is locked" (SQLite concurrency)
- Transaction support for multi-step operations
- Graceful degradation if DB unavailable (in-memory fallback)

### Testing Strategy

**Unit Tests:**
- Port pool allocation/release logic
- Database CRUD operations (in-memory SQLite)
- Terminal lifecycle state machine
- Configuration parsing

**Integration Tests:**
- Init command (requires Docker or VM)
- Service installation (systemd required)
- End-to-end terminal spawn/kill

**Manual Testing Checklist:**
```
Init Command:
□ Creates user successfully
□ Passwordless sudo works
□ All tools installed and in PATH
□ Language toolchains functional (go version, node -v)

Serve Command:
□ Server starts on specified port
□ Terminals spawn and are interactive
□ Can execute commands in all terminals
□ State persists to SQLite

Install Command:
□ Systemd service created
□ Service starts automatically
□ Service survives reboot
□ Service restarts on crash

Web UI:
□ Layout changes work (horizontal/vertical/grid)
□ Add/remove terminals dynamically
□ Save/load sessions
□ Preferences persist (theme, font)
□ Terminal renaming works
```

## Build & Deployment

### Build Process

```bash
# 1. Generate templ files
templ generate

# 2. Build binary
go build -o stratusshell main.go

# 3. Install to system
sudo cp stratusshell /usr/local/bin/
sudo chmod +x /usr/local/bin/stratusshell

# 4. Copy default config
sudo mkdir -p /etc/stratusshell
sudo cp configs/default.yaml /etc/stratusshell/
```

### Makefile

```makefile
.PHONY: generate build install test clean

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
	INTEGRATION_TESTS=1 go test ./test/integration/...

clean:
	rm -f stratusshell
	find . -name "*_templ.go" -delete
```

### Dependencies

```go
require (
    github.com/spf13/cobra v1.8.0
    github.com/a-h/templ v0.2.543
    github.com/sorenisanerd/gotty v1.6.0
    github.com/mattn/go-sqlite3 v1.14.19
    gopkg.in/yaml.v3 v3.0.1
)
```

## Migration from Current Codebase

### Migration Strategy

**Phase 1: CLI Structure**
- Add Cobra framework
- Move existing `main()` logic to `cmd/serve.go`
- Keep current functionality working
- Add stub `cmd/init.go` and `cmd/install.go`

**Phase 2: Provisioning**
- Implement `internal/provision/` package
- Add user creation logic
- Add tool installation logic
- Test `init` command independently

**Phase 3: Web UI Enhancement**
- Replace static HTML with templ components
- Add HTMX menubar and dropdowns
- Implement dynamic terminal management
- Keep GoTTY integration working

**Phase 4: Persistence**
- Add SQLite database
- Implement session save/restore
- Add preferences storage
- State restoration on startup

**Phase 5: Service Installation**
- Implement systemd service generation
- Test service lifecycle
- Verify restart behavior

### Backwards Compatibility

```bash
# Old way still works (run serve with defaults)
./stratusshell

# New way
stratusshell serve --port=8080
```

### Git Workflow

Development in feature branches:
```bash
git checkout -b feature/cobra-cli
git checkout -b feature/provisioning
git checkout -b feature/htmx-ui
git checkout -b feature/sqlite-persistence
git checkout -b feature/systemd-service
```

Each feature can be developed, tested, and merged independently.

## Open Questions & Future Enhancements

### Phase 2 Possibilities (Future)
- **Terminal history logging:** Capture all terminal output to SQLite for search
- **Multi-user mode:** Authentication, isolated sessions per user
- **Terminal sharing:** WebRTC-based collaborative terminals
- **Cloud integration:** Auto-provision on AWS/GCP/Azure instances
- **Dotfiles management:** Import/export dotfile templates
- **Resource monitoring:** CPU/memory/disk usage in menubar

### Technical Debt to Address
- GoTTY dependency: Consider replacing with native Go PTY + xterm.js for more control
- Port pool: Static range, could support dynamic discovery
- Single SQLite file: May want separate DBs for sessions vs preferences at scale

---

## Appendix: Command Reference

```bash
# Provision user with full dev environment
sudo stratusshell init --user=developer

# Test web UI manually
stratusshell serve --user=developer --port=8080

# Install as systemd service
sudo stratusshell install --user=developer --port=8080

# Service management
sudo systemctl status stratusshell-developer
sudo systemctl restart stratusshell-developer
sudo systemctl stop stratusshell-developer

# View logs
sudo journalctl -u stratusshell-developer -f
```
