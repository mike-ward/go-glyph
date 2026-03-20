//go:build ios

package glyph

/*
#include <CoreText/CoreText.h>
#include <CoreFoundation/CoreFoundation.h>

// ctRegisterFontURL registers a font file with Core Text.
static bool ctRegisterFontFile(const char *path) {
    CFStringRef pathStr = CFStringCreateWithCString(NULL, path,
        kCFStringEncodingUTF8);
    CFURLRef url = CFURLCreateWithFileSystemPath(NULL, pathStr,
        kCFURLPOSIXPathStyle, false);
    CFRelease(pathStr);
    if (!url) return false;
    bool ok = CTFontManagerRegisterFontsForURL(
        url, kCTFontManagerScopeProcess, NULL);
    CFRelease(url);
    return ok;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Context holds Core Text state for text shaping on iOS.
//
// Not safe for concurrent use.
type Context struct {
	scaleFactor float32
	scaleInv    float32
	metrics     metricsCache
}

// NewContext creates an iOS text context.
func NewContext(scaleFactor float32) (*Context, error) {
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	return &Context{
		scaleFactor: scaleFactor,
		scaleInv:    1.0 / scaleFactor,
		metrics:     newMetricsCache(256),
	}, nil
}

// Free releases resources.
func (ctx *Context) Free() {
	ctx.metrics = metricsCache{}
}

// ScaleFactor returns the DPI scale factor.
func (ctx *Context) ScaleFactor() float32 { return ctx.scaleFactor }

// AddFontFile registers a font file with Core Text.
func (ctx *Context) AddFontFile(path string) error {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))
	if !C.ctRegisterFontFile(cs) {
		return fmt.Errorf("CTFontManagerRegisterFontsForURL failed for %q", path)
	}
	return nil
}

// FontHeight returns ascent + descent in logical pixels.
func (ctx *Context) FontHeight(cfg TextConfig) (float32, error) {
	font := newCTFont(cfg.Style, ctx.scaleFactor)
	if font.ref == 0 {
		return 0, fmt.Errorf("failed to create CTFont")
	}
	defer font.close()

	ascent, descent, _ := font.metrics()
	return float32(ascent+descent) / ctx.scaleFactor, nil
}

// FontMetrics returns detailed metrics in logical pixels.
func (ctx *Context) FontMetrics(cfg TextConfig) (TextMetrics, error) {
	font := newCTFont(cfg.Style, ctx.scaleFactor)
	if font.ref == 0 {
		return TextMetrics{}, fmt.Errorf("failed to create CTFont")
	}
	defer font.close()

	ascent, descent, leading := font.metrics()
	sf := float64(ctx.scaleFactor)
	asc := float32(ascent / sf)
	dsc := float32(descent / sf)
	return TextMetrics{
		Ascender:  asc,
		Descender: dsc,
		Height:    asc + dsc,
		LineGap:   float32(leading / sf),
	}, nil
}

// ResolveFontName returns the resolved iOS font family name.
func (ctx *Context) ResolveFontName(fontDescStr string) (string, error) {
	family := resolveFontFamilyIOS(fontDescStr)
	return family, nil
}

// createFontDescription builds a ctFont from TextStyle. Caller
// must call close().
func (ctx *Context) createCTFont(style TextStyle) ctFont {
	return newCTFont(style, ctx.scaleFactor)
}
