package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthStatus represents the health check response
type HealthStatus struct {
	Status            string    `json:"status"`
	Timestamp         time.Time `json:"timestamp"`
	ActiveTerminals   int       `json:"active_terminals"`
	DatabaseConnected bool      `json:"database_connected"`
	UptimeSeconds     int64     `json:"uptime_seconds"`
}

var serverStartTime = time.Now()

// handleHealth returns the health status of the server
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:            "healthy",
		Timestamp:         time.Now(),
		ActiveTerminals:   len(s.terminalManager.GetTerminals()),
		DatabaseConnected: true,
		UptimeSeconds:     int64(time.Since(serverStartTime).Seconds()),
	}

	// Check database connection
	if err := s.db.Ping(); err != nil {
		status.Status = "unhealthy"
		status.DatabaseConnected = false
	}

	// Set status code based on health
	statusCode := http.StatusOK
	if status.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(status)
}

// MetricsStatus represents basic metrics
type MetricsStatus struct {
	TotalTerminalsSpawned int       `json:"total_terminals_spawned"`
	ActiveTerminals       int       `json:"active_terminals"`
	TotalSessions         int       `json:"total_sessions"`
	UptimeSeconds         int64     `json:"uptime_seconds"`
	Timestamp             time.Time `json:"timestamp"`
}

// handleMetrics returns basic metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Get total sessions from database
	sessions, err := s.db.GetAllSessions(r.Context())
	totalSessions := 0
	if err == nil {
		totalSessions = len(sessions)
	}

	metrics := MetricsStatus{
		TotalTerminalsSpawned: s.terminalManager.GetNextID() - 1, // nextID starts at 1
		ActiveTerminals:       len(s.terminalManager.GetTerminals()),
		TotalSessions:         totalSessions,
		UptimeSeconds:         int64(time.Since(serverStartTime).Seconds()),
		Timestamp:             time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}
