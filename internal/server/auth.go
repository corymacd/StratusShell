package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"sync"
	"time"
)

type contextKey string

const userContextKey contextKey = "user"

// Session represents an authenticated session
type Session struct {
	Token     string
	User      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// AuthManager manages authentication sessions
type AuthManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewAuthManager() *AuthManager {
	am := &AuthManager{
		sessions: make(map[string]*Session),
	}
	// Start cleanup goroutine
	go am.cleanupExpired()
	return am
}

func (am *AuthManager) generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (am *AuthManager) CreateSession(user string) (string, error) {
	token, err := am.generateToken()
	if err != nil {
		return "", err
	}

	session := &Session{
		Token:     token,
		User:      user,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	am.mu.Lock()
	am.sessions[token] = session
	am.mu.Unlock()

	return token, nil
}

func (am *AuthManager) ValidateSession(token string) (*Session, bool) {
	am.mu.RLock()
	session, exists := am.sessions[token]
	am.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(session.ExpiresAt) {
		am.mu.Lock()
		delete(am.sessions, token)
		am.mu.Unlock()
		return nil, false
	}

	return session, true
}

func (am *AuthManager) DeleteSession(token string) {
	am.mu.Lock()
	delete(am.sessions, token)
	am.mu.Unlock()
}

func (am *AuthManager) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		am.mu.Lock()
		now := time.Now()
		for token, session := range am.sessions {
			if now.After(session.ExpiresAt) {
				delete(am.sessions, token)
			}
		}
		am.mu.Unlock()
	}
}

// AuthMiddleware checks for valid authentication
func (s *Server) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// Redirect to login with return URL
			http.Redirect(w, r, "/login?user="+r.URL.Query().Get("user"), http.StatusSeeOther)
			return
		}

		// Validate session
		session, valid := s.authManager.ValidateSession(cookie.Value)
		if !valid {
			// Redirect to login with return URL
			http.Redirect(w, r, "/login?user="+r.URL.Query().Get("user"), http.StatusSeeOther)
			return
		}

		// Add session info to request context
		log.Printf("Authenticated request: user=%s, path=%s", session.User, r.URL.Path)

		// Add user to context for audit logging
		ctx := context.WithValue(r.Context(), userContextKey, session.User)
		next(w, r.WithContext(ctx))
	}
}
