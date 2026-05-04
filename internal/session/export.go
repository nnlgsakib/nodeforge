package session

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// excludedPatterns defines file patterns that must never be included in export tarballs.
var excludedPatterns = []string{
	".env",
	".env.*",
	".git",
	"config.yaml",
	".nforge",
	"*secret*",
	"*api_key*",
	"*secret_key*",
	"*credential*",
	"*.pem",
}

// isExcluded returns true if the given path matches any exclusion pattern.
func isExcluded(relPath string) bool {
	base := filepath.Base(relPath)
	for _, pattern := range excludedPatterns {
		if matched, err := filepath.Match(pattern, base); err != nil {
			log.Printf("warning: invalid exclusion pattern %q: %v", pattern, err)
		} else if matched {
			return true
		}
	}
	// Check path segments for directory patterns (e.g., .nforge)
	parts := strings.Split(filepath.ToSlash(relPath), "/")
	for _, part := range parts {
		for _, pattern := range excludedPatterns {
			if matched, err := filepath.Match(pattern, part); err != nil {
				log.Printf("warning: invalid exclusion pattern %q: %v", pattern, err)
			} else if matched {
				return true
			}
		}
	}
	return false
}

// sanitizeGraphJSON removes any embedded API keys or tokens from node outputs.
func sanitizeGraphJSON(graphJSON string) string {
	if graphJSON == "" {
		return "{}"
	}

	// Try as object first
	var graph map[string]interface{}
	if err := json.Unmarshal([]byte(graphJSON), &graph); err == nil {
		sanitizeMap(graph)
		clean, err := json.Marshal(graph)
		if err != nil {
			return "{}"
		}
		return string(clean)
	}

	// Try as array
	var graphArr []interface{}
	if err := json.Unmarshal([]byte(graphJSON), &graphArr); err == nil {
		sanitizeSlice(graphArr)
		clean, err := json.Marshal(graphArr)
		if err != nil {
			return "{}"
		}
		return string(clean)
	}

	// Not valid JSON — return empty graph rather than leak data
	return "{}"
}

// sanitizeMap recursively walks a map and redacts sensitive values.
func sanitizeMap(m map[string]interface{}) {
	sensitiveKeys := map[string]bool{
		"api_key":      true,
		"apikey":       true,
		"apiKey":       true,
		"token":        true,
		"secret":       true,
		"credential":   true,
		"password":     true,
		"authorization": true,
	}

	for k, v := range m {
		if sensitiveKeys[strings.ToLower(k)] {
			m[k] = "[REDACTED]"
			continue
		}
		switch val := v.(type) {
		case map[string]interface{}:
			sanitizeMap(val)
		case []interface{}:
			sanitizeSlice(val)
		}
	}
}

// sanitizeSlice recursively walks a slice and redacts sensitive values.
func sanitizeSlice(s []interface{}) {
	for _, v := range s {
		switch val := v.(type) {
		case map[string]interface{}:
			sanitizeMap(val)
		case []interface{}:
			sanitizeSlice(val)
		}
	}
}

// generateREADME creates a session summary README.md for the export tarball.
func generateREADME(sess *Session) string {
	var sb strings.Builder

	sb.WriteString("# Session Export\n\n")
	fmt.Fprintf(&sb, "**Project:** %s\n", sess.Name)
	fmt.Fprintf(&sb, "**Session ID:** %s\n", sess.ID)
	fmt.Fprintf(&sb, "**Created:** %s\n", sess.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&sb, "**Last Active:** %s\n", sess.LastActiveAt.Format(time.RFC3339))
	fmt.Fprintf(&sb, "**Status:** %s\n\n", sess.Status)

	if sess.Goal != "" {
		fmt.Fprintf(&sb, "## Goal\n\n%s\n\n", sess.Goal)
	}

	sb.WriteString("## Contents\n\n")
	sb.WriteString("- `graph.json` — The node graph for this session\n")
	sb.WriteString("- `workspace/` — Source files produced during the session\n")
	sb.WriteString("- `README.md` — This file\n\n")

	sb.WriteString("## Notes\n\n")
	sb.WriteString("API keys, configuration files, and secrets have been excluded from this export.\n")

	return sb.String()
}

// ExportSession creates a self-contained tarball of the session including graph JSON,
// workspace files, and an auto-generated README.md. Secrets and API keys are excluded.
// Returns the actual output path used.
func ExportSession(ctx context.Context, mgr *Manager, sessionID, outputPath string) (string, error) {
	// Load session from database
	sess, err := mgr.GetSession(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to load session: %w", err)
	}

	// Get workspace path
	workspacePath, err := mgr.WorkspacePath(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	actualPath := outputPath
	if actualPath == "" {
		timestamp := time.Now().Format("20060102T150405")
		actualPath = fmt.Sprintf("session-%s-%s.tar.gz", sessionID, timestamp)
	}
	if !strings.HasSuffix(actualPath, ".tar.gz") && !strings.HasSuffix(actualPath, ".tgz") {
		actualPath += ".tar.gz"
	}

	outDir := filepath.Dir(actualPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	outFile, err := os.Create(actualPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := writeExportTar(ctx, sess, workspacePath, outFile); err != nil {
		os.Remove(actualPath)
		return "", err
	}

	return actualPath, nil
}

// ExportSessionToWriter writes the export tarball to the provided io.Writer.
// This is used by the API endpoint for streaming responses.
func ExportSessionToWriter(ctx context.Context, mgr *Manager, sessionID string, w io.Writer) error {
	sess, err := mgr.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	workspacePath, err := mgr.WorkspacePath(sessionID)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	return writeExportTar(ctx, sess, workspacePath, w)
}

// writeExportTar writes the tarball to the given writer.
func writeExportTar(ctx context.Context, sess *Session, workspacePath string, w io.Writer) error {
	gzWriter := gzip.NewWriter(w)
	tarWriter := tar.NewWriter(gzWriter)

	// 1. Add graph.json (sanitized)
	cleanGraph := sanitizeGraphJSON(sess.GraphJSON)
	if err := addStringToTar(tarWriter, "graph.json", cleanGraph, 0644); err != nil {
		return fmt.Errorf("failed to add graph.json: %w", err)
	}

	// 2. Add README.md
	readme := generateREADME(sess)
	if err := addStringToTar(tarWriter, "README.md", readme, 0644); err != nil {
		return fmt.Errorf("failed to add README.md: %w", err)
	}

	// 3. Add workspace files (excluding secrets)
	if err := addWorkspaceToTar(ctx, tarWriter, workspacePath, "workspace"); err != nil {
		return fmt.Errorf("failed to add workspace files: %w", err)
	}

	// Close writers in order, capturing errors
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return nil
}

// addStringToTar writes a string as a file entry in the tar archive.
func addStringToTar(tw *tar.Writer, name, content string, mode int64) error {
	header := &tar.Header{
		Name: name,
		Size: int64(len(content)),
		Mode: mode,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write([]byte(content))
	return err
}

// addWorkspaceToTar walks the workspace directory and adds all non-excluded files to the tar.
func addWorkspaceToTar(ctx context.Context, tw *tar.Writer, workspacePath, tarPrefix string) error {
	return filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		// Check for context cancellation
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if err != nil {
			return err
		}

		// Compute relative path from workspace root
		relPath, err := filepath.Rel(workspacePath, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Normalize to forward slashes for tar archive (cross-platform)
		relPath = filepath.ToSlash(relPath)

		// Check exclusion patterns
		if isExcluded(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		tarPath := tarPrefix + "/" + relPath

		if info.IsDir() {
			header := &tar.Header{
				Name:     tarPath + "/",
				Mode:     0755,
				Typeflag: tar.TypeDir,
			}
			return tw.WriteHeader(header)
		}

		// Stream file content instead of loading into memory
		file, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				// File was deleted during walk — skip
				return nil
			}
			return fmt.Errorf("failed to open %s: %w", path, err)
		}
		defer file.Close()

		header := &tar.Header{
			Name: tarPath,
			Size: info.Size(),
			Mode: 0644,
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		return err
	})
}

// ReadWorkspaceFileForExport reads a workspace file for export purposes (used by API streaming).
// This is a convenience wrapper that reads the file content without creating a tarball.
func (m *Manager) ReadWorkspaceFileForExport(sessionID, relativePath string) ([]byte, error) {
	return m.ReadWorkspaceFile(sessionID, relativePath)
}

// GetWorkspaceSize calculates the total size of the workspace directory in bytes.
func (m *Manager) GetWorkspaceSize(sessionID string) (int64, error) {
	workspacePath, err := m.WorkspacePath(sessionID)
	if err != nil {
		return 0, err
	}

	var totalSize int64
	err = filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize, err
}
