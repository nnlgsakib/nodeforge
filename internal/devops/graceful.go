package devops

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/nnlgsakib/nodeforge/internal/session"
)

// ShutdownHandler manages graceful shutdown of the server and all active sessions.
type ShutdownHandler struct {
	sessionMgr   *session.Manager
	shuttingDown atomic.Int32
}

// NewShutdownHandler creates a new shutdown handler.
func NewShutdownHandler(mgr *session.Manager) *ShutdownHandler {
	return &ShutdownHandler{sessionMgr: mgr}
}

// IsShuttingDown returns true if shutdown is in progress.
func (h *ShutdownHandler) IsShuttingDown() bool {
	return h.shuttingDown.Load() == 1
}

// WaitForSignal blocks until SIGINT or SIGTERM is received, then triggers shutdown.
// The provided onShutdown callback is called once before session snapshots are saved.
func (h *ShutdownHandler) WaitForSignal(onShutdown func()) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	fmt.Printf("\nReceived signal: %v\n", sig)

	h.shuttingDown.Store(1)

	if onShutdown != nil {
		onShutdown()
	}

	// Snapshot all active sessions
	if h.sessionMgr != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.sessionMgr.SnapshotAllSessions(); err != nil {
				fmt.Printf("Warning: failed to snapshot sessions: %v\n", err)
			}
		}()

		// Wait for snapshots with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			fmt.Println("All sessions snapshotted successfully")
		case <-ctx.Done():
			fmt.Println("Warning: session snapshot timed out")
		}
	}

	return nil
}
