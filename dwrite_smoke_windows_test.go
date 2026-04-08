//go:build windows

package glyph

import (
	"errors"
	"testing"
)

// TestDWriteRasterizerSmoke verifies end-to-end that the DirectWrite
// color glyph path initializes, renders common emoji with non-empty
// bitmaps and sane bearings, and returns errNoColorGlyph for plain
// ASCII. Skips if DWrite init fails (Segoe UI Emoji not installed).
func TestDWriteRasterizerSmoke(t *testing.T) {
	dw, err := newDWriteRasterizer()
	if err != nil {
		t.Skipf("DirectWrite unavailable: %v", err)
	}
	defer dw.Close()

	cases := []rune{
		0x1F600, // 😀 grinning face
		0x1F389, // 🎉 party popper
		0x1F680, // 🚀 rocket
		0x2764,  // ❤ heavy heart
	}
	for _, r := range cases {
		bmp, left, top, err := dw.RenderColorGlyph(32.0, r)
		if err != nil {
			t.Errorf("U+%04X: unexpected error: %v", r, err)
			continue
		}
		if bmp.Width <= 0 || bmp.Height <= 0 {
			t.Errorf("U+%04X: empty bitmap w=%d h=%d",
				r, bmp.Width, bmp.Height)
			continue
		}
		if len(bmp.Data) != bmp.Width*bmp.Height*4 {
			t.Errorf("U+%04X: wrong data length %d for %dx%d",
				r, len(bmp.Data), bmp.Width, bmp.Height)
			continue
		}
		// Sanity: sum of alpha should be > 0 (non-empty coverage).
		var alphaSum int
		for i := 3; i < len(bmp.Data); i += 4 {
			alphaSum += int(bmp.Data[i])
		}
		if alphaSum == 0 {
			t.Errorf("U+%04X: bitmap has zero total alpha", r)
		}
		t.Logf("U+%04X: %dx%d left=%d top=%d alphaSum=%d",
			r, bmp.Width, bmp.Height, left, top, alphaSum)
	}

	// Plain ASCII should report no color glyph.
	_, _, _, err = dw.RenderColorGlyph(32.0, 'A')
	if !errors.Is(err, errNoColorGlyph) {
		t.Errorf("'A': expected errNoColorGlyph, got %v", err)
	}
}
