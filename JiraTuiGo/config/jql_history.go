package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const maxHistory = 200

type JqlEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	OriginalText   string    `json:"originalText"`
	EffectiveJql   string    `json:"effectiveJql"`
	WasAiTranslated bool     `json:"wasAiTranslated"`
}

type JqlHistory struct {
	entries []JqlEntry
}

func jqlHistoryPath() string {
	var base string
	if runtime.GOOS == "windows" {
		base = os.Getenv("USERPROFILE")
		if base == "" {
			base, _ = os.UserHomeDir()
		}
		return filepath.Join(base, "Documents", "JiraTuiGo", "jql-history.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Documents", "jiratui", "jql-history.json")
}

func (h *JqlHistory) Add(original, effective string, wasAI bool) {
	for i, e := range h.entries {
		if e.OriginalText == original {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			break
		}
	}

	entry := JqlEntry{
		Timestamp:       time.Now(),
		OriginalText:    original,
		EffectiveJql:    effective,
		WasAiTranslated: wasAI,
	}
	h.entries = append([]JqlEntry{entry}, h.entries...)

	if len(h.entries) > maxHistory {
		h.entries = h.entries[:maxHistory]
	}
}

// GetByRecentIndex returns 0=newest, 1=second newest, etc.
func (h *JqlHistory) GetByRecentIndex(idx int) (JqlEntry, bool) {
	if idx < 0 || idx >= len(h.entries) {
		return JqlEntry{}, false
	}
	return h.entries[idx], true
}

func (h *JqlHistory) Len() int {
	return len(h.entries)
}

func (h *JqlHistory) Load() error {
	path := jqlHistoryPath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &h.entries)
}

func (h *JqlHistory) Save() error {
	path := jqlHistoryPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(h.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func NewJqlHistory() *JqlHistory {
	return &JqlHistory{}
}
