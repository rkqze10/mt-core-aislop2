#!/usr/bin/env bash
# @command: canvas
# @title: PİKSEL ÇİZİMİ
# @category: DEMO
# @desc: 2D sanal piksel tamponuna dinamik geometrik çizim yapar
# @author: core-team
# @version: 1.5.0
# @icon: 🎨

source "$(dirname "$0")/sdk/mt.sh" 2>/dev/null

mt_box 2 2 36 6 "CANVAS MODU" "#00FF88"
mt_pixel 4 3 "⚡" "#00F0FF"
mt_pixel 6 3 "★" "#FF9E3B"
mt_pixel 8 3 "●" "#FF007F"
mt_pixel 10 3 "▲" "#00FF88"

mt_status "🎨 [Sanal Piksel Motoru] Tuvale 2D geometrik şekiller çizildi!" false
mt_toast "Tuval pikselleri güncellendi!" false

echo "=================================================="
echo "  2D SANAL FRAMEBUFFER TUVALİ ÇİZİLDİ! "
echo "=================================================="
echo "Sol üst köşedeki (2, 2) koordinatına kutu ve renkli"
echo "pikseller basıldı. Temizlemek için /clear_pixels yazabilirsiniz."
