//go:build windows

package glyph

// LoadGlyphConfig holds parameters for Windows glyph rasterization.
type LoadGlyphConfig struct {
	Index        uint32
	Codepoint    uint32
	TargetHeight int
	SubpixelBin  int
}

// LoadGlyphResult holds the output of a glyph load operation.
type LoadGlyphResult struct {
	Cached        CachedGlyph
	ResetOccurred bool
	ResetPage     int
}
