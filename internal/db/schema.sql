-- User preferences
-- Note: The 'id' column uses INTEGER PRIMARY KEY, which in SQLite is an alias for ROWID
-- and provides auto-incrementing behavior by default. Preferences are accessed via the
-- unique 'key' field, not by 'id'. The 'id' field is present for relational integrity
-- (e.g., if foreign keys reference this table); if not needed, it can be removed.
CREATE TABLE IF NOT EXISTS preferences (
    id INTEGER PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Saved sessions
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Terminal configurations within a session
CREATE TABLE IF NOT EXISTS session_terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    terminal_index INTEGER NOT NULL,
    title TEXT NOT NULL,
    shell TEXT DEFAULT '/bin/bash',
    working_dir TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Current active layout (singleton)
CREATE TABLE IF NOT EXISTS active_layout (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    layout_type TEXT NOT NULL CHECK (layout_type IN ('horizontal', 'vertical', 'grid')),
    terminal_count INTEGER NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Active terminals (current running state)
CREATE TABLE IF NOT EXISTS active_terminals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    port INTEGER UNIQUE NOT NULL,
    title TEXT NOT NULL,
    pid INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
