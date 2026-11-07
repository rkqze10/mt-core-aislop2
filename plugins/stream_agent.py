#!/usr/bin/env python3
# @name: stream_agent
# @desc: LLM / AI token streaming ve Olay İletişimi Simülasyonu
# @category: AI-AGENT
# @perm: stream, eventbus
# @tags: ai, llm, stream, agent, eventbus

import sys
import os
import time

sys.path.append(os.path.join(os.path.dirname(__file__), "sdk"))
from mt import MT, Color, Style

def main():
    mt = MT()

    query = " ".join(sys.argv[1:]) if len(sys.argv) > 1 else "MT Core mimarisini analiz et"

    mt.set_status(f"Düşünülüyor: '{query}'...", is_error=False)
    mt.emit_event("agent_started", {"query": query, "model": "gemini-2.5-flash"})

    # Streaming token token simülasyonu
    intro = f"{Style.badge('AI AGENT REASONING ENGINE', fg=Color.DARK, bg=Color.EMERALD)}\n"
    intro += f"{Style.color('Sorgu:', fg=Color.MUTED)} {Style.color(query, fg=Color.WHITE, bold=True)}\n\n"
    sys.stdout.write(intro)
    sys.stdout.flush()

    response_tokens = [
        "MT Core ", "TUI mimarisi, ", "Go dili ile ", "geliştirilmiş ", "çift katmanlı ",
        "bir hipervizördür. ", "\n\n",
        "Temel Bileşenler:\n",
        "• ", Style.color("Sanal Çerçeve Belleği (Framebuffer):", fg=Color.AMBER, bold=True), " 60 FPS piksel matrisi.\n",
        "• ", Style.color("IPC Motoru:", fg=Color.CYBER, bold=True), " Stdout üzerinden JSON tabanlı çift yönlü haberleşme.\n",
        "• ", Style.color("Olay Veri Yolu (Event Bus):", fg=Color.EMERALD, bold=True), " Eklentiler arası pub/sub koordinasyonu.\n",
        "• ", Style.color("Gerçek Zamanlı Akış (Streaming):", fg=Color.PINK, bold=True), " LLM token'larını anında ekrana yansıtma.\n\n",
        "İşlem başarıyla tamamlandı. Model durumu: Hazır."
    ]

    for token in response_tokens:
        mt.stream_append(token)
        # Doğrudan ekrana da basabilir veya host stream append ile güncelleyebilir
        time.sleep(0.04)

    mt.kv_set("LAST_AGENT_QUERY", query)
    mt.kv_set("LAST_AGENT_STATUS", "SUCCESS")
    mt.set_status("Yanıt tamamlandı.", is_error=False)
    mt.emit_event("agent_completed", {"query": query, "tokens": len(response_tokens)})

if __name__ == "__main__":
    main()
