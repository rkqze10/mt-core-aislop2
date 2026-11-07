/**
 * MT Core Node.js / JavaScript Super User SDK
 * A high-level, fluent library for building MT Core plugins in JavaScript/TypeScript.
 */

const fs = require('fs');

const Color = {
  AMBER: '#FF9E3B',
  EMERALD: '#00FF88',
  CYBER: '#00F0FF',
  PINK: '#FF007F',
  PURPLE: '#A855F7',
  YELLOW: '#EAB308',
  RED: '#EF4444',
  GREEN: '#22C55E',
  BLUE: '#3B82F6',
  WHITE: '#FAFAFA',
  MUTED: '#71717A',
  DARK: '#18181B'
};

const Style = {
  RESET: '\x1b[0m',
  BOLD: '\x1b[1m',
  DIM: '\x1b[2m',
  ITALIC: '\x1b[3m',
  UNDERLINE: '\x1b[4m',

  hexFg(hex) {
    hex = hex.replace(/^#/, '');
    if (hex.length === 6) {
      const r = parseInt(hex.slice(0, 2), 16);
      const g = parseInt(hex.slice(2, 4), 16);
      const b = parseInt(hex.slice(4, 6), 16);
      return `\x1b[38;2;${r};${g};${b}m`;
    }
    return '';
  },

  hexBg(hex) {
    hex = hex.replace(/^#/, '');
    if (hex.length === 6) {
      const r = parseInt(hex.slice(0, 2), 16);
      const g = parseInt(hex.slice(2, 4), 16);
      const b = parseInt(hex.slice(4, 6), 16);
      return `\x1b[48;2;${r};${g};${b}m`;
    }
    return '';
  },

  color(text, fg, bg, bold = false) {
    let out = '';
    if (bold) out += this.BOLD;
    if (fg) out += this.hexFg(fg);
    if (bg) out += this.hexBg(bg);
    out += String(text) + this.RESET;
    return out;
  },

  badge(text, fg = Color.WHITE, bg = Color.DARK) {
    return `${this.hexBg(bg)}${this.hexFg(fg)} ${text} ${this.RESET}`;
  },

  progressBar(current, total, width = 24, fgDone = Color.EMERALD, fgTodo = Color.MUTED) {
    const safeWidth = Math.max(1, width);
    const ratio = Math.max(0.0, Math.min(1.0, total > 0 ? current / total : 0.0));
    const filled = Math.floor(ratio * safeWidth);
    const empty = safeWidth - filled;
    const bar = this.color('━'.repeat(filled), fgDone, null, true) + this.color('━'.repeat(empty), fgTodo);
    const pct = Math.floor(ratio * 100);
    return `${bar} ${this.color(pct + '%', Color.WHITE, null, true)}`;
  },

  card(title, items, color = Color.AMBER, width = 44) {
    const safeWidth = Math.max(10, width);
    const lines = [];
    const titlePadding = Math.max(0, safeWidth - title.length - 6);
    lines.push(`╭─ ${this.color(` ${title} `, color, null, true)}${'─'.repeat(titlePadding)}╮`);
    for (const [k, v] of Object.entries(items)) {
      const kFmt = this.color(`  ${k.padEnd(16)}: `, Color.MUTED);
      const vFmt = this.color(String(v), Color.WHITE, null, true);
      lines.push(kFmt + vFmt);
    }
    lines.push(`╰${'─'.repeat(Math.max(0, safeWidth - 2))}╯`);
    return lines.join('\n');
  }
};

class MT {
  constructor() {
    this.width = parseInt(process.env.MT_WIDTH || '80', 10);
    this.height = parseInt(process.env.MT_HEIGHT || '24', 10);
    this.theme = process.env.MT_THEME || 'amber';
    this.pluginsDir = process.env.MT_PLUGINS_DIR || './plugins';
    this.isDaemon = process.env.MT_DAEMON === '1';
  }

  _emit(action, params = {}) {
    const payload = Object.assign({ action }, params);
    process.stdout.write(JSON.stringify(payload) + '\n');
  }

  // --- UI & Feedback ---
  toast(message, isError = false) {
    this._emit('toast', { message, is_error: isError });
  }

  setStatus(message, isError = false) {
    this._emit('set_status', { message, is_error: isError });
  }

  setBanner(banner) {
    this._emit('set_banner', { banner });
  }

  resetBanner() {
    this._emit('reset_banner');
  }

  setFooter(footer) {
    this._emit('set_footer', { footer });
  }

  resetFooter() {
    this._emit('reset_footer');
  }

  setInput(value) {
    this._emit('set_input_value', { value });
  }

  setPlaceholder(placeholder) {
    this._emit('set_input_placeholder', { placeholder });
  }

  modal(title, content) {
    this._emit('modal', { title, content });
  }

  hijackScreen(content) {
    this._emit('hijack_screen', { content });
  }

  releaseScreen() {
    this._emit('release_screen');
  }

  // --- Commands & Dynamic Injection ---
  registerCommand(command, title = '', category = 'SUPERUSER', desc = '', execScript = '') {
    this._emit('register_cmd', {
      command,
      title: title || command.toUpperCase(),
      category,
      desc,
      exec: execScript
    });
  }

  unregisterCommand(command) {
    this._emit('unregister_cmd', { command });
  }

  triggerCommand(command) {
    this._emit('trigger_cmd', { command });
  }

  launchTUI(binary, args = [], name = '', prompt = '') {
    this._emit('launch_tui', {
      binary,
      args,
      name: name || binary,
      prompt
    });
  }

  // --- Tabs & Workspace ---
  switchTab(index) {
    this._emit('switch_tab', { index });
  }

  closeAllTabs() {
    this._emit('close_all_tabs');
  }

  closeTab(index) {
    this._emit('force_close_tab', { index });
  }

  injectTab(title, cmd, output) {
    this._emit('inject_tab', { title, cmd, output });
  }

  updateTab(index, output) {
    this._emit('update_tab', { index, output });
  }

  setTabTitle(index, title) {
    this._emit('set_tab_title', { index, title });
  }

  scrollTab(index = -1, delta = 3) {
    this._emit('scroll_tab', { index, delta });
  }

  clearTab(index = -1) {
    this._emit('clear_tab', { index });
  }

  // --- Themes ---
  setTheme(name) {
    this._emit('theme', { name });
  }

  setThemeColor(field, hexColor) {
    this._emit('set_theme_color', { field, color: hexColor });
  }

  // --- 2D Pixel Framebuffer & Virtual Canvas ---
  pixel(x, y, char = '█', fg = '', bg = '') {
    this._emit('pixel', { x, y, char, fg, bg });
  }

  box(x, y, w, h, title = '', fg = '') {
    this._emit('box', { x, y, w, h, title, fg });
  }

  clearPixels() {
    this._emit('clear_pixels');
  }

  // --- Streaming & Real-Time Output ---
  stream(chunk) {
    this._emit('stream_append', { chunk });
  }

  streamAppend(chunk) {
    this._emit('stream_append', { chunk });
  }

  // --- Event Bus & Pub/Sub ---
  emitEvent(eventName, payload = null) {
    this._emit('emit_event', { event: eventName, data: payload });
  }

  // --- Rich Graphics & Visual Widgets ---
  renderTable(headers, rows) {
    this._emit('table', { headers, rows });
  }

  renderSparkline(values, title = '', color = '') {
    this._emit('sparkline', { values, title, color });
  }

  renderBarChart(data, title = '', color = '') {
    this._emit('barchart', { data, title, color });
  }

  // --- State & Persistence ---
  kvSet(key, value) {
    this._emit('kv_set', { key, value: String(value) });
    const cleanKey = 'MT_KV_' + key.replace(/[^a-zA-Z0-9]/g, '_').toUpperCase();
    process.env[cleanKey] = String(value);
  }

  kvGet(key, defaultValue = '') {
    const cleanKey = 'MT_KV_' + key.replace(/[^a-zA-Z0-9]/g, '_').toUpperCase();
    return process.env[cleanKey] !== undefined ? process.env[cleanKey] : defaultValue;
  }

  kvDelete(key) {
    this._emit('kv_del', { key });
    const cleanKey = 'MT_KV_' + key.replace(/[^a-zA-Z0-9]/g, '_').toUpperCase();
    delete process.env[cleanKey];
  }

  // --- Pipe Input ---
  getPipeInput() {
    try {
      if (!process.stdin.isTTY) {
        return fs.readFileSync(0, 'utf-8');
      }
    } catch (e) {
      // ignore
    }
    return process.env.MT_PIPE_INPUT || '';
  }
}

module.exports = { MT, Color, Style };
