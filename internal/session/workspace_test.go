package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitGitRepo_Success(t *testing.T) {
	tmpDir := t.TempDir()

	err := InitGitRepo(tmpDir)
	require.NoError(t, err)

	// Verify it's a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "should be a git repo after init")
	assert.Contains(t, string(output), ".git")

	// Verify git config is set
	cmd = exec.Command("git", "config", "user.email")
	cmd.Dir = tmpDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "nforge@local\n", string(output))

	cmd = exec.Command("git", "config", "user.name")
	cmd.Dir = tmpDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Equal(t, "NodeForge\n", string(output))
}

func TestInitGitRepo_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// Call twice — should not error
	err := InitGitRepo(tmpDir)
	require.NoError(t, err)

	err = InitGitRepo(tmpDir)
	require.NoError(t, err, "should be safe to call multiple times")
}

func TestInitGitRepo_NonExistentDir(t *testing.T) {
	err := InitGitRepo("/nonexistent/path/that/does/not/exist")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workspace directory does not exist")
}

func TestCopyFilePreservesPermissions(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create an executable file
	srcFile := filepath.Join(srcDir, "script.sh")
	err := os.WriteFile(srcFile, []byte("#!/bin/bash\necho hello"), 0755)
	require.NoError(t, err)

	dstFile := filepath.Join(dstDir, "script.sh")
	err = copyFile(srcFile, dstFile)
	require.NoError(t, err)

	// Verify permissions preserved
	srcInfo, err := os.Stat(srcFile)
	require.NoError(t, err)
	dstInfo, err := os.Stat(dstFile)
	require.NoError(t, err)
	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode(), "file permissions should be preserved")
}
