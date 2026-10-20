package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BreakdownRow is one prompt and how much it added to the context
type BreakdownRow struct {
	Label   string `json:"label"`   // what you asked; a short reply such as "y" is shown with the question it answered
	Actions string `json:"actions"` // what Claude did to answer: "read 4 files, ran 3 commands, edited 2 files"
	Reply   string `json:"reply"`   // Claude's one-line answer, for a tooltip
	Tokens  int    `json:"tokens"`
	At      int64  `json:"at"` // unix seconds of the prompt
	reply   string // Claude's last text reply in the turn
	acts    *actions
}

// Breakdown is what has filled a session's context: what it started with, then what each prompt added
type Breakdown struct {
	StartLabel  string         `json:"startLabel"` // "Carried over" after a compact, else "Starting context"
	StartAt     int64          `json:"startAt"`    // unix seconds the starting context began (compact or session start)
	Start       int            `json:"start"`      // tokens in the context at the first reply, before any prompt shown
	Rows        []BreakdownRow `json:"rows"`
	SmallCount  int            `json:"smallCount"`  // prompts under smallRowTokens, folded into one line
	SmallTokens int            `json:"smallTokens"` // what those added together
	Total       int            `json:"total"`
}

// actions counts what Claude did in one turn. Everything it reads or prints stays in the context,
// so this is why a turn was big.
type actions struct {
	read, edited map[string]bool
	seen         map[string]bool
	commands     int
	searches     int
	web          int
	agents       int
	other        int
}

func newActions() *actions {
	return &actions{read: map[string]bool{}, edited: map[string]bool{}, seen: map[string]bool{}}
}

func (a *actions) add(c toolCall) {
	if c.id != "" { // a message can be logged more than once as it streams
		if a.seen[c.id] {
			return
		}
		a.seen[c.id] = true
	}
	name, target := c.name, c.target
	switch name {
	case "Read":
		a.read[target] = true
	case "Edit", "Write", "MultiEdit", "NotebookEdit":
		a.edited[target] = true
	case "Bash", "PowerShell":
		a.commands++
	case "Grep", "Glob":
		a.searches++
	case "WebFetch", "WebSearch":
		a.web++
	case "Agent", "Task":
		a.agents++
	default:
		a.other++
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return strconv.Itoa(n) + " " + many
}

// String says it in words, or "just answered" when no tools were used
func (a *actions) String() string {
	var parts []string
	if n := len(a.read); n > 0 {
		parts = append(parts, "read "+plural(n, "file", "files"))
	}
	if a.commands > 0 {
		parts = append(parts, "ran "+plural(a.commands, "command", "commands"))
	}
	if n := len(a.edited); n > 0 {
		parts = append(parts, "edited "+plural(n, "file", "files"))
	}
	if a.searches > 0 {
		parts = append(parts, "searched "+plural(a.searches, "time", "times"))
	}
	if a.web > 0 {
		parts = append(parts, "fetched "+plural(a.web, "web page", "web pages"))
	}
	if a.agents > 0 {
		parts = append(parts, "started "+plural(a.agents, "helper", "helpers"))
	}
	if len(parts) == 0 {
		if a.other > 0 { // housekeeping calls (task lists and the like) are not worth listing beside real work
			return "Claude used other tools"
		}
		return "Claude just answered"
	}
	return "Claude " + strings.Join(parts, ", ")
}

type toolCall struct{ id, name, target string }

// toolUses lists the tool calls in an assistant message, each with the file or command it targeted
func toolUses(content json.RawMessage) (calls []toolCall) {
	var blocks []struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Name  string `json:"name"`
		Input struct {
			FilePath string `json:"file_path"`
			Path     string `json:"path"`
			Notebook string `json:"notebook_path"`
		} `json:"input"`
	}
	if json.Unmarshal(content, &blocks) != nil {
		return nil
	}
	for _, b := range blocks {
		if b.Type != "tool_use" {
			continue
		}
		target := b.Input.FilePath
		if target == "" {
			target = b.Input.Notebook
		}
		if target == "" {
			target = b.Input.Path
		}
		calls = append(calls, toolCall{id: b.ID, name: b.Name, target: target})
	}
	return calls
}

const (
	maxBreakdownRows = 40
	smallRowTokens   = 3000 // prompts that added less than this are folded into one line
)

type breakdownLine struct {
	Type             string `json:"type"`
	Subtype          string `json:"subtype"`
	Timestamp        string `json:"timestamp"`
	IsMeta           bool   `json:"isMeta"`
	IsSidechain      bool   `json:"isSidechain"`
	IsCompactSummary bool   `json:"isCompactSummary"`
	Message          struct {
		Usage   *Usage          `json:"usage"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// GetBreakdown reads one session's log and attributes growth in the context to the prompts
// that caused it: each reply records the whole context size, so a prompt's share is the
// difference between the size before and after its replies. The numbers are exact; only the
// labels are condensed. It starts over at the latest compact, since the summary replaced
// everything before it.
func (m *Monitor) GetBreakdown(id string) Breakdown {
	path := m.findSessionFile(id)
	if path == "" {
		return Breakdown{}
	}
	file, err := os.Open(path)
	if err != nil {
		return Breakdown{}
	}
	defer file.Close()
	return buildBreakdown(file)
}

// findSessionFile finds the log for a session id (its file name) under any project folder
func (m *Monitor) findSessionFile(id string) string {
	if id == "" || id != filepath.Base(id) || !strings.HasSuffix(id, ".jsonl") {
		return ""
	}
	dirs, err := os.ReadDir(m.claudeDir)
	if err != nil {
		return ""
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		p := filepath.Join(m.claudeDir, d.Name(), id)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func buildBreakdown(r io.Reader) Breakdown {
	var (
		b          Breakdown
		ctx        int
		started    bool // the first reply of this stretch has been seen
		curBase    int  // the context size when the current prompt was sent
		curIdx     = -1 // the current prompt's row, once the stretch has started
		pendingAsk string
		sawCompact bool
	)
	reset := func() {
		b = Breakdown{}
		ctx, started, curIdx, pendingAsk = 0, false, -1, ""
	}
	reader := bufio.NewReaderSize(r, 1024*1024)
	for {
		line, err := reader.ReadBytes(0x0a)
		if len(line) > 0 {
			var l breakdownLine
			if json.Unmarshal(line, &l) == nil && !l.IsSidechain {
				at := parseUnix(l.Timestamp)
				switch {
				case l.Type == "system" && l.Subtype == "compact_boundary":
					reset()
					sawCompact = true
					b.StartAt = at
				case l.Type == "user" && !l.IsMeta && !l.IsCompactSummary:
					// Before the first reply, a prompt is part of the starting context
					if text, ok := promptText(l.Message.Content); ok && started {
						b.Rows = append(b.Rows, BreakdownRow{Label: promptLabel(text, pendingAsk), At: at, acts: newActions()})
						curIdx, curBase = len(b.Rows)-1, ctx
					}
				case l.Type == "assistant" && l.Message.Usage != nil:
					u := l.Message.Usage
					ctx = u.InputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens
					if !started {
						started = true
						b.Start = ctx
						if b.StartAt == 0 {
							b.StartAt = at
						}
					} else if curIdx >= 0 {
						b.Rows[curIdx].Tokens = ctx - curBase
					}
					if curIdx >= 0 {
						for _, c := range toolUses(l.Message.Content) {
							b.Rows[curIdx].acts.add(c)
						}
					}
					if t := lastText(l.Message.Content); t != "" {
						pendingAsk = t
						if curIdx >= 0 {
							b.Rows[curIdx].reply = t
						}
					}
				}
			}
		}
		if err != nil {
			break
		}
	}
	b.Total = ctx
	b.StartLabel = "Starting context"
	if sawCompact {
		b.StartLabel = "Carried over"
	}
	rows := b.Rows[:0]
	for _, row := range b.Rows {
		switch {
		case row.Tokens >= smallRowTokens:
			row.Reply = firstSentence(row.reply)
			row.Actions = row.acts.String()
			rows = append(rows, row)
		case row.Tokens > 0:
			b.SmallCount++
			b.SmallTokens += row.Tokens
		}
	}
	if len(rows) > maxBreakdownRows {
		rows = rows[len(rows)-maxBreakdownRows:]
	}
	b.Rows = rows
	return b
}

func parseUnix(ts string) int64 {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return 0
	}
	return t.Unix()
}

// promptText returns what the user typed, or false for tool results, hooks, command output
// and other entries that are not a prompt
func promptText(content json.RawMessage) (string, bool) {
	var s string
	if json.Unmarshal(content, &s) == nil {
		return cleanPrompt(s)
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(content, &blocks) != nil {
		return "", false
	}
	var text []string
	for _, b := range blocks {
		if b.Type == "tool_result" {
			return "", false
		}
		if b.Type == "text" {
			text = append(text, b.Text)
		}
	}
	return cleanPrompt(strings.Join(text, " "))
}

func cleanPrompt(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "<") || strings.HasPrefix(s, "[Request interrupted") {
		return "", false
	}
	return s, true
}

// lastText is the text of an assistant message, ignoring tool calls
func lastText(content json.RawMessage) string {
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(content, &blocks) != nil {
		return ""
	}
	out := ""
	for _, b := range blocks {
		if b.Type == "text" && strings.TrimSpace(b.Text) != "" {
			out = b.Text
		}
	}
	return out
}

var spaces = regexp.MustCompile(`\s+`)
var sentenceEnd = regexp.MustCompile(`[^.!?\n]*\?`)

const labelWidth = 110

// promptLabel names a prompt by its first words. A short reply such as "y" says nothing on its
// own, so it is shown with the question it answered: the last question in Claude's message.
func promptLabel(prompt, ask string) string {
	p := spaces.ReplaceAllString(prompt, " ")
	if len([]rune(p)) <= 16 && strings.Count(p, " ") < 3 && ask != "" {
		q := lastQuestion(ask)
		if q == "" { // no question: fall back to what Claude opened with
			first := strings.SplitN(strings.TrimSpace(ask), "\n", 2)[0]
			q = strings.TrimSpace(spaces.ReplaceAllString(strings.NewReplacer("*", "", "`", "", "#", "").Replace(first), " "))
		}
		if q != "" {
			return cut(p, 10) + " → " + cut(q, labelWidth-3)
		}
	}
	return cut(p, labelWidth)
}

var codeBlock = regexp.MustCompile("(?s)```.*?```")

func lastQuestion(s string) string {
	s = codeBlock.ReplaceAllString(s, " ")
	m := sentenceEnd.FindAllString(s, -1)
	if len(m) == 0 {
		return ""
	}
	return strings.TrimSpace(spaces.ReplaceAllString(strings.NewReplacer("*", "", "`", "").Replace(m[len(m)-1]), " "))
}

func cut(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n-1])) + "…"
}

var (
	mdLink      = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	firstClause = regexp.MustCompile(`^(.{20,110}?[.!?:])\s`)
)

// firstSentence is how Claude opened its last reply in a turn, which is usually what it did:
// markdown stripped, the first sentence if there is one, otherwise the first hundred characters
func firstSentence(reply string) string {
	s := codeBlock.ReplaceAllString(reply, " ")
	s = mdLink.ReplaceAllString(s, "$1")
	s = strings.NewReplacer("*", "", "`", "", "#", "", "_", " ").Replace(s)
	s = strings.TrimSpace(spaces.ReplaceAllString(s, " "))
	if s == "" {
		return ""
	}
	if m := firstClause.FindStringSubmatch(s + " "); m != nil {
		s = m[1]
	}
	return cut(strings.TrimRight(s, ".:"), 110)
}
