// Update interval in milliseconds
const UPDATE_INTERVAL = 2000;

// Format token count (e.g., 133000 -> "133k")
function formatTokens(tokens) {
    if (tokens >= 1000) {
        return Math.floor(tokens / 1000) + 'k';
    }
    return tokens.toString();
}

// Get state class based on percentage
function getStateClass(percentage) {
    if (percentage >= 90) return 'state-critical';
    if (percentage >= 75) return 'state-danger';
    if (percentage >= 60) return 'state-warning';
    return 'state-good';
}

function escapeHtml(text) {
    return String(text).replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

// "+N" after a project name: other sessions on the same project, folded into this row
function othersTag(session) {
    const others = session.others || [];
    if (!others.length) return '';
    const tip = 'Also on this project: ' + others.map(o => `${o.percentage}%${o.title ? ' ' + o.title : ''}${o.idle ? ' (idle)' : ''}`).join('; ');
    return `<span class="others" title="${escapeHtml(tip)}">+${others.length}</span>`;
}

// What is using a session's context: its starting size, then what each prompt added. Opened by
// clicking a row's Used / Free line; numbers come straight from the session log.
const openBreakdowns = new Set();
const breakdownHtml = {};
const breakdownAt = {}; // used tokens when a breakdown was last fetched

function breakdownRows(b) {
    if (!b || !b.total) return '<div class="bd-empty">Nothing to show yet</div>';
    const now = Math.floor(Date.now() / 1000);
    const age = (t) => t ? formatDuration(now - t) : '';
    const row = (tokens, label, at, cls, tip, sub) => `<div class="bd-row ${cls || ''}" title="${escapeHtml(tip || label)}"><span class="bd-tok">${formatBigTokens(tokens)}</span><span class="bd-label"><span class="bd-you">${escapeHtml(label)}</span>${sub ? `<span class="bd-act">${escapeHtml(sub)}</span>` : ''}</span><span class="bd-age">${age(at)}</span></div>`;
    const biggest = Math.max(0, ...(b.rows || []).map(r => r.tokens));
    const small = b.smallCount ? `<div class="bd-row bd-small"><span class="bd-tok">${formatBigTokens(b.smallTokens)}</span><span class="bd-label"><span class="bd-you">${b.smallCount} smaller prompt${b.smallCount > 1 ? 's' : ''}</span></span><span class="bd-age"></span></div>` : '';
    return row(b.start, b.startLabel, b.startAt, 'bd-start', b.startLabel === 'Carried over' ? 'Summary, system prompt and tools at the start of this stretch' : 'System prompt, tools and your first prompt')
        + (b.rows || []).map(r => row(r.tokens, r.label, r.at, r.tokens === biggest && biggest > 0 ? 'bd-big' : '', r.reply ? "Claude's answer: " + r.reply : r.label, r.actions)).join('')
        + small
        + `<div class="bd-foot">What each prompt added, in your words, with what Claude did to answer. Largest is highlighted.</div>`;
}

async function loadBreakdown(id) {
    try {
        const b = await window.go.main.App.GetBreakdown(id);
        breakdownHtml[id] = breakdownRows(b);
        breakdownAt[id] = b ? b.total : 0;
    } catch (error) {
        console.error('Error fetching breakdown:', error);
        breakdownHtml[id] = '<div class="bd-empty">Could not read this session</div>';
    }
    const box = document.querySelector(`.breakdown[data-for="${CSS.escape(id)}"]`);
    if (box) { box.innerHTML = breakdownHtml[id]; scheduleFit(); }
}

async function toggleBreakdown(id) {
    const box = document.querySelector(`.breakdown[data-for="${CSS.escape(id)}"]`);
    if (openBreakdowns.has(id)) {
        openBreakdowns.delete(id);
        if (box) box.hidden = true;
        scheduleFit();
        return;
    }
    openBreakdowns.add(id);
    if (box) box.hidden = false;
    await loadBreakdown(id);
}

// Render a single session
function renderSession(session) {
    const stateClass = getStateClass(session.percentage);
    const id = escapeHtml(session.id);

    return `
        <div class="session ${stateClass}${session.idle ? ' idle' : ''}" data-id="${id}">
            <div class="session-header">
                <span class="grip" title="Drag to reorder">⋮⋮</span>
                <span class="project-name" title="${escapeHtml(session.title || session.projectPath)}">${escapeHtml(session.projectName)}${othersTag(session)}${session.title ? '<small class="session-title"><span class="title-text">' + escapeHtml(session.title) + '</span></small>' : ''}</span>
                <span class="pct-wrap">
                    <span class="percentage">${session.percentage}%</span>
                    ${session.idle ? `<small class="idle-tag" title="No activity since the last reading">idle ${formatDuration(session.idleSeconds)}</small>` : ''}
                </span>
                <button class="dismiss" title="Hide until it has new activity" onclick="dismissSession('${id}')">✕</button>
            </div>
            ${meterHtml(session.percentage)}
            <div class="session-footer clickable" title="What is using this context" onclick="toggleBreakdown('${id}')">
                <span>${LABELS.used}: ${formatTokens(session.usedTokens)}</span>
                <span>${LABELS.free}: ${formatTokens(session.freeTokens)}</span>
            </div>
            <div class="breakdown" data-for="${id}"${openBreakdowns.has(session.id) ? '' : ' hidden'}>${breakdownHtml[session.id] || ''}</div>
        </div>
    `;
}

// Hide a session until it sees new activity
async function dismissSession(id) {
    try {
        await window.go.main.App.DismissSession(id);
        lastSessionsHtml = null;
        updateSessions();
    } catch (error) {
        console.error('Error dismissing session:', error);
    }
}

// Manual refresh: poll now, then redraw (the call takes a few seconds)
async function refreshLimits() {
    const btn = document.getElementById('refreshLimits');
    if (btn) { btn.disabled = true; btn.textContent = 'Checking…'; }
    try {
        await window.go.main.App.RefreshLimits();
    } catch (error) {
        console.error('Error refreshing limits:', error);
    }
    updateLimits();
}

// Format seconds as "2h 14m" / "3d 4h" / "12m"
function formatDuration(seconds) {
    if (seconds <= 0) return 'now';
    const d = Math.floor(seconds / 86400);
    const h = Math.floor((seconds % 86400) / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    if (d > 0) return `${d}d ${h}h`;
    if (h > 0) return `${h}h ${m}m`;
    return `${Math.max(m, 1)}m`;
}

// The clock time a limit resets at: "5:20 PM" today, "Tomorrow 2:50 AM", or "Sun 12:00 PM" further out
function formatResetClock(epochSeconds) {
    const when = new Date(epochSeconds * 1000);
    const time = when.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
    const dayStart = (d) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
    const days = Math.round((dayStart(when) - dayStart(new Date())) / 86400000);
    if (days <= 0) return time;
    if (days === 1) return `Tomorrow ${time}`;
    return `${when.toLocaleDateString([], { weekday: 'short' })} ${time}`;
}

// Render one limit: a usage bar, the countdown, and the clock time it resets
function renderLimit(limit) {
    if (limit.expired) {
        // The window just reset and no new reading has arrived: say so rather than show a number
        const next = limit.resetsAt ? `next reset in ${formatDuration(limit.resetsAt - Math.floor(Date.now() / 1000))}` : 'starts with your next message';
        return `
        <div class="limit state-good">
            <div class="limit-header">
                <span class="limit-label">${limit.label}</span>
                <span class="percentage" title="Reset. Waiting for a fresh reading.">reset</span>
            </div>
            ${meterHtml(0)}
            <div class="limit-reset"><span>${next}</span></div>
        </div>`;
    }
    const pct = Math.min(100, Math.round(limit.percentage));
    const remaining = limit.resetsAt - Math.floor(Date.now() / 1000);

    return `
        <div class="limit ${getStateClass(pct)}">
            <div class="limit-header">
                <span class="limit-label">${limit.label}</span>
                <span class="percentage">${pct}%</span>
            </div>
            ${meterHtml(pct)}
            <div class="limit-reset"><span>${formatDuration(remaining)}</span><span class="limit-at" title="${new Date(limit.resetsAt * 1000).toLocaleString()}">${formatResetClock(limit.resetsAt)}</span></div>
        </div>
    `;
}

// Update the limits panel at the bottom
async function updateLimits() {
    const panel = document.getElementById('limits');
    try {
        const limits = await window.go.main.App.GetLimits();
        const before = panel.innerHTML.length;
        if (!limits || !limits.windows || limits.windows.length === 0) {
            panel.innerHTML = `
                <h2 class="section-title">${LABELS.limits}</h2>
                <div class="limits-empty">Shown after your next message in a Claude Code session. Turn on automatic checks in Settings (the gear, top right) to keep them current.</div>`;
            return;
        }
        panel.innerHTML = `
            <h2 class="section-title">${LABELS.limits}</h2>
            ${limits.windows.map(renderLimit).join('')}
            <div class="limits-asof">
                <span>${limits.paused ? 'Limit reached, not checking' : 'Last reported ' + formatDuration(Math.floor(Date.now() / 1000) - limits.asOf) + ' ago'}</span>
                ${limits.paused ? '' : '<button id="refreshLimits" class="refresh" title="Check now (uses one tiny request; needs automatic checks turned on in settings)" onclick="refreshLimits()">↻ Refresh</button>'}
            </div>`;
        if (panel.innerHTML.length !== before) scheduleFit();
    } catch (error) {
        console.error('Error fetching limits:', error);
    }
}

// Format a token count for the spend panel: 1.2M, 340k, 900
function formatBigTokens(n) {
    if (n >= 1e6) return (n / 1e6).toFixed(n >= 1e7 ? 0 : 1) + 'M';
    if (n >= 1e3) return Math.round(n / 1e3) + 'k';
    return String(n);
}

function formatMoney(x) {
    return '$' + (x >= 100 ? x.toFixed(0) : x.toFixed(2));
}

// API cost, on demand. Nothing is read until the button is clicked, and results stay until
// closed or refreshed. It only adds up the session logs on this computer: it spends no tokens
// and is information only, never a charge.
let usageResult = null;
let usageBusy = false;

async function checkUsage() {
    if (usageBusy) return;
    usageBusy = true;
    renderUsage();
    try {
        usageResult = await window.go.main.App.GetUsage();
    } catch (error) {
        console.error('Error fetching usage:', error);
    }
    usageBusy = false;
    renderUsage();
}

function closeUsage() {
    usageResult = null;
    renderUsage();
}

function renderUsage() {
    const panel = document.getElementById('usage');
    let html;
    if (!usageResult) {
        html = `<button class="usage-check" ${usageBusy ? 'disabled' : ''} onclick="checkUsage()" title="Adds up the token usage in your Claude Code logs for today and the last 7 days, priced at API rates. It reads files on this computer only: no tokens are spent and nothing is billed.">
            <span>${usageBusy ? 'Checking…' : '$ Check API cost'}</span><small>${usageBusy ? '' : 'free'}</small></button>`;
    } else {
        const u = usageResult;
        const rows = (u.periods || []).map(p => `
            <div class="usage-row" title="Fresh input ${formatBigTokens(p.inputTokens)}, output ${formatBigTokens(p.outputTokens)}, cache read ${formatBigTokens(p.cacheRead)}, cache write ${formatBigTokens(p.cacheWrite)}">
                <span class="limit-label">${p.label}</span>
                <span>${formatBigTokens(p.totalTokens)} tok</span>
                <span class="percentage">${u.approx ? '~' : ''}${formatMoney(p.cost)}</span>
            </div>
            <div class="usage-sub">cache saved ${formatMoney(p.cacheSaved)} (${formatBigTokens(p.cacheSavedTok)} tok)</div>`).join('');
        html = `<div class="usage-head"><span class="section-title">API cost</span>
                <span><button class="refresh" onclick="checkUsage()" title="Check again">${usageBusy ? '…' : '↻'}</button> <button class="dismiss" onclick="closeUsage()" title="Close">✕</button></span></div>
            ${rows}
            <div class="limits-asof"><span>Information only, not a charge. Estimated at API prices.</span></div>`;
    }
    if (panel.innerHTML.length !== html.length) { panel.innerHTML = html; scheduleFit(); } else { panel.innerHTML = html; }
}

// Render empty state
function renderEmpty() {
    return `
        <div class="empty-state">
            <div class="icon">💤</div>
            <p>No active Claude sessions<br>Open a project in Claude Code, in the terminal or the desktop app, and it will appear here</p>
        </div>
    `;
}

let lastSessionsHtml = null;
let draggingRow = null;
let appCfg = {};

// Fit the window's height to its content, so there is never a blank gap and nothing is cut off.
// Turned off by "Keep my window size" in Settings. Only runs when the content changes.
let fitTimer = null;
function scheduleFit() {
    clearTimeout(fitTimer);
    fitTimer = setTimeout(fitWindow, 300);
}

async function fitWindow() {
    const panel = document.getElementById('overlay');
    if (appCfg.fixedSize || !panel.hidden || draggingRow) return;
    try {
        const container = document.querySelector('.container');
        const previous = container.style.height;
        container.style.height = 'auto';
        const natural = container.offsetHeight;
        container.style.height = previous;

        const delta = natural - window.innerHeight;
        if (Math.abs(delta) < 6) return;
        const size = await window.runtime.WindowGetSize();
        const limit = Math.floor(window.screen.availHeight * 0.92);
        const height = Math.max(300, Math.min(limit, size.h + delta));
        if (Math.abs(height - size.h) >= 6) window.runtime.WindowSetSize(size.w, height);
    } catch (error) { /* leave the window as it is */ }
}

// Update the UI with session data
async function updateSessions() {
    if (draggingRow) return; // don't redraw under the user's pointer

    const container = document.getElementById('sessions');
    const status = document.getElementById('status');

    try {
        const sessions = await window.go.main.App.GetSessions();
        const list = sessions || [];
        const signature = list.map(x => `${x.id}|${x.projectName}|${x.title}|${x.idle ? 'idle' : 'live'}|${(x.others || []).length}`).join('\n');

        if (list.length && signature === lastSessionsHtml && container.querySelectorAll('.session').length === list.length) {
            // Same rows, same names: refresh only the numbers and meters
            list.forEach((session) => {
                const row = [...container.querySelectorAll('.session')].find(r => r.dataset.id === session.id);
                if (row) updateRow(row, session);
            });
        } else {
            container.innerHTML = list.length ? list.map(renderSession).join('') : renderEmpty();
            lastSessionsHtml = signature;
            applyTicker(container);
            scheduleFit();
        }
        status.textContent = (!sessions || sessions.length === 0) ? 'No sessions'
            : `${sessions.length} session${sessions.length > 1 ? 's' : ''}`;

    } catch (error) {
        console.error('Error fetching sessions:', error);
        status.textContent = 'Error';
    }
}

// Refresh one row's numbers without rebuilding it (keeps the title ticker running)
function updateRow(row, session) {
    row.className = `session ${getStateClass(session.percentage)}${session.idle ? ' idle' : ''}${row.classList.contains('dragging') ? ' dragging' : ''}`;
    const idleTag = row.querySelector('.idle-tag');
    if (idleTag) idleTag.textContent = `idle ${formatDuration(session.idleSeconds)}`;
    row.querySelector('.percentage').textContent = session.percentage + '%';
    row.querySelector('.meter').outerHTML = meterHtml(session.percentage);
    const spans = row.querySelectorAll('.session-footer span');
    spans[0].textContent = `${LABELS.used}: ${formatTokens(session.usedTokens)}`;
    spans[1].textContent = `${LABELS.free}: ${formatTokens(session.freeTokens)}`;
    if (openBreakdowns.has(session.id) && breakdownAt[session.id] !== session.usedTokens) loadBreakdown(session.id);
}

// A session title too long for its row slowly scrolls to the left and back, so it can be read in full
function applyTicker(container) {
    container.querySelectorAll('.session-title').forEach((box) => {
        const text = box.firstElementChild;
        if (!text) return;
        text.classList.remove('scroll');
        const distance = text.scrollWidth - box.clientWidth;
        if (distance > 2) {
            text.style.setProperty('--dist', distance + 'px');
            text.style.animationDuration = (distance / 22 + 5) + 's'; // ~22px a second, plus pauses at each end
            text.classList.add('scroll');
        }
    });
}

let tickerTimer = null;
window.addEventListener('resize', () => {
    clearTimeout(tickerTimer);
    tickerTimer = setTimeout(() => applyTicker(document.getElementById('sessions')), 250);
});

// Drag a row by its grip to reorder; the new order is saved and stays put
function setupDragging() {
    const container = document.getElementById('sessions');

    container.addEventListener('pointerdown', (e) => {
        const grip = e.target.closest('.grip');
        if (!grip || e.button !== 0) return;
        const row = grip.closest('.session');
        if (!row) return;

        e.preventDefault();
        draggingRow = row;
        row.classList.add('dragging');
        document.body.classList.add('is-dragging');

        const move = (ev) => {
            const siblings = [...container.querySelectorAll('.session')].filter(r => r !== row);
            const next = siblings.find(r => {
                const box = r.getBoundingClientRect();
                return ev.clientY < box.top + box.height / 2;
            });
            if (next) container.insertBefore(row, next);
            else container.appendChild(row);

            // Scroll when dragging near the edge of the list
            const box = container.getBoundingClientRect();
            if (ev.clientY < box.top + 24) container.scrollTop -= 8;
            else if (ev.clientY > box.bottom - 24) container.scrollTop += 8;
        };

        const finish = async () => {
            document.removeEventListener('pointermove', move);
            document.removeEventListener('pointerup', finish);
            document.removeEventListener('pointercancel', finish);
            row.classList.remove('dragging');
            document.body.classList.remove('is-dragging');
            const ids = [...container.querySelectorAll('.session')].map(r => r.dataset.id);
            try {
                await window.go.main.App.SetOrder(ids);
            } catch (error) {
                console.error('Error saving order:', error);
            }
            draggingRow = null;
            lastSessionsHtml = null;
            updateSessions();
        };

        document.addEventListener('pointermove', move);
        document.addEventListener('pointerup', finish);
        document.addEventListener('pointercancel', finish);
    });
}

// Initialize
document.addEventListener('DOMContentLoaded', async () => {
    setupDragging();

    // Apply the saved look first so the window never flashes the wrong theme
    try {
        const cfg = await window.go.main.App.GetSettings();
        appCfg = cfg;
        applyAppearance(cfg);
        if (!cfg.setupDone) openSetup();
    } catch (error) {
        console.error('Error loading settings:', error);
        applyAppearance({});
    }

    updateSessions();
    updateLimits();
    renderUsage();

    setInterval(updateSessions, UPDATE_INTERVAL);
    setInterval(updateLimits, UPDATE_INTERVAL);
});
