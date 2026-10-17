package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// An idle session keeps its last known reading until it is resumed or dismissed.
func TestIdleSessionsStayListed(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home) // Windows
	t.Setenv("HOME", home)        // everything else

	projects := filepath.Join(home, ".claude", "projects", "C--Users-me-projects-app")
	if err := os.MkdirAll(projects, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(projects, "s1.jsonl")
	appendFile(t, path, assistantLine(`C:\\Users\\me\\projects\\app`, 2, 4000))
	m := &Monitor{claudeDir: filepath.Join(home, ".claude", "projects")}

	// Active: listed, not idle
	got := m.GetActiveSessions()
	if len(got) != 1 || got[0].Idle || got[0].UsedTokens != 4002 {
		t.Fatalf("active session: %+v", got)
	}

	// Two hours of silence: still listed, marked idle, same reading
	old := time.Now().Add(-2 * time.Hour)
	os.Chtimes(path, old, old)
	got = m.GetActiveSessions()
	if len(got) != 1 || !got[0].Idle || got[0].UsedTokens != 4002 || got[0].IdleSeconds < 7000 {
		t.Fatalf("idle session: %+v", got)
	}

	// Resumed: new content, no longer idle, new reading
	appendFile(t, path, assistantLine(`C:\\Users\\me\\projects\\app`, 3, 9000))
	got = m.GetActiveSessions()
	if len(got) != 1 || got[0].Idle || got[0].UsedTokens != 9003 {
		t.Fatalf("resumed session: %+v", got)
	}

	// Idle again, then dismissed: gone, and forgotten
	os.Chtimes(path, old, old)
	dismissSession(got[0].ID)
	time.Sleep(1100 * time.Millisecond) // dismissal is compared in whole seconds
	if got = m.GetActiveSessions(); len(got) != 0 {
		t.Fatalf("dismissed idle session still listed: %+v", got)
	}
	if _, still := loadTracked()["s1.jsonl"]; still {
		t.Fatal("dismissed idle session is still tracked")
	}

	// A session never seen active is not listed, however recent its history
	other := filepath.Join(projects, "never-seen.jsonl")
	appendFile(t, other, assistantLine(`C:\\Users\\me\\projects\\app`, 1, 100))
	os.Chtimes(other, old, old)
	if got = m.GetActiveSessions(); len(got) != 0 {
		t.Fatalf("an old, never-tracked session appeared: %+v", got)
	}
}
