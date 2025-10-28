package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	conn *sql.DB
	path string
}

// Open opens or creates the SQLite database
func Open(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open database
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn, path: dbPath}

	// Run migrations
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Initialize singleton active_layout if not exists
	if err := db.initializeActiveLayout(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize active layout: %w", err)
	}

	return db, nil
}

func (db *DB) migrate() error {
	_, err := db.conn.Exec(schemaSQL)
	return err
}

func (db *DB) initializeActiveLayout() error {
	// Insert default layout if table is empty
	_, err := db.conn.Exec(`
		INSERT OR IGNORE INTO active_layout (id, layout_type, terminal_count)
		VALUES (1, 'horizontal', 2)
	`)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Ping() error {
	return db.conn.Ping()
}
