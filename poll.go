package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Terminal sessions report plan limits through the status line, but the Claude desktop
// app doesn't run the status line. To cover that gap, a tiny headless `claude -p` call
// emits a rate_limit_event with both windows. It only runs when it won't start a window
// on its own (a session was active recently) and the numbers have gone stale.

const (
	pollEvery        = 10 * time.Minute // minimum gap between headless calls, and staleness threshold
	pollActiveWithin = 30 * time.Minute // only poll if a session log changed this recently
	pollTimeout      = 90 * time.Second
)

type polledState struct {
	mu          sync.Mutex
	limits      Limits
	lastAttempt time.Time
}

var polled polledState

func (m *Monitor) polledLimits() Limits {
	polled.mu.Lock()
	defer polled.mu.Unlock()

	out := Limits{AsOf: polled.limits.AsOf}
	now := time.Now().Unix()
	for _, w := range polled.limits.Windows {
		if w.ResetsAt > now {
			out.Windows = append(out.Windows, w)
		}
	}
	return out
}

// anySessionActive reports whether any session log changed within the given window
func (m *Monitor) anySessionActive(within time.Duration) bool {
	cutoff := time.Now().Add(-within)
	dirs, err := os.ReadDir(m.claudeDir)
	if err != nil {
		return false
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		files, err := os.ReadDir(filepath.Join(m.claudeDir, d.Name()))
		if err != nil {
			continue
		}
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			if info, err := f.Info(); err == nil && info.ModTime().After(cutoff) {
				return true
			}
		}
	}
	return false
}

// MaybePollLimits runs the headless call if the numbers are stale and a session is active.
func (m *Monitor) MaybePollLimits() {
	// Opt-in: each probe is a real (tiny) request that counts against the user's plan
	if cfg := loadConfig(); cfg.PollLimits == nil || !*cfg.PollLimits {
		return
	}

	polled.mu.Lock()
	recentAttempt := time.Since(polled.lastAttempt) < pollEvery
	polled.mu.Unlock()
	if recentAttempt {
		return
	}

	// Fresh data from a terminal session's status line means no call is needed
	if files := m.statusFileLimits(); files.AsOf > 0 && time.Since(time.Unix(files.AsOf, 0)) < pollEvery {
		return
	}
	if !m.anySessionActive(pollActiveWithin) {
		return
	}

	polled.mu.Lock()
	polled.lastAttempt = time.Now()
	polled.mu.Unlock()

	if limits, ok := runHeadlessProbe(); ok {
		polled.mu.Lock()
		polled.limits = limits
		polled.mu.Unlock()
	}
}

func findClaude() string {
	if p, err := exec.LookPath("claude"); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	for _, name := range []string{"claude.exe", "claude"} {
		p := filepath.Join(home, ".local", "bin", name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

type rateLimitEvent struct {
	Type string `json:"type"`
	Info struct {
		Windows map[string]struct {
			Utilization float64 `json:"utilization"` // fraction, 0.63 = 63%
			ResetsAt    int64   `json:"resetsAt"`
		} `json:"unifiedWindows"`
	} `json:"rate_limit_info"`
}

func runHeadlessProbe() (Limits, bool) {
	var out Limits
	claude := findClaude()
	if claude == "" {
		return out, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), pollTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, claude, "-p", "reply with: ok",
		"--model", "haiku",
		"--output-format", "stream-json", "--verbose",
		"--no-session-persistence")
	cmd.Dir = os.TempDir() // keep it out of any project
	hideWindow(cmd)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return out, false
	}

	scanner := bufio.NewScanner(&stdout)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		var ev rateLimitEvent
		if json.Unmarshal(scanner.Bytes(), &ev) != nil || ev.Type != "rate_limit_event" {
			continue
		}
		// Fixed order: 5-hour first
		for _, w := range []struct{ key, label string }{{"five_hour", "5-hour"}, {"seven_day", "Weekly"}} {
			if v, ok := ev.Info.Windows[w.key]; ok && v.ResetsAt > 0 {
				out.Windows = append(out.Windows, LimitInfo{Label: w.label, Percentage: v.Utilization * 100, ResetsAt: v.ResetsAt})
			}
		}
	}
	if len(out.Windows) == 0 {
		return out, false
	}
	out.AsOf = time.Now().Unix()
	return out, true
}
