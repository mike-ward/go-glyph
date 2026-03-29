//go:build windows

package glyph

// drawLayoutImpl is a no-op stub for Windows.
func (r *Renderer) drawLayoutImpl(_ Layout, _, _ float32,
	_ AffineTransform, _ *GradientConfig) {
}
