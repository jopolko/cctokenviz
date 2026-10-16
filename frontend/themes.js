// Themes are sets of CSS variables applied to :root. Everything in style.css reads these,
// so adding a theme is just adding an entry here.
const THEMES = {
    midnight: { name: 'Midnight', vars: { bg: '#14161c', panel: '#1c1f27', panel2: '#181b22', border: '#2b3040', track: '#2b3040', text: '#e6e9f0', dim: '#aab1c3', strong: '#ffffff', accent: '#7aa2f7', good: '#4ade80', warn: '#fbbf24', danger: '#f87171', crit: '#ff4d4d' } },
    graphite: { name: 'Graphite', vars: { bg: '#1a1a1a', panel: '#242424', panel2: '#202020', border: '#333333', track: '#333333', text: '#e0e0e0', dim: '#c0c0c0', strong: '#ffffff', accent: '#60a5fa', good: '#4ade80', warn: '#fbbf24', danger: '#f87171', crit: '#ff4444' } },
    light: { name: 'Daylight', vars: { bg: '#f5f6f8', panel: '#ffffff', panel2: '#ffffff', border: '#d9dde5', track: '#e3e6ec', text: '#1f2430', dim: '#5b6475', strong: '#111827', accent: '#2563eb', good: '#16a34a', warn: '#d97706', danger: '#dc2626', crit: '#b91c1c' } },
    paper: { name: 'Paper', vars: { bg: '#f4ecd8', panel: '#fbf6e9', panel2: '#f9f2e0', border: '#d8cba8', track: '#e4d9bb', text: '#3b3226', dim: '#6b5d44', strong: '#2a2118', accent: '#a0522d', good: '#4d7c0f', warn: '#b45309', danger: '#b91c1c', crit: '#991b1b' } },
    nord: { name: 'Nord', vars: { bg: '#2e3440', panel: '#3b4252', panel2: '#353c4a', border: '#4c566a', track: '#4c566a', text: '#eceff4', dim: '#c3cad8', strong: '#ffffff', accent: '#88c0d0', good: '#a3be8c', warn: '#ebcb8b', danger: '#d08770', crit: '#bf616a' } },
    dracula: { name: 'Dracula', vars: { bg: '#282a36', panel: '#343746', panel2: '#2f3241', border: '#44475a', track: '#44475a', text: '#f8f8f2', dim: '#c9cbe0', strong: '#ffffff', accent: '#bd93f9', good: '#50fa7b', warn: '#f1fa8c', danger: '#ffb86c', crit: '#ff5555' } },
    solarDark: { name: 'Solarized Dark', vars: { bg: '#002b36', panel: '#073642', panel2: '#04323d', border: '#0f4a58', track: '#0f4a58', text: '#eee8d5', dim: '#93a1a1', strong: '#fdf6e3', accent: '#b58900', good: '#859900', warn: '#b58900', danger: '#cb4b16', crit: '#dc322f' } },
    solarLight: { name: 'Solarized Light', vars: { bg: '#fdf6e3', panel: '#eee8d5', panel2: '#f3edda', border: '#d6cfb8', track: '#ddd6c1', text: '#586e75', dim: '#657b83', strong: '#073642', accent: '#268bd2', good: '#859900', warn: '#b58900', danger: '#cb4b16', crit: '#dc322f' } },
    gruvbox: { name: 'Gruvbox', vars: { bg: '#282828', panel: '#3c3836', panel2: '#32302f', border: '#504945', track: '#504945', text: '#ebdbb2', dim: '#bdae93', strong: '#fbf1c7', accent: '#fabd2f', good: '#b8bb26', warn: '#fabd2f', danger: '#fe8019', crit: '#fb4934' } },
    synthwave: { name: 'Synthwave', vars: { bg: '#1a1033', panel: '#2a1b4d', panel2: '#231544', border: '#4b2a85', track: '#3a2266', text: '#f5e9ff', dim: '#c8aef0', strong: '#ffffff', accent: '#ff2fd0', good: '#00f0ff', warn: '#ffe66d', danger: '#ff7a59', crit: '#ff2e63' } },
    terminal: { name: 'Terminal', vars: { bg: '#000000', panel: '#04110a', panel2: '#020a06', border: '#0b3a1d', track: '#0b2a17', text: '#33ff7a', dim: '#1fbf5a', strong: '#7dffa8', accent: '#33ff7a', good: '#33ff7a', warn: '#e6ff4d', danger: '#ff9a3d', crit: '#ff3d3d' } },
    ocean: { name: 'Ocean', vars: { bg: '#0b1d2a', panel: '#112b3d', panel2: '#0e2433', border: '#1d4a66', track: '#1b3e55', text: '#dff3ff', dim: '#9cc4dc', strong: '#ffffff', accent: '#2ec4b6', good: '#5eead4', warn: '#ffd166', danger: '#ff9f6e', crit: '#ff5d73' } },
    sunset: { name: 'Sunset', vars: { bg: '#2a1420', panel: '#3a1c2c', panel2: '#33182a', border: '#5c2b44', track: '#4a2338', text: '#ffe9dc', dim: '#e0b9a6', strong: '#ffffff', accent: '#ff8c42', good: '#ffd166', warn: '#ff9f1c', danger: '#ff6b6b', crit: '#ff2e4d' } },
    contrast: { name: 'High contrast', vars: { bg: '#000000', panel: '#000000', panel2: '#000000', border: '#ffffff', track: '#333333', text: '#ffffff', dim: '#dddddd', strong: '#ffffff', accent: '#ffff00', good: '#00ff00', warn: '#ffff00', danger: '#ff8c00', crit: '#ff0000' } },
};
const DEFAULT_THEME = 'midnight';

// Fonts that ship with Windows; the list falls back to the system UI font if one is missing.
const FONTS = [
    { id: 'system', name: 'System default', stack: "'Segoe UI', system-ui, -apple-system, sans-serif" },
    { id: 'bahnschrift', name: 'Bahnschrift', stack: "Bahnschrift, 'Segoe UI', sans-serif" },
    { id: 'trebuchet', name: 'Trebuchet MS', stack: "'Trebuchet MS', 'Segoe UI', sans-serif" },
    { id: 'verdana', name: 'Verdana', stack: "Verdana, 'Segoe UI', sans-serif" },
    { id: 'tahoma', name: 'Tahoma', stack: "Tahoma, 'Segoe UI', sans-serif" },
    { id: 'candara', name: 'Candara', stack: "Candara, 'Segoe UI', sans-serif" },
    { id: 'georgia', name: 'Georgia', stack: "Georgia, 'Times New Roman', serif" },
    { id: 'palatino', name: 'Palatino', stack: "'Palatino Linotype', Palatino, Georgia, serif" },
    { id: 'consolas', name: 'Consolas', stack: "Consolas, 'Courier New', monospace" },
    { id: 'cascadia', name: 'Cascadia Mono', stack: "'Cascadia Mono', 'Cascadia Code', Consolas, monospace" },
    { id: 'courier', name: 'Courier New', stack: "'Courier New', Courier, monospace" },
];
const DEFAULT_FONT = 'system';
const DEFAULT_FONT_SIZE = 13;

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
    const font = FONTS.find(f => f.id === cfg.font) || FONTS[0];
    root.style.setProperty('--font', font.stack);
    root.style.setProperty('--fs', (cfg.fontSize || DEFAULT_FONT_SIZE) + 'px');
}
