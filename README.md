# CCTokenViz

A lightweight desktop app that shows how much room you have left in Claude Code: context window usage across all your active sessions, plus your 5-hour and weekly plan limits.

**It replaces the status line for the desktop app.** In the Claude Code terminal, a `/statusline` can show context remaining, 5-hour usage and 7-day usage. The Claude desktop app has no status line, so those numbers aren't available there. CCTokenViz puts them in a small always-on-top window, for the desktop app and the terminal alike.

**[Download for Windows](https://github.com/jopolko/cctokenviz/releases/latest)** (a single portable `.exe`, no installer needed)

Forked from [gosparq/claudewatch](https://github.com/gosparq/claudewatch) (MIT). Not affiliated with Anthropic.

<p align="center">
  <img src="docs/screenshot.png" alt="CCTokenViz in the Terminal theme: context per project, 5-hour and weekly limits, and the Check API cost button at the bottom" width="224">
</p>

## Why?

When working with AI coding assistants on complex tasks, context management becomes critical. Claude Code has a finite context window (200k tokens, or 1M on extended-context models), and when it fills up, Claude automatically compacts the conversation—often losing important details you wanted to preserve.

**CCTokenViz** gives you visibility into your context usage across all active sessions, so you can intentionally save your work before Claude does it for you.

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

- 🖥️ **Floating window** - Always on top, stays visible while you work; its height fits what it shows
- 📊 **Real-time monitoring** - Updates every 2 seconds, from local files only
- 📁 **One row per project** - Monitors all active Claude Code sessions automatically. Two sessions on the same project fold into one row (the one you used most recently) with a small **+N** tag; a setting shows every session separately
- 🏷️ **Project names** - Labels each row with the project the session is working in, with the session title underneath
- 💤 **Idle sessions** - A session with no real conversation for a while stays listed, dimmed, with its last reading
- ✕ **Hide a session** - Hide one row until that session has new activity; or **hide a project for good** from Settings
- 📈 **Limits** - 5-hour and weekly usage with reset countdowns, for Pro/Max plans, with a **Refresh** button. A window that just reset shows as "reset" until a fresh reading arrives (see [Limits](#limits-where-the-numbers-come-from))
- 💲 **API cost** - A "Check API cost" button at the bottom, on demand: tokens, estimated cost at API prices and what prompt caching saved, for today and the last 7 days. Free (it reads your local logs) and information only (see [API cost](#api-cost-on-demand))
- 📐 **Context windows** - Detected per project (200k or 1M), or set your own size per project or as a default
- ↕️ **Your order** - Drag the ⋮⋮ grip on a row to reorder; the order stays put
- 🎨 **Themes** - Ten full styles (Linux terminal, sci-fi, Bitcoin, AI, cowboy, anime, kawaii, garden, celestial, lab notebook) each with its own meter, fonts, headings and wording, three plain themes, an accent color picker, and a font and size selector
- 🧭 **Setup wizard** - A first-run walkthrough, and a gear icon for changing anything afterwards
- 🔁 **Upgrades from HeadRoom** - This app was called HeadRoom; your settings and state move over automatically (see [Upgrading from HeadRoom](#upgrading-from-headroom))

Works with the Claude Code terminal and the Claude desktop app (Code tab); both write the same session logs.

## Installation

> **Note**: The original app was only tested on macOS. The limits, labels, dismiss, reordering and settings features were developed and tested on Windows 11 against the Claude desktop app and the terminal CLI; they are untested on macOS and Linux. Linux requires building from source on a Linux machine—contributions welcome!

### Pre-built Binaries

Download the latest Windows build from the [Releases page](https://github.com/jopolko/cctokenviz/releases/latest): unzip it and run `CCTokenViz.exe`. Binaries are published as release downloads, not stored in the repository.

### Build from Source

**Prerequisites:**
- [Go 1.21+](https://go.dev/dl/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **macOS**: Xcode command line tools (`xcode-select --install`)
- **Windows**: WebView2 (usually pre-installed on Windows 10/11)
- **Linux**: `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev`

```bash
# Clone or download this project
cd cctokenviz

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

1. **Launch the app** - Double-click `CCTokenViz.app` (macOS) or `CCTokenViz.exe` (Windows). The first launch opens a short setup wizard; CCTokenViz is a single executable and needs nothing installed beyond what Windows already has (the WebView2 runtime that Windows 10/11 ship). On Windows, add a Start Menu shortcut and pin it to the taskbar for one-click launch.
2. **Window auto-positions** - Opens in the top-right corner, always on top (drag to reposition).
3. **Work normally** - Open projects in Claude Code, in the terminal or the Claude desktop app. Each active project shows up as a row; there is nothing to register.
4. **Watch the colors** - Each bar changes color as it fills (green, yellow, orange, red, and pulsing red near the top), so you can see at a glance when a session is filling up and save your work before Claude compacts it.

### How projects show up

CCTokenViz doesn't need to be pointed at your project folders. Claude Code (terminal and desktop app) records every session in `~/.claude/projects/`, and CCTokenViz lists every session it has seen active. Open a project in Claude and its row appears within a couple of seconds.

**What counts as activity.** A session is active when its conversation moves: a message from you or from Claude. Claude Code also writes bookkeeping to session logs (titles, hook summaries, background-task notices), sometimes for sessions nobody is using; those writes are ignored, so they never make a session look live or bring back one you hid.

**Idle sessions keep their last reading.** After `activeMinutes` (default 60) without activity, a session's row stays, dimmed and marked `idle 2h 10m`, showing the last known context. Nothing is polled for an idle session (its log isn't changing, so there is nothing to read), and the row goes back to normal the moment you resume that session. It stays until you remove it with the ✕, even across restarts.

**Row names.** Each row is labelled with the project the session is working on, with the session's title underneath. The project is worked out, in order, from folders you added in Settings, the first folder under a directory named `projects` (`C:\Users\me\projects\my-app\src` gives `my-app`), or the nearest enclosing folder with a `.git`, `go.mod`, `package.json` etc. It follows the session if it moves between projects, using the most common project among the last several files and folders it touched, so a session that glances at another project doesn't flip its row. If nothing matches, the row uses the working folder's own name.

**One row per project.** If several sessions are working on the same project (for example a resumed session beside the original), CCTokenViz shows the one whose conversation moved most recently and adds a small **+N** after the name; hover it to see the others' context. To see every session as its own row, tick "Show every session as its own row" in Settings (`perSession` in the config).

**Showing only some projects.** By default every active project is listed. To limit it, open Settings (the gear at the top right) and choose "Only these projects", or set `projects` in `~/.cctokenviz.json`. Names are matched against the row names above, ignoring case. To remove one project for good, tick **hide** next to it in Settings (`hiddenProjects`).

**Order.** Rows stay where you put them. Drag the ⋮⋮ grip on a row to move it up or down; the order is remembered. A new session is added at the bottom.

**Hiding a session.** Click the **✕** on a row to hide that session. It stays hidden until it has new real activity (a message from you or Claude), then comes back on its own, so reopening a project in Claude and working in it brings its row back. If a session keeps returning because it is genuinely active, use **hide** on its project in Settings instead.

### What it monitors

The app scans `~/.claude/projects/` for active sessions and takes the last assistant message's input and cache tokens as the current context size, shown as a percentage of the context window. It reads local files only.

**Why it may differ from Claude's own meter.** The count is the same one Claude Code's status line uses: input plus cache tokens of the latest response, not counting that response's output. CCTokenViz reads the log every 2 seconds, and in a long turn every tool call writes a new entry, so it can run ahead of the desktop app's context display, which appears to update less often. In one check the app was about four minutes behind, and the two also differed by about 1% for reasons I couldn't determine. Compare right after a turn finishes and they should agree within a percent or two.

## Settings

Open Settings from the gear at the top right any time. **Theme** has ten styles and three plain themes. A style is a whole personality, not a palette: its own meter, fonts, background pattern, heading decoration, logo and wording, and the picker shows each one as a live miniature.

| Style | Meter | Flavor |
|-------|-------|--------|
| **Terminal** (Linux) | Segmented LED blocks | `$ context`, lowercase `used`/`free`, scanlines, monospace |
| **Starship** (sci-fi) | Glowing gauge with tick marks | `[ CONTEXT ]`, cut corners, grid |
| **Satoshi** (Bitcoin) | Minted coins | `₿` headings, `Spent`/`Left`, orange on black |
| **Neural** (AI) | Chain of nodes on a wire | `context_window`, `rate_limits`, monospace |
| **Frontier** (cowboy) | A belt of cartridges | `★ CONTEXT ★`, slab serif, wood grain |
| **Shonen** (anime) | Slanted power gauge | `CONTEXT!`, impact lettering, speed lines, `Power used`/`Power left` |
| **Sweet** (kawaii) | Hearts | `♡ Context ♡`, handwriting headings, pastel pink, hearts and sparkles |
| **Bloom** (garden and eco) | Leaves | `❦ Context ❦`, italic serif, sage and cream |
| **Celestial** | Stars | `☾ Context ☽`, starfield, gold on indigo |
| **Lab Notebook** (scientific) | Ruled gauge | `Fig. 1 — Context`, graph paper, serif |

The plain themes are Midnight, Daylight and High contrast. They are original designs and don't use any brand's characters or logos.

Every option in Settings:

| Section | Option | What it does |
|---------|--------|--------------|
| **Theme** | Style / plain theme | The look, shown as live miniatures |
| | Accent color, "Use theme's" | Override the highlight color, or go back to the theme's |
| | Font, Text size | Fonts that ship with Windows (nothing is bundled); 11-17 px |
| **Projects** | Show every active session / Only these projects | Show all, or pick projects; Select all / Deselect all |
| | Show every session as its own row | Turn off the one-row-per-project folding |
| | hide (per project) | Never show that project, however active |
| | Context box (per project) | Override that project's context window: leave empty to detect, or type 200k, 1M, 750k |
| | Add a folder… / ✕ | Point at a single project or a folder of projects (e.g. outside `~/projects`) |
| **Context window** | Default | Window size used when a project's can't be detected |
| **Limits** | Check my 5-hour and weekly limits automatically, Frequency | Opt in to background limit checks (about 1K tokens each; see [Limits](#limits-where-the-numbers-come-from)) |
| **Window** | Keep my window size | Stop the window fitting its height to its content |
| (bottom) | Run setup again | Replays the wizard |

Changes save as you make them. There is no Settings option for API cost: it is a button in the main window.

Settings are stored in `~/.cctokenviz.json`, which you never need to edit by hand but can. It is re-read on every refresh, so edits apply immediately:

```json
{
  "projects": ["my-app", "my-api"],
  "activeMinutes": 30,
  "contextTokens": 1000000,
  "projectFolders": ["D:\\work"],
  "projectContext": { "my-app": 1000000 },
  "perSession": false,
  "hiddenProjects": ["scratch"],
  "pollLimits": false,
  "pollMinutes": 10,
  "theme": "midnight",
  "accent": "",
  "font": "system",
  "fontSize": 13,
  "setupDone": true,
  "fixedSize": false
}
```

| Key | Default | Meaning |
|-----|---------|---------|
| `projects` | all | Only show sessions in these project folders (the first folder under a `projects` directory, taken from the session's working directory and the files and commands Claude is working on, latest wins) |
| `perSession` | false | Show every session as its own row instead of one row per project (the most recently active, with a +N tag) |
| `hiddenProjects` | none | Projects that are never shown, even with new activity. Tick "hide" next to a project in Settings. (The ✕ on a row only hides that session until it has new activity.) |
| `activeMinutes` | 60 | How long a session can go without activity before its row is marked idle (it stays listed) |
| `contextTokens` | 200000 | Fallback context window, for projects with no detected size and no override (use 1000000 for 1M-context models) |
| `projectFolders` | none | Folders you added in Settings: a single project, or a folder that holds several. Projects are also found from your Claude Code sessions, from any `projects` directory they use, and by the nearest enclosing folder with a `.git`, `go.mod`, `package.json` etc. |
| `projectContext` | detect | Per-project window override in tokens, e.g. `{"my-app": 750000}` (Settings accepts 200k, 1M, 750k, any size). Without one, a project's size comes from what its sessions report; usage beyond 200k bumps it to 1M |
| `pollLimits` | false | Opt in to background limit refreshes, see [headless polling](#limits-option-b-headless-polling-opt-in) |
| `pollMinutes` | 10 | Minutes between automatic polls (minimum 5; halved above 85% usage) |
| `theme`, `accent`, `font`, `fontSize` | midnight, theme's, theme's, 13 | Appearance: a theme id: `linux`, `scifi`, `bitcoin`, `ai`, `cowboy`, `anime`, `sweet`, `bloom`, `celestial`, `lab`, `midnight`, `light` or `contrast`; `accent` is an optional `#rrggbb`; `font` empty or `theme` uses the theme's own |
| `setupDone` | false | Set once the wizard has been completed |
| `fixedSize` | false | By default the window fits its height to what it is showing; set true to keep your own size |

## API cost (on demand)

At the bottom of the window there is a **Check API cost** button. Nothing is read until you click it, and the result stays until you close it (✕) or check again (↻). **It is information only, not a charge, and it spends no tokens:** it just adds up the token usage already recorded in your Claude Code session logs on this computer (including sub-agents) for **today** and the **last 7 days**, so you can compare your usage with what the same work would cost on the pay-as-you-go or enterprise API. It shows:

- total tokens, and the **estimated cost** at public API prices (input, output, cache reads and cache writes priced separately),
- how much **prompt caching saved**: cache-read tokens priced at the normal input rate, minus what they actually cost.

It only reads files already on your computer. On a Pro or Max plan the dollar figure is what the same usage would cost on the API, not what you pay; it is an estimate, and a model it doesn't recognise is priced as a mid-tier model (marked with ~).

To regenerate the screenshots above from the real frontend with demo data, run `python scripts/screenshots.py` (needs Edge or Chrome).

## Limits: where the numbers come from

CCTokenViz can use two sources for the limits. You only need one. Context bars need none of this; they read local session logs and cost nothing.

| Source | Shows | Authentication | Cost |
|--------|-------|----------------|------|
| **Status line export** | 5-hour and weekly plan limits, plus each session's real context window size | None. Claude Code runs your status line script and hands it the data. | Free. No extra requests. |
| **Headless polling** (opt-in) | 5-hour and weekly plan limits | Your own `claude` login, used by the official CLI on your machine. CCTokenViz never reads or stores it. | A small real request per poll: it slowly draws on your plan, or bills your API key. See [What polling costs](#what-polling-costs). |

### Limits, option A: status line export

Claude Code only reports your 5-hour and weekly plan usage (Pro/Max plans) to its **status line command**. Have yours save the JSON it receives:

- Copy `scripts/cctokenviz-statusline.py` somewhere and set it as your status line in `~/.claude/settings.json`:
  ```json
  "statusLine": { "type": "command", "command": "python /path/to/cctokenviz-statusline.py" }
  ```
- Or, if you already have a status line script, copy the `export_for_cctokenviz` function from that file into it and call it with the parsed JSON.

The script writes `~/.cctokenviz/status-<session_id>.json`. (Before the rename to CCTokenViz it wrote to `~/.headroom/`; if you copied the old script, change that path.)

**Limitations:** the Claude desktop app doesn't run the status line, so only terminal sessions export. And the limit numbers inside the file only change after a response, so they can lag behind the real values. "Last reported X ago" shows how old the reading is.

### Limits, option B: headless polling (opt-in)

Turn on "Check my 5-hour and weekly limits automatically" in the setup wizard or settings (`"pollLimits": true`). While one of your sessions has been active in the last 30 minutes, CCTokenViz runs one headless Claude Code call:

```
claude -p "reply with: ok" --model haiku --output-format stream-json --verbose --no-session-persistence \
       --system-prompt "Reply with the single word ok." --tools "" --disable-slash-commands \
       --strict-mcp-config --setting-sources ""
```

Claude Code emits a `rate_limit_event` with both windows, which CCTokenViz reads.

#### What polling costs

Polling is a slow drip, not a one-off, but a small one. A plain headless call would send Claude Code's whole system prompt and tool list, about 39,000 tokens, just to say "ok". CCTokenViz's poll replaces the system prompt with one line and turns off tools, skills, MCP servers and settings files, which brings a poll down to **under 1,000 tokens**: I measured 887, against 39,214 for the plain call, and it still returns both the 5-hour and weekly windows. (It does not use `--bare`, which would stop the CLI from reading your subscription login. If your `claude` is too old to know these options, CCTokenViz falls back to the plain call.) At the default of one poll every 10 minutes while you're working, that's about 6 polls and roughly 5,000 tokens an hour. CCTokenViz never charges anything itself; what you pay depends on what the `claude` command is signed in to:

| `claude` is signed in with | What a poll costs |
|----------------------------|-------------------|
| **A Pro or Max subscription** | No money. Each poll takes a tiny slice of your 5-hour and weekly limits, so CCTokenViz's own checking uses a little of the quota it is showing you. I couldn't measure the share, because the usage figure only moves in whole percents; at under 1,000 tokens a poll it should be hard to notice, but it is not zero. |
| **An Anthropic API key** (an `ANTHROPIC_API_KEY` set, or a Console login) | Real money, but very little: roughly $0.0002 per poll at API list prices, about a tenth of a cent an hour at the default. |
| **A subscription with extra usage switched on** | Free until the 5-hour limit is used up. At 100%, a poll goes through and is billed to extra usage. CCTokenViz stops polling once it sees a limit at 100% and starts again after the reset, but it can only react after it has seen the 100%. |

CCTokenViz can't tell which of these applies to you. To keep the drip small, it only polls while one of your sessions has been active in the last 30 minutes, and you can lower the frequency or turn polling off in Settings. The status line export (option A, above) costs nothing, at the price of only updating from terminal sessions.

- It uses the official `claude` CLI under your own login, on your machine. CCTokenViz never reads or stores your credentials.
- It never runs while you're idle, so it won't start a new 5-hour window by itself.
- It leaves no session log (`--no-session-persistence`) and runs from a temp folder, so it doesn't appear as a session.
- When several sources report, CCTokenViz shows the latest window and the highest percentage, since a status line file can hold numbers from before its last refresh.

It's off by default because it spends a little of your plan. Review [Anthropic's terms](https://www.anthropic.com/legal) and decide for yourself; this project is not affiliated with Anthropic and can't speak for their policies. CCTokenViz deliberately does **not** call undocumented usage endpoints with your OAuth token.

**How often should it refresh?** Context changes by the second and is free, so it updates every 2 seconds. Plan limits change slowly: a 5-hour window moves roughly 1% every 3 minutes at full burn, so frequent polling buys little.

- **Default: every 10 minutes while you're active** (about 6 polls an hour, roughly 5,000 tokens, about a tenth of a cent an hour at API list prices). That's fresh enough to plan a compact or a break.
- **Faster when it matters:** above 85% on any limit it polls twice as often, down to a floor of 5 minutes, since that's when a fresher number changes what you do.
- **Exact when you need it:** the **Refresh** button checks right now (it needs automatic checks turned on in Settings). Use it before a big decision rather than polling harder all day.
- **After a reset:** when a window's reset time passes it shows as **reset** (weekly also shows the next reset; the 5-hour window "starts with your next message") instead of disappearing, and a check follows within about 2 minutes while you are active. A window that reset more than a day ago, with no reading since, is dropped.
- **Tune it:** `pollMinutes` in the config (default 10, minimum 5). Set `pollLimits` to false to turn polling off entirely.

## Upgrading from HeadRoom

This app was called HeadRoom. On first launch it moves `~/.headroom.json` to `~/.cctokenviz.json` and `~/.headroom/` to `~/.cctokenviz/` (merging into the new folder if it already exists), so your settings, row order and hidden sessions carry over. If you use the status line export, change the path in your script from `.headroom` to `.cctokenviz`, or copy the new `scripts/cctokenviz-statusline.py`.

## Project Structure

```
cctokenviz/
├── main.go           # Wails entry point
├── app.go            # App bindings and the limits poller
├── monitor.go        # Session log parsing, activity, one row per project
├── projects.go       # Which project a path belongs to
├── config.go         # ~/.cctokenviz.json
├── limits.go         # Plan limits from status line exports, reset handling
├── poll.go           # Opt-in headless limit refresh
├── usage.go          # API cost: tokens and estimated price from local logs
├── order.go          # Remembers the order rows are dragged into
├── dismiss.go        # Per-session hide
├── tracked.go        # Remembers sessions seen active, so idle ones stay listed
├── migrate.go        # Moves ~/.headroom* to ~/.cctokenviz*
├── scripts/
│   ├── cctokenviz-statusline.py  # Status line script that feeds CCTokenViz
│   ├── relaunch.ps1              # Stop, rebuild and restart (development)
│   └── screenshots.py            # Regenerates docs/ screenshots from demo data
├── docs/             # Screenshot and its demo-data stub
├── frontend/
│   ├── index.html    # UI structure
│   ├── style.css     # Styling
│   ├── themes.js     # The ten styles and three plain themes
│   ├── main.js       # Rows, limits and API cost
│   └── settings.js   # Settings panel and setup wizard
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

On Windows, `powershell -File scripts/relaunch.ps1` stops the running copy (a running exe blocks the build), rebuilds `build\bin\CCTokenViz.exe` and starts it, so a pinned taskbar icon always opens the latest build. `go test ./...` runs the tests.

## Troubleshooting

### "No active Claude sessions"
- A session only appears once CCTokenViz has seen it active (a change in its log while CCTokenViz is running, or within `activeMinutes` of launching it)
- If you set `projects` in `~/.cctokenviz.json`, only those project names are shown; check `hiddenProjects` too
- A session you hid with ✕ stays hidden until it has new activity (reset by deleting `~/.cctokenviz/dismissed.json`)
- Check that Claude Code is storing sessions in `~/.claude/projects/`

### A project keeps coming back after I hide it
The ✕ hides one session until it has real new activity. If the session is still open and working, it returns. Tick **hide** next to the project in Settings to remove it for good.

### Two rows for one project
By default sessions on the same project fold into one row with a +N tag. If you see two rows, either "Show every session as its own row" is on, or the sessions were judged to be on different projects (check the names).

### Limits show "reset" or are missing
"Reset" means that plan window just reset and no fresh reading has arrived yet. Turn on automatic checks in Settings, press Refresh, or send a message in a terminal session with the status line export. See [Limits](#limits-where-the-numbers-come-from).

### Build errors
- Run `wails doctor` to diagnose missing dependencies
- Ensure Go and Wails are in your PATH
- "Access denied" when building on Windows: close the running `CCTokenViz.exe` first (or use `scripts/relaunch.ps1`)

## License

MIT License - Feel free to modify and distribute.

## Contributing

Pull requests welcome! Ideas for future features:
- [ ] System tray mode
- [ ] Export context summary
- [ ] MCP server integration
