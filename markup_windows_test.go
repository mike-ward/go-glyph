//go:build windows

package glyph

import "testing"

func TestMarkupParseHexColor6(t *testing.T) {
	c, ok := markupParseHexColor("#FF8000")
	if !ok {
		t.Fatal("expected ok")
	}
	if c.R != 0xFF || c.G != 0x80 || c.B != 0x00 || c.A != 255 {
		t.Errorf("got %+v", c)
	}
}

func TestMarkupParseHexColor3(t *testing.T) {
	c, ok := markupParseHexColor("#F80")
	if !ok {
		t.Fatal("expected ok for #RGB")
	}
	if c.R != 0xFF || c.G != 0x88 || c.B != 0x00 || c.A != 255 {
		t.Errorf("got %+v", c)
	}
}

func TestMarkupParseHexColor8(t *testing.T) {
	c, ok := markupParseHexColor("#FF800080")
	if !ok {
		t.Fatal("expected ok for #RRGGBBAA")
	}
	if c.R != 0xFF || c.G != 0x80 || c.B != 0x00 || c.A != 0x80 {
		t.Errorf("got %+v", c)
	}
}

func TestMarkupParseHexColor4(t *testing.T) {
	c, ok := markupParseHexColor("#F808")
	if !ok {
		t.Fatal("expected ok for #RGBA")
	}
	if c.R != 0xFF || c.G != 0x88 || c.B != 0x00 || c.A != 0x88 {
		t.Errorf("got %+v", c)
	}
}

func TestMarkupParseHexColorInvalid(t *testing.T) {
	for _, s := range []string{"", "#", "#GG0000", "#12345", "red"} {
		if _, ok := markupParseHexColor(s); ok {
			t.Errorf("expected failure for %q", s)
		}
	}
}

func TestMarkupNamedSize(t *testing.T) {
	tests := []struct {
		name string
		want float32
	}{
		{"xx-small", 6.9},
		{"small", 10},
		{"medium", 12},
		{"large", 14.4},
		{"xx-large", 20.7},
	}
	for _, tt := range tests {
		sz, ok := markupNamedSize(tt.name)
		if !ok || sz != tt.want {
			t.Errorf("markupNamedSize(%q) = %v, %v; want %v", tt.name, sz, ok, tt.want)
		}
	}
	if _, ok := markupNamedSize("bogus"); ok {
		t.Error("expected failure for unknown size name")
	}
}

func TestMarkupApplyBoldItalic(t *testing.T) {
	base := TextStyle{Typeface: TypefaceRegular}

	b := markupApplyBold(base)
	if b.Typeface != TypefaceBold {
		t.Errorf("bold: got %v", b.Typeface)
	}

	bi := markupApplyItalic(b)
	if bi.Typeface != TypefaceBoldItalic {
		t.Errorf("bold+italic: got %v", bi.Typeface)
	}

	i := markupApplyItalic(base)
	if i.Typeface != TypefaceItalic {
		t.Errorf("italic: got %v", i.Typeface)
	}

	ib := markupApplyBold(i)
	if ib.Typeface != TypefaceBoldItalic {
		t.Errorf("italic+bold: got %v", ib.Typeface)
	}
}

func TestMarkupReplaceFontFamily(t *testing.T) {
	got := markupReplaceFontFamily("Sans 12", "Consolas")
	if got != "Consolas 12" {
		t.Errorf("got %q", got)
	}
	got = markupReplaceFontFamily("Sans", "Consolas")
	if got != "Consolas" {
		t.Errorf("no size: got %q", got)
	}
}

func TestMarkupReplaceFontSize(t *testing.T) {
	got := markupReplaceFontSize("Sans 12", 20)
	if got != "Sans 20" {
		t.Errorf("got %q", got)
	}
	got = markupReplaceFontSize("Sans", 14)
	if got != "Sans 14" {
		t.Errorf("no existing size: got %q", got)
	}
}

func TestParsePangoMarkupBasic(t *testing.T) {
	base := TextStyle{FontName: "Sans 12", Color: Color{0, 0, 0, 255}}
	runs, err := parsePangoMarkup("hello <b>bold</b> world", base)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}
	if runs[0].Text != "hello " {
		t.Errorf("run0 text: %q", runs[0].Text)
	}
	if runs[1].Text != "bold" || runs[1].Style.Typeface != TypefaceBold {
		t.Errorf("run1: text=%q typeface=%v", runs[1].Text, runs[1].Style.Typeface)
	}
	if runs[2].Text != " world" {
		t.Errorf("run2 text: %q", runs[2].Text)
	}
}

func TestParsePangoMarkupNested(t *testing.T) {
	base := TextStyle{FontName: "Sans 12"}
	runs, err := parsePangoMarkup("<b><i>bi</i></b>", base)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Style.Typeface != TypefaceBoldItalic {
		t.Errorf("expected BoldItalic, got %v", runs[0].Style.Typeface)
	}
}

func TestParsePangoMarkupSpanColor(t *testing.T) {
	base := TextStyle{FontName: "Sans 12", Color: Color{0, 0, 0, 255}}
	runs, err := parsePangoMarkup(`<span foreground="#FF0000">red</span>`, base)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Style.Color.R != 255 || runs[0].Style.Color.G != 0 {
		t.Errorf("color: %+v", runs[0].Style.Color)
	}
}

func TestParsePangoMarkupPlainText(t *testing.T) {
	base := TextStyle{FontName: "Sans 12"}
	runs, err := parsePangoMarkup("no markup", base)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 || runs[0].Text != "no markup" {
		t.Errorf("plain text: %+v", runs)
	}
}

func TestParsePangoMarkupMalformed(t *testing.T) {
	base := TextStyle{FontName: "Sans 12"}

	// Unclosed tag with content before the error.
	runs, err := parsePangoMarkup("hello <b>bold", base)
	if err == nil {
		t.Error("expected error for unclosed tag")
	}
	if len(runs) == 0 {
		t.Error("expected partial runs before error")
	}

	// Malformed entity mid-stream.
	runs, err = parsePangoMarkup("hello <b>bold</b> <invalid&>rest", base)
	if err == nil {
		t.Error("expected error for malformed XML")
	}
	if len(runs) < 2 {
		t.Errorf("expected at least 2 partial runs, got %d", len(runs))
	}
}
