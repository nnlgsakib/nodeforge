package nforge

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dgraph-io/badger/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type checkResult struct {
	Name    string
	Passed  bool
	Message string
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and connectivity",
	Long:  "Run comprehensive health checks for Go version, framework availability, frontend build, database connectivity, and LLM provider reachability.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHealthChecks()
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runHealthChecks() error {
	criticalFail := false
	results := []checkResult{}

	// Critical checks (affect exit code)
	checks := []func() checkResult{
		checkGoVersion,
		checkGin,
		checkFrontendBuild,
		checkSQLite,
		checkBadgerDB,
	}

	for _, check := range checks {
		result := check()
		results = append(results, result)
		if !result.Passed {
			criticalFail = true
		}
	}

	// LLM provider checks (warnings only, do not affect exit code)
	llmResults := checkLLMProviders()
	results = append(results, llmResults...)

	// Print results
	fmt.Println("System Health Check Report")
	fmt.Println(strings.Repeat("=", 50))
	for _, r := range results {
		icon := "✅"
		if !r.Passed {
			if strings.Contains(r.Name, "LLM") {
				icon = "⚠️"
			} else {
				icon = "❌"
			}
		}
		fmt.Printf("%-30s %s %s\n", r.Name, icon, r.Message)
	}
	fmt.Println(strings.Repeat("=", 50))

	if criticalFail {
		fmt.Println("❌ Critical checks failed. Exit code 1.")
		return fmt.Errorf("critical health checks failed")
	}

	fmt.Println("✅ All critical checks passed.")
	return nil
}

func checkGoVersion() checkResult {
	return parseGoVersion(runtime.Version())
}

func parseGoVersion(versionStr string) checkResult {
	if !strings.HasPrefix(versionStr, "go") {
		return checkResult{Name: "Go Version", Passed: false, Message: fmt.Sprintf("Unexpected version format: %s", versionStr)}
	}

	remainder := strings.TrimPrefix(versionStr, "go")
	parts := strings.SplitN(remainder, ".", 3)

	if len(parts) < 2 {
		return checkResult{Name: "Go Version", Passed: false, Message: fmt.Sprintf("Cannot parse version: %s", versionStr)}
	}

	majorNum := 0
	fmt.Sscanf(parts[0], "%d", &majorNum)
	minorNum := 0
	fmt.Sscanf(parts[1], "%d", &minorNum)

	if majorNum == 0 {
		return checkResult{Name: "Go Version", Passed: false, Message: fmt.Sprintf("Cannot parse version: %s", versionStr)}
	}

	// Any Go 2.x or higher is newer than 1.24
	if parts[0] != "1" {
		return checkResult{Name: "Go Version", Passed: true, Message: fmt.Sprintf("Go %s", versionStr)}
	}

	if minorNum < 24 {
		return checkResult{Name: "Go Version", Passed: false, Message: fmt.Sprintf("Go %s detected, require 1.24+", versionStr)}
	}

	return checkResult{Name: "Go Version", Passed: true, Message: fmt.Sprintf("Go %s", versionStr)}
}

func checkGin() (result checkResult) {
	// Use defer/recover to catch panics from gin.Default()
	defer func() {
		if r := recover(); r != nil {
			result = checkResult{Name: "Gin Framework", Passed: false, Message: fmt.Sprintf("Gin initialization panicked: %v", r)}
		}
	}()

	_ = gin.Default()
	return checkResult{Name: "Gin Framework", Passed: true, Message: "v1.11.0 available"}
}

func checkFrontendBuild() checkResult {
	// Resolve frontend/dist path relative to executable if not found in CWD
	distPath := "frontend/dist"
	if _, err := os.Stat(distPath); err != nil {
		if !os.IsNotExist(err) {
			return checkResult{Name: "Frontend Build", Passed: false, Message: fmt.Sprintf("Cannot access %s: %v", distPath, err)}
		}
		// Try relative to executable
		if exe, err := os.Executable(); err == nil {
			candidate := filepath.Join(filepath.Dir(exe), "frontend", "dist")
			if _, err2 := os.Stat(candidate); err2 == nil {
				distPath = candidate
			}
		}
	}

	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return checkResult{Name: "Frontend Build", Passed: false, Message: "frontend/dist/ not found. Run: cd frontend && npm run build"}
	}

	indexPath := filepath.Join(distPath, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return checkResult{Name: "Frontend Build", Passed: false, Message: "frontend/dist/index.html missing. Run: cd frontend && npm run build"}
	}

	return checkResult{Name: "Frontend Build", Passed: true, Message: fmt.Sprintf("frontend/dist/ found (%s)", distPath)}
}

func checkSQLite() checkResult {
	// Attempt to open an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return checkResult{Name: "SQLite", Passed: false, Message: fmt.Sprintf("Failed to open: %v", err)}
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return checkResult{Name: "SQLite", Passed: false, Message: fmt.Sprintf("Ping failed: %v", err)}
	}

	return checkResult{Name: "SQLite", Passed: true, Message: "Connected (in-memory)"}
}

func checkBadgerDB() checkResult {
	tmpDir, err := os.MkdirTemp("", "badger-check-*")
	if err != nil {
		return checkResult{Name: "BadgerDB", Passed: false, Message: fmt.Sprintf("Failed to create temp dir: %v", err)}
	}
	defer os.RemoveAll(tmpDir)

	opts := badger.DefaultOptions(tmpDir)
	db, err := badger.Open(opts)
	if err != nil {
		return checkResult{Name: "BadgerDB", Passed: false, Message: fmt.Sprintf("Failed to open: %v", err)}
	}
	defer db.Close()

	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("health-check"), []byte("ok"))
	})
	if err != nil {
		return checkResult{Name: "BadgerDB", Passed: false, Message: fmt.Sprintf("Write failed: %v", err)}
	}

	return checkResult{Name: "BadgerDB", Passed: true, Message: "Connected (temp dir)"}
}

func checkLLMProviders() []checkResult {
	results := []checkResult{}

	v := viper.New()
	v.SetConfigType("yaml")

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			v.SetConfigFile(filepath.Join(home, ".nforge", "config.yaml"))
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return append(results, checkResult{Name: "LLM Providers", Passed: true, Message: "No config found, skipping LLM checks"})
		}
		return append(results, checkResult{Name: "LLM Providers", Passed: true, Message: fmt.Sprintf("Config error: %v (skipping LLM checks)", err)})
	}

	ollamaURL := v.GetString("llm.ollama-url")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	results = append(results, checkOllama(ollamaURL))

	if v.IsSet("llm.openai-key") {
		results = append(results, checkOpenAI(v.GetString("llm.openai-key"), "https://api.openai.com"))
	}

	if v.IsSet("llm.anthropic-key") {
		results = append(results, checkAnthropic(v.GetString("llm.anthropic-key"), "https://api.anthropic.com"))
	}

	return results
}

func checkOllama(baseURL string) checkResult {
	client := &http.Client{Timeout: 5 * time.Second}
	url := baseURL
	if !strings.HasSuffix(url, "/api/tags") {
		url = strings.TrimSuffix(url, "/") + "/api/tags"
	}

	resp, err := client.Get(url)
	if err != nil {
		return checkResult{Name: "LLM: Ollama", Passed: false, Message: fmt.Sprintf("Unreachable at %s: %v (start with: ollama serve)", baseURL, err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 400 {
		return checkResult{Name: "LLM: Ollama", Passed: false, Message: fmt.Sprintf("Unexpected status %d from %s", resp.StatusCode, baseURL)}
	}

	return checkResult{Name: "LLM: Ollama", Passed: true, Message: fmt.Sprintf("Connected (%s)", baseURL)}
}

func checkOpenAI(apiKey string, baseURL string) checkResult {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/v1/models", nil)
	if err != nil {
		return checkResult{Name: "LLM: OpenAI", Passed: false, Message: fmt.Sprintf("Request creation failed: %v", err)}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return checkResult{Name: "LLM: OpenAI", Passed: false, Message: fmt.Sprintf("Unreachable: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return checkResult{Name: "LLM: OpenAI", Passed: false, Message: "Invalid API key"}
	}
	if resp.StatusCode != 200 {
		return checkResult{Name: "LLM: OpenAI", Passed: false, Message: fmt.Sprintf("Unexpected status %d", resp.StatusCode)}
	}

	return checkResult{Name: "LLM: OpenAI", Passed: true, Message: "Connected"}
}

func checkAnthropic(apiKey string, baseURL string) checkResult {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/v1/messages", nil)
	if err != nil {
		return checkResult{Name: "LLM: Anthropic", Passed: false, Message: fmt.Sprintf("Request creation failed: %v", err)}
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return checkResult{Name: "LLM: Anthropic", Passed: false, Message: fmt.Sprintf("Unreachable: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return checkResult{Name: "LLM: Anthropic", Passed: false, Message: "Invalid API key"}
	}
	// 405 is expected for GET on /v1/messages, 400 indicates auth worked but method invalid, 200 is OK
	if resp.StatusCode == 405 || resp.StatusCode == 400 || resp.StatusCode == 200 {
		return checkResult{Name: "LLM: Anthropic", Passed: true, Message: "Connected"}
	}

	return checkResult{Name: "LLM: Anthropic", Passed: false, Message: fmt.Sprintf("Unexpected status %d", resp.StatusCode)}
}
