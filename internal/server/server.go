package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/corymacd/cloud-dev-cli-env/internal/db"
	"github.com/corymacd/cloud-dev-cli-env/internal/ui"
)

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
	mux.Handle("/static/", http.FileServer(http.Dir(".")))

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
