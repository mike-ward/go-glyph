//go:build android

package glyph

/*
#include <ft2build.h>
#include FT_FREETYPE_H
#include <stdlib.h>
*/
import "C"
import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unsafe"
)

// Context holds FreeType state for text shaping on Android.
//
// Not safe for concurrent use.
type Context struct {
	ftLib         C.FT_Library
	scaleFactor   float32
	scaleInv      float32
	metrics       metricsCache
	fontPaths     map[string]string
	fallbackPaths []string // script fallback fonts (CJK, Arabic, etc.)
}

// NewContext creates an Android text context.
func NewContext(scaleFactor float32) (*Context, error) {
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}

	var lib C.FT_Library
	if rc := C.FT_Init_FreeType(&lib); rc != 0 {
		return nil, fmt.Errorf("FT_Init_FreeType failed: %d", rc)
	}

	ctx := &Context{
		ftLib:       lib,
		scaleFactor: scaleFactor,
		scaleInv:    1.0 / scaleFactor,
		metrics:     newMetricsCache(256),
		fontPaths:   make(map[string]string),
	}
	setFTLib(lib)
	ctx.parseSystemFonts()
	setFTFontPaths(ctx.fontPaths)
	setFTScriptFallbacks(ctx.fallbackPaths)
	return ctx, nil
}

// Free releases resources.
func (ctx *Context) Free() {
	if ctx.ftLib != nil {
		C.FT_Done_FreeType(ctx.ftLib)
		ctx.ftLib = nil
	}
	ctx.metrics = metricsCache{}
	ctx.fontPaths = nil
}

// ScaleFactor returns the DPI scale factor.
func (ctx *Context) ScaleFactor() float32 { return ctx.scaleFactor }

// AddFontFile registers a font file with FreeType by loading it
// temporarily to extract the family name.
func (ctx *Context) AddFontFile(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var face C.FT_Face
	if rc := C.FT_New_Face(ctx.ftLib, cPath, 0, &face); rc != 0 {
		return fmt.Errorf("FT_New_Face failed for %q: %d", path, rc)
	}
	family := C.GoString(face.family_name)
	style := C.GoString(face.style_name)
	C.FT_Done_Face(face)

	key := family + "-" + style
	ctx.fontPaths[key] = path
	ctx.fontPaths[family] = path
	return nil
}

// FontHeight returns ascent + descent in logical pixels.
func (ctx *Context) FontHeight(cfg TextConfig) (float32, error) {
	font := newFTFont(ctx.ftLib, ctx.fontPaths, cfg.Style, ctx.scaleFactor)
	if font.face == nil {
		return 0, fmt.Errorf("failed to create FT font")
	}
	defer font.close()

	ascent, descent, _ := font.metrics()
	return float32(ascent+descent) / ctx.scaleFactor, nil
}

// FontMetrics returns detailed metrics in logical pixels.
func (ctx *Context) FontMetrics(cfg TextConfig) (TextMetrics, error) {
	font := newFTFont(ctx.ftLib, ctx.fontPaths, cfg.Style, ctx.scaleFactor)
	if font.face == nil {
		return TextMetrics{}, fmt.Errorf("failed to create FT font")
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

// ResolveFontName returns the resolved Android font family name.
func (ctx *Context) ResolveFontName(fontDescStr string) (string, error) {
	family := resolveFontFamilyAndroid(fontDescStr)
	return family, nil
}

// createFTFont builds an ftFont from TextStyle. Caller must call
// close().
func (ctx *Context) createFTFont(style TextStyle) ftFont {
	return newFTFont(ctx.ftLib, ctx.fontPaths, style, ctx.scaleFactor)
}

// parseSystemFonts reads /system/etc/fonts.xml to populate the
// fontPaths map.
func (ctx *Context) parseSystemFonts() {
	f, err := os.Open("/system/etc/fonts.xml")
	if err != nil {
		// Fallback: populate known defaults.
		ctx.populateDefaultFontPaths()
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var currentFamily string
	isNamedFamily := false
	seenFallback := make(map[string]bool)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Parse <family name="sans-serif"> or <family> (unnamed).
		if strings.HasPrefix(line, "<family") {
			if idx := strings.Index(line, `name="`); idx >= 0 {
				rest := line[idx+6:]
				if end := strings.Index(rest, `"`); end >= 0 {
					currentFamily = rest[:end]
				}
				isNamedFamily = true
			} else {
				currentFamily = "_fallback_"
				isNamedFamily = false
			}
			continue
		}

		// Parse <font ...>FontFile.ttf</font>
		if strings.HasPrefix(line, "<font") && currentFamily != "" {
			if start := strings.Index(line, ">"); start >= 0 {
				rest := line[start+1:]
				if end := strings.Index(rest, "</font>"); end >= 0 {
					fileName := strings.TrimSpace(rest[:end])
					if fileName == "" {
						continue
					}
					path := "/system/fonts/" + fileName
					if isNamedFamily {
						ctx.fontPaths[currentFamily] = path
						stem := strings.TrimSuffix(fileName, ".ttf")
						stem = strings.TrimSuffix(stem, ".otf")
						stem = strings.TrimSuffix(stem, ".ttc")
						ctx.fontPaths[stem] = path
					} else if !seenFallback[path] {
						ctx.fallbackPaths = append(
							ctx.fallbackPaths, path)
						seenFallback[path] = true
					}
				}
			}
			continue
		}

		if strings.HasPrefix(line, "</family>") {
			currentFamily = ""
			isNamedFamily = false
		}
	}

	// Ensure defaults exist even if XML parsing missed them.
	ctx.populateDefaultFontPaths()
}

// populateDefaultFontPaths sets fallback paths for common fonts.
func (ctx *Context) populateDefaultFontPaths() {
	defaults := map[string]string{
		"Roboto":            "/system/fonts/Roboto-Regular.ttf",
		"Roboto-Regular":    "/system/fonts/Roboto-Regular.ttf",
		"Roboto-Bold":       "/system/fonts/Roboto-Bold.ttf",
		"Roboto-Italic":     "/system/fonts/Roboto-Italic.ttf",
		"Roboto-BoldItalic": "/system/fonts/Roboto-BoldItalic.ttf",
		"NotoSerif":            "/system/fonts/NotoSerif-Regular.ttf",
		"NotoSerif-Regular":    "/system/fonts/NotoSerif-Regular.ttf",
		"NotoSerif-Bold":       "/system/fonts/NotoSerif-Bold.ttf",
		"NotoSerif-Italic":     "/system/fonts/NotoSerif-Italic.ttf",
		"NotoSerif-BoldItalic": "/system/fonts/NotoSerif-BoldItalic.ttf",
		"DroidSansMono":         "/system/fonts/DroidSansMono.ttf",
		"DroidSansMono-Regular": "/system/fonts/DroidSansMono.ttf",
		"sans-serif":            "/system/fonts/Roboto-Regular.ttf",
		"serif":                 "/system/fonts/NotoSerif-Regular.ttf",
		"monospace":             "/system/fonts/DroidSansMono.ttf",
	}
	for k, v := range defaults {
		if _, ok := ctx.fontPaths[k]; !ok {
			ctx.fontPaths[k] = v
		}
	}
}
