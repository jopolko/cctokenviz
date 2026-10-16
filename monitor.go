package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
		Usage *Usage `json:"usage"`
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

	// Sort by modification time (newest first)
	sort.Slice(allSessions, func(i, j int) bool {
		return allSessions[i].modTime.After(allSessions[j].modTime)
	})

	// Parse each session, keeping only configured projects
	for _, sf := range allSessions {
		if len(sessions) >= MaxProjects {
			break
		}
		info := m.parseSession(sf.path, sf.project)
		if info != nil && cfg.allows(info.ProjectName) && !isDismissed(info.ID, sf.modTime.Unix()) {
			info.LastUpdated = sf.modTime.Format("15:04:05")
			sessions = append(sessions, *info)
		}
	}

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
}

var (
	sessionCache   = map[string]*sessionState{}
	sessionCacheMu sync.Mutex
)

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
			if name := projectFromCwd(msg.Cwd); name != "" {
				st.projectFolder = name
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

	// Prefer the window size Claude Code reports for this session; fall back to the default
	maxTokens := MaxContextTokens
	if cfg := loadConfig(); cfg.ContextTokens > 0 {
		maxTokens = cfg.ContextTokens
	}
	if size := contextSizeFor(strings.TrimSuffix(filepath.Base(filePath), ".jsonl")); size > 0 {
		maxTokens = size
	}
	percentage := (totalUsed * 100) / maxTokens
	freeTokens := maxTokens - totalUsed

	if freeTokens < 0 {
		freeTokens = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	// Use folder name from cwd if available, otherwise fall back to directory name
	projectName := filepath.Base(cwd)
	if projectName == "" || projectName == "." {
		projectName = projectDir
	}
	if projectFolder != "" {
		projectName = projectFolder
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

