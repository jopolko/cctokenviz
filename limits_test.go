package main

import "testing"

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
