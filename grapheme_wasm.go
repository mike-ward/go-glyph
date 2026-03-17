//go:build js && wasm

package glyph

import (
	"syscall/js"
	"unicode/utf8"
)

// graphemeCluster represents one user-perceived character.
type graphemeCluster struct {
	text  string
	byteI int
	byteL int
}

var (
	segChecked   bool
	segAvailable bool
	segmenter    js.Value
)

// getGraphemeSegmenter returns a cached Intl.Segmenter for
// grapheme cluster segmentation. Returns false if the browser
// does not support the API.
func getGraphemeSegmenter() (js.Value, bool) {
	if segChecked {
		return segmenter, segAvailable
	}
	segChecked = true
	intl := js.Global().Get("Intl")
	if intl.IsUndefined() || intl.IsNull() {
		return js.Value{}, false
	}
	cls := intl.Get("Segmenter")
	if cls.IsUndefined() || cls.IsNull() {
		return js.Value{}, false
	}
	opts := js.Global().Get("Object").New()
	opts.Set("granularity", "grapheme")
	segmenter = cls.New(js.Undefined(), opts)
	segAvailable = true
	return segmenter, true
}

// segmentGraphemes splits text into grapheme clusters using the
// browser's Intl.Segmenter API. Falls back to rune-by-rune
// splitting when the API is unavailable.
func segmentGraphemes(text string) []graphemeCluster {
	seg, ok := getGraphemeSegmenter()
	if !ok {
		return segmentByRune(text)
	}
	segments := seg.Call("segment", text)
	arr := js.Global().Get("Array").Call("from", segments)
	n := arr.Length()
	clusters := make([]graphemeCluster, n)
	byteIdx := 0
	for i := range n {
		s := arr.Index(i).Get("segment").String()
		clusters[i] = graphemeCluster{
			text:  s,
			byteI: byteIdx,
			byteL: len(s),
		}
		byteIdx += len(s)
	}
	return clusters
}

// segmentByRune is the fallback when Intl.Segmenter is
// unavailable. Each rune becomes its own cluster.
func segmentByRune(text string) []graphemeCluster {
	clusters := make([]graphemeCluster, 0,
		utf8.RuneCountInString(text))
	byteIdx := 0
	for _, r := range text {
		s := string(r)
		clusters = append(clusters, graphemeCluster{
			text:  s,
			byteI: byteIdx,
			byteL: len(s),
		})
		byteIdx += len(s)
	}
	return clusters
}

// glyphText extracts the original cluster text for a glyph.
// In WASM layout, Index stores the byte offset and Codepoint
// stores the byte length into the layout text.
func glyphText(text string, g Glyph) string {
	start := int(g.Index)
	end := start + int(g.Codepoint)
	if start >= 0 && end <= len(text) {
		return text[start:end]
	}
	return ""
}
