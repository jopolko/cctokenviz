package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config is the user's settings, stored in ~/.cctokenviz.json. The setup wizard and the
// settings panel edit it; it is re-read on every refresh, so hand edits apply live too.
type Config struct {
	// What to show
	Projects      []string `json:"projects"`      // only these project folders; empty shows all
	ActiveMinutes int      `json:"activeMinutes"` // how recently a session must have changed to be listed
	ContextTokens int      `json:"contextTokens"` // context window for sessions that haven't reported one
	ProjectFolders []string `json:"projectFolders"` // folders you added: a folder of projects, or a single project
	// Per-project window override (0 or missing = detect). Used when a session has not
	// exported its own size.
	ProjectContext map[string]int `json:"projectContext"`
	// Projects never shown, however active (unlike the X on a row, which only hides a session until it has new activity)
	HiddenProjects []string `json:"hiddenProjects"`

	// Plan limits
	PollLimits  bool `json:"pollLimits"`  // opt in to the background headless call that refreshes limits
	PollMinutes int  `json:"pollMinutes"` // gap between automatic polls (default 10, minimum 5)

	// API spend panel
	ShowUsage   bool `json:"showUsage"`   // show the API spend panel (off by default)
	UsageFolded bool `json:"usageFolded"` // the panel is folded to its header (expanded by default)

	// Look
	Theme     string `json:"theme"`
	Accent    string `json:"accent"` // optional #rrggbb override for headings and highlights
	Font      string `json:"font"`
	FontSize  int    `json:"fontSize"`
	SetupDone bool   `json:"setupDone"`

	// Window
	FixedSize bool `json:"fixedSize"` // true keeps the window size you set; false fits its height to the content
}

var configMu sync.Mutex

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cctokenviz.json")
}

func loadConfig() Config {
	cfg := Config{ActiveMinutes: ActiveMinutes}
	if p := configPath(); p != "" {
		if data, err := os.ReadFile(p); err == nil {
			// Notepad and PowerShell can add a UTF-8 byte order mark, which JSON parsing rejects
			json.Unmarshal(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf")), &cfg)
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
	if strings.Join(loadConfig().ProjectFolders, "|") != strings.Join(cfg.ProjectFolders, "|") {
		resetSessionCache() // project names are worked out while reading logs
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, append(data, '\n'), 0o644)
}

func (c Config) allows(project string) bool {
	for _, h := range c.HiddenProjects {
		if strings.EqualFold(h, project) {
			return false
		}
	}
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

// contextFor returns the override set for a project, or 0
func (c Config) contextFor(project string) int {
	for k, v := range c.ProjectContext {
		if strings.EqualFold(k, project) {
			return v
		}
	}
	return 0
}

// projectsRoot returns the "projects" directory that cwd sits in, or "".
func projectsRoot(cwd string) string {
	norm := strings.ReplaceAll(cwd, "\\", "/")
	parts := strings.Split(norm, "/")
	for i, p := range parts {
		if strings.EqualFold(p, "projects") && i+1 < len(parts) && parts[i+1] != "" {
			return filepath.FromSlash(strings.Join(parts[:i+1], "/"))
		}
	}
	return ""
}
