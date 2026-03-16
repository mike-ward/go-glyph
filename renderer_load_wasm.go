//go:build js && wasm

package glyph

// LoadGlyphConfig holds parameters for WASM glyph rasterization.
type LoadGlyphConfig struct {
	Index        uint32
	Codepoint    uint32
	TargetHeight int
	SubpixelBin  int
	CSSFont      string
}

// LoadGlyphResult holds the output of a glyph load operation.
type LoadGlyphResult struct {
	Cached        CachedGlyph
	ResetOccurred bool
	ResetPage     int
}
