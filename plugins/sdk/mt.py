"""
MT Core Python Super User SDK (Super User x1000)
A high-level, fluent library for building MT Core plugins with full terminal control.
"""

import sys
import os
import json
from typing import List, Optional, Tuple, Dict, Any

class Color:
    AMBER   = "#FF9E3B"
    EMERALD = "#00FF88"
    CYBER   = "#00F0FF"
    PINK    = "#FF007F"
    PURPLE  = "#A855F7"
    YELLOW  = "#EAB308"
    RED     = "#EF4444"
    GREEN   = "#22C55E"
    BLUE    = "#3B82F6"
    WHITE   = "#FAFAFA"
    MUTED   = "#71717A"
    DARK    = "#18181B"

class Style:
    RESET     = "\033[0m"
    BOLD      = "\033[1m"
    DIM       = "\033[2m"
    ITALIC    = "\033[3m"
    UNDERLINE = "\033[4m"

    @staticmethod
    def hex_fg(hex_code: str) -> str:
        hex_code = hex_code.lstrip("#")
        if len(hex_code) == 6:
            r = int(hex_code[0:2], 16)
            g = int(hex_code[2:4], 16)
            b = int(hex_code[4:6], 16)
            return f"\033[38;2;{r};{g};{b}m"
        return ""

    @staticmethod
    def hex_bg(hex_code: str) -> str:
        hex_code = hex_code.lstrip("#")
        if len(hex_code) == 6:
            r = int(hex_code[0:2], 16)
            g = int(hex_code[2:4], 16)
            b = int(hex_code[4:6], 16)
            return f"\033[48;2;{r};{g};{b}m"
        return ""

    @classmethod
    def color(cls, text: str, fg: str = None, bg: str = None, bold: bool = False) -> str:
        out = ""
        if bold:
            out += cls.BOLD
        if fg:
            out += cls.hex_fg(fg)
        if bg:
            out += cls.hex_bg(bg)
        out += str(text) + cls.RESET
        return out

    @classmethod
    def badge(cls, text: str, fg: str = Color.WHITE, bg: str = Color.DARK) -> str:
        return f"{cls.hex_bg(bg)}{cls.hex_fg(fg)} {text} {cls.RESET}"

class MT:
    """Primary Super User x1000 host controller."""

    def __init__(self):
        self.width = int(os.environ.get("MT_WIDTH", "80"))
        self.height = int(os.environ.get("MT_HEIGHT", "24"))
        self.theme = os.environ.get("MT_THEME", "amber")
        self.plugins_dir = os.environ.get("MT_PLUGINS_DIR", "./plugins")
        self.is_daemon = os.environ.get("MT_DAEMON", "0") == "1"

    def _emit(self, action: str, **kwargs) -> None:
        payload = {"action": action, **kwargs}
        sys.stdout.write(json.dumps(payload) + "\n")
        sys.stdout.flush()

    # --- UI & Feedback ---
    def toast(self, message: str, is_error: bool = False) -> None:
        self._emit("toast", message=message, is_error=is_error)

    def set_status(self, message: str, is_error: bool = False) -> None:
        self._emit("set_status", message=message, is_error=is_error)

    def set_banner(self, banner_text: str) -> None:
        self._emit("set_banner", banner=banner_text)

    def reset_banner(self) -> None:
        self._emit("reset_banner")

    def set_footer(self, footer_text: str) -> None:
        self._emit("set_footer", footer=footer_text)

    def reset_footer(self) -> None:
        self._emit("reset_footer")

    def set_input(self, value: str) -> None:
        self._emit("set_input_value", value=value)

    def set_placeholder(self, placeholder: str) -> None:
        self._emit("set_input_placeholder", placeholder=placeholder)

    def modal(self, title: str, content: str) -> None:
        self._emit("modal", title=title, content=content)

    def hijack_screen(self, content: str) -> None:
        self._emit("hijack_screen", content=content)

    def release_screen(self) -> None:
        self._emit("release_screen")

    # --- Commands & Dynamic Injection ---
    def register_command(self, command: str, title: str = "", category: str = "SUPERUSER", desc: str = "", exec_script: str = "") -> None:
        self._emit("register_cmd", command=command, title=title or command.upper(), category=category, desc=desc, exec=exec_script)

    def unregister_command(self, command: str) -> None:
        self._emit("unregister_cmd", command=command)

    def trigger_command(self, command_line: str) -> None:
        self._emit("trigger_cmd", command=command_line)

    def launch_tui(self, binary: str, args: List[str] = None, name: str = "", prompt: str = "") -> None:
        self._emit("launch_tui", binary=binary, args=args or [], name=name or binary, prompt=prompt)

    # --- Tabs & Workspace ---
    def switch_tab(self, index: int) -> None:
        self._emit("switch_tab", index=index)

    def close_all_tabs(self) -> None:
        self._emit("close_all_tabs")

    def close_tab(self, index: int) -> None:
        self._emit("force_close_tab", index=index)

    def inject_tab(self, title: str, cmd: str, output: str) -> None:
        self._emit("inject_tab", title=title, cmd=cmd, output=output)

    def update_tab(self, index: int, output: str) -> None:
        self._emit("update_tab", index=index, output=output)

    def set_tab_title(self, index: int, title: str) -> None:
        self._emit("set_tab_title", index=index, title=title)

    def scroll_tab(self, index: int = -1, delta: int = 3) -> None:
        self._emit("scroll_tab", index=index, delta=delta)

    # --- Themes ---
    def set_theme(self, name: str) -> None:
        self._emit("theme", name=name)

    def set_theme_color(self, field: str, hex_color: str) -> None:
        self._emit("set_theme_color", field=field, color=hex_color)

    # --- 2D Pixel Framebuffer & Virtual Canvas ---
    def pixel(self, x: int, y: int, char: str = "█", fg: str = None, bg: str = None) -> None:
        self._emit("pixel", x=x, y=y, char=char, fg=fg or "", bg=bg or "")

    def box(self, x: int, y: int, w: int, h: int, title: str = "", fg: str = None) -> None:
        self._emit("box", x=x, y=y, w=w, h=h, title=title, fg=fg or "")

    def clear_pixels(self) -> None:
        self._emit("clear_pixels")

    # --- Streaming & Real-Time Output ---
    def stream(self, chunk: str) -> None:
        """Stream real-time tokens/chunks directly to the current viewport."""
        self._emit("stream_append", chunk=chunk)

    def stream_append(self, chunk: str) -> None:
        """Alias for stream()."""
        self._emit("stream_append", chunk=chunk)

    # --- Event Bus & Pub/Sub ---
    def emit_event(self, event_name: str, payload: Any = None) -> None:
        """Publish an event across the MT Core event bus."""
        self._emit("emit_event", event=event_name, data=payload)

    # --- Rich Graphics & Visual Widgets ---
    def render_table(self, headers: List[str], rows: List[List[str]]) -> None:
        """Request the host kernel to render a styled UTF-8 table."""
        self._emit("table", headers=headers, rows=rows)

    def render_sparkline(self, values: List[float], title: str = "", color: str = "") -> None:
        """Request the host kernel to render an 8-level UTF-8 sparkline."""
        self._emit("sparkline", values=values, title=title, color=color)

    def render_barchart(self, data: Dict[str, float], title: str = "", color: str = "") -> None:
        """Request the host kernel to render a proportional horizontal bar chart."""
        self._emit("barchart", data=data, title=title, color=color)

    def clear_tab(self, index: int = -1) -> None:
        """Clear output content of the specified tab (-1 for active tab)."""
        self._emit("clear_tab", index=index)

    # --- State & Persistence ---
    def kv_set(self, key: str, value: str) -> None:
        self._emit("kv_set", key=key, value=value)
        clean_key = "MT_KV_" + "".join(c.upper() if c.isalnum() else "_" for c in key)
        os.environ[clean_key] = value

    def kv_get(self, key: str, default: str = "") -> str:
        clean_key = "MT_KV_" + "".join(c.upper() if c.isalnum() else "_" for c in key)
        return os.environ.get(clean_key, default)

    def kv_delete(self, key: str) -> None:
        self._emit("kv_del", key=key)
        clean_key = "MT_KV_" + "".join(c.upper() if c.isalnum() else "_" for c in key)
        if clean_key in os.environ:
            del os.environ[clean_key]

    # --- Pipe Input ---
    def get_pipe_input(self) -> str:
        try:
            if not sys.stdin.isatty():
                return sys.stdin.read()
        except Exception:
            pass
        return os.environ.get("MT_PIPE_INPUT", "")

    # --- Visual Widgets Helpers ---
    @staticmethod
    def progress_bar(current: float, total: float, width: int = 24, fg_done: str = Color.EMERALD, fg_todo: str = Color.MUTED) -> str:
        safe_width = max(1, width)
        ratio = max(0.0, min(1.0, current / total if total > 0 else 0.0))
        filled = int(ratio * safe_width)
        empty = safe_width - filled
        bar = Style.color("━" * filled, fg=fg_done, bold=True) + Style.color("━" * empty, fg=fg_todo)
        pct = int(ratio * 100)
        return f"{bar} {Style.color(f'{pct}%', fg=Color.WHITE, bold=True)}"

    @staticmethod
    def sparkline_local(values: List[float], width: int = 20) -> str:
        """Pure-Python inline sparkline generator."""
        if not values:
            return ""
        ticks = [" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"]
        min_v = min(values)
        max_v = max(values)
        span = max_v - min_v
        res = []
        for v in values[-width:]:
            if span <= 0:
                res.append(ticks[0])
            else:
                idx = int(((v - min_v) / span) * (len(ticks) - 1))
                idx = max(0, min(len(ticks) - 1, idx))
                res.append(ticks[idx])
        return "".join(res)

    @staticmethod
    def card(title: str, items: Dict[str, str], color: str = Color.AMBER, width: int = 44) -> str:
        safe_width = max(10, width)
        lines = []
        top_border = "╭─ " + Style.color(f" {title} ", fg=color, bold=True) + "─" * max(0, safe_width - len(title) - 6) + "╮"
        lines.append(top_border)
        for k, v in items.items():
            k_fmt = Style.color(f"  {k:<16}: ", fg=Color.MUTED)
            v_fmt = Style.color(str(v), fg=Color.WHITE, bold=True)
            lines.append(k_fmt + v_fmt)
        lines.append("╰" + "─" * max(0, safe_width - 2) + "╯")
        return "\n".join(lines)

