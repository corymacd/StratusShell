package db

import (
	"database/sql"
	"fmt"
)

func (db *DB) GetPreference(key string) (string, error) {
	var value string
	err := db.conn.QueryRow("SELECT value FROM preferences WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (db *DB) SetPreference(key, value string) error {
	_, err := db.conn.Exec(`
		INSERT INTO preferences (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, updated_at = CURRENT_TIMESTAMP
	`, key, value, value)
	return err
}

func (db *DB) GetAllPreferences() (map[string]string, error) {
	rows, err := db.conn.Query("SELECT key, value FROM preferences")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		prefs[key] = value
	}
	return prefs, rows.Err()
}
