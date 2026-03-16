//go:build js && wasm

package glyph

// WASM stubs for layout_attrs.go types.
// Style parsing is handled by layout_wasm.go's buildLayout
// and LayoutRichText.

// runAttributes holds parsed visual properties.
type runAttributes struct {
	Color            Color
	BgColor          Color
	HasBgColor       bool
	HasUnderline     bool
	HasStrikethrough bool
	IsObject         bool
	ObjectID         string
}
