//go:build windows

package glyph

import "fmt"

// LayoutText is a stub for Windows.
func (ctx *Context) LayoutText(_ string, _ TextConfig) (Layout, error) {
	return Layout{}, fmt.Errorf("glyph: Windows text shaping not yet implemented")
}

// LayoutRichText is a stub for Windows.
func (ctx *Context) LayoutRichText(_ RichText, _ TextConfig) (Layout, error) {
	return Layout{}, fmt.Errorf("glyph: Windows text shaping not yet implemented")
}
