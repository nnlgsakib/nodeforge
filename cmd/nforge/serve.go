package nforge

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nnlgsakib/nodeforge/internal/canvas"
	nfcontext "github.com/nnlgsakib/nodeforge/internal/context"
	"github.com/nnlgsakib/nodeforge/internal/engine"
	"github.com/nnlgsakib/nodeforge/internal/llm"
	"github.com/nnlgsakib/nodeforge/internal/session"
	"github.com/spf13/cobra"
)

// wsClient represents a WebSocket client connection
type wsClient struct {
	hub    *wsHub
	conn   *websocket.Conn
	send   chan []byte
	mu     sync.Mutex
}

// wsHub manages all WebSocket clients and message broadcasting
type wsHub struct {
	clients        map[*wsClient]bool
	broadcast      chan []byte
	register       chan *wsClient
	unregister     chan *wsClient
	graphGen       *engine.Generator
	store          *nfcontext.Store
	statusChecker  *llm.StatusChecker
	mu             sync.RWMutex
	clientCount    atomic.Int64 // Story 2.7: Track active connections for monitoring
}

// newWSHub creates a new WebSocket hub
func newWSHub() *wsHub {
	return &wsHub{
		clients:    make(map[*wsClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *wsClient),
		unregister: make(chan *wsClient),
	}
}

// run starts the hub's message processing loop
func (h *wsHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if !h.clients[client] {
				h.clients[client] = true
				h.clientCount.Add(1)
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if h.clients[client] {
				delete(h.clients, client)
				close(client.send)
				h.clientCount.Add(-1)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Drop message if client can't keep up (latency <50ms requirement)
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// broadcastGraphUpdate sends a graph update to all connected clients
func (h *wsHub) broadcastGraphUpdate(graph *engine.Graph) {
	data, err := json.Marshal(map[string]interface{}{
		"type":   "graph_update",
		"nodes":  graph.Nodes,
		"edges":  graph.Edges,
		"goal":   graph.Goal,
	})
	if err != nil {
		return
	}
	h.broadcast <- data
}

// BroadcastNodeUpdate sends a node status update to all clients
func (h *wsHub) BroadcastNodeUpdate(nodeID, status string, progress float64) {
	data, _ := json.Marshal(map[string]interface{}{
		"type":     "node_update",
		"nodeId":   nodeID,
		"status":   status,
		"progress": progress,
	})
	h.broadcast <- data
}

// BroadcastEdgeUpdate sends an edge update to all clients
func (h *wsHub) BroadcastEdgeUpdate(source, target string, tension float64) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    "edge_update",
		"source":  source,
		"target":  target,
		"tension": tension,
	})
	if err != nil {
		return
	}
	h.broadcast <- data
}

// BroadcastRaw sends raw bytes to all connected clients
func (h *wsHub) BroadcastRaw(data []byte) {
	if len(data) == 0 {
		return
	}
	h.broadcast <- data
}

// ClientCount returns the number of active WebSocket connections (Story 2.7)
func (h *wsHub) ClientCount() int64 {
	return h.clientCount.Load()
}

// broadcastSkillInstalled broadcasts a skill installation success event to all clients.
func (h *wsHub) broadcastSkillInstalled(skillID string) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    "skill_installed",
		"skillId": skillID,
	})
	if err != nil {
		return
	}
	h.broadcast <- data
}

// broadcastSkillInstallFailed broadcasts a skill installation failure to all clients.
func (h *wsHub) broadcastSkillInstallFailed(skillID, message string) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    "skill_install_failed",
		"skillId": skillID,
		"message": message,
	})
	if err != nil {
		return
	}
	h.broadcast <- data
}

var (
	servePort   string
	distFS      embed.FS
	frontendFS  fs.FS
	shuttingDown int32 // accessed via sync/atomic
)

func SetDistFS(embeddedFS embed.FS) error {
	distFS = embeddedFS
	subFS, err := fs.Sub(distFS, "frontend/dist")
	if err != nil {
		return fmt.Errorf("SetDistFS: failed to create sub-FS: %w", err)
	}
	frontendFS = subFS
	return nil
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the NodeForge web UI + API server",
	Long:  "Starts the Gin server with REST API, WebSocket hub, and embedded React frontend",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

func init() {
	serveCmd.Flags().StringVar(&servePort, "port", "8080", "Port to run the server on (default: 8080)")
	serveCmd.RegisterFlagCompletionFunc("port", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"8080", "9090", "3000"}, cobra.ShellCompDirectiveNoFileComp
	})
}

func validatePort(port string) bool {
	if port == "" {
		return false
	}
	for _, c := range port {
		if c < '0' || c > '9' {
			return false
		}
	}
	// Check valid port range 1-65535
	portNum := 0
	for _, c := range port {
		portNum = portNum*10 + int(c-'0')
	}
	return portNum >= 1 && portNum <= 65535
}

func runServer() error {
	startTime := time.Now()

	// Validate port
	if !validatePort(servePort) {
		return fmt.Errorf("invalid port: %s (must be 1-65535)", servePort)
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)
	if verboseMode {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Gin router with middleware
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// CORS middleware — allow localhost origins and empty Origin (non-browser clients)
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" || origin == "http://localhost:5173" || origin == "http://localhost:"+servePort {
			if origin != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Initialize session manager and register API routes
	sessionMgr := session.NewManager(".")
	canvas.RegisterAPIRoutes(r, sessionMgr)
	registerSkillRoutes(r)

	// Initialize WebSocket hub, graph generator, and context store
	hub := newWSHub()
	SetWSHub(hub)
	go hub.run()

	// Initialize context store (BadgerDB)
	store, err := nfcontext.NewStore(nfcontext.DefaultStorePath("."))
	if err != nil {
		fmt.Printf("Warning: failed to initialize context store: %v\n", err)
	}
	if store != nil {
		defer store.Close()
		hub.store = store
	}

	// Initialize LLM providers
	var providers []llm.LLMProvider

	ollamaCfg := &llm.ProviderConfig{
		Type:     llm.ProviderOllama,
		BaseURL:  "http://localhost:11434",
		Model:    "llama3",
		Timeout:  30 * time.Second,
	}
	if ollamaProv, err := llm.NewProvider(ollamaCfg); err == nil {
		providers = append(providers, ollamaProv)
	}

	// TODO: Initialize other providers from config (OpenAI, Anthropic, etc.)

	// Initialize status checker and start background checks
	if len(providers) > 0 {
		hub.statusChecker = llm.NewStatusChecker(providers)
		// Use a detached context — status checks are best-effort and should not block shutdown
		go hub.statusChecker.CheckAndBroadcast(context.Background(), func(data []byte) {
			hub.broadcast <- data
		})
	}

	// Set graph generator if Ollama is available
	if len(providers) > 0 {
		hub.graphGen = engine.NewGenerator(providers[0], store)
	}

	// WebSocket upgrader
	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "" || origin == "http://localhost:5173" || origin == "http://localhost:"+servePort
		},
	}

	// WebSocket endpoint with hub integration
	r.GET("/ws", func(c *gin.Context) {
		conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &wsClient{hub: hub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		// Send initial connection confirmation
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"connected"}`))

		// Send provider status if checker is available
		if client.hub.statusChecker != nil {
			statusData, err := client.hub.statusChecker.FormatWebSocketMessage(context.Background())
			if err == nil && statusData != nil {
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				conn.WriteMessage(websocket.TextMessage, statusData)
			}
		}

		// Start write pump
		go func() {
			defer func() {
				conn.Close()
				client.hub.unregister <- client
			}()
			for msg := range client.send {
				conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond)) // <50ms latency
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			}
		}()

		// Read pump: handle incoming messages
		defer func() {
			client.hub.unregister <- client
			conn.Close()
		}()

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
					return
				}
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsCloseError(err, websocket.CloseNoStatusReceived) || strings.Contains(err.Error(), "close") {
						conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
					}
					return
				}

				// Parse incoming message
				var wsMsg struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}
				if err := json.Unmarshal(msg, &wsMsg); err != nil {
					continue
				}

				// Handle goal message: generate graph
				if wsMsg.Type == "goal" && hub.graphGen != nil {
					graph, err := hub.graphGen.Generate(c.Request.Context(), wsMsg.Text)
					if err == nil {
						// Save graph to BadgerDB
						if hub.store != nil {
							hub.store.SaveGraph(c.Request.Context(), graph.ID, graph)
						}
						// Broadcast graph update to all clients
						hub.broadcastGraphUpdate(graph)
					}
				}
			}
		}
	})

	// Health check endpoint
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   version,
		})
	})

	// Readiness check
	r.GET("/readyz", func(c *gin.Context) {
		if atomic.LoadInt32(&shuttingDown) == 1 {
			c.JSON(503, gin.H{
				"status": "not ready",
				"checks": gin.H{"shutdown": "in progress"},
			})
			return
		}
		status := "ready"
		httpStatus := 200
		checks := gin.H{}
		checks["gin_server"] = "up"
		if frontendFS != nil {
			checks["frontend_embedded"] = "ok"
		} else {
			checks["frontend_embedded"] = "missing"
			status = "not ready"
			httpStatus = 503
		}
		c.JSON(httpStatus, gin.H{
			"status": status,
			"checks": checks,
		})
	})

	// Liveness check
	r.GET("/livez", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "alive",
			"uptime":    time.Since(startTime).String(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Metrics endpoint (Prometheus placeholder)
	r.GET("/metrics", func(c *gin.Context) {
		c.String(200, "# Prometheus metrics placeholder - full implementation in Story 6.5\nws_connections_total %d\n", hub.ClientCount())
	})

	// Get monologue history for a session
	r.GET("/api/v1/sessions/:id/monologue", func(c *gin.Context) {
		sessionID := c.Param("id")
		if sessionID == "" {
			c.JSON(400, gin.H{"error": "missing session ID"})
			return
		}
		if hub.store == nil {
			c.JSON(500, gin.H{"error": "store not available"})
			return
		}
		messages, err := hub.store.GetMonologueHistory(c.Request.Context(), sessionID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to get monologue history"})
			return
		}
		c.JSON(200, messages)
	})

	// Serve embedded frontend
	if frontendFS != nil {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if path == "/ws" || (len(path) >= 7 && path[:7] == "/api/v1") {
				c.Next()
				return
			}

			// Sanitize path: reject any path with ".." (literal or URL-encoded) to prevent traversal
			filename := strings.TrimPrefix(path, "/")
			decoded, err := url.PathUnescape(filename)
			if err != nil {
				decoded = filename
			}
			if strings.Contains(decoded, "..") {
				c.String(400, "Bad request")
				return
			}
			if strings.Contains(filename, "..") {
				c.String(400, "Bad request")
				return
			}

			if filename == "" {
				filename = "index.html"
			}
			f, err := frontendFS.Open(filename)
			if err == nil {
				defer f.Close()
				info, err := f.Stat()
				if err != nil {
					c.String(500, "Internal server error")
					return
				}
				if !info.IsDir() {
					data, err := io.ReadAll(f)
					if err != nil {
						c.String(500, "Internal server error")
						return
					}
					contentType := "application/octet-stream"
					ext := strings.ToLower(filename)
					switch {
					case strings.HasSuffix(ext, ".html"):
						contentType = "text/html"
					case strings.HasSuffix(ext, ".css"):
						contentType = "text/css"
					case strings.HasSuffix(ext, ".js"):
						contentType = "application/javascript"
					case strings.HasSuffix(ext, ".ico"):
						contentType = "image/x-icon"
					case strings.HasSuffix(ext, ".png"):
						contentType = "image/png"
					case strings.HasSuffix(ext, ".jpg"), strings.HasSuffix(ext, ".jpeg"):
						contentType = "image/jpeg"
					case strings.HasSuffix(ext, ".svg"):
						contentType = "image/svg+xml"
					case strings.HasSuffix(ext, ".json"):
						contentType = "application/json"
					case strings.HasSuffix(ext, ".woff"):
						contentType = "font/woff"
					case strings.HasSuffix(ext, ".woff2"):
						contentType = "font/woff2"
					case strings.HasSuffix(ext, ".ttf"):
						contentType = "font/ttf"
					}
					c.Data(200, contentType, data)
					return
				}
			}

			// Fallback to index.html for SPA routing
			indexFile, err := frontendFS.Open("index.html")
			if err == nil {
				defer indexFile.Close()
				data, err := io.ReadAll(indexFile)
				if err != nil {
					c.String(500, "Internal server error")
					return
				}
				c.Header("Content-Type", "text/html")
				c.Data(200, "text/html", data)
				return
			}

			c.String(404, "Frontend not found. Build the frontend first with npm run build.")
		})
	} else {
		r.NoRoute(func(c *gin.Context) {
			c.String(503, "Frontend not built. Run 'npm run build' in the frontend directory.")
		})
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + servePort,
		Handler: r,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("Starting server on :%s\n", servePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
	}()

	select {
	case <-quit:
		fmt.Println("Shutting down server...")
		atomic.StoreInt32(&shuttingDown, 1)
		// Signal all hub-connected WebSocket clients to exit
		hub.mu.Lock()
		for client := range hub.clients {
			client.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
			close(client.send)
			client.conn.Close()
		}
		hub.mu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}
		fmt.Println("Server exited")
		return nil
	case err := <-errCh:
		return fmt.Errorf("server failed to start: %w", err)
	}
}
