#!/usr/bin/env python3
# @name: chart_demo
# @desc: Terminal veri görselleştirme (Tablo, Sparkline, Çubuk Grafik)
# @category: ANALYTICS
# @perm: net, stream
# @tags: analytics, charts, tables, telemetry

import sys
import os
import time

sys.path.append(os.path.join(os.path.dirname(__file__), "sdk"))
from mt import MT, Color, Style

def main():
    mt = MT()

    print(Style.badge("MT CORE ANALYTICS ENGINE", fg=Color.DARK, bg=Color.CYBER))
    print(Style.color("Terminal İçi Grafik & Veri Görselleştirme Konsolu\n", fg=Color.MUTED))

    # 1. Bilgi Kartı
    card = MT.card("SİSTEM TELEMETRİSİ", {
        "Düğüm": "worker-node-01",
        "Durum": Style.color("AKTİF", fg=Color.EMERALD, bold=True),
        "Yük (Load)": "0.42, 0.38, 0.29",
        "Bellek (RAM)": MT.progress_bar(5.4, 16.0, width=16),
        "Disk (NVMe)": MT.progress_bar(128.0, 512.0, width=16, fg_done=Color.CYBER)
    }, color=Color.CYBER, width=54)
    print(card)
    print()

    # 2. Host Tarafından Render Edilen Tablo
    mt.render_table(
        headers=["Servis", "Protokol", "Port", "Gecikme (ms)", "Durum"],
        rows=[
            ["api-gateway", "HTTP/2", "8080", "1.2ms", "HEALTHY"],
            ["auth-service", "gRPC", "50051", "0.8ms", "HEALTHY"],
            ["postgres-cluster", "TCP", "5432", "2.4ms", "HEALTHY"],
            ["redis-cache", "RESP", "6379", "0.3ms", "HEALTHY"],
            ["kafka-broker", "PLAINTEXT", "9092", "4.1ms", "DEGRADED"]
        ]
    )

    # 3. Host Sparkline (Zaman Serisi Trendi)
    latencies = [12.0, 14.5, 11.0, 9.8, 15.2, 22.0, 45.1, 80.0, 65.4, 30.2, 18.0, 14.0, 12.5, 11.8, 10.2]
    mt.render_sparkline(latencies, title="API Gateway 15-Dk Gecikme Trendi (ms)", color=Color.EMERALD)

    # 4. Host Çubuk Grafik (Hacim Dağılımı)
    req_distribution = {
        "/api/v1/auth": 45000.0,
        "/api/v1/query": 128000.0,
        "/api/v1/stream": 82000.0,
        "/api/v1/mutate": 19000.0,
        "/healthz": 240000.0
    }
    mt.render_barchart(req_distribution, title="Son 1 Saat Endpoint İstek Hacmi", color=Color.AMBER)

    mt.toast("Grafik panosu başarıyla derlendi.")
    mt.emit_event("telemetry_updated", {"node": "worker-node-01", "status": "ok"})

if __name__ == "__main__":
    main()
