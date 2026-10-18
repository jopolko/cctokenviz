package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UsageSummary is what Claude Code has used over a period, priced at public API rates.
// For a subscription this is notional (what the same tokens would cost on the API), not a bill.
type UsageSummary struct {
	Label         string  `json:"label"`
	InputTokens   int     `json:"inputTokens"` // fresh input, not served from cache
	OutputTokens  int     `json:"outputTokens"`
	CacheRead     int     `json:"cacheRead"` // input served from the prompt cache
	CacheWrite    int     `json:"cacheWrite"`
	TotalTokens   int     `json:"totalTokens"`
	Cost          float64 `json:"cost"`          // USD, estimated
	CacheSaved    float64 `json:"cacheSaved"`    // USD saved by cache reads versus paying full input price
	CacheSavedTok int     `json:"cacheSavedTok"` // tokens that did not need to be processed fresh
}

// UsageReport is the panel's data: today and the last seven days.
type UsageReport struct {
	Periods []UsageSummary `json:"periods"`
	Approx  bool           `json:"approx"` // some model had no known price and a mid-tier rate was assumed
}

// price is USD per million tokens
type price struct{ in, out float64 }

func priceFor(model string) (price, bool) {
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "opus"):
		for _, old := range []string{"opus-4-1", "opus-4-0", "opus-4-2025", "claude-3-opus", "opus-4@"} {
			if strings.Contains(m, old) {
				return price{15, 75}, true
			}
		}
		if m == "claude-opus-4" {
			return price{15, 75}, true
		}
		return price{5, 25}, true
	case strings.Contains(m, "sonnet"):
		return price{3, 15}, true
	case strings.Contains(m, "haiku"):
		if strings.Contains(m, "3-5") {
			return price{0.8, 4}, true
		}
		return price{1, 5}, true
	}
	return price{3, 15}, false // unknown model: assume mid-tier so the total is in the right range
}

type usageEntry struct {
	at    time.Time
	model string
	u     Usage1
}

// Usage1 mirrors Usage; kept separate so entries stay small and comparable
type Usage1 struct{ in, out, read, write int }

type usageFile struct {
	offset  int64
	entries map[string]usageEntry // by message id: a message is logged several times as it streams, keep the last
}

var (
	usageFiles   = map[string]*usageFile{}
	usageFilesMu sync.Mutex
)

type usageLine struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
	Type      string `json:"type"`
	Message   struct {
		ID    string `json:"id"`
		Model string `json:"model"`
		Usage *Usage `json:"usage"`
	} `json:"message"`
}

// readUsageFile reads only what was appended since last time.
func readUsageFile(path string) *usageFile {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil
	}
	uf := usageFiles[path]
	if uf == nil || stat.Size() < uf.offset {
		uf = &usageFile{entries: map[string]usageEntry{}}
		usageFiles[path] = uf
	}
	if stat.Size() == uf.offset {
		return uf
	}
	if _, err := f.Seek(uf.offset, io.SeekStart); err != nil {
		return uf
	}
	r := bufio.NewReaderSize(f, 1024*1024)
	n := 0
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			break
		}
		uf.offset += int64(len(line))
		// Cheap filter before decoding: only assistant lines carry usage
		if !bytes.Contains(line, []byte(`"usage"`)) {
			continue
		}
		var l usageLine
		if json.Unmarshal(line, &l) != nil || l.Type != "assistant" || l.Message.Usage == nil {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, l.Timestamp)
		if err != nil {
			continue
		}
		model := l.Message.Model
		if model == "<synthetic>" {
			continue
		}
		key := l.Message.ID + "|" + l.RequestID
		if l.Message.ID == "" {
			n++
			key = "line" + strconv.Itoa(n)
		}
		u := l.Message.Usage
		uf.entries[key] = usageEntry{at: at, model: model, u: Usage1{u.InputTokens, u.OutputTokens, u.CacheReadInputTokens, u.CacheCreationInputTokens}}
	}
	return uf
}

// GetUsage tallies tokens and estimated cost for today and the last seven days, across every
// session log (including sub-agents) touched in that time.
func (m *Monitor) GetUsage() UsageReport {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := now.Add(-7 * 24 * time.Hour)
	periods := []struct {
		label string
		since time.Time
	}{{"Today", midnight}, {"7 days", weekAgo}}
	out := UsageReport{Periods: make([]UsageSummary, len(periods))}
	for i, p := range periods {
		out.Periods[i].Label = p.label
	}

	usageFilesMu.Lock()
	defer usageFilesMu.Unlock()

	seen := map[string]bool{}
	dirs, _ := os.ReadDir(m.claudeDir)
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		dir := filepath.Join(m.claudeDir, d.Name())
		files, _ := os.ReadDir(dir)
		for _, fe := range files {
			if !strings.HasSuffix(fe.Name(), ".jsonl") {
				continue
			}
			info, err := fe.Info()
			if err != nil || info.ModTime().Before(weekAgo) {
				continue
			}
			path := filepath.Join(dir, fe.Name())
			seen[path] = true
			uf := readUsageFile(path)
			if uf == nil {
				continue
			}
			for _, e := range uf.entries {
				p, known := priceFor(e.model)
				if !known {
					out.Approx = true
				}
				cost := (float64(e.u.in)*p.in + float64(e.u.out)*p.out +
					float64(e.u.read)*p.in*0.1 + float64(e.u.write)*p.in*1.25) / 1e6
				saved := float64(e.u.read) * p.in * 0.9 / 1e6
				for i, per := range periods {
					if e.at.Before(per.since) {
						continue
					}
					s := &out.Periods[i]
					s.InputTokens += e.u.in
					s.OutputTokens += e.u.out
					s.CacheRead += e.u.read
					s.CacheWrite += e.u.write
					s.TotalTokens += e.u.in + e.u.out + e.u.read + e.u.write
					s.Cost += cost
					s.CacheSaved += saved
					s.CacheSavedTok += e.u.read
				}
			}
		}
	}
	// Forget logs that fell out of the window
	for path := range usageFiles {
		if !seen[path] {
			delete(usageFiles, path)
		}
	}
	return out
}
