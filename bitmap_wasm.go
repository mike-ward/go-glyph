//go:build js && wasm

package glyph

import (
	"fmt"
	"math"
)

const MaxGlyphSize = 256
const maxAllocationSize = 1024 * 1024 * 1024

// Bitmap holds RGBA pixel data for a rasterized glyph.
type Bitmap struct {
	Width    int
	Height   int
	Channels int
	Data     []byte
}

func checkAllocationSize(w, h, channels int) (int64, error) {
	size := int64(w) * int64(h) * int64(channels)
	if size <= 0 {
		return 0, fmt.Errorf("invalid allocation size: %dx%dx%d",
			w, h, channels)
	}
	if size > int64(math.MaxInt32) {
		return 0, fmt.Errorf("allocation overflow: %d bytes", size)
	}
	if size > maxAllocationSize {
		return 0, fmt.Errorf("allocation exceeds 1GB limit: %d bytes", size)
	}
	return size, nil
}
