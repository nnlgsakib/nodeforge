package session

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HeartbeatConfig configures the heartbeat system.
type HeartbeatConfig struct {
	// Timeout is the duration after which a session is considered zombie.
	// Default: 5 minutes.
	Timeout time.Duration

	// CheckInterval is how often the zombie detection loop runs.
	// Default: 60 seconds.
	CheckInterval time.Duration
}

// DefaultHeartbeatConfig returns the default heartbeat configuration.
func DefaultHeartbeatConfig() HeartbeatConfig {
	return HeartbeatConfig{
		Timeout:       5 * time.Minute,
		CheckInterval: 60 * time.Second,
	}
}

// HeartbeatMonitor manages heartbeat tracking for active sessions.
type HeartbeatMonitor struct {
	mgr      *Manager
	cfg      HeartbeatConfig
	mu       sync.RWMutex
	stopCh   chan struct{}
	stopped  chan struct{}
	lastBeat map[string]time.Time // in-memory heartbeat map
	beatCh   chan string          // buffered channel for heartbeat writes
}

// NewHeartbeatMonitor creates a new heartbeat monitor.
func NewHeartbeatMonitor(mgr *Manager, cfg HeartbeatConfig) *HeartbeatMonitor {
	if cfg.Timeout == 0 {
		cfg = DefaultHeartbeatConfig()
	}
	mon := &HeartbeatMonitor{
		mgr:      mgr,
		cfg:      cfg,
		stopCh:   make(chan struct{}),
		stopped:  make(chan struct{}),
		lastBeat: make(map[string]time.Time),
		beatCh:   make(chan string, 64), // buffer to absorb bursts
	}
	// Start the dedicated writer goroutine
	go mon.persistLoop()
	return mon
}

// Start begins the heartbeat monitoring goroutines.
// It starts a zombie cleanup loop that runs every CheckInterval.
func (h *HeartbeatMonitor) Start(ctx context.Context) {
	go h.zombieCleanupLoop(ctx)
}

// Stop gracefully stops the heartbeat monitor.
func (h *HeartbeatMonitor) Stop() {
	close(h.stopCh)
	<-h.stopped
}

// Beat records a heartbeat for the given session ID.
// It updates the in-memory map synchronously and queues a DB write.
func (h *HeartbeatMonitor) Beat(sessionID string) {
	h.mu.Lock()
	h.lastBeat[sessionID] = time.Now().UTC()
	h.mu.Unlock()

	// Queue DB write (non-blocking with drop on full)
	select {
	case h.beatCh <- sessionID:
	default:
		// Channel full — DB is slow, skip this beat (next one will catch up)
	}
}

// persistLoop runs a single goroutine that batches and persists heartbeats to SQLite.
func (h *HeartbeatMonitor) persistLoop() {
	defer func() {
		// Drain remaining beats before exiting
		for sessionID := range h.beatCh {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = h.mgr.UpdateHeartbeat(ctx, sessionID)
			cancel()
		}
	}()

	for {
		select {
		case sessionID := <-h.beatCh:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if err := h.mgr.UpdateHeartbeat(ctx, sessionID); err != nil {
				fmt.Printf("Warning: heartbeat persist failed for %s: %v\n", sessionID, err)
			}
			cancel()
		case <-h.stopCh:
			close(h.beatCh)
			return
		}
	}
}

// zombieCleanupLoop runs periodically to find and clean up zombie sessions.
func (h *HeartbeatMonitor) zombieCleanupLoop(ctx context.Context) {
	defer close(h.stopped)

	ticker := time.NewTicker(h.cfg.CheckInterval)
	defer ticker.Stop()

	// Run immediately on first start
	h.runZombieCleanup(ctx)

	for {
		select {
		case <-ticker.C:
			h.runZombieCleanup(ctx)
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		}
	}
}

func (h *HeartbeatMonitor) runZombieCleanup(ctx context.Context) {
	zombieIDs, err := h.mgr.CleanupZombieSessions(ctx, h.cfg.Timeout)
	if err != nil {
		fmt.Printf("[heartbeat] zombie cleanup failed: %v\n", err)
		return
	}

	if len(zombieIDs) > 0 {
		fmt.Printf("[heartbeat] marked %d session(s) as zombie\n", len(zombieIDs))
	}

	// Clean up in-memory tracking for zombies and prune non-running sessions
	h.mu.Lock()
	for id := range h.lastBeat {
		// Remove zombies and any session not seen recently (prune stale entries)
		isZombie := false
		for _, zid := range zombieIDs {
			if id == zid {
				isZombie = true
				break
			}
		}
		if isZombie || time.Since(h.lastBeat[id]) > h.cfg.Timeout*2 {
			delete(h.lastBeat, id)
		}
	}
	h.mu.Unlock()
}

// GetLastBeat returns the last heartbeat time for a session.
func (h *HeartbeatMonitor) GetLastBeat(sessionID string) time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastBeat[sessionID]
}
