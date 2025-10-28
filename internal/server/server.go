package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/corymacd/StratusShell/internal/audit"
	"github.com/corymacd/StratusShell/internal/db"
	"github.com/corymacd/StratusShell/internal/middleware"
	"github.com/corymacd/StratusShell/internal/ui"
)

type Server struct {
	port            int
	db              *db.DB
	terminalManager *TerminalManager
	authManager     *AuthManager
	auditLogger     *audit.Logger
	rateLimiter     *middleware.RateLimiter
	csrfProtection  *middleware.CSRFProtection
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

	// Create auth manager
	am := NewAuthManager()

	// Create audit logger
	al := audit.NewLogger()

	// Create rate limiter: 100 requests per minute per IP
	rl := middleware.NewRateLimiter(100, time.Minute)

	// Create CSRF protection
	csrf := middleware.NewCSRFProtection()

	s := &Server{
		port:            port,
		db:              database,
		terminalManager: tm,
		authManager:     am,
		auditLogger:     al,
		rateLimiter:     rl,
		csrfProtection:  csrf,
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
	// Static files - only serve from static/ directory
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Health and metrics - public
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/metrics", s.handleMetrics)

	// Auth routes - public with rate limiting
	mux.HandleFunc("/login", s.rateLimiter.Limit(s.handleLogin))
	mux.HandleFunc("/logout", s.rateLimiter.Limit(s.handleLogout))

	// Terminal proxy - requires auth + rate limiting
	mux.HandleFunc("/term/", s.rateLimiter.Limit(s.AuthMiddleware(s.handleTerminalProxy)))

	// Main page - requires auth + rate limiting
	mux.HandleFunc("/", s.rateLimiter.Limit(s.AuthMiddleware(s.handleIndex)))

	// API routes - all require auth + rate limiting + CSRF protection for state changes
	mux.HandleFunc("/api/layout", s.rateLimiter.Limit(s.AuthMiddleware(s.handleGetLayout)))
	mux.HandleFunc("/api/layout/horizontal", s.rateLimiter.Limit(s.AuthMiddleware(s.csrfProtection.Protect(s.handleLayoutHorizontal))))
	mux.HandleFunc("/api/layout/vertical", s.rateLimiter.Limit(s.AuthMiddleware(s.csrfProtection.Protect(s.handleLayoutVertical))))
	mux.HandleFunc("/api/layout/grid", s.rateLimiter.Limit(s.AuthMiddleware(s.csrfProtection.Protect(s.handleLayoutGrid))))

	mux.HandleFunc("/api/terminals/add", s.rateLimiter.Limit(s.AuthMiddleware(s.csrfProtection.Protect(s.handleAddTerminal))))
	mux.HandleFunc("/api/terminal/", s.rateLimiter.Limit(s.AuthMiddleware(s.csrfProtection.Protect(s.handleTerminalAction))))

	mux.HandleFunc("/api/session/save-modal", s.rateLimiter.Limit(s.AuthMiddleware(s.handleSaveSessionModal)))
	mux.HandleFunc("/api/session/save", s.rateLimiter.Limit(s.AuthMiddleware(s.csrfProtection.Protect(s.handleSaveSession))))
	mux.HandleFunc("/api/session/list-modal", s.rateLimiter.Limit(s.AuthMiddleware(s.handleListSessionsModal)))
	mux.HandleFunc("/api/session/load/", s.rateLimiter.Limit(s.AuthMiddleware(s.csrfProtection.Protect(s.handleLoadSession))))
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
	ctx := context.Background()

	// Clean up stale terminal records from previous crashes
	if err := s.db.ClearActiveTerminals(ctx); err != nil {
		log.Printf("Warning: failed to clear stale terminal records: %v", err)
	}

	layout, err := s.db.GetActiveLayout(ctx)
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

func (s *Server) handleTerminalProxy(w http.ResponseWriter, r *http.Request) {
	// Extract port from path: /term/{port}/...
	path := strings.TrimPrefix(r.URL.Path, "/term/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid terminal port", http.StatusBadRequest)
		return
	}

	port, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid terminal port", http.StatusBadRequest)
		return
	}

	// Find terminal by port to get credential
	var credential string
	terminals := s.terminalManager.GetTerminals()
	for _, t := range terminals {
		if t.Port == port {
			credential = t.Credential
			break
		}
	}

	if credential == "" {
		http.Error(w, "Terminal not found", http.StatusNotFound)
		return
	}

	// Create reverse proxy to localhost:{port}
	target, err := url.Parse(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		http.Error(w, "Invalid proxy target", http.StatusInternalServerError)
		return
	}

	// Strip /term/{port} prefix and proxy to target
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Modify request to add Basic Auth header
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Add Basic Auth using the terminal's credential
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(credential)))
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/term/%d", port))
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	proxy.ServeHTTP(w, r)
}
