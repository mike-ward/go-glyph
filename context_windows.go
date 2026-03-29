//go:build windows

package glyph

import "fmt"

// Context is a stub for Windows. Text shaping is not yet
// implemented on this platform.
//
// Not safe for concurrent use.
type Context struct {
	scaleFactor float32
	scaleInv    float32
	metrics     metricsCache
}

// NewContext returns an error because Windows text shaping is
// not yet implemented.
func NewContext(scaleFactor float32) (*Context, error) {
	return nil, fmt.Errorf("glyph: Windows text shaping not yet implemented")
}

// Free releases resources.
func (ctx *Context) Free() {}

// ScaleFactor returns the DPI scale factor.
func (ctx *Context) ScaleFactor() float32 { return ctx.scaleFactor }

// AddFontFile is a no-op stub.
func (ctx *Context) AddFontFile(_ string) error { return nil }

// FontHeight is a stub that returns zero.
func (ctx *Context) FontHeight(_ TextConfig) (float32, error) {
	return 0, fmt.Errorf("glyph: Windows text shaping not yet implemented")
}

// FontMetrics is a stub that returns zero metrics.
func (ctx *Context) FontMetrics(_ TextConfig) (TextMetrics, error) {
	return TextMetrics{}, fmt.Errorf("glyph: Windows text shaping not yet implemented")
}

// ResolveFontName is a stub that returns the input name.
func (ctx *Context) ResolveFontName(name string) (string, error) {
	return name, nil
}
