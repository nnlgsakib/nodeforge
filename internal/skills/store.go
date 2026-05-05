package skills

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS installed_skills (
	skill_id TEXT PRIMARY KEY,
	installed_at TIMESTAMP NOT NULL,
	version TEXT NOT NULL
);
`

// Store manages persistent storage of installed skills using SQLite.
type Store struct {
	db *sql.DB
}

// NewStore creates or opens a SQLite store for installed skills.
func NewStore(dbPath string) (*Store, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("skills: resolve home dir: %w", err)
		}
		dbPath = filepath.Join(home, ".nforge", "skills.db")
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("skills: create dir %q: %w", dir, err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("skills: open db: %w", err)
	}

	// SQLite is file-based; serialize access with a single connection
	db.SetMaxOpenConns(1)

	// Enable WAL mode for concurrent writes
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("skills: enable WAL: %w", err)
	}

	// Create schema
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("skills: create schema: %w", err)
	}

	return &Store{db: db}, nil
}

// InstalledSkill represents a single installed skill record.
type InstalledSkill struct {
	SkillID     string
	InstalledAt time.Time
	Version     string
}

// Insert adds a skill to the installed skills table.
func (s *Store) Insert(skillID, version string) error {
	now := time.Now()
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO installed_skills (skill_id, installed_at, version) VALUES (?, ?, ?)",
		skillID, now, version,
	)
	if err != nil {
		return fmt.Errorf("skills: insert %q: %w", skillID, err)
	}
	return nil
}

// Delete removes a skill from the installed skills table.
func (s *Store) Delete(skillID string) error {
	_, err := s.db.Exec("DELETE FROM installed_skills WHERE skill_id = ?", skillID)
	if err != nil {
		return fmt.Errorf("skills: delete %q: %w", skillID, err)
	}
	return nil
}

// Exists checks if a skill is installed.
func (s *Store) Exists(skillID string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM installed_skills WHERE skill_id = ?", skillID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("skills: exists %q: %w", skillID, err)
	}
	return count > 0, nil
}

// List returns all installed skills.
func (s *Store) List() ([]InstalledSkill, error) {
	rows, err := s.db.Query("SELECT skill_id, installed_at, version FROM installed_skills ORDER BY installed_at DESC")
	if err != nil {
		return nil, fmt.Errorf("skills: list: %w", err)
	}
	defer rows.Close()

	result := make([]InstalledSkill, 0)
	for rows.Next() {
		var sk InstalledSkill
		if err := rows.Scan(&sk.SkillID, &sk.InstalledAt, &sk.Version); err != nil {
			return nil, fmt.Errorf("skills: scan row: %w", err)
		}
		result = append(result, sk)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("skills: iterate rows: %w", err)
	}
	return result, nil
}

// IsInstalled is a convenience method matching the old installedSkills map API.
func (s *Store) IsInstalled(skillID string) bool {
	exists, err := s.Exists(skillID)
	return err == nil && exists
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
