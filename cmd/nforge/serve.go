package nforge

import (
	"context"
	"embed"
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
	"github.com/nnlgsakib/nodeforge/internal/session"
	"github.com/spf13/cobra"
)

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

	// WebSocket upgrader (initialized here after flag parsing to avoid data race)
	websocketUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == "" || origin == "http://localhost:5173" || origin == "http://localhost:"+servePort
		},
	}

	// Track active WebSocket connections for graceful shutdown
	type connEntry struct {
		conn *websocket.Conn
		done chan struct{}
	}
	var (
		connsMu sync.Mutex
		conns   = make(map[*websocket.Conn]chan struct{})
	)

	addConn := func(conn *websocket.Conn, done chan struct{}) {
		connsMu.Lock()
		conns[conn] = done
		connsMu.Unlock()
	}
	removeConn := func(conn *websocket.Conn) {
		connsMu.Lock()
		delete(conns, conn)
		connsMu.Unlock()
	}

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		conn, err := websocketUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		done := make(chan struct{})
		addConn(conn, done)
		defer removeConn(conn)
		defer conn.Close()

		// Set initial read deadline and pong handler for keepalive
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// Send periodic pings to keep connection alive
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
					return
				}
			default:
				_, _, err := conn.ReadMessage()
				if err != nil {
					// Echo close frame per WebSocket spec
					if websocket.IsCloseError(err, websocket.CloseNoStatusReceived) || strings.Contains(err.Error(), "close") {
						conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
					}
					return
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
		c.String(200, "# Prometheus metrics placeholder - full implementation in Story 6.5\n")
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
		// Signal all tracked WebSocket connections to exit
		connsMu.Lock()
		for conn, done := range conns {
			conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Second))
			close(done)
			conn.Close()
		}
		connsMu.Unlock()

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
