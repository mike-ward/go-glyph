//go:build android

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
		{"Font 12.5", 12},
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
		{"Bold", "Bold"},
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

func TestResolveFontFamilyAndroid(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Sans 12", "Roboto"},
		{"sans-serif Bold 14", "Roboto"},
		{"Serif 11", "NotoSerif"},
		{"Monospace 10", "DroidSansMono"},
		{"mono Bold 12", "DroidSansMono"},
		{"system 16", "Roboto"},
		{"Fira Code 12", "Fira Code"},
		{"", "Roboto"},
	}
	for _, tt := range tests {
		got := resolveFontFamilyAndroid(tt.name)
		if got != tt.want {
			t.Errorf("resolveFontFamilyAndroid(%q) = %q, want %q",
				tt.name, got, tt.want)
		}
	}
}
