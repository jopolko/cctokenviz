package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Rows keep the order you drag them into. The order is a list of session ids in
// ~/.headroom/order.json; sessions not in the list go after the ones that are,
// newest activity first.

var orderMu sync.Mutex

func orderPath() string {
	dir := statusDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "order.json")
}

func loadOrder() []string {
	var ids []string
	if p := orderPath(); p != "" {
		if data, err := os.ReadFile(p); err == nil {
			json.Unmarshal(data, &ids)
		}
	}
	return ids
}

// applyOrder sorts sessions by the saved order; unknown sessions keep their incoming order, at the end
func applyOrder(sessions []SessionInfo) []SessionInfo {
	orderMu.Lock()
	saved := loadOrder()
	orderMu.Unlock()

	rank := make(map[string]int, len(saved))
	for i, id := range saved {
		rank[id] = i
	}

	out := make([]SessionInfo, 0, len(sessions))
	var unknown []SessionInfo
	byID := make(map[string]SessionInfo, len(sessions))
	for _, s := range sessions {
		byID[s.ID] = s
		if _, ok := rank[s.ID]; !ok {
			unknown = append(unknown, s)
		}
	}
	for _, id := range saved {
		if s, ok := byID[id]; ok {
			out = append(out, s)
		}
	}
	return append(out, unknown...)
}

// saveOrder stores the dragged order, keeping the places of sessions that aren't on screen right now
func saveOrder(visible []string) {
	orderMu.Lock()
	defer orderMu.Unlock()
	p := orderPath()
	if p == "" {
		return
	}

	seen := make(map[string]bool, len(visible))
	merged := make([]string, 0, len(visible))
	for _, id := range visible {
		if !seen[id] {
			seen[id] = true
			merged = append(merged, id)
		}
	}
	for _, id := range loadOrder() {
		if !seen[id] {
			seen[id] = true
			merged = append(merged, id)
		}
	}
	if len(merged) > 200 {
		merged = merged[:200]
	}

	os.MkdirAll(filepath.Dir(p), 0o755)
	if data, err := json.Marshal(merged); err == nil {
		os.WriteFile(p, data, 0o644)
	}
}
