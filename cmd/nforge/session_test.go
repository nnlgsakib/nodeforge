package nforge

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunListSessions_NoSessions(t *testing.T) {
	tmpDir := t.TempDir()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runListSessions(tmpDir)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runListSessions failed: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	if !strings.Contains(output, "No sessions found") {
		t.Errorf("expected 'No sessions found' in output, got: %s", output)
	}
}

func TestRunResumeSession_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	err := runResumeSession(tmpDir, "sess-nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session, got nil")
	}
	if !strings.Contains(err.Error(), "session not found") {
		t.Errorf("expected 'session not found' error, got: %v", err)
	}
}

func TestRunExportSession_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	err := runExportSession(tmpDir, "sess-nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent session, got nil")
	}
	if !strings.Contains(err.Error(), "failed to load session") {
		t.Errorf("expected 'failed to load session' error, got: %v", err)
	}
}

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
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes_Exabyte(t *testing.T) {
	// Ensure no panic on very large values
	result := formatBytes(1 << 62)
	if result == "" {
		t.Error("formatBytes returned empty string for large value")
	}
	if !strings.Contains(result, "EB") {
		t.Errorf("expected EB unit for large value, got: %s", result)
	}
}
