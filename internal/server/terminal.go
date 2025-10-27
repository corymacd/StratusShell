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
