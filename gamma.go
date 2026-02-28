package glyph

import "math"

// gammaTable is a pre-computed lookup table for gamma correction
// applied during grayscale glyph rasterization.
// Gamma value 1.8 matches the V implementation.
var gammaTable [256]byte

func init() {
	const gamma = 1.8
	invGamma := 1.0 / gamma
	for i := range 256 {
		gammaTable[i] = byte(math.Round(math.Pow(float64(i)/255.0, invGamma) * 255.0))
	}
}
