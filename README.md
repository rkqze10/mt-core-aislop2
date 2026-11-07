# 🤖 MT Core - Modular Terminal Shell & Super User Plugin Engine

Terminal UI (TUI) tabanlı komut kabuğu, çok sekmeli çalışma alanı ve **Super User x1000** yeteneklerine sahip dinamik eklenti motoru.

---

## 🚀 Çalıştırma

```bash
go run main.go
```

Komut isteminde `/` yazarak mevcut tüm komutları görebilirsiniz:
* `/help` : Tüm dahili ve dinamik komutları listeler.
* `/sys` : Sistem telemetrisi ve bellek durumunu görüntüler.
* `/theme <amber|emerald|cyber|mono>` : Renk temasını değiştirir.
* `/sh <komut>` : Arka planda OS komutlarını çalıştırır (örn: `/sh ls -la`).
* `/tui <htop|nano|vim>` : Tam ekran interaktif TUI programlarını güvenli devir ile açar.

---

## ⚡ Super User x1000 Eklenti Motoru

Eklentiler (`plugins/` dizinindeki Python, Shell, Go veya Binary dosyaları), MT Core çekirdeğini canlıda manipüle edebilir:

* **Komut Kaydı / Silme:** Eklentiler runtime sırasında sisteme yeni `/komut`lar ekleyebilir veya silebilir.
* **Canlı UI Kontrolü:** Banner, durum mesajı, prompt, altbilgi (footer) ve tam ekran çıktısı eklenti tarafından canlıda değiştirilebilir.
* **Sekme & Girdi Yönetimi:** Sekme başlığı, çıktısı, sekme değiştirme, kaydırma ve imleç girdileri eklentilerce yönetilebilir.
* **Çekirdek İnceleme (Kernel Introspection):** Tüm modüller, kancalar ve çekirdek durumu JSON formatında sorgulanabilir.

### Örnek Python Eklentisi (`plugins/hello.py`):
```python
# @command: hello
# @title: SAY HELLO
# @category: FUN
# @desc: Greets the user in Super User fashion

print("Hello world! Welcome to MT Core Super User x1000.")
```

---

## 📖 Yapay Zeka (AI) ve Geliştirici Eklenti Rehberi
Eklenti geliştirme kuralları, başlık etiketleri, IPC protokolü, SDK referansı ve otonom AI ajanları için rehberin tamamı için [**PLUGIN_README.md**](PLUGIN_README.md) dosyasına göz atabilirsiniz.

