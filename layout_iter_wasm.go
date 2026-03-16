//go:build js && wasm

package glyph

// WASM stubs for layout_iter.go types and helpers.
// processRun, computeHitTestRects, computeLines, extractLogAttrs
// are handled inline by layout_wasm.go's buildLayout.

// runMetrics holds underline/strikethrough positioning.
type runMetrics struct {
	UndPos      float64
	UndThick    float64
	StrikePos   float64
	StrikeThick float64
}
