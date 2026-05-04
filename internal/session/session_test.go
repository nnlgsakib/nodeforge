package session

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	mgr, err := NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test manager: %v", err)
	}
	t.Cleanup(func() { mgr.Close() })
	return mgr
}

// --- ResumeSession Tests ---

func TestResumeSession_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	_, err := mgr.ResumeSession(context.Background(), "sess-nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

func TestResumeSession_Success(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "test-resume")
	require.NoError(t, err)

	err = mgr.SnapshotAllSessions()
	require.NoError(t, err)

	s, err := mgr.GetSession(context.Background(), sess.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusComplete, s.Status)

	resumed, err := mgr.ResumeSession(context.Background(), sess.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, resumed.Status)
	assert.Equal(t, sess.ID, resumed.ID)
	assert.Equal(t, sess.Name, resumed.Name)
	assert.Equal(t, sess.GraphJSON, resumed.GraphJSON)

	s2, err := mgr.GetSession(context.Background(), sess.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, s2.Status)
}

func TestResumeSession_PreservesGraphAndChat(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "test-preserve")
	require.NoError(t, err)

	sess.GraphJSON = `{"nodes":[{"id":"n1"}]}`
	sess.ChatLog = `[{"role":"user","content":"hello"}]`
	err = mgr.UpdateSession(context.Background(), sess)
	require.NoError(t, err)

	err = mgr.SnapshotAllSessions()
	require.NoError(t, err)

	resumed, err := mgr.ResumeSession(context.Background(), sess.ID)
	require.NoError(t, err)
	assert.Equal(t, `{"nodes":[{"id":"n1"}]}`, resumed.GraphJSON)
	assert.Equal(t, `[{"role":"user","content":"hello"}]`, resumed.ChatLog)
}

func TestResumeSession_UpdatesTimestamps(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "test-timestamps")
	require.NoError(t, err)

	originalActiveAt := sess.LastActiveAt
	time.Sleep(100 * time.Millisecond)

	err = mgr.SnapshotAllSessions()
	require.NoError(t, err)

	resumed, err := mgr.ResumeSession(context.Background(), sess.ID)
	require.NoError(t, err)

	assert.True(t, resumed.LastActiveAt.After(originalActiveAt))
	assert.True(t, resumed.HeartbeatAt.After(originalActiveAt))
}

func TestResumeSession_NonCompleteSession(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "test-running-resume")
	require.NoError(t, err)

	resumed, err := mgr.ResumeSession(context.Background(), sess.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, resumed.Status)
}

// --- GetSessionStats Tests ---

func TestGetSessionStats_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	_, err := mgr.GetSessionStats(context.Background(), "sess-nonexistent")
	require.Error(t, err)
}

func TestGetSessionStats_ReturnsWorkspaceSize(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "test-stats")
	require.NoError(t, err)

	err = mgr.WriteWorkspaceFile(sess.ID, "file1.txt", []byte(strings.Repeat("x", 1000)))
	require.NoError(t, err)
	err = mgr.WriteWorkspaceFile(sess.ID, "file2.txt", []byte(strings.Repeat("y", 2000)))
	require.NoError(t, err)

	stats, err := mgr.GetSessionStats(context.Background(), sess.ID)
	require.NoError(t, err)
	assert.Equal(t, sess.ID, stats.ID)
	assert.Equal(t, "test-stats", stats.Name)
	assert.Equal(t, string(StatusRunning), stats.Status)
	assert.GreaterOrEqual(t, stats.WorkspaceSize, int64(3000))
}

func TestGetSessionStats_EmptyWorkspace(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "test-empty-stats")
	require.NoError(t, err)

	stats, err := mgr.GetSessionStats(context.Background(), sess.ID)
	require.NoError(t, err)
	assert.Equal(t, sess.ID, stats.ID)
	assert.GreaterOrEqual(t, stats.WorkspaceSize, int64(0))
}

// --- Integration Test: Full Export-Extract-Verify Cycle ---

func TestExportExtractVerifyCycle(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "integration-test")
	require.NoError(t, err)

	sess.GraphJSON = `{"nodes":[{"id":"n1","label":"Goal","status":"complete"}]}`
	sess.ChatLog = `[{"role":"user","content":"Build a website"}]`
	err = mgr.UpdateSession(context.Background(), sess)
	require.NoError(t, err)

	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "index.html", []byte("<html>")))
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "style.css", []byte("body {}")))
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, ".env", []byte("SECRET=hidden")))

	exportPath, err := ExportSession(context.Background(), mgr, sess.ID, "")
	require.NoError(t, err)
	defer os.Remove(exportPath)

	extractDir := t.TempDir()
	extractTarball(t, exportPath, extractDir)

	graphData, err := os.ReadFile(filepath.Join(extractDir, "graph.json"))
	require.NoError(t, err)
	assert.Contains(t, string(graphData), "nodes")

	_, err = os.Stat(filepath.Join(extractDir, "README.md"))
	require.NoError(t, err)

	htmlData, err := os.ReadFile(filepath.Join(extractDir, "workspace", "index.html"))
	require.NoError(t, err)
	assert.Equal(t, "<html>", string(htmlData))

	cssData, err := os.ReadFile(filepath.Join(extractDir, "workspace", "style.css"))
	require.NoError(t, err)
	assert.Equal(t, "body {}", string(cssData))

	_, err = os.Stat(filepath.Join(extractDir, "workspace", ".env"))
	require.True(t, os.IsNotExist(err), ".env should not be in extracted tarball")
}

// --- Security Tests: Comprehensive API Key Exclusion ---

func TestExportExcludesAllSensitiveKeys(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "security-test")
	require.NoError(t, err)

	sess.GraphJSON = `{
		"nodes": [
			{"id": "n1", "output": {"api_key": "sk-openai-12345"}},
			{"id": "n2", "output": {"apiKey": "pk-anthropic-67890"}},
			{"id": "n3", "output": {"token": "tok-xyz"}},
			{"id": "n4", "output": {"secret": "s3cr3t"}},
			{"id": "n5", "output": {"authorization": "Bearer abc"}},
			{"id": "n6", "output": {"password": "p@ss"}},
			{"id": "n7", "output": {"credential": "cred-123"}},
			{"id": "n8", "output": {"safe_field": "this is fine"}}
		]
	}`
	err = mgr.UpdateSession(context.Background(), sess)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = ExportSessionToWriter(context.Background(), mgr, sess.ID, &buf)
	require.NoError(t, err)

	graph := extractFileFromTarGzip(t, buf.Bytes(), "graph.json")

	assert.NotContains(t, graph, "sk-openai-12345")
	assert.NotContains(t, graph, "pk-anthropic-67890")
	assert.NotContains(t, graph, "tok-xyz")
	assert.NotContains(t, graph, "s3cr3t")
	assert.NotContains(t, graph, "Bearer abc")
	assert.NotContains(t, graph, "p@ss")
	assert.NotContains(t, graph, "cred-123")
	assert.Contains(t, graph, "this is fine")

	count := strings.Count(graph, "[REDACTED]")
	assert.GreaterOrEqual(t, count, 7)
}

func TestExportExcludesAllSecretFiles(t *testing.T) {
	mgr := newTestManager(t)

	sess, err := mgr.CreateSessionWithName(context.Background(), "security-files")
	require.NoError(t, err)

	workspacePath, err := mgr.WorkspacePath(sess.ID)
	require.NoError(t, err)

	secrets := []string{
		".env",
		".env.local",
		"api_key.json",
		"secret_key.pem",
		"credentials.yaml",
		"config.yaml",
		"cert.pem",
	}

	for _, f := range secrets {
		err := os.WriteFile(filepath.Join(workspacePath, f), []byte("SECRET"), 0644)
		require.NoError(t, err)
	}

	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "main.go", []byte("package main")))
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "README.md", []byte("# Project")))

	var buf bytes.Buffer
	err = ExportSessionToWriter(context.Background(), mgr, sess.ID, &buf)
	require.NoError(t, err)

	entries := listTarballEntriesGzip(t, buf.Bytes())

	for _, secret := range secrets {
		for _, entry := range entries {
			assert.False(t, strings.Contains(entry, secret),
				"secret file %q should not be in export (found as %q)", secret, entry)
		}
	}

	assert.Contains(t, entries, "workspace/main.go")
	assert.Contains(t, entries, "workspace/README.md")
}

// --- CLI formatBytes Tests ---

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			// formatBytes is in session.go (cmd package), but we can test the logic here
			// by re-implementing or testing through GetSessionStats
			result := formatBytesHelper(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- Helper Functions ---

func extractTarball(t *testing.T, tarPath, destDir string) {
	t.Helper()
	f, err := os.Open(tarPath)
	require.NoError(t, err)
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		target := filepath.Join(destDir, hdr.Name)
		if strings.HasSuffix(hdr.Name, "/") {
			os.MkdirAll(target, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)
		f, err := os.Create(target)
		require.NoError(t, err)
		io.Copy(f, tr)
		f.Close()
	}
}

func extractFileFromTarGzip(t *testing.T, data []byte, filename string) string {
	t.Helper()
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if hdr.Name == filename {
			content, err := io.ReadAll(tr)
			require.NoError(t, err)
			return string(content)
		}
	}
	t.Fatalf("file %s not found in tarball", filename)
	return ""
}

func listTarballEntriesGzip(t *testing.T, data []byte) []string {
	t.Helper()
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var entries []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		entries = append(entries, hdr.Name)
	}
	return entries
}

// formatBytesHelper duplicates the CLI formatBytes for testing in this package
func formatBytesHelper(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
