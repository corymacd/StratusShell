package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/corymacd/cloud-dev-cli-env/internal/ui"
)

func (s *Server) handleGetLayout(w http.ResponseWriter, r *http.Request) {
	terminals := s.terminalManager.GetTerminals()
	layout, err := s.db.GetActiveLayout()
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
	if err := s.terminalManager.ApplyLayout(layoutType); err != nil {
		s.handleError(w, r, err, "Failed to apply layout")
		return
	}
	s.handleGetLayout(w, r)
}

func (s *Server) handleAddTerminal(w http.ResponseWriter, r *http.Request) {
	terminals := s.terminalManager.GetTerminals()
	title := fmt.Sprintf("Terminal %d", len(terminals)+1)

	_, err := s.terminalManager.SpawnTerminal(title, "/bin/bash", "")
	if err != nil {
		s.handleError(w, r, err, "Failed to add terminal")
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
			s.handleError(w, r, err, "Failed to delete terminal")
			return
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodPost:
		if len(parts) > 1 && parts[1] == "rename" {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				s.handleError(w, r, err, "Failed to parse form")
				return
			}
			newTitle := strings.TrimSpace(r.FormValue("title"))
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

			// Persist title change to database
			if terminal.DBID > 0 {
				if err := s.db.UpdateActiveTerminalTitle(terminal.DBID, newTitle); err != nil {
					log.Printf("Warning: failed to update terminal title in db: %v", err)
				}
			}

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
		s.handleError(w, r, fmt.Errorf("name required"), "Session name is required")
		return
	}

	// Create session
	sessionID, err := s.db.CreateSession(name, description)
	if err != nil {
		s.handleError(w, r, err, "Failed to save session")
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
			s.handleError(w, r, err, "Failed to spawn new terminals for session")
			return
		}
		newTerminals = append(newTerminals, term)
	}

	// Now that new terminals are ready, kill old ones
	for _, t := range oldTerminals {
		s.terminalManager.KillTerminal(t.ID)
	}

	// Update layout
	layoutType := "horizontal"
	if len(sessionTerminals) > 2 {
		layoutType = "grid"
	}
	s.db.UpdateActiveLayout(layoutType, len(sessionTerminals))

	s.handleGetLayout(w, r)
}

func (s *Server) handleError(w http.ResponseWriter, r *http.Request, err error, userMsg string) {
	log.Printf("Error: %v", err)
	w.Header().Set("HX-Retarget", "#modal")
	w.Header().Set("HX-Reswap", "innerHTML")
	w.WriteHeader(http.StatusInternalServerError)
	ui.ErrorToast(userMsg).Render(r.Context(), w)
}
