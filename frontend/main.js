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

// Render a single session
function renderSession(session) {
    const stateClass = getStateClass(session.percentage);
    const id = escapeHtml(session.id);

    return `
        <div class="session ${stateClass}" data-id="${id}">
            <div class="session-header">
                <span class="grip" title="Drag to reorder">⋮⋮</span>
                <span class="project-name" title="${escapeHtml(session.title || session.projectPath)}">${escapeHtml(session.projectName)}${session.title ? ' <small class="session-title">' + escapeHtml(session.title) + '</small>' : ''}</span>
                <span class="percentage">${session.percentage}%</span>
                <button class="dismiss" title="Hide until it has new activity" onclick="dismissSession('${id}')">✕</button>
            </div>
            ${meterHtml(session.percentage)}
            <div class="session-footer">
                <span>${LABELS.used}: ${formatTokens(session.usedTokens)}</span>
                <span>${LABELS.free}: ${formatTokens(session.freeTokens)}</span>
            </div>
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
    const pct = Math.min(100, Math.round(limit.percentage));
    const remaining = limit.resetsAt - Math.floor(Date.now() / 1000);

    return `
        <div class="limit ${getStateClass(pct)}">
            <div class="limit-header">
                <span class="limit-label">${limit.label}</span>
                <span class="percentage">${pct}%</span>
            </div>
            ${meterHtml(pct)}
            <div class="limit-reset"><span>${LABELS.resets} ${formatDuration(remaining)}</span><span class="limit-at" title="${new Date(limit.resetsAt * 1000).toLocaleString()}">${formatResetClock(limit.resetsAt)}</span></div>
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

// Render empty state
function renderEmpty() {
    return `
        <div class="empty-state">
            <div class="icon">💤</div>
            <p>No active Claude sessions<br>Open a project in Claude Code and it will appear here</p>
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
    const lastUpdate = document.getElementById('lastUpdate');

    try {
        const sessions = await window.go.main.App.GetSessions();
        const html = (!sessions || sessions.length === 0) ? renderEmpty() : sessions.map(renderSession).join('');

        // Only touch the DOM when something changed, so rows don't flicker
        if (html !== lastSessionsHtml) {
            container.innerHTML = html;
            lastSessionsHtml = html;
            scheduleFit();
        }
        status.textContent = (!sessions || sessions.length === 0) ? 'No sessions'
            : `${sessions.length} session${sessions.length > 1 ? 's' : ''}`;

        lastUpdate.textContent = new Date().toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
        lastUpdate.title = 'Last refreshed';
    } catch (error) {
        console.error('Error fetching sessions:', error);
        status.textContent = 'Error';
    }
}

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

// Reset alert history
async function resetAlerts() {
    try {
        await window.go.main.App.ResetAlerts();
        // Visual feedback
        const btn = document.querySelector('footer button');
        btn.textContent = '✓';
        setTimeout(() => { btn.textContent = '🔔'; }, 1000);
    } catch (error) {
        console.error('Error resetting alerts:', error);
    }
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

    setInterval(updateSessions, UPDATE_INTERVAL);
    setInterval(updateLimits, UPDATE_INTERVAL);
});
