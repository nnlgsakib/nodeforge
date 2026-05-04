package canvas

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nnlgsakib/nodeforge/internal/session"
)

// SessionRequest represents the request body for creating a session
type SessionRequest struct {
	ProjectName string `json:"projectName" binding:"required"`
	Goal        string `json:"goal"`
}

// AutoSaveRequest represents the request body for updating session state
type AutoSaveRequest struct {
	GraphJSON *string `json:"graphJson"`
	ChatLog   *string `json:"chatLog"`
	Status    string  `json:"status,omitempty"`
	ClearData bool    `json:"clearData"` // When true, null/empty pointer fields are treated as "clear"
}

// SessionResponse represents the response for a created session
type SessionResponse struct {
	SessionID   string `json:"sessionId"`
	ProjectName string `json:"projectName"`
	Status      string `json:"status"`
	Workspace   string `json:"workspace"`
	Goal        string `json:"goal"`
	CreatedAt   string `json:"createdAt"`
	LastActive  string `json:"lastActive"`
}

// ListSessionsResponse represents the response for listing sessions
type ListSessionsResponse struct {
	Data []SessionResponse `json:"data"`
}

// RegisterAPIRoutes registers all API routes for the canvas
func RegisterAPIRoutes(r *gin.Engine, sessionMgr *session.Manager) {
	api := r.Group("/api/v1")
	{
		api.POST("/sessions", createSession(sessionMgr))
		api.GET("/sessions", listSessions(sessionMgr))
		api.GET("/sessions/:id", getSession(sessionMgr))
		api.PUT("/sessions/:id/auto-save", autoSaveSession(sessionMgr))
	}
}

func createSession(mgr *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		ctx := context.Background()
		sess, err := mgr.CreateSessionWithName(ctx, req.ProjectName)
		if err != nil {
			log.Printf("createSession error: %v", err)
			if err.Error() == "project name cannot be empty" || err.Error() == "project name must not contain path separators" {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
			}
			return
		}

		// If goal is provided, update it atomically
		if req.Goal != "" {
			sess.Goal = req.Goal
			if err := mgr.UpdateSession(ctx, sess); err != nil {
				log.Printf("failed to save goal for session %s: %v", sess.ID, err)
				// Don't fail the creation — goal is supplementary
			}
		}

		c.JSON(http.StatusCreated, SessionResponse{
			SessionID:   sess.ID,
			ProjectName: sess.Name,
			Status:      string(sess.Status),
			Workspace:   sess.Workspace,
			Goal:        sess.Goal,
			CreatedAt:   sess.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastActive:  sess.LastActiveAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func listSessions(mgr *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		sessions, err := mgr.ListSessions(ctx)
		if err != nil {
			log.Printf("listSessions error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sessions"})
			return
		}

		resp := make([]SessionResponse, 0, len(sessions))
		for _, s := range sessions {
			resp = append(resp, SessionResponse{
				SessionID:   s.ID,
				ProjectName: s.Name,
				Status:      string(s.Status),
				Workspace:   s.Workspace,
				Goal:        s.Goal,
				CreatedAt:   s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				LastActive:  s.LastActiveAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		c.JSON(http.StatusOK, ListSessionsResponse{Data: resp})
	}
}

func getSession(mgr *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing session ID"})
			return
		}

		ctx := context.Background()
		sess, err := mgr.GetSession(ctx, id)
		if err != nil {
			if err.Error() == "session database not available" {
				log.Printf("getSession error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "session service unavailable"})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			}
			return
		}

		c.JSON(http.StatusOK, SessionResponse{
			SessionID:   sess.ID,
			ProjectName: sess.Name,
			Status:      string(sess.Status),
			Workspace:   sess.Workspace,
			Goal:        sess.Goal,
			CreatedAt:   sess.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastActive:  sess.LastActiveAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func autoSaveSession(mgr *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing session ID"})
			return
		}

		var req AutoSaveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		ctx := context.Background()

		// Verify session exists
		_, err := mgr.GetSession(ctx, id)
		if err != nil {
			if err.Error() == "session database not available" {
				log.Printf("autoSaveSession error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "session service unavailable"})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			}
			return
		}

		// Handle clearing: if ClearData is true, nil pointers mean "set to empty"
		graphJSON := req.GraphJSON
		chatLog := req.ChatLog
		if req.ClearData {
			emptyGraph := ""
			emptyChat := ""
			if graphJSON == nil {
				graphJSON = &emptyGraph
			}
			if chatLog == nil {
				chatLog = &emptyChat
			}
		}

		var status *session.SessionStatus
		if req.Status != "" {
			s := session.SessionStatus(req.Status)
			switch s {
			case session.StatusRunning, session.StatusComplete, session.StatusFailed, session.StatusPaused:
				status = &s
			default:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
				return
			}
		}

		// Atomically update session state
		if err := mgr.UpdateSessionState(ctx, id, graphJSON, chatLog, status); err != nil {
			log.Printf("autoSaveSession update error: %v", err)
			if err.Error() == "session not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save session state"})
			}
			return
		}

		// Return updated session
		sess, err := mgr.GetSession(ctx, id)
		if err != nil {
			log.Printf("autoSaveSession get error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updated session"})
			return
		}

		c.JSON(http.StatusOK, SessionResponse{
			SessionID:   sess.ID,
			ProjectName: sess.Name,
			Status:      string(sess.Status),
			Workspace:   sess.Workspace,
			Goal:        sess.Goal,
			CreatedAt:   sess.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastActive:  sess.LastActiveAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}
