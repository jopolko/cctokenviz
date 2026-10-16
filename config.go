package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Config is read from ~/.headroom.json on every refresh, so edits apply live.
//
//	{"projects": ["my-app", "my-api"], "activeMinutes": 30}
//
// projects: only show sessions whose project folder (under ~/projects) matches. Empty shows all.
// activeMinutes: how recently a session must have been modified to be listed.
//
// contextTokens: context window size for sessions that haven't exported one (default 200000;
// use 1000000 for 1M-context models). Sessions using the status line export report their own.
//
// pollLimits: opt in (true) to a background headless `claude -p` call that refreshes plan limits.
// Off by default. See the README for what it does and what it costs.
// pollMinutes: gap between automatic polls (default 10, minimum 5; halved above 85% usage).
// credits: balances shown in the panel. "console" is a snapshot (balance in USD, asOf as YYYY-MM-DD);
// with ANTHROPIC_ADMIN_KEY set, spend since asOf is subtracted. "claudeAi" is shown as typed.
type Config struct {
	Projects      []string `json:"projects"`
	ActiveMinutes int      `json:"activeMinutes"`
	ContextTokens int      `json:"contextTokens"`
	PollLimits    *bool    `json:"pollLimits"`
	PollMinutes   int      `json:"pollMinutes"`
	Credits       struct {
		Console *struct {
			Balance float64 `json:"balance"`
			AsOf    string  `json:"asOf"`
		} `json:"console"`
		ClaudeAi string `json:"claudeAi"`
	} `json:"credits"`
}

func loadConfig() Config {
	cfg := Config{ActiveMinutes: ActiveMinutes}
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}
	data, err := os.ReadFile(filepath.Join(home, ".headroom.json"))
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	if cfg.ActiveMinutes <= 0 {
		cfg.ActiveMinutes = ActiveMinutes
	}
	return cfg
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
