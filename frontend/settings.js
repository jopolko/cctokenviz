// Setup wizard (first run) and settings panel. Both edit the same settings object and
// save it to ~/.headroom.json as you go, so there is no "apply" button to forget.

let cfg = {};
let knownProjects = [];
let hasCli = false;
let wizardStep = 0;
let wizardMode = false;
let savedSize = null;
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
        const size = await window.runtime.WindowGetSize();
        const pos = await window.runtime.WindowGetPosition();
        savedSize = { size, pos };
        window.runtime.WindowSetSize(Math.max(size.w, 380), Math.max(size.h, 640));
        window.runtime.WindowCenter();
    } catch (error) {
        savedSize = null;
    }
}

function restoreWindow() {
    try {
        if (savedSize) {
            window.runtime.WindowSetSize(savedSize.size.w, savedSize.size.h);
            window.runtime.WindowSetPosition(savedSize.pos.x, savedSize.pos.y);
        }
    } catch (error) { /* window stays as is */ }
    savedSize = null;
}

async function loadState() {
    cfg = await window.go.main.App.GetSettings();
    cfg.projects = cfg.projects || [];
    try { knownProjects = (await window.go.main.App.KnownProjects()) || []; } catch (e) { knownProjects = []; }
    try { hasCli = await window.go.main.App.HasClaudeCLI(); } catch (e) { hasCli = false; }
}

function closeOverlay() {
    clearTimeout(saveTimer);
    window.go.main.App.SaveSettings(cfg).catch(() => {});
    overlay().hidden = true;
    overlay().innerHTML = '';
    wizardMode = false;
    appCfg = { ...cfg };
    restoreWindow();
    scheduleFit();
    lastSessionsHtml = null;
    updateSessions();
    updateLimits();
}

/* ---------- Sections (shared by the wizard and the settings panel) ---------- */

function lookSection() {
    const card = ([id]) => themeCardHtml(id, (cfg.theme || DEFAULT_THEME) === id);
    const group = (name) => Object.entries(THEMES).filter(([, t]) => t.group === name).map(card).join('');
    const cards = THEME_GROUPS.map(([id, label]) => `<div class="group-label">${label}</div><div class="theme-grid">${group(id)}</div>`).join('');

    const fonts = FONTS.map(f => `<option value="${f.id}" ${((cfg.font || 'theme') === f.id) ? 'selected' : ''} style="font-family:${f.stack || 'inherit'}">${f.name}</option>`).join('');
    const accent = /^#[0-9a-fA-F]{6}$/.test(cfg.accent || '') ? cfg.accent : (THEMES[cfg.theme] || THEMES[DEFAULT_THEME]).vars.accent;

    return `
        <div class="section">
            <h2 class="section-title">Look</h2>
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

function projectsSection() {
    const all = !cfg.projects.length;
    const names = [...new Set([...knownProjects, ...cfg.projects])].sort((a, b) => a.toLowerCase().localeCompare(b.toLowerCase()));
    const list = names.length
        ? names.map(n => `<label><input type="checkbox" data-project="${escapeHtml(n)}" ${cfg.projects.includes(n) ? 'checked' : ''} ${all ? 'disabled' : ''}> ${escapeHtml(n)}</label>`).join('')
        : '<div class="hint" style="margin-left:0">No projects found yet. Open one in Claude Code and it will show up here.</div>';

    return `
        <div class="section">
            <h2 class="section-title">Projects</h2>
            <label><input type="radio" name="pscope" value="all" ${all ? 'checked' : ''}> Show every active session</label>
            <label><input type="radio" name="pscope" value="some" ${all ? '' : 'checked'}> Only these projects</label>
            <div class="checklist">${list}</div>
            <div class="hint">Projects appear automatically as you open them in Claude Code, in the terminal or the desktop app. Click the ✕ on a row to hide a session until it has new activity.</div>
        </div>`;
}

function contextSection() {
    const size = cfg.contextTokens || 200000;
    return `
        <div class="section">
            <h2 class="section-title">Context window</h2>
            <label><input type="radio" name="ctx" value="200000" ${size === 200000 ? 'checked' : ''}> 200k tokens</label>
            <label><input type="radio" name="ctx" value="1000000" ${size === 1000000 ? 'checked' : ''}> 1M tokens (extended context)</label>
            <div class="hint">Used for the percentage when a session doesn't report its own size. Pick the window your model has.</div>
        </div>`;
}

function alertsSection() {
    const warn = cfg.alertWarn || 75;
    const urgent = cfg.alertUrgent || 90;
    const opts = (list, current) => list.map(v => `<option value="${v}" ${v === current ? 'selected' : ''}>${v}%</option>`).join('');
    return `
        <div class="section">
            <h2 class="section-title">Alerts</h2>
            <label><input type="checkbox" id="alertsOn" ${cfg.alertsOff ? '' : 'checked'}> Notify me when a session's context is filling up</label>
            <div class="hint">A Windows notification appears the first time a session passes each level, so you can save your work before Claude compacts it. Sessions already past a level when HeadRoom starts don't alert.</div>
            <div class="row"><span>First alert</span><select id="alertWarn" ${cfg.alertsOff ? 'disabled' : ''}>${opts([50, 60, 70, 75, 80], warn)}</select></div>
            <div class="row"><span>Urgent</span><select id="alertUrgent" ${cfg.alertsOff ? 'disabled' : ''}>${opts([80, 85, 90, 95], urgent)}</select></div>
            <label><input type="checkbox" id="soundOn" ${cfg.soundOff ? '' : 'checked'} ${cfg.alertsOff ? 'disabled' : ''}> Play a sound</label>
            <div class="row"><button class="btn" data-action="test-alert" ${cfg.alertsOff ? 'disabled' : ''}>Send a test alert</button></div>
        </div>`;
}

function windowSection() {
    return `
        <div class="section">
            <h2 class="section-title">Window</h2>
            <label><input type="checkbox" id="fixedSize" ${cfg.fixedSize ? 'checked' : ''}> Keep my window size</label>
            <div class="hint">By default HeadRoom fits its height to what it is showing, so there's no empty space. Tick this to size it yourself.</div>
        </div>`;
}

function limitsSection() {
    const minutes = cfg.pollMinutes || 10;
    const opts = [5, 10, 15, 30].map(m => `<option value="${m}" ${minutes === m ? 'selected' : ''}>every ${m} minutes</option>`).join('');
    return `
        <div class="section">
            <h2 class="section-title">Limits</h2>
            <label><input type="checkbox" id="poll" ${cfg.pollLimits ? 'checked' : ''} ${hasCli ? '' : 'disabled'}> Check my 5-hour and weekly limits automatically</label>
            <div class="hint">${hasCli ? '<span class="ok">Claude Code found.</span>' : '<span class="bad">Claude Code (the claude command) was not found, so this is unavailable.</span>'}
                Each check runs one tiny request under your own Claude login: about 40,000 mostly cached tokens, roughly $0.002 at API prices. It slowly draws on your plan limits (or bills your API key, if the claude command uses one), and HeadRoom never sees or stores your login.
                It only checks while a session is active, and stops once a limit is used up.</div>
            <div class="row"><span>Frequency</span><select id="pollMin" ${cfg.pollLimits ? '' : 'disabled'}>${opts}</select></div>
            <div class="hint">Off by default. Without it, limits update only when a terminal session reports them (see the README) and the ↻ Refresh button does nothing.</div>
        </div>`;
}

/* ---------- Settings panel ---------- */

async function openSettings() {
    await loadState();
    wizardMode = false;
    await growWindow();
    renderSettings();
    overlay().hidden = false;
}

function renderSettings() {
    const top = overlay().scrollTop;
    overlay().innerHTML = `
        <h2>Settings</h2>
        <p class="lead">Changes save as you make them.</p>
        ${lookSection()}${projectsSection()}${contextSection()}${limitsSection()}${alertsSection()}${windowSection()}
        <div class="nav">
            <button class="btn" data-action="rerun">Run setup again</button>
            <span class="spacer"></span>
            <button class="btn primary" data-action="close">Done</button>
        </div>`;
    overlay().scrollTop = top;
}

/* ---------- Setup wizard ---------- */

const WIZARD_STEPS = [
    () => `
        <h2>Welcome to HeadRoom</h2>
        <p class="lead">HeadRoom shows how much room you have left in Claude Code: the context window of each project you have open, and your 5-hour and weekly limits.</p>
        <p class="lead">This takes about a minute. Everything stays on your computer, and you can change any of it later from the gear at the top right.</p>`,
    () => `<h2>Pick a look</h2><p class="lead">Choose a theme, then adjust the accent color and font if you like.</p>${lookSection()}`,
    () => `<h2>Which projects?</h2><p class="lead">HeadRoom finds your sessions on its own. Limit it to the projects you care about, or show everything.</p>${projectsSection()}`,
    () => `<h2>Context size</h2><p class="lead">How big is your model's context window?</p>${contextSection()}`,
    () => `<h2>Keeping limits current</h2><p class="lead">Your 5-hour and weekly limits only come from Claude Code itself. Choose whether HeadRoom may ask for them in the background.</p>${limitsSection()}`,
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
            case 'next': wizardStep++; renderWizard(); break;
            case 'back': wizardStep--; renderWizard(); break;
            case 'finish': cfg.setupDone = true; persist(); closeOverlay(); break;
            case 'close': closeOverlay(); break;
            case 'test-alert': window.go.main.App.TestAlert(); break;
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
        } else if (t.dataset && t.dataset.project !== undefined) {
            const name = t.dataset.project;
            cfg.projects = t.checked ? [...new Set([...cfg.projects, name])] : cfg.projects.filter(p => p !== name);
            persist();
            // An empty selection would silently mean "show everything", so reflect that
            if (!cfg.projects.length) rerender();
        } else if (t.name === 'ctx') {
            cfg.contextTokens = parseInt(t.value, 10);
            persist();
        } else if (t.id === 'poll') {
            cfg.pollLimits = t.checked;
            if (t.checked && !cfg.pollMinutes) cfg.pollMinutes = 10;
            persist();
            rerender();
        } else if (t.id === 'alertsOn') {
            cfg.alertsOff = !t.checked;
            persist();
            rerender();
        } else if (t.id === 'soundOn') {
            cfg.soundOff = !t.checked;
            persist();
        } else if (t.id === 'alertWarn') {
            cfg.alertWarn = parseInt(t.value, 10);
            if (cfg.alertUrgent <= cfg.alertWarn) cfg.alertUrgent = Math.min(95, cfg.alertWarn + 5);
            persist();
            rerender();
        } else if (t.id === 'alertUrgent') {
            cfg.alertUrgent = parseInt(t.value, 10);
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
