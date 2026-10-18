package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	MaxContextTokens = 200000 // fallback; see contextTokens in the config and the status line export
	MaxProjects      = 10
	ActiveMinutes    = 60
)

// SessionInfo represents a Claude Code session's context usage
type SessionInfo struct {
	ID          string `json:"id"`
	ProjectName string `json:"projectName"`
	ProjectPath string `json:"projectPath"`
	Title       string `json:"title"`
	UsedTokens  int    `json:"usedTokens"`
	FreeTokens  int    `json:"freeTokens"`
	Percentage  int    `json:"percentage"`
	LastUpdated string `json:"lastUpdated"`
	Idle        bool   `json:"idle"`        // no activity for a while: the last known context is shown
	IdleSeconds int64  `json:"idleSeconds"` // how long since the log last changed
}

// Monitor handles scanning and parsing Claude session files
type Monitor struct {
	claudeDir string
}

// Message structures for parsing JSONL
type AssistantMessage struct {
	Type    string `json:"type"`
	Cwd     string `json:"cwd"`
	AiTitle string `json:"aiTitle"`
	Message struct {
		Usage   *Usage          `json:"usage"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// NewMonitor creates a new monitor instance
func NewMonitor() *Monitor {
	homeDir, _ := os.UserHomeDir()
	claudeDir := filepath.Join(homeDir, ".claude", "projects")

	// Handle Windows path if needed
	if runtime.GOOS == "windows" {
		claudeDir = filepath.Join(homeDir, ".claude", "projects")
	}

	return &Monitor{
		claudeDir: claudeDir,
	}
}

// GetActiveSessions returns all active Claude sessions
func (m *Monitor) GetActiveSessions() []SessionInfo {
	var sessions []SessionInfo

	// Find all project directories
	projectDirs, err := os.ReadDir(m.claudeDir)
	if err != nil {
		return sessions
	}

	type sessionFile struct {
		path    string
		modTime time.Time
		project string
		idle    bool
	}

	var allSessions []sessionFile

	cfg := loadConfig()
	cutoff := time.Now().Add(-time.Duration(cfg.ActiveMinutes) * time.Minute)

	for _, projectDir := range projectDirs {
		if !projectDir.IsDir() {
			continue
		}

		projectPath := filepath.Join(m.claudeDir, projectDir.Name())
		files, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			// Skip agent files and non-jsonl files
			if !strings.HasSuffix(file.Name(), ".jsonl") || strings.HasPrefix(file.Name(), "agent-") {
				continue
			}

			filePath := filepath.Join(projectPath, file.Name())
			info, err := file.Info()
			if err != nil {
				continue
			}

			// Only include recently modified files
			if info.ModTime().After(cutoff) {
				allSessions = append(allSessions, sessionFile{
					path:    filePath,
					modTime: info.ModTime(),
					project: projectDir.Name(),
				})
			}
		}
	}

	// Remember sessions seen active, and bring back the ones that have gone idle
	trackedMu.Lock()
	tracked := loadTracked()
	changed := false
	activeIDs := map[string]bool{}
	for _, sf := range allSessions {
		id := filepath.Base(sf.path)
		activeIDs[id] = true
		if e, ok := tracked[id]; !ok || e.Path != sf.path || e.Project != sf.project {
			tracked[id] = trackedEntry{Path: sf.path, Project: sf.project}
			changed = true
		}
	}

	var idle []sessionFile
	for id, e := range tracked {
		if activeIDs[id] {
			continue
		}
		info, err := os.Stat(e.Path)
		if err != nil { // the log is gone
			delete(tracked, id)
			changed = true
			continue
		}
		idle = append(idle, sessionFile{path: e.Path, modTime: info.ModTime(), project: e.Project, idle: true})
	}
	// Keep only the most recent idle sessions
	sort.Slice(idle, func(i, j int) bool { return idle[i].modTime.After(idle[j].modTime) })
	if len(idle) > maxIdleTracked {
		for _, sf := range idle[maxIdleTracked:] {
			delete(tracked, filepath.Base(sf.path))
			changed = true
		}
		idle = idle[:maxIdleTracked]
	}
	allSessions = append(allSessions, idle...)

	// Sort by modification time (newest first)
	sort.Slice(allSessions, func(i, j int) bool {
		return allSessions[i].modTime.After(allSessions[j].modTime)
	})

	// Parse each session, keeping only configured projects. An idle session that you hid with the
	// X is forgotten; resuming it later makes it active again and it comes back.
	for _, sf := range allSessions {
		info := m.parseSession(sf.path, sf.project)
		if info == nil {
			continue
		}
		if isDismissed(info.ID, sf.modTime.Unix()) {
			if sf.idle {
				delete(tracked, info.ID)
				changed = true
			}
			continue
		}
		if len(sessions) >= MaxProjects || !cfg.allows(info.ProjectName) {
			continue
		}
		info.LastUpdated = sf.modTime.Format("15:04:05")
		info.Idle = sf.idle
		if sf.idle {
			info.IdleSeconds = int64(time.Since(sf.modTime).Seconds())
		}
		sessions = append(sessions, *info)
	}

	if changed {
		saveTracked(tracked)
	}
	trackedMu.Unlock()

	return sessions
}

// sessionState is what we have learned from a session log so far. Logs are append-only,
// so each refresh reads only the bytes added since the last one.
type sessionState struct {
	offset        int64 // bytes consumed, always ending on a line boundary
	lastUsage     *Usage
	cwd           string
	projectFolder string
	title         string
	signals       []string // recent project signals, newest last (see addSignal)
	lastCwdName   string   // the project the working folder last pointed at
}

const signalWindow = 8

// addSignal records that the session just pointed at a project (by moving into it or by
// touching its files) and returns the project to show: the most common one among the last few
// signals, the newest winning a tie. A session that works across two projects used to flip
// its row between them with every file it touched, which made rows appear and disappear.
func (st *sessionState) addSignal(name string) string {
	st.signals = append(st.signals, name)
	if len(st.signals) > signalWindow {
		st.signals = st.signals[len(st.signals)-signalWindow:]
	}
	counts := map[string]int{}
	best := ""
	for _, s := range st.signals {
		counts[s]++
		if best == "" || counts[s] >= counts[best] {
			best = s
		}
	}
	return best
}

var (
	sessionCache   = map[string]*sessionState{}
	sessionCacheMu sync.Mutex

	projectRoots   = map[string]bool{} // "projects" folders seen in session logs
)

func resetSessionCache() {
	sessionCacheMu.Lock()
	sessionCache = map[string]*sessionState{}
	sessionCacheMu.Unlock()
}

// parseSession returns context stats for a session file, reading only what is new since last time
func (m *Monitor) parseSession(filePath, projectDir string) *SessionInfo {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil
	}

	folders := loadConfig().ProjectFolders

	sessionCacheMu.Lock()
	defer sessionCacheMu.Unlock()

	st := sessionCache[filePath]
	if st == nil || stat.Size() < st.offset {
		// First sight, or the file shrank (replaced or rewritten): start over
		st = &sessionState{}
		sessionCache[filePath] = st
	}

	if stat.Size() > st.offset {
		if _, err := file.Seek(st.offset, io.SeekStart); err != nil {
			return nil
		}
		reader := bufio.NewReaderSize(file, 1024*1024)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				// A trailing partial line (no newline yet) is left for the next refresh
				break
			}
			st.offset += int64(len(line))

			var msg AssistantMessage
			if json.Unmarshal(line, &msg) != nil {
				continue
			}

			// Capture cwd from the first message that has it
			if st.cwd == "" && msg.Cwd != "" {
				st.cwd = msg.Cwd
			}

			// Sessions often start in a parent folder and move into a project,
			// so the latest project folder seen wins
			if name := projectFromPath(msg.Cwd, folders); name != "" && name != st.lastCwdName {
				st.lastCwdName = name // only a move into a project counts, not every line sitting in it
				st.projectFolder = st.addSignal(name)
			}
			// A session's folder is not always where its work is (it may start in a parent
			// folder), so the paths Claude is actually touching count too; the latest signal wins
			if msg.Type == "assistant" {
				if name := projectFromToolUse(msg.Message.Content, folders); name != "" {
					st.projectFolder = st.addSignal(name)
				}
			}
			if msg.AiTitle != "" {
				st.title = msg.AiTitle
			}

			if msg.Type == "assistant" && msg.Message.Usage != nil {
				st.lastUsage = msg.Message.Usage
			}
		}
	}

	lastUsage, cwd, projectFolder, title := st.lastUsage, st.cwd, st.projectFolder, st.title

	if lastUsage == nil {
		return nil
	}

	// Calculate context usage (matches official Claude Code status bar hook)
	totalUsed := lastUsage.InputTokens + lastUsage.CacheReadInputTokens + lastUsage.CacheCreationInputTokens

	// Use the project's real window size. Order: what this session reported, a per-project
	// override from Settings, what other sessions in the project reported, then the default.
	// Usage beyond 200k proves the window is bigger than the default.
	projectName := filepath.Base(cwd)
	if projectName == "" || projectName == "." {
		projectName = projectDir
	}
	if projectFolder != "" {
		projectName = projectFolder
	}

	cfg := loadConfig()
	maxTokens := MaxContextTokens
	if cfg.ContextTokens > 0 {
		maxTokens = cfg.ContextTokens
	}
	if size := cfg.contextFor(projectName); size > 0 {
		maxTokens = size
	} else if size := contextSizeForProject(projectName); size > 0 {
		maxTokens = size
	}
	if size := contextSizeFor(strings.TrimSuffix(filepath.Base(filePath), ".jsonl")); size > 0 {
		maxTokens = size
	}
	if totalUsed > maxTokens {
		maxTokens = 1000000
	}
	percentage := (totalUsed * 100) / maxTokens
	freeTokens := maxTokens - totalUsed

	if freeTokens < 0 {
		freeTokens = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	return &SessionInfo{
		ID:          filepath.Base(filePath),
		ProjectName: projectName,
		ProjectPath: projectDir,
		Title:       title,
		UsedTokens:  totalUsed,
		FreeTokens:  freeTokens,
		Percentage:  percentage,
	}
}


// KnownProjects returns the project names found in sessions from the last two weeks,
// for the project picker. It parses at most the three newest logs per folder.
func (m *Monitor) KnownProjects() []string {
	var names []string
	seen := map[string]bool{}
	dirs, err := os.ReadDir(m.claudeDir)
	if err != nil {
		return names
	}
	cutoff := time.Now().Add(-14 * 24 * time.Hour)

	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		type logFile struct {
			path string
			mod  time.Time
		}
		var logs []logFile
		files, err := os.ReadDir(filepath.Join(m.claudeDir, d.Name()))
		if err != nil {
			continue
		}
		for _, f := range files {
			if !strings.HasSuffix(f.Name(), ".jsonl") || strings.HasPrefix(f.Name(), "agent-") {
				continue
			}
			if info, err := f.Info(); err == nil && info.ModTime().After(cutoff) {
				logs = append(logs, logFile{filepath.Join(m.claudeDir, d.Name(), f.Name()), info.ModTime()})
			}
		}
		sort.Slice(logs, func(i, j int) bool { return logs[i].mod.After(logs[j].mod) })
		if len(logs) > 3 {
			logs = logs[:3]
		}
		for _, l := range logs {
			if info := m.parseSession(l.path, d.Name()); info != nil && !seen[info.ProjectName] {
				seen[info.ProjectName] = true
				names = append(names, info.ProjectName)
			}
		}
	}
	// Also offer project folders that have no recent session
	sessionCacheMu.Lock()
	seenDirs := make([]string, 0, len(projectRoots))
	for r := range projectRoots {
		seenDirs = append(seenDirs, r)
	}
	sessionCacheMu.Unlock()
	for _, n := range discoverableProjects(loadConfig().ProjectFolders, seenDirs) {
		if !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	sort.Slice(names, func(i, j int) bool { return strings.ToLower(names[i]) < strings.ToLower(names[j]) })
	return names
}

var toolPathRe = regexp.MustCompile(`(?:[A-Za-z]:[\\/]|/[A-Za-z]/|/(?:home|Users|mnt|opt|srv|var|workspace|tmp)/)[^\s"'&|;<>()*?]+`)

// projectFromToolUse returns the project named by the last file path or command in an
// assistant message's tool calls, or "".
func projectFromToolUse(content json.RawMessage, folders []string) string {
	if len(content) == 0 || content[0] != '[' || !bytes.Contains(content, []byte(`"tool_use"`)) {
		return ""
	}
	var blocks []struct {
		Type  string          `json:"type"`
		Input json.RawMessage `json:"input"`
	}
	if json.Unmarshal(content, &blocks) != nil {
		return ""
	}
	name := ""
	for _, b := range blocks {
		if b.Type != "tool_use" {
			continue
		}
		// Decode the JSON string escapes so paths read normally
		var in map[string]any
		if json.Unmarshal(b.Input, &in) != nil {
			continue
		}
		for _, key := range []string{"file_path", "path", "notebook_path", "command"} {
			v, _ := in[key].(string)
			if key == "command" {
				for _, m := range toolPathRe.FindAllString(v, -1) {
					if n := projectFromPath(strings.TrimRight(m, ".,:"), folders); n != "" {
						name = n
					}
				}
			} else if n := projectFromPath(v, folders); n != "" {
				name = n
			}
		}
	}
	return name
}
