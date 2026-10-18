package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// A session that has been active while CCTokenViz could see it is "tracked": it stays on the
// list, dimmed and marked idle, with its last known context, until you resume it (it goes back
// to normal) or hide it with the X. The list is kept in ~/.cctokenviz/tracked.json so it
// survives restarts. Idle sessions cost nothing to keep: their logs don't change, so
// nothing is re-read.

// maxIdleTracked bounds how many idle sessions are remembered
const maxIdleTracked = 40

type trackedEntry struct {
	Path    string `json:"path"`    // the session's log file
	Project string `json:"project"` // the folder it lives in under ~/.claude/projects
}

var trackedMu sync.Mutex

func trackedPath() string {
	dir := statusDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "tracked.json")
}

func loadTracked() map[string]trackedEntry {
	tracked := map[string]trackedEntry{}
	if p := trackedPath(); p != "" {
		if data, err := os.ReadFile(p); err == nil {
			json.Unmarshal(data, &tracked)
		}
	}
	return tracked
}

func saveTracked(tracked map[string]trackedEntry) {
	p := trackedPath()
	if p == "" {
		return
	}
	os.MkdirAll(filepath.Dir(p), 0o755)
	if data, err := json.Marshal(tracked); err == nil {
		os.WriteFile(p, data, 0o644)
	}
}
