//go:build !ios

package gpu

/*
#cgo CFLAGS: -fobjc-arc
#cgo darwin,arm64 CFLAGS: -I/opt/homebrew/include/SDL2 -D_THREAD_SAFE
#cgo darwin,amd64 CFLAGS: -I/usr/local/include/SDL2 -D_THREAD_SAFE
#cgo LDFLAGS: -framework Metal -framework QuartzCore -framework Foundation
#include "metal_darwin.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// gpuCtx wraps the opaque C MetalCtx pointer.
type gpuCtx struct {
	ptr *C.MetalCtx
}

func gpuInitGo(sdlWin unsafe.Pointer, dpiScale float32) (*gpuCtx, error) {
	ctx := C.metalInit(sdlWin, C.float(dpiScale))
	if ctx == nil {
		return nil, fmt.Errorf("gpu: metalInit failed")
	}
	return &gpuCtx{ptr: ctx}, nil
}

func (m *gpuCtx) newTexture(w, h int) uint64 {
	return uint64(C.metalNewTex(m.ptr, C.int(w), C.int(h)))
}

func (m *gpuCtx) updateTexture(id uint64, data []byte, w, h int) {
	if len(data) == 0 {
		return
	}
	C.metalUpdateTex(m.ptr, C.uint64_t(id),
		unsafe.Pointer(&data[0]), C.int(w), C.int(h))
}

func (m *gpuCtx) deleteTexture(id uint64) {
	C.metalDeleteTex(m.ptr, C.uint64_t(id))
}

func (m *gpuCtx) render(verts []Vertex, cmds []drawCmd,
	clearR, clearG, clearB, clearA float32,
	logicalW, logicalH int) error {

	var vp, cp unsafe.Pointer
	vc := len(verts)
	cc := len(cmds)
	if vc > 0 {
		vp = unsafe.Pointer(&verts[0])
	}
	if cc > 0 {
		cp = unsafe.Pointer(&cmds[0])
	}
	rc := C.metalRender(m.ptr,
		vp, C.int(vc),
		cp, C.int(cc),
		C.float(clearR), C.float(clearG),
		C.float(clearB), C.float(clearA),
		C.int(logicalW), C.int(logicalH))
	if rc != 0 {
		return fmt.Errorf("gpu: metalRender failed")
	}
	return nil
}

func (m *gpuCtx) drawableSize() (int, int) {
	var w, h C.int
	C.metalGetDrawableSize(m.ptr, &w, &h)
	return int(w), int(h)
}

func (m *gpuCtx) destroy() {
	if m.ptr != nil {
		C.metalDestroy(m.ptr)
		m.ptr = nil
	}
}

// WindowFlag returns SDL_WINDOW_METAL.
func WindowFlag() uint32 {
	return uint32(C.metalWindowFlag())
}

// WindowDrawableSize returns the physical drawable size for
// an SDL2 Metal window. sdlWindow is unsafe.Pointer to SDL_Window.
func WindowDrawableSize(sdlWindow unsafe.Pointer) (int, int) {
	var w, h C.int
	C.metalWindowDrawableSize(sdlWindow, &w, &h)
	return int(w), int(h)
}
