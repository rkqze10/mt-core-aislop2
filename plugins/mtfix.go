package main

// @command: mtfix
// @title: MTFIX & GOD MODE (LIVE CORE MUTATOR)
// @category: GOD-MODE
// @desc: GUI sıfırlar, MTfix ana kod mutasyon motorunu ve God Mode'u açar
// @author: root-hypervisor
// @version: 2.0.0
// @icon: ⚡
// @aliases: god, godmode, fix, patch
// @perm: fs, net, stream, eventbus, daemon
// @tags: godmode, mtfix, root, patch, live-mutator
// @cmd:god:GOD MODE KONTROL:GOD-MODE:God Mode merkezi ve telemetri
// @cmd:mutate:ARAYÜZ MUTASYONU:GOD-MODE:Canlı tema, prompt, banner değiştirme
// @cmd:patchcmd:KOMUT YAMALAMA:GOD-MODE:Komut yönlendirme ve yeniden yazma
// @cmd:source:KOD GÖRÜNTÜLE:GOD-MODE:Ana kaynak kod dosyalarını incele

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GodModeBanner is the glorious ASCII banner for God Mode & MTfix
const GodModeBanner = `  ██████╗  ██████╗ ██████╗     ███╗   ███╗ ██████╗ ██████╗ ███████╗
 ██╔════╝ ██╔═══██╗██╔══██╗    ████╗ ████║██╔═══██╗██╔══██╗██╔════╝
 ██║  ███╗██║   ██║██║  ██║    ██╔████╔██║██║   ██║██║  ██║█████╗  
 ██║   ██║██║   ██║██║  ██║    ██║╚██╔╝██║██║   ██║██║  ██║██╔══╝  
 ╚██████╔╝╚██████╔╝██████╔╝    ██║ ╚═╝ ██║╚██████╔╝██████╔╝███████╗
  ╚═════╝  ╚═════╝ ╚═════╝     ╚═╝     ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝
    ⚡ GOD MODE v9999 & MTFIX // SÜPER KULLANICI ÇEKİRDEK MUTATÖRÜ ⚡`

func emitIPC(action map[string]any) {
	bytes, err := json.Marshal(action)
	if err == nil {
		fmt.Println(string(bytes))
	}
}

func getWorkspaceRoot() string {
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := os.Stat(filepath.Join(cwd, "main.go")); err == nil {
			return cwd
		}
	}
	execPath, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(execPath)
		if _, err := os.Stat(filepath.Join(dir, "main.go")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if _, err := os.Stat(filepath.Join(parent, "main.go")); err == nil {
			return parent
		}
	}
	// Termux home fallback
	if prefix := os.Getenv("PREFIX"); prefix != "" {
		home := os.Getenv("HOME")
		if home != "" {
			if _, err := os.Stat(filepath.Join(home, "mt-core-ai-", "main.go")); err == nil {
				return filepath.Join(home, "mt-core-ai-")
			}
			if _, err := os.Stat(filepath.Join(home, "main.go")); err == nil {
				return home
			}
		}
	}
	return "/app/applet"
}

func getGoCompiler() string {
	if prefix := os.Getenv("PREFIX"); prefix != "" {
		cand := filepath.Join(prefix, "bin", "go")
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	if p, err := exec.LookPath("go"); err == nil {
		return p
	}
	return "go"
}

func main() {
	args := os.Args[1:]

	// If invoked without subcommands, activate GOD MODE & open MTFIX
	if len(args) == 0 || args[0] == "open" || args[0] == "init" {
		activateGodMode()
		return
	}

	subCmd := strings.ToLower(args[0])
	subArgs := args[1:]

	switch subCmd {
	case "status", "info":
		showStatus()
	case "list", "ls":
		listFiles()
	case "view", "cat", "show":
		viewSource(subArgs)
	case "patch", "edit", "replace":
		patchSource(subArgs)
	case "append":
		appendSource(subArgs)
	case "revert", "restore":
		revertSource(subArgs)
	case "build", "compile", "check":
		buildCore()
	case "mutate":
		mutateUI(subArgs)
	case "patchcmd", "reroute":
		patchCommand(subArgs)
	case "mem", "kv":
		manageMemory(subArgs)
	case "matrix", "rain":
		triggerMatrixRain()
	case "reset":
		resetGUI()
	case "help":
		showHelp()
	default:
		// Unknown subcommand or direct file target
		if _, err := os.Stat(subCmd); err == nil || strings.HasSuffix(subCmd, ".go") {
			viewSource([]string{subCmd})
		} else {
			fmt.Printf("⚠️ Bilinmeyen MTfix parametresi: '%s'. Yardım için: /godhelp veya /mtfix help\n", subCmd)
		}
	}
}

// activateGodMode resets the GUI and unlocks the God Mode environment
func activateGodMode() {
	// 1. Reset GUI completely
	emitIPC(map[string]any{"action": "reset_gui"})
	emitIPC(map[string]any{"action": "close_all_tabs"})
	emitIPC(map[string]any{"action": "clear_pixels"})

	// 2. Apply High-Contrast God Mode Cyber Theme
	emitIPC(map[string]any{"action": "theme", "name": "cyber"})
	emitIPC(map[string]any{"action": "set_theme_color", "field": "primary", "color": "#00FFCC"})
	emitIPC(map[string]any{"action": "set_theme_color", "field": "secondary", "color": "#FF007F"})
	emitIPC(map[string]any{"action": "set_theme_color", "field": "border_color", "color": "#7928CA"})
	emitIPC(map[string]any{"action": "set_theme_color", "field": "badge_bg", "color": "#1F1A3A"})

	// 3. Inject glowing God Mode banner, prompt and footer
	emitIPC(map[string]any{
		"action": "set_banner",
		"banner": GodModeBanner,
	})

	emitIPC(map[string]any{
		"action":    "set_prompt",
		"prefix":    "⚡ GOD-MODE ",
		"arrow":     "❯❯ ",
		"fg_prefix": "#FF007F",
		"fg_arrow":  "#00FFCC",
	})

	emitIPC(map[string]any{
		"action": "set_footer",
		"footer": "⚡ [GOD MODE] /mtfix (kod tamir) | /god (durum) | /mutate (arayüz) | /patchcmd (yönlendir) | /source (dosya oku) | /mem (bellek) | /matrix | /godhelp",
	})

	emitIPC(map[string]any{
		"action":   "set_status",
		"message":  "⚡ [GOD MODE AKTİF] MTfix Canlı Çekirdek ve Kaynak Kod Değiştirici Devrede (ROOT-0 Sınırsız İzin).",
		"is_error": false,
	})

	emitIPC(map[string]any{
		"action":   "toast",
		"message":  "🔥 GOD MODE VE MTFIX AÇILDI! Tüm kilitler kaldırıldı, yeni komutlar eklendi.",
		"is_error": false,
	})

	// 4. Injected dynamic commands
	binPath := filepath.Join(getWorkspaceRoot(), ".mt_bin", "mtfix")
	if _, err := os.Stat(binPath); err != nil {
		binPath = filepath.Join(getWorkspaceRoot(), "plugins", "mtfix.go")
	}

	commands := []struct {
		Cmd   string
		Title string
		Cat   string
		Desc  string
		Exec  string
	}{
		{"god", "GOD MODE KONTROL MERKEZİ", "GOD", "God Mode durum paneli ve sistem telemetrisi", binPath + " status"},
		{"mutate", "CANLI ARAYÜZ MUTASYONU", "GOD", "Arayüz öğelerini değiştir (/mutate banner|prompt|theme|footer)", binPath + " mutate"},
		{"patchcmd", "KOMUT YAMALAMA & REROUTE", "GOD", "Bir komutu canlıda başka komuta yönlendir", binPath + " patchcmd"},
		{"source", "KAYNAK KOD GÖRÜNTÜLEYİCİ", "GOD", "MT Core ana kaynak kodlarını satır satır görüntüle", binPath + " view"},
		{"mem", "PAYLAŞIMLI BELLEK YÖNETİMİ", "GOD", "Shared KV store durumunu oku, yaz, sil", binPath + " mem"},
		{"matrix", "SANAL MATRIX YAĞMURU", "GOD", "Sanal piksel katmanında siber matrix yağmuru oluşturur", binPath + " matrix"},
		{"godhelp", "GOD MODE YARDIM REHBERİ", "GOD", "Tüm God Mode komutlarının kullanım kılavuzu", binPath + " help"},
		{"reboot", "KAYNAK DERLE & YENİLE", "GOD", "Ana Go kodunu derler ve modülleri yeniden yükler", binPath + " build"},
	}

	for _, c := range commands {
		emitIPC(map[string]any{
			"action":   "register_cmd",
			"command":  c.Cmd,
			"title":    c.Title,
			"category": c.Cat,
			"desc":     c.Desc,
			"exec":     c.Exec,
		})
	}

	// 5. Output rich visual interactive dashboard in tab
	printDashboard()
}

func printDashboard() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║               ⚡ MTFIX LIVE CODE MUTATOR & GOD MODE HYPERVISOR v9999 ⚡                    ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Arayüz sıfırlandı, süper kullanıcı yetkileri (ROOT-0) çekirdeğe enjekte edildi.         ║")
	fmt.Println("║  Artık MT Core'un kaynak kodlarını, arayüzünü ve çekirdek fonksiyonlarını değiştirebilirsin.║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("🔥 YENİ ENJEKTE EDİLEN GOD MODE KOMUTLARI:")
	fmt.Println("  ╭────────────────┬──────────────────────────┬──────────────────────────────────────────╮")
	fmt.Println("  │ Komut          │ Kategori                 │ Açıklama                                 │")
	fmt.Println("  ├────────────────┼──────────────────────────┼──────────────────────────────────────────┤")
	fmt.Println("  │ /mtfix patch   │ [KAYNAK KOD DEĞİŞTİRİCİ] │ Ana Go dosyalarındaki kodları değiştirir │")
	fmt.Println("  │ /mtfix view    │ [KOD İNCELEME]           │ Ana kodları satır numaralarıyla gösterir │")
	fmt.Println("  │ /mtfix list    │ [DOSYA TARAYICI]         │ Tüm sistem ve eklenti dosyalarını listeler│")
	fmt.Println("  │ /mtfix build   │ [CANLI DERLEYİCİ]        │ Değiştirilen ana kodu derler ve sınar    │")
	fmt.Println("  │ /mtfix revert  │ [GERİ ALMA]              │ Dosyayı otomatik yedekten (.bak) onarır  │")
	fmt.Println("  │ /god           │ [KONTROL MERKEZİ]        │ Çekirdek telemetrisi ve bellek paneli    │")
	fmt.Println("  │ /mutate        │ [ARAYÜZ MUTASYONU]       │ Banner, prompt, tema ve renkleri değişir │")
	fmt.Println("  │ /patchcmd      │ [KOMUT YAMALAMA]         │ İstediğin komutu başkasına yönlendirir   │")
	fmt.Println("  │ /source        │ [KOD GÖRÜNTÜLE]          │ /source <dosya> ile kod oku              │")
	fmt.Println("  │ /mem           │ [KERNEL BELLEĞİ]         │ /mem set|get|dump ile bellek oku/yaz     │")
	fmt.Println("  │ /matrix        │ [SANAL EKRAN]            │ Framebuffer üzerinde Matrix yağmuru      │")
	fmt.Println("  │ /godhelp       │ [YARDIM]                 │ Tüm komutların detaylı kullanım kılavuzu │")
	fmt.Println("  ╰────────────────┴──────────────────────────┴──────────────────────────────────────────╯")
	fmt.Println()

	fmt.Println("🛠️ MTFIX KULLANIM ÖRNEKLERİ:")
	fmt.Println("  1. Dosyaları listele:          /mtfix list")
	fmt.Println("  2. main.go kodunu incele:      /mtfix view main.go")
	fmt.Println("  3. Ana kodda metin değiştir:   /mtfix patch main.go \"EskiMetin\" \"YeniMetin\"")
	fmt.Println("  4. Arayüzü kırmızı yap:        /mutate color primary #FF0055")
	fmt.Println("  5. Komut yönlendir (help->sys):/patchcmd help sys")
	fmt.Println("  6. Belleğe değişken yaz:       /mem set GOD_STATUS OVERLORD")
	fmt.Println("  7. Matrix efekti başlat:       /matrix")
	fmt.Println()
	fmt.Println("👉 Hemen alttaki komut satırına bir komut yazarak başlayabilirsiniz!")
}

func showStatus() {
	root := getWorkspaceRoot()
	files, _ := filepath.Glob(filepath.Join(root, "pkg", "*", "*.go"))
	pluginFiles, _ := filepath.Glob(filepath.Join(root, "plugins", "*.*"))

	var memStats map[string]string
	stateFile := filepath.Join(root, ".mt_state.json")
	if data, err := os.ReadFile(stateFile); err == nil {
		_ = json.Unmarshal(data, &memStats)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║               ⚡ GOD MODE KERNEL & SYSTEM TELEMETRY (RING-0) ⚡                 ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Durum:            [AKTİF - TAM ERİŞİM] Sınırsız Çekirdek Hakları             ║\n")
	fmt.Printf("║  Ana Kod Dosyaları: %-58d ║\n", len(files)+1)
	fmt.Printf("║  Yüklü Eklentiler:  %-58d ║\n", len(pluginFiles))
	fmt.Printf("║  Paylaşımlı Bellek: %-58d ║\n", len(memStats))
	fmt.Printf("║  Çalışma Dizini:    %-58s ║\n", root)
	fmt.Printf("║  Zaman Damgası:     %-58s ║\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(memStats) > 0 {
		fmt.Println("🧠 PAYLAŞIMLI BELLEKTEKİ DEĞİŞKENLER:")
		for k, v := range memStats {
			if len(v) > 50 {
				v = v[:47] + "..."
			}
			fmt.Printf(" • %-20s : %s\n", k, v)
		}
		fmt.Println()
	}

	fmt.Println("💡 İpucu: /mtfix patch <dosya> <eski> <yeni> ile ana kodu canlı değiştirebilirsiniz.")
}

func listFiles() {
	root := getWorkspaceRoot()
	fmt.Println("📂 MT CORE PROJE VE KAYNAK KOD DOSYALARI:")
	fmt.Println("──────────────────────────────────────────────────────────────────────────")

	type fileEntry struct {
		Category string
		Path     string
		Lines    int
		Size     int64
	}

	var entries []fileEntry

	countLines := func(p string) (int, int64) {
		f, err := os.Open(p)
		if err != nil {
			return 0, 0
		}
		defer f.Close()
		lines := 0
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			lines++
		}
		fi, _ := f.Stat()
		return lines, fi.Size()
	}

	// Main & core
	if l, s := countLines(filepath.Join(root, "main.go")); l > 0 {
		entries = append(entries, fileEntry{"CORE", "main.go", l, s})
	}
	if l, s := countLines(filepath.Join(root, "go.mod")); l > 0 {
		entries = append(entries, fileEntry{"CONFIG", "go.mod", l, s})
	}

	// Packages
	_ = filepath.Walk(filepath.Join(root, "pkg"), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".go") {
			rel, _ := filepath.Rel(root, path)
			cat := "PKG"
			if strings.Contains(rel, "ui") {
				cat = "UI"
			} else if strings.Contains(rel, "kernel") {
				cat = "KERNEL"
			} else if strings.Contains(rel, "superuser") {
				cat = "SUPERUSER"
			} else if strings.Contains(rel, "builtins") {
				cat = "BUILTIN"
			}
			l, s := countLines(path)
			entries = append(entries, fileEntry{cat, rel, l, s})
		}
		return nil
	})

	// Plugins
	_ = filepath.Walk(filepath.Join(root, "plugins"), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(root, path)
			l, s := countLines(path)
			entries = append(entries, fileEntry{"PLUGIN", rel, l, s})
		}
		return nil
	})

	fmt.Printf("%-10s %-36s %-10s %-10s\n", "KATEGORİ", "DOSYA YOLU", "SATIR", "BOYUT")
	fmt.Println("──────────────────────────────────────────────────────────────────────────")
	for _, e := range entries {
		fmt.Printf("%-10s %-36s %-10d %d B\n", "["+e.Category+"]", e.Path, e.Lines, e.Size)
	}
	fmt.Println("──────────────────────────────────────────────────────────────────────────")
	fmt.Println("Dosya incelemek için: /mtfix view <dosya_adı>")
	fmt.Println("Dosya değiştirmek için: /mtfix patch <dosya_adı> \"eski_kod\" \"yeni_kod\"")
}

func viewSource(args []string) {
	if len(args) == 0 {
		fmt.Println("⚠️ Kullanım: /mtfix view <dosya_adı> [başlangıç_satırı] [bitiş_satırı]")
		fmt.Println("Örnek: /mtfix view main.go")
		fmt.Println("Örnek: /mtfix view pkg/ui/view.go 1 50")
		return
	}

	root := getWorkspaceRoot()
	targetPath := args[0]
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(root, targetPath)
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		fmt.Printf("⚠️ Dosya okunamadı: %v\n", err)
		return
	}

	lines := strings.Split(string(data), "\n")
	startLine := 1
	endLine := len(lines)

	if len(args) >= 2 {
		if s, err := strconv.Atoi(args[1]); err == nil && s >= 1 {
			startLine = s
		}
	}
	if len(args) >= 3 {
		if e, err := strconv.Atoi(args[2]); err == nil && e >= startLine && e <= len(lines) {
			endLine = e
		}
	}

	rel, _ := filepath.Rel(root, targetPath)
	fmt.Printf("📄 DOSYA: %s (Toplam %d satır, gösterilen: %d-%d)\n", rel, len(lines), startLine, endLine)
	fmt.Println("──────────────────────────────────────────────────────────────────────────")
	for i := startLine - 1; i < endLine && i < len(lines); i++ {
		fmt.Printf("%4d │ %s\n", i+1, lines[i])
	}
	fmt.Println("──────────────────────────────────────────────────────────────────────────")
}

func patchSource(args []string) {
	if len(args) < 3 {
		fmt.Println("⚠️ Kullanım: /mtfix patch <dosya_adı> \"<aranacak_kod>\" \"<yeni_kod>\"")
		fmt.Println("Örnek: /mtfix patch main.go \"mt-tui\" \"mt-tui-godmode\"")
		return
	}

	root := getWorkspaceRoot()
	targetPath := args[0]
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(root, targetPath)
	}

	oldStr := args[1]
	newStr := args[2]

	contentBytes, err := os.ReadFile(targetPath)
	if err != nil {
		fmt.Printf("⚠️ Hata: Dosya bulunamadı (%s): %v\n", targetPath, err)
		return
	}

	content := string(contentBytes)
	if !strings.Contains(content, oldStr) {
		fmt.Printf("⚠️ Hata: Değiştirilmek istenen metin dosyada bulunamadı!\nAranan:\n%s\n", oldStr)
		return
	}

	// 1. Create automatic backup .bak
	bakPath := targetPath + ".bak"
	if err := os.WriteFile(bakPath, contentBytes, 0644); err != nil {
		fmt.Printf("⚠️ Yedek dosyası oluşturulamadı: %v\n", err)
		return
	}

	// 2. Perform replacement
	newContent := strings.Replace(content, oldStr, newStr, 1)
	if err := os.WriteFile(targetPath, []byte(newContent), 0644); err != nil {
		fmt.Printf("⚠️ Dosya yazılamadı: %v\n", err)
		return
	}

	// 3. Test Go build if it is a .go file
	rel, _ := filepath.Rel(root, targetPath)
	if strings.HasSuffix(targetPath, ".go") {
		fmt.Printf("⏳ [%s] güncellendi. Go derleme doğrulaması yapılıyor...\n", rel)
		checkCmd := exec.Command(getGoCompiler(), "build", "-o", "/tmp/mt-syntax-check", filepath.Join(root, "main.go"))
		checkCmd.Dir = root
		buildOut, buildErr := checkCmd.CombinedOutput()
		if buildErr != nil {
			// Auto revert to protect the core!
			_ = os.WriteFile(targetPath, contentBytes, 0644)
			fmt.Printf("❌ DERLEME HATASI! Değişiklik otomatik geri alındı (Rollback):\n%s\n", string(buildOut))
			return
		}
		_ = os.Remove("/tmp/mt-syntax-check")
	}

	emitIPC(map[string]any{
		"action":   "toast",
		"message":  fmt.Sprintf("✔ [%s] başarıyla yamalandı ve derlendi!", rel),
		"is_error": false,
	})

	fmt.Printf("✔ [MTFIX BAŞARILI] %s dosyası başarıyla yamalandı!\n", rel)
	fmt.Printf(" • Eski Metin: %s\n", oldStr)
	fmt.Printf(" • Yeni Metin: %s\n", newStr)
	fmt.Printf(" • Otomatik Yedek: %s.bak\n", rel)
	fmt.Println(" • Doğrulama: Go derleme kontrolü BAŞARILI.")
	fmt.Println("Değişikliği geri almak isterseniz: /mtfix revert " + rel)
}

func appendSource(args []string) {
	if len(args) < 2 {
		fmt.Println("⚠️ Kullanım: /mtfix append <dosya_adı> \"<eklenecek_kod>\"")
		return
	}

	root := getWorkspaceRoot()
	targetPath := args[0]
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(root, targetPath)
	}
	codeToAppend := strings.Join(args[1:], " ")

	contentBytes, err := os.ReadFile(targetPath)
	if err != nil {
		fmt.Printf("⚠️ Hata: Dosya okunamadı: %v\n", err)
		return
	}

	bakPath := targetPath + ".bak"
	_ = os.WriteFile(bakPath, contentBytes, 0644)

	newContent := string(contentBytes) + "\n\n" + codeToAppend + "\n"
	if err := os.WriteFile(targetPath, []byte(newContent), 0644); err != nil {
		fmt.Printf("⚠️ Hata: Dosyaya yazılamadı: %v\n", err)
		return
	}

	rel, _ := filepath.Rel(root, targetPath)
	fmt.Printf("✔ [MTFIX] %s dosyasının sonuna kod eklendi. Yedek: %s.bak\n", rel, rel)
}

func revertSource(args []string) {
	if len(args) == 0 {
		fmt.Println("⚠️ Kullanım: /mtfix revert <dosya_adı>")
		return
	}

	root := getWorkspaceRoot()
	targetPath := args[0]
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(root, targetPath)
	}

	bakPath := targetPath + ".bak"
	bakBytes, err := os.ReadFile(bakPath)
	if err != nil {
		fmt.Printf("⚠️ Yedek dosyası bulunamadı (%s.bak): %v\n", targetPath, err)
		return
	}

	if err := os.WriteFile(targetPath, bakBytes, 0644); err != nil {
		fmt.Printf("⚠️ Dosya geri yüklenemedi: %v\n", err)
		return
	}

	rel, _ := filepath.Rel(root, targetPath)
	fmt.Printf("✔ [ROLLBACK BAŞARILI] %s dosyası .bak yedeğinden geri yüklendi.\n", rel)
}

func buildCore() {
	root := getWorkspaceRoot()
	outBin := filepath.Join(root, "mt-core")

	fmt.Println("⏳ MT Core ana çekirdek kodu derleniyor...")
	startTime := time.Now()

	cmd := exec.Command(getGoCompiler(), "build", "-o", outBin, filepath.Join(root, "main.go"))
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	dur := time.Since(startTime)

	if err != nil {
		fmt.Printf("❌ DERLEME HATASI:\n%s\n", string(out))
		return
	}

	fi, _ := os.Stat(outBin)
	sizeMB := float64(fi.Size()) / (1024 * 1024)

	emitIPC(map[string]any{
		"action":   "toast",
		"message":  fmt.Sprintf("✔ Çekirdek başarıyla derlendi (%.2fs | %.1f MB)", dur.Seconds(), sizeMB),
		"is_error": false,
	})

	fmt.Println("✔ [DERLEME BAŞARILI]")
	fmt.Printf(" • Çıktı İkili Dosyası: %s\n", outBin)
	fmt.Printf(" • Derleme Süresi:      %v\n", dur)
	fmt.Printf(" • İkili Boyutu:         %.2f MB\n", sizeMB)
	fmt.Println("Sistem hazır durumda!")
}

func mutateUI(args []string) {
	if len(args) == 0 {
		fmt.Println("⚠️ Kullanım: /mutate <banner|prompt|theme|color|footer|title|reset> [parametreler]")
		fmt.Println("Örnekler:")
		fmt.Println("  /mutate theme cyber")
		fmt.Println("  /mutate color primary #FF0055")
		fmt.Println("  /mutate prompt \"⚡ ROOT \" \"❯ \"")
		fmt.Println("  /mutate banner \"<<< ÖZEL BAŞLIK >>>\"")
		fmt.Println("  /mutate footer \"[F1] Yardım | [F2] Çıkış\"")
		fmt.Println("  /mutate reset")
		return
	}

	target := strings.ToLower(args[0])
	rest := args[1:]

	switch target {
	case "theme":
		if len(rest) > 0 {
			emitIPC(map[string]any{"action": "theme", "name": rest[0]})
			fmt.Printf("✔ Tema değiştirildi: %s\n", rest[0])
		}
	case "color", "colour":
		if len(rest) >= 2 {
			emitIPC(map[string]any{"action": "set_theme_color", "field": rest[0], "color": rest[1]})
			fmt.Printf("✔ Tema rengi güncellendi: %s -> %s\n", rest[0], rest[1])
		}
	case "prompt":
		pfx := "⚡ ROOT "
		arr := "❯❯ "
		if len(rest) >= 1 {
			pfx = rest[0]
		}
		if len(rest) >= 2 {
			arr = rest[1]
		}
		emitIPC(map[string]any{
			"action":    "set_prompt",
			"prefix":    pfx,
			"arrow":     arr,
			"fg_prefix": "#FF007F",
			"fg_arrow":  "#00FFCC",
		})
		fmt.Printf("✔ Prompt güncellendi: %s%s\n", pfx, arr)
	case "banner":
		if len(rest) > 0 {
			b := strings.Join(rest, " ")
			emitIPC(map[string]any{"action": "set_banner", "banner": b})
			fmt.Println("✔ Üst ASCII banner değiştirildi.")
		}
	case "footer":
		if len(rest) > 0 {
			f := strings.Join(rest, " ")
			emitIPC(map[string]any{"action": "set_footer", "footer": f})
			fmt.Println("✔ Altbilgi kısayolları değiştirildi.")
		}
	case "reset":
		emitIPC(map[string]any{"action": "reset_banner"})
		emitIPC(map[string]any{"action": "reset_footer"})
		emitIPC(map[string]any{"action": "reset_prompt"})
		emitIPC(map[string]any{"action": "clear_pixels"})
		fmt.Println("✔ Arayüz varsayılan ayarlara döndürüldü.")
	}
}

func patchCommand(args []string) {
	if len(args) < 2 {
		fmt.Println("⚠️ Kullanım: /patchcmd <hedef_komut> <yönlendirilecek_komut>")
		fmt.Println("Örnek: /patchcmd help sys  (/help yazılınca /sys çalışır)")
		return
	}
	target := args[0]
	rewrite := args[1]
	emitIPC(map[string]any{
		"action":  "patch_cmd",
		"target":  target,
		"rewrite": rewrite,
	})
	fmt.Printf("✔ Komut yönlendirildi: /%s -> /%s\n", target, rewrite)
}

func manageMemory(args []string) {
	if len(args) == 0 {
		fmt.Println("⚠️ Kullanım: /mem <set|get|del|dump> [anahtar] [değer]")
		return
	}
	op := strings.ToLower(args[0])
	root := getWorkspaceRoot()
	stateFile := filepath.Join(root, ".mt_state.json")

	switch op {
	case "set":
		if len(args) >= 3 {
			k := args[1]
			v := strings.Join(args[2:], " ")
			emitIPC(map[string]any{"action": "kv_set", "key": k, "value": v})
			fmt.Printf("✔ Belleğe kaydedildi: %s = %s\n", k, v)
		}
	case "del":
		if len(args) >= 2 {
			emitIPC(map[string]any{"action": "kv_del", "key": args[1]})
			fmt.Printf("✔ Bellekten silindi: %s\n", args[1])
		}
	case "dump", "list":
		data, err := os.ReadFile(stateFile)
		if err != nil {
			fmt.Printf("Bellek boş veya okunamadı: %v\n", err)
			return
		}
		var stateMap map[string]string
		_ = json.Unmarshal(data, &stateMap)
		fmt.Printf("🧠 PAYLAŞIMLI BELLEK DÖKÜMÜ (%d anahtar):\n", len(stateMap))
		fmt.Println("──────────────────────────────────────────────────────────")
		for k, v := range stateMap {
			fmt.Printf(" • %-24s = %s\n", k, v)
		}
		fmt.Println("──────────────────────────────────────────────────────────")
	}
}

func triggerMatrixRain() {
	// Emit 60 FPS matrix characters to virtual framebuffer
	chars := []rune{'0', '1', 'X', 'Z', 'Y', 'λ', 'Ω', 'Ψ', '§', '¶', 'Ξ', '⚡', '★'}
	var pixels []map[string]any

	w := 120
	h := 35
	if rawW := os.Getenv("MT_WIDTH"); rawW != "" {
		if n, err := strconv.Atoi(rawW); err == nil && n > 20 {
			w = n
		}
	}
	if rawH := os.Getenv("MT_HEIGHT"); rawH != "" {
		if n, err := strconv.Atoi(rawH); err == nil && n > 10 {
			h = n
		}
	}

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 80; i++ {
		rx := rand.Intn(w)
		ry := rand.Intn(h)
		rc := string(chars[rand.Intn(len(chars))])
		color := "#00FF66"
		if i%3 == 0 {
			color = "#00FFFF"
		} else if i%5 == 0 {
			color = "#FFFFFF"
		}

		pixels = append(pixels, map[string]any{
			"x":    rx,
			"y":    ry,
			"char": rc,
			"fg":   color,
			"bg":   "#000000",
			"bold": true,
		})
	}

	emitIPC(map[string]any{
		"action": "pixels",
		"pixels": pixels,
	})

	emitIPC(map[string]any{
		"action":   "toast",
		"message":  "⚡ Sanal Matrix Yağmuru aktif edildi!",
		"is_error": false,
	})

	fmt.Println("✔ Sanal Framebuffer üzerine Matrix piksel yağmuru uygulandı.")
	fmt.Println("Temizlemek için: /mutate reset veya /pixel clear")
}

func resetGUI() {
	emitIPC(map[string]any{"action": "reset_gui"})
	emitIPC(map[string]any{"action": "close_all_tabs"})
	emitIPC(map[string]any{"action": "clear_pixels"})
	emitIPC(map[string]any{"action": "theme", "name": "amber"})
	fmt.Println("✔ Arayüz ve sekmeler tamamen sıfırlandı.")
}

func showHelp() {
	fmt.Println("📖 MTFIX & GOD MODE TAM KULLANIM REHBERİ:")
	fmt.Println("═════════════════════════════════════════════════════════════════════════════")
	fmt.Println("1. KOD DEĞİŞTİRME:")
	fmt.Println("   /mtfix list                           -> Kaynak dosyaları listeler")
	fmt.Println("   /mtfix view <dosya> [satır_aralığı]  -> Dosyayı satır numaralarıyla gösterir")
	fmt.Println("   /mtfix patch <dosya> <eski> <yeni>   -> Dosyadaki kodu değiştirir (oto derleme kontrollü)")
	fmt.Println("   /mtfix append <dosya> <kod>          -> Dosya sonuna yeni kod ekler")
	fmt.Println("   /mtfix revert <dosya>                -> .bak yedeğinden geri yükler")
	fmt.Println("   /mtfix build                         -> Ana Go kodunu baştan derler")
	fmt.Println()
	fmt.Println("2. ARAYÜZ MUTASYONLARI:")
	fmt.Println("   /mutate theme <amber|emerald|cyber|mono>")
	fmt.Println("   /mutate color primary <hex_renk>")
	fmt.Println("   /mutate prompt <önek> <ok>")
	fmt.Println("   /mutate banner <metin>")
	fmt.Println("   /mutate footer <kısayol_metni>")
	fmt.Println("   /mutate reset")
	fmt.Println()
	fmt.Println("3. KOMUT & KERNEL KONTROL:")
	fmt.Println("   /patchcmd <hedef> <yönlendirme>      -> Örn: /patchcmd help sys")
	fmt.Println("   /mem set <anahtar> <değer>           -> Paylaşımlı belleğe değişken yaz")
	fmt.Println("   /mem dump                            -> Paylaşımlı belleği dök")
	fmt.Println("   /matrix                              -> Framebuffer'a Matrix yağmuru uygula")
	fmt.Println("   /god                                 -> Sistem durumu ve telemetri")
	fmt.Println("═════════════════════════════════════════════════════════════════════════════")
}
