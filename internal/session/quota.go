package session

import (
	"context"
	"fmt"
)

const (
	// DefaultMaxSessions is the default maximum number of concurrent sessions.
	DefaultMaxSessions = 100

	// DefaultMaxWorkspaceSize is the default maximum workspace size per session (500MB).
	DefaultMaxWorkspaceSize = 500 * 1024 * 1024 // 500MB
)

// QuotaConfig holds session quota settings.
type QuotaConfig struct {
	MaxSessions     int
	MaxWorkspaceSize int64 // bytes
}

// DefaultQuotaConfig returns the default quota configuration.
func DefaultQuotaConfig() QuotaConfig {
	return QuotaConfig{
		MaxSessions:     DefaultMaxSessions,
		MaxWorkspaceSize: DefaultMaxWorkspaceSize,
	}
}

// CheckQuota verifies that creating a new session or writing to a workspace
// does not exceed configured quotas. Returns an error if quota is exceeded.
func (m *Manager) CheckQuota(ctx context.Context, sessionID string) error {
	// Check max sessions limit
	if err := m.checkMaxSessions(ctx); err != nil {
		return err
	}

	// Check workspace size limit (only if session exists)
	if sessionID != "" {
		if err := m.checkWorkspaceSize(sessionID); err != nil {
			return err
		}
	}

	return nil
}

// checkMaxSessions verifies that the total number of sessions does not exceed the limit.
func (m *Manager) checkMaxSessions(ctx context.Context) error {
	sessions, err := m.ListSessions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list sessions for quota check: %w", err)
	}

	if len(sessions) > m.quota.MaxSessions {
		return fmt.Errorf("session quota exceeded: maximum %d sessions reached (NFR-17)", m.quota.MaxSessions)
	}

	return nil
}

// checkWorkspaceSize verifies that the workspace size does not exceed the limit.
func (m *Manager) checkWorkspaceSize(sessionID string) error {
	size, err := m.GetWorkspaceSize(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get workspace size: %w", err)
	}

	if size > m.quota.MaxWorkspaceSize {
		return fmt.Errorf("workspace size limit exceeded: %d bytes >= %d bytes (NFR-17)", size, m.quota.MaxWorkspaceSize)
	}

	return nil
}

// CheckQuotaForCreation checks if a new session can be created within quota limits.
// This should be called before CreateSessionWithName.
func (m *Manager) CheckQuotaForCreation(ctx context.Context) error {
	return m.checkMaxSessions(ctx)
}
