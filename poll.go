package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
// on its own (a session was active recently), at most once per poll interval.

const (
	defaultPollMinutes = 10 // gap between headless calls, see pollMinutes in the config
	minPollMinutes     = 5
	busyPercent        = 85 // above this, poll twice as often: that's when limits drive decisions
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

// MaybePollLimits runs the headless call if a session is active and the last attempt is old enough.
func (m *Monitor) MaybePollLimits() {
	// Opt-in: each probe is a real (tiny) request that counts against the user's plan
	if cfg := loadConfig(); !cfg.PollLimits {
		return
	}

	if m.GetLimits().Paused {
		return
	}

	// pollInterval reads the merged limits, which takes the same lock, so call it before locking
	interval := m.pollInterval()
	polled.mu.Lock()
	recentAttempt := time.Since(polled.lastAttempt) < interval
	polled.mu.Unlock()
	if recentAttempt {
		return
	}

	if !m.anySessionActive(pollActiveWithin) {
		return
	}

	m.pollNow()
}

// RefreshLimitsNow is the manual refresh button: one probe now, still opt-in, and rate
// limited to one per 15 seconds so a stuck button can't spam requests.
func (m *Monitor) RefreshLimitsNow() {
	if cfg := loadConfig(); !cfg.PollLimits {
		return
	}
	if m.GetLimits().Paused {
		return
	}
	polled.mu.Lock()
	tooSoon := time.Since(polled.lastAttempt) < 15*time.Second
	polled.mu.Unlock()
	if tooSoon {
		return
	}
	m.pollNow()
}

// pollInterval is the configured gap (default 10 min, floor 5), halved when any limit is
// nearly used up, since that is when a fresher number changes what you do.
func (m *Monitor) pollInterval() time.Duration {
	mins := loadConfig().PollMinutes
	if mins <= 0 {
		mins = defaultPollMinutes
	}
	if mins < minPollMinutes {
		mins = minPollMinutes
	}
	d := time.Duration(mins) * time.Minute
	for _, w := range m.GetLimits().Windows {
		if w.Percentage >= busyPercent {
			if d /= 2; d < minPollMinutes*time.Minute {
				d = minPollMinutes * time.Minute
			}
			break
		}
	}
	return d
}

func (m *Monitor) pollNow() {
	polled.mu.Lock()
	polled.lastAttempt = time.Now()
	polled.mu.Unlock()

	limits, err := runHeadlessProbe()
	if err != nil {
		pollLog("probe failed: " + err.Error())
		return
	}
	pollLog(fmt.Sprintf("probe ok: %d windows", len(limits.Windows)))
	polled.mu.Lock()
	polled.limits = limits
	polled.mu.Unlock()
}

// pollLog appends to ~/.headroom/poll.log (trimmed when large) so a quiet failure can be diagnosed
func pollLog(msg string) {
	dir := statusDir()
	if dir == "" {
		return
	}
	os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, "poll.log")
	if info, err := os.Stat(path); err == nil && info.Size() > 100*1024 {
		os.Remove(path)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
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

func runHeadlessProbe() (Limits, error) {
	var out Limits
	claude := findClaude()
	if claude == "" {
		return out, errors.New("claude CLI not found")
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
	// A non-zero exit can still carry a usable rate_limit_event, so parse what we got first
	runErr := cmd.Run()

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
		if runErr != nil {
			return out, runErr
		}
		return out, errors.New("no rate_limit_event in output")
	}
	out.AsOf = time.Now().Unix()
	return out, nil
}
