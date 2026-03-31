//go:build windows

package glyph

// runAttributes holds parsed visual properties from a layout run.
type runAttributes struct {
	Color            Color
	BgColor          Color
	HasBgColor       bool
	HasUnderline     bool
	HasStrikethrough bool
	IsObject         bool
	ObjectID         string
}
