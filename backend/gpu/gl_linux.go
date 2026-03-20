//go:build linux && !android

package gpu

/*
#cgo CFLAGS: -I/usr/include/SDL2 -D_REENTRANT
#cgo LDFLAGS: -lSDL2 -lGL
#include "gl_linux.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// gpuCtx wraps the opaque C GLCtx pointer.
type gpuCtx struct {
	ptr *C.GLCtx
}

func gpuInitGo(sdlWin unsafe.Pointer, dpiScale float32) (*gpuCtx, error) {
	ctx := C.glCtxInit(sdlWin, C.float(dpiScale))
	if ctx == nil {
		return nil, fmt.Errorf("gpu: glCtxInit failed")
	}
	return &gpuCtx{ptr: ctx}, nil
}

func (m *gpuCtx) newTexture(w, h int) uint64 {
	return uint64(C.glCtxNewTex(m.ptr, C.int(w), C.int(h)))
}

func (m *gpuCtx) updateTexture(id uint64, data []byte, w, h int) {
	C.glCtxUpdateTex(m.ptr, C.uint64_t(id),
		unsafe.Pointer(&data[0]), C.int(w), C.int(h))
}

func (m *gpuCtx) deleteTexture(id uint64) {
	C.glCtxDeleteTex(m.ptr, C.uint64_t(id))
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
	rc := C.glCtxRender(m.ptr,
		vp, C.int(vc),
		cp, C.int(cc),
		C.float(clearR), C.float(clearG),
		C.float(clearB), C.float(clearA),
		C.int(logicalW), C.int(logicalH))
	if rc != 0 {
		return fmt.Errorf("gpu: glCtxRender failed")
	}
	return nil
}

func (m *gpuCtx) drawableSize() (int, int) {
	var w, h C.int
	C.glCtxGetDrawableSize(m.ptr, &w, &h)
	return int(w), int(h)
}

func (m *gpuCtx) destroy() {
	if m.ptr != nil {
		C.glCtxDestroy(m.ptr)
		m.ptr = nil
	}
}

// WindowFlag returns SDL_WINDOW_OPENGL.
func WindowFlag() uint32 {
	return 0x00000002 // SDL_WINDOW_OPENGL
}

// WindowDrawableSize returns the physical drawable size for
// an SDL2 OpenGL window. sdlWindow is unsafe.Pointer to SDL_Window.
func WindowDrawableSize(sdlWindow unsafe.Pointer) (int, int) {
	var w, h C.int
	C.SDL_GL_GetDrawableSize((*C.SDL_Window)(sdlWindow), &w, &h)
	return int(w), int(h)
}
