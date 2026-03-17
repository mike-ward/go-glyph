//go:build ios

package glyph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	MaxTextLength       = 10240
	MaxTextureDimension = 16384
	MinFontSize         = float32(0.1)
	MaxFontSize         = float32(500.0)
)

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

func ValidateFontPath(path string, location string) error {
	if len(path) == 0 {
		return fmt.Errorf("empty font path not allowed at %s",
			location)
	}
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if part == ".." {
			return fmt.Errorf(
				"path traversal (..) not allowed in font path at %s",
				location)
		}
	}
	cleaned := filepath.Clean(path)
	if _, err := os.Stat(cleaned); err != nil {
		return fmt.Errorf("font file not accessible: %q at %s: %w",
			path, location, err)
	}
	return nil
}

func ValidateSize(size, min, max float32,
	name, location string) error {
	if size < min || size > max {
		return fmt.Errorf("%s %g out of range [%g, %g] at %s",
			name, size, min, max, location)
	}
	return nil
}

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
