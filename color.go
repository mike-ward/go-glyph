package glyph

import "math"

// Color is an RGBA color with 8-bit channels.
type Color struct {
	R, G, B, A uint8
}

// LerpColor linearly interpolates between two colors. t is clamped to [0,1].
func LerpColor(a, b Color, t float32) Color {
	tc := float32(math.Min(math.Max(float64(t), 0), 1))
	inv := 1.0 - tc
	return Color{
		R: uint8(float32(a.R)*inv + float32(b.R)*tc),
		G: uint8(float32(a.G)*inv + float32(b.G)*tc),
		B: uint8(float32(a.B)*inv + float32(b.B)*tc),
		A: uint8(float32(a.A)*inv + float32(b.A)*tc),
	}
}
