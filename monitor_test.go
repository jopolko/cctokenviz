package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func assistantLine(cwd string, input, cacheRead int) string {
	return `{"type":"assistant","cwd":"` + cwd + `","message":{"usage":{"input_tokens":` +
		itoa(input) + `,"cache_read_input_tokens":` + itoa(cacheRead) + `,"output_tokens":5}}}` + "\n"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{byte('0' + n%10)}, b...)
	}
	return string(b)
}

func appendFile(t *testing.T, path, text string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		t.Fatal(err)
	}
}

func TestParseSessionIncremental(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "abc.jsonl")
	m := &Monitor{claudeDir: dir}

	appendFile(t, path, assistantLine(`C:\\Users\\me\\projects\\my-app`, 2, 1000))
	info := m.parseSession(path, "proj")
	if info == nil || info.UsedTokens != 1002 || info.ProjectName != "my-app" {
		t.Fatalf("first parse: %+v", info)
	}

	// A half-written line must not be consumed or change the result
	full := assistantLine(`C:\\Users\\me\\projects\\my-app`, 3, 5000)
	appendFile(t, path, full[:len(full)/2])
	if info = m.parseSession(path, "proj"); info.UsedTokens != 1002 {
		t.Fatalf("partial line was consumed: %+v", info)
	}

	// Once the line is completed, it is picked up
	appendFile(t, path, full[len(full)/2:])
	if info = m.parseSession(path, "proj"); info.UsedTokens != 5003 {
		t.Fatalf("completed line not read: %+v", info)
	}

	// A rewritten, shorter file starts over
	os.WriteFile(path, []byte(assistantLine(`C:\\Users\\me\\projects\\other`, 1, 10)), 0o644)
	if info = m.parseSession(path, "proj"); info.UsedTokens != 11 || info.ProjectName != "other" {
		t.Fatalf("shrunk file not reparsed: %+v", info)
	}
}

func TestAddSignalDoesNotFlipOnOneStrayFile(t *testing.T) {
	st := &sessionState{}
	for i := 0; i < 5; i++ {
		st.addSignal("lookout")
	}
	if got := st.addSignal("joshuaopolko"); got != "lookout" {
		t.Fatalf("one stray file flipped the project to %q", got)
	}
	// a real move wins once it has been going for a while
	var got string
	for i := 0; i < 6; i++ {
		got = st.addSignal("joshuaopolko")
	}
	if got != "joshuaopolko" {
		t.Fatalf("sustained work in a new project did not take over: %q", got)
	}
}

func TestFoldByProjectKeepsMostRecentlyActive(t *testing.T) {
	in := []SessionInfo{
		{ID: "old", ProjectName: "lookout", ActivityAt: 100, Percentage: 14, Title: "resume"},
		{ID: "new", ProjectName: "Lookout", ActivityAt: 200, Percentage: 18},
		{ID: "x", ProjectName: "other", ActivityAt: 50},
	}
	out := foldByProject(in)
	if len(out) != 2 {
		t.Fatalf("want 2 rows, got %d", len(out))
	}
	if out[0].ID != "new" || len(out[0].Others) != 1 || out[0].Others[0].Percentage != 14 {
		t.Fatalf("wrong fold: %+v", out[0])
	}
}

func TestAnySessionActiveIgnoresBookkeepingWrites(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)
	claude := filepath.Join(home, ".claude", "projects", "p")
	os.MkdirAll(claude, 0o755)

	fresh := time.Now().UTC().Format(time.RFC3339Nano)
	stale := time.Now().Add(-5 * time.Hour).UTC().Format(time.RFC3339Nano)
	assistant := func(ts string) string {
		return `{"type":"assistant","timestamp":"` + ts + `","cwd":"` + strings.ReplaceAll(home, `\`, `/`) + `","message":{"id":"m","model":"x","usage":{"input_tokens":1,"output_tokens":1,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}` + "\n"
	}
	bookkeeping := `{"type":"custom-title","title":"x"}` + "\n" + `{"type":"system","subtype":"stop_hook_summary","timestamp":"` + fresh + `"}` + "\n" +
		`{"type":"user","timestamp":"` + fresh + `","message":{"content":"<task-notification>done</task-notification>"}}` + "\n"
	m := &Monitor{claudeDir: filepath.Join(home, ".claude", "projects")}

	// Log modified just now, but the last real message was hours ago and the new lines are bookkeeping
	os.WriteFile(filepath.Join(claude, "idle.jsonl"), []byte(assistant(stale)+bookkeeping), 0o644)
	if m.anySessionActive(30 * time.Minute) {
		t.Fatal("bookkeeping writes made an idle session look active")
	}

	// A real, recent message counts
	os.WriteFile(filepath.Join(claude, "busy.jsonl"), []byte(assistant(fresh)), 0o644)
	if !m.anySessionActive(30 * time.Minute) {
		t.Fatal("a recent real message should count as active")
	}
}
