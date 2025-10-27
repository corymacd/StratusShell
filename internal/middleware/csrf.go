package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

// CSRFProtection implements CSRF token validation
type CSRFProtection struct {
	tokens map[string]time.Time // token -> expiration
	mu     sync.RWMutex
}

// NewCSRFProtection creates a new CSRF protection instance
func NewCSRFProtection() *CSRFProtection {
	csrf := &CSRFProtection{
		tokens: make(map[string]time.Time),
	}

	// Start cleanup goroutine
	go csrf.cleanupTokens()

	return csrf
}

// generateToken creates a new CSRF token
func (csrf *CSRFProtection) generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetToken returns a valid CSRF token for the session
// If a valid token exists in the cookie, it's reused; otherwise a new one is created
func (csrf *CSRFProtection) GetToken(w http.ResponseWriter, r *http.Request) (string, error) {
	// Check if there's already a valid token in cookie
	cookie, err := r.Cookie("csrf_token")
	if err == nil {
		csrf.mu.RLock()
		expiration, exists := csrf.tokens[cookie.Value]
		csrf.mu.RUnlock()

		if exists && time.Now().Before(expiration) {
			return cookie.Value, nil
		}
	}

	// Generate new token
	token, err := csrf.generateToken()
	if err != nil {
		return "", err
	}

	// Store token with 24 hour expiration
	expiration := time.Now().Add(24 * time.Hour)
	csrf.mu.Lock()
	csrf.tokens[token] = expiration
	csrf.mu.Unlock()

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})

	return token, nil
}

// ValidateToken checks if the provided token is valid
func (csrf *CSRFProtection) ValidateToken(token string) bool {
	csrf.mu.RLock()
	defer csrf.mu.RUnlock()

	expiration, exists := csrf.tokens[token]
	if !exists {
		return false
	}

	if time.Now().After(expiration) {
		return false
	}

	return true
}

// cleanupTokens removes expired tokens
func (csrf *CSRFProtection) cleanupTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		csrf.mu.Lock()
		now := time.Now()
		for token, expiration := range csrf.tokens {
			if now.After(expiration) {
				delete(csrf.tokens, token)
			}
		}
		csrf.mu.Unlock()
	}
}

// Protect returns a middleware that validates CSRF tokens for state-changing requests
func (csrf *CSRFProtection) Protect(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only check CSRF for state-changing methods
		if r.Method == http.MethodPost || r.Method == http.MethodPut ||
			r.Method == http.MethodDelete || r.Method == http.MethodPatch {

			// Get token from header or form
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				// Try to get from form data (for regular form submissions)
				token = r.FormValue("csrf_token")
			}
			if token == "" {
				// Try to get from cookie for comparison
				cookie, err := r.Cookie("csrf_token")
				if err != nil {
					http.Error(w, "CSRF token missing", http.StatusForbidden)
					return
				}
				token = cookie.Value
			}

			// Validate token
			if !csrf.ValidateToken(token) {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}
		}

		next(w, r)
	}
}
