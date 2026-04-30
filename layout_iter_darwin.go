//go:build darwin && !glyph_pango

package glyph

// runMetrics holds underline/strikethrough positioning.
type runMetrics struct {
	UndPos      float64
	UndThick    float64
	StrikePos   float64
	StrikeThick float64
}
