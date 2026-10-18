package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestGetUsageDedupesAndPrices(t *testing.T) {
	dir := t.TempDir()
	proj := filepath.Join(dir, "p")
	os.MkdirAll(proj, 0o755)
	ts := time.Now().Format(time.RFC3339Nano)
	line := func(out int) string {
		return `{"type":"assistant","timestamp":"` + ts + `","requestId":"r1","message":{"id":"m1","model":"claude-sonnet-4-5","usage":{"input_tokens":1000,"output_tokens":` + strconv.Itoa(out) + `,"cache_read_input_tokens":10000,"cache_creation_input_tokens":0}}}` + "\n"
	}
	// the same message logged twice while streaming: the last one counts
	os.WriteFile(filepath.Join(proj, "s.jsonl"), []byte(line(10)+line(500)), 0o644)
	u := (&Monitor{claudeDir: dir}).GetUsage()
	p := u.Periods[0]
	if p.OutputTokens != 500 || p.InputTokens != 1000 {
		t.Fatalf("dedupe failed: %+v", p)
	}
	// 1000*3 + 500*15 + 10000*3*0.1 = 13500 per million
	if want := 0.0135; p.Cost < want-1e-9 || p.Cost > want+1e-9 {
		t.Fatalf("cost %v want %v", p.Cost, want)
	}
}
