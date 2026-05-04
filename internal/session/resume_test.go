package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResumeSession(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	ctx := context.Background()

	// Create a session
	sess, err := mgr.CreateSessionWithName(ctx, "test-resume")
	require.NoError(t, err)
	originalActiveAt := sess.LastActiveAt

	// Mark it as complete (simulating shutdown)
	err = mgr.UpdateSessionStatus(ctx, sess.ID, StatusComplete)
	require.NoError(t, err)

	// Small delay to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Resume the session
	resumed, err := mgr.ResumeSession(ctx, sess.ID)
	require.NoError(t, err)

	assert.Equal(t, StatusRunning, resumed.Status)
	assert.True(t, resumed.LastActiveAt.After(originalActiveAt), "LastActiveAt should be updated on resume")
	assert.False(t, resumed.HeartbeatAt.IsZero(), "HeartbeatAt should be set on resume")
}

func TestResumeSessionNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	ctx := context.Background()
	_, err = mgr.ResumeSession(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSnapshotAllSessions(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	ctx := context.Background()

	// Create two running sessions
	sess1, err := mgr.CreateSessionWithName(ctx, "snap-test-1")
	require.NoError(t, err)
	sess2, err := mgr.CreateSessionWithName(ctx, "snap-test-2")
	require.NoError(t, err)

	// Snapshot all
	err = mgr.SnapshotAllSessions()
	require.NoError(t, err)

	// Verify both are now complete
	s1, _ := mgr.GetSession(ctx, sess1.ID)
	s2, _ := mgr.GetSession(ctx, sess2.ID)
	assert.Equal(t, StatusComplete, s1.Status)
	assert.Equal(t, StatusComplete, s2.Status)
}

func TestCleanupZombieSessions(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	ctx := context.Background()

	// Create a session
	sess, err := mgr.CreateSessionWithName(ctx, "zombie-test")
	require.NoError(t, err)

	// Set heartbeat to 10 minutes ago (beyond 5min timeout)
	mgr.mu.Lock()
	_, err = mgr.db.Exec("UPDATE sessions SET heartbeat_at = ? WHERE id = ?",
		time.Now().UTC().Add(-10*time.Minute), sess.ID)
	mgr.mu.Unlock()
	require.NoError(t, err)

	// Run zombie cleanup with 5min timeout
	zombieIDs, err := mgr.CleanupZombieSessions(ctx, 5*time.Minute)
	require.NoError(t, err)
	assert.Len(t, zombieIDs, 1)
	assert.Equal(t, sess.ID, zombieIDs[0])

	// Verify session is now zombie
	s, err := mgr.GetSession(ctx, sess.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusZombie, s.Status)
}

func TestCleanupZombieSessions_NoZombies(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	ctx := context.Background()

	// Create a session with recent heartbeat
	sess, err := mgr.CreateSessionWithName(ctx, "no-zombie-test")
	require.NoError(t, err)
	err = mgr.UpdateHeartbeat(ctx, sess.ID)
	require.NoError(t, err)

	// Run zombie cleanup with 5min timeout
	zombieIDs, err := mgr.CleanupZombieSessions(ctx, 5*time.Minute)
	require.NoError(t, err)
	assert.Empty(t, zombieIDs)
}

func TestListZombieSessions(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	ctx := context.Background()

	// Create and zombie a session
	sess, err := mgr.CreateSessionWithName(ctx, "list-zombie-test")
	require.NoError(t, err)
	_, err = mgr.CleanupZombieSessions(ctx, 0) // 0 timeout = immediate
	require.NoError(t, err)

	// List zombies
	zombies, err := mgr.ListZombieSessions(ctx)
	require.NoError(t, err)
	assert.Len(t, zombies, 1)
	assert.Equal(t, sess.ID, zombies[0].ID)
	assert.Equal(t, StatusZombie, zombies[0].Status)
}

func TestUpdateHeartbeat(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "heartbeat-test")
	require.NoError(t, err)

	// Set old heartbeat
	oldTime := time.Now().UTC().Add(-1 * time.Hour)
	mgr.mu.Lock()
	_, _ = mgr.db.Exec("UPDATE sessions SET heartbeat_at = ? WHERE id = ?", oldTime, sess.ID)
	mgr.mu.Unlock()

	// Update heartbeat
	err = mgr.UpdateHeartbeat(ctx, sess.ID)
	require.NoError(t, err)

	// Verify heartbeat was updated
	s, err := mgr.GetSession(ctx, sess.ID)
	require.NoError(t, err)
	assert.True(t, s.HeartbeatAt.After(oldTime))
}

func TestHeartbeatMonitor(t *testing.T) {
	tmpDir := t.TempDir()
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "monitor-test")
	require.NoError(t, err)

	cfg := HeartbeatConfig{
		Timeout:       100 * time.Millisecond,
		CheckInterval: 50 * time.Millisecond,
	}
	mon := NewHeartbeatMonitor(mgr, cfg)
	mon.Start(ctx)
	defer mon.Stop()

	// Record a heartbeat
	mon.Beat(sess.ID)

	// Wait for cleanup cycle to run and mark as zombie
	time.Sleep(300 * time.Millisecond)

	// Session should be zombie
	s, err := mgr.GetSession(ctx, sess.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusZombie, s.Status)
}

func TestSessionStatusZombie(t *testing.T) {
	assert.Equal(t, SessionStatus("zombie"), StatusZombie)
}

func TestSchemaMigration(t *testing.T) {
	// Test that opening an existing DB adds new columns
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, ".nforge", "sessions.db")
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	// Create DB with old schema (without snapshot/heartbeat_at)
	db, err := openDB(dbPath)
	require.NoError(t, err)
	db.Close()

	// Reopen - should still work (CREATE TABLE IF NOT EXISTS)
	mgr, err := NewManager(tmpDir)
	require.NoError(t, err)
	defer mgr.Close()

	// Verify new columns exist
	var hasSnapshot, hasHeartbeatAt bool
	err = mgr.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name='snapshot'`).Scan(&hasSnapshot)
	require.NoError(t, err)
	assert.True(t, hasSnapshot)

	err = mgr.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name='heartbeat_at'`).Scan(&hasHeartbeatAt)
	require.NoError(t, err)
	assert.True(t, hasHeartbeatAt)
}
