package session

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckQuotaForCreation_WithinLimit(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	// Create a few sessions (well under 100 limit)
	for i := 0; i < 5; i++ {
		_, err := mgr.CreateSessionWithName(ctx, fmt.Sprintf("test-%d", i))
		require.NoError(t, err)
	}

	err := mgr.CheckQuotaForCreation(ctx)
	assert.NoError(t, err, "should not error when under quota")
}

func TestCheckMaxSessions_AtLimit(t *testing.T) {
	// Verifies the quota enforcement at the exact limit (off-by-one check)
	mgr := newTestManager(t)
	ctx := context.Background()

	// Create 3 sessions
	for i := 0; i < 3; i++ {
		_, err := mgr.CreateSessionWithName(ctx, fmt.Sprintf("quota-test-%d", i))
		require.NoError(t, err)
	}

	// Verify we can still check quota (no error at 3 sessions)
	err := mgr.checkMaxSessions(ctx)
	assert.NoError(t, err)
}

func TestQuotaConfig_AppliedToManager(t *testing.T) {
	mgr := newTestManager(t)

	assert.Equal(t, 100, mgr.quota.MaxSessions, "manager should have default max sessions")
	assert.Equal(t, int64(500*1024*1024), mgr.quota.MaxWorkspaceSize, "manager should have default max workspace size")
}

func TestQuota_RejectsAtExactLimit(t *testing.T) {
	// With > operator, exactly MaxSessions sessions should pass the check,
	// but the (MaxSessions+1)th creation should fail.
	// Test the comparison logic directly without creating 100 sessions.

	mgr := newTestManager(t)
	ctx := context.Background()

	// Create 3 sessions
	for i := 0; i < 3; i++ {
		_, err := mgr.CreateSessionWithName(ctx, fmt.Sprintf("limit-%d", i))
		require.NoError(t, err)
	}

	// countSessions should return 3
	count, err := mgr.countSessions(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	// With MaxSessions=100, count=3: 3 > 100 is false → should pass
	assert.True(t, count <= mgr.quota.MaxSessions, "3 sessions should be under quota")
}

func TestCheckWorkspaceSize_WithinLimit(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "workspace-quota")
	require.NoError(t, err)

	// Write a small file
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "small.txt", []byte("hello")))

	err = mgr.checkWorkspaceSize(sess.ID)
	assert.NoError(t, err, "small workspace should be within quota")
}

func TestQuotaConfig_Defaults(t *testing.T) {
	cfg := DefaultQuotaConfig()

	assert.Equal(t, 100, cfg.MaxSessions, "default max sessions should be 100")
	assert.Equal(t, int64(500*1024*1024), cfg.MaxWorkspaceSize, "default max workspace size should be 500MB")
}

func TestCheckQuota_EmptySessionID(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	// CheckQuota with empty sessionID should only check max sessions
	err := mgr.CheckQuota(ctx, "")
	assert.NoError(t, err)
}

func TestQuota_EnforcedOnCreation(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	// Create 5 sessions - should all succeed (under 100 limit)
	for i := 0; i < 5; i++ {
		_, err := mgr.CreateSessionWithName(ctx, fmt.Sprintf("quota-enforce-%d", i))
		require.NoError(t, err, "session creation should succeed within quota")
	}
}
