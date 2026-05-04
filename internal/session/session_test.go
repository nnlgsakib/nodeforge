package session

import "testing"

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	mgr, err := NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test manager: %v", err)
	}
	t.Cleanup(func() { mgr.Close() })
	return mgr
}
