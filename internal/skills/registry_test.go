package skills

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryClient_FetchSkills_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer test-api-key", auth)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[
			{"id":"skill-a","name":"Test Skill A","slug":"skill-a","description":"test desc","category":"dev","categorySlug":"dev","stars":4.5,"downloads":100,"author":"test","version":"1.0.0","icon":"a","tags":["test"]},
			{"id":"skill-b","name":"Test Skill B","slug":"skill-b","description":"another test","category":"test","categorySlug":"test","stars":3.0,"downloads":50,"author":"test","version":"2.0.0","icon":"b","tags":["other"]}
		]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	})
	skills, err := client.FetchSkills("", "test")
	require.NoError(t, err)
	assert.Len(t, skills, 2)
	assert.Equal(t, "skill-a", skills[0].ID)
	assert.Equal(t, "skill-b", skills[1].ID)
	assert.Equal(t, 4.5, skills[0].Rating)
}

func TestRegistryClient_FetchSkills_Caching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[{"id":"s1","name":"Skill One","slug":"s1","description":"a test skill","category":"c","categorySlug":"c","stars":4.0,"downloads":1,"author":"a","version":"1.0.0","icon":"i","tags":[]}]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
	})

	// Search always hits API
	_, err := client.FetchSkills("", "skill")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second search call also hits API (no cache short-circuit for search)
	_, err = client.FetchSkills("", "skill")
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestRegistryClient_FetchSkills_InitialLoadUsesCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[{"id":"s1","name":"Skill One","slug":"s1","description":"d","category":"c","categorySlug":"c","stars":4.0,"downloads":1,"author":"a","version":"1.0.0","icon":"i","tags":[]}]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
	})

	// First call (no params) hits API and caches
	_, err := client.FetchSkills("", "")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// Second call (no params) uses cache
	_, err = client.FetchSkills("", "")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount) // no additional HTTP call
}

func TestRegistryClient_FetchSkillsFresh_BypassesCache(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[{"id":"s1","name":"Skill One","slug":"s1","description":"d","category":"c","categorySlug":"c","stars":4.0,"downloads":1,"author":"a","version":"1.0.0","icon":"i","tags":[]}]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
	})

	// First call populates cache
	_, err := client.FetchSkills("", "")
	require.NoError(t, err)
	assert.Equal(t, 1, callCount)

	// FetchSkillsFresh bypasses cache even with no params
	_, err = client.FetchSkillsFresh("", "")
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestRegistryClient_FetchSkills_FilterCategory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify category query parameter
		category := r.URL.Query().Get("category")
		assert.Equal(t, "dev", category)

		w.Header().Set("Content-Type", "application/json")
		// Return only the dev category skill
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[
			{"id":"s1","name":"S1","slug":"s1","description":"d","category":"dev","categorySlug":"dev","stars":4.0,"downloads":1,"author":"a","version":"1.0.0","icon":"i","tags":[]}
		]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
	})
	skills, err := client.FetchSkills("dev", "")
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "s1", skills[0].ID)
}

func TestRegistryClient_FetchSkill_ByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[
			{"id":"target-skill","name":"Target","slug":"target-skill","description":"found","category":"c","categorySlug":"c","stars":4.0,"downloads":1,"author":"a","version":"1.0.0","icon":"i","tags":[]}
		]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
	})
	skill, err := client.FetchSkill("target-skill")
	require.NoError(t, err)
	assert.Equal(t, "target-skill", skill.ID)
	assert.Equal(t, "found", skill.Description)
}

func TestRegistryClient_FetchSkill_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
	})
	_, err := client.FetchSkill("nonexistent")
	assert.ErrorIs(t, err, ErrSkillNotFound)
}

func TestRegistryClient_FetchSkills_FallbackToLocalCache(t *testing.T) {
	// Create a local cache file
	tmpDir := t.TempDir()
	cacheData := `[{"id":"cached-skill","name":"Cached","version":"1.0.0","description":"from cache","author":"a","category":"c","rating":4.0,"ratingCount":1,"downloads":1,"icon":"i","tags":[]}]`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "skills.json"), []byte(cacheData), 0o644))

	// Create a client pointing to a non-existent server (will fail immediately)
	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL:  "http://127.0.0.1:1",
		CacheDir: tmpDir,
	})
	// Force cache expiration so it tries the server
	client.cache.expiresAt = time.Now().Add(-1 * time.Hour)

	skills, err := client.FetchSkills("", "cached")
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "cached-skill", skills[0].ID)
}

func TestRegistryClient_FetchSkills_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
	})
	skills, err := client.FetchSkills("", "")
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestRegistryClient_SetAPIKey(t *testing.T) {
	client := NewRegistryClient("", "")
	assert.Equal(t, "", client.GetAPIKey())

	client.SetAPIKey("sk_live_abcdef1234567890")
	assert.Equal(t, "sk_l...7890", client.GetAPIKey())

	client.SetAPIKey("")
	assert.Equal(t, "", client.GetAPIKey())
}

func TestRegistryClient_GetAPIKey_Masked(t *testing.T) {
	client := NewRegistryClient("", "")

	// Short key
	client.SetAPIKey("short")
	assert.Equal(t, "****", client.GetAPIKey())

	// Long key
	client.SetAPIKey("sk_live_skillsmp_ABCDEFGHIJKLMNOPQRSTUVWXYZ_12345678")
	assert.Equal(t, "sk_l...5678", client.GetAPIKey())
}

func TestRegistryClient_AnonymousAccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No auth header should be sent
		auth := r.Header.Get("Authorization")
		assert.Empty(t, auth)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"skills":[{"id":"anon","name":"Anon","slug":"anon","description":"d","category":"c","categorySlug":"c","stars":4.0,"downloads":1,"author":"a","version":"1.0.0","icon":"i","tags":[]}]}}`))
	}))
	defer server.Close()

	client := NewRegistryClientWithConfig(RegistryConfig{
		BaseURL: server.URL,
		// No API key
	})
	skills, err := client.FetchSkills("", "anon")
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "anon", skills[0].ID)
}

func TestFilterSkills(t *testing.T) {
	skills := []SkillManifest{
		{ID: "s1", Name: "Skill One", Category: "dev", Tags: []string{"test", "review"}},
		{ID: "s2", Name: "Skill Two", Category: "test", Tags: []string{"automation"}},
		{ID: "s3", Name: "Special Tool", Category: "dev", Tags: []string{"security"}},
	}

	// Filter by category
	result := filterSkills(skills, "dev", "")
	assert.Len(t, result, 2)

	// Filter by search (name)
	result = filterSkills(skills, "", "special")
	assert.Len(t, result, 1)
	assert.Equal(t, "s3", result[0].ID)

	// Filter by search (tag)
	result = filterSkills(skills, "", "automation")
	assert.Len(t, result, 1)
	assert.Equal(t, "s2", result[0].ID)

	// No filter
	result = filterSkills(skills, "", "")
	assert.Len(t, result, 3)
}

func TestRegistryCache_TTL(t *testing.T) {
	cache := &registryCache{}

	// Set cache
	cache.set([]SkillManifest{{ID: "s1"}})

	// Should be valid
	assert.NotNil(t, cache.get())

	// Expire cache
	cache.expiresAt = time.Now().Add(-1 * time.Hour)
	assert.Nil(t, cache.get())
}

func TestRegistryClient_SaveAndLoadLocalCache(t *testing.T) {
	tmpDir := t.TempDir()
	client := NewRegistryClientWithConfig(RegistryConfig{
		CacheDir: tmpDir,
	})

	skills := []SkillManifest{
		{ID: "cache-test", Name: "Cache Test", Version: "1.0.0", Description: "test", Author: "a", Category: "c", Rating: 4.0, RatingCount: 1, Downloads: 1, Icon: "i", Tags: []string{"t"}},
	}

	err := client.saveLocalCache(skills)
	require.NoError(t, err)

	loaded, err := client.loadLocalCache()
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, "cache-test", loaded[0].ID)
}
