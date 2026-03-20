//go:build js && wasm

package glyph

// ValidateFontPath is a no-op under WASM (no filesystem).
func ValidateFontPath(_ string, _ string) error { return nil }
