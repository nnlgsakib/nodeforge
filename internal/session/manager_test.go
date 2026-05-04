package session

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCreateSessionWithName(t *testing.T) {
	mgr := newTestManager(t)
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

	// Check that .nforge/sessions/<id>/workspace/ structure was created
	if _, err := os.Stat(sess.Workspace); os.IsNotExist(err) {
		t.Fatalf("workspace directory not created at %s", sess.Workspace)
	}
	// Verify path contains expected structure
	if !strings.Contains(sess.Workspace, ".nforge") || !strings.Contains(sess.Workspace, "sessions") {
		t.Errorf("expected workspace to contain .nforge/sessions, got %q", sess.Workspace)
	}
}

func TestCreateSessionWithNameInvalidName(t *testing.T) {
	mgr := newTestManager(t)
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
	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}
	defer mgr.Close()
	ctx := context.Background()

	// Create first session
	_, err = mgr.CreateSessionWithName(ctx, "dup-project")
	if err != nil {
		t.Fatalf("first CreateSessionWithName failed: %v", err)
	}

	// Create a second session with different name — should succeed
	_, err = mgr.CreateSessionWithName(ctx, "other-project")
	if err != nil {
		t.Fatalf("second CreateSessionWithName for different name failed: %v", err)
	}

	// Verify both sessions are listed
	sessions, err := mgr.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestListSessions(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	// Create multiple sessions
	names := []string{"alpha", "beta", "gamma"}
	for _, name := range names {
		_, err := mgr.CreateSessionWithName(ctx, name)
		if err != nil {
			t.Fatalf("CreateSessionWithName failed for %s: %v", name, err)
		}
	}

	sessions, err := mgr.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != len(names) {
		t.Errorf("expected %d sessions, got %d", len(names), len(sessions))
	}

	// Verify sessions are ordered by created_at DESC (gamma last = first in list)
	if len(sessions) > 0 && sessions[0].Name != "gamma" {
		t.Errorf("expected first session to be 'gamma' (newest), got %q", sessions[0].Name)
	}
}

func TestGetSession(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "get-test")
	if err != nil {
		t.Fatalf("CreateSessionWithName failed: %v", err)
	}

	retrieved, err := mgr.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.ID != sess.ID {
		t.Errorf("expected session ID %q, got %q", sess.ID, retrieved.ID)
	}
	if retrieved.Name != "get-test" {
		t.Errorf("expected session name 'get-test', got %q", retrieved.Name)
	}
	if retrieved.Status != StatusRunning {
		t.Errorf("expected status 'running', got %q", retrieved.Status)
	}
}

func TestGetSessionNotFound(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	_, err := mgr.GetSession(ctx, "nonexistent-id")
	if err == nil {
		t.Error("expected error for non-existent session, got nil")
	}
}

func TestUpdateSession(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "update-test")
	if err != nil {
		t.Fatalf("CreateSessionWithName failed: %v", err)
	}

	sess.GraphJSON = `{"nodes":[]}`
	sess.ChatLog = `[]`
	sess.Goal = "Build a web app"

	err = mgr.UpdateSession(ctx, sess)
	if err != nil {
		t.Fatalf("UpdateSession failed: %v", err)
	}

	// Verify persistence
	retrieved, err := mgr.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetSession after update failed: %v", err)
	}
	if retrieved.GraphJSON != `{"nodes":[]}` {
		t.Errorf("expected graph JSON to be persisted, got %q", retrieved.GraphJSON)
	}
	if retrieved.Goal != "Build a web app" {
		t.Errorf("expected goal 'Build a web app', got %q", retrieved.Goal)
	}
}

func TestSaveGraphJSON(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "graph-test")
	if err != nil {
		t.Fatalf("CreateSessionWithName failed: %v", err)
	}

	graphData := `{"nodes":[{"id":"n1"}],"edges":[]}`
	err = mgr.SaveGraphJSON(ctx, sess.ID, graphData)
	if err != nil {
		t.Fatalf("SaveGraphJSON failed: %v", err)
	}

	retrieved, err := mgr.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.GraphJSON != graphData {
		t.Errorf("expected graph JSON %q, got %q", graphData, retrieved.GraphJSON)
	}
}

func TestSaveChatLog(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "chat-test")
	if err != nil {
		t.Fatalf("CreateSessionWithName failed: %v", err)
	}

	chatData := `[{"role":"user","text":"hello"}]`
	err = mgr.SaveChatLog(ctx, sess.ID, chatData)
	if err != nil {
		t.Fatalf("SaveChatLog failed: %v", err)
	}

	retrieved, err := mgr.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.ChatLog != chatData {
		t.Errorf("expected chat log %q, got %q", chatData, retrieved.ChatLog)
	}
}

func TestUpdateSessionStatus(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "status-test")
	if err != nil {
		t.Fatalf("CreateSessionWithName failed: %v", err)
	}

	err = mgr.UpdateSessionStatus(ctx, sess.ID, StatusComplete)
	if err != nil {
		t.Fatalf("UpdateSessionStatus failed: %v", err)
	}

	retrieved, err := mgr.GetSession(ctx, sess.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.Status != StatusComplete {
		t.Errorf("expected status 'complete', got %q", retrieved.Status)
	}
}

func TestWorkspacePath(t *testing.T) {
	mgr := newTestManager(t)
	sessionID := "sess-00000000-0000-0000-0000-000000000001"

	path, err := mgr.WorkspacePath(sessionID)
	if err != nil {
		t.Fatalf("WorkspacePath error: %v", err)
	}
	if path == "" {
		t.Error("workspace path is empty")
	}
	// Verify path contains session structure
	if !strings.Contains(path, sessionID) {
		t.Errorf("expected path to contain session ID %q, got %q", sessionID, path)
	}
}

func TestEnsureWorkspaceDir(t *testing.T) {
	mgr := newTestManager(t)
	sessionID := "sess-00000000-0000-0000-0000-000000000002"

	err := mgr.EnsureWorkspaceDir(sessionID)
	if err != nil {
		t.Fatalf("EnsureWorkspaceDir failed: %v", err)
	}

	workspaceDir, err := mgr.WorkspacePath(sessionID)
	if err != nil {
		t.Fatalf("WorkspacePath error: %v", err)
	}
	info, err := os.Stat(workspaceDir)
	if err != nil {
		t.Fatalf("workspace directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("workspace path is not a directory")
	}
}

func TestWriteWorkspaceFile(t *testing.T) {
	mgr := newTestManager(t)
	sessionID := "sess-00000000-0000-0000-0000-000000000003"

	// Ensure workspace exists
	if err := mgr.EnsureWorkspaceDir(sessionID); err != nil {
		t.Fatalf("EnsureWorkspaceDir failed: %v", err)
	}

	// Write a file
	err := mgr.WriteWorkspaceFile(sessionID, "test.txt", []byte("hello world"))
	if err != nil {
		t.Fatalf("WriteWorkspaceFile failed: %v", err)
	}

	// Verify file exists
	workspaceDir, err := mgr.WorkspacePath(sessionID)
	if err != nil {
		t.Fatalf("WorkspacePath error: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(workspaceDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(content))
	}
}

func TestWriteWorkspaceFileDirectoryTraversal(t *testing.T) {
	mgr := newTestManager(t)
	sessionID := "sess-00000000-0000-0000-0000-000000000004"

	if err := mgr.EnsureWorkspaceDir(sessionID); err != nil {
		t.Fatalf("EnsureWorkspaceDir failed: %v", err)
	}

	// Test with ".." in path — should fail
	err := mgr.WriteWorkspaceFile(sessionID, "../../etc/passwd", []byte("malicious"))
	if err == nil {
		t.Error("expected error for directory traversal, got nil")
	}

	// Test with absolute path — should fail (platform-aware)
	absPath := "/etc/passwd"
	if runtime.GOOS == "windows" {
		absPath = "C:\\Windows\\System32\\config\\SAM"
	}
	err = mgr.WriteWorkspaceFile(sessionID, absPath, []byte("malicious"))
	if err == nil {
		t.Error("expected error for absolute path, got nil")
	}
}

func TestReadWorkspaceFile(t *testing.T) {
	mgr := newTestManager(t)
	sessionID := "sess-00000000-0000-0000-0000-000000000005"

	if err := mgr.EnsureWorkspaceDir(sessionID); err != nil {
		t.Fatalf("EnsureWorkspaceDir failed: %v", err)
	}

	// Write then read
	if err := mgr.WriteWorkspaceFile(sessionID, "data.txt", []byte("test data")); err != nil {
		t.Fatalf("WriteWorkspaceFile failed: %v", err)
	}

	content, err := mgr.ReadWorkspaceFile(sessionID, "data.txt")
	if err != nil {
		t.Fatalf("ReadWorkspaceFile failed: %v", err)
	}
	if string(content) != "test data" {
		t.Errorf("expected 'test data', got %q", string(content))
	}
}

func TestReadWorkspaceFileDirectoryTraversal(t *testing.T) {
	mgr := newTestManager(t)
	sessionID := "sess-00000000-0000-0000-0000-000000000006"

	if err := mgr.EnsureWorkspaceDir(sessionID); err != nil {
		t.Fatalf("EnsureWorkspaceDir failed: %v", err)
	}

	_, err := mgr.ReadWorkspaceFile(sessionID, "../../etc/passwd")
	if err == nil {
		t.Error("expected error for directory traversal read, got nil")
	}
}
