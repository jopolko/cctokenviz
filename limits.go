package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The Claude Code status line script drops its raw JSON into
// ~/.headroom/status-<session_id>.json (see ~/.claude/statusline-command.py).
// That JSON carries the plan rate limits and the real context window size.

type statusFile struct {
	SessionID     string `json:"session_id"`
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
}

func statusDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".headroom")
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

// GetLimits returns the freshest limits from either the status line files or the headless poll.
func (m *Monitor) GetLimits() Limits {
	fromFiles := m.statusFileLimits()
	fromPoll := m.polledLimits()
	if fromPoll.AsOf > fromFiles.AsOf && len(fromPoll.Windows) > 0 {
		return fromPoll
	}
	return fromFiles
}

// statusFileLimits reads the most recently written status file that carries rate limits.
// Limits are per account, so any session's copy will do; the newest is the freshest.
func (m *Monitor) statusFileLimits() Limits {
	var out Limits
	dir := statusDir()
	if dir == "" {
		return out
	}
	files, err := filepath.Glob(filepath.Join(dir, "status-*.json"))
	if err != nil {
		return out
	}

	var newest *statusFile
	var newestTime time.Time
	for _, f := range files {
		if strings.HasSuffix(f, ".tmp") {
			continue
		}
		info, err := os.Stat(f)
		if err != nil || !info.ModTime().After(newestTime) {
			continue
		}
		s := readStatus(f)
		if s == nil || (s.RateLimits.FiveHour == nil && s.RateLimits.SevenDay == nil) {
			continue
		}
		newest, newestTime = s, info.ModTime()
	}
	if newest == nil {
		return out
	}

	out.AsOf = newestTime.Unix()
	now := time.Now().Unix()
	add := func(label string, w *rateWindow) {
		// Claude Code drops a window once it resets; do the same for stale files
		if w != nil && w.ResetsAt > now {
			out.Windows = append(out.Windows, LimitInfo{Label: label, Percentage: w.UsedPercentage, ResetsAt: w.ResetsAt})
		}
	}
	add("5-hour", newest.RateLimits.FiveHour)
	add("Weekly", newest.RateLimits.SevenDay)
	return out
}
