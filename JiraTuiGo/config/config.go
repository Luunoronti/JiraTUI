package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type ConnectionConfig struct {
	BaseURL        string `json:"baseUrl"`
	Email          string `json:"email"`
	TokenProtected string `json:"tokenProtected"`
	AuthType       string `json:"authType"`
}

type AppearanceConfig struct {
	ThemeName string `json:"themeName"`
}

type BehaviorConfig struct {
	DefaultJql         string    `json:"defaultJql"`
	PageSize           int       `json:"pageSize"`
	AutoRefreshSeconds int       `json:"autoRefreshSeconds"`
	LastUpdateCheck    time.Time `json:"lastUpdateCheck,omitempty"`
}

type AiConfig struct {
	Adapter          string `json:"adapter"`
	BaseURL          string `json:"baseUrl"`
	Model            string `json:"model"`
	ApiKeyProtected  string `json:"apiKeyProtected"`
}

type ColumnVisibilityConfig struct {
	Key      bool `json:"key"`
	Type     bool `json:"type"`
	Priority bool `json:"priority"`
	Status   bool `json:"status"`
	Assignee bool `json:"assignee"`
	Summary  bool `json:"summary"`
}

type AppConfig struct {
	Conn ConnectionConfig `json:"connection"`
	Appearance AppearanceConfig       `json:"appearance"`
	Behavior   BehaviorConfig         `json:"behavior"`
	AI         AiConfig               `json:"ai"`
	Columns    ColumnVisibilityConfig `json:"columns"`
}

func configPath() string {
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
	return filepath.Join(base, "JiraTuiGo", "config.json")
}

func defaults() *AppConfig {
	return &AppConfig{
		Conn: ConnectionConfig{
			AuthType: "ApiToken",
		},
		Appearance: AppearanceConfig{
			ThemeName: "Dark",
		},
		Behavior: BehaviorConfig{
			DefaultJql: "assignee = currentUser() AND statusCategory != Done",
			PageSize:   50,
		},
		AI: AiConfig{
			Adapter: "anthropic",
			Model:   "claude-sonnet-4-5",
		},
		Columns: ColumnVisibilityConfig{
			Key:      true,
			Type:     true,
			Priority: true,
			Status:   true,
			Assignee: true,
			Summary:  true,
		},
	}
}

func Load() (*AppConfig, error) {
	cfg := defaults()
	path := configPath()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func (cfg *AppConfig) Save() error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func ConfigPath() string {
	return configPath()
}
