//go:build !js && !ios && !android && !windows

package glyph

import "testing"

func TestContextCreation(t *testing.T) {
	ctx, err := NewContext(1.0)
	if err != nil {
		t.Fatalf("NewContext: %v", err)
	}
	ctx.Free()
}

func TestFontHeightSanity(t *testing.T) {
	ctx, err := NewContext(1.0)
	if err != nil {
		t.Skip("Pango/FreeType not available")
	}
	defer ctx.Free()

	cfg := TextConfig{
		Style: TextStyle{FontName: "Sans 20"},
	}
	h, err := ctx.FontHeight(cfg)
	if err != nil {
		t.Fatalf("FontHeight: %v", err)
	}
	if h < 15.0 || h > 40.0 {
		t.Errorf("Sans 20 height=%f, want 15-40", h)
	}
}

func TestFontHeightPixels(t *testing.T) {
	ctx, err := NewContext(1.0)
	if err != nil {
		t.Skip("Pango/FreeType not available")
	}
	defer ctx.Free()

	cfg := TextConfig{
		Style: TextStyle{FontName: "Sans 20px"},
	}
	h, err := ctx.FontHeight(cfg)
	if err != nil {
		t.Fatalf("FontHeight: %v", err)
	}
	if h < 18.0 || h > 30.0 {
		t.Errorf("Sans 20px height=%f, want 18-30", h)
	}
}

func TestFontMetrics(t *testing.T) {
	ctx, err := NewContext(1.0)
	if err != nil {
		t.Skip("Pango/FreeType not available")
	}
	defer ctx.Free()

	cfg := TextConfig{
		Style: TextStyle{FontName: "Sans 20"},
	}
	m, err := ctx.FontMetrics(cfg)
	if err != nil {
		t.Fatalf("FontMetrics: %v", err)
	}
	if m.Ascender <= 0 {
		t.Errorf("ascender=%f, want > 0", m.Ascender)
	}
	if m.Descender <= 0 {
		t.Errorf("descender=%f, want > 0", m.Descender)
	}
	if m.Height <= 0 {
		t.Errorf("height=%f, want > 0", m.Height)
	}
}

func TestFontHeightCaching(t *testing.T) {
	ctx, err := NewContext(1.0)
	if err != nil {
		t.Skip("Pango/FreeType not available")
	}
	defer ctx.Free()

	cfg := TextConfig{
		Style: TextStyle{FontName: "Sans 20"},
	}
	h1, err := ctx.FontHeight(cfg)
	if err != nil {
		t.Fatalf("FontHeight first call: %v", err)
	}
	h2, err := ctx.FontHeight(cfg)
	if err != nil {
		t.Fatalf("FontHeight second call: %v", err)
	}
	if h1 != h2 {
		t.Errorf("cached mismatch: %f != %f", h1, h2)
	}
}

func TestResolveFontName(t *testing.T) {
	ctx, err := NewContext(1.0)
	if err != nil {
		t.Skip("Pango/FreeType not available")
	}
	defer ctx.Free()

	name, err := ctx.ResolveFontName("Sans 12")
	if err != nil {
		t.Fatalf("ResolveFontName: %v", err)
	}
	if name == "" {
		t.Error("resolved name is empty")
	}
}
