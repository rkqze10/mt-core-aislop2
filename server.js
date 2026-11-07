const http = require('http');
const { exec, spawn } = require('child_process');
const fs = require('fs');
const path = require('path');
const url = require('url');

const PORT = 3000;
const ROOT_DIR = __dirname;
const BIN_DIR = path.join(ROOT_DIR, '.mt_bin');
const MTFIX_BIN = path.join(BIN_DIR, 'mtfix');

let godModeActive = false;
let currentBanner = `  ██████╗  ██████╗ ██████╗     ███╗   ███╗ ██████╗ ██████╗ ███████╗
 ██╔════╝ ██╔═══██╗██╔══██╗    ████╗ ████║██╔═══██╗██╔══██╗██╔════╝
 ██║  ███╗██║   ██║██║  ██║    ██╔████╔██║██║   ██║██║  ██║█████╗  
 ██║   ██║██║   ██║██║  ██║    ██║╚██╔╝██║██║   ██║██║  ██║██╔══╝  
 ╚██████╔╝╚██████╔╝██████╔╝    ██║ ╚═╝ ██║╚██████╔╝██████╔╝███████╗
  ╚═════╝  ╚═════╝ ╚═════╝     ╚═╝     ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝
    ⚡ GOD MODE v9999 & MTFIX // SÜPER KULLANICI ÇEKİRDEK MUTATÖRÜ ⚡`;

let currentPrompt = { prefix: '⚡ GOD-MODE ', arrow: '❯❯ ', color: '#00FFCC' };
let currentTheme = 'cyber';

// Ensure .mt_bin exists and mtfix is built
function ensureMtfixBinary(cb) {
  if (fs.existsSync(MTFIX_BIN)) {
    return cb(null, MTFIX_BIN);
  }
  fs.mkdirSync(BIN_DIR, { recursive: true });
  exec(`go build -o "${MTFIX_BIN}" plugins/mtfix.go`, { cwd: ROOT_DIR }, (err) => {
    if (err) {
      console.error('mtfix build error:', err);
      return cb(null, 'go run plugins/mtfix.go');
    }
    cb(null, MTFIX_BIN);
  });
}

function ansiToHtml(text) {
  if (!text) return '';
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/\x1b\[1m/g, '<span style="font-weight:bold;">')
    .replace(/\x1b\[31m/g, '<span style="color:#FF5555;">')
    .replace(/\x1b\[32m/g, '<span style="color:#50FA7B;">')
    .replace(/\x1b\[33m/g, '<span style="color:#F1FA8C;">')
    .replace(/\x1b\[34m/g, '<span style="color:#BD93F9;">')
    .replace(/\x1b\[35m/g, '<span style="color:#FF79C6;">')
    .replace(/\x1b\[36m/g, '<span style="color:#8BE9FD;">')
    .replace(/\x1b\[0m/g, '</span>')
    .replace(/\x1b\[\d+;?\d*m/g, ''); // strip remaining ANSI sequences
}

const HTML_PAGE = `<!DOCTYPE html>
<html lang="tr">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>MT Core Hypervisor & MTfix God Mode</title>
  <style>
    :root {
      --bg: #0b0d13;
      --bg-panel: #131722;
      --bg-input: #080a0f;
      --border: #2d3748;
      --primary: #00ffcc;
      --secondary: #ff007f;
      --accent: #7928ca;
      --text: #e2e8f0;
      --text-muted: #718096;
      --font-mono: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background-color: var(--bg);
      color: var(--text);
      font-family: var(--font-mono);
      font-size: 13px;
      line-height: 1.5;
      height: 100vh;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }
    header {
      background: var(--bg-panel);
      border-bottom: 1px solid var(--border);
      padding: 8px 16px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 12px;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: bold;
      color: var(--primary);
      text-transform: uppercase;
      letter-spacing: 1px;
    }
    .badge {
      background: #1f1a3a;
      border: 1px solid var(--accent);
      color: var(--primary);
      font-size: 11px;
      padding: 2px 8px;
      border-radius: 4px;
      font-weight: 600;
    }
    .badge-god {
      background: rgba(255, 0, 127, 0.2);
      border: 1px solid var(--secondary);
      color: #ff5599;
      animation: pulse 2s infinite;
    }
    @keyframes pulse {
      0%, 100% { opacity: 1; filter: drop-shadow(0 0 4px var(--secondary)); }
      50% { opacity: 0.7; }
    }
    .quick-actions {
      display: flex;
      gap: 6px;
      flex-wrap: wrap;
    }
    button {
      background: #1a202c;
      color: #cbd5e0;
      border: 1px solid var(--border);
      padding: 5px 10px;
      border-radius: 4px;
      font-family: inherit;
      font-size: 11px;
      font-weight: 600;
      cursor: pointer;
      display: inline-flex;
      align-items: center;
      gap: 4px;
      transition: all 0.15s ease;
    }
    button:hover {
      background: #2d3748;
      border-color: var(--primary);
      color: var(--primary);
    }
    button.btn-god {
      background: linear-gradient(135deg, #ff007f, #7928ca);
      border: 1px solid #ff5599;
      color: #ffffff;
      font-weight: bold;
      box-shadow: 0 0 12px rgba(255, 0, 127, 0.4);
    }
    button.btn-god:hover {
      background: linear-gradient(135deg, #ff1a8c, #8b3dd8);
      box-shadow: 0 0 18px rgba(255, 0, 127, 0.7);
    }
    .main-layout {
      flex: 1;
      display: flex;
      overflow: hidden;
    }
    .terminal-container {
      flex: 3;
      display: flex;
      flex-direction: column;
      background: var(--bg);
      border-right: 1px solid var(--border);
      overflow: hidden;
    }
    .terminal-output {
      flex: 1;
      padding: 16px;
      overflow-y: auto;
      white-space: pre-wrap;
      word-break: break-all;
      font-family: var(--font-mono);
      font-size: 13px;
    }
    .terminal-output .banner-text {
      color: var(--secondary);
      font-weight: bold;
      margin-bottom: 12px;
      text-shadow: 0 0 8px rgba(255, 0, 127, 0.3);
    }
    .terminal-input-bar {
      background: var(--bg-input);
      border-top: 1px solid var(--border);
      padding: 8px 12px;
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .prompt-label {
      color: var(--secondary);
      font-weight: bold;
      white-space: nowrap;
    }
    .prompt-arrow {
      color: var(--primary);
      font-weight: bold;
    }
    #cmdInput {
      flex: 1;
      background: transparent;
      border: none;
      outline: none;
      color: #ffffff;
      font-family: inherit;
      font-size: 13px;
    }
    .sidebar {
      flex: 2;
      background: var(--bg-panel);
      display: flex;
      flex-direction: column;
      overflow: hidden;
      min-width: 340px;
      max-width: 480px;
    }
    .sidebar-header {
      padding: 10px 14px;
      border-bottom: 1px solid var(--border);
      font-weight: bold;
      color: var(--primary);
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
    .sidebar-content {
      flex: 1;
      overflow-y: auto;
      padding: 14px;
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    .card {
      background: #0f131c;
      border: 1px solid var(--border);
      border-radius: 6px;
      padding: 12px;
    }
    .card-title {
      font-weight: bold;
      color: var(--secondary);
      margin-bottom: 8px;
      font-size: 12px;
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .field-group {
      margin-bottom: 10px;
    }
    .field-label {
      font-size: 11px;
      color: var(--text-muted);
      margin-bottom: 4px;
      display: block;
    }
    select, input[type="text"], textarea {
      width: 100%;
      background: #080a0f;
      border: 1px solid var(--border);
      color: #e2e8f0;
      padding: 6px 8px;
      border-radius: 4px;
      font-family: inherit;
      font-size: 12px;
      outline: none;
    }
    select:focus, input[type="text"]:focus, textarea:focus {
      border-color: var(--primary);
      box-shadow: 0 0 6px rgba(0, 255, 204, 0.2);
    }
    textarea {
      resize: vertical;
      min-height: 60px;
    }
    .toast {
      position: fixed;
      top: 50px;
      right: 20px;
      background: #1a202c;
      border: 1px solid var(--secondary);
      color: #fff;
      padding: 10px 16px;
      border-radius: 6px;
      box-shadow: 0 4px 16px rgba(0,0,0,0.5);
      z-index: 1000;
      display: none;
      animation: fadeIn 0.2s ease;
    }
    @keyframes fadeIn { from { opacity: 0; transform: translateY(-10px); } to { opacity: 1; transform: translateY(0); } }
    .status-bar {
      background: var(--bg-panel);
      border-top: 1px solid var(--border);
      padding: 4px 14px;
      font-size: 11px;
      color: var(--text-muted);
      display: flex;
      justify-content: space-between;
    }
  </style>
</head>
<body>
  <div id="toast" class="toast"></div>

  <header>
    <div class="brand">
      <span>● MT CORE HYPERVISOR</span>
      <span class="badge" id="themeBadge">CYBER</span>
      <span class="badge badge-god" id="godBadge">ROOT-0 / GOD MODE</span>
    </div>
    <div class="quick-actions">
      <button class="btn-god" onclick="runCmd('/mtfix')">⚡ GOD MODE'U AÇ (MTFIX)</button>
      <button onclick="runCmd('/mtfix list')">📂 Dosyalar</button>
      <button onclick="runCmd('/mtfix view main.go')">🔍 main.go</button>
      <button onclick="runCmd('/god')">📊 Durum</button>
      <button onclick="runCmd('/mem dump')">🧠 Bellek</button>
      <button onclick="runCmd('/matrix')">🌧️ Matrix</button>
      <button onclick="runCmd('/mutate theme cyber')">🎨 Cyber</button>
      <button onclick="runCmd('/mtfix build')">🔨 Derle</button>
      <button onclick="runCmd('/mtfix reset')">🔄 Sıfırla</button>
    </div>
  </header>

  <div class="main-layout">
    <div class="terminal-container">
      <div class="terminal-output" id="termOutput"></div>
      <div class="terminal-input-bar">
        <span class="prompt-label" id="promptPrefix">⚡ GOD-MODE</span>
        <span class="prompt-arrow">❯❯</span>
        <input type="text" id="cmdInput" placeholder="komut yazın (/mtfix, /god, /mutate, /mtfix patch main.go, /help)..." autofocus autocomplete="off" />
        <button onclick="submitCmd()">Çalıştır ↵</button>
      </div>
    </div>

    <div class="sidebar">
      <div class="sidebar-header">
        <span>⚡ MTFIX CANLI KOD DEĞİŞTİRİCİ</span>
        <span style="font-size:11px; color:#a0aec0;">[OTO GERİ ALMA GÜVENLİĞİ]</span>
      </div>
      <div class="sidebar-content">
        <div class="card">
          <div class="card-title">🛠️ Hızlı Kod Yamalayıcı (Hot Patch)</div>
          <div class="field-group">
            <label class="field-label">Hedef Dosya</label>
            <select id="patchFile">
              <option value="main.go">main.go (Ana Başlatıcı)</option>
              <option value="pkg/ui/view.go">pkg/ui/view.go (Arayüz Görüntüleyici)</option>
              <option value="pkg/ui/model.go">pkg/ui/model.go (Çekirdek Modeli)</option>
              <option value="pkg/ui/update.go">pkg/ui/update.go (Olay İşleyici)</option>
              <option value="pkg/kernel/engine.go">pkg/kernel/engine.go (Kernel Motoru)</option>
              <option value="pkg/superuser/ipc.go">pkg/superuser/ipc.go (IPC Dağıtıcı)</option>
              <option value="plugins/mtfix.go">plugins/mtfix.go (MTfix Motoru)</option>
            </select>
          </div>
          <div class="field-group">
            <label class="field-label">Değiştirilecek Eski Metin / Kod</label>
            <textarea id="patchOld" placeholder="Dosyada bulunan orijinal kod parçası..."></textarea>
          </div>
          <div class="field-group">
            <label class="field-label">Yeni Kod Parçası</label>
            <textarea id="patchNew" placeholder="Yerine yazılacak yeni kod..."></textarea>
          </div>
          <button class="btn-god" style="width:100%; justify-content:center;" onclick="applyHotPatch()">
            ⚡ Yamayı Uygula ve Go Derlemesini Sına
          </button>
        </div>

        <div class="card">
          <div class="card-title">🎨 Canlı Arayüz Mutasyonu</div>
          <div class="field-group">
            <label class="field-label">Prompt Ön Eki</label>
            <input type="text" id="mutPrompt" value="⚡ GOD-MODE " />
          </div>
          <div class="field-group">
            <label class="field-label">Ana Tema Rengi (Primary Hex)</label>
            <input type="text" id="mutColor" value="#00FFCC" />
          </div>
          <button style="width:100%; justify-content:center;" onclick="applyUIMutation()">
            ✔ Arayüzü Canlıda Güncelle
          </button>
        </div>

        <div class="card">
          <div class="card-title">📖 God Mode Kılavuzu</div>
          <p style="font-size:11px; color:var(--text-muted); line-height:1.6;">
            • <b>/mtfix patch &lt;f&gt; &quot;a&quot; &quot;b&quot;</b>: Ana kodu değiştirir.<br/>
            • <b>/mtfix view &lt;f&gt;</b>: Dosyayı satır satır inceler.<br/>
            • <b>/mtfix revert &lt;f&gt;</b>: .bak yedeğinden geri yükler.<br/>
            • <b>/mutate prompt|banner|theme</b>: Arayüzü boyar.<br/>
            • <b>/patchcmd &lt;hedef&gt; &lt;yeni&gt;</b>: Komut yönlendirir.<br/>
            • <b>/mem set &lt;k&gt; &lt;v&gt;</b>: Çekirdek belleğine yazar.<br/>
            • <b>/matrix</b>: Framebuffer üzerine yağmur yağdırır.
          </p>
        </div>
      </div>
    </div>
  </div>

  <div class="status-bar">
    <span id="statusText">⚡ MT Core v12.0 hazır. God Mode devrede. Komut için '/' yazın.</span>
    <span>Çekirdek: Go 1.19 + Hypervisor Ring-0 | Port 3000</span>
  </div>

  <script>
    const term = document.getElementById('termOutput');
    const cmdInput = document.getElementById('cmdInput');
    const toast = document.getElementById('toast');
    let history = [];
    let histIdx = -1;

    function showToast(msg) {
      toast.innerText = msg;
      toast.style.display = 'block';
      setTimeout(() => { toast.style.display = 'none'; }, 3000);
    }

    function appendOutput(rawText, isCmd = false) {
      const div = document.createElement('div');
      div.style.marginBottom = '6px';
      if (isCmd) {
        div.innerHTML = '<span style="color:#ff007f; font-weight:bold;">⚡ GOD-MODE ❯❯</span> <span style="color:#fff; font-weight:bold;">' + rawText + '</span>';
      } else {
        div.innerHTML = rawText;
      }
      term.appendChild(div);
      term.scrollTop = term.scrollHeight;
    }

    async function runCmd(cmd) {
      if (!cmd) return;
      appendOutput(cmd, true);
      history.push(cmd);
      histIdx = history.length;
      try {
        const res = await fetch('/api/exec', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ command: cmd })
        });
        const data = await res.json();
        if (data.output) {
          appendOutput(data.output);
        }
        if (data.toast) {
          showToast(data.toast);
        }
        if (data.prompt) {
          document.getElementById('promptPrefix').innerText = data.prompt;
        }
        if (data.status) {
          document.getElementById('statusText').innerText = data.status;
        }
      } catch (err) {
        appendOutput('<span style="color:#ff5555;">Hata: ' + err.message + '</span>');
      }
    }

    function submitCmd() {
      const val = cmdInput.value.trim();
      if (!val) return;
      cmdInput.value = '';
      runCmd(val);
    }

    cmdInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        submitCmd();
      } else if (e.key === 'ArrowUp') {
        if (history.length > 0 && histIdx > 0) {
          histIdx--;
          cmdInput.value = history[histIdx];
        }
      } else if (e.key === 'ArrowDown') {
        if (histIdx < history.length - 1) {
          histIdx++;
          cmdInput.value = history[histIdx];
        } else {
          histIdx = history.length;
          cmdInput.value = '';
        }
      }
    });

    async function applyHotPatch() {
      const file = document.getElementById('patchFile').value;
      const oldText = document.getElementById('patchOld').value;
      const newText = document.getElementById('patchNew').value;
      if (!oldText || !newText) {
        showToast('⚠️ Eski ve yeni metin alanları zorunludur!');
        return;
      }
      runCmd('/mtfix patch ' + file + ' "' + oldText.replace(/"/g, '\\"') + '" "' + newText.replace(/"/g, '\\"') + '"');
    }

    function applyUIMutation() {
      const prompt = document.getElementById('mutPrompt').value;
      const color = document.getElementById('mutColor').value;
      runCmd('/mutate prompt "' + prompt + '" "❯❯ "');
      runCmd('/mutate color primary ' + color);
    }

    // Auto-launch God Mode on initial open
    window.onload = () => {
      runCmd('/mtfix');
    };
  </script>
</body>
</html>`;

function executeCommand(cmdStr, cb) {
  const parts = cmdStr.trim().split(/\s+/);
  let mainCmd = parts[0];
  if (mainCmd.startsWith('/')) {
    mainCmd = mainCmd.substring(1);
  }

  ensureMtfixBinary((err, mtfixExec) => {
    let child;
    const env = Object.assign({}, process.env, {
      MT_SUPERUSER: '1',
      MT_WIDTH: '120',
      MT_HEIGHT: '35',
      MT_THEME: currentTheme
    });

    if (mainCmd === 'mtfix' || mainCmd === 'god' || mainCmd === 'godmode' || mainCmd === 'fix' || mainCmd === 'patch') {
      const args = parts.slice(1);
      child = spawn(mtfixExec, args, { cwd: ROOT_DIR, env });
    } else if (mainCmd === 'mutate') {
      const args = ['mutate'].concat(parts.slice(1));
      child = spawn(mtfixExec, args, { cwd: ROOT_DIR, env });
    } else if (mainCmd === 'patchcmd') {
      const args = ['patchcmd'].concat(parts.slice(1));
      child = spawn(mtfixExec, args, { cwd: ROOT_DIR, env });
    } else if (mainCmd === 'source') {
      const args = ['view'].concat(parts.slice(1));
      child = spawn(mtfixExec, args, { cwd: ROOT_DIR, env });
    } else if (mainCmd === 'mem') {
      const args = ['mem'].concat(parts.slice(1));
      child = spawn(mtfixExec, args, { cwd: ROOT_DIR, env });
    } else if (mainCmd === 'matrix') {
      const args = ['matrix'];
      child = spawn(mtfixExec, args, { cwd: ROOT_DIR, env });
    } else if (mainCmd === 'godhelp') {
      const args = ['help'];
      child = spawn(mtfixExec, args, { cwd: ROOT_DIR, env });
    } else if (mainCmd === 'help') {
      child = spawn('go', ['run', 'main.go'], { cwd: ROOT_DIR, env });
    } else {
      // Shell fallback
      const rawSh = cmdStr.startsWith('/') ? cmdStr.substring(1) : cmdStr;
      child = spawn('bash', ['-c', rawSh], { cwd: ROOT_DIR, env });
    }

    let stdout = '';
    let stderr = '';
    child.stdout.on('data', d => { stdout += d.toString(); });
    child.stderr.on('data', d => { stderr += d.toString(); });

    child.on('close', code => {
      let toastMsg = null;
      let promptVal = null;
      let statusVal = null;
      let cleanOutputLines = [];

      const lines = (stdout + (stderr ? '\n' + stderr : '')).split('\n');
      for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed.startsWith('{') && trimmed.endsWith('}')) {
          try {
            const parsed = JSON.parse(trimmed);
            if (parsed.action === 'toast' && parsed.message) {
              toastMsg = parsed.message;
            } else if (parsed.action === 'set_prompt' && parsed.prefix) {
              promptVal = parsed.prefix;
            } else if (parsed.action === 'set_status' && parsed.message) {
              statusVal = parsed.message;
            } else if (parsed.action === 'theme' && parsed.name) {
              currentTheme = parsed.name;
            }
          } catch (e) {}
        } else {
          cleanOutputLines.push(line);
        }
      }

      const htmlFormatted = ansiToHtml(cleanOutputLines.join('\n'));
      cb(null, {
        output: htmlFormatted,
        toast: toastMsg,
        prompt: promptVal,
        status: statusVal,
        exitCode: code
      });
    });
  });
}

const server = http.createServer((req, res) => {
  const parsedUrl = url.parse(req.url, true);

  if (req.method === 'GET' && parsedUrl.pathname === '/') {
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(HTML_PAGE);
    return;
  }

  if (req.method === 'GET' && parsedUrl.pathname === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ status: 'ok', godMode: true }));
    return;
  }

  if (req.method === 'POST' && parsedUrl.pathname === '/api/exec') {
    let body = '';
    req.on('data', chunk => { body += chunk; });
    req.on('end', () => {
      try {
        const payload = JSON.parse(body || '{}');
        const cmd = payload.command || '/mtfix';
        executeCommand(cmd, (err, result) => {
          res.writeHead(200, { 'Content-Type': 'application/json; charset=utf-8' });
          res.end(JSON.stringify(result));
        });
      } catch (e) {
        res.writeHead(400, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ error: e.message }));
      }
    });
    return;
  }

  res.writeHead(404, { 'Content-Type': 'text/plain' });
  res.end('Not Found');
});

server.listen(PORT, '0.0.0.0', () => {
  console.log(`⚡ MT Core Hypervisor & MTfix God Mode web server listening on http://0.0.0.0:${PORT}`);
  // Ensure mtfix binary is ready
  ensureMtfixBinary((err, p) => {
    console.log(`✔ MTfix binary ready at ${p}`);
  });
});
