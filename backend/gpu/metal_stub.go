//go:build !darwin

package gpu

import (
	"fmt"
	"unsafe"
)

type metalCtx struct{}

func metalInitGo(_ unsafe.Pointer, _ float32) (*metalCtx, error) {
	return nil, fmt.Errorf("gpu: Metal backend requires macOS")
}

func (m *metalCtx) newTexture(_, _ int) uint64                         { return 0 }
func (m *metalCtx) updateTexture(_ uint64, _ []byte, _, _ int)         {}
func (m *metalCtx) deleteTexture(_ uint64)                             {}
func (m *metalCtx) render(_ []Vertex, _ []drawCmd, _, _, _, _ float32, _, _ int) error {
	return fmt.Errorf("gpu: Metal backend requires macOS")
}
func (m *metalCtx) drawableSize() (int, int) { return 0, 0 }
func (m *metalCtx) destroy()                 {}

func WindowFlag() uint32                                  { return 0 }
func WindowDrawableSize(_ unsafe.Pointer) (int, int)      { return 0, 0 }
