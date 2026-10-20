package main

import (
	"strings"
	"testing"
)

func bdAssistant(ts string, ctx int, text string) string {
	return `{"type":"assistant","timestamp":"` + ts + `","message":{"usage":{"input_tokens":1,"cache_read_input_tokens":` + itoa(ctx) +
		`,"output_tokens":5},"content":[{"type":"text","text":"` + text + `"}]}}` + "\n"
}

func bdUser(ts, text string) string {
	return `{"type":"user","timestamp":"` + ts + `","message":{"content":"` + text + `"}}` + "\n"
}

func TestBreakdownLabelsRowsByClaudesReply(t *testing.T) {
	log := bdUser("2026-10-19T10:00:00Z", "first prompt") +
		bdAssistant("2026-10-19T10:00:05Z", 20000, "ok") +
		bdUser("2026-10-19T10:10:00Z", "add the banner to the readme please") +
		bdAssistant("2026-10-19T10:10:05Z", 30000, "working") +
		`{"type":"user","timestamp":"2026-10-19T10:10:06Z","message":{"content":[{"type":"tool_result","content":"x"}]}}` + "\u000a" +
		bdAssistant("2026-10-19T10:10:09Z", 35000, "The README now has the banner under the title. I also fixed a typo.") +
		bdUser("2026-10-19T10:20:00Z", "y") +
		bdAssistant("2026-10-19T10:20:05Z", 50000, "**Pushed** to [the fork](https://example.com/x) as a single commit.") +
		bdUser("2026-10-19T10:30:00Z", "tiny") +
		bdAssistant("2026-10-19T10:30:05Z", 51000, "ok")
	b := buildBreakdown(strings.NewReader(log))
	if b.Start != 20001 || b.StartLabel != "Starting context" || b.Total != 51001 {
		t.Fatalf("start/total: %+v", b)
	}
	if len(b.Rows) != 2 {
		t.Fatalf("rows: %+v", b.Rows)
	}
	if b.Rows[0].Tokens != 15000 || b.Rows[0].Label != "The README now has the banner under the title" || b.Rows[0].Prompt != "add the banner to the readme please" {
		t.Errorf("row 0: %+v", b.Rows[0])
	}
	if b.Rows[1].Tokens != 15000 || b.Rows[1].Label != "Pushed to the fork as a single commit" || b.Rows[1].Prompt != "y" {
		t.Errorf("a bare \"y\" is named by what Claude did, with the markdown stripped: %+v", b.Rows[1])
	}
	if b.SmallCount != 1 || b.SmallTokens != 1000 {
		t.Errorf("small prompts should be folded: %+v", b)
	}
}

func TestPromptLabelFallbackForShortReply(t *testing.T) {
	got := promptLabel("y", "Done. Do you want me to push it?")
	if got != "y → Do you want me to push it?" {
		t.Errorf("got %q", got)
	}
}

func TestBreakdownStartsOverAtCompact(t *testing.T) {
	log := bdUser("2026-10-19T10:00:00Z", "old work") + bdAssistant("2026-10-19T10:00:05Z", 150000, "ok") +
		`{"type":"system","subtype":"compact_boundary","timestamp":"2026-10-19T11:00:00Z"}` + "\n" +
		`{"type":"user","isCompactSummary":true,"timestamp":"2026-10-19T11:00:01Z","message":{"content":"summary"}}` + "\n" +
		bdUser("2026-10-19T11:01:00Z", "next thing") + bdAssistant("2026-10-19T11:01:05Z", 25000, "ok") +
		bdUser("2026-10-19T11:05:00Z", "one more thing to do") + bdAssistant("2026-10-19T11:05:05Z", 31000, "ok")
	b := buildBreakdown(strings.NewReader(log))
	if b.StartLabel != "Carried over" || b.Start != 25001 || len(b.Rows) != 1 || b.Rows[0].Tokens != 6000 {
		t.Fatalf("after compact: %+v", b)
	}
}
