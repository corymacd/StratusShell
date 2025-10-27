package db

import (
	"time"
)

type Session struct {
	ID          int
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SessionTerminal struct {
	ID            int
	SessionID     int
	TerminalIndex int
	Title         string
	Shell         string
	WorkingDir    string
}

func (db *DB) CreateSession(name, description string) (int, error) {
	result, err := db.conn.Exec(`
		INSERT INTO sessions (name, description) VALUES (?, ?)
	`, name, description)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (db *DB) GetSession(id int) (*Session, error) {
	s := &Session{}
	err := db.conn.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM sessions WHERE id = ?
	`, id).Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (db *DB) GetAllSessions() ([]*Session, error) {
	rows, err := db.conn.Query(`
		SELECT id, name, description, created_at, updated_at
		FROM sessions ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		s := &Session{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func (db *DB) SaveSessionTerminal(sessionID, index int, title, shell, workingDir string) error {
	_, err := db.conn.Exec(`
		INSERT INTO session_terminals (session_id, terminal_index, title, shell, working_dir)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, index, title, shell, workingDir)
	return err
}

func (db *DB) GetSessionTerminals(sessionID int) ([]*SessionTerminal, error) {
	rows, err := db.conn.Query(`
		SELECT id, session_id, terminal_index, title, shell, working_dir
		FROM session_terminals WHERE session_id = ? ORDER BY terminal_index
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terminals []*SessionTerminal
	for rows.Next() {
		t := &SessionTerminal{}
		if err := rows.Scan(&t.ID, &t.SessionID, &t.TerminalIndex, &t.Title, &t.Shell, &t.WorkingDir); err != nil {
			return nil, err
		}
		terminals = append(terminals, t)
	}
	return terminals, rows.Err()
}
