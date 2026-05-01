package session

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateSessionWithName(t *testing.T) {
	mgr := NewManager(t.TempDir())
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-session")
	if err != nil {
		t.Fatalf("CreateSessionWithName failed: %v", err)
	}
	if sess == nil {
		t.Fatal("session is nil")
	}
	if sess.Name != "test-session" {
		t.Errorf("expected session name 'test-session', got %q", sess.Name)
	}
	if sess.ID == "" {
		t.Error("session ID is empty")
	}
	// Verify UUID format (sess-UUID)
	if !strings.HasPrefix(sess.ID, "sess-") {
		t.Errorf("expected session ID to start with 'sess-', got %q", sess.ID)
	}
	if sess.Workspace == "" {
		t.Error("session workspace is empty")
	}

	// Check that .nforge directory was created
	nforgeDir := filepath.Join(sess.Workspace, ".nforge")
	if _, err := os.Stat(nforgeDir); os.IsNotExist(err) {
		t.Fatalf(".nforge directory not created at %s", nforgeDir)
	}
}

func TestCreateSessionWithNameInvalidName(t *testing.T) {
	mgr := NewManager(t.TempDir())
	ctx := context.Background()

	// Test empty name
	_, err := mgr.CreateSessionWithName(ctx, "")
	if err == nil {
		t.Error("expected error for empty project name, got nil")
	}

	// Test path traversal in name
	_, err = mgr.CreateSessionWithName(ctx, "../../etc")
	if err == nil {
		t.Error("expected error for path traversal in project name, got nil")
	}

	// Test name with slashes
	_, err = mgr.CreateSessionWithName(ctx, "foo/bar")
	if err == nil {
		t.Error("expected error for project name with slashes, got nil")
	}
}

func TestCreateSessionWithNameDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)
	ctx := context.Background()

	// Create first session
	_, err := mgr.CreateSessionWithName(ctx, "dup-project")
	if err != nil {
		t.Fatalf("first CreateSessionWithName failed: %v", err)
	}

	// Try to create duplicate — should fail since .nforge already exists
	_, err = mgr.CreateSessionWithName(ctx, "dup-project")
	if err == nil {
		t.Error("expected error for duplicate project (existing .nforge/), got nil")
	}
}
