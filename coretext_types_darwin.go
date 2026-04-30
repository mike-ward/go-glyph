//go:build darwin && !glyph_pango

package glyph

/*
#include <CoreText/CoreText.h>
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>

// ctFontCreateFromStyle creates a CTFont from family name, size,
// and symbolic traits.
static CTFontRef ctCreateFont(const char *family, CGFloat size,
    bool bold, bool italic) {
    CFStringRef fam = CFStringCreateWithCString(NULL, family,
        kCFStringEncodingUTF8);
    CTFontRef base = CTFontCreateWithName(fam, size, NULL);
    CFRelease(fam);
    if (!bold && !italic) return base;

    CTFontSymbolicTraits traits = 0;
    if (bold) traits |= kCTFontBoldTrait;
    if (italic) traits |= kCTFontItalicTrait;
    CTFontRef styled = CTFontCreateCopyWithSymbolicTraits(
        base, size, NULL, traits, traits);
    if (styled) {
        CFRelease(base);
        return styled;
    }
    // Trait application failed; return base font as-is.
    return base;
}

// ctFontGetMetrics returns ascent, descent, leading.
static void ctFontGetMetrics(CTFontRef font,
    CGFloat *ascent, CGFloat *descent, CGFloat *leading) {
    *ascent = CTFontGetAscent(font);
    *descent = CTFontGetDescent(font);
    *leading = CTFontGetLeading(font);
}

// ctApplyOpenTypeFeatures returns a copy of base with the supplied
// OpenType feature tags applied. tagsBlob is count*4 bytes of ASCII
// tag chars; values is count int32 enable values. Releases base on
// success and returns the styled copy. Returns base unchanged on
// failure.
static CTFontRef ctApplyOpenTypeFeatures(CTFontRef base,
    const char *tagsBlob, const int *values, int count) {
    if (!base || count <= 0) {
        return base;
    }
    CFMutableArrayRef arr = CFArrayCreateMutable(NULL, count,
        &kCFTypeArrayCallBacks);
    for (int i = 0; i < count; i++) {
        char tag[5];
        memcpy(tag, tagsBlob + i*4, 4);
        tag[4] = 0;
        CFStringRef cTag = CFStringCreateWithCString(NULL, tag,
            kCFStringEncodingASCII);
        int v = values[i];
        CFNumberRef cVal = CFNumberCreate(NULL, kCFNumberIntType, &v);
        const void *keys[2] = {
            (const void *)kCTFontOpenTypeFeatureTag,
            (const void *)kCTFontOpenTypeFeatureValue,
        };
        const void *vals[2] = { cTag, cVal };
        CFDictionaryRef d = CFDictionaryCreate(NULL, keys, vals, 2,
            &kCFTypeDictionaryKeyCallBacks,
            &kCFTypeDictionaryValueCallBacks);
        CFArrayAppendValue(arr, d);
        CFRelease(cTag);
        CFRelease(cVal);
        CFRelease(d);
    }
    const void *attrKeys[1] = {
        (const void *)kCTFontFeatureSettingsAttribute,
    };
    const void *attrVals[1] = { arr };
    CFDictionaryRef attrs = CFDictionaryCreate(NULL,
        attrKeys, attrVals, 1,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks);
    CTFontDescriptorRef desc = CTFontDescriptorCreateWithAttributes(attrs);
    CFRelease(arr);
    CFRelease(attrs);
    CTFontRef styled = CTFontCreateCopyWithAttributes(base,
        CTFontGetSize(base), NULL, desc);
    CFRelease(desc);
    if (styled) {
        CFRelease(base);
        return styled;
    }
    return base;
}
*/
import "C"
import (
	"strings"
	"unsafe"
)

// ctFont wraps a CTFontRef with a Go-friendly interface.
type ctFont struct {
	ref C.CTFontRef
}

func resolveCTFontParams(style TextStyle, scaleFactor float32) (
	family string, size float64, bold, italic bool,
) {
	family = resolveFontFamilyDarwin(style.FontName)

	rawSize := style.Size
	if rawSize <= 0 {
		rawSize = parseSizeFromFontName(style.FontName)
	}
	if rawSize <= 0 {
		rawSize = 16
	}
	size = float64(rawSize) * float64(scaleFactor)

	bold = style.Typeface == TypefaceBold ||
		style.Typeface == TypefaceBoldItalic
	italic = style.Typeface == TypefaceItalic ||
		style.Typeface == TypefaceBoldItalic

	lower := strings.ToLower(style.FontName)
	if !bold && strings.Contains(lower, " bold") {
		bold = true
	}
	if !italic && strings.Contains(lower, " italic") {
		italic = true
	}

	return family, size, bold, italic
}

// newCTFont creates a Core Text font from a TextStyle.
func newCTFont(style TextStyle, scaleFactor float32) ctFont {
	family, size, bold, italic := resolveCTFontParams(style, scaleFactor)

	cFamily := C.CString(family)
	defer C.free(unsafe.Pointer(cFamily))

	ref := C.ctCreateFont(cFamily,
		C.CGFloat(size),
		C.bool(bold), C.bool(italic))
	if ref != 0 && style.Features != nil {
		ref = applyOpenTypeFeatures(ref, style.Features.OpenTypeFeatures)
	}
	return ctFont{ref: ref}
}

// applyOpenTypeFeatures applies user-supplied OpenType feature tags
// (e.g. liga, tnum, calt) to the font via
// kCTFontFeatureSettingsAttribute. Returns the styled CTFontRef
// (may equal base on failure or empty input). The "subs" / "sups"
// tags are skipped here because LayoutRichText handles them by
// scaling the font size and shifting the baseline; passing them
// through to CoreText as well would double-apply.
func applyOpenTypeFeatures(base C.CTFontRef, feats []FontFeature) C.CTFontRef {
	if len(feats) == 0 {
		return base
	}
	tags := make([]byte, 0, len(feats)*4)
	vals := make([]C.int, 0, len(feats))
	for _, f := range feats {
		if f.Tag == "subs" || f.Tag == "sups" {
			continue
		}
		t := f.Tag
		switch len(t) {
		case 4:
			tags = append(tags, t[0], t[1], t[2], t[3])
		case 0, 1, 2, 3:
			padded := []byte(t + "    ")[:4]
			tags = append(tags, padded...)
		default:
			tags = append(tags, t[0], t[1], t[2], t[3])
		}
		vals = append(vals, C.int(f.Value))
	}
	count := len(vals)
	if count == 0 {
		return base
	}
	return C.ctApplyOpenTypeFeatures(base,
		(*C.char)(unsafe.Pointer(&tags[0])),
		(*C.int)(unsafe.Pointer(&vals[0])),
		C.int(count))
}

// close releases the CTFont.
func (f *ctFont) close() {
	if f.ref != 0 {
		C.CFRelease(C.CFTypeRef(f.ref))
		f.ref = 0
	}
}

// metrics returns ascent, descent, leading in Core Text units.
func (f ctFont) metrics() (ascent, descent, leading float64) {
	var a, d, l C.CGFloat
	C.ctFontGetMetrics(f.ref, &a, &d, &l)
	return float64(a), float64(d), float64(l)
}

// resolveFontFamilyDarwin maps generic Pango-style font names to
// macOS / iOS system font families. SF Mono ships in 10.15+; older
// macOS targets fall back to Menlo via CoreText's font matcher when
// "SF Mono" is unavailable.
func resolveFontFamilyDarwin(fontName string) string {
	family := parseFamilyFromFontName(fontName)
	if family == "" {
		return ".AppleSystemUIFont"
	}
	switch strings.ToLower(family) {
	case "sans", "sans-serif", "system":
		return ".AppleSystemUIFont"
	case "serif":
		return "New York"
	case "monospace", "mono":
		return "SF Mono"
	default:
		return family
	}
}

// parseSizeFromFontName extracts trailing numeric size from Pango
// font name like "Sans Bold 18".
func parseSizeFromFontName(name string) float32 {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return 0
	}
	last := parts[len(parts)-1]
	var sz float32
	for _, c := range last {
		if c >= '0' && c <= '9' {
			sz = sz*10 + float32(c-'0')
		} else if c == '.' {
			// Skip fractional parsing for simplicity.
			break
		} else {
			return 0
		}
	}
	return sz
}

// parseFamilyFromFontName extracts the family portion from a Pango
// font name, stripping trailing size and style keywords.
func parseFamilyFromFontName(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return ""
	}

	end := len(parts)
	// Strip trailing number (size).
	if sz := parseSizeFromFontName(name); sz > 0 {
		end--
	}

	styleWords := map[string]bool{
		"bold": true, "italic": true, "oblique": true,
		"light": true, "medium": true, "semibold": true,
		"heavy": true, "ultrabold": true, "ultralight": true,
		"condensed": true, "expanded": true, "regular": true,
	}
	for end > 0 && styleWords[strings.ToLower(parts[end-1])] {
		end--
	}
	if end == 0 {
		end = 1
	}
	return strings.Join(parts[:end], " ")
}
