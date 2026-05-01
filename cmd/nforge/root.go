package nforge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var version = "dev" // Set via ldflags at build time
var verboseMode bool
var configPath string

var rootCmd = &cobra.Command{
	Use:   "nforge",
	Short: "NodeForge OS - Spec-driven development workbench",
	Version: version,
}

func init() {
	// Persistent flags
	rootCmd.PersistentFlags().BoolVar(&verboseMode, "verbose", getEnvBool("NFORGE_VERBOSE", false), "Enable debug logging (env: NFORGE_VERBOSE)")
	rootCmd.PersistentFlags().StringVar(&configPath, "config-path", getDefaultConfigPath(), "Config file path (env: NFORGE_CONFIG)")

	// Register subcommands
	rootCmd.AddCommand(serveCmd)
}

func getDefaultConfigPath() string {
	if v := os.Getenv("NFORGE_CONFIG"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".nforge", "config.yaml")
}

func validateConfigPath(p string) error {
	if p == "" {
		return fmt.Errorf("config path not set and cannot determine home directory")
	}
	if info, err := os.Stat(p); err == nil && info.IsDir() {
		return fmt.Errorf("config path is a directory, expected a file path: %s", p)
	}
	return nil
}

func getEnvBool(key string, defaultVal bool) bool {
	v := strings.ToLower(os.Getenv(key))
	if v == "" {
		return defaultVal
	}
	return v == "true" || v == "1" || v == "yes"
}

func Execute() error {
	return rootCmd.Execute()
}
