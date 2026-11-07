package framebuffer

import (
	"github.com/mattn/go-runewidth"
)

// IsWideRune determines if a rune occupies 2 terminal cells (East Asian Wide / Fullwidth, Emoji, etc.).
func IsWideRune(r rune) bool {
	return runewidth.RuneWidth(r) == 2
}

// IsZeroWidthRune checks if a rune takes up zero visual columns (e.g. control chars, zero-width space, format runes).
func IsZeroWidthRune(r rune) bool {
	return runewidth.RuneWidth(r) == 0
}

// RuneDisplayWidth returns the visual column width of a rune on standard terminal emulators (0, 1, or 2).
func RuneDisplayWidth(r rune) int {
	w := runewidth.RuneWidth(r)
	if w < 0 {
		return 0
	}
	return w
}
