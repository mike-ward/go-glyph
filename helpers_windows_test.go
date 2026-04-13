//go:build windows

package glyph

import "testing"

func TestParseFontDesc(t *testing.T) {
	tests := []struct {
		style      TextStyle
		wantFamily string
		wantSize   float32
	}{
		{TextStyle{FontName: "Sans 14"}, "Segoe UI", 14},
		{TextStyle{FontName: "Consolas 10"}, "Consolas", 10},
		{TextStyle{FontName: "Times New Roman Bold 12"}, "Times New Roman", 12},
		{TextStyle{FontName: "Sans Bold Italic 16"}, "Segoe UI", 16},
		{TextStyle{FontName: ""}, "Segoe UI", 12},
		{TextStyle{FontName: "Mono"}, "Consolas", 12},
		{TextStyle{FontName: "Serif 18"}, "Times New Roman", 18},
		{TextStyle{FontName: "Arial", Size: 20}, "Arial", 20},
		{TextStyle{FontName: "Arial 10", Size: 20}, "Arial", 20}, // explicit Size wins
	}
	for _, tt := range tests {
		family, size := parseFontDesc(tt.style)
		if family != tt.wantFamily || size != tt.wantSize {
			t.Errorf("parseFontDesc(%q, Size=%v) = (%q, %v); want (%q, %v)",
				tt.style.FontName, tt.style.Size, family, size, tt.wantFamily, tt.wantSize)
		}
	}
}

func TestMapWindowsFamily(t *testing.T) {
	tests := map[string]string{
		"sans":       "Segoe UI",
		"Sans-Serif": "Segoe UI",
		"serif":      "Times New Roman",
		"monospace":  "Consolas",
		"mono":       "Consolas",
		"Arial":      "Arial",
	}
	for input, want := range tests {
		if got := mapWindowsFamily(input); got != want {
			t.Errorf("mapWindowsFamily(%q) = %q; want %q", input, got, want)
		}
	}
}

func TestWinFontParams(t *testing.T) {
	style := TextStyle{FontName: "Sans 14", Typeface: TypefaceBold}
	family, heightPx, weight, italic := winFontParams(style, 2.0)
	if family != "Segoe UI" {
		t.Errorf("family = %q", family)
	}
	if heightPx >= 0 {
		t.Errorf("heightPx should be negative, got %d", heightPx)
	}
	if weight != _FW_BOLD {
		t.Errorf("weight = %d; want %d", weight, _FW_BOLD)
	}
	if italic {
		t.Error("unexpected italic")
	}

	// Italic
	style.Typeface = TypefaceItalic
	_, _, w, i := winFontParams(style, 1.0)
	if w != _FW_NORMAL || !i {
		t.Errorf("italic: weight=%d, italic=%v", w, i)
	}

	// BoldItalic
	style.Typeface = TypefaceBoldItalic
	_, _, w, i = winFontParams(style, 1.0)
	if w != _FW_BOLD || !i {
		t.Errorf("bolditalic: weight=%d, italic=%v", w, i)
	}
}

func TestIsEmojiRune(t *testing.T) {
	emojis := []rune{'😀', '🚀', '❤', '⌚', '🀄'}
	for _, r := range emojis {
		if !isEmojiRune(r) {
			t.Errorf("expected emoji: %U", r)
		}
	}
	nonEmoji := []rune{'A', '1', ' ', '中', 'é'}
	for _, r := range nonEmoji {
		if isEmojiRune(r) {
			t.Errorf("unexpected emoji: %U", r)
		}
	}
}

func TestWinDetectSubSup(t *testing.T) {
	// No features.
	s := TextStyle{}
	if winDetectSubSup(s) != winSubSupNone {
		t.Error("expected none")
	}

	// Subscript.
	s.Features = &FontFeatures{
		OpenTypeFeatures: []FontFeature{{Tag: "subs", Value: 1}},
	}
	if winDetectSubSup(s) != winSubSupSub {
		t.Error("expected sub")
	}

	// Superscript.
	s.Features = &FontFeatures{
		OpenTypeFeatures: []FontFeature{{Tag: "sups", Value: 1}},
	}
	if winDetectSubSup(s) != winSubSupSup {
		t.Error("expected sup")
	}

	// Disabled feature (Value=0).
	s.Features = &FontFeatures{
		OpenTypeFeatures: []FontFeature{{Tag: "subs", Value: 0}},
	}
	if winDetectSubSup(s) != winSubSupNone {
		t.Error("expected none for disabled feature")
	}
}

func TestWinScaleFontSize(t *testing.T) {
	got := winScaleFontSize("Sans 20", 0.65)
	if got != "Sans 13" {
		t.Errorf("got %q", got)
	}
	// No size component.
	got = winScaleFontSize("Sans", 0.65)
	if got != "Sans" {
		t.Errorf("no size: got %q", got)
	}
}
