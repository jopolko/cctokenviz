package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMergeLimitsKeepsJustResetWeekly(t *testing.T) {
	now := int64(1_000_000)
	weekAgoReset := now - 3600 // reset an hour ago
	src := []Limits{{AsOf: now - 7200, Windows: []LimitInfo{
		{Label: "Weekly", Percentage: 84, ResetsAt: weekAgoReset},
		{Label: "5-hour", Percentage: 18, ResetsAt: now + 3600},
	}}}
	out := mergeLimits(src, now)
	if len(out.Windows) != 2 {
		t.Fatalf("want both windows, got %+v", out.Windows)
	}
	var weekly LimitInfo
	for _, w := range out.Windows {
		if w.Label == "Weekly" {
			weekly = w
		}
	}
	if !weekly.Expired || weekly.Percentage != 0 || weekly.ResetsAt != weekAgoReset+7*86400 {
		t.Fatalf("weekly should show as just reset with the next reset a week on: %+v", weekly)
	}
	if out.Paused {
		t.Fatal("an expired 84% window must not pause polling")
	}
}

func TestMergeLimitsFreshReadingReplacesExpired(t *testing.T) {
	now := int64(1_000_000)
	out := mergeLimits([]Limits{
		{Windows: []LimitInfo{{Label: "Weekly", Percentage: 84, ResetsAt: now - 3600}}},
		{Windows: []LimitInfo{{Label: "Weekly", Percentage: 1, ResetsAt: now + 6*86400}}},
	}, now)
	if len(out.Windows) != 1 || out.Windows[0].Expired || out.Windows[0].Percentage != 1 {
		t.Fatalf("fresh reading should win: %+v", out.Windows)
	}
}

func TestMergeLimitsFiveHourExpiredHasNoReset(t *testing.T) {
	now := int64(1_000_000)
	out := mergeLimits([]Limits{{Windows: []LimitInfo{{Label: "5-hour", Percentage: 90, ResetsAt: now - 60}}}}, now)
	if len(out.Windows) != 1 || !out.Windows[0].Expired || out.Windows[0].ResetsAt != 0 {
		t.Fatalf("5-hour should be expired with unknown next reset: %+v", out.Windows)
	}
}

func TestContextSizeForModelUsesNewestReport(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".cctokenviz")
	os.MkdirAll(dir, 0o755)
	write := func(name, model string, size int, age time.Duration) {
		body, _ := json.Marshal(map[string]any{
			"session_id":     name,
			"cwd":            home,
			"model":          map[string]any{"id": model},
			"context_window": map[string]any{"context_window_size": size},
		})
		p := filepath.Join(dir, "status-"+name+".json")
		os.WriteFile(p, body, 0o644)
		at := time.Now().Add(-age)
		os.Chtimes(p, at, at)
	}
	write("a", "model-x", 200000, 2*time.Hour)
	write("b", "model-x", 1000000, time.Minute) // newer report wins
	write("c", "model-y", 200000, time.Minute)

	projSizeMu.Lock()
	projSizeMemo = nil // force a rescan of the temp folder
	projSizeMu.Unlock()

	if got := contextSizeForModel("model-x"); got != 1000000 {
		t.Fatalf("model-x: got %d, want 1000000", got)
	}
	if got := contextSizeForModel("model-y"); got != 200000 {
		t.Fatalf("model-y: got %d, want 200000", got)
	}
	if got := contextSizeForModel("unknown"); got != 0 {
		t.Fatalf("unknown model should be 0, got %d", got)
	}

	projSizeMu.Lock()
	projSizeMemo = nil // don't leak the temp folder's sizes into other tests
	projSizeMu.Unlock()
}
