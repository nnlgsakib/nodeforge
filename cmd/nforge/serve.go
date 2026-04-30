package nforge

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var distFS embed.FS

func SetDistFS(fs embed.FS) {
	distFS = fs
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the NodeForge server",
	Run: func(cmd *cobra.Command, args []string) {
		r := gin.Default()

		RegisterRoutes(r)

		dist, err := fs.Sub(distFS, "frontend/dist")
		if err != nil {
			panic(err)
		}
		r.StaticFS("/assets", http.FS(dist))

		r.NoRoute(func(c *gin.Context) {
			data, err := distFS.ReadFile("frontend/dist/index.html")
			if err != nil {
				c.AbortWithStatus(404)
				return
			}
			c.Data(200, "text/html; charset=utf-8", data)
		})

		r.Run(":8080")
	},
}

func init() {
	serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}
