package nforge

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nforge",
	Short: "NodeForge - Visual AI Workflow Engine",
	Long:  `NodeForge (nfv2) is a visual AI workflow engine with n8n-style graphs, LLM integration, and a Go plugin system.`,
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file path")
}

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}
}

func Execute() error {
	return rootCmd.Execute()
}
