#!/usr/bin/env bash
# MT Core Shell Super User SDK (Super User x1000)

_mt_escape_json() {
  local s="$1"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\n'/\\n}"
  s="${s//$'\r'/}"
  s="${s//$'\t'/\\t}"
  printf '%s' "$s"
}

_mt_bool() {
  local val="$1"
  if [ "$val" = "true" ] || [ "$val" = "1" ]; then
    printf 'true'
  else
    printf 'false'
  fi
}

mt_toast() {
  local msg="$(_mt_escape_json "$1")"
  local is_err="$(_mt_bool "${2:-false}")"
  printf '{"action":"toast","message":"%s","is_error":%s}\n' "$msg" "$is_err"
}

mt_status() {
  local msg="$(_mt_escape_json "$1")"
  local is_err="$(_mt_bool "${2:-false}")"
  printf '{"action":"set_status","message":"%s","is_error":%s}\n' "$msg" "$is_err"
}

mt_banner() {
  local banner="$(_mt_escape_json "$1")"
  printf '{"action":"set_banner","banner":"%s"}\n' "$banner"
}

mt_register_cmd() {
  local cmd="$(_mt_escape_json "$1")"
  local title="$(_mt_escape_json "$2")"
  local cat="$(_mt_escape_json "$3")"
  local desc="$(_mt_escape_json "$4")"
  local exec_cmd="$(_mt_escape_json "$5")"
  printf '{"action":"register_cmd","command":"%s","title":"%s","category":"%s","desc":"%s","exec":"%s"}\n' "$cmd" "$title" "$cat" "$desc" "$exec_cmd"
}

mt_pixel() {
  local x="$1"
  local y="$2"
  local ch="$(_mt_escape_json "${3:-█}")"
  local fg="$(_mt_escape_json "${4:-}")"
  local bg="$(_mt_escape_json "${5:-}")"
  printf '{"action":"pixel","x":%d,"y":%d,"char":"%s","fg":"%s","bg":"%s"}\n' "$x" "$y" "$ch" "$fg" "$bg"
}

mt_box() {
  local x="$1"
  local y="$2"
  local w="$3"
  local h="$4"
  local title="$(_mt_escape_json "${5:-}")"
  local fg="$(_mt_escape_json "${6:-}")"
  printf '{"action":"box","x":%d,"y":%d,"w":%d,"h":%d,"title":"%s","fg":"%s"}\n' "$x" "$y" "$w" "$h" "$title" "$fg"
}

mt_clear_pixels() {
  printf '{"action":"clear_pixels"}\n'
}

mt_switch_tab() {
  local idx="$1"
  printf '{"action":"switch_tab","index":%d}\n' "$idx"
}

mt_set_theme_color() {
  local field="$(_mt_escape_json "$1")"
  local color="$(_mt_escape_json "$2")"
  printf '{"action":"set_theme_color","field":"%s","color":"%s"}\n' "$field" "$color"
}

mt_kv_set() {
  local key="$(_mt_escape_json "$1")"
  local val="$(_mt_escape_json "$2")"
  printf '{"action":"kv_set","key":"%s","value":"%s"}\n' "$key" "$val"
}

mt_kv_del() {
  local key="$(_mt_escape_json "$1")"
  printf '{"action":"kv_del","key":"%s"}\n' "$key"
}

mt_stream() {
  local chunk="$(_mt_escape_json "$1")"
  printf '{"action":"stream_append","chunk":"%s"}\n' "$chunk"
}

mt_emit_event() {
  local evt="$(_mt_escape_json "$1")"
  local data="$(_mt_escape_json "$2")"
  printf '{"action":"emit_event","event":"%s","data":"%s"}\n' "$evt" "$data"
}

mt_clear_tab() {
  local idx="${1:--1}"
  printf '{"action":"clear_tab","index":%d}\n' "$idx"
}

