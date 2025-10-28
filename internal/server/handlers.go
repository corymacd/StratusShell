package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/corymacd/StratusShell/internal/audit"
	"github.com/corymacd/StratusShell/internal/ui"
	"github.com/corymacd/StratusShell/internal/validation"
)

// getActor extracts the authenticated user from request context
func (s *Server) getActor(r *http.Request) string {
	if user, ok := r.Context().Value(userContextKey).(string); ok {
		return user
	}
	return "unknown"
}

func (s *Server) handleGetLayout(w http.ResponseWriter, r *http.Request) {
	terminals := s.terminalManager.GetTerminals()
	layout, err := s.db.GetActiveLayout(r.Context())
	if err != nil {
		s.handleError(w, r, err, "Failed to get layout")
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
	actor := s.getActor(r)

	// Validate layout type
	if err := validation.ValidateLayoutType(layoutType); err != nil {
		s.auditLogger.LogLayoutChange(actor, layoutType, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Invalid layout type")
		return
	}

	if err := s.terminalManager.ApplyLayout(layoutType); err != nil {
		s.auditLogger.LogLayoutChange(actor, layoutType, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Failed to apply layout")
		return
	}

	s.auditLogger.LogLayoutChange(actor, layoutType, audit.OutcomeSuccess, nil)
	s.handleGetLayout(w, r)
}

func (s *Server) handleAddTerminal(w http.ResponseWriter, r *http.Request) {
	actor := s.getActor(r)
	terminals := s.terminalManager.GetTerminals()
	title := fmt.Sprintf("Terminal %d", len(terminals)+1)

	terminal, err := s.terminalManager.SpawnTerminal(title, "/bin/bash", "")
	if err != nil {
		s.auditLogger.LogTerminalSpawn(actor, -1, title, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Failed to add terminal")
		return
	}

	s.auditLogger.LogTerminalSpawn(actor, terminal.ID, title, audit.OutcomeSuccess, nil)
	s.handleGetLayout(w, r)
}

func (s *Server) handleTerminalAction(w http.ResponseWriter, r *http.Request) {
	actor := s.getActor(r)

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

	// Validate terminal ID
	if err := validation.ValidateTerminalID(id); err != nil {
		s.handleError(w, r, err, "Invalid terminal ID")
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := s.terminalManager.KillTerminal(id); err != nil {
			s.auditLogger.LogTerminalKill(actor, id, audit.OutcomeFailure, err)
			s.handleError(w, r, err, "Failed to delete terminal")
			return
		}
		s.auditLogger.LogTerminalKill(actor, id, audit.OutcomeSuccess, nil)
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		if len(parts) > 1 && parts[1] == "rename" {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				s.handleError(w, r, err, "Failed to parse form")
				return
			}
			newTitle := validation.SanitizeString(r.FormValue("title"))

			// Validate title
			if err := validation.ValidateTerminalTitle(newTitle); err != nil {
				s.auditLogger.LogTerminalRename(actor, id, "", newTitle, audit.OutcomeFailure, err)
				s.handleError(w, r, err, "Invalid terminal title")
				return
			}

			terminal, ok := s.terminalManager.GetTerminal(id)
			if !ok {
				http.Error(w, "Terminal not found", http.StatusNotFound)
				return
			}

			oldTitle := terminal.Title
			terminal.Title = newTitle

			// Persist title change to database
			if terminal.DBID > 0 {
				if err := s.db.UpdateActiveTerminalTitle(r.Context(), terminal.DBID, newTitle); err != nil {
					log.Printf("Warning: failed to update terminal title in db: %v", err)
				}
			}

			s.auditLogger.LogTerminalRename(actor, id, oldTitle, newTitle, audit.OutcomeSuccess, nil)
			w.WriteHeader(http.StatusOK)
		}
	}
}

func (s *Server) handleSaveSessionModal(w http.ResponseWriter, r *http.Request) {
	ui.SaveSessionModal().Render(r.Context(), w)
}

func (s *Server) handleSaveSession(w http.ResponseWriter, r *http.Request) {
	actor := s.getActor(r)
	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, err, "Failed to parse form")
		return
	}
	name := validation.SanitizeString(r.FormValue("name"))
	description := validation.SanitizeString(r.FormValue("description"))

	// Validate inputs
	if err := validation.ValidateSessionName(name); err != nil {
		s.auditLogger.LogSessionCreate(actor, -1, name, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Invalid session name")
		return
	}

	if err := validation.ValidateSessionDescription(description); err != nil {
		s.auditLogger.LogSessionCreate(actor, -1, name, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Invalid session description")
		return
	}

	// Create session
	sessionID, err := s.db.CreateSession(r.Context(), name, description)
	if err != nil {
		s.auditLogger.LogSessionCreate(actor, -1, name, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Failed to save session")
		return
	}

	// Save all current terminals
	terminals := s.terminalManager.GetTerminals()
	for i, t := range terminals {
		if err := s.db.SaveSessionTerminal(r.Context(), sessionID, i, t.Title, t.Shell, t.WorkingDir); err != nil {
			log.Printf("Warning: failed to save terminal %d: %v", t.ID, err)
		}
	}

	s.auditLogger.LogSessionCreate(actor, sessionID, name, audit.OutcomeSuccess, nil)
	ui.SuccessMessage("Session saved successfully").Render(r.Context(), w)
}

func (s *Server) handleListSessionsModal(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.db.GetAllSessions(r.Context())
	if err != nil {
		s.handleError(w, r, err, "Failed to load sessions")
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
	actor := s.getActor(r)

	// Extract session ID
	path := strings.TrimPrefix(r.URL.Path, "/api/session/load/")
	sessionID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Validate session ID
	if err := validation.ValidateSessionID(sessionID); err != nil {
		s.auditLogger.LogSessionLoad(actor, sessionID, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Invalid session ID")
		return
	}

	// Get session terminals
	sessionTerminals, err := s.db.GetSessionTerminals(r.Context(), sessionID)
	if err != nil {
		s.auditLogger.LogSessionLoad(actor, sessionID, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Failed to load session")
		return
	}

	// Store old terminals to be killed later
	oldTerminals := s.terminalManager.GetTerminals()

	// Spawn new terminals from session first (transactional approach)
	newTerminals := make([]*Terminal, 0, len(sessionTerminals))
	for _, st := range sessionTerminals {
		term, err := s.terminalManager.SpawnTerminal(st.Title, st.Shell, st.WorkingDir)
		if err != nil {
			log.Printf("Error: failed to spawn terminal for session: %v", err)
			// Rollback: clean up any terminals that were successfully spawned
			for _, t := range newTerminals {
				s.terminalManager.KillTerminal(t.ID)
			}
			s.auditLogger.LogSessionLoad(actor, sessionID, audit.OutcomeFailure, err)
			s.handleError(w, r, err, "Failed to spawn new terminals for session")
			return
		}
		newTerminals = append(newTerminals, term)
	}

	// Now that new terminals are ready, kill old ones
	for _, t := range oldTerminals {
		if err := s.terminalManager.KillTerminal(t.ID); err != nil {
			log.Printf("Warning: failed to kill old terminal %d: %v", t.ID, err)
		}
	}

	// Update layout: restore session's original layout type if available
	sessionLayoutType, err := s.db.GetSessionLayoutType(sessionID)
	layoutType := sessionLayoutType
	if err != nil || layoutType == "" {
		// Fallback to previous logic if not found
		layoutType = "horizontal"
		if len(sessionTerminals) > 2 {
			layoutType = "grid"
		}
	}
	s.db.UpdateActiveLayout(r.Context(), layoutType, len(sessionTerminals))

	s.auditLogger.LogSessionLoad(actor, sessionID, audit.OutcomeSuccess, nil)
	s.handleGetLayout(w, r)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Get current system user
	user := r.URL.Query().Get("user")
	if user == "" {
		user = "anonymous"
	}

	// Create session
	token, err := s.authManager.CreateSession(user)
	if err != nil {
		s.auditLogger.LogAuthLogin(user, audit.OutcomeFailure, err)
		s.handleError(w, r, err, "Failed to create session")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})

	s.auditLogger.LogAuthLogin(user, audit.OutcomeSuccess, nil)

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	actor := s.getActor(r)

	// Get session cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		s.authManager.DeleteSession(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	s.auditLogger.LogAuthLogout(actor, audit.OutcomeSuccess)

	// Redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) handleError(w http.ResponseWriter, r *http.Request, err error, userMsg string) {
	log.Printf("Error: %v", err)
	w.Header().Set("HX-Retarget", "#modal")
	w.Header().Set("HX-Reswap", "innerHTML")
	w.WriteHeader(http.StatusInternalServerError)
	ui.ErrorToast(userMsg).Render(r.Context(), w)
}
