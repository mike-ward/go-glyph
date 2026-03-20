//go:build !js

package glyph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateFontPath validates a font file path for safety and
// existence.
func ValidateFontPath(path string, location string) error {
	if len(path) == 0 {
		return fmt.Errorf("empty font path not allowed at %s",
			location)
	}
	for part := range strings.SplitSeq(filepath.ToSlash(path), "/") {
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
