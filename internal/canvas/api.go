package canvas

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nnlgsakib/nodeforge/internal/session"
)

// SessionRequest represents the request body for creating a session
type SessionRequest struct {
	ProjectName string `json:"projectName" binding:"required"`
}

// SessionResponse represents the response for a created session
type SessionResponse struct {
	SessionID   string `json:"sessionId"`
	ProjectName string `json:"projectName"`
	Workspace   string `json:"workspace"`
}

// RegisterAPIRoutes registers all API routes for the canvas
func RegisterAPIRoutes(r *gin.Engine, sessionMgr *session.Manager) {
	api := r.Group("/api/v1")
	{
		api.POST("/sessions", createSession(sessionMgr))
	}
}

func createSession(mgr *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}

		ctx := context.Background()
		sess, err := mgr.CreateSessionWithName(ctx, req.ProjectName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, SessionResponse{
			SessionID:   sess.ID,
			ProjectName: sess.Name,
			Workspace:   sess.Workspace,
		})
	}
}
