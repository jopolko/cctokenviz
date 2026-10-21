// Stands in for the Wails backend so the real frontend can be rendered in a plain browser with
// demo data (no real project names). Used by scripts/screenshots.py. URL: ?theme=linux&spend=1
(function () {
    const q = new URLSearchParams(location.search);
    const now = Math.floor(Date.now() / 1000);
    const cfg = {
        projects: [], activeMinutes: 30, contextTokens: 200000, projectFolders: [], projectContext: null,
        hiddenProjects: [], pollLimits: true, pollMinutes: 10, 
        theme: q.get('theme') || 'linux', accent: '', font: '', fontSize: 0, setupDone: true, fixedSize: true,
    };
    const s = (id, name, title, pct, idle) => ({
        id, projectName: name, projectPath: name, title, usedTokens: Math.round(pct * 2000),
        freeTokens: 200000 - Math.round(pct * 2000), percentage: pct, lastUpdated: '10:42:07',
        idle: !!idle, idleSeconds: idle ? 1260 : 0,
    });
    const sessions = [
        s('a.jsonl', 'web-app', 'Add dark mode toggle', 34),
        s('b.jsonl', 'api-server', 'Fix pagination bug', 68),
        s('c.jsonl', 'mobile-client', 'Refactor navigation', 91),
        s('d.jsonl', 'docs-site', 'Update the docs', 22, true),
    ];
    const limits = { windows: [
        { label: '5-hour', percentage: 47, resetsAt: now + 2 * 3600 + 14 * 60 },
        { label: 'Weekly', percentage: 29, resetsAt: now + 3 * 86400 + 4 * 3600 },
    ], asOf: now - 40, paused: false };
    const usage = { approx: false, periods: [
        { label: 'Today', inputTokens: 1200, outputTokens: 61000, cacheRead: 7400000, cacheWrite: 310000, totalTokens: 7772200, cost: 4.12, cacheSaved: 19.98, cacheSavedTok: 7400000 },
        { label: '7 days', inputTokens: 9800, outputTokens: 402000, cacheRead: 52000000, cacheWrite: 2100000, totalTokens: 54511800, cost: 28.6, cacheSaved: 140.4, cacheSavedTok: 52000000 },
    ] };
    const bdNow = Math.floor(Date.now() / 1000);
    const breakdown = { startLabel: 'Carried over', startAt: bdNow - 7200, start: 38000, total: 144000, smallCount: 9, smallTokens: 11000, rows: [
        { label: 'Fix the broken imports in the uploader module after we renamed the shared helper functions', actions: 'Claude read 3 files, edited 2 files', reply: 'Renamed the helper and fixed the broken imports', tokens: 14000, at: bdNow - 6600 },
        { label: 'Yes, go ahead and run the whole test suite again now that the retry change is finished', actions: 'Claude ran 4 commands, read 2 files', reply: 'All 48 tests pass after the retry change', tokens: 41000, at: bdNow - 5400 },
        { label: 'Add retry with exponential backoff to the uploader so that a flaky network does not lose files', actions: 'Claude read 5 files, edited 3 files, ran 2 commands', reply: 'Added retry with backoff to the uploader', tokens: 21000, at: bdNow - 3000 },
        { label: 'Why is the production build so slow lately, and what can we cache to make it faster?', actions: 'Claude ran 6 commands', reply: 'The build is slow because the images are not cached', tokens: 19000, at: bdNow - 1500 },
    ] };
    const ok = (v) => () => Promise.resolve(v);
    window.go = { main: { App: {
        GetSettings: ok(cfg), SaveSettings: ok(), GetSessions: ok(sessions), GetLimits: ok(limits), GetUsage: ok(usage), GetBreakdown: ok(breakdown),
        RefreshLimits: ok(limits), SetOrder: ok(), DismissSession: ok(), KnownProjects: ok([]), DetectedContext: ok({}),
        HasClaudeCLI: ok(true), EnvCheck: ok({ logs: true, cli: true, statusline: true }), ExpandForPanel: ok({ w: 240, h: 600, x: 0, y: 0 }), SetSettingsOpen: ok(), AddProjectFolder: ok(''),
    } } };
    // ?spend=1 clicks the on-demand "Check API cost" button, as a user would
    if (q.get('spend') === '1') document.addEventListener('DOMContentLoaded', () => setTimeout(() => window.checkUsage && window.checkUsage(), 400));
    // ?only=usage shows just the API cost section (for the README), hiding the rest of the window
    if (q.get('only') === 'usage') document.addEventListener('DOMContentLoaded', () => { const st = document.createElement('style'); st.textContent = '#main > header, .context, #limits { display: none !important; }'; document.head.appendChild(st); });
    // ?breakdown=1 opens the first row's context breakdown, on a row sized to match it, alone in the window
    if (q.get('breakdown') === '1') {
        sessions[0].percentage = 72; sessions[0].usedTokens = 144000; sessions[0].freeTokens = 56000;
        document.addEventListener('DOMContentLoaded', () => { const st = document.createElement('style'); st.textContent = '#main > header, #limits, .session ~ .session, #usage { display: none !important; }'; document.head.appendChild(st); });
    }
    if (q.get('breakdown') === '1') document.addEventListener('DOMContentLoaded', () => setTimeout(() => { const f = document.querySelector('.session-footer'); if (f) f.click(); }, 700));
    // ?settings=1 opens Settings (?themes=1 with the theme list expanded)
    if (q.get('settings') === '1') document.addEventListener('DOMContentLoaded', () => setTimeout(() => {
        window.openSettings().then(() => { if (q.get('themes') === '1') document.querySelector('[data-action=themes-toggle]').click(); });
    }, 500));
    window.runtime = { WindowGetSize: ok({ w: 240, h: 600 }), WindowSetSize: ok(), WindowGetPosition: ok({ x: 0, y: 0 }),
        WindowSetPosition: ok(), WindowCenter: ok() };
})();
