package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// The Claude Code status line script drops its raw JSON into
// ~/.cctokenviz/status-<session_id>.json (see ~/.claude/statusline-command.py).
// That JSON carries the plan rate limits and the real context window size.

type statusFile struct {
	SessionID     string `json:"session_id"`
	Cwd           string `json:"cwd"`
	ContextWindow struct {
		Size int `json:"context_window_size"`
	} `json:"context_window"`
	RateLimits struct {
		FiveHour *rateWindow `json:"five_hour"`
		SevenDay *rateWindow `json:"seven_day"`
	} `json:"rate_limits"`
}

type rateWindow struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

// LimitInfo is one plan rate-limit window
type LimitInfo struct {
	Label      string  `json:"label"`
	Percentage float64 `json:"percentage"`
	ResetsAt   int64   `json:"resetsAt"` // unix seconds
}

// Limits holds the account-wide plan limits; empty slices until a session has reported them
type Limits struct {
	Windows []LimitInfo `json:"windows"`
	AsOf    int64       `json:"asOf"` // unix seconds when a session last reported these numbers
	Paused  bool        `json:"paused"` // a window is at 100%: polling stops so probes can't spend extra-usage credits
}

func statusDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cctokenviz")
}

func readStatus(path string) *statusFile {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var s statusFile
	if json.Unmarshal(data, &s) != nil {
		return nil
	}
	return &s
}

// contextSizeFor returns the real context window size reported for a session, or 0 if unknown
func contextSizeFor(sessionID string) int {
	dir := statusDir()
	if dir == "" {
		return 0
	}
	s := readStatus(filepath.Join(dir, "status-"+sessionID+".json"))
	if s == nil {
		return 0
	}
	return s.ContextWindow.Size
}

var (
	projSizeMu   sync.Mutex
	projSizeAt   time.Time
	projSizeMemo map[string]int
)

// contextSizeForProject returns the window size most recently reported by any session in the
// project (matched case-insensitively), or 0 if no session there has reported one. Scans are
// cached for a few seconds because every session row asks.
func contextSizeForProject(project string) int {
	projSizeMu.Lock()
	defer projSizeMu.Unlock()
	if projSizeMemo == nil || time.Since(projSizeAt) > 5*time.Second {
		projSizeMemo = map[string]int{}
		projSizeAt = time.Now()
		if dir := statusDir(); dir != "" {
			files, _ := filepath.Glob(filepath.Join(dir, "status-*.json"))
			newest := map[string]time.Time{}
			roots := loadConfig().ProjectFolders
			for _, f := range files {
				info, err := os.Stat(f)
				if err != nil {
					continue
				}
				s := readStatus(f)
				if s == nil || s.ContextWindow.Size <= 0 {
					continue
				}
				name := strings.ToLower(projectFromPath(s.Cwd, roots))
				if name == "" {
					continue
				}
				if info.ModTime().After(newest[name]) {
					newest[name] = info.ModTime()
					projSizeMemo[name] = s.ContextWindow.Size
				}
			}
		}
	}
	return projSizeMemo[strings.ToLower(project)]
}

// detectedContextSizes returns the window size each project's sessions reported, by lowercase project name
func detectedContextSizes() map[string]int {
	contextSizeForProject("") // refresh the scan if it is stale
	projSizeMu.Lock()
	defer projSizeMu.Unlock()
	out := make(map[string]int, len(projSizeMemo))
	for k, v := range projSizeMemo {
		out[k] = v
	}
	return out
}

// GetLimits merges every source: each session's status file plus the headless poll.
//
// A status file's modification time says when the status line last ran, not how old its
// rate-limit numbers are (they only change after an API response), so a recently written
// file can hold stale numbers. Instead, per window: the later reset time wins (a newer
// window), and within the same window the higher percentage wins (usage only goes up).
func (m *Monitor) GetLimits() Limits {
	sources := m.statusFileSources()
	if p := m.polledLimits(); len(p.Windows) > 0 {
		sources = append(sources, p)
	}

	var out Limits
	for _, label := range []string{"5-hour", "Weekly"} {
		var best *LimitInfo
		var bestAsOf int64
		for _, src := range sources {
			for i := range src.Windows {
				w := &src.Windows[i]
				if w.Label != label {
					continue
				}
				if best == nil || w.ResetsAt > best.ResetsAt ||
					(w.ResetsAt == best.ResetsAt && w.Percentage > best.Percentage) {
					best, bestAsOf = w, src.AsOf
				}
			}
		}
		if best != nil {
			if best.Percentage >= 100 {
				out.Paused = true
			}
			out.Windows = append(out.Windows, *best)
			if bestAsOf > out.AsOf {
				out.AsOf = bestAsOf
			}
		}
	}
	return out
}

// statusFileSources returns one Limits per status file that carries rate limits,
// with AsOf set to the file's modification time.
func (m *Monitor) statusFileSources() []Limits {
	var out []Limits
	dir := statusDir()
	if dir == "" {
		return out
	}
	files, err := filepath.Glob(filepath.Join(dir, "status-*.json"))
	if err != nil {
		return out
	}

	now := time.Now().Unix()
	for _, f := range files {
		if strings.HasSuffix(f, ".tmp") {
			continue
		}
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		s := readStatus(f)
		if s == nil {
			continue
		}

		src := Limits{AsOf: info.ModTime().Unix()}
		add := func(label string, w *rateWindow) {
			// Claude Code drops a window once it resets; do the same for stale files
			if w != nil && w.ResetsAt > now {
				src.Windows = append(src.Windows, LimitInfo{Label: label, Percentage: w.UsedPercentage, ResetsAt: w.ResetsAt})
			}
		}
		add("5-hour", s.RateLimits.FiveHour)
		add("Weekly", s.RateLimits.SevenDay)
		if len(src.Windows) > 0 {
			out = append(out, src)
		}
	}
	return out
}
