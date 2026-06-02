package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// IssueMeta holds local per-issue metadata (not stored in Jira).
type IssueMeta struct {
	Hidden   bool      `json:"hidden"`
	HiddenAt time.Time `json:"hiddenAt,omitempty"`
}

// IssueMetaStore persists local issue metadata alongside config.json.
type IssueMetaStore struct {
	Issues map[string]IssueMeta `json:"issues"`
}

// issueMetaPath returns the path to issue-meta.json, using the same base
// directory as configPath().
func issueMetaPath() string {
	var base string
	if runtime.GOOS == "windows" {
		base = os.Getenv("APPDATA")
		if base == "" {
			base, _ = os.UserHomeDir()
		}
	} else {
		base = os.Getenv("XDG_CONFIG_HOME")
		if base == "" {
			home, _ := os.UserHomeDir()
			base = filepath.Join(home, ".config")
		}
	}
	return filepath.Join(base, "JiraTuiGo", "issue-meta.json")
}

// NewIssueMetaStore returns an empty store ready to use.
func NewIssueMetaStore() *IssueMetaStore {
	return &IssueMetaStore{
		Issues: make(map[string]IssueMeta),
	}
}

// IsHidden reports whether the issue with the given key is hidden.
func (s *IssueMetaStore) IsHidden(key string) bool {
	if s == nil {
		return false
	}
	return s.Issues[key].Hidden
}

// ToggleHidden flips the hidden flag for the given key and returns the new state.
func (s *IssueMetaStore) ToggleHidden(key string) (nowHidden bool) {
	meta := s.Issues[key]
	meta.Hidden = !meta.Hidden
	if meta.Hidden {
		meta.HiddenAt = time.Now()
	} else {
		meta.HiddenAt = time.Time{}
	}
	s.Issues[key] = meta
	return meta.Hidden
}

// Load reads the store from issueMetaPath(). Missing file is not an error.
func (s *IssueMetaStore) Load() error {
	data, err := os.ReadFile(issueMetaPath())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}

// Save writes the store to issueMetaPath(), creating directories as needed.
func (s *IssueMetaStore) Save() error {
	path := issueMetaPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
