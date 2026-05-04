package session

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// UpdateSessionState atomically updates graph, chat, and/or status under one lock.
// This avoids the TOCTOU race of fetch-then-update: each call is a single SQL UPDATE.
func (m *Manager) UpdateSessionState(ctx context.Context, sessionID string, graphJSON, chatLog *string, status *SessionStatus) error {
	if m.db == nil {
		return fmt.Errorf("session database not available")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify session exists
	var exists bool
	if err := m.db.QueryRowContext(ctx, "SELECT 1 FROM sessions WHERE id = ?", sessionID).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("session not found: %s", sessionID)
		}
		return fmt.Errorf("failed to verify session: %w", err)
	}

	now := time.Now().UTC()
	_, err := m.db.ExecContext(ctx,
		"UPDATE sessions SET graph_json = COALESCE(?, graph_json), chat_log = COALESCE(?, chat_log), status = COALESCE(?, status), heartbeat_at = ?, last_active_at = ? WHERE id = ?",
		graphJSON, chatLog, status, now, now, sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}
	return nil
}

// SaveGraphJSON persists the graph state to the session.
// Deprecated: use UpdateSessionState for atomic updates.
func (m *Manager) SaveGraphJSON(ctx context.Context, sessionID, graphJSON string) error {
	return m.UpdateSessionState(ctx, sessionID, &graphJSON, nil, nil)
}

// SaveChatLog persists the chat log to the session.
// Deprecated: use UpdateSessionState for atomic updates.
func (m *Manager) SaveChatLog(ctx context.Context, sessionID, chatLog string) error {
	return m.UpdateSessionState(ctx, sessionID, nil, &chatLog, nil)
}

// UpdateSessionStatus updates the status of a session.
// Deprecated: use UpdateSessionState for atomic updates.
func (m *Manager) UpdateSessionStatus(ctx context.Context, sessionID string, status SessionStatus) error {
	return m.UpdateSessionState(ctx, sessionID, nil, nil, &status)
}
