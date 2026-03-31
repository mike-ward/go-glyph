//go:build windows

package glyph

// FreeType load target constants (stubs for Windows).
const (
	FTLoadTargetNormal = 0
	FTLoadTargetLight  = 1 << 16
	FTLoadTargetMono   = 1 << 17
	FTLoadTargetLCD    = 3 << 16
)

// SubpixelBins is the number of fractional-pixel bins for glyph
// positioning. Must be a power of two.
const SubpixelBins = 4

// FTSubpixelUnit is the FreeType 26.6 shift per subpixel bin.
const FTSubpixelUnit = 16 // 64 / SubpixelBins
