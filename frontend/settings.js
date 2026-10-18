// Setup wizard (first run) and settings panel. Both edit the same settings object and
// save it to ~/.cctokenviz.json as you go, so there is no "apply" button to forget.

let cfg = {};
let knownProjects = [];
let detectedCtx = {};
let hasCli = false;
let wizardStep = 0;
let wizardMode = false;
let saveTimer = null;

const overlay = () => document.getElementById('overlay');

function persist() {
    clearTimeout(saveTimer);
    saveTimer = setTimeout(async () => {
        try {
            await window.go.main.App.SaveSettings(cfg);
        } catch (error) {
            console.error('Error saving settings:', error);
        }
    }, 200);
}

// The panels need more room than the compact monitor window, so grow it (and keep it on
// screen) while one is open, then put the window back exactly as it was.
async function growWindow() {
    try {
        // The window drops down from where it is, anchored at the gear, and stays on screen.
        // The app remembers the original size and position.
        await window.go.main.App.ExpandForPanel(380, 640);
    } catch (error) { /* leave the window as it is */ }
}

async function restoreWindow() {
    try {
        await window.go.main.App.RestoreFromPanel();
    } catch (error) { /* window stays as is */ }
}

async function loadState() {
    cfg = await window.go.main.App.GetSettings();
    cfg.projects = cfg.projects || [];
    try { knownProjects = (await window.go.main.App.KnownProjects()) || []; } catch (e) { knownProjects = []; }
    try { detectedCtx = (await window.go.main.App.DetectedContext()) || {}; } catch (e) { detectedCtx = {}; }
    cfg.projectFolders = cfg.projectFolders || [];
    try { hasCli = await window.go.main.App.HasClaudeCLI(); } catch (e) { hasCli = false; }
}

// Pick up folders added or removed, and projects found in them
async function refreshProjects() {
    const fresh = await window.go.main.App.GetSettings();
    cfg.projectFolders = fresh.projectFolders || [];
    try { knownProjects = (await window.go.main.App.KnownProjects()) || []; } catch (e) { /* keep the old list */ }
    rerender();
}

async function closeOverlay() {
    clearTimeout(saveTimer);
    window.go.main.App.SaveSettings(cfg).catch(() => {});
    overlay().hidden = true;
    overlay().innerHTML = '';
    overlay().style.top = '';
    themesOpen = false;
    wizardMode = false;
    window.go.main.App.SetSettingsOpen(false).catch(() => {});
    appCfg = { ...cfg };
    await restoreWindow(); // size back first, so the auto-fit that follows starts from it
    scheduleFit();
    lastSessionsHtml = null;
    updateSessions();
    updateLimits();
}

/* ---------- Sections (shared by the wizard and the settings panel) ---------- */

// The theme picker is long, so it starts folded to one line naming the current theme
let themesOpen = false;

function lookSection() {
    const current = THEMES[cfg.theme] || THEMES[DEFAULT_THEME];
    const head = `<button class="theme-toggle" data-action="themes-toggle" aria-expanded="${themesOpen}">
                <span>${current.icon ? current.icon + ' ' : ''}<b>${current.name}</b> <small>${cfg.theme ? current.tag : 'follows Windows'}</small></span>
                <span class="chev">${themesOpen ? 'Hide ▾' : 'Change ▸'}</span></button>`;
    if (!themesOpen) {
        return `
        <div class="section">
            <h2 class="section-title">Theme</h2>
            ${head}
        </div>`;
    }
    const card = ([id]) => themeCardHtml(id, (cfg.theme || DEFAULT_THEME) === id);
    const group = (name) => Object.entries(THEMES).filter(([, t]) => t.group === name).map(card).join('');
    const cards = THEME_GROUPS.map(([id, label]) => `<div class="group-label">${label}</div><div class="theme-grid">${group(id)}</div>`).join('');

    const fonts = FONTS.map(f => `<option value="${f.id}" ${((cfg.font || 'theme') === f.id) ? 'selected' : ''} style="font-family:${f.stack || 'inherit'}">${f.name}</option>`).join('');
    const accent = /^#[0-9a-fA-F]{6}$/.test(cfg.accent || '') ? cfg.accent : (THEMES[cfg.theme] || THEMES[DEFAULT_THEME]).vars.accent;

    return `
        <div class="section">
            <h2 class="section-title">Theme</h2>
            ${head}
            ${cards}
            <div class="row"><span>Accent</span>
                <input type="color" id="accent" value="${accent}">
                <button class="btn" data-action="accent-reset">Use theme's</button></div>
            <div class="row"><span>Font</span><select id="font">${fonts}</select></div>
            <div class="row"><span>Size</span>
                <input type="range" id="fontSize" min="11" max="17" step="1" value="${cfg.fontSize || DEFAULT_FONT_SIZE}">
                <b id="fontSizeLabel">${cfg.fontSize || DEFAULT_FONT_SIZE}px</b></div>
        </div>`;
}

function projectNames() {
    return [...new Set([...knownProjects, ...cfg.projects])].sort((a, b) => a.toLowerCase().localeCompare(b.toLowerCase()));
}

// 1000000 -> "1M", 200000 -> "200k", 750000 -> "750k"
function formatTokens(n) {
    if (!n) return '';
    if (n >= 1000000) return `${+(n / 1000000).toFixed(2)}M`;
    return `${Math.round(n / 1000)}k`;
}

// "1m", "200k", "750,000" -> tokens; "" or "auto" -> 0; anything else -> null (invalid)
function parseTokens(text) {
    const t = String(text).trim().toLowerCase().replace(/,/g, '');
    if (t === '' || t === 'auto') return 0;
    const m = t.match(/^(\d+(?:\.\d+)?)\s*([km])?$/);
    if (!m) return null;
    const n = parseFloat(m[1]) * (m[2] === 'm' ? 1000000 : m[2] === 'k' ? 1000 : 1);
    return n >= 1000 ? Math.round(n) : null;
}

const CTX_LIST = '<datalist id="ctxsizes"><option value="200k"><option value="500k"><option value="1M"></datalist>';

function projectsSection() {
    const all = !cfg.projects.length;
    const names = projectNames();
    const list = names.length
        ? names.map(n => {
            const set = (cfg.projectContext || {})[n] || 0;
            const found = detectedCtx[n.toLowerCase()] || 0;
            const hint = found ? `Auto · ${formatTokens(found)}` : 'Auto';
            const hidden = (cfg.hiddenProjects || []).some(h => h.toLowerCase() === n.toLowerCase());
            return `<div class="proj-row"><label><input type="checkbox" data-project="${escapeHtml(n)}" ${cfg.projects.includes(n) ? 'checked' : ''} ${all ? 'disabled' : ''}> ${escapeHtml(n)}</label>
                <label class="hide-box" title="Never show this project, even when it has new activity"><input type="checkbox" data-hide="${escapeHtml(n)}" ${hidden ? 'checked' : ''}> hide</label>
                <input class="ctx-input" list="ctxsizes" data-pctx="${escapeHtml(n)}" value="${formatTokens(set)}" placeholder="${hint}" title="Context window for this project. Leave empty to detect it; or type a size such as 200k, 1M or 750k"></div>`;
        }).join('')
        : '<div class="hint" style="margin-left:0">No projects found yet. Open one in Claude Code, or add a folder below.</div>';

    const folders = (cfg.projectFolders || []).map((f, i) =>
        `<div class="proj-row"><span class="folder-path" title="${escapeHtml(f)}">${escapeHtml(f)}</span><a href="#" data-action="folder-remove" data-i="${i}" title="Stop looking here">✕</a></div>`).join('');

    return `
        <div class="section">
            <h2 class="section-title">Projects</h2>
            <label><input type="radio" name="pscope" value="all" ${all ? 'checked' : ''}> Show every active session</label>
            <label><input type="radio" name="pscope" value="some" ${all ? '' : 'checked'}> Only these projects</label>
            ${names.length ? `<div class="hint" style="margin:6px 0 0 24px"><a href="#" data-action="projects-all">Select all</a> · <a href="#" data-action="projects-none">Deselect all</a></div>` : ''}
            <label style="margin-top:6px"><input type="checkbox" id="perSession" ${cfg.perSession ? 'checked' : ''}> Show every session as its own row</label>
            <div class="hint" style="margin:2px 0 0 24px">Off: one row per project, the session you used most recently, with a small +N for others on that project.</div>
            <div class="checklist">${list}</div>${CTX_LIST}
            <div class="hint">Projects are found from your Claude Code sessions and from any folder you add. The box on the right is that project's context window: left empty it is detected from the session (shown in grey); type a size like 200k, 1M or 750k to override. Tick “hide” to never show a project. The ✕ on a row only hides that session until it has new activity.</div>
            ${folders ? `<div class="checklist" style="max-height:90px">${folders}</div>` : ''}
            <div class="hint" style="margin:6px 0 0 24px"><a href="#" data-action="folder-add">Add a folder…</a> a single project, or a folder that holds several</div>
        </div>`;
}

function contextSection() {
    return `
        <div class="section">
            <h2 class="section-title">Context window</h2>
            <div class="row"><span>Default</span>
                <input class="ctx-input" list="ctxsizes" id="ctxDefault" value="${formatTokens(cfg.contextTokens || 200000)}" title="Type a size such as 200k, 1M or 750k"></div>
            <div class="hint">Used only when a project's size can't be detected and you haven't set one above. Type the window your model has: 200k, 1M, 750k.</div>
        </div>`;
}

function windowSection() {
    return `
        <div class="section">
            <h2 class="section-title">Window</h2>
            <label><input type="checkbox" id="fixedSize" ${cfg.fixedSize ? 'checked' : ''}> Keep my window size</label>
            <div class="hint">By default CCTokenViz fits its height to what it is showing, so there's no empty space. Tick this to size it yourself.</div>
        </div>`;
}

// The full explanation is for the setup wizard; Settings gets a one-line version
function limitsSection(brief) {
    const minutes = cfg.pollMinutes || 10;
    const opts = [5, 10, 15, 30].map(m => `<option value="${m}" ${minutes === m ? 'selected' : ''}>every ${m} minutes</option>`).join('');
    return `
        <div class="section">
            <h2 class="section-title">Limits</h2>
            <label><input type="checkbox" id="poll" ${cfg.pollLimits ? 'checked' : ''} ${hasCli ? '' : 'disabled'}> Check my 5-hour and weekly limits automatically</label>
            ${brief ? `<div class="hint">${hasCli ? '<span class="ok">Claude Code found.</span>' : '<span class="bad">Claude Code not found, so this is unavailable.</span>'} Each check uses about 1K tokens of your plan. Off by default.</div>` : `<div class="hint">${hasCli ? '<span class="ok">Claude Code found.</span>' : '<span class="bad">Claude Code (the claude command) was not found, so this is unavailable.</span>'}
                Each check runs one tiny request under your own Claude login: under 1,000 tokens, roughly $0.0002 at API prices. It slowly draws on your plan limits (or bills your API key, if the claude command uses one), and CCTokenViz never sees or stores your login.
                It only checks while a session is active, and stops once a limit is used up.</div>`}
            <div class="row"><span>Frequency</span><select id="pollMin" ${cfg.pollLimits ? '' : 'disabled'}>${opts}</select></div>
            ${brief ? '' : '<div class="hint">Off by default. Without it, limits update only when a terminal session reports them (see the README) and the ↻ Refresh button does nothing.</div>'}
        </div>`;
}

/* ---------- Settings panel ---------- */

// The gear opens Settings, and closes it again when it is already open
function toggleSettings() {
    if (!overlay().hidden && !wizardMode) closeOverlay();
    else openSettings();
}

async function openSettings() {
    await loadState();
    wizardMode = false;
    await growWindow();
    renderSettings();
    // Settings sits under the header so the title and gear stay visible, like a drop-down
    const header = document.querySelector('header');
    overlay().style.top = header ? header.getBoundingClientRect().bottom + 'px' : '';
    overlay().hidden = false;
    window.go.main.App.SetSettingsOpen(true).catch(() => {}); // so the window's X closes this panel, not the app
}

function renderSettings() {
    const top = overlay().scrollTop;
    overlay().innerHTML = `
        <h2>Settings</h2>
        <p class="lead">Changes save as you make them.</p>
        ${lookSection()}${projectsSection()}${contextSection()}${limitsSection(true)}${windowSection()}
        <div class="nav">
            <button class="btn" data-action="rerun">Run setup again</button>
            <span class="spacer"></span>
            <button class="btn primary" data-action="close">Save</button>
        </div>`;
    overlay().scrollTop = top;
}

/* ---------- Setup wizard ---------- */

const WIZARD_STEPS = [
    () => `
        <h2>Welcome to CCTokenViz</h2>
        <p class="lead">CCTokenViz shows how much room you have left in Claude Code: the context window of each project you have open, and your 5-hour and weekly limits.</p>
        <p class="lead">This takes about a minute. Everything stays on your computer, and you can change any of it later from the gear at the top right.</p>`,
    () => `<h2>Pick a look</h2><p class="lead">Open the list to browse the styles, or skip this and change it later from Settings. Until you pick one, it follows Windows: Midnight in dark mode, Daylight in light mode.</p>${lookSection()}`,
    () => `<h2>Which projects?</h2><p class="lead">CCTokenViz finds your sessions on its own. Limit it to the projects you care about, or show everything.</p>${projectsSection()}`,
    () => `<h2>Context size</h2><p class="lead">How big is your model's context window?</p>${contextSection()}`,
    () => `<h2>Keeping limits current</h2><p class="lead">Your 5-hour and weekly limits only come from Claude Code itself. Choose whether CCTokenViz may ask for them in the background.</p><p class="lead"><b>Recommended if you want live 5-hour and weekly numbers.</b> Each check costs under 1,000 tokens. It stays off until you tick the box.</p>${limitsSection()}`,
    () => `
        <h2>You're set</h2>
        <p class="lead">Your sessions will show up under <b>Context</b>, and your limits under <b>Limits</b>. Drag the ⋮⋮ grip on a row to reorder it, and click ✕ to hide a session.</p>
        <p class="lead">Open Settings any time from the gear at the top right.</p>`,
];

async function openSetup() {
    await loadState();
    wizardMode = true;
    wizardStep = 0;
    await growWindow();
    renderWizard();
    overlay().hidden = false;
}

function renderWizard() {
    const last = wizardStep === WIZARD_STEPS.length - 1;
    const dots = WIZARD_STEPS.map((_, i) => `<span class="${i <= wizardStep ? 'on' : ''}"></span>`).join('');
    const top = overlay().scrollTop;
    overlay().innerHTML = `
        <div class="dots">${dots}</div>
        ${WIZARD_STEPS[wizardStep]()}
        <div class="nav">
            ${wizardStep > 0 ? '<button class="btn" data-action="back">Back</button>' : ''}
            <span class="spacer"></span>
            ${last ? '<button class="btn primary" data-action="finish">Finish</button>' : '<button class="btn primary" data-action="next">Next</button>'}
        </div>`;
    overlay().scrollTop = top;
}

function rerender() {
    if (wizardMode) renderWizard(); else renderSettings();
}

/* ---------- Input handling (one set of listeners for both) ---------- */

function setupOverlayEvents() {
    const root = overlay();

    root.addEventListener('click', async (e) => {
        const card = e.target.closest('[data-theme]');
        if (card) {
            cfg.theme = card.dataset.theme;
            cfg.accent = ''; // a new theme brings its own accent
            applyAppearance(cfg);
            persist();
            rerender();
            return;
        }
        const action = e.target.closest('[data-action]');
        if (!action) return;
        switch (action.dataset.action) {
            case 'accent-reset': cfg.accent = ''; applyAppearance(cfg); persist(); rerender(); break;
            case 'folder-add': {
                e.preventDefault();
                try {
                    await window.go.main.App.SaveSettings(cfg);
                    const dir = await window.go.main.App.AddProjectFolder();
                    if (dir) await refreshProjects();
                } catch (err) { console.error(err); }
                break;
            }
            case 'folder-remove':
                e.preventDefault();
                cfg.projectFolders.splice(parseInt(action.dataset.i, 10), 1);
                await window.go.main.App.SaveSettings(cfg);
                await refreshProjects();
                break;
            case 'themes-toggle': e.preventDefault(); themesOpen = !themesOpen; rerender(); break;
            case 'projects-all': e.preventDefault(); cfg.projects = projectNames(); persist(); rerender(); break;
            case 'projects-none': e.preventDefault(); cfg.projects = []; persist(); rerender(); break;
            case 'next': wizardStep++; renderWizard(); break;
            case 'back': wizardStep--; renderWizard(); break;
            case 'finish': cfg.setupDone = true; persist(); closeOverlay(); break;
            case 'close': closeOverlay(); break;
            case 'rerun': wizardMode = true; wizardStep = 0; renderWizard(); break;
        }
    });

    root.addEventListener('input', (e) => {
        if (e.target.id === 'accent') {
            cfg.accent = e.target.value;
            applyAppearance(cfg);
            persist();
        } else if (e.target.id === 'fontSize') {
            cfg.fontSize = parseInt(e.target.value, 10);
            document.getElementById('fontSizeLabel').textContent = cfg.fontSize + 'px';
            applyAppearance(cfg);
            persist();
        }
    });

    root.addEventListener('change', (e) => {
        const t = e.target;
        if (t.id === 'font') {
            cfg.font = t.value;
            applyAppearance(cfg);
            persist();
        } else if (t.name === 'pscope') {
            if (t.value === 'all') cfg.projects = [];
            else if (!cfg.projects.length) cfg.projects = knownProjects.slice(0, 1);
            persist();
            rerender();
        } else if (t.dataset && t.dataset.hide !== undefined) {
            const name = t.dataset.hide;
            const rest = (cfg.hiddenProjects || []).filter(h => h.toLowerCase() !== name.toLowerCase());
            cfg.hiddenProjects = t.checked ? [...rest, name] : rest;
            persist();
        } else if (t.dataset && t.dataset.project !== undefined) {
            const name = t.dataset.project;
            cfg.projects = t.checked ? [...new Set([...cfg.projects, name])] : cfg.projects.filter(p => p !== name);
            persist();
            // An empty selection would silently mean "show everything", so reflect that
            if (!cfg.projects.length) rerender();
        } else if (t.dataset && t.dataset.pctx !== undefined) {
            cfg.projectContext = cfg.projectContext || {};
            const v = parseTokens(t.value);
            if (v === null) { rerender(); return; } // not a size: put the old value back
            if (v) cfg.projectContext[t.dataset.pctx] = v; else delete cfg.projectContext[t.dataset.pctx];
            t.value = formatTokens(v);
            persist();
        } else if (t.id === 'ctxDefault') {
            const v = parseTokens(t.value);
            if (v === null || v === 0) { rerender(); return; }
            cfg.contextTokens = v;
            t.value = formatTokens(v);
            persist();
        } else if (t.id === 'poll') {
            cfg.pollLimits = t.checked;
            if (t.checked && !cfg.pollMinutes) cfg.pollMinutes = 10;
            persist();
            rerender();
        } else if (t.id === 'perSession') {
            cfg.perSession = t.checked;
            persist();
        } else if (t.id === 'fixedSize') {
            cfg.fixedSize = t.checked;
            persist();
        } else if (t.id === 'pollMin') {
            cfg.pollMinutes = parseInt(t.value, 10);
            persist();
        }
    });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && !root.hidden && !wizardMode) closeOverlay();
    });
}

document.addEventListener('DOMContentLoaded', setupOverlayEvents);

// The window's X while Settings is open comes back here as an event instead of quitting
document.addEventListener('DOMContentLoaded', () => {
    try {
        window.runtime.EventsOn('settings:close', () => { if (!overlay().hidden && !wizardMode) closeOverlay(); });
    } catch (error) { /* not running inside the app */ }
});
