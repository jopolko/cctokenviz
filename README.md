# Headroom

A lightweight desktop app that shows how much room you have left in Claude Code: context window usage across all your active sessions, plus your 5-hour and weekly plan limits.

Forked from [gosparq/claudewatch](https://github.com/gosparq/claudewatch) (MIT). Not affiliated with Anthropic.

![Headroom Screenshot](screenshot.png)

## Why?

When working with AI coding assistants on complex tasks, context management becomes critical. Claude Code has a finite context window (200k tokens, or 1M on extended-context models), and when it fills up, Claude automatically compacts the conversation—often losing important details you wanted to preserve.

**Headroom** gives you visibility into your context usage across all active sessions, so you can intentionally save your work before Claude does it for you.

When you get the warning, tell Claude:

*Write everything we did so far to progress.md, ensure to note the approach we're taking, the steps we've done so far, and the current failure we're working on.*

### The RPI Framework

This tool pairs well with the **Research-Plan-Implement (RPI) Framework**—a structured methodology for working with Claude Code:

| Phase | Purpose |
|-------|---------|
| **Research** | Deeply understand the codebase before making changes |
| **Plan** | Create detailed, validated implementation plans |
| **Implement** | Execute plans methodically, phase by phase |

Without structure, AI-assisted development often becomes chaotic: context gets lost between sessions, no clear methodology for multi-file changes, and painful handoffs when stopping mid-task.

**Learn more:**
- [Watch the RPI methodology video](https://www.youtube.com/watch?v=rmvDxxNubIg)
- [RPI Framework for Claude Code](https://github.com/acampb/claude-rpi-framework)

## Features

- 🖥️ **Floating window** - Always on top, stays visible while you work
- 📊 **Real-time monitoring** - Updates every 2 seconds
- 🔔 **Smart alerts** - Notification + sound at 75% and 90% context usage
- 🌍 **Cross-platform** - Works on macOS, Windows, and Linux
- 🚀 **Lightweight** - ~10MB binary, minimal resource usage
- 📁 **Multi-project** - Monitors all active Claude Code sessions automatically
- 🏷️ **Project names** - Labels sessions by the project folder under `~/projects` they're working in, with the session title underneath
- ✕ **Dismiss** - Hide a session until it has new activity
- 📈 **Plan limits** - 5-hour and weekly usage with reset countdowns, for Pro/Max plans (see [Plan limits](#plan-limits))
- 💳 **Credits** - Optional Console credit estimate and claude.ai credit display (see [Credits](#credits))

Works with the Claude Code terminal and the Claude desktop app (Code tab); both write the same session logs.

## Installation

> **Note**: The original app was only tested on macOS. The limits, credits, labels and dismiss features were developed and tested on Windows 11 against the Claude desktop app and the terminal CLI; they are untested on macOS and Linux. Linux requires building from source on a Linux machine—contributions welcome!

### Pre-built Binaries

Download from the project's GitHub Releases page.

### Build from Source

**Prerequisites:**
- [Go 1.21+](https://go.dev/dl/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **macOS**: Xcode command line tools (`xcode-select --install`)
- **Windows**: WebView2 (usually pre-installed on Windows 10/11)
- **Linux**: `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev`

```bash
# Clone or download this project
cd headroom

# Development mode (hot reload)
make dev

# Build for your current platform
make build

# Build for all platforms
make build-all
```

### Install (macOS)

```bash
make install-mac
# App will be available in /Applications
```

## Usage

1. **Launch the app** - Double-click `headroom.app` (macOS) or `headroom.exe` (Windows)
2. **Window auto-positions** - Opens in the top-right corner (drag to reposition if needed)
3. **Work normally** - The app monitors all Claude Code sessions in the background
4. **Get alerted** - At 75% and 90% context, you'll receive:
   - System notification
   - Audio alert
   - Voice warning (macOS at 90%)

### What it monitors

The app scans `~/.claude/projects/` for active sessions (modified recently) and takes the last assistant message's input and cache tokens as the current context size, shown as a percentage of the context window. It reads local files only.

## Configuration

Settings live in `~/.headroom.json` and are re-read on every refresh, so edits apply without a restart. Every key is optional:

```json
{
  "projects": ["my-app", "my-api"],
  "activeMinutes": 30,
  "contextTokens": 1000000,
  "pollLimits": false,
  "credits": {
    "console": { "balance": 12.34, "asOf": "2026-01-31" },
    "claudeAi": "$20.00"
  }
}
```

| Key | Default | Meaning |
|-----|---------|---------|
| `projects` | all | Only show sessions in these project folders (the first folder under a `projects` directory in the session's working directory) |
| `activeMinutes` | 60 | How recently a session must have changed to be listed |
| `contextTokens` | 200000 | Context window to measure against, for sessions that haven't exported their own size (use 1000000 for 1M-context models) |
| `pollLimits` | false | Opt in to background limit refreshes, see below |
| `credits` | none | Balances to display, see [Credits](#credits) |

## Plan limits

Claude Code only reports your 5-hour and weekly plan usage (Pro/Max plans) to its **status line command**, so Headroom needs one of two sources. Use either or both; it shows whichever is fresher.

### Option A: status line export (no extra requests)

Terminal sessions run your status line script on every refresh. Have it save the JSON it receives:

- Copy `scripts/headroom-statusline.py` somewhere and set it as your status line in `~/.claude/settings.json`:
  ```json
  "statusLine": { "type": "command", "command": "python /path/to/headroom-statusline.py" }
  ```
- Or, if you already have a status line script, copy the `export_for_headroom` function from that file into it and call it with the parsed JSON.

The script writes `~/.headroom/status-<session_id>.json`. As a bonus, sessions that export also report their real context window size.

**Limitation:** the Claude desktop app doesn't run the status line, so its sessions never export. Limits stay as fresh as your most recent terminal session ("Last reported X ago" shows how fresh).

### Option B: headless polling (opt-in)

Set `"pollLimits": true`. When the limits are stale and one of your sessions has been active in the last 30 minutes, Headroom runs one tiny headless Claude Code call:

```
claude -p "reply with: ok" --model haiku --output-format stream-json --verbose --no-session-persistence
```

Claude Code emits a `rate_limit_event` with both windows, which Headroom reads. Things to know:

- It uses the official `claude` CLI under **your own login**, on your machine. Headroom never reads or stores your credentials.
- Each probe is a real request. It's very small, but it counts against your plan's usage.
- It never runs while you're idle, so it won't start a new 5-hour window by itself. It runs at most once every 10 minutes.
- It leaves no session log (`--no-session-persistence`) and runs from a temp folder, so it doesn't appear as a session.

It's off by default because it spends a little of your quota. Review [Anthropic's terms](https://www.anthropic.com/legal) and decide for yourself; this project is not affiliated with Anthropic and can't speak for their policies.

Headroom deliberately does **not** call undocumented usage endpoints with your OAuth token.

## Credits

Both are optional and shown in a "Credits" panel.

- **claude.ai credits** (`credits.claudeAi`): free text you update yourself, shown as typed. There is no API for this balance.
- **Console (API) credits** (`credits.console`): the Console doesn't expose a credit balance through its API, so this is an estimate. Enter your balance and the date you read it (`asOf`, `YYYY-MM-DD`). If you set the environment variable `ANTHROPIC_ADMIN_KEY`, Headroom subtracts spend since that date, using the documented [Usage and Cost Admin API](https://platform.claude.com/docs/en/manage-claude/usage-cost-api), and shows the result with a `≈`. Without a key it shows your snapshot. Update the snapshot whenever you add credit.
  - Admin API keys are for organization accounts. Create one in the Console and set it yourself; Headroom only reads it from the environment and never writes it anywhere.
  - The Admin API is unavailable for individual (non-organization) accounts.

## Project Structure

```
headroom/
├── main.go           # Wails entry point
├── app.go            # Core app + notifications
├── monitor.go        # JSONL parsing logic
├── config.go         # ~/.headroom.json
├── limits.go         # Plan limits from status line exports
├── poll.go           # Opt-in headless limit refresh
├── credits.go        # Console credit estimate (Admin API)
├── dismiss.go        # Per-session dismiss
├── scripts/          # Status line script that feeds Headroom
├── frontend/
│   ├── index.html    # UI structure
│   ├── style.css     # Styling
│   └── main.js       # Update logic
├── wails.json        # Wails config
├── go.mod            # Go dependencies
└── Makefile          # Build commands
```

## Development

```bash
# Run with hot reload
make dev

# The app will open and refresh when you edit files
```

## Troubleshooting

### "No active Claude sessions"
- Sessions must be modified within the last 60 minutes
- Check that Claude Code is storing sessions in `~/.claude/projects/`

### Notifications not working
- **macOS**: Grant notification permissions in System Preferences
- **Windows**: Check notification settings in Windows Settings
- **Linux**: Ensure `notify-send` is installed

### Build errors
- Run `wails doctor` to diagnose missing dependencies
- Ensure Go and Wails are in your PATH

## License

MIT License - Feel free to modify and distribute.

## Contributing

Pull requests welcome! Ideas for future features:
- [ ] System tray mode
- [ ] Configurable alert thresholds
- [ ] Export context summary
- [ ] MCP server integration
