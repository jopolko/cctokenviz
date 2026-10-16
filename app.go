package main

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	monitor      *Monitor
	alertTracker map[string]map[int]bool // sessionID -> threshold -> alerted
	mu           sync.Mutex
	stopChan     chan struct{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		alertTracker: make(map[string]map[int]bool),
		stopChan:     make(chan struct{}),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.monitor = NewMonitor()

	// Start background monitoring for alerts
	go a.backgroundMonitor()
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

// DismissSession hides a session until it sees new activity
func (a *App) DismissSession(id string) {
	dismissSession(id)
}

// backgroundMonitor checks for threshold crossings and triggers alerts
func (a *App) backgroundMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			// Refreshes plan limits via a tiny headless call, only when stale and a session is active
			go a.monitor.MaybePollLimits()
			sessions := a.monitor.GetActiveSessions()
			for _, session := range sessions {
				a.checkThresholds(session)
			}
		}
	}
}

// alertThresholds returns the two configured thresholds, with defaults
func alertThresholds(cfg Config) (int, int) {
	warn, urgent := cfg.AlertWarn, cfg.AlertUrgent
	if warn <= 0 {
		warn = 75
	}
	if urgent <= 0 {
		urgent = 90
	}
	if urgent <= warn {
		urgent = warn + 1
	}
	return warn, urgent
}

// checkThresholds alerts when a session first crosses a threshold. A session that is already
// past a threshold when HeadRoom first sees it (at launch, or after re-arming) is recorded
// silently, so starting the app never fires a burst of old alerts.
func (a *App) checkThresholds(session SessionInfo) {
	cfg := loadConfig()
	warn, urgent := alertThresholds(cfg)

	a.mu.Lock()
	defer a.mu.Unlock()

	seen, known := a.alertTracker[session.ID]
	if !known {
		seen = make(map[int]bool)
		a.alertTracker[session.ID] = seen
	}

	for _, threshold := range []int{warn, urgent} {
		if session.Percentage < threshold || seen[threshold] {
			continue
		}
		seen[threshold] = true
		if known && !cfg.AlertsOff {
			go a.triggerAlert(session, threshold, cfg)
		}
	}
}

// triggerAlert sends OS-specific notifications and sounds
func (a *App) triggerAlert(session SessionInfo, threshold int, cfg Config) {
	title := "HeadRoom"
	message := session.ProjectName + " - " + strconv.Itoa(threshold) + "% context used"

	_, urgent := alertThresholds(cfg)
	if threshold >= urgent {
		message = session.ProjectName + ": context almost full (" + strconv.Itoa(session.Percentage) + "%)"
	} else {
		message = session.ProjectName + ": save context soon (" + strconv.Itoa(session.Percentage) + "%)"
	}

	switch runtime.GOOS {
	case "darwin":
		a.macAlert(title, message, threshold)
	case "windows":
		a.windowsAlert(title, message, !cfg.SoundOff)
	case "linux":
		a.linuxAlert(title, message, threshold)
	}
}

// macAlert triggers macOS notification and sound
func (a *App) macAlert(title, message string, threshold int) {
	// Play sound
	soundFile := "/System/Library/Sounds/Glass.aiff"
	if threshold >= 90 {
		soundFile = "/System/Library/Sounds/Sosumi.aiff"
	}
	exec.Command("afplay", soundFile).Start()

	// Show notification
	script := `display notification "` + message + `" with title "` + title + `" sound name "Glass"`
	exec.Command("osascript", "-e", script).Run()

	// Voice alert at 90%
	if threshold >= 90 {
		exec.Command("say", "Claude context almost full").Start()
	}
}

// windowsAlert shows a toast and optionally plays a sound. Everything runs in hidden PowerShell
// processes. The toast is sent under Windows PowerShell's own app id: Windows silently drops
// toasts from an app id it doesn't know, and registering one needs an installer.
func (a *App) windowsAlert(title, message string, sound bool) {
	esc := func(x string) string {
		return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;").Replace(x)
	}
	audio := `<audio silent="true"/>`
	if sound {
		audio = `<audio src="ms-winsoundevent:Notification.Default"/>`
	}
	ps := `
	[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
	[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
	$template = @"
	<toast>
		<visual>
			<binding template="ToastText02">
				<text id="1">` + esc(title) + `</text>
				<text id="2">` + esc(message) + `</text>
			</binding>
		</visual>
		` + audio + `
	</toast>
"@
	$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
	$xml.LoadXml($template)
	$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
	[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe").Show($toast)
	`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", ps)
	hideWindow(cmd)
	cmd.Run()
}

// linuxAlert triggers Linux notification and sound
func (a *App) linuxAlert(title, message string, threshold int) {
	// Try notify-send (most common)
	exec.Command("notify-send", title, message).Run()

	// Try to play a sound
	sounds := []string{
		"/usr/share/sounds/freedesktop/stereo/complete.oga",
		"/usr/share/sounds/gnome/default/alerts/glass.ogg",
		"/usr/share/sounds/ubuntu/stereo/message.ogg",
	}

	for _, sound := range sounds {
		if err := exec.Command("paplay", sound).Start(); err == nil {
			break
		}
	}

	// Fallback to terminal bell
	exec.Command("echo", "-e", "\a").Run()
}

// ResetAlerts re-arms alerts. Sessions already past a threshold stay quiet; each one alerts
// again only when it next crosses a threshold it hasn't been recorded at.
func (a *App) ResetAlerts() {
	a.mu.Lock()
	a.alertTracker = make(map[string]map[int]bool)
	a.mu.Unlock()
	for _, session := range a.monitor.GetActiveSessions() {
		a.checkThresholds(session) // first sight: record silently
	}
}

// TestAlert sends one alert so you can check notifications and sound work
func (a *App) TestAlert() {
	cfg := loadConfig()
	cfg.AlertsOff = false
	_, urgent := alertThresholds(cfg)
	go a.triggerAlert(SessionInfo{ProjectName: "Test alert", Percentage: urgent}, urgent, cfg)
}
