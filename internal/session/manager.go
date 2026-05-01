package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

// Session represents a project workspace session
type Session struct {
	ID        string
	Name      string
	Workspace string
}

// Manager handles session creation and management
type Manager struct {
	workspaceRoot string
	mu           sync.Mutex
}

// NewManager creates a new session manager
func NewManager(workspaceRoot string) *Manager {
	return &Manager{workspaceRoot: workspaceRoot}
}

// CreateSessionWithName creates a new session with the specified name
func (m *Manager) CreateSessionWithName(ctx context.Context, name string) (*Session, error) {
	// Sanitize project name: reject path separators and empty names
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}
	if filepath.Base(name) != name {
		return nil, fmt.Errorf("project name must not contain path separators")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create project directory
	projectDir := filepath.Join(m.workspaceRoot, name)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	// Initialize .nforge workspace
	if err := InitProjectWorkspace(projectDir, name); err != nil {
		return nil, fmt.Errorf("failed to initialize workspace: %w", err)
	}

	return &Session{
		ID:        generateSessionID(),
		Name:      name,
		Workspace: projectDir,
	}, nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("sess-%s", uuid.New().String())
}
