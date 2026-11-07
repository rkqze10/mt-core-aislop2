# MT Core Plugins (Gelişmiş Çoklu Çalışma Zamanlı Eklenti Sistemi)

Bu dizin MT Core eklentilerini (**Python**, **Node.js / JavaScript**, **Bash / Shell**, **Go** ve derlenmiş ikili dosyalar) içerir.

> 🤖 **Yapay Zeka (AI) ve Geliştiriciler İçin Kapsamlı Rehber:**  
> Eklenti geliştirme mimarisi, Node.js V8 çalışma zamanı, Super User x1000 IPC protokolü, gerçek zamanlı LLM akışı (Streaming), Olay Veri Yolu (Event Bus), veri görselleştirme (Tablo, Sparkline, Çubuk Grafik) ve SDK referansı için ana dizindeki [**PLUGIN_README.md**](../PLUGIN_README.md) dosyasını inceleyin.

---

### Desteklenen Diller ve Çalışma Zamanları:
1. **Python (`.py`):** `python3` üzerinden çalıştırılır. SDK: `sdk/mt.py`.
2. **Node.js (`.js`, `.mjs`, `.cjs`):** `node` V8 çalışma zamanı üzerinden çalıştırılır. SDK: `sdk/mt.js`.
3. **Shell (`.sh`):** `bash`/`sh` üzerinden çalıştırılır. SDK: `sdk/mt.sh`.
4. **Go (`.go`):** Çalışma anında otomatik olarak `go build` ile derlenir ve yürütülür.
5. **Yerel İkili Dosyalar (ELF/PE/Mach-O):** Doğrudan işletim sistemi prosesi olarak başlatılır.

### Yeni Nesil Yetenekler:
- ⚡ **Gerçek Zamanlı Akış (Streaming Output):** `stream_append` ile LLM token veya canlı log akışı.
- 📡 **Olay Veri Yolu (Event Bus):** Eklentiler arası asenkron pub/sub haberleşmesi (`emit_event`).
- 📊 **Veri Görselleştirme:** Çekirdek tarafından render edilen UTF-8 Tablolar, 8-kademeli Sparkline'lar (` ▂▃▄▅▆▇█`) ve Çubuk Grafikler.
- 🛡️ **İzin & Yetki Modeli:** `@perm: net, fs, stream, eventbus, daemon` başlık etiketleri ve `@tags: ...` etiketleme.
- 🕹️ **Sanal Piksel Framebuffer:** 60 FPS piksel matrisine doğrudan erişim (`pixel`, `box`, `draw_batch`).
