package framebuffer

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// VirtualFramebuffer provides a 2D grid for pixel-level terminal manipulation.
type VirtualFramebuffer struct {
	Width  int
	Height int
	grid   [][]Cell
}

// New creates a new VirtualFramebuffer with specified dimensions.
func New(w, h int) *VirtualFramebuffer {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	fb := &VirtualFramebuffer{
		Width:  w,
		Height: h,
		grid:   make([][]Cell, h),
	}
	for y := 0; y < h; y++ {
		fb.grid[y] = make([]Cell, w)
	}
	return fb
}

func isCellActive(c Cell) bool {
	return c.Char != 0 && (c.Char != ' ' || c.Bg != nil)
}

// SetPixel modifies a single coordinate in the virtual framebuffer.
func (fb *VirtualFramebuffer) SetPixel(x, y int, ch rune, fg, bg lipgloss.TerminalColor, bold bool) {
	if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
		return
	}
	fb.grid[y][x] = Cell{
		Char: ch,
		Fg:   fg,
		Bg:   bg,
		Bold: bold,
	}
}

// GetPixel retrieves the cell at the given coordinate.
func (fb *VirtualFramebuffer) GetPixel(x, y int) (Cell, bool) {
	if x < 0 || x >= fb.Width || y < 0 || y >= fb.Height {
		return Cell{}, false
	}
	return fb.grid[y][x], true
}

// DrawText prints a string starting at (x, y) with proper multi-column and multi-line handling.
func (fb *VirtualFramebuffer) DrawText(x, y int, text string, fg, bg lipgloss.TerminalColor, bold bool) {
	if y >= fb.Height {
		return
	}
	col := x
	row := y
	for _, r := range text {
		if r == '\r' {
			col = x
			continue
		}
		if r == '\n' {
			col = x
			row++
			if row >= fb.Height {
				break
			}
			continue
		}
		if row < 0 || row >= fb.Height {
			continue
		}

		w := RuneDisplayWidth(r)
		if w == 0 {
			continue
		}

		// Prevent wide characters from overflowing right border and wrapping
		if w == 2 && col+1 >= fb.Width {
			col += w
			continue
		}

		if col >= 0 && col < fb.Width {
			fb.SetPixel(col, row, r, fg, bg, bold)
			if w == 2 && col+1 < fb.Width {
				fb.SetPixel(col+1, row, ' ', fg, bg, false)
			}
		}
		col += w
	}
}

// DrawBox renders a bordered box with an optional title.
func (fb *VirtualFramebuffer) DrawBox(x, y, w, h int, title string, borderFg lipgloss.TerminalColor, bold bool) {
	if x < 0 || y < 0 || w < 2 || h < 2 || x >= fb.Width || y >= fb.Height {
		return
	}
	maxX := x + w - 1
	maxY := y + h - 1

	if maxX >= fb.Width {
		maxX = fb.Width - 1
	}
	if maxY >= fb.Height {
		maxY = fb.Height - 1
	}

	if maxX <= x || maxY <= y {
		return
	}

	fb.SetPixel(x, y, '╭', borderFg, nil, bold)
	fb.SetPixel(maxX, y, '╮', borderFg, nil, bold)
	fb.SetPixel(x, maxY, '╰', borderFg, nil, bold)
	fb.SetPixel(maxX, maxY, '╯', borderFg, nil, bold)

	for curX := x + 1; curX < maxX; curX++ {
		fb.SetPixel(curX, y, '─', borderFg, nil, bold)
		fb.SetPixel(curX, maxY, '─', borderFg, nil, bold)
	}
	for curY := y + 1; curY < maxY; curY++ {
		fb.SetPixel(x, curY, '│', borderFg, nil, bold)
		fb.SetPixel(maxX, curY, '│', borderFg, nil, bold)
	}

	if title != "" && x+2 < maxX {
		// Strictly constrain title to fit within the box width
		maxTitleW := maxX - (x + 2) - 2
		if maxTitleW > 0 {
			tRunes := []rune(title)
			if len(tRunes) > maxTitleW {
				tRunes = tRunes[:maxTitleW]
			}
			t := " " + string(tRunes) + " "
			fb.DrawText(x+2, y, t, borderFg, nil, true)
		}
	}
}

// CompositeOnANSI blends the virtual framebuffer cells onto rendered ANSI text without breaking terminal escape codes.
func (fb *VirtualFramebuffer) CompositeOnANSI(rawScreen string) string {
	lines := strings.Split(rawScreen, "\n")
	var result strings.Builder

	for y := 0; y < fb.Height; y++ {
		var srcLine string
		if y < len(lines) {
			srcLine = lines[y]
		}

		rowHasPixels := false
		maxPixelCol := 0
		for x := 0; x < fb.Width; x++ {
			if isCellActive(fb.grid[y][x]) {
				rowHasPixels = true
				maxPixelCol = x
			}
		}

		if !rowHasPixels {
			result.WriteString(srcLine)
			if y < fb.Height-1 {
				result.WriteRune('\n')
			}
			continue
		}

		var rowOut strings.Builder
		runes := []rune(srcLine)
		rLen := len(runes)
		col := 0
		i := 0

		for i < rLen {
			if runes[i] == 0x1b && i+1 < rLen {
				if runes[i+1] == '[' {
					end := i + 2
					for end < rLen && (runes[end] < 0x40 || runes[end] > 0x7e) {
						end++
					}
					if end < rLen {
						end++
					}
					rowOut.WriteString(string(runes[i:end]))
					i = end
					continue
				} else if runes[i+1] == ']' {
					end := i + 2
					for end < rLen {
						if runes[end] == 0x07 {
							end++
							break
						}
						if runes[end] == 0x1b && end+1 < rLen && runes[end+1] == '\\' {
							end += 2
							break
						}
						end++
					}
					rowOut.WriteString(string(runes[i:end]))
					i = end
					continue
				} else {
					end := i + 2
					if end <= rLen {
						rowOut.WriteString(string(runes[i:end]))
						i = end
						continue
					}
				}
			}

			w := RuneDisplayWidth(runes[i])
			if w == 0 {
				rowOut.WriteRune(runes[i])
				i++
				continue
			}

			if col < fb.Width && isCellActive(fb.grid[y][col]) {
				c := fb.grid[y][col]
				st := lipgloss.NewStyle()
				if c.Fg != nil {
					st = st.Foreground(c.Fg)
				}
				if c.Bg != nil {
					st = st.Background(c.Bg)
				}
				if c.Bold {
					st = st.Bold(true)
				}
				rowOut.WriteString(st.Render(string(c.Char)))

				pixelW := RuneDisplayWidth(c.Char)
				if w == 2 && pixelW == 1 {
					rowOut.WriteRune(' ')
				}
				col += w
			} else {
				rowOut.WriteRune(runes[i])
				col += w
			}
			i++
		}

		for col <= maxPixelCol && col < fb.Width {
			if isCellActive(fb.grid[y][col]) {
				c := fb.grid[y][col]
				st := lipgloss.NewStyle()
				if c.Fg != nil {
					st = st.Foreground(c.Fg)
				}
				if c.Bg != nil {
					st = st.Background(c.Bg)
				}
				if c.Bold {
					st = st.Bold(true)
				}
				rowOut.WriteString(st.Render(string(c.Char)))
				pw := RuneDisplayWidth(c.Char)
				if pw < 1 {
					pw = 1
				}
				col += pw
			} else {
				rowOut.WriteRune(' ')
				col++
			}
		}

		result.WriteString(rowOut.String())
		if y < fb.Height-1 {
			result.WriteRune('\n')
		}
	}

	return result.String()
}
