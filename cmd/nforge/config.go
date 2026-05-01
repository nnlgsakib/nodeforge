package nforge

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var supportedKeys = map[string]bool{
	"llm.openai-key":    true,
	"llm.anthropic-key": true,
	"llm.ollama-url":    true,
	"server.port":        true,
	"llm.default-model": true,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration (set/get values)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration key to a value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if !isSupportedKey(key) {
			return fmt.Errorf("unsupported config key: %s. Supported keys: %v", key, getSupportedKeys())
		}

		if value == "" {
			return fmt.Errorf("value cannot be empty for key: %s", key)
		}

		if err := validateConfigPath(configPath); err != nil {
			return err
		}

		// Initialize Viper
		viper.Reset()
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		// Create config directory if needed
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Read existing config if it exists
		if _, err := os.Stat(configPath); err == nil {
			if err := viper.ReadInConfig(); err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}
		}

		// Convert value to appropriate type
		var val interface{}
		switch key {
		case "server.port":
			i, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("server.port must be an integer: %w", err)
			}
			if i < 1 || i > 65535 {
				return fmt.Errorf("server.port must be a valid port (1-65535), got: %d", i)
			}
			val = i
		default:
			val = value
		}

		viper.Set(key, val)

		// Write config (creates file if it doesn't exist)
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Printf("Set %s = %v\n", key, val)
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		if !isSupportedKey(key) {
			return fmt.Errorf("unsupported config key: %s. Supported keys: %v", key, getSupportedKeys())
		}

		if err := validateConfigPath(configPath); err != nil {
			return err
		}

		// Initialize Viper
		viper.Reset()
		viper.SetConfigFile(configPath)
		viper.SetConfigType("yaml")

		// Check if config file exists
		if _, err := os.Stat(configPath); err != nil {
			return fmt.Errorf("config file not found at %s", configPath)
		}

		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		val := viper.Get(key)
		if val == nil {
			return fmt.Errorf("key %s not found in config", key)
		}

		fmt.Println(val)
		return nil
	},
}

func init() {
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(getCmd)
	rootCmd.AddCommand(configCmd)
}

func isSupportedKey(key string) bool {
	_, ok := supportedKeys[key]
	return ok
}

func getSupportedKeys() []string {
	keys := make([]string, 0, len(supportedKeys))
	for k := range supportedKeys {
		keys = append(keys, k)
	}
	return keys
}
