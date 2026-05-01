package nforge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIsSupportedKey(t *testing.T) {
	assert.True(t, isSupportedKey("llm.openai-key"))
	assert.True(t, isSupportedKey("llm.anthropic-key"))
	assert.True(t, isSupportedKey("llm.ollama-url"))
	assert.True(t, isSupportedKey("server.port"))
	assert.True(t, isSupportedKey("llm.default-model"))
	assert.False(t, isSupportedKey("invalid-key"))
	assert.False(t, isSupportedKey("llm.openai-key.bad"))
}

func TestConfigSetAndGet(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "config.yaml")
	configPath = tmpConfig
	defer func() { configPath = getDefaultConfigPath() }()

	// Test set valid key
	err := setCmd.RunE(setCmd, []string{"llm.openai-key", "sk-test123"})
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(tmpConfig)
	assert.NoError(t, err)

	// Verify value via Viper
	viper.Reset()
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	assert.NoError(t, err)
	assert.Equal(t, "sk-test123", viper.GetString("llm.openai-key"))

	// Test get valid key
	// Capture output? Since getCmd prints to stdout, we can test via Viper
	val := viper.Get("llm.openai-key")
	assert.Equal(t, "sk-test123", val)

	// Test set server.port as integer
	err = setCmd.RunE(setCmd, []string{"server.port", "8080"})
	assert.NoError(t, err)

	viper.Reset()
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	assert.NoError(t, err)
	assert.Equal(t, 8080, viper.GetInt("server.port"))
}

func TestConfigSetInvalidKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath = filepath.Join(tmpDir, "config.yaml")
	defer func() { configPath = getDefaultConfigPath() }()

	err := setCmd.RunE(setCmd, []string{"invalid-key", "value"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config key")
}

func TestConfigGetNonExistentKey(t *testing.T) {
	tmpDir := t.TempDir()
	configPath = filepath.Join(tmpDir, "config.yaml")
	defer func() { configPath = getDefaultConfigPath() }()

	// Create empty config
	f, _ := os.Create(configPath)
	f.Close()

	err := getCmd.RunE(getCmd, []string{"llm.openai-key"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in config")
}

func TestConfigSetServerPortInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	configPath = filepath.Join(tmpDir, "config.yaml")
	defer func() { configPath = getDefaultConfigPath() }()

	err := setCmd.RunE(setCmd, []string{"server.port", "not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an integer")
}
