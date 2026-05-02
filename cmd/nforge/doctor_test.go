package nforge

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGoVersion(t *testing.T) {
	tests := []struct {
		name       string
		versionStr string
		expectPass bool
	}{
		{"go1.24", "go1.24", true},
		{"go1.26.2", "go1.26.2", true},
		{"go2.0", "go2.0", true},
		{"go1.23", "go1.23", false},
		{"go1.22.5", "go1.22.5", false},
		{"invalid", "invalid", false},
		{"go", "go", false},
		{"go1.24rc1", "go1.24rc1", true},
		{"go1.23beta", "go1.23beta", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGoVersion(tt.versionStr)
			assert.Equal(t, tt.expectPass, result.Passed, "version string: %s", tt.versionStr)
		})
	}
}

func TestCheckSQLite(t *testing.T) {
	result := checkSQLite()
	assert.True(t, result.Passed, "SQLite check should pass with in-memory DB")
}

func TestCheckBadgerDB(t *testing.T) {
	result := checkBadgerDB()
	if !result.Passed {
		if strings.Contains(result.Message, "not enough space") || strings.Contains(result.Message, "Insufficient disk space") {
			t.Skip("Skipping BadgerDB test: insufficient disk space")
		}
	}
	assert.True(t, result.Passed, "BadgerDB check should pass with temp dir, got: "+result.Message)
}

func TestCheckGin(t *testing.T) {
	result := checkGin()
	assert.True(t, result.Passed, "Gin check should pass when gin is available")
	assert.Contains(t, result.Message, "v1.11.0")
}

func TestCheckFrontendBuild(t *testing.T) {
	result := checkFrontendBuild()
	// We don't know if frontend/dist exists in the test environment
	// Just verify the function runs without panic
	if !result.Passed {
		assert.Contains(t, result.Message, "not found")
	}
}

func TestCheckOllama(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := checkOllama(server.URL)
	assert.True(t, result.Passed, "Ollama check should pass with mock server")
	assert.Contains(t, result.Message, "Connected")
}

func TestCheckOllamaNotRunning(t *testing.T) {
	result := checkOllama("http://localhost:19999")
	assert.False(t, result.Passed, "Ollama check should fail when not running")
	assert.Contains(t, result.Message, "Unreachable")
}

func TestCheckOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := checkOpenAI("sk-test-key", server.URL)
	assert.True(t, result.Passed, "OpenAI check should pass with valid key and mock server")
}

func TestCheckOpenAIInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	result := checkOpenAI("invalid-key", server.URL)
	assert.False(t, result.Passed, "OpenAI check should fail with invalid key")
	assert.Contains(t, result.Message, "Invalid API key")
}

func TestCheckAnthropic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// GET on /v1/messages returns 405
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	result := checkAnthropic("sk-test-key", server.URL)
	// 405 is treated as success (connectivity confirmed)
	assert.True(t, result.Passed, "Anthropic check should pass with 405 (method not allowed expected)")
}
