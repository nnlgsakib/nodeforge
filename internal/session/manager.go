package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

const maxProjectNameLen = 128

var (
	// ErrWorkspaceExists is returned when the project workspace directory already exists.
	ErrWorkspaceExists = errors.New("workspace directory already exists")
)

// Manager handles session creation and management with SQLite persistence
type Manager struct {
	workspaceRoot string
	db            *sql.DB
	mu            sync.Mutex
}

// NewManager creates a new session manager with SQLite backend.
// Returns an error if the database cannot be initialized.
func NewManager(workspaceRoot string) (*Manager, error) {
	dbPath := filepath.Join(workspaceRoot, ".nforge", "sessions.db")
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open session database: %w", err)
	}

	return &Manager{workspaceRoot: workspaceRoot, db: db}, nil
}

// CreateSessionWithName creates a new session with the specified name
func (m *Manager) CreateSessionWithName(ctx context.Context, name string) (*Session, error) {
	// Sanitize project name
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}
	if len(name) > maxProjectNameLen {
		return nil, fmt.Errorf("project name must not exceed %d characters", maxProjectNameLen)
	}
	if filepath.Base(name) != name {
		return nil, fmt.Errorf("project name must not contain path separators")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate project name
	var exists bool
	if err := m.db.QueryRow("SELECT 1 FROM sessions WHERE name = ?", name).Scan(&exists); err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for duplicate project: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("project %q already exists: %w", name, ErrWorkspaceExists)
	}

	// Create session directory structure per spec: .nforge/sessions/<session-id>/workspace/
	sessionID := generateSessionID()
	projectDir := filepath.Join(m.workspaceRoot, ".nforge", "sessions", sessionID)
	workspaceDir := filepath.Join(projectDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Initialize .nforge workspace metadata
	if err := InitProjectWorkspace(projectDir, name); err != nil {
		if !errors.Is(err, ErrWorkspaceExists) {
			return nil, fmt.Errorf("failed to initialize workspace: %w", err)
		}
		// Workspace already exists — this shouldn't happen since we just created it,
		// but tolerate it to avoid race conditions.
	}

	now := time.Now().UTC()

	sess := &Session{
		ID:           sessionID,
		Name:         name,
		Status:       StatusRunning,
		Goal:         "",
		Workspace:    workspaceDir,
		CreatedAt:    now,
		LastActiveAt: now,
	}

	// Persist to SQLite
	if err := m.saveSession(sess); err != nil {
		return nil, fmt.Errorf("failed to persist session: %w", err)
	}

	return sess, nil
}

// ListSessions returns all sessions ordered by creation time (newest first)
func (m *Manager) ListSessions(ctx context.Context) ([]Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows, err := m.db.QueryContext(ctx,
		"SELECT id, name, status, goal, workspace_path, graph_json, chat_log, created_at, last_active_at FROM sessions ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		err := rows.Scan(&s.ID, &s.Name, &s.Status, &s.Goal, &s.Workspace,
			&s.GraphJSON, &s.ChatLog, &s.CreatedAt, &s.LastActiveAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

// GetSession returns a single session by ID
func (m *Manager) GetSession(ctx context.Context, id string) (*Session, error) {
	if m.db == nil {
		return nil, fmt.Errorf("session database not available")
	}

	var s Session
	err := m.db.QueryRowContext(ctx,
		"SELECT id, name, status, goal, workspace_path, graph_json, chat_log, created_at, last_active_at FROM sessions WHERE id = ?",
		id).Scan(&s.ID, &s.Name, &s.Status, &s.Goal, &s.Workspace,
		&s.GraphJSON, &s.ChatLog, &s.CreatedAt, &s.LastActiveAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &s, nil
}

// UpdateSession updates a session's state (graph, chat, status)
func (m *Manager) UpdateSession(ctx context.Context, sess *Session) error {
	if m.db == nil {
		return fmt.Errorf("session database not available")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sess.LastActiveAt = time.Now().UTC()
	return m.saveSession(sess)
}

// saveSession inserts or updates a session in the database
func (m *Manager) saveSession(sess *Session) error {
	_, err := m.db.Exec(`
		INSERT OR REPLACE INTO sessions (id, name, status, goal, workspace_path, graph_json, chat_log, created_at, last_active_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, sess.ID, sess.Name, sess.Status, sess.Goal, sess.Workspace,
		sess.GraphJSON, sess.ChatLog, sess.CreatedAt, sess.LastActiveAt)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("sess-%s", uuid.New().String())
}

// Close closes the session database connection
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
