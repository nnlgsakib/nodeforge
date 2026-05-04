package canvas

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nnlgsakib/nodeforge/internal/session"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *session.Manager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpDir := t.TempDir()
	mgr, err := session.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create session manager: %v", err)
	}
	t.Cleanup(func() { mgr.Close() })

	RegisterAPIRoutes(r, mgr)
	return r, mgr
}

func TestCreateSession_Success(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := bytes.NewReader([]byte(`{"projectName":"test-project"}`))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.ProjectName != "test-project" {
		t.Errorf("expected projectName test-project, got %s", resp.ProjectName)
	}
	if resp.SessionID == "" {
		t.Error("expected non-empty session ID")
	}
}

func TestCreateSession_WithGoal(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := bytes.NewReader([]byte(`{"projectName":"test-project","goal":"Build a web app"}`))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Goal != "Build a web app" {
		t.Errorf("expected goal 'Build a web app', got %q", resp.Goal)
	}
}

func TestCreateSession_MissingProjectName(t *testing.T) {
	r, _ := setupTestRouter(t)

	body := bytes.NewReader([]byte(`{}`))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateSession_DuplicateProjectName(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create first project
	body1 := bytes.NewReader([]byte(`{"projectName":"dup-project"}`))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body1)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", w.Code)
	}

	// Try to create duplicate with fresh body
	body2 := bytes.NewReader([]byte(`{"projectName":"dup-project"}`))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/sessions", body2)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for duplicate, got %d: %s", w.Code, w.Body.String())
	}
}

func TestListSessions_Empty(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/sessions", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ListSessionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected empty list, got %d items", len(resp.Data))
	}
}

func TestListSessions_AfterCreate(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create two projects
	for _, name := range []string{"project-a", "project-b"} {
		body := bytes.NewReader([]byte(`{"projectName":"` + name + `"}`))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("create %s: expected 201, got %d", name, w.Code)
		}
	}

	// List sessions
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/sessions", nil)
	r.ServeHTTP(w, req)

	var resp ListSessionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(resp.Data))
	}
}

func TestGetSession_NotFound(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/sessions/nonexistent-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetSession_Success(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create a session first
	body := bytes.NewReader([]byte(`{"projectName":"get-test"}`))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var created SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	// Get the session
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/sessions/"+created.SessionID, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.SessionID != created.SessionID {
		t.Errorf("expected session ID %s, got %s", created.SessionID, resp.SessionID)
	}
}

func TestAutoSaveSession_UpdateGraph(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create session
	body := bytes.NewReader([]byte(`{"projectName":"autosave-test"}`))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var created SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	// Auto-save graph
	saveBody := bytes.NewReader([]byte(`{"graphJson":"{\"nodes\":[]}"}`))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/sessions/"+created.SessionID+"/auto-save", saveBody)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAutoSaveSession_ClearData(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create session with graph
	body := bytes.NewReader([]byte(`{"projectName":"clear-test"}`))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var created SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	// Save some data first
	saveBody := bytes.NewReader([]byte(`{"graphJson":"{\"nodes\":[1]}"}`))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/sessions/"+created.SessionID+"/auto-save", saveBody)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Clear with clearData flag
	clearBody := bytes.NewReader([]byte(`{"clearData":true}`))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/sessions/"+created.SessionID+"/auto-save", clearBody)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAutoSaveSession_InvalidStatus(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create session
	body := bytes.NewReader([]byte(`{"projectName":"status-test"}`))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var created SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	// Try invalid status
	saveBody := bytes.NewReader([]byte(`{"status":"invalid-status"}`))
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/sessions/"+created.SessionID+"/auto-save", saveBody)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid status, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAutoSaveSession_MissingID(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/sessions//auto-save", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Gin's routing depends on the route definition; with :id param, empty id returns 404
	// But our handler should catch it
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("expected 400 or 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMain(m *testing.M) {
	// Suppress gin debug output
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}
