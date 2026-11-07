package framebuffer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRuneDisplayWidth(t *testing.T) {
	if w := RuneDisplayWidth('a'); w != 1 {
		t.Errorf("RuneDisplayWidth('a') = %d; want 1", w)
	}
	if w := RuneDisplayWidth('🚀'); w != 2 {
		t.Errorf("RuneDisplayWidth('🚀') = %d; want 2", w)
	}
	if w := RuneDisplayWidth('ç'); w != 1 {
		t.Errorf("RuneDisplayWidth('ç') = %d; want 1", w)
	}
	if w := RuneDisplayWidth(0); w != 0 {
		t.Errorf("RuneDisplayWidth(0) = %d; want 0", w)
	}
}

func TestVirtualFramebuffer_DrawBox(t *testing.T) {
	fb := New(10, 5)
	fb.DrawBox(0, 0, 10, 5, "TEST", lipgloss.Color("#FF9E3B"), true)

	c00, _ := fb.GetPixel(0, 0)
	if c00.Char != '╭' {
		t.Errorf("expected '╭' at (0,0), got %c", c00.Char)
	}
	c90, _ := fb.GetPixel(9, 0)
	if c90.Char != '╮' {
		t.Errorf("expected '╮' at (9,0), got %c", c90.Char)
	}
	c04, _ := fb.GetPixel(0, 4)
	if c04.Char != '╰' {
		t.Errorf("expected '╰' at (0,4), got %c", c04.Char)
	}
	c94, _ := fb.GetPixel(9, 4)
	if c94.Char != '╯' {
		t.Errorf("expected '╯' at (9,4), got %c", c94.Char)
	}

	fb.DrawBox(-2, -2, 5, 5, "", nil, false)
}

func TestVirtualFramebuffer_DrawBox_TitleTruncation(t *testing.T) {
	fb := New(10, 5)
	// Long title that exceeds box width
	fb.DrawBox(0, 0, 10, 5, "VERY_LONG_TITLE_EXCEEDING_WIDTH", lipgloss.Color("#FF9E3B"), true)

	// Corner '╮' must remain intact and not overwritten
	c90, _ := fb.GetPixel(9, 0)
	if c90.Char != '╮' {
		t.Errorf("expected corner '╮' to be preserved, got %c", c90.Char)
	}
}

func TestVirtualFramebuffer_DrawTextMultiLine(t *testing.T) {
	fb := New(20, 5)
	fb.DrawText(0, 0, "Line 1\nLine 2", nil, nil, false)

	c00, _ := fb.GetPixel(0, 0)
	if c00.Char != 'L' {
		t.Errorf("expected 'L' at (0,0), got %c", c00.Char)
	}
	c01, _ := fb.GetPixel(0, 1)
	if c01.Char != 'L' {
		t.Errorf("expected 'L' at (0,1), got %c", c01.Char)
	}
}

func TestVirtualFramebuffer_CompositeOnANSI(t *testing.T) {
	fb := New(20, 3)
	fb.SetPixel(2, 1, 'X', lipgloss.Color("#FF0000"), nil, false)

	rawANSI := "Line 0\nLine 1 is here\nLine 2"
	composited := fb.CompositeOnANSI(rawANSI)

	lines := strings.Split(composited, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[1], "X") {
		t.Errorf("expected 'X' in line 1, got %q", lines[1])
	}
}

func TestVirtualFramebuffer_ColoredSpaceActive(t *testing.T) {
	fb := New(10, 2)
	// Space with background color must be recognized as active
	fb.SetPixel(2, 0, ' ', nil, lipgloss.Color("#00FF88"), false)

	rawANSI := "Hello World\nLine 2"
	composited := fb.CompositeOnANSI(rawANSI)
	if composited == rawANSI {
		t.Errorf("expected composited output to differ from raw when background pixel is set")
	}
}
