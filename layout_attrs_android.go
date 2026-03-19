//go:build android

package glyph

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
