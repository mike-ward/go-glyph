package glyph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	// MaxTextLength is the maximum text input length (10KB) for DoS prevention.
	MaxTextLength = 10240
	// MaxTextureDimension is the maximum texture size in pixels.
	MaxTextureDimension = 16384
	// MinFontSize is the minimum font size in points.
	MinFontSize = float32(0.1)
	// MaxFontSize is the maximum font size in points.
	MaxFontSize = float32(500.0)
)

// ValidateTextInput validates text for UTF-8, non-empty, and length.
// Returns error if invalid.
func ValidateTextInput(text string, maxLen int, location string) error {
	if len(text) == 0 {
		return fmt.Errorf("empty string not allowed at %s", location)
	}
	if len(text) > maxLen {
		return fmt.Errorf("text exceeds max length %d bytes at %s", maxLen, location)
	}
	if !utf8.ValidString(text) {
		return fmt.Errorf("invalid UTF-8 encoding at %s", location)
	}
	if strings.ContainsRune(text, '\x00') {
		return fmt.Errorf("null byte in text at %s", location)
	}
	return nil
}

// ValidateFontPath validates a font file path for safety and existence.
func ValidateFontPath(path string, location string) error {
	if len(path) == 0 {
		return fmt.Errorf("empty font path not allowed at %s", location)
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal (..) not allowed in font path at %s", location)
	}
	cleaned := filepath.Clean(path)
	if _, err := os.Stat(cleaned); os.IsNotExist(err) {
		return fmt.Errorf("font file does not exist: %q at %s", path, location)
	}
	return nil
}

// ValidateSize validates a numeric size against min/max bounds.
func ValidateSize(size, min, max float32, name, location string) error {
	if size < min || size > max {
		return fmt.Errorf("%s %g out of range [%g, %g] at %s", name, size, min, max, location)
	}
	return nil
}

// ValidateDimension validates an integer dimension (width/height).
func ValidateDimension(dim int, name, location string) error {
	if dim <= 0 {
		return fmt.Errorf("%s must be positive, got %d at %s", name, dim, location)
	}
	if dim > MaxTextureDimension {
		return fmt.Errorf("%s %d exceeds max %d at %s", name, dim, MaxTextureDimension, location)
	}
	return nil
}
