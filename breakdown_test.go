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

func bdTool(ts, id, name, target string) string {
	return `{"type":"assistant","timestamp":"` + ts + `","message":{"usage":{"input_tokens":1,"cache_read_input_tokens":25000,"output_tokens":5},"content":[{"type":"tool_use","id":"` + id +
		`","name":"` + name + `","input":{"file_path":"` + target + `"}}]}}` + "\u000a"
}

func TestBreakdownNamesTheAskAndWhatClaudeDid(t *testing.T) {
	log := bdUser("2026-10-19T10:00:00Z", "first prompt") +
		bdAssistant("2026-10-19T10:00:05Z", 20000, "ok") +
		bdUser("2026-10-19T10:10:00Z", "add the banner to the readme please") +
		bdTool("2026-10-19T10:10:02Z", "t1", "Read", "a.go") +
		bdTool("2026-10-19T10:10:02Z", "t1", "Read", "a.go") + // the same call logged twice
		bdTool("2026-10-19T10:10:03Z", "t2", "Read", "b.go") +
		bdTool("2026-10-19T10:10:04Z", "t3", "Bash", "") +
		bdTool("2026-10-19T10:10:05Z", "t4", "Edit", "a.go") +
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
	r := b.Rows[0]
	if r.Tokens != 15000 || r.Label != "add the banner to the readme please" || r.Actions != "Claude read 2 files, ran 1 command, edited 1 file" ||
		r.Reply != "The README now has the banner under the title" {
		t.Errorf("row 0: %+v", r)
	}
	r = b.Rows[1]
	if r.Tokens != 15000 || r.Actions != "Claude just answered" || r.Reply != "Pushed to the fork as a single commit" {
		t.Errorf("row 1: %+v", r)
	}
	if !strings.HasPrefix(r.Label, "y") {
		t.Errorf("a short reply keeps what you typed: %+v", r)
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
