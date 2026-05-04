package session

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SessionStatus represents the state of a session
type SessionStatus string

const (
	StatusRunning  SessionStatus = "running"
	StatusComplete SessionStatus = "complete"
	StatusFailed   SessionStatus = "failed"
	StatusPaused   SessionStatus = "paused"
	StatusZombie   SessionStatus = "zombie"
)

// Session represents a project workspace session with persisted state
type Session struct {
	ID           string        `json:"sessionId"`
	Name         string        `json:"projectName"`
	Status       SessionStatus `json:"status"`
	Goal         string        `json:"goal"`
	Workspace    string        `json:"workspace"`
	GraphJSON    string        `json:"graphJson,omitempty"`
	ChatLog      string        `json:"chatLog,omitempty"`
	Snapshot     string        `json:"snapshot,omitempty"`
	HeartbeatAt  time.Time     `json:"heartbeatAt,omitempty"`
	CreatedAt    time.Time     `json:"createdAt"`
	LastActiveAt time.Time     `json:"lastActive"`
}

// SQLite schema initialization
const schemaSQL = `
CREATE TABLE IF NOT EXISTS sessions (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'running',
	goal TEXT NOT NULL DEFAULT '',
	workspace_path TEXT NOT NULL,
	graph_json TEXT DEFAULT '{}',
	chat_log TEXT DEFAULT '[]',
	snapshot TEXT DEFAULT '',
	heartbeat_at DATETIME,
	created_at DATETIME NOT NULL,
	last_active_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);
CREATE INDEX IF NOT EXISTS idx_sessions_heartbeat_at ON sessions(heartbeat_at);
`

// initDB initializes the SQLite database and creates tables if they don't exist
func initDB(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	if err != nil {
		return err
	}
	// Run migrations for existing databases that may be missing columns
	return migrateSchema(db)
}

// migrateSchema adds missing columns to existing databases (e.g. from older versions)
func migrateSchema(db *sql.DB) error {
	// Check and add missing columns that may not exist in older databases
	migrations := []struct {
		col  string
		typ  string
		def  string
	}{
		{"snapshot", "TEXT", "DEFAULT ''"},
		{"heartbeat_at", "DATETIME", ""},
	}

	for _, m := range migrations {
		var count int
		if err := db.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name=?",
			m.col,
		).Scan(&count); err != nil {
			return fmt.Errorf("failed to check column %s: %w", m.col, err)
		}
		if count == 0 {
			sql := fmt.Sprintf("ALTER TABLE sessions ADD COLUMN %s %s %s", m.col, m.typ, m.def)
			if _, err := db.Exec(sql); err != nil {
				return fmt.Errorf("failed to add column %s: %w", m.col, err)
			}
		}
	}

	// Ensure indexes exist (CREATE INDEX IF NOT EXISTS is safe on existing indexes)
	_, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
		CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at);
		CREATE INDEX IF NOT EXISTS idx_sessions_heartbeat_at ON sessions(heartbeat_at);
	`)
	return err
}

// openDB opens a SQLite database connection
func openDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session database: %w", err)
	}
	// Enable WAL mode for better concurrent performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}
	// Set busy timeout so concurrent writers retry instead of failing immediately
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}
	if err := initDB(db); err != nil {
		return nil, fmt.Errorf("failed to initialize session schema: %w", err)
	}
	return db, nil
}
