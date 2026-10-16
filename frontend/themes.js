// A theme is a whole personality, not just a palette. Each one sets:
//   vars    - the color palette (CSS variables on :root)
//   look    - corner radius, border width, background pattern, heading decoration, logo,
//             glow, panel cut-outs and the heading font and weight
//   meter   - the indicator: its kind (block, led, heart, star, leaf, bullet, node, slant,
//             smooth...), how many segments, and optional ticks
//   font    - body font id from FONTS (headFont for headings)
//   labels  - the words for Context, Limits, Used, Free and Resets
// style.css reads all of it, so a new theme is just a new entry here.

const FLAT_LOOK = { radius: 8, bw: 1, bg: 'none', bgSize: 'auto', pre: '', post: '', logo: '', glow: false, caps: true, clip: 'none', headWeight: 800, headStyle: 'normal', headSpacing: '0.14em' };
const SMOOTH = { kind: 'smooth', n: 0 };
const PLAIN_LABELS = { context: 'Context', limits: 'Limits', used: 'Used', free: 'Free', resets: 'Resets in' };

// Pattern helpers (inline SVG and gradients, so nothing is downloaded)
const gridBg = (rgba, size) => ({ bg: `linear-gradient(${rgba} 1px, transparent 1px), linear-gradient(90deg, ${rgba} 1px, transparent 1px)`, bgSize: `${size}px ${size}px` });
const dotBg = (rgba, size) => ({ bg: `radial-gradient(${rgba} 1.2px, transparent 1.6px)`, bgSize: `${size}px ${size}px` });
const scanBg = (rgba) => ({ bg: `repeating-linear-gradient(0deg, ${rgba} 0 1px, transparent 1px 3px)`, bgSize: 'auto' });
const emojiBg = (a, b, size, opacity) => {
    const svg = `<svg xmlns='http://www.w3.org/2000/svg' width='${size}' height='${size}'>` +
        `<text x='8' y='30' font-size='20' opacity='${opacity}'>${a}</text>` +
        `<text x='${size / 2 + 8}' y='${size - 14}' font-size='16' opacity='${opacity}'>${b}</text></svg>`;
    return { bg: `url("data:image/svg+xml,${encodeURIComponent(svg)}")`, bgSize: `${size}px ${size}px` };
};
const starsBg = () => {
    const dots = [[12, 18, 1.2], [60, 40, 0.8], [110, 14, 1.4], [150, 70, 0.9], [36, 96, 1], [96, 120, 1.3], [140, 150, 0.8], [20, 150, 1.1], [70, 170, 0.7]]
        .map(([x, y, r]) => `<circle cx='${x}' cy='${y}' r='${r}' fill='white' opacity='0.55'/>`).join('');
    const svg = `<svg xmlns='http://www.w3.org/2000/svg' width='180' height='190'>${dots}</svg>`;
    return { bg: `url("data:image/svg+xml,${encodeURIComponent(svg)}")`, bgSize: '180px 190px' };
};

const THEMES = {
    /* ---------- Styles ---------- */
    linux: {
        name: 'Terminal', tag: 'For the Linux geek', group: 'styles', icon: '>_', font: 'cascadia', headFont: 'cascadia',
        vars: { bg: '#0a0e0a', panel: '#0d140d', panel2: '#0a110a', border: '#1d3a1d', track: '#142214', text: '#b6f5b6', dim: '#5fbf5f', strong: '#e6ffe6', accent: '#39ff14', good: '#39ff14', warn: '#e6ff4d', danger: '#ffb000', crit: '#ff4040' },
        look: { ...FLAT_LOOK, radius: 0, ...scanBg('rgba(0,0,0,0.28)'), pre: '$ ', logo: '>_ ', caps: false, headWeight: 700, headSpacing: '0.04em' },
        meter: { kind: 'block', n: 20 },
        labels: { context: 'context', limits: 'limits', used: 'used', free: 'free', resets: 'resets in' },
    },
    scifi: {
        name: 'Starship', tag: 'Bridge console', group: 'styles', icon: '◈', font: 'bahnsemi', headFont: 'bahnsemi',
        vars: { bg: '#060b1a', panel: '#0b1630', panel2: '#08112a', border: '#18407a', track: '#0f2347', text: '#cfe8ff', dim: '#86b4e4', strong: '#eaf6ff', accent: '#29d9ff', good: '#3dffc4', warn: '#ffd24a', danger: '#ff8a3d', crit: '#ff3b5c' },
        look: { ...FLAT_LOOK, radius: 0, ...gridBg('rgba(41,217,255,0.07)', 24), pre: '[ ', post: ' ]', logo: '◈ ', glow: true, headSpacing: '0.22em',
            clip: 'polygon(0 7px, 7px 0, 100% 0, 100% calc(100% - 7px), calc(100% - 7px) 100%, 0 100%)' },
        meter: { kind: 'smooth', n: 0, ticks: true, glow: true },
        labels: { ...PLAIN_LABELS, context: 'Context', limits: 'Limits', resets: 'T-minus' },
    },
    bitcoin: {
        name: 'Satoshi', tag: 'Number go up', group: 'styles', icon: '₿', font: 'franklin', headFont: 'franklin',
        vars: { bg: '#0d0d0d', panel: '#191919', panel2: '#141414', border: '#4a3414', track: '#2a2a2a', text: '#f2f2f2', dim: '#c2b08a', strong: '#ffffff', accent: '#f7931a', good: '#f7931a', warn: '#ffd166', danger: '#ff6b35', crit: '#ff3b3b' },
        look: { ...FLAT_LOOK, radius: 6, ...dotBg('rgba(247,147,26,0.12)', 26), pre: '₿ ', logo: '₿ ', glow: true, headSpacing: '0.1em' },
        meter: { kind: 'coin', n: 12 },
        labels: { ...PLAIN_LABELS, used: 'Spent', free: 'Left' },
    },
    ai: {
        name: 'Neural', tag: 'For the AI freak', group: 'styles', icon: '◉', font: 'consolas', headFont: 'consolas',
        vars: { bg: '#0b0820', panel: '#15113a', panel2: '#110e30', border: '#3b2f8f', track: '#241c5e', text: '#e8e4ff', dim: '#aca3f2', strong: '#ffffff', accent: '#a78bfa', good: '#34d399', warn: '#fbbf24', danger: '#fb7185', crit: '#f43f5e' },
        look: { ...FLAT_LOOK, radius: 12, ...dotBg('rgba(167,139,250,0.16)', 28), pre: '◉ ', logo: '◉ ', glow: true, caps: false, headWeight: 700, headSpacing: '0.03em' },
        meter: { kind: 'node', n: 14 },
        labels: { context: 'context_window', limits: 'rate_limits', used: 'used', free: 'free', resets: 'resets_in' },
    },
    cowboy: {
        name: 'Frontier', tag: 'Yeehaw', group: 'styles', icon: '★', font: 'rockwell', headFont: 'rockwell',
        vars: { bg: '#3b2616', panel: '#4a301c', panel2: '#43291a', border: '#8a5a2b', track: '#2e1d10', text: '#f3e2c3', dim: '#d2b184', strong: '#fff4dc', accent: '#e0a03a', good: '#9bc53d', warn: '#e0a03a', danger: '#d9622b', crit: '#c0392b' },
        look: { ...FLAT_LOOK, radius: 3, bw: 2, bg: 'repeating-linear-gradient(0deg, rgba(0,0,0,0.10) 0 2px, transparent 2px 7px)', pre: '★ ', post: ' ★', logo: '🤠 ', headWeight: 800 },
        meter: { kind: 'bullet', n: 12 },
        labels: { ...PLAIN_LABELS, free: 'Left', resets: 'Back in' },
    },
    anime: {
        name: 'Shonen', tag: 'Anime energy', group: 'styles', icon: '⚡', font: 'bahnschrift', headFont: 'impact',
        vars: { bg: '#101018', panel: '#1b1b2a', panel2: '#161622', border: '#ff3b3b', track: '#2a2a3c', text: '#ffffff', dim: '#ffd54a', strong: '#ffffff', accent: '#ff3b3b', good: '#3dd6ff', warn: '#ffd54a', danger: '#ff8a00', crit: '#ff1744' },
        look: { ...FLAT_LOOK, radius: 0, bw: 3, bg: 'repeating-conic-gradient(from 0deg at 50% 50%, rgba(255,255,255,0.05) 0deg 2deg, transparent 2deg 6deg)', post: '!', logo: '⚡ ', headWeight: 400, headStyle: 'italic', headSpacing: '0.08em' },
        meter: { kind: 'slant', n: 18 },
        labels: { ...PLAIN_LABELS, used: 'Power used', free: 'Power left' },
    },
    sweet: {
        name: 'Sweet', tag: 'Hearts and pastels', group: 'styles', icon: '💗', font: 'candara', headFont: 'inkfree',
        vars: { bg: '#fff0f6', panel: '#ffffff', panel2: '#fff7fb', border: '#ffb6d5', track: '#ffe0ee', text: '#6b3a55', dim: '#a0678a', strong: '#4a2040', accent: '#ff5fa2', good: '#ff5fa2', warn: '#ffa94d', danger: '#ff7aa2', crit: '#e8336d' },
        look: { ...FLAT_LOOK, radius: 20, ...emojiBg('💗', '✨', 100, 0.2), pre: '♡ ', post: ' ♡', logo: '💗 ', caps: false, headWeight: 400, headSpacing: '0.04em' },
        meter: { kind: 'heart', n: 10 },
        labels: { context: 'Context', limits: 'Limits', used: 'Used', free: 'Left', resets: 'Resets in' },
    },
    bloom: {
        name: 'Bloom', tag: 'Garden and eco', group: 'styles', icon: '🌿', font: 'constantia', headFont: 'constantia',
        vars: { bg: '#f1f0e4', panel: '#fbfaf2', panel2: '#f6f5e8', border: '#c5d1ae', track: '#dfe6cf', text: '#34453a', dim: '#6b7f62', strong: '#233328', accent: '#5f8f4e', good: '#5f9a4e', warn: '#e0a83a', danger: '#d9776b', crit: '#c0453a' },
        look: { ...FLAT_LOOK, radius: 16, ...emojiBg('🍃', '🌼', 104, 0.22), pre: '❦ ', post: ' ❦', logo: '🌿 ', caps: false, headWeight: 700, headStyle: 'italic', headSpacing: '0.04em' },
        meter: { kind: 'leaf', n: 12 },
        labels: { context: 'Context', limits: 'Limits', used: 'Used', free: 'Fresh', resets: 'Renews in' },
    },
    celestial: {
        name: 'Celestial', tag: 'Moon and stars', group: 'styles', icon: '☾', font: 'corbel', headFont: 'constantia',
        vars: { bg: '#120d2b', panel: '#1d1646', panel2: '#181239', border: '#4a3c96', track: '#2c2468', text: '#ece6ff', dim: '#b6a9ee', strong: '#ffffff', accent: '#ffd166', good: '#ffd166', warn: '#ff9f68', danger: '#ff7aa8', crit: '#ff4d6d' },
        look: { ...FLAT_LOOK, radius: 14, ...starsBg(), pre: '☾ ', post: ' ☽', logo: '☾ ', glow: true, caps: false, headWeight: 700, headStyle: 'italic', headSpacing: '0.06em' },
        meter: { kind: 'star', n: 10 },
        labels: { ...PLAIN_LABELS, free: 'Left', resets: 'Rises again in' },
    },
    lab: {
        name: 'Lab Notebook', tag: 'Scientific method', group: 'styles', icon: '⚗', font: 'cambria', headFont: 'cambria',
        vars: { bg: '#f7f9fc', panel: '#ffffff', panel2: '#ffffff', border: '#9db4d6', track: '#e4ebf5', text: '#1d2b44', dim: '#4d6288', strong: '#0f1b30', accent: '#1f5fbf', good: '#1f8a5b', warn: '#c77a00', danger: '#c8432f', crit: '#a1211a' },
        look: { ...FLAT_LOOK, radius: 2, ...gridBg('rgba(31,95,191,0.10)', 16), logo: '⚗ ', caps: false, headWeight: 700, headSpacing: '0.02em' },
        meter: { kind: 'smooth', n: 0, ticks: true },
        labels: { context: 'Fig. 1 — Context', limits: 'Fig. 2 — Limits', used: 'Used', free: 'Free', resets: 'Reset after' },
    },

    /* ---------- Simple ---------- */
    midnight: { name: 'Midnight', tag: 'Plain dark', group: 'simple', icon: '', font: 'system', headFont: 'system', look: FLAT_LOOK, meter: SMOOTH, labels: PLAIN_LABELS, vars: { bg: '#14161c', panel: '#1c1f27', panel2: '#181b22', border: '#2b3040', track: '#2b3040', text: '#e6e9f0', dim: '#aab1c3', strong: '#ffffff', accent: '#7aa2f7', good: '#4ade80', warn: '#fbbf24', danger: '#f87171', crit: '#ff4d4d' } },
    light: { name: 'Daylight', tag: 'Plain light', group: 'simple', icon: '', font: 'system', headFont: 'system', look: FLAT_LOOK, meter: SMOOTH, labels: PLAIN_LABELS, vars: { bg: '#f5f6f8', panel: '#ffffff', panel2: '#ffffff', border: '#d9dde5', track: '#e3e6ec', text: '#1f2430', dim: '#5b6475', strong: '#111827', accent: '#2563eb', good: '#16a34a', warn: '#d97706', danger: '#dc2626', crit: '#b91c1c' } },
    contrast: { name: 'High contrast', tag: 'Easiest to read', group: 'simple', icon: '', font: 'system', headFont: 'system', look: FLAT_LOOK, meter: SMOOTH, labels: PLAIN_LABELS, vars: { bg: '#000000', panel: '#000000', panel2: '#000000', border: '#ffffff', track: '#333333', text: '#ffffff', dim: '#dddddd', strong: '#ffffff', accent: '#ffff00', good: '#00ff00', warn: '#ffff00', danger: '#ff8c00', crit: '#ff0000' } },
};
const DEFAULT_THEME = 'midnight';
const THEME_GROUPS = [['styles', 'Styles'], ['simple', 'Simple']];

// Fonts that ship with Windows or Office; each stack falls back to something close, so a
// missing font never breaks the layout.
const FONTS = [
    { id: 'theme', name: "Theme's own", stack: '' },
    { id: 'system', name: 'System default', stack: "'Segoe UI', system-ui, -apple-system, sans-serif" },
    { id: 'bahnschrift', name: 'Bahnschrift', stack: "Bahnschrift, 'Segoe UI', sans-serif" },
    { id: 'bahnsemi', name: 'Bahnschrift Semi-condensed', stack: "'Bahnschrift SemiCondensed', Bahnschrift, 'Segoe UI', sans-serif" },
    { id: 'franklin', name: 'Franklin Gothic', stack: "'Franklin Gothic Medium', 'Segoe UI Semibold', 'Segoe UI', sans-serif" },
    { id: 'trebuchet', name: 'Trebuchet MS', stack: "'Trebuchet MS', 'Segoe UI', sans-serif" },
    { id: 'verdana', name: 'Verdana', stack: "Verdana, 'Segoe UI', sans-serif" },
    { id: 'tahoma', name: 'Tahoma', stack: "Tahoma, 'Segoe UI', sans-serif" },
    { id: 'candara', name: 'Candara', stack: "Candara, 'Trebuchet MS', 'Segoe UI', sans-serif" },
    { id: 'corbel', name: 'Corbel', stack: "Corbel, Candara, 'Segoe UI', sans-serif" },
    { id: 'inkfree', name: 'Ink Free (handwriting)', stack: "'Ink Free', 'Segoe Print', 'Segoe Script', cursive" },
    { id: 'impact', name: 'Impact', stack: "Impact, 'Arial Black', 'Segoe UI', sans-serif" },
    { id: 'georgia', name: 'Georgia', stack: "Georgia, 'Times New Roman', serif" },
    { id: 'constantia', name: 'Constantia', stack: "Constantia, Georgia, 'Times New Roman', serif" },
    { id: 'cambria', name: 'Cambria', stack: "Cambria, Georgia, 'Times New Roman', serif" },
    { id: 'palatino', name: 'Palatino', stack: "'Palatino Linotype', Palatino, Georgia, serif" },
    { id: 'rockwell', name: 'Rockwell (slab)', stack: "Rockwell, 'Rockwell Nova', 'Palatino Linotype', Georgia, serif" },
    { id: 'consolas', name: 'Consolas', stack: "Consolas, 'Courier New', monospace" },
    { id: 'cascadia', name: 'Cascadia Mono', stack: "'Cascadia Mono', 'Cascadia Code', Consolas, monospace" },
    { id: 'courier', name: 'Courier New', stack: "'Courier New', Courier, monospace" },
];
const DEFAULT_FONT_SIZE = 13;

// The theme in force: main.js reads it when it draws meters and wording
let LABELS = PLAIN_LABELS;
let CURRENT_THEME = THEMES[DEFAULT_THEME];

function fontStack(id) {
    const f = FONTS.find(x => x.id === id);
    return f && f.stack ? f.stack : FONTS[1].stack;
}

// A meter as HTML. The same markup is used for the live rows, the limits and the theme previews.
function meterHtml(pct, theme) {
    const m = (theme || CURRENT_THEME).meter;
    const value = Math.max(0, Math.min(100, pct));
    if (m.kind === 'smooth') {
        return `<div class="meter smooth ${m.ticks ? 'ticks' : ''} ${m.glow ? 'glowing' : ''}"><div class="fill" style="width:${value}%"></div></div>`;
    }
    let lit = Math.round(value / 100 * m.n);
    if (value > 0 && lit === 0) lit = 1;
    let cells = '';
    for (let i = 0; i < m.n; i++) cells += `<i class="${i < lit ? 'on' : 'off'}"></i>`;
    return `<div class="meter segs" data-kind="${m.kind}">${cells}</div>`;
}

// Apply theme, accent, font and size from the saved settings
function applyAppearance(cfg) {
    const root = document.documentElement;
    const theme = THEMES[cfg.theme] || THEMES[DEFAULT_THEME];
    CURRENT_THEME = theme;
    for (const [key, value] of Object.entries(theme.vars)) {
        root.style.setProperty('--' + key, value);
    }
    if (cfg.accent && /^#[0-9a-fA-F]{6}$/.test(cfg.accent)) {
        root.style.setProperty('--accent', cfg.accent);
    }

    const look = theme.look;
    root.style.setProperty('--radius', look.radius + 'px');
    root.style.setProperty('--bw', look.bw + 'px');
    root.style.setProperty('--bg-image', look.bg);
    root.style.setProperty('--bg-size', look.bgSize);
    root.style.setProperty('--title-pre', JSON.stringify(look.pre));
    root.style.setProperty('--title-post', JSON.stringify(look.post));
    root.style.setProperty('--logo', JSON.stringify(look.logo));
    root.style.setProperty('--title-case', look.caps ? 'uppercase' : 'none');
    root.style.setProperty('--clip', look.clip);
    root.style.setProperty('--head-weight', look.headWeight);
    root.style.setProperty('--head-style', look.headStyle);
    root.style.setProperty('--head-spacing', look.headSpacing);
    root.dataset.glow = look.glow ? '1' : '0';

    // Your own font wins; otherwise the theme brings its body and heading fonts
    const own = cfg.font && cfg.font !== 'theme' ? cfg.font : null;
    root.style.setProperty('--font', fontStack(own || theme.font));
    root.style.setProperty('--font-head', fontStack(own || theme.headFont));
    root.style.setProperty('--fs', (cfg.fontSize || DEFAULT_FONT_SIZE) + 'px');

    LABELS = theme.labels;
    const ct = document.getElementById('contextTitle');
    if (ct) ct.textContent = LABELS.context;
    if (typeof lastSessionsHtml !== 'undefined') lastSessionsHtml = null; // words and meters changed, redraw
}
