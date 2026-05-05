package nforge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nnlgsakib/nodeforge/internal/session"
	"github.com/nnlgsakib/nodeforge/internal/skills"
	"github.com/spf13/cobra"
)

// skillRegistry is a simple in-memory registry of available skills.
// In production, this would be backed by a database or remote marketplace.
var skillRegistry = map[string]*skills.SkillManifest{
	"skill-code-review": {
		ID: "skill-code-review", Name: "Code Review", Version: "1.0.0",
		Description: "Automated code review with best practices",
		Author:      "NodeForge", Category: "Development", Rating: 4.5, RatingCount: 128,
		Downloads: 1200, Icon: "code-review", Tags: []string{"review", "quality"},
	},
	"skill-test-generator": {
		ID: "skill-test-generator", Name: "Test Generator", Version: "1.2.0",
		Description: "Generate unit and integration tests automatically",
		Author:      "NodeForge", Category: "Testing", Rating: 4.2, RatingCount: 95,
		Downloads: 980, Icon: "test-gen", Tags: []string{"testing", "automation"},
		Dependencies: []string{"skill-code-review"},
	},
	"skill-doc-writer": {
		ID: "skill-doc-writer", Name: "Documentation Writer", Version: "0.9.0",
		Description: "Auto-generate documentation from source code",
		Author:      "NodeForge", Category: "Documentation", Rating: 4.0, RatingCount: 64,
		Downloads: 540, Icon: "doc-writer", Tags: []string{"docs", "automation"},
	},
	"skill-refactor": {
		ID: "skill-refactor", Name: "Refactor Assistant", Version: "2.0.0",
		Description: "AI-powered code refactoring suggestions",
		Author:      "NodeForge", Category: "Development", Rating: 4.7, RatingCount: 210,
		Downloads: 1850, Icon: "refactor", Tags: []string{"refactor", "ai"},
		Dependencies: []string{"skill-code-review"},
	},
	"skill-security-scan": {
		ID: "skill-security-scan", Name: "Security Scanner", Version: "1.1.0",
		Description: "Scan code for common security vulnerabilities (OWASP Top 10)",
		Author:      "NodeForge", Category: "Security", Rating: 4.8, RatingCount: 312,
		Downloads: 2400, Icon: "security", Tags: []string{"security", "owasp"},
	},
	"skill-performance": {
		ID: "skill-performance", Name: "Performance Optimizer", Version: "1.0.0",
		Description: "Identify and fix performance bottlenecks",
		Author:      "NodeForge", Category: "Development", Rating: 4.3, RatingCount: 88,
		Downloads: 720, Icon: "performance", Tags: []string{"performance", "optimization"},
		Dependencies: []string{"skill-code-review"},
	},
}

var installedSkills = make(map[string]bool)
var installedMu sync.RWMutex
var abRunner = skills.NewABTestRunner()
var skillWSHub *wsHub // reference to WebSocket hub for broadcasting
var skillRegistryClient *skills.RegistryClient
var skillStore *skills.Store // SQLite-backed store for installed skills

func init() {
	// Register A/B test for code-review skill
	abRunner.RegisterTest(&skills.ABTestConfig{
		SkillID: "skill-code-review",
		Variants: []skills.ABTestVariant{
			{ID: "v1", Name: "Standard Review", Weight: 0.5},
			{ID: "v2", Name: "Deep Review", Weight: 0.5},
		},
	})

	// Pre-install some skills
	installedSkills["skill-code-review"] = true

	// Load skills from filesystem if internal/skills/ directory exists
	loadSkillsFromFS()

	// Initialize remote registry client
	registryURL := os.Getenv("NFORGE_SKILL_REGISTRY_URL")
	apiKey := os.Getenv("NFORGE_SKILL_REGISTRY_API_KEY")

	// Try loading API key from config file if not in env
	if apiKey == "" {
		if home, err := os.UserHomeDir(); err == nil {
			cfgPath := filepath.Join(home, ".nforge", "skill-config.json")
			if data, err := os.ReadFile(cfgPath); err == nil {
				var cfg struct {
					APIKey string `json:"apiKey"`
				}
				if json.Unmarshal(data, &cfg) == nil && cfg.APIKey != "" {
					apiKey = cfg.APIKey
				}
			}
		}
	}

	skillRegistryClient = skills.NewRegistryClient(registryURL, "")
	if apiKey != "" {
		skillRegistryClient.SetAPIKey(apiKey)
	}

	// Initialize SQLite store for installed skills
	var err error
	skillStore, err = skills.NewStore("")
	if err != nil {
		// Log warning but continue with in-memory fallback
		fmt.Printf("Warning: skills database unavailable (%v), using in-memory store\n", err)
	} else {
		// Migrate pre-installed skills to SQLite
		if !skillStore.IsInstalled("skill-code-review") {
			_ = skillStore.Insert("skill-code-review", "1.0.0")
		}
		// Sync in-memory map from SQLite for backward compatibility
		installedList, err := skillStore.List()
		if err == nil {
			installedMu.Lock()
			for _, sk := range installedList {
				installedSkills[sk.SkillID] = true
			}
			installedMu.Unlock()
		}
	}
}

// loadSkillsFromFS scans internal/skills/ for skill.json manifests and registers them.
func loadSkillsFromFS() {
	skillsDir := filepath.Join("internal", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		// Directory doesn't exist or isn't readable — skip, use in-memory registry
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		m, err := skills.LoadManifest(filepath.Join(skillsDir, entry.Name()))
		if err != nil {
			continue // skip invalid manifests
		}
		skillRegistry[m.ID] = m
	}
}

// SetWSHub sets the WebSocket hub reference for broadcasting skill install status.
func SetWSHub(h *wsHub) {
	skillWSHub = h
}

func initSkillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "skill",
		Short: "Manage skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Usage: nforge skill [list|install] <skill-id>")
			fmt.Println("  list     - List available skills")
			fmt.Println("  install  - Install a skill by ID")
			return nil
		},
	}
}

func init() {
	skillCmd := initSkillCmd()
	rootCmd.AddCommand(skillCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			searchFlag, _ := cmd.Flags().GetString("search")
			categoryFlag, _ := cmd.Flags().GetString("category")
			// Try fetching from registry client
			if skillRegistryClient != nil {
				sk, err := skillRegistryClient.FetchSkills(categoryFlag, searchFlag)
				if err == nil {
					for _, s := range sk {
						installedMu.RLock()
						installed := installedSkills[s.ID]
						installedMu.RUnlock()
						installedMark := ""
						if installed {
							installedMark = " (installed)"
						}
						fmt.Printf("  %s - %s v%s%s\n", s.Name, s.ID, s.Version, installedMark)
					}
					return nil
				}
				// Fallback to local if registry fails
				fmt.Printf("Warning: registry unavailable (%v), showing local skills\n", err)
			}

			// Fallback to local in-memory registry
			ids := make([]string, 0, len(skillRegistry))
			for id := range skillRegistry {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, id := range ids {
				s := skillRegistry[id]
				installed := ""
				if installedSkills[s.ID] {
					installed = " (installed)"
				}
				fmt.Printf("  %s - %s v%s%s\n", s.Name, s.ID, s.Version, installed)
			}
			return nil
		},
	}
	skillCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("search", "s", "", "Search skills by name or description")
	listCmd.Flags().StringP("category", "c", "", "Filter skills by category")

	installCmd := &cobra.Command{
		Use:   "install <skill-id>",
		Short: "Install a skill and its dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillID := args[0]

			fmt.Printf("Fetching manifest for %q...\n", skillID)

			depTree, err := skills.ResolveDependencies(skillID, func(id string) (*skills.SkillManifest, error) {
				// Try registry client first
				if skillRegistryClient != nil {
					s, err := skillRegistryClient.FetchSkill(id)
					if err == nil {
						return s, nil
					}
				}
				// Fallback to local registry
				s, ok := skillRegistry[id]
				if !ok {
					return nil, skills.ErrSkillNotFound
				}
				return s, nil
			})
			if err != nil {
				return fmt.Errorf("resolve dependencies: %w", err)
			}

			fmt.Printf("Installing dependencies:\n")
			for i, id := range depTree {
				fmt.Printf("  [%d/%d] %s\n", i+1, len(depTree), id)
			}

			fmt.Println("Installing skills...")
			// Pre-fetch versions outside the lock to avoid blocking reads during HTTP calls
			versions := make(map[string]string)
			for _, id := range depTree {
				if !installedSkills[id] {
					if m, err := skillRegistryClient.FetchSkill(id); err == nil && m != nil {
						versions[id] = m.Version
					} else if lm, ok := skillRegistry[id]; ok {
						versions[id] = lm.Version
					}
				}
			}
			installedMu.Lock()
			for _, id := range depTree {
				if !installedSkills[id] {
					installedSkills[id] = true
					// Persist to SQLite if available
					if skillStore != nil {
						_ = skillStore.Insert(id, versions[id])
					}
					fmt.Printf("  installed: %s\n", id)
				} else {
					fmt.Printf("  already installed: %s\n", id)
				}
			}
			installedMu.Unlock()
			fmt.Printf("Done. %d skill(s) installed.\n", len(depTree))
			return nil
		},
	}
	skillCmd.AddCommand(installCmd)
}

func registerSessionMonologueRoute(r *gin.Engine, hub *wsHub, sessionMgr *session.Manager) {
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
}

func registerSkillRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/skills")
	{
		api.GET("", listSkills)
		api.GET("/:id", getSkill)
		api.POST("/install", installSkill)
		api.GET("/config", getSkillConfig)
		api.POST("/config/apikey", setSkillAPIKey)
		api.GET("/abtest", getABTestMetrics)
		api.POST("/abtest/select", selectABTestVariant)
		api.POST("/abtest/metrics", recordABTestMetrics)
	}
}

// listSkills returns all available skills, fetching from the remote registry with fallback to local cache.
func listSkills(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("search")
	sortBy := c.Query("sort")     // rating|downloads|name
	refresh := c.Query("refresh") // "1" to bypass cache

	// Try fetching from remote registry client
	var allSkills []skills.SkillManifest
	var err error
	if skillRegistryClient != nil {
		if refresh == "1" {
			allSkills, err = skillRegistryClient.FetchSkillsFresh(category, search)
		} else {
			allSkills, err = skillRegistryClient.FetchSkills(category, search)
		}
	} else {
		err = fmt.Errorf("registry client not initialized")
	}
	if err != nil || len(allSkills) == 0 {
		if err != nil {
			fmt.Printf("Warning: registry client failed (%v), falling back to local skills\n", err)
		}
		// Fallback to local in-memory registry
		allSkills = getLocalSkills(category, search)
	}

	// Ensure we never return nil slice (would serialize as null)
	if allSkills == nil {
		allSkills = []skills.SkillManifest{}
	}

	result := buildSkillResponse(allSkills)

	// Apply sorting
	sortSkills(result, sortBy)

	c.JSON(http.StatusOK, gin.H{"skills": result})
}

// getLocalSkills returns skills from the in-memory registry as fallback.
func getLocalSkills(category, search string) []skills.SkillManifest {
	installedMu.RLock()
	defer installedMu.RUnlock()

	result := make([]skills.SkillManifest, 0)
	for _, s := range skillRegistry {
		if category != "" && s.Category != category {
			continue
		}
		if search != "" {
			q := search
			match := strings.Contains(strings.ToLower(s.Name), strings.ToLower(q)) ||
				strings.Contains(strings.ToLower(s.Description), strings.ToLower(q))
			for _, t := range s.Tags {
				if strings.Contains(strings.ToLower(t), strings.ToLower(q)) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		result = append(result, *s)
	}
	return result
}

// buildSkillResponse converts SkillManifest slice to API response format.
func buildSkillResponse(sk []skills.SkillManifest) []gin.H {
	result := make([]gin.H, 0, len(sk))
	for _, s := range sk {
		deps := s.Dependencies
		if deps == nil {
			deps = []string{}
		}
		tags := s.Tags
		if tags == nil {
			tags = []string{}
		}
		installedMu.RLock()
		installed := installedSkills[s.ID]
		installedMu.RUnlock()

		result = append(result, gin.H{
			"id":           s.ID,
			"name":         s.Name,
			"version":      s.Version,
			"description":  s.Description,
			"author":       s.Author,
			"category":     s.Category,
			"rating":       s.Rating,
			"ratingCount":  s.RatingCount,
			"downloads":    s.Downloads,
			"icon":         s.Icon,
			"tags":         tags,
			"dependencies": deps,
			"installed":    installed,
		})
	}
	return result
}

// sortSkills applies sorting to the response based on the sort parameter.
func sortSkills(result []gin.H, sortBy string) {
	switch sortBy {
	case "rating":
		sort.Slice(result, func(i, j int) bool {
			ri, _ := result[i]["rating"].(float64)
			rj, _ := result[j]["rating"].(float64)
			return ri > rj
		})
	case "downloads", "installs":
		sort.Slice(result, func(i, j int) bool {
			di, _ := result[i]["downloads"].(int)
			dj, _ := result[j]["downloads"].(int)
			return di > dj
		})
	default: // name
		sort.Slice(result, func(i, j int) bool {
			ni, _ := result[i]["name"].(string)
			nj, _ := result[j]["name"].(string)
			return ni < nj
		})
	}
}

// getSkill returns a single skill by ID.
func getSkill(c *gin.Context) {
	skillID := c.Param("id")
	if skillID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing skill ID"})
		return
	}

	// Try fetching from remote registry client
	var skill *skills.SkillManifest
	var fetchErr error
	if skillRegistryClient != nil {
		skill, fetchErr = skillRegistryClient.FetchSkill(skillID)
	} else {
		fetchErr = fmt.Errorf("registry client not initialized")
	}
	if fetchErr == nil {
		deps := skill.Dependencies
		if deps == nil {
			deps = []string{}
		}
		installedMu.RLock()
		installed := installedSkills[skill.ID]
		installedMu.RUnlock()

		c.JSON(http.StatusOK, gin.H{
			"id":           skill.ID,
			"name":         skill.Name,
			"version":      skill.Version,
			"description":  skill.Description,
			"author":       skill.Author,
			"category":     skill.Category,
			"rating":       skill.Rating,
			"ratingCount":  skill.RatingCount,
			"downloads":    skill.Downloads,
			"icon":         skill.Icon,
			"tags":         skill.Tags,
			"dependencies": deps,
			"installed":    installed,
		})
		return
	}

	// Fallback to local registry
	installedMu.RLock()
	s, ok := skillRegistry[skillID]
	installed := installedSkills[skillID]
	installedMu.RUnlock()

	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}

	deps := s.Dependencies
	if deps == nil {
		deps = []string{}
	}
	c.JSON(http.StatusOK, gin.H{
		"id":           s.ID,
		"name":         s.Name,
		"version":      s.Version,
		"description":  s.Description,
		"author":       s.Author,
		"category":     s.Category,
		"rating":       s.Rating,
		"ratingCount":  s.RatingCount,
		"downloads":    s.Downloads,
		"icon":         s.Icon,
		"tags":         s.Tags,
		"dependencies": deps,
		"installed":    installed,
	})
}

// installSkill installs a skill and its dependencies.
func installSkill(c *gin.Context) {
	var req struct {
		SkillID string `json:"skillId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "skillId is required"})
		return
	}
	if strings.TrimSpace(req.SkillID) == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "skillId must not be empty"})
		return
	}

	// Resolve dependency tree using registry client
	depTree, err := skills.ResolveDependencies(req.SkillID, func(id string) (*skills.SkillManifest, error) {
		// Try registry client first
		if skillRegistryClient != nil {
			s, err := skillRegistryClient.FetchSkill(id)
			if err == nil {
				return s, nil
			}
		}
		// Fallback to local registry
		s, ok := skillRegistry[id]
		if !ok {
			return nil, fmt.Errorf("skill %q not found in registry", id)
		}
		return s, nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	installed := []string{}
	// Pre-fetch versions outside the lock to avoid blocking reads during HTTP calls
	versions := make(map[string]string)
	installedMu.RLock()
	for _, id := range depTree {
		if !installedSkills[id] {
			if m, err := skillRegistryClient.FetchSkill(id); err == nil && m != nil {
				versions[id] = m.Version
			} else if lm, ok := skillRegistry[id]; ok {
				versions[id] = lm.Version
			}
		}
	}
	installedMu.RUnlock()

	installedMu.Lock()
	for _, id := range depTree {
		if !installedSkills[id] {
			installedSkills[id] = true
			// Persist to SQLite if available
			if skillStore != nil {
				_ = skillStore.Insert(id, versions[id])
			}
			installed = append(installed, id)
		}
	}
	installedMu.Unlock()

	// Broadcast install status via WebSocket hub
	if skillWSHub != nil {
		for _, id := range installed {
			skillWSHub.broadcastSkillInstalled(id)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "skill installed successfully",
		"installed": installed,
		"total":     len(installed),
	})
}

// getABTestMetrics returns A/B test metrics for Prometheus scraping.
func getABTestMetrics(c *gin.Context) {
	skillID := c.Query("skillId")
	if skillID == "" {
		// Return all metrics for all registered tests
		result := make(map[string]map[string]*skills.ABTestMetrics)
		for id := range abRunner.GetAllTests() {
			result[id] = abRunner.GetMetrics(id)
		}
		c.JSON(http.StatusOK, gin.H{"metrics": result})
		return
	}

	metrics := abRunner.GetMetrics(skillID)
	if len(metrics) == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no metrics for skill " + skillID})
		return
	}
	c.JSON(http.StatusOK, gin.H{"skillId": skillID, "metrics": metrics})
}

// selectABTestVariant routes a request to a variant for A/B testing (FR46).
func selectABTestVariant(c *gin.Context) {
	var req struct {
		SkillID string `json:"skillId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "skillId is required"})
		return
	}

	variantID := abRunner.SelectVariant(req.SkillID)
	if variantID == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no A/B test configured for skill " + req.SkillID})
		return
	}

	c.JSON(http.StatusOK, gin.H{"skillId": req.SkillID, "variant": variantID})
}

// recordABTestMetrics records metrics for an A/B test variant.
func recordABTestMetrics(c *gin.Context) {
	var req struct {
		SkillID   string  `json:"skillId" binding:"required"`
		VariantID string  `json:"variantId" binding:"required"`
		Success   bool    `json:"success"`
		Duration  float64 `json:"durationMs"`
		Tokens    int     `json:"tokens"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	abRunner.RecordMetrics(req.SkillID, req.VariantID, req.Success, req.Duration, req.Tokens)
	c.JSON(http.StatusOK, gin.H{"message": "metrics recorded"})
}

// getSkillConfig returns the current skill registry configuration (API key masked).
func getSkillConfig(c *gin.Context) {
	apiKey := ""
	if skillRegistryClient != nil {
		apiKey = skillRegistryClient.GetAPIKey()
	}
	c.JSON(http.StatusOK, gin.H{
		"apiKey": apiKey,
	})
}

// setSkillAPIKey saves the SkillsMP API key for authenticated registry access.
func setSkillAPIKey(c *gin.Context) {
	var req struct {
		APIKey string `json:"apiKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "apiKey is required"})
		return
	}

	if skillRegistryClient == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "registry client not initialized"})
		return
	}

	skillRegistryClient.SetAPIKey(req.APIKey)

	// Persist to config file
	cfgPath := filepath.Join(os.Getenv("HOME"), ".nforge", "skill-config.json")
	if home, err := os.UserHomeDir(); err == nil {
		cfgPath = filepath.Join(home, ".nforge", "skill-config.json")
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o700); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create config directory"})
		return
	}
	cfgData, err := json.Marshal(gin.H{"apiKey": req.APIKey})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal config"})
		return
	}
	if err := os.WriteFile(cfgPath, cfgData, 0o600); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save API key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key saved successfully"})
}
