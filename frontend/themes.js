// A theme is more than colors. Each one sets:
//   vars    - the color palette (CSS variables on :root)
//   look    - shape and decoration: corner radius, border width, bar style, background
//                pattern, heading decoration, logo glyph, glow
//   font    - a font id from FONTS, used unless you pick your own
//   labels  - the words for Context, Limits, Used, Free and Resets
// Everything in style.css reads these, so adding a theme is just adding an entry here.

const FLAT_LOOK = { radius: 8, bw: 1, bar: 'solid', bg: 'none', bgSize: 'auto', pre: '', post: '', logo: '', glow: false, caps: true };
const PLAIN_LABELS = { context: 'Context', limits: 'Limits', used: 'Used', free: 'Free', resets: 'Resets in' };

// Pattern helpers (inline SVG and gradients, so nothing is downloaded)
const gridBg = (rgba, size) => ({ bg: `linear-gradient(${rgba} 1px, transparent 1px), linear-gradient(90deg, ${rgba} 1px, transparent 1px)`, bgSize: `${size}px ${size}px` });
const dotBg = (rgba, size) => ({ bg: `radial-gradient(${rgba} 1.2px, transparent 1.6px)`, bgSize: `${size}px ${size}px` });
const scanBg = (rgba) => ({ bg: `repeating-linear-gradient(0deg, ${rgba} 0 1px, transparent 1px 3px)`, bgSize: 'auto' });
const emojiBg = (a, b, size) => {
    const svg = `<svg xmlns='http://www.w3.org/2000/svg' width='${size}' height='${size}'>` +
        `<text x='8' y='30' font-size='20' opacity='0.16'>${a}</text>` +
        `<text x='${size / 2 + 8}' y='${size - 14}' font-size='16' opacity='0.16'>${b}</text></svg>`;
    return { bg: `url("data:image/svg+xml,${encodeURIComponent(svg)}")`, bgSize: `${size}px ${size}px` };
};

const THEMES = {
    /* ----- Styles ----- */
    scifi: {
        name: 'Starship', tag: 'Bridge console', group: 'styles', icon: '◈', font: 'bahnschrift',
        vars: { bg: '#060b1a', panel: '#0b1630', panel2: '#08112a', border: '#18407a', track: '#0f2347', text: '#cfe8ff', dim: '#86b4e4', strong: '#eaf6ff', accent: '#29d9ff', good: '#3dffc4', warn: '#ffd24a', danger: '#ff8a3d', crit: '#ff3b5c' },
        look: { ...FLAT_LOOK, radius: 2, bar: 'blocks', ...gridBg('rgba(41,217,255,0.07)', 24), pre: '[ ', post: ' ]', logo: '◈ ', glow: true },
        labels: { ...PLAIN_LABELS, resets: 'Resets in' },
    },
    linux: {
        name: 'Terminal', tag: 'For the Linux geek', group: 'styles', icon: '>_', font: 'cascadia',
        vars: { bg: '#0a0e0a', panel: '#0d140d', panel2: '#0a110a', border: '#1d3a1d', track: '#142214', text: '#b6f5b6', dim: '#5fbf5f', strong: '#e6ffe6', accent: '#39ff14', good: '#39ff14', warn: '#e6ff4d', danger: '#ffb000', crit: '#ff4040' },
        look: { ...FLAT_LOOK, radius: 0, bar: 'blocks', ...scanBg('rgba(0,0,0,0.28)'), pre: '$ ', post: '', logo: '>_ ', caps: false },
        labels: { context: 'context', limits: 'limits', used: 'used', free: 'free', resets: 'resets in' },
    },
    bitcoin: {
        name: 'Satoshi', tag: 'Number go up', group: 'styles', icon: '₿', font: 'bahnschrift',
        vars: { bg: '#0d0d0d', panel: '#191919', panel2: '#141414', border: '#4a3414', track: '#2a2a2a', text: '#f2f2f2', dim: '#c2b08a', strong: '#ffffff', accent: '#f7931a', good: '#f7931a', warn: '#ffd166', danger: '#ff6b35', crit: '#ff3b3b' },
        look: { ...FLAT_LOOK, radius: 6, ...dotBg('rgba(247,147,26,0.12)', 26), logo: '₿ ', glow: true },
        labels: { ...PLAIN_LABELS, used: 'Spent', free: 'Left' },
    },
    ai: {
        name: 'Neural', tag: 'For the AI freak', group: 'styles', icon: '✦', font: 'cascadia',
        vars: { bg: '#0b0820', panel: '#15113a', panel2: '#110e30', border: '#3b2f8f', track: '#241c5e', text: '#e8e4ff', dim: '#aca3f2', strong: '#ffffff', accent: '#a78bfa', good: '#34d399', warn: '#fbbf24', danger: '#fb7185', crit: '#f43f5e' },
        look: { ...FLAT_LOOK, radius: 12, bar: 'glow', ...dotBg('rgba(167,139,250,0.16)', 28), pre: '✦ ', logo: '✦ ', glow: true },
        labels: { ...PLAIN_LABELS, context: 'Context window', used: 'Tokens used', free: 'Tokens free' },
    },
    cowboy: {
        name: 'Frontier', tag: 'Yeehaw', group: 'styles', icon: '★', font: 'palatino',
        vars: { bg: '#3b2616', panel: '#4a301c', panel2: '#43291a', border: '#8a5a2b', track: '#2e1d10', text: '#f3e2c3', dim: '#d2b184', strong: '#fff4dc', accent: '#e0a03a', good: '#9bc53d', warn: '#e0a03a', danger: '#d9622b', crit: '#c0392b' },
        look: { ...FLAT_LOOK, radius: 4, bw: 2, bar: 'rope', bg: 'repeating-linear-gradient(0deg, rgba(0,0,0,0.10) 0 2px, transparent 2px 7px)', pre: '★ ', post: ' ★', logo: '★ ' },
        labels: { ...PLAIN_LABELS, free: 'Left', resets: 'Back in' },
    },
    sweet: {
        name: 'Sweet', tag: 'Kawaii pastels', group: 'styles', icon: '🌸', font: 'candara',
        vars: { bg: '#fff0f6', panel: '#ffffff', panel2: '#fff7fb', border: '#ffb6d5', track: '#ffe0ee', text: '#6b3a55', dim: '#a0678a', strong: '#4a2040', accent: '#ff5fa2', good: '#3fc3a0', warn: '#ffa94d', danger: '#ff7aa2', crit: '#e8336d' },
        look: { ...FLAT_LOOK, radius: 18, ...emojiBg('🌸', '✨', 96), pre: '✿ ', post: ' ✿', logo: '🌸 ', caps: false },
        labels: { context: 'Context', limits: 'Limits', used: 'Used', free: 'Left', resets: 'Resets in' },
    },
    anime: {
        name: 'Shonen', tag: 'Anime energy', group: 'styles', icon: '⚡', font: 'bahnschrift',
        vars: { bg: '#101018', panel: '#1b1b2a', panel2: '#161622', border: '#ff3b3b', track: '#2a2a3c', text: '#ffffff', dim: '#ffd54a', strong: '#ffffff', accent: '#ff3b3b', good: '#3dd6ff', warn: '#ffd54a', danger: '#ff8a00', crit: '#ff1744' },
        look: { ...FLAT_LOOK, radius: 0, bw: 3, bar: 'stripes', bg: 'repeating-conic-gradient(from 0deg at 50% 50%, rgba(255,255,255,0.05) 0deg 2deg, transparent 2deg 6deg)', post: '!', logo: '⚡ ', glow: false },
        labels: { ...PLAIN_LABELS, used: 'Power used', free: 'Power left' },
    },
    arcade: {
        name: '8-bit', tag: 'Insert coin', group: 'styles', icon: '👾', font: 'consolas',
        vars: { bg: '#000000', panel: '#10102a', panel2: '#0a0a1e', border: '#5a5aff', track: '#202048', text: '#ffffff', dim: '#a3a8ff', strong: '#ffffff', accent: '#ffe600', good: '#00ff66', warn: '#ffe600', danger: '#ff8a00', crit: '#ff0044' },
        look: { ...FLAT_LOOK, radius: 0, bw: 2, bar: 'pixel', ...gridBg('rgba(90,90,255,0.10)', 16), pre: '▶ ', logo: '👾 ' },
        labels: { ...PLAIN_LABELS, used: 'Used', free: 'HP left', resets: 'Next round in' },
    },
    cyber: {
        name: 'Neon City', tag: 'Cyberpunk', group: 'styles', icon: '▮', font: 'bahnschrift',
        vars: { bg: '#0a0a0f', panel: '#14141c', panel2: '#101018', border: '#6b6500', track: '#23232e', text: '#f5f5f5', dim: '#7df9ff', strong: '#ffffff', accent: '#fcee0a', good: '#00f0ff', warn: '#fcee0a', danger: '#ff2a6d', crit: '#ff003c' },
        look: { ...FLAT_LOOK, radius: 0, bar: 'glow', bg: 'repeating-linear-gradient(135deg, rgba(252,238,10,0.05) 0 2px, transparent 2px 10px)', pre: '// ', logo: '▮ ', glow: true },
        labels: PLAIN_LABELS,
    },
    vapor: {
        name: 'Vaporwave', tag: 'Sunset grid', group: 'styles', icon: '🌴', font: 'bahnschrift',
        vars: { bg: '#1a1033', panel: '#2a1b4d', panel2: '#231544', border: '#4b2a85', track: '#3a2266', text: '#f5e9ff', dim: '#cdb6f5', strong: '#ffffff', accent: '#ff2fd0', good: '#00f0ff', warn: '#ffe66d', danger: '#ff7a59', crit: '#ff2e63' },
        look: { ...FLAT_LOOK, radius: 10, bar: 'glow', ...gridBg('rgba(255,47,208,0.10)', 28), logo: '🌴 ', glow: true },
        labels: PLAIN_LABELS,
    },

    /* ----- Classic color schemes ----- */
    midnight: { name: 'Midnight', tag: 'Deep blue', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#14161c', panel: '#1c1f27', panel2: '#181b22', border: '#2b3040', track: '#2b3040', text: '#e6e9f0', dim: '#aab1c3', strong: '#ffffff', accent: '#7aa2f7', good: '#4ade80', warn: '#fbbf24', danger: '#f87171', crit: '#ff4d4d' } },
    graphite: { name: 'Graphite', tag: 'Neutral dark', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#1a1a1a', panel: '#242424', panel2: '#202020', border: '#333333', track: '#333333', text: '#e0e0e0', dim: '#c0c0c0', strong: '#ffffff', accent: '#60a5fa', good: '#4ade80', warn: '#fbbf24', danger: '#f87171', crit: '#ff4444' } },
    light: { name: 'Daylight', tag: 'Clean light', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#f5f6f8', panel: '#ffffff', panel2: '#ffffff', border: '#d9dde5', track: '#e3e6ec', text: '#1f2430', dim: '#5b6475', strong: '#111827', accent: '#2563eb', good: '#16a34a', warn: '#d97706', danger: '#dc2626', crit: '#b91c1c' } },
    paper: { name: 'Paper', tag: 'Warm light', group: 'classic', icon: '', font: 'georgia', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#f4ecd8', panel: '#fbf6e9', panel2: '#f9f2e0', border: '#d8cba8', track: '#e4d9bb', text: '#3b3226', dim: '#6b5d44', strong: '#2a2118', accent: '#a0522d', good: '#4d7c0f', warn: '#b45309', danger: '#b91c1c', crit: '#991b1b' } },
    nord: { name: 'Nord', tag: 'Arctic', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#2e3440', panel: '#3b4252', panel2: '#353c4a', border: '#4c566a', track: '#4c566a', text: '#eceff4', dim: '#c3cad8', strong: '#ffffff', accent: '#88c0d0', good: '#a3be8c', warn: '#ebcb8b', danger: '#d08770', crit: '#bf616a' } },
    dracula: { name: 'Dracula', tag: 'Purple night', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#282a36', panel: '#343746', panel2: '#2f3241', border: '#44475a', track: '#44475a', text: '#f8f8f2', dim: '#c9cbe0', strong: '#ffffff', accent: '#bd93f9', good: '#50fa7b', warn: '#f1fa8c', danger: '#ffb86c', crit: '#ff5555' } },
    solarDark: { name: 'Solarized Dark', tag: 'Classic', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#002b36', panel: '#073642', panel2: '#04323d', border: '#0f4a58', track: '#0f4a58', text: '#eee8d5', dim: '#93a1a1', strong: '#fdf6e3', accent: '#b58900', good: '#859900', warn: '#b58900', danger: '#cb4b16', crit: '#dc322f' } },
    solarLight: { name: 'Solarized Light', tag: 'Classic', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#fdf6e3', panel: '#eee8d5', panel2: '#f3edda', border: '#d6cfb8', track: '#ddd6c1', text: '#586e75', dim: '#657b83', strong: '#073642', accent: '#268bd2', good: '#859900', warn: '#b58900', danger: '#cb4b16', crit: '#dc322f' } },
    gruvbox: { name: 'Gruvbox', tag: 'Retro warm', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#282828', panel: '#3c3836', panel2: '#32302f', border: '#504945', track: '#504945', text: '#ebdbb2', dim: '#bdae93', strong: '#fbf1c7', accent: '#fabd2f', good: '#b8bb26', warn: '#fabd2f', danger: '#fe8019', crit: '#fb4934' } },
    ocean: { name: 'Ocean', tag: 'Teal depths', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#0b1d2a', panel: '#112b3d', panel2: '#0e2433', border: '#1d4a66', track: '#1b3e55', text: '#dff3ff', dim: '#9cc4dc', strong: '#ffffff', accent: '#2ec4b6', good: '#5eead4', warn: '#ffd166', danger: '#ff9f6e', crit: '#ff5d73' } },
    sunset: { name: 'Sunset', tag: 'Warm glow', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#2a1420', panel: '#3a1c2c', panel2: '#33182a', border: '#5c2b44', track: '#4a2338', text: '#ffe9dc', dim: '#e0b9a6', strong: '#ffffff', accent: '#ff8c42', good: '#ffd166', warn: '#ff9f1c', danger: '#ff6b6b', crit: '#ff2e4d' } },
    contrast: { name: 'High contrast', tag: 'Accessibility', group: 'classic', icon: '', font: 'system', look: FLAT_LOOK, labels: PLAIN_LABELS, vars: { bg: '#000000', panel: '#000000', panel2: '#000000', border: '#ffffff', track: '#333333', text: '#ffffff', dim: '#dddddd', strong: '#ffffff', accent: '#ffff00', good: '#00ff00', warn: '#ffff00', danger: '#ff8c00', crit: '#ff0000' } },
};
const DEFAULT_THEME = 'midnight';

// Fonts that ship with Windows; each falls back to the system UI font if it is missing.
const FONTS = [
    { id: 'theme', name: "Theme's own", stack: '' },
    { id: 'system', name: 'System default', stack: "'Segoe UI', system-ui, -apple-system, sans-serif" },
    { id: 'bahnschrift', name: 'Bahnschrift', stack: "Bahnschrift, 'Segoe UI', sans-serif" },
    { id: 'trebuchet', name: 'Trebuchet MS', stack: "'Trebuchet MS', 'Segoe UI', sans-serif" },
    { id: 'verdana', name: 'Verdana', stack: "Verdana, 'Segoe UI', sans-serif" },
    { id: 'tahoma', name: 'Tahoma', stack: "Tahoma, 'Segoe UI', sans-serif" },
    { id: 'candara', name: 'Candara', stack: "Candara, 'Trebuchet MS', 'Segoe UI', sans-serif" },
    { id: 'georgia', name: 'Georgia', stack: "Georgia, 'Times New Roman', serif" },
    { id: 'palatino', name: 'Palatino', stack: "'Palatino Linotype', Palatino, Georgia, serif" },
    { id: 'consolas', name: 'Consolas', stack: "Consolas, 'Courier New', monospace" },
    { id: 'cascadia', name: 'Cascadia Mono', stack: "'Cascadia Mono', 'Cascadia Code', Consolas, monospace" },
    { id: 'courier', name: 'Courier New', stack: "'Courier New', Courier, monospace" },
];
const DEFAULT_FONT_SIZE = 13;

// The words the current theme uses; main.js reads these when it draws rows and limits
let LABELS = PLAIN_LABELS;

function fontStack(id) {
    const f = FONTS.find(x => x.id === id);
    return f && f.stack ? f.stack : FONTS[1].stack;
}

// Apply theme, accent, font and size from the saved settings
function applyAppearance(cfg) {
    const root = document.documentElement;
    const theme = THEMES[cfg.theme] || THEMES[DEFAULT_THEME];
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
    root.dataset.bar = look.bar;
    root.dataset.glow = look.glow ? '1' : '0';

    // Your own font wins; otherwise the theme brings one
    const own = cfg.font && cfg.font !== 'theme' ? cfg.font : theme.font;
    root.style.setProperty('--font', fontStack(own));
    root.style.setProperty('--fs', (cfg.fontSize || DEFAULT_FONT_SIZE) + 'px');

    LABELS = theme.labels;
    const ct = document.getElementById('contextTitle');
    if (ct) ct.textContent = LABELS.context;
    if (typeof lastSessionsHtml !== 'undefined') lastSessionsHtml = null; // words changed, redraw rows
}
