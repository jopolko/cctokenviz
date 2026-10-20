package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	monitor  *Monitor
	stopChan chan struct{}

	panelMu   sync.Mutex
	panelPrev *WindowRect // the window as it was before a panel grew it; nil when no panel is open

	settingsOpen atomic.Bool  // the Settings panel is showing, so the window's X should close it, not the app
	closeAskedAt atomic.Int64 // unix nanos of the last close we turned into "close Settings"
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		stopChan: make(chan struct{}),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.monitor = NewMonitor()

	// Keep the plan limits fresh in the background (opt-in, see poll.go)
	go a.limitsPoller()
}

// domReady is called when the DOM is ready, positions window in top right
func (a *App) domReady(ctx context.Context) {
	// Get screen dimensions
	screens, err := wailsRuntime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return
	}

	// Use primary screen
	screen := screens[0]
	for _, s := range screens {
		if s.IsPrimary {
			screen = s
			break
		}
	}

	// Window width (must match main.go)
	windowWidth := 240

	// Calculate top right position with small margin
	margin := 20
	x := screen.Size.Width - windowWidth - margin
	y := margin

	wailsRuntime.WindowSetPosition(ctx, x, y)
}

// WindowRect is a window's size and position, in the units the Wails window calls use
type WindowRect struct {
	W int `json:"w"`
	H int `json:"h"`
	X int `json:"x"`
	Y int `json:"y"`
}

// taskbarAllowance keeps a panel clear of the taskbar, which Wails' screen size includes
const taskbarAllowance = 56

// ExpandForPanel grows the window so Settings has room, as a drop-down under the gear: the
// top edge stays put and the right edge stays put (the gear is at the top right), so the window
// widens to the left and extends downward. If that would run off the screen it slides up or
// left just enough, and is shortened only if it still does not fit. It returns the window's
// previous size and position so the caller can restore them.
func (a *App) ExpandForPanel(minW, minH int) WindowRect {
	a.panelMu.Lock()
	defer a.panelMu.Unlock()
	if a.panelPrev != nil { // already grown (e.g. Settings, then "Run setup again"): keep the original
		return *a.panelPrev
	}
	w, h := wailsRuntime.WindowGetSize(a.ctx)
	x, y := wailsRuntime.WindowGetPosition(a.ctx)
	prev := WindowRect{W: w, H: h, X: x, Y: y}
	a.panelPrev = &prev

	newW, newH := max(w, minW), max(h, minH)
	newX, newY := x+w-newW, y // keep the right edge, so the gear stays where it was

	if screens, err := wailsRuntime.ScreenGetAll(a.ctx); err == nil && len(screens) > 0 {
		scr := screens[0]
		for _, s := range screens {
			if s.IsCurrent {
				scr = s
				break
			}
		}
		maxH := scr.Size.Height - taskbarAllowance
		newW = min(newW, scr.Size.Width)
		newH = min(newH, maxH)
		newX = max(0, min(newX, scr.Size.Width-newW))
		newY = max(0, min(newY, maxH-newH))
	}

	wailsRuntime.WindowSetSize(a.ctx, newW, newH)
	wailsRuntime.WindowSetPosition(a.ctx, newX, newY)
	return prev
}

// RestoreFromPanel puts the window back to the size and position it had before ExpandForPanel,
// kept on screen. The original is remembered here, in one place, rather than by the page.
func (a *App) RestoreFromPanel() {
	a.panelMu.Lock()
	defer a.panelMu.Unlock()
	prev := a.panelPrev
	if prev == nil {
		return
	}
	a.panelPrev = nil
	x, y := prev.X, prev.Y
	if screens, err := wailsRuntime.ScreenGetAll(a.ctx); err == nil && len(screens) > 0 {
		scr := screens[0]
		for _, s := range screens {
			if s.IsCurrent {
				scr = s
				break
			}
		}
		x = max(0, min(x, scr.Size.Width-prev.W))
		y = max(0, min(y, scr.Size.Height-taskbarAllowance-prev.H))
	}
	wailsRuntime.WindowSetSize(a.ctx, prev.W, prev.H)
	wailsRuntime.WindowSetPosition(a.ctx, x, y)
}

// SetSettingsOpen is told by the frontend when the Settings panel opens or closes
func (a *App) SetSettingsOpen(open bool) {
	a.settingsOpen.Store(open)
}

// beforeClose runs when the window's X is clicked. While Settings is open it closes Settings
// and keeps the app running. A second click within three seconds quits anyway, so a stuck
// panel can never trap the window.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	if !a.settingsOpen.Load() {
		return false
	}
	now := time.Now().UnixNano()
	if now-a.closeAskedAt.Load() < int64(3*time.Second) {
		return false
	}
	a.closeAskedAt.Store(now)
	wailsRuntime.EventsEmit(ctx, "settings:close")
	return true
}

// shutdown is called when the app closes
func (a *App) shutdown(ctx context.Context) {
	close(a.stopChan)
}

// GetSessions returns all active Claude sessions with their context usage
func (a *App) GetSessions() []SessionInfo {
	return applyOrder(a.monitor.GetActiveSessions())
}

// GetLimits returns the account's 5-hour and weekly plan limits, if a session has reported them
func (a *App) GetLimits() Limits {
	return a.monitor.GetLimits()
}

// RefreshLimits polls plan limits now (if polling is enabled) and returns the merged result
func (a *App) RefreshLimits() Limits {
	a.monitor.RefreshLimitsNow()
	return a.monitor.GetLimits()
}

// GetUsage returns token and estimated cost totals for today and the last 7 days
func (a *App) GetUsage() UsageReport {
	return a.monitor.GetUsage()
}

// GetSettings returns the saved settings
func (a *App) GetSettings() Config {
	return loadConfig()
}

// SaveSettings stores settings from the wizard or settings panel
func (a *App) SaveSettings(cfg Config) error {
	return saveConfig(cfg)
}

// SetOrder remembers the order the rows were dragged into
func (a *App) SetOrder(ids []string) {
	saveOrder(ids)
}

// HasClaudeCLI reports whether the claude command is available for limit polling
func (a *App) HasClaudeCLI() bool {
	return findClaude() != ""
}

// EnvCheck reports what CCTokenViz can find on this machine, for the setup wizard: Claude Code session
// logs (what the context meters read), the claude command (automatic limit checks) and status line data
func (a *App) EnvCheck() map[string]bool {
	logs := false
	if dirs, err := os.ReadDir(a.monitor.claudeDir); err == nil {
		for _, d := range dirs {
			if d.IsDir() {
				logs = true
				break
			}
		}
	}
	status := false
	if dir := statusDir(); dir != "" {
		files, _ := filepath.Glob(filepath.Join(dir, "status-*.json"))
		status = len(files) > 0
	}
	return map[string]bool{"logs": logs, "cli": findClaude() != "", "statusline": status}
}

// KnownProjects lists project names seen in recent sessions, for the project picker
func (a *App) KnownProjects() []string {
	return a.monitor.KnownProjects()
}

// DetectedContext returns the context window size each project's sessions reported, keyed by lowercase project name
func (a *App) DetectedContext() map[string]int {
	return detectedContextSizes()
}

// AddProjectFolder opens a folder picker and remembers the choice. The folder can be a
// single project or a folder that holds projects. It returns the chosen path, or "" if cancelled.
func (a *App) AddProjectFolder() (string, error) {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Choose a project folder, or a folder that holds your projects",
	})
	if err != nil || dir == "" {
		return "", err
	}
	cfg := loadConfig()
	for _, f := range cfg.ProjectFolders {
		if strings.EqualFold(f, dir) {
			return dir, nil
		}
	}
	cfg.ProjectFolders = append(cfg.ProjectFolders, dir)
	return dir, saveConfig(cfg)
}

// DismissSession hides a session until it sees new activity
func (a *App) DismissSession(id string) {
	dismissSession(id)
}

// limitsPoller asks for fresh plan limits every few seconds; MaybePollLimits decides whether
// a call is actually due (opt-in, a session active, and not too soon after the last one)
func (a *App) limitsPoller() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			go a.monitor.MaybePollLimits()
		}
	}
}

// GetBreakdown returns what has filled one session's context: its starting size, then what each prompt added
func (a *App) GetBreakdown(id string) Breakdown {
	return a.monitor.GetBreakdown(id)
}
