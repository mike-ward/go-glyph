//go:build windows

package glyph

import "unsafe"

// Stub RAII types for Windows — no CGo, no FreeType/Pango.

// FTLibrary is a no-op stub.
type FTLibrary struct{}

func InitFreeType() (FTLibrary, error) { return FTLibrary{}, nil }
func (l *FTLibrary) Close()            {}

// FTFace is a no-op stub.
type FTFace struct{}

func (f *FTFace) FacePtr() unsafe.Pointer { return nil }

// FTStroker is a no-op stub.
type FTStroker struct{}

func NewFTStroker(_ FTLibrary) (FTStroker, error) { return FTStroker{}, nil }
func (s *FTStroker) Close()                       {}

// PangoFontMapW is a no-op stub.
type PangoFontMapW struct{}

func NewPangoFT2FontMap() PangoFontMapW              { return PangoFontMapW{} }
func (m PangoFontMapW) SetResolution(_, _ float64)   {}
func (m PangoFontMapW) CreateContext() PangoContextW { return PangoContextW{} }
func (m *PangoFontMapW) Close()                      {}

// PangoFCFontMapConfigChanged is a no-op on Windows.
func PangoFCFontMapConfigChanged(_ PangoFontMapW) {}

// PangoContextW is a no-op stub.
type PangoContextW struct{}

func (c *PangoContextW) Close() {}

// PangoLayoutW is a no-op stub.
type PangoLayoutW struct{}

func (l *PangoLayoutW) Close() {}

// PangoFontDescW is a no-op stub.
type PangoFontDescW struct{}

func (d *PangoFontDescW) Close()         {}
func (d PangoFontDescW) SetSize(_ int)   {}
func (d PangoFontDescW) SetWeight(_ int) {}
func (d PangoFontDescW) SetStyle(_ int)  {}

// PangoFontW is a no-op stub.
type PangoFontW struct{}

// PangoFontMetricsW is a no-op stub.
type PangoFontMetricsW struct{}

// PangoTabArrayW is a no-op stub.
type PangoTabArrayW struct{}

func NewPangoTabArray(_ int, _ bool) PangoTabArrayW { return PangoTabArrayW{} }
func (t *PangoTabArrayW) Close()                    {}

// PangoAttrListW is a no-op stub.
type PangoAttrListW struct{}

func (a *PangoAttrListW) Close() {}

// PangoLayoutIterW is a no-op stub.
type PangoLayoutIterW struct{}

// getFontFamilyName returns a placeholder on Windows.
func getFontFamilyName(_ unsafe.Pointer) string { return "Unknown" }
