package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultRegistryURL  = "https://skillsmp.com/api/v1/skills/search"
	defaultCacheTTL     = 5 * time.Minute
	defaultTimeout      = 10 * time.Second
	defaultMaxRetries   = 3
	defaultCacheDirName = "skills-cache"
	defaultFetchLimit   = 100
)

// SkillsMPResponse is the response format from the SkillsMP API.
type SkillsMPResponse struct {
	Success bool              `json:"success"`
	Data    *SkillsMPData     `json:"data,omitempty"`
	Error   *SkillsMPAPIError `json:"error,omitempty"`
}

// SkillsMPData wraps the skills array in the API response.
type SkillsMPData struct {
	Skills []SkillsMPSkill `json:"skills"`
}

// SkillsMPSkill represents a skill from the SkillsMP API response.
type SkillsMPSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	CategorySlug string `json:"categorySlug"`
	Stars       float64  `json:"stars"`
	Downloads   int      `json:"downloads"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags"`
	Icon        string   `json:"icon,omitempty"`
}

// SkillsMPAPIError represents an error from the SkillsMP API.
type SkillsMPAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RegistryConfig holds configuration for the registry client.
type RegistryConfig struct {
	BaseURL string
	APIKey  string
	CacheDir string
}

// RegistryClient fetches skill manifests from the SkillsMP API.
type RegistryClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	cache      *registryCache
	cacheDir   string
}

// registryCache holds in-memory cached skills with TTL.
type registryCache struct {
	mu        sync.RWMutex
	skills    []SkillManifest
	expiresAt time.Time
}

// NewRegistryClient creates a new registry client with caching and retry logic.
func NewRegistryClient(baseURL, cacheDir string) *RegistryClient {
	return NewRegistryClientWithConfig(RegistryConfig{
		BaseURL:  baseURL,
		CacheDir: cacheDir,
	})
}

// NewRegistryClientWithConfig creates a registry client with full configuration including API key.
func NewRegistryClientWithConfig(cfg RegistryConfig) *RegistryClient {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultRegistryURL
	}
	if cfg.CacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		cfg.CacheDir = filepath.Join(home, ".nforge", defaultCacheDirName)
	}

	// Resolve API key from config, env var, or leave empty for anonymous access
	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("NFORGE_SKILL_REGISTRY_API_KEY")
	}

	return &RegistryClient{
		baseURL: cfg.BaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		cache:    &registryCache{},
		cacheDir: cfg.CacheDir,
	}
}

// SetAPIKey updates the API key for authenticated requests.
func (c *RegistryClient) SetAPIKey(key string) {
	c.apiKey = key
}

// GetAPIKey returns the current API key (masked for display).
func (c *RegistryClient) GetAPIKey() string {
	if c.apiKey == "" {
		return ""
	}
	// Show first 4 and last 4 chars for verification
	if len(c.apiKey) > 12 {
		return c.apiKey[:4] + "..." + c.apiKey[len(c.apiKey)-4:]
	}
	return "****"
}

// FetchSkills fetches skills from the SkillsMP API with optional category and search filters.
// When search is non-empty, always hits the API directly (no cache short-circuit).
// When both search and category are empty, checks cache first for the initial load.
func (c *RegistryClient) FetchSkills(category, search string) ([]SkillManifest, error) {
	return c.fetchSkills(category, search, false)
}

// FetchSkillsFresh fetches skills from the API, bypassing all caches.
// Use this for explicit user-initiated refresh actions.
func (c *RegistryClient) FetchSkillsFresh(category, search string) ([]SkillManifest, error) {
	return c.fetchSkills(category, search, true)
}

func (c *RegistryClient) fetchSkills(category, search string, forceRefresh bool) ([]SkillManifest, error) {
	// Only check cache on initial load (no search, no category, no force refresh)
	if !forceRefresh && search == "" && category == "" {
		if skills := c.cache.get(); skills != nil {
			return skills, nil
		}
	}

	// Always call API for search or category filter (server-side filtering is more comprehensive)
	skills, err := c.fetchFromAPI(search, category)
	if err == nil {
		c.cache.set(skills)
		_ = c.saveLocalCache(skills)
		if search != "" || category != "" {
			return filterSkills(skills, category, search), nil
		}
		return skills, nil
	}
	// Fallback to local cache if API fails
	if skills, err := c.loadLocalCache(); err == nil {
		if search != "" || category != "" {
			return filterSkills(skills, category, search), nil
		}
		return skills, nil
	}
	return nil, fmt.Errorf("skills: failed to fetch skills from registry and local cache: %w", err)
}

// FetchSkill fetches a single skill by ID from the registry.
func (c *RegistryClient) FetchSkill(skillID string) (*SkillManifest, error) {
	// Call API directly with skill ID as search term
	skills, err := c.fetchFromAPI(skillID, "")
	if err == nil {
		for _, s := range skills {
			if s.ID == skillID {
				return &s, nil
			}
		}
		return nil, ErrSkillNotFound
	}
	
	// Fallback to local cache
	if skills, err := c.loadLocalCache(); err == nil {
		for _, s := range skills {
			if s.ID == skillID {
				return &s, nil
			}
		}
	}
	
	return nil, ErrSkillNotFound
}

// fetchFromAPI does the actual HTTP call with retry logic.
func (c *RegistryClient) fetchFromAPI(search, category string) ([]SkillManifest, error) {
	var lastErr error
	for attempt := 1; attempt <= defaultMaxRetries; attempt++ {
		skills, err := c.doFetch(search, category)
		if err == nil {
			return skills, nil
		}
		lastErr = err
		// Brief backoff before retry
		time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
	}
	return nil, fmt.Errorf("skills: %d attempts failed: %w", defaultMaxRetries, lastErr)
}

// doFetch performs a single HTTP request to the SkillsMP API.
func (c *RegistryClient) doFetch(search, category string) ([]SkillManifest, error) {
	// Build query parameters
	params := url.Values{}
	// SkillsMP API requires "q" parameter - use empty string to list all
	params.Set("q", search)
	params.Set("limit", strconv.Itoa(defaultFetchLimit))
	params.Set("sortBy", "stars")
	if category != "" {
		params.Set("category", category)
	}

	// Properly merge query params with baseURL (handles baseURL that already has ?)
	reqURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("skills: parse baseURL %q: %w", c.baseURL, err)
	}
	existing := reqURL.Query()
	for k, v := range params {
		existing[k] = v
	}
	reqURL.RawQuery = existing.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("skills: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	// Add API key if available
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("skills: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		// Parse error response if possible
		var apiErr SkillsMPResponse
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != nil {
			return nil, fmt.Errorf("skills: registry error %s: %s", apiErr.Error.Code, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("skills: registry returned %d: %s", resp.StatusCode, string(body))
	}

	var result SkillsMPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("skills: decode response: %w", err)
	}

	if !result.Success {
		if result.Error != nil {
			return nil, fmt.Errorf("skills: registry error %s: %s", result.Error.Code, result.Error.Message)
		}
		return nil, fmt.Errorf("skills: registry returned success=false")
	}

	// Convert SkillsMP skills to internal SkillManifest format
	if result.Data == nil {
		return []SkillManifest{}, nil
	}

	skills := make([]SkillManifest, 0, len(result.Data.Skills))
	for _, s := range result.Data.Skills {
		skills = append(skills, SkillManifest{
			ID:          s.ID,
			Name:        s.Name,
			Version:     s.Version,
			Description: s.Description,
			Author:      s.Author,
			Category:    s.Category,
			Rating:      s.Stars,
			RatingCount: 0, // Not provided by API
			Downloads:   s.Downloads,
			Icon:        s.Icon,
			Tags:        s.Tags,
		})
	}

	return skills, nil
}

// filterSkills applies category and search filters to a skill list.
func filterSkills(skills []SkillManifest, category, search string) []SkillManifest {
	if category == "" && search == "" {
		return skills
	}

	result := make([]SkillManifest, 0)
	for _, s := range skills {
		if category != "" && s.Category != category {
			continue
		}
		if search != "" {
			q := search
			match := containsIgnoreCase(s.Name, q) ||
				containsIgnoreCase(s.Description, q) ||
				containsAnyIgnoreCase(s.Tags, q)
			if !match {
				continue
			}
		}
		result = append(result, s)
	}
	return result
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func containsAnyIgnoreCase(tags []string, substr string) bool {
	for _, t := range tags {
		if containsIgnoreCase(t, substr) {
			return true
		}
	}
	return false
}

// get returns cached skills if still valid (within TTL).
func (rc *registryCache) get() []SkillManifest {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	if time.Now().Before(rc.expiresAt) && rc.skills != nil {
		cp := make([]SkillManifest, len(rc.skills))
		for i, s := range rc.skills {
			// Deep copy slice fields to prevent shared mutation
			if s.Tags != nil {
				s.Tags = append([]string(nil), s.Tags...)
			}
			if s.Dependencies != nil {
				s.Dependencies = append([]string(nil), s.Dependencies...)
			}
			cp[i] = s
		}
		return cp
	}
	return nil
}

// set stores skills in cache with TTL.
func (rc *registryCache) set(skills []SkillManifest) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.skills = skills
	rc.expiresAt = time.Now().Add(defaultCacheTTL)
}

// saveLocalCache persists skills to a JSON file for offline fallback.
// Uses atomic write (temp file + rename) to prevent corrupt JSON on concurrent writes.
func (c *RegistryClient) saveLocalCache(skills []SkillManifest) error {
	if err := os.MkdirAll(c.cacheDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(c.cacheDir, "skills.json")
	data, err := json.Marshal(skills)
	if err != nil {
		return err
	}
	// Atomic write: write to temp file, then rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// loadLocalCache loads skills from the local JSON cache file.
func (c *RegistryClient) loadLocalCache() ([]SkillManifest, error) {
	path := filepath.Join(c.cacheDir, "skills.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var skills []SkillManifest
	if err := json.Unmarshal(data, &skills); err != nil {
		return nil, err
	}
	if skills == nil {
		skills = []SkillManifest{}
	}
	return skills, nil
}
