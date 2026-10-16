package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Sessions you X out stay hidden until they see new activity.
// Stored as session id -> unix time dismissed in ~/.headroom/dismissed.json.

var dismissMu sync.Mutex

func dismissedPath() string {
	dir := statusDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "dismissed.json")
}

func loadDismissed() map[string]int64 {
	m := map[string]int64{}
	if p := dismissedPath(); p != "" {
		if data, err := os.ReadFile(p); err == nil {
			json.Unmarshal(data, &m)
		}
	}
	return m
}

// isDismissed is true if the session was dismissed and hasn't been modified since
func isDismissed(sessionID string, modUnix int64) bool {
	dismissMu.Lock()
	defer dismissMu.Unlock()
	at, ok := loadDismissed()[sessionID]
	return ok && modUnix <= at
}

func dismissSession(sessionID string) {
	dismissMu.Lock()
	defer dismissMu.Unlock()
	p := dismissedPath()
	if p == "" {
		return
	}
	m := loadDismissed()
	m[sessionID] = time.Now().Unix()
	os.MkdirAll(filepath.Dir(p), 0o755)
	if data, err := json.Marshal(m); err == nil {
		os.WriteFile(p, data, 0o644)
	}
}
