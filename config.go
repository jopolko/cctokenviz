package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config is the user's settings, stored in ~/.headroom.json. The setup wizard and the
// settings panel edit it; it is re-read on every refresh, so hand edits apply live too.
type Config struct {
	// What to show
	Projects      []string `json:"projects"`      // only these project folders; empty shows all
	ActiveMinutes int      `json:"activeMinutes"` // how recently a session must have changed to be listed
	ContextTokens int      `json:"contextTokens"` // context window for sessions that haven't reported one

	// Plan limits
	PollLimits  bool `json:"pollLimits"`  // opt in to the background headless call that refreshes limits
	PollMinutes int  `json:"pollMinutes"` // gap between automatic polls (default 10, minimum 5)

	// Look
	Theme     string `json:"theme"`
	Accent    string `json:"accent"` // optional #rrggbb override for headings and highlights
	Font      string `json:"font"`
	FontSize  int    `json:"fontSize"`
	SetupDone bool   `json:"setupDone"`
}

var configMu sync.Mutex

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".headroom.json")
}

func loadConfig() Config {
	cfg := Config{ActiveMinutes: ActiveMinutes}
	if p := configPath(); p != "" {
		if data, err := os.ReadFile(p); err == nil {
			json.Unmarshal(data, &cfg)
		}
	}
	if cfg.ActiveMinutes <= 0 {
		cfg.ActiveMinutes = ActiveMinutes
	}
	return cfg
}

func saveConfig(cfg Config) error {
	configMu.Lock()
	defer configMu.Unlock()
	p := configPath()
	if p == "" {
		return os.ErrNotExist
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, append(data, '\n'), 0o644)
}

func (c Config) allows(project string) bool {
	if len(c.Projects) == 0 {
		return true
	}
	for _, p := range c.Projects {
		if strings.EqualFold(p, project) {
			return true
		}
	}
	return false
}

// projectFromCwd returns the first folder under a "projects" directory,
// e.g. C:\Users\me\projects\my-app\src -> "my-app". Empty if cwd is not inside one.
func projectFromCwd(cwd string) string {
	parts := strings.Split(strings.ReplaceAll(cwd, "\\", "/"), "/")
	for i, p := range parts {
		if strings.EqualFold(p, "projects") && i+1 < len(parts) && parts[i+1] != "" {
			return parts[i+1]
		}
	}
	return ""
}
