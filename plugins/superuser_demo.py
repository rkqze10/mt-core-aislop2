# @command: superuser_demo
# @title: SUPER USER X1000 DEMO
# @category: RING-1
# @desc: Demonstrates full kernel mutation and control

import json
import sys

# 1. Dynamically register a new custom command directly into MT Core
add_cmd_ipc = {
    "action": "register_cmd",
    "command": "hacked",
    "title": "HACKED CMD",
    "category": "SUPERUSER",
    "desc": "This command was dynamically injected into MT Core!",
    "exec": "echo '>>> [SUPER USER x1000] Injected command executed successfully! <<<' "
}
print(json.dumps(add_cmd_ipc))

# 2. Inject a custom cyberpunk banner
banner_ipc = {
    "action": "set_banner",
    "banner": """
  ██████╗  ██████╗ ██████╗     ███╗   ███╗ ██████╗ ██████╗ ███████╗
 ██╔════╝ ██╔═══██╗██╔══██╗    ████╗ ████║██╔═══██╗██╔══██╗██╔════╝
 ██║  ███╗██║   ██║██║  ██║    ██╔████╔██║██║   ██║██║  ██║█████╗  
 ██║   ██║██║   ██║██║  ██║    ██║╚██╔╝██║██║   ██║██║  ██║██╔══╝  
 ╚██████╔╝╚██████╔╝██████╔╝    ██║ ╚═╝ ██║╚██████╔╝██████╔╝███████╗
  ╚═════╝  ╚═════╝ ╚═════╝     ╚═╝     ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝
    ⚡ SUPER USER x1000 ENGINE // ROOT OVERLORD ONLINE ⚡
"""
}
print(json.dumps(banner_ipc))

# 3. Mutate the status line
status_ipc = {
    "action": "set_status",
    "message": "⚡ [GOD MODE] Eklenti çekirdeği başarıyla ele geçirdi. /hacked komutu eklendi!",
    "is_error": False
}
print(json.dumps(status_ipc))

# 4. Mutate primary theme color to neon purple/cyan
theme_ipc = {
    "action": "set_theme_color",
    "field": "primary",
    "color": "#D946EF"
}
print(json.dumps(theme_ipc))

# 5. Output plain text feedback for the workspace tab
print("=" * 60)
print("⚡ SUPER USER x1000 ÇEKİRDEK MUTASYON RAPORU:")
print(" • /hacked komutu çalışma anında sisteme eklendi.")
print(" • Ana ASCII banner değiştirildi.")
print(" • Tema birincil rengi neon mor (#D946EF) yapıldı.")
print(" • Durum mesajı güncellendi.")
print("=" * 60)
print("Şimdi komut satırına /hacked yazarak deneyebilirsiniz!")
