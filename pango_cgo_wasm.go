//go:build js && wasm

package glyph

import "unsafe"

// Stub RAII types for WASM — no CGo, no FreeType/Pango.

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

// PangoContextW is a no-op stub.
type PangoContextW struct{}

func (c *PangoContextW) Close() {}

// PangoLayoutW is a no-op stub.
type PangoLayoutW struct{}

func (l *PangoLayoutW) Close() {}

// PangoFontDescW is a no-op stub.
type PangoFontDescW struct{}

func (d *PangoFontDescW) Close()                {}
func (d PangoFontDescW) SetSize(_ int)          {}
func (d PangoFontDescW) SetWeight(_ int)        {}
func (d PangoFontDescW) SetStyle(_ int)         {}
func (d PangoFontDescW) SetVariations(_ string) {}

// PangoAttrListW is a no-op stub.
type PangoAttrListW struct{}

func NewPangoAttrList() PangoAttrListW { return PangoAttrListW{} }
func (a *PangoAttrListW) Close()       {}

// PangoLayoutIterW is a no-op stub.
type PangoLayoutIterW struct{}

func (it *PangoLayoutIterW) Close() {}

// PangoTabArrayW is a no-op stub.
type PangoTabArrayW struct{}

func NewPangoTabArray(_ int) PangoTabArrayW { return PangoTabArrayW{} }
func (t PangoTabArrayW) SetTab(_, _ int)    {}
func (t *PangoTabArrayW) Close()            {}

// PangoFontW is a no-op stub.
type PangoFontW struct{}

func (f *PangoFontW) Close() {}

// PangoFontMetricsW is a no-op stub.
type PangoFontMetricsW struct{}

func (m *PangoFontMetricsW) Close() {}

// getFontFamilyName returns "Unknown" in WASM.
func getFontFamilyName(_ unsafe.Pointer) string { return "Unknown" }
