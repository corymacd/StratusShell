package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/corymacd/StratusShell/internal/db"
)

type Terminal struct {
	ID          int
	DBID        int // Database primary key
	Port        int
	Title       string
	Shell       string
	WorkingDir  string
	Credential  string // GoTTY authentication credential
	GoTTYServer *GoTTYServer
	CreatedAt   time.Time
}

type TerminalManager struct {
	terminals    map[int]*Terminal
	portPool     *PortPool
	db           *db.DB
	mu           sync.RWMutex
	nextID       int
	maxTerminals int
	activeTabID  int // Track the currently active tab
}

func NewTerminalManager(db *db.DB) *TerminalManager {
	return &TerminalManager{
		terminals:    make(map[int]*Terminal),
		portPool:     NewPortPool(0, 0), // Use ephemeral ports
		db:           db,
		nextID:       1,
		maxTerminals: 10, // Maximum 10 concurrent terminals
		activeTabID:  0,  // No active tab initially
	}
}

// generateCredential creates a random credential for GoTTY authentication
func generateCredential() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Format as username:password
	username := "term"
	password := base64.URLEncoding.EncodeToString(b)
	return fmt.Sprintf("%s:%s", username, password), nil
}

func (tm *TerminalManager) SpawnTerminal(title, shell, workingDir string) (*Terminal, error) {
	tm.mu.Lock()
	// Check if we've reached the maximum number of terminals
	if len(tm.terminals) >= tm.maxTerminals {
		tm.mu.Unlock()
		return nil, fmt.Errorf("maximum number of terminals (%d) reached", tm.maxTerminals)
	}
	tm.mu.Unlock()

	port, err := tm.portPool.Allocate()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}

	// Generate authentication credential
	credential, err := generateCredential()
	if err != nil {
		tm.portPool.Release(port)
		return nil, fmt.Errorf("failed to generate credential: %w", err)
	}

	// Create GoTTY server using library
	ctx := context.Background()
	gottyServer, err := NewGoTTYServer(ctx, port, credential, title, shell, workingDir)
	if err != nil {
		tm.portPool.Release(port)
		return nil, fmt.Errorf("failed to start gotty server: %w", err)
	}

	tm.mu.Lock()
	terminalID := tm.nextID
	tm.nextID++
	tm.mu.Unlock()

	terminal := &Terminal{
		ID:          terminalID,
		Port:        port,
		Title:       title,
		Shell:       shell,
		WorkingDir:  workingDir,
		Credential:  credential,
		GoTTYServer: gottyServer,
		CreatedAt:   time.Now(),
	}

	// Save to database (PID is 0 since we're using library, not external process)
	dbID, err := tm.db.SaveActiveTerminal(ctx, terminal.Port, terminal.Title, 0)
	if err != nil {
		log.Printf("Warning: failed to save terminal to db: %v", err)
	} else {
		terminal.DBID = dbID
	}

	tm.mu.Lock()
	tm.terminals[terminal.ID] = terminal
	// Set as active tab if it's the first terminal or no active tab
	if tm.activeTabID == 0 || len(tm.terminals) == 1 {
		tm.activeTabID = terminal.ID
	}
	tm.mu.Unlock()

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
	
	// If we're closing the active tab, switch to another tab
	if tm.activeTabID == id {
		// Find another terminal to make active
		tm.activeTabID = 0
		for _, t := range tm.terminals {
			tm.activeTabID = t.ID
			break
		}
	}
	tm.mu.Unlock()

	// Stop GoTTY server gracefully
	if err := terminal.GoTTYServer.Stop(); err != nil {
		log.Printf("Warning: error stopping GoTTY server: %v", err)
	}

	// Release port
	tm.portPool.Release(terminal.Port)

	// Remove from database using the correct database ID
	if terminal.DBID > 0 {
		if err := tm.db.DeleteActiveTerminal(context.Background(), terminal.DBID); err != nil {
			log.Printf("Warning: failed to delete terminal from db: %v", err)
		}
	}

	return nil
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

func (tm *TerminalManager) GetActiveTabID() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.activeTabID
}

func (tm *TerminalManager) SetActiveTabID(id int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.activeTabID = id
}

func (tm *TerminalManager) GetNextID() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.nextID
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
	if err := tm.db.UpdateActiveLayout(context.Background(), layoutType, targetCount); err != nil {
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
