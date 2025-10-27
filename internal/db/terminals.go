package db

import (
	"context"
	"time"
)

type ActiveTerminal struct {
	ID        int
	Port      int
	Title     string
	PID       int
	CreatedAt time.Time
}

type ActiveLayout struct {
	LayoutType    string
	TerminalCount int
}

func (db *DB) SaveActiveTerminal(ctx context.Context, port int, title string, pid int) (int, error) {
	result, err := db.conn.ExecContext(ctx, `
		INSERT INTO active_terminals (port, title, pid) VALUES (?, ?, ?)
	`, port, title, pid)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (db *DB) GetActiveTerminals(ctx context.Context) ([]*ActiveTerminal, error) {
	rows, err := db.conn.QueryContext(ctx, `
		SELECT id, port, title, pid, created_at
		FROM active_terminals ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var terminals []*ActiveTerminal
	for rows.Next() {
		t := &ActiveTerminal{}
		if err := rows.Scan(&t.ID, &t.Port, &t.Title, &t.PID, &t.CreatedAt); err != nil {
			return nil, err
		}
		terminals = append(terminals, t)
	}
	return terminals, rows.Err()
}

func (db *DB) UpdateActiveTerminalTitle(ctx context.Context, id int, title string) error {
	_, err := db.conn.ExecContext(ctx, "UPDATE active_terminals SET title = ? WHERE id = ?", title, id)
	return err
}

func (db *DB) DeleteActiveTerminal(ctx context.Context, id int) error {
	_, err := db.conn.ExecContext(ctx, "DELETE FROM active_terminals WHERE id = ?", id)
	return err
}

func (db *DB) ClearActiveTerminals(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, "DELETE FROM active_terminals")
	return err
}

func (db *DB) GetActiveLayout(ctx context.Context) (*ActiveLayout, error) {
	layout := &ActiveLayout{}
	err := db.conn.QueryRowContext(ctx, `
		SELECT layout_type, terminal_count FROM active_layout WHERE id = 1
	`).Scan(&layout.LayoutType, &layout.TerminalCount)
	if err != nil {
		return nil, err
	}
	return layout, nil
}

func (db *DB) UpdateActiveLayout(ctx context.Context, layoutType string, terminalCount int) error {
	_, err := db.conn.ExecContext(ctx, `
		UPDATE active_layout SET layout_type = ?, terminal_count = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, layoutType, terminalCount)
	return err
}
