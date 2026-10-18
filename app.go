package main

import (
	"context"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	monitor  *Monitor
	stopChan chan struct{}
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
