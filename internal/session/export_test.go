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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportSession_CreatesValidTarball(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-export")
	require.NoError(t, err)

	// Write some workspace files
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "main.go", []byte("package main")))
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "readme.txt", []byte("hello")))

	// Export to a buffer using ExportSessionToWriter
	var buf bytes.Buffer
	err = ExportSessionToWriter(ctx, mgr, sess.ID, &buf)
	require.NoError(t, err)
	assert.True(t, buf.Len() > 0, "tarball should not be empty")

	// Verify tarball contents
	gr, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	defer gr.Close()

	tr := tar.NewReader(gr)
	files := make(map[string]bool)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		files[hdr.Name] = true
		t.Logf("tarball entry: %s", hdr.Name)
	}

	assert.True(t, files["graph.json"], "tarball should contain graph.json")
	assert.True(t, files["README.md"], "tarball should contain README.md")
	assert.True(t, files["workspace/main.go"], "tarball should contain workspace files")
	assert.True(t, files["workspace/readme.txt"], "tarball should contain workspace files")
}

func TestExportSession_ExcludesSecrets(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-secrets")
	require.NoError(t, err)

	// Write workspace files including secrets
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "main.go", []byte("package main")))
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, ".env", []byte("API_KEY=sk-1234")))
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "config.yaml", []byte("api_key: secret")))

	var buf bytes.Buffer
	err = ExportSessionToWriter(ctx, mgr, sess.ID, &buf)
	require.NoError(t, err)

	gr, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	defer gr.Close()

	tr := tar.NewReader(gr)
	files := make(map[string]bool)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		files[hdr.Name] = true
	}

	assert.True(t, files["workspace/main.go"], "should include normal files")
	assert.False(t, files["workspace/.env"], "should exclude .env")
	assert.False(t, files["workspace/config.yaml"], "should exclude config.yaml")
}

func TestExportSession_SanitizesGraphJSON(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-sanitize")
	require.NoError(t, err)

	// Inject graph with API key
	sess.GraphJSON = `{"nodes": [{"id": "n1", "output": "result", "api_key": "sk-secret123"}]}`
	require.NoError(t, mgr.UpdateSession(ctx, sess))

	var buf bytes.Buffer
	err = ExportSessionToWriter(ctx, mgr, sess.ID, &buf)
	require.NoError(t, err)

	gr, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if hdr.Name == "graph.json" {
			content, err := io.ReadAll(tr)
			require.NoError(t, err)
			assert.NotContains(t, string(content), "sk-secret123", "API key should be redacted")
			assert.Contains(t, string(content), "[REDACTED]", "API key should be replaced with [REDACTED]")
		}
	}
}

func TestExportSession_GeneratesREADME(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-readme")
	require.NoError(t, err)
	sess.Goal = "Build a web app"
	require.NoError(t, mgr.UpdateSession(ctx, sess))

	var buf bytes.Buffer
	err = ExportSessionToWriter(ctx, mgr, sess.ID, &buf)
	require.NoError(t, err)

	gr, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if hdr.Name == "README.md" {
			content, err := io.ReadAll(tr)
			require.NoError(t, err)
			readme := string(content)
			assert.Contains(t, readme, "test-readme")
			assert.Contains(t, readme, "Build a web app")
			assert.Contains(t, readme, "API keys, configuration files, and secrets have been excluded")
		}
	}
}

func TestExportSession_FileOutput(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-file-export")
	require.NoError(t, err)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "export.tar.gz")

	actualPath, err := ExportSession(ctx, mgr, sess.ID, outputPath)
	require.NoError(t, err)
	assert.Equal(t, outputPath, actualPath, "should return the specified output path")

	// Verify file exists and is valid
	stat, err := os.Stat(actualPath)
	require.NoError(t, err)
	assert.True(t, stat.Size() > 0, "output file should not be empty")
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{".env", true},
		{"config.yaml", true},
		{"my_secret.txt", true},
		{"api_key.json", true},
		{"secret_key.pem", true},
		{"credentials.json", true},
		{"cert.pem", true},
		{".nforge/config.yaml", true},
		{"main.go", false},
		{"readme.md", false},
		{"src/index.ts", false},
		{"keyboard.js", false},
		{"primaryKey.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, isExcluded(tt.path))
		})
	}
}

func TestSanitizeGraphJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		checks   []string // should NOT contain
		contains string   // should contain
	}{
		{
			name:     "redacts api_key",
			input:    `{"api_key": "sk-123"}`,
			checks:   []string{"sk-123"},
			contains: "[REDACTED]",
		},
		{
			name:     "redacts token",
			input:    `{"token": "abc"}`,
			checks:   []string{"abc"},
			contains: "[REDACTED]",
		},
		{
			name:     "redacts nested secrets",
			input:    `{"config": {"secret": "hidden"}}`,
			checks:   []string{"hidden"},
			contains: "[REDACTED]",
		},
		{
			name:     "empty string",
			input:    "",
			checks:   []string{},
			contains: "{}",
		},
		{
			name:     "invalid JSON",
			input:    "{not json}",
			checks:   []string{},
			contains: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeGraphJSON(tt.input)
			for _, check := range tt.checks {
				assert.NotContains(t, result, check)
			}
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestGetWorkspaceSize(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-size")
	require.NoError(t, err)

	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "a.txt", []byte("hello")))
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "b.txt", []byte("world!")))

	size, err := mgr.GetWorkspaceSize(sess.ID)
	require.NoError(t, err)
	assert.True(t, size > 0, "workspace size should be positive")
}

func TestExportSession_NonExistentSession(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	var buf bytes.Buffer
	err := ExportSessionToWriter(ctx, mgr, "sess-nonexistent", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load session")
}

func TestExportSession_WithContextDir(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-context")
	require.NoError(t, err)

	// Simulate .nforge/context.db/ directory (should be excluded)
	workspacePath, err := mgr.WorkspacePath(sess.ID)
	require.NoError(t, err)
	contextDir := filepath.Join(filepath.Dir(workspacePath), "..", ".nforge", "context.db")
	require.NoError(t, os.MkdirAll(contextDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(contextDir, "000001.sst"), []byte("data"), 0644))

	var buf bytes.Buffer
	err = ExportSessionToWriter(ctx, mgr, sess.ID, &buf)
	require.NoError(t, err)

	// Verify context.db is not in tarball
	gr, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		assert.False(t, strings.Contains(hdr.Name, "context.db"), "context.db should be excluded")
	}
}

func TestExportSession_ContextCancellation(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-cancel")
	require.NoError(t, err)

	// Write many files to make export take time
	for i := 0; i < 10; i++ {
		require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, fmt.Sprintf("file%d.txt", i), []byte("content")))
	}

	// Create a cancelled context
	ctx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	var buf bytes.Buffer
	err = ExportSessionToWriter(ctx, mgr, sess.ID, &buf)
	assert.Error(t, err, "export should fail with cancelled context")
	assert.Contains(t, err.Error(), "context canceled", "error should mention context cancellation")
}

func TestExportSession_SkipsSymlinks(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	sess, err := mgr.CreateSessionWithName(ctx, "test-symlink")
	require.NoError(t, err)

	// Write a real file
	require.NoError(t, mgr.WriteWorkspaceFile(sess.ID, "real.txt", []byte("hello")))

	// Create a symlink (skip on Windows where symlinks require admin)
	workspacePath, err := mgr.WorkspacePath(sess.ID)
	require.NoError(t, err)
	linkPath := filepath.Join(workspacePath, "link.txt")
	if err := os.Symlink(filepath.Join(workspacePath, "real.txt"), linkPath); err != nil {
		t.Skip("symlinks not supported")
	}

	var buf bytes.Buffer
	err = ExportSessionToWriter(ctx, mgr, sess.ID, &buf)
	require.NoError(t, err)

	gr, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	defer gr.Close()

	tr := tar.NewReader(gr)
	entries := make(map[string]bool)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		entries[hdr.Name] = true
	}

	assert.True(t, entries["workspace/real.txt"], "should include real file")
	assert.False(t, entries["workspace/link.txt"], "should exclude symlink")
}
