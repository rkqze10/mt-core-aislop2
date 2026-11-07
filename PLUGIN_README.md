# 🧠 MT Core Plugin Development Guide for AI Agents & Developers
> **Target Audience:** Autonomous AI Agents (Claude, GPT, Gemini), LLM Code Assistants, and Human Core Developers.  
> **Repository:** `rkqze10/mt-core-ai-`  
> **Architecture:** Go TUI Kernel (Bubble Tea / Lip Gloss) + Dynamic Multi-Language IPC Hypervisor (Super User x1000).

---

## 📌 Executive Summary

MT Core is not an ordinary shell: it is an **advanced, modular terminal hypervisor and operating environment** featuring a ring-0 level superuser host. External plugins written in **Python**, **Node.js / JavaScript**, **Bash/Shell**, **Go Source**, or **Native Binaries** can run seamlessly inside MT Core, inject new shell commands at runtime, directly manipulate terminal framebuffers, mutate UI themes, manage workspace tabs, stream background daemons, broadcast events across an internal bus, stream real-time LLM tokens without screen tearing, and render rich graphical widgets (tables, sparklines, bar charts).

When an AI generates or edits a plugin in MT Core, it must adhere to the **manifest headers**, the **JSON-over-stdout IPC protocol**, and the **environment variable contract** defined in this document.

---

## 📁 1. Directory Structure & Plugin Discovery

All plugins live in the `/plugins` directory:

```text
/plugins/
├── sdk/
│   ├── mt.py               # Official Python SDK (Super User x1000 controller)
│   ├── mt.js               # Official Node.js SDK (JavaScript / TypeScript controller)
│   └── mt.sh               # Official Bash SDK (POSIX helpers)
├── hello.py                # Minimal Python plugin
├── monitor.py              # System monitoring & visual telemetry plugin
├── superuser_demo.py       # Full kernel & UI mutation demo
├── chart_demo.py           # Rich visual graphics (tables, sparklines, bar charts)
├── stream_agent.py         # Real-time token streaming and Event Bus demo
├── node_demo.js            # First-class Node.js V8 plugin demo
├── canvas.sh               # Bash 2D virtual framebuffer canvas demo
└── <your_plugin>.<py|js|sh|go> # Any new plugin placed here is auto-loaded!
```

### Discovery & Lifecycle Rules:
1. **Dynamic Scanning:** Any executable script or binary placed inside `/plugins/` is automatically detected and parsed by `pkg/plugin/dynamic.go`.
2. **Multi-Language Runtime Matrix:**
   - `.py` → Executed via `python3` (or `python`).
   - `.js`, `.mjs`, `.cjs` → Executed via `node` (Node.js V8 engine).
   - `.sh` → Executed via `bash` (or `sh`).
   - `.go` → Automatically compiled on-the-fly to a binary in `~/.mt/bin/` by MT Core via `go build`.
   - Any executable without extension → Executed directly as a native OS binary.
3. **Execution Modes:**
   - **Synchronous Command (Default):** Executed when the user runs `/<command>` in the shell. Timeout is 15 seconds by default (customizable via `@timeout`).
   - **Background Daemon:** Runs continuously in the background upon MT Core launch if `@daemon: true` is specified.
   - **Streaming Mode:** Emits incremental tokens or logs via `stream_append` to update the active viewport in real time.

---

## 🏷️ 2. Plugin Manifest Headers (Mandatory Metadata)

MT Core inspects the first **100 lines** of the plugin source code for metadata tags. Comment syntax can be `#` (Python/Shell) or `//` (JavaScript/Go/C).

### Supported Manifest Tags

| Tag | Example | Description |
|---|---|---|
| `@command:` or `@name:` | `# @command: netstat` | The primary shell slash-command (e.g. `/netstat`). |
| `@title:` | `# @title: NETWORK MONITOR` | Human-readable title displayed in autocomplete and tabs. |
| `@category:` | `# @category: NET` | Grouping category (`SYS`, `NET`, `AI`, `ANALYTICS`, `RUNTIME`, `TOOL`, etc.). |
| `@desc:` | `# @desc: Real-time socket visualizer` | Short description displayed in `/help` and dropdown search. |
| `@author:` | `# @author: ai-agent-42` | Author credit string. |
| `@version:` | `# @version: 1.0.0` | Semantic version string. |
| `@icon:` | `# @icon: 🌐` | Icon or emoji displayed next to command title. |
| `@aliases:` or `@alias:` | `# @aliases: ns, sockets` | Comma-separated command aliases. |
| `@requires:` or `@require:` | `# @requires: curl, jq` | Required system dependencies or tools. |
| `@timeout:` | `# @timeout: 8s` | Maximum execution duration (Go duration string: `5s`, `30s`, `1m`). Default: `15s`. |
| `@daemon:` | `# @daemon: true` | If `true` or `1`, MT Core runs this script continuously in background. |
| `@perm:` or `@permission:` | `# @perm: net, stream, eventbus` | Requested capabilities (`net`, `fs`, `stream`, `eventbus`, `daemon`). |
| `@hook:` or `@hooks:` | `# @hook: on_load, on_unload` | Declares lifecycle event hooks for the plugin. |
| `@tags:` or `@tag:` | `# @tags: telemetry, network, charts`| Comma-separated searchable tags for plugin indexing. |

### Capability Bitmasks Automatically Granted

When `@perm:` is declared in the manifest, MT Core activates corresponding kernel security privileges:
- `net` / `network` → `CapNetwork` (Permits outbound socket / HTTP requests).
- `fs` / `filesystem` → `CapFileSystem` (Permits disk reads/writes in workspace).
- `stream` / `streaming` → `CapStream` (Permits viewport token streaming).
- `event` / `eventbus` → `CapEventBus` (Permits publishing/subscribing to Event Bus).
- `daemon` → `CapDaemon` (Permits persistent daemon execution).

### Compact Multi-Command Tag (`@cmd:`)
To register multiple subcommands from a single file:
```python
# @cmd:<command_name>:<Title>:<CATEGORY>:<Description>
# @cmd:ip:IP ADDRESS:NET:Show local and public IP
# @cmd:ping:NETWORK PING:NET:Ping remote host
```

---

## 🌐 3. Environment Variables Injected by MT Core

When MT Core launches your plugin, it populates standard environment variables:

| Variable | Description |
|---|---|
| `MT_WIDTH` | Terminal column width (integer, e.g. `120`). |
| `MT_HEIGHT` | Terminal row height (integer, e.g. `35`). |
| `MT_THEME` | Active theme name (`amber`, `emerald`, `cyber`, `mono`, etc.). |
| `MT_INVOKED_CMD` | The command string used to trigger the plugin. |
| `MT_PIPE_INPUT` | Up to 1024 characters of piped standard input (from `/cmd1 \| /cmd2`). |
| `MT_PLUGINS_DIR` | Absolute path to the `/plugins` directory. |
| `MT_SUPERUSER` | Set to `"1"` (denotes ring-0 privileges). |
| `MT_DAEMON` | Set to `"1"` if currently executing in daemon mode (`--daemon`). |
| `MT_KV_<KEY>` | All shared persistent key-value store entries automatically exported into environment! |

---

## ⚡ 4. IPC Protocol (Super User x1000 Actions)

MT Core intercepts plugin stdout lines.
- **Rule 1:** Any line that starts with `{` and ends with `}` is parsed as an **IPC JSON Action**.
- **Rule 2:** Any regular text line (not JSON) is treated as **visual output** and rendered in the active terminal workspace tab.
- **Rule 3:** You can mix IPC commands and regular terminal output in the same plugin run!

### 📡 Complete List of Supported IPC Actions

#### A. Real-Time Streaming & Event Bus
| Action | Parameters | Description |
|---|---|---|
| `stream_append` | `chunk: string` | Streams an incremental token or text chunk directly into the active viewport without clearing screen. |
| `emit_event` | `event: string, data: any` | Publishes an event on the MT Core event bus. Persisted to `EVENT_<NAME>` in shared state. |

#### B. Rich Graphics & Visual Analytics
| Action | Parameters | Description |
|---|---|---|
| `table` | `headers: []string, rows: [][]string` | Renders an auto-aligned table with rounded UTF-8 borders (`╭┬╮├┼┤╰┴╯`). |
| `sparkline` | `values: []float64, title?: string, color?: string` | Renders an 8-level UTF-8 sparkline (` ▂▃▄▅▆▇█`) for latency, memory, or metric trends. |
| `barchart` | `data: {key: float64}, title?: string, color?: string` | Renders a proportional horizontal bar chart with percentage bars. |
| `clear_tab` | `index?: int` | Clears all rendered content in the specified tab (-1 for active tab). |

#### C. Live UI & Frame Mutation
| Action | Parameters | Description |
|---|---|---|
| `toast` | `message: string, is_error: bool` | Triggers a top-right floating toast notification. |
| `set_status` | `message: string, is_error: bool` | Sets the persistent status indicator in the bottom bar. |
| `set_banner` | `banner: string` | Overrides the top ASCII logo/header in real time. |
| `reset_banner` | None | Resets the top header back to default MT Core branding. |
| `set_footer` | `footer: string` | Overrides bottom shortcut hints. |
| `reset_footer` | None | Restores standard bottom hints. |
| `modal` | `title: string, content: string` | Displays a centered modal dialog box. |
| `hijack_screen`| `content: string` | Completely takes over the terminal view with custom content. |
| `release_screen`| None | Releases the fullscreen hijack back to workspace tabs. |
| `set_input_value`| `value: string` | Prepopulates or alters the command line input box. |
| `set_input_placeholder`| `placeholder: string` | Updates the input box placeholder prompt. |

#### D. Dynamic Command Injection
| Action | Parameters | Description |
|---|---|---|
| `register_cmd`| `command: string, title: string, category: string, desc: string, exec?: string` | Injects a new slash command into MT Core on-the-fly! |
| `unregister_cmd`| `command: string` | Removes a registered dynamic command from the kernel. |
| `trigger_cmd` | `command: string` | Programmatically triggers any command as if typed by the user. |

#### E. 2D Virtual Framebuffer Manipulation (Canvas)
| Action | Parameters | Description |
|---|---|---|
| `pixel` | `x: int, y: int, char: string, fg: hex, bg: hex` | Mutates an individual cell in the 60 FPS framebuffer. |
| `pixels` | `pixels: [{x, y, char, fg, bg}, ...]` | Batch draws an array of cells at once. |
| `box` | `x: int, y: int, w: int, h: int, title: string, fg: hex` | Draws a framed geometric box with optional title. |
| `clear_pixels` | None | Wipes all virtual canvas mutations. |

#### F. Tabs, Workspaces & Theme Control
| Action | Parameters | Description |
|---|---|---|
| `switch_tab` | `index: int` | Switches active focus to tab `index` (0-indexed). |
| `inject_tab` | `title: string, cmd: string, output: string` | Spawns a brand new tab with custom title and content. |
| `update_tab` | `index: int, output: string` | Overwrites content of an existing tab. |
| `set_tab_title`| `index: int, title: string` | Renames an existing tab title. |
| `scroll_tab` | `index: int, delta: int` | Programmatically scrolls tab viewport up or down. |
| `close_all_tabs`| None | Closes all open workspace tabs. |
| `theme` | `name: string` (`amber`, `emerald`, `cyber`, `mono`) | Switches the active global theme. |
| `set_theme_color`| `field: string, color: hex` | Hot-patches individual theme colors (`primary`, `border`, `bg_dark`, etc.). |

#### G. Shared State (KV Store)
| Action | Parameters | Description |
|---|---|---|
| `kv_set` | `key: string, value: string` | Persists a key-value pair across all plugins and commands. |
| `kv_del` | `key: string` | Removes a key from the shared KV store. |

---

## 🛠️ 5. SDK Reference & Examples

### A. Node.js SDK (`plugins/sdk/mt.js`)
First-class support for JavaScript & TypeScript plugins:

```javascript
#!/usr/bin/env node
// @name: sysinfo
// @desc: JavaScript System Diagnostic Tool
// @category: DIAG
// @perm: net, stream, eventbus
// @tags: node, diag, memory

const path = require('path');
const { MT, Color, Style } = require(path.join(__dirname, 'sdk', 'mt.js'));

function main() {
  const mt = new MT();

  console.log(Style.badge('SYSTEM DIAGNOSTICS', Color.DARK, Color.CYBER));

  // Render a styled info card
  console.log(Style.card('MEMORY STATS', {
    'Platform': process.platform,
    'Uptime': `${Math.floor(process.uptime())}s`,
    'Heap Used': `${Math.round(process.memoryUsage().heapUsed / 1024 / 1024)} MB`
  }, Color.CYBER));

  // Render a native host table
  mt.renderTable(
    ['Worker', 'State', 'Tasks', 'Load'],
    [
      ['worker-01', 'ONLINE', '42', '12%'],
      ['worker-02', 'ONLINE', '87', '34%']
    ]
  );

  // Broadcast an event
  mt.emitEvent('diag_complete', { ok: true, timestamp: Date.now() });

  // Floating notification
  mt.toast('System diagnostics compiled successfully.');
}

main();
```

---

### B. Python SDK (`plugins/sdk/mt.py`)
Fluent Python controller with rich terminal helpers:

```python
#!/usr/bin/env python3
# @name: agent_query
# @desc: LLM Reasoning and Token Streaming
# @category: AI-AGENT
# @perm: stream, eventbus
# @tags: ai, llm, stream

import sys, os, time
sys.path.append(os.path.join(os.path.dirname(__file__), "sdk"))
from mt import MT, Color, Style

def main():
    mt = MT()
    query = " ".join(sys.argv[1:]) or "Cluster Status"

    mt.set_status(f"Thinking: '{query}'...", is_error=False)
    mt.emit_event("query_started", {"query": query})

    tokens = ["Gathering telemetry... ", "Nodes: 12/12 online. ", "Latency nominal (1.2ms)."]
    for t in tokens:
        mt.stream_append(t)
        time.sleep(0.05)

    # Render a Sparkline of latency
    mt.render_sparkline([12.4, 11.2, 14.0, 9.8, 15.1, 10.2], title="Cluster Latency (ms)", color=Color.EMERALD)

    mt.kv_set("LAST_QUERY", query)
    mt.set_status("Done", is_error=False)

if __name__ == "__main__":
    main()
```

---

### C. Bash / POSIX Shell SDK (`plugins/sdk/mt.sh`)
Lightweight, zero-dependency Shell functions:

```bash
#!/usr/bin/env bash
# @name: diskcheck
# @desc: Disk usage alert
# @category: SYS

source "$(dirname "$0")/sdk/mt.sh"

USAGE=$(df -h / | awk 'NR==2 {print $5}')
mt_toast "Root Disk Usage: $USAGE" false
mt_status "Disk Check: $USAGE OK" false
mt_stream "Checked / partition: $USAGE in use.\n"
```

---

## 🚀 6. Best Practices for AI Agents

1. **Always declare manifest headers:** Ensure `@command:`, `@title:`, `@category:`, `@desc:`, and `@perm:` are in the first 100 lines.
2. **Handle piping:** Read from stdin via `mt.get_pipe_input()` or `$MT_PIPE_INPUT` so your plugin supports unix pipelines (`/sysinfo | /myplugin`).
3. **Use ANSI color constants:** Use `Color.AMBER`, `Color.EMERALD`, `Color.CYBER`, etc., to blend seamlessly with the active MT Core color scheme.
4. **Use JSON stdout only for IPC:** Any output that should not mutate the UI must be plain text, or emitted via `mt.stream_append()`.
5. **Persist state wisely:** Use `mt.kv_set(key, val)` for data that downstream plugins or subsequent runs will need.
