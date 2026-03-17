//go:build ios

package glyph

/*
#include <CoreGraphics/CoreGraphics.h>
#include <CoreText/CoreText.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>

// GlyphRenderCtx holds shared state for glyph rasterization.
typedef struct {
    CGContextRef ctx;
    void *data;
    CTLineRef line;
    CFAttributedStringRef astr;
    CFDictionaryRef attrs;
    CFStringRef str;
    CTFontRef font;
    CGFloat minX, minY;
} GlyphRenderCtx;

// cgSetupGlyph creates font, attributed string, measures bounds,
// and allocates a bitmap context. Returns zeroed ctx on failure.
static GlyphRenderCtx cgSetupGlyph(const char *text,
    const char *family, CGFloat fontSize, bool bold, bool italic,
    int pad, int *outW, int *outH, int *outLeft, int *outTop) {

    GlyphRenderCtx r = {0};

    CFStringRef fam = CFStringCreateWithCString(NULL, family,
        kCFStringEncodingUTF8);
    CTFontRef baseFont = CTFontCreateWithName(fam, fontSize, NULL);
    CFRelease(fam);

    CTFontRef font = baseFont;
    if (bold || italic) {
        CTFontSymbolicTraits traits = 0;
        if (bold) traits |= kCTFontBoldTrait;
        if (italic) traits |= kCTFontItalicTrait;
        CTFontRef styled = CTFontCreateCopyWithSymbolicTraits(
            baseFont, fontSize, NULL, traits, traits);
        if (styled) {
            CFRelease(baseFont);
            font = styled;
        }
    }

    CFStringRef str = CFStringCreateWithCString(NULL, text,
        kCFStringEncodingUTF8);
    if (!str) {
        CFRelease(font);
        *outW = 0; *outH = 0;
        return r;
    }

    CFStringRef keys[] = { kCTFontAttributeName };
    CFTypeRef vals[] = { font };
    CFDictionaryRef attrs = CFDictionaryCreate(NULL,
        (const void **)keys, (const void **)vals, 1,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks);
    CFAttributedStringRef astr = CFAttributedStringCreate(
        NULL, str, attrs);
    CTLineRef line = CTLineCreateWithAttributedString(astr);

    CGRect bounds = CTLineGetBoundsWithOptions(line,
        kCTLineBoundsUseGlyphPathBounds);
    CGFloat minX = floor(CGRectGetMinX(bounds));
    CGFloat maxX = ceil(CGRectGetMaxX(bounds));
    CGFloat minY = floor(CGRectGetMinY(bounds));
    CGFloat maxY = ceil(CGRectGetMaxY(bounds));
    int w = (int)(maxX - minX) + pad * 2;
    int h = (int)(maxY - minY) + pad * 2;
    if (w < 1) w = 1;
    if (h < 1) h = 1;
    if (w > 256) w = 256;
    if (h > 256) h = 256;

    *outW = w;
    *outH = h;
    *outLeft = (int)minX - pad;
    *outTop = (int)maxY + pad;

    size_t bytesPerRow = w * 4;
    void *data = calloc(h, bytesPerRow);
    if (!data) {
        CFRelease(line);
        CFRelease(astr);
        CFRelease(attrs);
        CFRelease(str);
        CFRelease(font);
        *outW = 0; *outH = 0;
        return r;
    }

    CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
    CGContextRef ctx = CGBitmapContextCreate(data, w, h, 8,
        bytesPerRow, cs,
        kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big);
    CGColorSpaceRelease(cs);

    if (!ctx) {
        free(data);
        CFRelease(line);
        CFRelease(astr);
        CFRelease(attrs);
        CFRelease(str);
        CFRelease(font);
        *outW = 0; *outH = 0;
        return r;
    }

    r.ctx = ctx;
    r.data = data;
    r.line = line;
    r.astr = astr;
    r.attrs = attrs;
    r.str = str;
    r.font = font;
    r.minX = minX;
    r.minY = minY;
    return r;
}

static void cgCleanupGlyph(GlyphRenderCtx *r) {
    CGContextRelease(r->ctx);
    CFRelease(r->line);
    CFRelease(r->astr);
    CFRelease(r->attrs);
    CFRelease(r->str);
    CFRelease(r->font);
}

// cgRenderGlyph rasterizes a text string into an RGBA bitmap.
// Returns bitmap data (caller must free), width and height.
static void* cgRenderGlyph(const char *text, const char *family,
    CGFloat fontSize, bool bold, bool italic, CGFloat subpixelShift,
    int *outW, int *outH, int *outLeft, int *outTop) {

    const int pad = 2;
    GlyphRenderCtx r = cgSetupGlyph(text, family, fontSize,
        bold, italic, pad, outW, outH, outLeft, outTop);
    if (!r.ctx) return NULL;

    CGContextSetRGBFillColor(r.ctx, 1, 1, 1, 1);
    CGContextSetTextDrawingMode(r.ctx, kCGTextFill);

    CGFloat baselineY = -r.minY + pad;
    CGFloat baselineX = -r.minX + pad + subpixelShift;
    CGContextSetTextPosition(r.ctx, baselineX, baselineY);
    CTLineDraw(r.line, r.ctx);

    cgCleanupGlyph(&r);
    return r.data;
}

// cgRenderStrokedGlyph rasterizes a stroked text string.
static void* cgRenderStrokedGlyph(const char *text,
    const char *family, CGFloat fontSize,
    bool bold, bool italic, CGFloat strokeWidth, CGFloat subpixelShift,
    int *outW, int *outH, int *outLeft, int *outTop) {

    int pad = (int)ceil(strokeWidth) + 4;
    GlyphRenderCtx r = cgSetupGlyph(text, family, fontSize,
        bold, italic, pad, outW, outH, outLeft, outTop);
    if (!r.ctx) return NULL;

    CGContextSetRGBStrokeColor(r.ctx, 1, 1, 1, 1);
    CGContextSetRGBFillColor(r.ctx, 0, 0, 0, 0);
    CGContextSetLineWidth(r.ctx, strokeWidth);
    CGContextSetLineJoin(r.ctx, kCGLineJoinRound);
    CGContextSetLineCap(r.ctx, kCGLineCapRound);
    CGContextSetTextDrawingMode(r.ctx, kCGTextStroke);

    CGFloat baselineY = -r.minY + pad;
    CGFloat baselineX = -r.minX + pad + subpixelShift;
    CGContextSetTextPosition(r.ctx, baselineX, baselineY);
    CTLineDraw(r.line, r.ctx);

    cgCleanupGlyph(&r);
    return r.data;
}
*/
import "C"
import (
	"fmt"
	"math"
	"unsafe"
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
		return 0, fmt.Errorf("allocation exceeds 1GB limit: %d bytes",
			size)
	}
	return size, nil
}

// loadGlyphCG rasterizes a character using Core Graphics.
func loadGlyphCG(atlas *GlyphAtlas, ch string, item Item,
	subpixelBin int, scaleFactor float32) (LoadGlyphResult, error) {

	family, fontSize, bold, italic := resolveCTFontParams(
		item.Style, scaleFactor)

	cText := C.CString(ch)
	defer C.free(unsafe.Pointer(cText))
	cFamily := C.CString(family)
	defer C.free(unsafe.Pointer(cFamily))

	var w, h, left, top C.int
	subpixelShift := C.CGFloat(float64(subpixelBin) / 4.0)

	data := C.cgRenderGlyph(cText, cFamily,
		C.CGFloat(fontSize), C.bool(bold), C.bool(italic),
		subpixelShift,
		&w, &h, &left, &top)

	if data == nil || w == 0 || h == 0 {
		return LoadGlyphResult{}, nil
	}
	defer C.free(data)

	width := int(w)
	height := int(h)
	bmpSize, err := checkAllocationSize(width, height, 4)
	if err != nil {
		return LoadGlyphResult{}, err
	}
	goData := C.GoBytes(data, C.int(bmpSize))

	// Convert premultiplied RGBA to white + alpha for tinting.
	for i := 0; i < len(goData); i += 4 {
		a := goData[i+3]
		goData[i+0] = 255
		goData[i+1] = 255
		goData[i+2] = 255
		goData[i+3] = a
	}

	bmp := Bitmap{
		Width:    width,
		Height:   height,
		Channels: 4,
		Data:     goData,
	}

	cached, resetOccurred, resetPage, err := atlas.InsertBitmap(
		bmp, int(left), int(top))
	if err != nil {
		return LoadGlyphResult{}, err
	}

	return LoadGlyphResult{
		Cached:        cached,
		ResetOccurred: resetOccurred,
		ResetPage:     resetPage,
	}, nil
}

// loadStrokedGlyphCG rasterizes a stroked character.
func loadStrokedGlyphCG(atlas *GlyphAtlas, ch string, item Item,
	strokeWidth float32, subpixelBin int,
	scaleFactor float32) (LoadGlyphResult, error) {

	family, fontSize, bold, italic := resolveCTFontParams(
		item.Style, scaleFactor)

	cText := C.CString(ch)
	defer C.free(unsafe.Pointer(cText))
	cFamily := C.CString(family)
	defer C.free(unsafe.Pointer(cFamily))

	var w, h, left, top C.int
	subpixelShift := C.CGFloat(float64(subpixelBin) / 4.0)

	data := C.cgRenderStrokedGlyph(cText, cFamily,
		C.CGFloat(fontSize), C.bool(bold), C.bool(italic),
		C.CGFloat(float64(strokeWidth)*float64(scaleFactor)),
		subpixelShift,
		&w, &h, &left, &top)

	if data == nil || w == 0 || h == 0 {
		return LoadGlyphResult{}, nil
	}
	defer C.free(data)

	width := int(w)
	height := int(h)
	bmpSize, err := checkAllocationSize(width, height, 4)
	if err != nil {
		return LoadGlyphResult{}, err
	}
	goData := C.GoBytes(data, C.int(bmpSize))

	for i := 0; i < len(goData); i += 4 {
		a := goData[i+3]
		goData[i+0] = 255
		goData[i+1] = 255
		goData[i+2] = 255
		goData[i+3] = a
	}

	bmp := Bitmap{
		Width:    width,
		Height:   height,
		Channels: 4,
		Data:     goData,
	}

	cached, resetOccurred, resetPage, err := atlas.InsertBitmap(
		bmp, int(left), int(top))
	if err != nil {
		return LoadGlyphResult{}, err
	}

	return LoadGlyphResult{
		Cached:        cached,
		ResetOccurred: resetOccurred,
		ResetPage:     resetPage,
	}, nil
}

// Bicubic scaling helpers (same as native).

func cubicHermite(p0, p1, p2, p3, t float32) float32 {
	a := -0.5*p0 + 1.5*p1 - 1.5*p2 + 0.5*p3
	b := p0 - 2.5*p1 + 2.0*p2 - 0.5*p3
	c := -0.5*p0 + 0.5*p2
	d := p1
	return a*t*t*t + b*t*t + c*t + d
}

func getPixelRGBAPremul(src []byte, w, h, x, y int) (r, g, b, a float32) {
	if w <= 0 || h <= 0 {
		return
	}
	cx := max(0, min(x, w-1))
	cy := max(0, min(y, h-1))
	idx := (cy*w + cx) * 4
	if idx < 0 || idx+3 >= len(src) {
		return
	}
	rr := float32(src[idx+0])
	gg := float32(src[idx+1])
	bb := float32(src[idx+2])
	aa := float32(src[idx+3])
	f := aa / 255.0
	return rr * f, gg * f, bb * f, aa
}

func ScaleBitmapBicubic(src []byte, srcW, srcH, dstW, dstH int) []byte {
	if dstW <= 0 || dstH <= 0 || srcW <= 0 || srcH <= 0 {
		return nil
	}
	dstSize := int64(dstW) * int64(dstH) * 4
	if dstSize > int64(math.MaxInt32) || dstSize <= 0 {
		return nil
	}

	dst := make([]byte, dstSize)
	xScale := float32(srcW) / float32(dstW)
	yScale := float32(srcH) / float32(dstH)

	for y := range dstH {
		srcY := float32(y) * yScale
		y0 := int(srcY)
		yDiff := srcY - float32(y0)

		for x := range dstW {
			srcX := float32(x) * xScale
			x0 := int(srcX)
			xDiff := srcX - float32(x0)

			dstIdx := (y*dstW + x) * 4

			var colR, colG, colB, colA [4]float32

			for i := -1; i <= 2; i++ {
				rowY := y0 + i
				r0, g0, b0, a0 := getPixelRGBAPremul(src, srcW, srcH, x0-1, rowY)
				r1, g1, b1, a1 := getPixelRGBAPremul(src, srcW, srcH, x0+0, rowY)
				r2, g2, b2, a2 := getPixelRGBAPremul(src, srcW, srcH, x0+1, rowY)
				r3, g3, b3, a3 := getPixelRGBAPremul(src, srcW, srcH, x0+2, rowY)

				j := i + 1
				colR[j] = cubicHermite(r0, r1, r2, r3, xDiff)
				colG[j] = cubicHermite(g0, g1, g2, g3, xDiff)
				colB[j] = cubicHermite(b0, b1, b2, b3, xDiff)
				colA[j] = cubicHermite(a0, a1, a2, a3, xDiff)
			}

			finalR := cubicHermite(colR[0], colR[1], colR[2], colR[3], yDiff)
			finalG := cubicHermite(colG[0], colG[1], colG[2], colG[3], yDiff)
			finalB := cubicHermite(colB[0], colB[1], colB[2], colB[3], yDiff)
			finalA := cubicHermite(colA[0], colA[1], colA[2], colA[3], yDiff)

			finalA = max(0, min(finalA, 255))

			if finalA > 0 {
				f := 255.0 / finalA
				finalR *= f
				finalG *= f
				finalB *= f
			}

			dst[dstIdx+0] = byte(max(0, min(finalR, 255)))
			dst[dstIdx+1] = byte(max(0, min(finalG, 255)))
			dst[dstIdx+2] = byte(max(0, min(finalB, 255)))
			dst[dstIdx+3] = byte(finalA)
		}
	}
	return dst
}
