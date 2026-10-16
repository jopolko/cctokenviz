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

1. **Launch the app** - Double-click `headroom.app` (macOS) or `headroom.exe` (Windows). On Windows, add a Start Menu shortcut and pin it to the taskbar for one-click launch.
2. **Window auto-positions** - Opens in the top-right corner, always on top (drag to reposition).
3. **Work normally** - Open projects in Claude Code, in the terminal or the Claude desktop app. Each active session shows up as a row on its own; there is nothing to register.
4. **Get alerted** - At 75% and 90% context you'll receive:
   - System notification
   - Audio alert
   - Voice warning (macOS at 90%)

### How projects show up

Headroom doesn't need to be pointed at your project folders. Claude Code (terminal and desktop app) records every session in `~/.claude/projects/`, and Headroom lists any session that has had activity in the last `activeMinutes` (default 60). Open a project in Claude and its row appears within a couple of seconds; stop using it and the row drops off after the cutoff.

**Row names.** Each row is labelled with the project folder the session is working in, with the session's title underneath. The project is the first folder under a directory named `projects` in the session's working directory (`C:\Users\me\projects\my-app\src` gives `my-app`), and it follows the session if it moves between projects. If the working directory isn't under a `projects` folder, the row uses that folder's own name.

**Showing only some projects.** By default every active session is listed. To limit it, list the project names in `~/.headroom.json`:

```json
{ "projects": ["my-app", "my-api"] }
```

Names are matched against the row names above, ignoring case, and edits apply on the next refresh.

**Hiding a session.** Click the **✕** on a row to hide it. It stays hidden until that session has new activity, then comes back on its own, so reopening a project in Claude and working in it brings its row back. A brand-new session in the same project always appears as a new row. Hidden sessions don't trigger alerts.

Several rows with the same project name are separate sessions (each conversation is its own log); the title under the name tells them apart.

### What it monitors

The app scans `~/.claude/projects/` for active sessions and takes the last assistant message's input and cache tokens as the current context size, shown as a percentage of the context window. It reads local files only.

**Why it may differ from Claude's own meter.** The count is the same one Claude Code's status line uses: input plus cache tokens of the latest response, not counting that response's output. Headroom reads the log every 2 seconds, and in a long turn every tool call writes a new entry, so it can run ahead of the desktop app's context display, which appears to update less often. In one check the app was about four minutes behind, and the two also differed by about 1% for reasons I couldn't determine. Compare right after a turn finishes and they should agree within a percent or two.

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
| `pollLimits` | false | Opt in to background limit refreshes, see [headless polling](#plan-limits-option-b-headless-polling-opt-in) |
| `pollMinutes` | 10 | Minutes between automatic polls (minimum 5; halved above 85% usage) |
| `credits` | none | Balances to display, see [Credits](#credits) |

## Plan limits and credits: where the numbers come from

Headroom can use three different sources. They are not interchangeable: plan limits come from the first two, Console credits from the third. Context bars need none of this; they read local session logs and cost nothing.

| Source | Shows | Authentication | Cost |
|--------|-------|----------------|------|
| **Status line export** | 5-hour and weekly plan limits, plus each session's real context window size | None. Claude Code runs your status line script and hands it the data. | Free. No extra requests. |
| **Headless polling** (opt-in) | 5-hour and weekly plan limits | Your own `claude` login, used by the official CLI on your machine. Headroom never reads or stores it. | One small real request per poll, counted against your plan. See below. |
| **Admin API key** (optional) | Console (API) spend, used to estimate your API credit balance | An admin key you create and put in `ANTHROPIC_ADMIN_KEY` | Free. It only reads usage reports; it makes no model calls. |

### Plan limits, option A: status line export

Claude Code only reports your 5-hour and weekly plan usage (Pro/Max plans) to its **status line command**. Have yours save the JSON it receives:

- Copy `scripts/headroom-statusline.py` somewhere and set it as your status line in `~/.claude/settings.json`:
  ```json
  "statusLine": { "type": "command", "command": "python /path/to/headroom-statusline.py" }
  ```
- Or, if you already have a status line script, copy the `export_for_headroom` function from that file into it and call it with the parsed JSON.

The script writes `~/.headroom/status-<session_id>.json`.

**Limitations:** the Claude desktop app doesn't run the status line, so only terminal sessions export. And the limit numbers inside the file only change after a response, so they can lag behind the real values. "Last reported X ago" shows how old the reading is.

### Plan limits, option B: headless polling (opt-in)

Set `"pollLimits": true`. While one of your sessions has been active in the last 30 minutes, Headroom runs one headless Claude Code call:

```
claude -p "reply with: ok" --model haiku --output-format stream-json --verbose --no-session-persistence
```

Claude Code emits a `rate_limit_event` with both windows, which Headroom reads.

**What it costs.** Each probe is a real Haiku request. Claude Code sends its normal system prompt and tool definitions even for a one-word reply, so one probe is about 40,000 input tokens (mostly cache reads) and a handful of output tokens, roughly $0.002 at API list prices. On a subscription you pay no money, but every probe counts toward your plan usage. I could not measure its share of your limit because the usage figure only moves in whole percents; expect it to be small, but it is not zero.

- It uses the official `claude` CLI under your own login, on your machine. Headroom never reads or stores your credentials.
- It never runs while you're idle, so it won't start a new 5-hour window by itself.
- It leaves no session log (`--no-session-persistence`) and runs from a temp folder, so it doesn't appear as a session.
- When several sources report, Headroom shows the latest window and the highest percentage, since a status line file can hold numbers from before its last refresh.

It's off by default because it spends a little of your plan. Review [Anthropic's terms](https://www.anthropic.com/legal) and decide for yourself; this project is not affiliated with Anthropic and can't speak for their policies. Headroom deliberately does **not** call undocumented usage endpoints with your OAuth token.

**How often should it refresh?** Context changes by the second and is free, so it updates every 2 seconds. Plan limits change slowly: a 5-hour window moves roughly 1% every 3 minutes at full burn, so frequent polling buys little.

- **Default: every 10 minutes while you're active** (about 6 probes an hour, roughly 240k mostly cached tokens, about a cent an hour at API list prices). That's fresh enough to plan a compact or a break.
- **Faster when it matters:** above 85% on any limit it polls twice as often, down to a floor of 5 minutes, since that's when a fresher number changes what you do.
- **Exact when you need it:** the **Refresh** button checks right now. Use it before a big decision rather than polling harder all day.
- **Tune it:** `pollMinutes` in the config (default 10, minimum 5). Set `pollLimits` to false to turn polling off entirely.

### Credits

Both are optional and shown in a "Credits" panel.

- **claude.ai credits** (`credits.claudeAi`): free text you update yourself, shown as typed. There is no API for this balance.
- **Console (API) credits** (`credits.console`): the Console doesn't expose a credit balance through its API, so this is an estimate. Enter your balance and the date you read it (`asOf`, `YYYY-MM-DD`). If you set the environment variable `ANTHROPIC_ADMIN_KEY`, Headroom subtracts spend since that date, using the documented [Usage and Cost Admin API](https://platform.claude.com/docs/en/manage-claude/usage-cost-api), and shows the result with a `≈`. Without a key it shows your snapshot. Update the snapshot whenever you add credit.
  - Admin API keys are for organization accounts. Create one in the Console and set it yourself; Headroom only reads it from the environment and never writes it anywhere.
  - The Admin API is unavailable for individual (non-organization) accounts.
  - Reading usage reports is free; they make no model calls.

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
- Sessions must have had activity within `activeMinutes` (default 60)
- If you set `projects` in `~/.headroom.json`, only those project names are shown
- A session you hid with ✕ stays hidden until it has new activity (reset by deleting `~/.headroom/dismissed.json`)
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
