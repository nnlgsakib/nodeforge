package nforge

import (
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

	installCmd := &cobra.Command{
		Use:   "install <skill-id>",
		Short: "Install a skill and its dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillID := args[0]
			depTree, err := skills.ResolveDependencies(skillID, func(id string) (*skills.SkillManifest, error) {
				s, ok := skillRegistry[id]
				if !ok {
					return nil, skills.ErrSkillNotFound
				}
				return s, nil
			})
			if err != nil {
				return fmt.Errorf("resolve dependencies: %w", err)
			}
			installedMu.Lock()
			for _, id := range depTree {
				if !installedSkills[id] {
					installedSkills[id] = true
					fmt.Printf("  installed: %s\n", id)
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
		api.POST("/install", installSkill)
		api.GET("/abtest", getABTestMetrics)
		api.POST("/abtest/select", selectABTestVariant)
		api.POST("/abtest/metrics", recordABTestMetrics)
	}
}

// listSkills returns all available skills from the registry.
func listSkills(c *gin.Context) {
	category := c.Query("category")

	installedMu.RLock()
	var result []gin.H
	for _, s := range skillRegistry {
		if category != "" && s.Category != category {
			continue
		}
		installed := installedSkills[s.ID]
		deps := s.Dependencies
		if deps == nil {
			deps = []string{}
		}
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
			"tags":         s.Tags,
			"dependencies": deps,
			"installed":    installed,
		})
	}
	installedMu.RUnlock()

	// Sort by name for stable response
	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})

	c.JSON(http.StatusOK, gin.H{"skills": result})
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

	// Resolve dependency tree
	depTree, err := skills.ResolveDependencies(req.SkillID, func(id string) (*skills.SkillManifest, error) {
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
	installedMu.Lock()
	for _, id := range depTree {
		if !installedSkills[id] {
			installedSkills[id] = true
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
