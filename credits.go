package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// Credits is what the panel shows under the plan limits.
type Credits struct {
	ClaudeAi string `json:"claudeAi"` // as typed in the config, e.g. "CA$81.92"

	HasConsole       bool    `json:"hasConsole"`
	ConsoleBalance   float64 `json:"consoleBalance"`   // snapshot minus spend since snapshot, if known
	ConsoleSnapshot  float64 `json:"consoleSnapshot"`  // the balance you typed
	ConsoleAsOf      string  `json:"consoleAsOf"`      // YYYY-MM-DD
	ConsoleSpentSince float64 `json:"consoleSpentSince"` // USD spent since asOf; 0 when unknown
	ConsoleEstimated bool    `json:"consoleEstimated"` // true when spend since the snapshot was subtracted
	ConsoleNote      string  `json:"consoleNote"`      // why it isn't estimated, if so
}

const costRefresh = 10 * time.Minute

type costCache struct {
	mu      sync.Mutex
	asOf    string
	fetched time.Time
	spent   float64
	err     string
}

var costs costCache

// GetCredits builds the credits display from the config plus, when an admin key is set,
// spend since the snapshot date from the Admin API cost report.
func GetCredits() Credits {
	cfg := loadConfig()
	out := Credits{ClaudeAi: cfg.Credits.ClaudeAi}
	c := cfg.Credits.Console
	if c == nil {
		return out
	}

	out.HasConsole = true
	out.ConsoleSnapshot = c.Balance
	out.ConsoleBalance = c.Balance
	out.ConsoleAsOf = c.AsOf

	key := os.Getenv("ANTHROPIC_ADMIN_KEY")
	if key == "" {
		out.ConsoleNote = "Set ANTHROPIC_ADMIN_KEY to track spend"
		return out
	}

	spent, errMsg := cachedSpendSince(key, c.AsOf)
	if errMsg != "" {
		out.ConsoleNote = errMsg
		return out
	}
	out.ConsoleEstimated = true
	out.ConsoleSpentSince = spent
	out.ConsoleBalance = c.Balance - spent
	return out
}

func cachedSpendSince(key, asOf string) (float64, string) {
	costs.mu.Lock()
	defer costs.mu.Unlock()

	if costs.asOf == asOf && time.Since(costs.fetched) < costRefresh {
		return costs.spent, costs.err
	}
	costs.asOf = asOf
	costs.fetched = time.Now()
	costs.spent, costs.err = fetchSpendSince(key, asOf)
	return costs.spent, costs.err
}

type costReport struct {
	Data []struct {
		Results []struct {
			Amount json.Number `json:"amount"` // USD in cents, as a decimal string
		} `json:"results"`
	} `json:"data"`
	HasMore  bool   `json:"has_more"`
	NextPage string `json:"next_page"`
}

func fetchSpendSince(key, asOf string) (float64, string) {
	start, err := time.Parse("2006-01-02", asOf)
	if err != nil {
		return 0, "credits.console.asOf must be YYYY-MM-DD"
	}

	var totalCents float64
	page := ""
	for i := 0; i < 20; i++ { // bounded pagination
		q := url.Values{}
		q.Set("starting_at", start.UTC().Format(time.RFC3339))
		q.Set("ending_at", time.Now().UTC().Format(time.RFC3339))
		q.Set("limit", "31")
		if page != "" {
			q.Set("page", page)
		}
		req, _ := http.NewRequest("GET", "https://api.anthropic.com/v1/organizations/cost_report?"+q.Encode(), nil)
		req.Header.Set("x-api-key", key)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("User-Agent", "headroom/1.0")

		resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
		if err != nil {
			return 0, "Cost report unreachable"
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return 0, fmt.Sprintf("Cost report error %d (check admin key)", resp.StatusCode)
		}

		var r costReport
		if json.Unmarshal(body, &r) != nil {
			return 0, "Cost report not understood"
		}
		for _, b := range r.Data {
			for _, res := range b.Results {
				if v, err := res.Amount.Float64(); err == nil {
					totalCents += v
				}
			}
		}
		if !r.HasMore || r.NextPage == "" {
			break
		}
		page = r.NextPage
	}
	return totalCents / 100, ""
}
