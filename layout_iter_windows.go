//go:build windows

package glyph

// runMetrics holds font metrics for a single run.
type runMetrics struct {
	Ascent      float64
	Descent     float64
	StrikePos   float64
	StrikeThick float64
}
