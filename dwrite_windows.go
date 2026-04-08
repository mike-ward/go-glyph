//go:build windows

package glyph

/*
#cgo LDFLAGS: -ldwrite -lgdi32

#define COBJMACROS
#define CINTERFACE
#include <initguid.h>
#include <windows.h>
#include <dwrite_3.h>
#include <stdlib.h>
#include <string.h>

// MinGW headers do not always define DWRITE_E_NOCOLOR. The actual
// runtime value (verified empirically against Windows 10 22H2) is
// 0x8898500CL — Microsoft documentation lists 0x8898500DL but DWrite
// returns 0x8898500CL in practice for "no color glyph for this run".
#ifndef DWRITE_E_NOCOLOR
#define DWRITE_E_NOCOLOR ((HRESULT)0x8898500CL)
#endif

// DWriteCtx owns all long-lived DirectWrite COM objects. Created once
// per process and shared across the Windows GDI context. Guarded by a
// Go-side mutex because the bitmap render target / glyph analysis
// operations are not safe for concurrent use.
//
// Factory4 (Windows 10 1709+) is required because Segoe UI Emoji on
// modern Windows 10/11 ships as COLR v1; Factory2's TranslateColorGlyphRun
// returns S_OK with zero layers for v1 glyphs. Factory4 down-converts
// COLR v1 paint trees into flat COLR alpha layers when the caller
// requests COLR but not COLR_PAINT_TREE.
typedef struct DWriteCtx {
    IDWriteFactory4*       factory;
    IDWriteFontCollection* sysCollection;
    IDWriteFontFace*       emojiFace;
    int                    emojiFaceValid;
} DWriteCtx;

static void dwrite_ctx_free(DWriteCtx* ctx) {
    if (!ctx) return;
    if (ctx->emojiFace) {
        IDWriteFontFace_Release(ctx->emojiFace);
    }
    if (ctx->sysCollection) {
        IDWriteFontCollection_Release(ctx->sysCollection);
    }
    if (ctx->factory) {
        IDWriteFactory4_Release(ctx->factory);
    }
    free(ctx);
}

// dwrite_ctx_new creates a DirectWrite factory and preloads the
// Segoe UI Emoji font face. Returns NULL on any failure (caller
// silently falls back to GDI rendering).
static DWriteCtx* dwrite_ctx_new(void) {
    DWriteCtx* ctx = (DWriteCtx*)calloc(1, sizeof(DWriteCtx));
    if (!ctx) return NULL;

    HRESULT hr = DWriteCreateFactory(DWRITE_FACTORY_TYPE_SHARED,
        &IID_IDWriteFactory4, (IUnknown**)&ctx->factory);
    if (FAILED(hr) || !ctx->factory) {
        free(ctx);
        return NULL;
    }

    // Use the base IDWriteFactory's GetSystemFontCollection for the
    // legacy collection type. Factory3+ has a different signature
    // returning IDWriteFontCollection1.
    hr = IDWriteFactory_GetSystemFontCollection(
        (IDWriteFactory*)ctx->factory, &ctx->sysCollection, FALSE);
    if (FAILED(hr) || !ctx->sysCollection) {
        dwrite_ctx_free(ctx);
        return NULL;
    }

    // Preload the Segoe UI Emoji font face. This is the only font we
    // use for color glyph rendering — the assumption is that anything
    // isEmojiRune() flags should be rendered from this font.
    const WCHAR* name = L"Segoe UI Emoji";
    UINT32 idx = 0;
    BOOL exists = FALSE;
    hr = IDWriteFontCollection_FindFamilyName(ctx->sysCollection,
        name, &idx, &exists);
    if (FAILED(hr) || !exists) {
        dwrite_ctx_free(ctx);
        return NULL;
    }

    IDWriteFontFamily* fam = NULL;
    hr = IDWriteFontCollection_GetFontFamily(ctx->sysCollection, idx, &fam);
    if (FAILED(hr) || !fam) {
        dwrite_ctx_free(ctx);
        return NULL;
    }

    IDWriteFont* font = NULL;
    hr = IDWriteFontFamily_GetFirstMatchingFont(fam,
        DWRITE_FONT_WEIGHT_NORMAL,
        DWRITE_FONT_STRETCH_NORMAL,
        DWRITE_FONT_STYLE_NORMAL, &font);
    IDWriteFontFamily_Release(fam);
    if (FAILED(hr) || !font) {
        dwrite_ctx_free(ctx);
        return NULL;
    }

    hr = IDWriteFont_CreateFontFace(font, &ctx->emojiFace);
    IDWriteFont_Release(font);
    if (FAILED(hr) || !ctx->emojiFace) {
        dwrite_ctx_free(ctx);
        return NULL;
    }

    ctx->emojiFaceValid = 1;
    return ctx;
}

// Return codes for dwrite_render_color_glyph.
#define DWRITE_RENDER_OK       0
#define DWRITE_RENDER_FAIL     1
#define DWRITE_RENDER_NOCOLOR  2

// DWriteLayer holds the rasterized alpha texture for a single color
// glyph layer along with its bounds and color. Used during composition.
typedef struct DWriteLayer {
    IDWriteGlyphRunAnalysis* analysis;
    RECT                     bounds;
    DWRITE_COLOR_F           color;
} DWriteLayer;

static void dwrite_free_layers(DWriteLayer* layers, int count) {
    if (!layers) return;
    for (int i = 0; i < count; i++) {
        if (layers[i].analysis) {
            IDWriteGlyphRunAnalysis_Release(layers[i].analysis);
        }
    }
    free(layers);
}

// dwrite_render_color_glyph rasterizes a single codepoint into a
// premultiplied BGRA bitmap. The returned buffer must be released
// via dwrite_free_pixels.
//
// On success, *pixels is a heap buffer of (*w * *h * 4) bytes in
// premultiplied BGRA layout. *left and *top give the glyph bearing
// in pixels relative to the pen origin / baseline (FreeType convention:
// left = pen-X to leftmost pixel, top = baseline-Y up to topmost pixel).
static int dwrite_render_color_glyph(
    DWriteCtx* ctx,
    float emSizePx,
    unsigned int codepoint,
    unsigned char** pixels,
    int* w, int* h,
    int* left, int* top
) {
    if (!ctx || !ctx->emojiFaceValid || emSizePx <= 0.0f) {
        return DWRITE_RENDER_FAIL;
    }

    UINT32 cp = (UINT32)codepoint;
    UINT16 glyphIndex = 0;
    HRESULT hr = IDWriteFontFace_GetGlyphIndices(ctx->emojiFace,
        &cp, 1, &glyphIndex);
    if (FAILED(hr) || glyphIndex == 0) {
        return DWRITE_RENDER_NOCOLOR;
    }

    FLOAT advance = 0.0f;
    DWRITE_GLYPH_OFFSET goffset = {0.0f, 0.0f};
    DWRITE_GLYPH_RUN run;
    memset(&run, 0, sizeof(run));
    run.fontFace      = ctx->emojiFace;
    run.fontEmSize    = emSizePx;
    run.glyphCount    = 1;
    run.glyphIndices  = &glyphIndex;
    run.glyphAdvances = &advance;
    run.glyphOffsets  = &goffset;
    run.isSideways    = FALSE;
    run.bidiLevel     = 0;

    // Use Factory4's TranslateColorGlyphRun (Win10 1709+). We omit
    // COLR_PAINT_TREE because (a) older Windows runtimes reject it as
    // E_INVALIDARG and (b) we cannot rasterize paint trees via
    // CreateGlyphRunAnalysis anyway. Including the bitmap formats
    // tells DWrite to surface CBDT/sbix layers as well.
    D2D1_POINT_2F baselineOrigin = {0.0f, 0.0f};
    DWRITE_GLYPH_IMAGE_FORMATS desiredFormats =
        DWRITE_GLYPH_IMAGE_FORMATS_TRUETYPE |
        DWRITE_GLYPH_IMAGE_FORMATS_CFF |
        DWRITE_GLYPH_IMAGE_FORMATS_COLR |
        DWRITE_GLYPH_IMAGE_FORMATS_SVG |
        DWRITE_GLYPH_IMAGE_FORMATS_PNG |
        DWRITE_GLYPH_IMAGE_FORMATS_JPEG |
        DWRITE_GLYPH_IMAGE_FORMATS_TIFF |
        DWRITE_GLYPH_IMAGE_FORMATS_PREMULTIPLIED_B8G8R8A8;
    IDWriteColorGlyphRunEnumerator1* colorEnum = NULL;
    hr = IDWriteFactory4_TranslateColorGlyphRun(ctx->factory,
        baselineOrigin,
        &run,
        NULL,
        desiredFormats,
        DWRITE_MEASURING_MODE_NATURAL,
        NULL,
        0,
        &colorEnum);
    if (hr == DWRITE_E_NOCOLOR) {
        return DWRITE_RENDER_NOCOLOR;
    }
    if (FAILED(hr) || !colorEnum) {
        return DWRITE_RENDER_FAIL;
    }

    DWriteLayer* layers = NULL;
    int layerCount = 0;
    int layerCap = 0;
    int unionL = 0, unionT = 0, unionR = 0, unionB = 0;
    int haveUnion = 0;

    for (;;) {
        BOOL haveRun = FALSE;
        hr = IDWriteColorGlyphRunEnumerator1_MoveNext(colorEnum, &haveRun);
        if (FAILED(hr) || !haveRun) break;

        DWRITE_COLOR_GLYPH_RUN1 const* colorRun = NULL;
        hr = IDWriteColorGlyphRunEnumerator1_GetCurrentRun(colorEnum, &colorRun);
        if (FAILED(hr) || !colorRun || colorRun->glyphRun.glyphCount == 0) {
            continue;
        }

        // Skip layers we cannot rasterize via CreateGlyphRunAnalysis.
        // We can only consume vector outlines (TRUETYPE/CFF/COLR);
        // bitmap and SVG formats need different decoders.
        if (colorRun->glyphImageFormat &
            (DWRITE_GLYPH_IMAGE_FORMATS_PNG |
             DWRITE_GLYPH_IMAGE_FORMATS_JPEG |
             DWRITE_GLYPH_IMAGE_FORMATS_TIFF |
             DWRITE_GLYPH_IMAGE_FORMATS_PREMULTIPLIED_B8G8R8A8 |
             DWRITE_GLYPH_IMAGE_FORMATS_SVG |
             DWRITE_GLYPH_IMAGE_FORMATS_COLR_PAINT_TREE)) {
            continue;
        }

        IDWriteGlyphRunAnalysis* ana = NULL;
        // Use the base IDWriteFactory signature (pixelsPerDip variant).
        // Factory3+ override has a different parameter list and its
        // COBJMACRO in MinGW's dwrite_3.h is emitted with a broken name,
        // so we cast down to the base interface and call that slot.
        // NATURAL_SYMMETRIC produces a ClearType-style 3x1 texture that
        // we average to grayscale below — we can't use ALIASED_1x1
        // because that only works with the ALIASED rendering mode.
        hr = IDWriteFactory_CreateGlyphRunAnalysis(
            (IDWriteFactory*)ctx->factory,
            &colorRun->glyphRun,
            1.0f,
            NULL,
            DWRITE_RENDERING_MODE_NATURAL_SYMMETRIC,
            DWRITE_MEASURING_MODE_NATURAL,
            colorRun->baselineOriginX,
            colorRun->baselineOriginY,
            &ana);
        if (FAILED(hr) || !ana) continue;

        RECT rc = {0, 0, 0, 0};
        hr = IDWriteGlyphRunAnalysis_GetAlphaTextureBounds(ana,
            DWRITE_TEXTURE_CLEARTYPE_3x1, &rc);
        if (FAILED(hr) || rc.right <= rc.left || rc.bottom <= rc.top) {
            IDWriteGlyphRunAnalysis_Release(ana);
            continue;
        }

        if (layerCount >= layerCap) {
            int newCap = layerCap == 0 ? 4 : layerCap * 2;
            DWriteLayer* nl = (DWriteLayer*)realloc(layers,
                (size_t)newCap * sizeof(DWriteLayer));
            if (!nl) {
                IDWriteGlyphRunAnalysis_Release(ana);
                break;
            }
            layers = nl;
            layerCap = newCap;
        }
        DWriteLayer* L = &layers[layerCount++];
        L->analysis = ana;
        L->bounds   = rc;
        L->color    = colorRun->runColor;

        if (!haveUnion) {
            unionL = rc.left;   unionT = rc.top;
            unionR = rc.right;  unionB = rc.bottom;
            haveUnion = 1;
        } else {
            if (rc.left   < unionL) unionL = rc.left;
            if (rc.top    < unionT) unionT = rc.top;
            if (rc.right  > unionR) unionR = rc.right;
            if (rc.bottom > unionB) unionB = rc.bottom;
        }
    }
    IDWriteColorGlyphRunEnumerator1_Release(colorEnum);

    if (!haveUnion || layerCount == 0) {
        dwrite_free_layers(layers, layerCount);
        return DWRITE_RENDER_NOCOLOR;
    }

    int outW = unionR - unionL;
    int outH = unionB - unionT;
    if (outW <= 0 || outH <= 0 || outW > 4096 || outH > 4096) {
        dwrite_free_layers(layers, layerCount);
        return DWRITE_RENDER_FAIL;
    }

    size_t accBytes = (size_t)outW * (size_t)outH * 4;
    unsigned char* acc = (unsigned char*)calloc(1, accBytes);
    if (!acc) {
        dwrite_free_layers(layers, layerCount);
        return DWRITE_RENDER_FAIL;
    }

    // Composite each layer onto the accumulator in premultiplied BGRA.
    // Layer color is straight (non-premultiplied) DWRITE_COLOR_F.
    // CreateAlphaTexture with CLEARTYPE_3x1 returns 3 bytes per pixel
    // (subpixel coverage); we average those to a single grayscale alpha.
    for (int i = 0; i < layerCount; i++) {
        DWriteLayer* L = &layers[i];
        int lw = L->bounds.right - L->bounds.left;
        int lh = L->bounds.bottom - L->bounds.top;
        if (lw <= 0 || lh <= 0) continue;

        size_t bufSize = (size_t)lw * (size_t)lh * 3;
        unsigned char* alphaBuf = (unsigned char*)malloc(bufSize);
        if (!alphaBuf) continue;

        hr = IDWriteGlyphRunAnalysis_CreateAlphaTexture(L->analysis,
            DWRITE_TEXTURE_CLEARTYPE_3x1, &L->bounds,
            alphaBuf, (UINT32)bufSize);
        if (FAILED(hr)) {
            free(alphaBuf);
            continue;
        }

        float cr = L->color.r;
        float cg = L->color.g;
        float cb = L->color.b;
        float ca = L->color.a;
        // A layer with its runColor.a == -1.0f (encoded as 0 in some
        // DWrite builds) signals "use foreground color". Treat that
        // plus any fully-zero color as opaque white so the layer at
        // least contributes coverage.
        if (ca <= 0.0f && cr <= 0.0f && cg <= 0.0f && cb <= 0.0f) {
            cr = 1.0f; cg = 1.0f; cb = 1.0f; ca = 1.0f;
        }
        if (ca < 0.0f) ca = 1.0f;
        if (ca > 1.0f) ca = 1.0f;

        int offX = L->bounds.left - unionL;
        int offY = L->bounds.top  - unionT;

        for (int py = 0; py < lh; py++) {
            int dy = py + offY;
            if (dy < 0 || dy >= outH) continue;
            const unsigned char* srcRow = alphaBuf + (size_t)py * lw * 3;
            unsigned char* dstRow = acc + ((size_t)dy * outW + offX) * 4;
            for (int px = 0; px < lw; px++) {
                // Average the three subpixel coverage bytes to get a
                // grayscale alpha (we don't want subpixel artifacts in
                // a color emoji texture).
                unsigned int a8 =
                    ((unsigned int)srcRow[px*3 + 0] +
                     (unsigned int)srcRow[px*3 + 1] +
                     (unsigned int)srcRow[px*3 + 2]) / 3;
                if (a8 == 0) continue;

                // Total coverage for this pixel, 0..1.
                float pa = ((float)a8 / 255.0f) * ca;
                // Premultiplied source in 0..255.
                int srcA = (int)(pa * 255.0f + 0.5f);
                int srcB = (int)(pa * cb * 255.0f + 0.5f);
                int srcG = (int)(pa * cg * 255.0f + 0.5f);
                int srcR = (int)(pa * cr * 255.0f + 0.5f);
                if (srcA > 255) srcA = 255;
                if (srcB > 255) srcB = 255;
                if (srcG > 255) srcG = 255;
                if (srcR > 255) srcR = 255;

                unsigned char* dp = dstRow + (size_t)px * 4;
                int invA = 255 - srcA;
                int nB = srcB + ((int)dp[0] * invA + 127) / 255;
                int nG = srcG + ((int)dp[1] * invA + 127) / 255;
                int nR = srcR + ((int)dp[2] * invA + 127) / 255;
                int nA = srcA + ((int)dp[3] * invA + 127) / 255;
                if (nB > 255) nB = 255;
                if (nG > 255) nG = 255;
                if (nR > 255) nR = 255;
                if (nA > 255) nA = 255;
                dp[0] = (unsigned char)nB;
                dp[1] = (unsigned char)nG;
                dp[2] = (unsigned char)nR;
                dp[3] = (unsigned char)nA;
            }
        }

        free(alphaBuf);
    }

    dwrite_free_layers(layers, layerCount);

    *pixels = acc;
    *w      = outW;
    *h      = outH;
    // unionL, unionT are in pixel-space with origin at the pen baseline.
    // FreeType bitmap_left  = pixels from pen-X to leftmost glyph pixel.
    // FreeType bitmap_top   = pixels from baseline UP to topmost pixel.
    *left   = unionL;
    *top    = -unionT;
    return DWRITE_RENDER_OK;
}

static void dwrite_free_pixels(unsigned char* p) { free(p); }
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"
)

// dwriteRasterizer renders color (COLR) glyphs via DirectWrite, which
// classic GDI cannot do. Used as the emoji rendering path on Windows.
//
// All DirectWrite operations go through a single process-wide factory
// and are serialized by mu — IDWriteGlyphRunAnalysis is not documented
// as thread-safe and we'd rather pay the lock cost than debug a rare
// corruption later.
type dwriteRasterizer struct {
	mu  sync.Mutex
	ctx *C.DWriteCtx
}

// errNoColorGlyph signals that the requested codepoint has no color
// glyph in the Segoe UI Emoji font. The caller should fall back to GDI
// monochrome rendering.
var errNoColorGlyph = errors.New("glyph: no color glyph for codepoint")

// newDWriteRasterizer initializes the DirectWrite factory and preloads
// the Segoe UI Emoji font face. Returns an error on any failure; the
// caller should fall back to GDI-only rendering in that case.
func newDWriteRasterizer() (*dwriteRasterizer, error) {
	ctx := C.dwrite_ctx_new()
	if ctx == nil {
		return nil, errors.New("glyph: DirectWrite initialization failed")
	}
	return &dwriteRasterizer{ctx: ctx}, nil
}

// Close releases all DirectWrite COM objects.
func (d *dwriteRasterizer) Close() {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ctx != nil {
		C.dwrite_ctx_free(d.ctx)
		d.ctx = nil
	}
}

// RenderColorGlyph rasterizes a single color glyph. emSizePx is the
// font em size in physical pixels (already scaled by DPI). The returned
// Bitmap holds RGBA premultiplied pixels and uses the same bearing
// convention as the FreeType/GDI paths (left = pen-X → left edge,
// top = baseline → top edge, positive = above baseline).
func (d *dwriteRasterizer) RenderColorGlyph(
	emSizePx float32, codepoint rune,
) (Bitmap, int, int, error) {
	if d == nil || d.ctx == nil {
		return Bitmap{}, 0, 0, errors.New("glyph: DirectWrite not initialized")
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	var (
		cPixels *C.uchar
		cW, cH  C.int
		cLeft   C.int
		cTop    C.int
	)
	rc := C.dwrite_render_color_glyph(
		d.ctx,
		C.float(emSizePx),
		C.uint(codepoint),
		&cPixels,
		&cW, &cH,
		&cLeft, &cTop,
	)
	switch rc {
	case C.DWRITE_RENDER_OK:
		// fall through
	case C.DWRITE_RENDER_NOCOLOR:
		return Bitmap{}, 0, 0, errNoColorGlyph
	default:
		return Bitmap{}, 0, 0, errors.New("glyph: DirectWrite rasterization failed")
	}
	if cPixels == nil {
		return Bitmap{}, 0, 0, errors.New("glyph: DirectWrite returned nil pixels")
	}
	defer C.dwrite_free_pixels(cPixels)

	w := int(cW)
	h := int(cH)
	if w <= 0 || h <= 0 {
		return Bitmap{}, 0, 0, errNoColorGlyph
	}
	total := w * h * 4
	// Copy from C heap, swizzling BGRA → RGBA. The premultiplication
	// from the C side is preserved.
	src := unsafe.Slice((*byte)(unsafe.Pointer(cPixels)), total)
	data := make([]byte, total)
	for i := 0; i < total; i += 4 {
		data[i+0] = src[i+2] // R
		data[i+1] = src[i+1] // G
		data[i+2] = src[i+0] // B
		data[i+3] = src[i+3] // A
	}

	return Bitmap{
		Width:    w,
		Height:   h,
		Channels: 4,
		Data:     data,
	}, int(cLeft), int(cTop), nil
}
