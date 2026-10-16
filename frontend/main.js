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

// Render a single session
function renderSession(session) {
    const stateClass = getStateClass(session.percentage);
    
    return `
        <div class="session ${stateClass}">
            <div class="session-header">
                <span class="project-name" title="${session.title || session.projectPath}">${session.projectName}${session.title ? ' <small class="session-title">' + session.title + '</small>' : ''}</span>
                <span class="percentage">${session.percentage}%</span>
                <button class="dismiss" title="Hide until it has new activity" onclick="dismissSession('${session.id}')">✕</button>
            </div>
            <div class="progress-container">
                <div class="progress-bar" style="width: ${session.percentage}%"></div>
            </div>
            <div class="session-footer">
                <span>Used: ${formatTokens(session.usedTokens)}</span>
                <span>Free: ${formatTokens(session.freeTokens)}</span>
            </div>
        </div>
    `;
}

// Hide a session until it sees new activity
async function dismissSession(id) {
    try {
        await window.go.main.App.DismissSession(id);
        updateSessions();
    } catch (error) {
        console.error('Error dismissing session:', error);
    }
}

function formatMoney(n) {
    return '$' + n.toFixed(2);
}

// Credit balances under the limits: Console (estimated from spend) and claude.ai (as typed)
async function updateCredits() {
    const panel = document.getElementById('credits');
    try {
        const c = await window.go.main.App.GetCredits();
        if (!c || (!c.hasConsole && !c.claudeAi)) {
            panel.innerHTML = '';
            panel.style.display = 'none';
            return;
        }
        panel.style.display = '';

        let rows = '';
        if (c.hasConsole) {
            const stateClass = c.consoleBalance < 1 ? 'state-critical' : c.consoleBalance < 5 ? 'state-warning' : 'state-good';
            const approx = c.consoleEstimated ? '≈ ' : '';
            const sub = c.consoleEstimated
                ? `${formatMoney(c.consoleSnapshot)} on ${c.consoleAsOf}, ${formatMoney(c.consoleSpentSince)} spent since`
                : (c.consoleNote || `Snapshot from ${c.consoleAsOf}`);
            rows += `
                <div class="credit ${stateClass}">
                    <div class="limit-header">
                        <span class="limit-label">Console (API)</span>
                        <span class="percentage">${approx}${formatMoney(Math.max(c.consoleBalance, 0))}</span>
                    </div>
                    <div class="limit-reset">${sub}</div>
                </div>`;
        }
        if (c.claudeAi) {
            rows += `
                <div class="credit state-good">
                    <div class="limit-header">
                        <span class="limit-label">claude.ai credits</span>
                        <span class="percentage">${c.claudeAi}</span>
                    </div>
                </div>`;
        }
        panel.innerHTML = `<div class="limits-title">Credits</div>${rows}`;
    } catch (error) {
        console.error('Error fetching credits:', error);
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

// Render one plan limit: a usage bar and the time until it resets
function renderLimit(limit) {
    const pct = Math.min(100, Math.round(limit.percentage));
    const remaining = limit.resetsAt - Math.floor(Date.now() / 1000);

    return `
        <div class="limit ${getStateClass(pct)}">
            <div class="limit-header">
                <span class="limit-label">${limit.label}</span>
                <span class="percentage">${pct}%</span>
            </div>
            <div class="progress-container limit-bar">
                <div class="progress-bar" style="width: ${pct}%"></div>
            </div>
            <div class="limit-reset">Resets in ${formatDuration(remaining)}</div>
        </div>
    `;
}

// Update the plan limits panel at the bottom
async function updateLimits() {
    const panel = document.getElementById('limits');
    try {
        const limits = await window.go.main.App.GetLimits();
        if (!limits || !limits.windows || limits.windows.length === 0) {
            panel.innerHTML = `
                <div class="limits-title">Plan limits</div>
                <div class="limits-empty">Shown after your next message in any Claude Code session</div>`;
            return;
        }
        panel.innerHTML = `
            <div class="limits-title">Plan limits</div>
            ${limits.windows.map(renderLimit).join('')}
            <div class="limits-asof">
                <span>${limits.paused ? 'Limit reached, not checking' : 'Last reported ' + formatDuration(Math.floor(Date.now() / 1000) - limits.asOf) + ' ago'}</span>
                ${limits.paused ? '' : '<button id="refreshLimits" class="refresh" title="Check now (uses one tiny request; needs pollLimits)" onclick="refreshLimits()">↻ Refresh</button>'}
            </div>`;
    } catch (error) {
        console.error('Error fetching limits:', error);
    }
}

// Render empty state
function renderEmpty() {
    return `
        <div class="empty-state">
            <div class="icon">💤</div>
            <p>No active Claude sessions<br>Sessions modified in the last 60 min will appear here</p>
        </div>
    `;
}

// Update the UI with session data
async function updateSessions() {
    const container = document.getElementById('sessions');
    const status = document.getElementById('status');
    const lastUpdate = document.getElementById('lastUpdate');
    
    try {
        // Call Go backend
        const sessions = await window.go.main.App.GetSessions();
        
        if (!sessions || sessions.length === 0) {
            container.innerHTML = renderEmpty();
            status.textContent = 'No sessions';
        } else {
            container.innerHTML = sessions.map(renderSession).join('');
            status.textContent = `${sessions.length} session${sessions.length > 1 ? 's' : ''}`;
        }
        
        // Update timestamp
        const now = new Date();
        lastUpdate.textContent = `Updated ${now.toLocaleTimeString()}`;
        
    } catch (error) {
        console.error('Error fetching sessions:', error);
        status.textContent = 'Error';
    }
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
document.addEventListener('DOMContentLoaded', () => {
    // Initial update
    updateSessions();
    updateLimits();

    // Periodic updates
    setInterval(updateSessions, UPDATE_INTERVAL);
    setInterval(updateLimits, UPDATE_INTERVAL);
    updateCredits();
    setInterval(updateCredits, 60 * 1000);
});
