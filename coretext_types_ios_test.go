//go:build ios

package glyph

import "testing"

func TestParseSizeFromFontName(t *testing.T) {
	tests := []struct {
		name string
		want float32
	}{
		{"Sans Bold 18", 18},
		{"Monospace 12", 12},
		{"Sans", 0},
		{"", 0},
		{"Sans Bold", 0},
		{"Serif 0", 0},
		{"Mono 100", 100},
		{"Font 12.5", 12}, // fractional truncated at dot
	}
	for _, tt := range tests {
		got := parseSizeFromFontName(tt.name)
		if got != tt.want {
			t.Errorf("parseSizeFromFontName(%q) = %v, want %v",
				tt.name, got, tt.want)
		}
	}
}

func TestParseFamilyFromFontName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Sans Bold 18", "Sans"},
		{"Noto Sans Bold Italic 14", "Noto Sans"},
		{"Monospace 12", "Monospace"},
		{"Liberation Mono Bold", "Liberation Mono"},
		{"Sans", "Sans"},
		{"Bold", "Bold"}, // lone style word preserved (end==1)
		{"", ""},
		{"Fira Code Light 11", "Fira Code"},
		{"Serif Regular 16", "Serif"},
	}
	for _, tt := range tests {
		got := parseFamilyFromFontName(tt.name)
		if got != tt.want {
			t.Errorf("parseFamilyFromFontName(%q) = %q, want %q",
				tt.name, got, tt.want)
		}
	}
}

func TestResolveFontFamilyIOS(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Sans 12", ".AppleSystemUIFont"},
		{"sans-serif Bold 14", ".AppleSystemUIFont"},
		{"Serif 11", "New York"},
		{"Monospace 10", "SF Mono"},
		{"mono Bold 12", "SF Mono"},
		{"system 16", ".AppleSystemUIFont"},
		{"Fira Code 12", "Fira Code"},
		{"", ".AppleSystemUIFont"},
	}
	for _, tt := range tests {
		got := resolveFontFamilyIOS(tt.name)
		if got != tt.want {
			t.Errorf("resolveFontFamilyIOS(%q) = %q, want %q",
				tt.name, got, tt.want)
		}
	}
}
