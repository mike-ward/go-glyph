package sdl2

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework QuartzCore
#import <QuartzCore/QuartzCore.h>

static void
setPresentsWithTransaction(void *layer) {
	CAMetalLayer *ml = (__bridge CAMetalLayer *)layer;
	ml.presentsWithTransaction = YES;
}
*/
import "C"

import (
	"github.com/veandco/go-sdl2/sdl"
)

// SyncMetalLayer sets presentsWithTransaction on the
// renderer's CAMetalLayer so that frame presentation is
// synchronized with Core Animation transactions. This
// eliminates the 1-frame compositor stretch during live
// window resize on macOS.
func SyncMetalLayer(renderer *sdl.Renderer) {
	layer, err := renderer.GetMetalLayer()
	if err != nil || layer == nil {
		return
	}
	C.setPresentsWithTransaction(layer)
}
