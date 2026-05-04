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

// NodeTypes defines the valid node types for completion
var NodeTypes = []string{"Goal", "Spec", "Plan", "Implement", "Test", "Review"}

// NodeTypeDescriptions maps node types to their descriptions
var NodeTypeDescriptions = map[string]string{
	"Goal":      "Top-level goal node",
	"Spec":      "Specification node",
	"Plan":      "Planning node",
	"Implement": "Implementation node",
	"Test":      "Testing node",
	"Review":    "Review node",
}

var rootCmd = &cobra.Command{
	Use:     "nforge",
	Short:   "NodeForge OS - Spec-driven development workbench",
	Version: version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: false,
	},
}

func init() {
	// Persistent flags
	rootCmd.PersistentFlags().BoolVar(&verboseMode, "verbose", getEnvBool("NFORGE_VERBOSE", false), "Enable debug logging (env: NFORGE_VERBOSE)")
	rootCmd.PersistentFlags().StringVar(&configPath, "config-path", getDefaultConfigPath(), "Config file path (env: NFORGE_CONFIG)")

	// Register flag completions
	rootCmd.RegisterFlagCompletionFunc("config-path", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault // Default file/directory completion
	})

	// Initialize and customize completion command
	rootCmd.InitDefaultCompletionCmd()
	// Hide the completion subcommand from help (AC1 specifies exact 8 subcommands)
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "completion" {
			cmd.Hidden = true
			cmd.Long = `Generate shell completion scripts for nforge.

Installation instructions:
  Bash:
    source <(nforge completion bash)
    # Or add to ~/.bashrc:
    echo 'source <(nforge completion bash)' >> ~/.bashrc

  Zsh:
    source <(nforge completion zsh)
    # Or add to ~/.zshrc:
    echo 'source <(nforge completion zsh)' >> ~/.zshrc

  PowerShell:
    . (nforge completion powershell)
    # Or add to your profile:
    Add-Content $PROFILE ". (nforge completion powershell)"
`
			break
		}
	}

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
