package glyph

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// MaxTextLength is the maximum text input length (10KB) for
	// DoS prevention.
	MaxTextLength = 10240
	// MaxTextureDimension is the maximum texture size in pixels.
	MaxTextureDimension = 16384
	// MinFontSize is the minimum font size in points.
	MinFontSize = float32(0.1)
	// MaxFontSize is the maximum font size in points.
	MaxFontSize = float32(500.0)
)

// ValidateTextInput validates text for UTF-8, non-empty, and length.
func ValidateTextInput(text string, maxLen int, location string) error {
	if len(text) == 0 {
		return fmt.Errorf("empty string not allowed at %s", location)
	}
	if len(text) > maxLen {
		return fmt.Errorf("text exceeds max length %d bytes at %s",
			maxLen, location)
	}
	if !utf8.ValidString(text) {
		return fmt.Errorf("invalid UTF-8 encoding at %s", location)
	}
	if strings.ContainsRune(text, '\x00') {
		return fmt.Errorf("null byte in text at %s", location)
	}
	return nil
}

// ValidateSize validates a numeric size against min/max bounds.
func ValidateSize(size, minVal, maxVal float32,
	name, location string) error {
	if size < minVal || size > maxVal {
		return fmt.Errorf("%s %g out of range [%g, %g] at %s",
			name, size, minVal, maxVal, location)
	}
	return nil
}

// ValidateDimension validates an integer dimension (width/height).
func ValidateDimension(dim int, name, location string) error {
	if dim <= 0 {
		return fmt.Errorf("%s must be positive, got %d at %s",
			name, dim, location)
	}
	if dim > MaxTextureDimension {
		return fmt.Errorf("%s %d exceeds max %d at %s",
			name, dim, MaxTextureDimension, location)
	}
	return nil
}
