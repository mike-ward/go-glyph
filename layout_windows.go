//go:build windows

package glyph

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// LayoutText shapes and wraps text using GDI measureText.
func (ctx *Context) LayoutText(text string, cfg TextConfig) (Layout, error) {
	if len(text) == 0 {
		return Layout{}, nil
	}
	if err := ValidateTextInput(text, MaxTextLength, "LayoutText"); err != nil {
		return Layout{}, err
	}

	if cfg.UseMarkup {
		runs, err := parsePangoMarkup(text, cfg.Style)
		if err != nil {
			runs = []StyleRun{{Text: text, Style: cfg.Style}}
		}
		markupCfg := cfg
		markupCfg.UseMarkup = false
		return ctx.LayoutRichText(RichText{Runs: runs}, markupCfg)
	}

	ctx.selectFont(cfg.Style)
	return ctx.buildLayout(text, cfg, nil), nil
}

// LayoutRichText shapes multi-styled text.
func (ctx *Context) LayoutRichText(rt RichText, cfg TextConfig) (Layout, error) {
	if len(rt.Runs) == 0 {
		return Layout{}, nil
	}
	for _, run := range rt.Runs {
		if err := ValidateTextInput(run.Text, MaxTextLength, "LayoutRichText"); err != nil {
			return Layout{}, err
		}
	}

	// Build full text and per-character style overrides.
	var fullText strings.Builder
	type runRange struct {
		start int
		end   int
		style TextStyle
	}
	ranges := make([]runRange, 0, len(rt.Runs))
	idx := 0
	for _, run := range rt.Runs {
		fullText.WriteString(run.Text)
		n := len(run.Text)
		ranges = append(ranges, runRange{start: idx, end: idx + n, style: run.Style})
		idx += n
	}
	text := fullText.String()
	if len(text) == 0 {
		return Layout{}, nil
	}

	// Build override map: byte-index → style for characters with
	// non-default styles.
	overrides := make(map[int]TextStyle)
	for _, rr := range ranges {
		for i := rr.start; i < rr.end; {
			_, sz := utf8.DecodeRuneInString(text[i:])
			overrides[i] = rr.style
			i += sz
		}
	}

	ctx.selectFont(cfg.Style)
	return ctx.buildLayout(text, cfg, overrides), nil
}

// buildLayout builds a Layout struct by measuring characters with GDI.
func (ctx *Context) buildLayout(text string, cfg TextConfig,
	overrides map[int]TextStyle) Layout {

	if cfg.Orientation == OrientationVertical {
		return ctx.buildVerticalLayout(text, cfg, overrides)
	}

	scale := ctx.scaleFactor
	pixelScale := 1.0 / float64(scale)

	tm := ctx.gdi.getTextMetrics()
	fontAscent := float64(tm.TmAscent)
	fontDescent := float64(tm.TmDescent)
	lineHeight := fontAscent + fontDescent + float64(tm.TmExternalLeading)

	baseColor := cfg.Style.Color
	if baseColor.A == 0 {
		baseColor = Color{0, 0, 0, 255}
	}

	// Segment text into grapheme clusters.
	clusters := segmentGraphemes(text)
	if len(clusters) == 0 {
		return Layout{}
	}

	// Measure each cluster.
	type measuredChar struct {
		text  string
		byteI int
		byteL int
		width float64
	}
	chars := make([]measuredChar, 0, len(clusters))
	var lastMeasureStyle *TextStyle
	for _, cl := range clusters {
		if overrides != nil {
			if s, ok := overrides[cl.byteI]; ok {
				adjusted := winApplySubSup(s)
				if lastMeasureStyle == nil || adjusted != *lastMeasureStyle {
					ctx.selectFont(adjusted)
					sCopy := adjusted
					lastMeasureStyle = &sCopy
				}
			}
		}
		w, _ := ctx.gdi.measureString(cl.text)
		chars = append(chars, measuredChar{
			text:  cl.text,
			byteI: cl.byteI,
			byteL: cl.byteL,
			width: float64(w),
		})
	}
	if lastMeasureStyle != nil {
		ctx.selectFont(cfg.Style) // Restore base font.
	}

	// Apply letter spacing.
	letterSpacing := float64(cfg.Style.LetterSpacing) * float64(scale)

	// Word-wrap into lines.
	maxWidth := float64(cfg.Block.Width) * float64(scale)
	wrapEnabled := cfg.Block.Width > 0 && cfg.Block.Wrap != WrapNone

	type lineInfo struct {
		startChar int
		endChar   int
		width     float64
	}
	var lines []lineInfo
	lineStart := 0
	lineW := 0.0
	lastBreak := -1

	for ci, ch := range chars {
		cw := ch.width + letterSpacing
		if ch.text == "\n" {
			lines = append(lines, lineInfo{startChar: lineStart, endChar: ci, width: lineW})
			lineStart = ci + 1
			lineW = 0
			lastBreak = -1
			continue
		}
		if ch.text == " " || ch.text == "\t" {
			lastBreak = ci
		}
		if wrapEnabled && lineW+cw > maxWidth && ci > lineStart {
			if cfg.Block.Wrap == WrapWord || cfg.Block.Wrap == WrapWordChar {
				if lastBreak > lineStart {
					lines = append(lines, lineInfo{startChar: lineStart, endChar: lastBreak + 1, width: lineW})
					lineStart = lastBreak + 1
					lineW = 0
					lastBreak = -1
					// Re-measure from new line start.
					for ri := lineStart; ri <= ci; ri++ {
						lineW += chars[ri].width + letterSpacing
					}
					continue
				}
			}
			// WrapChar or fallback.
			lines = append(lines, lineInfo{startChar: lineStart, endChar: ci, width: lineW})
			lineStart = ci
			lineW = cw
			lastBreak = -1
			continue
		}
		lineW += cw
	}
	if lineStart <= len(chars) {
		lines = append(lines, lineInfo{startChar: lineStart, endChar: len(chars), width: lineW})
	}

	// Build layout items, glyphs, char rects, etc.
	var allGlyphs []Glyph
	var charRects []CharRect
	charRectByIndex := make(map[int]int)
	var logAttrs []LogAttr
	logAttrByIndex := make(map[int]int)
	var items []Item
	var layoutLines []Line

	lineSpacing := float64(cfg.Block.LineSpacing) * float64(scale)
	indentPx := float64(cfg.Block.Indent) * float64(scale)
	penY := 0.0

	for li, line := range lines {
		lineY := penY

		// Compute alignment offset.
		alignOffset := 0.0
		if wrapEnabled && maxWidth > 0 {
			switch cfg.Block.Align {
			case AlignCenter:
				alignOffset = (maxWidth - line.width) / 2
			case AlignRight:
				alignOffset = maxWidth - line.width
			}
		}

		// Apply indent to first line.
		indent := 0.0
		if li == 0 {
			indent = indentPx
		}

		// Compute line byte range.
		startByteIdx := 0
		if line.startChar < len(chars) {
			startByteIdx = chars[line.startChar].byteI
		} else if len(chars) > 0 {
			last := chars[len(chars)-1]
			startByteIdx = last.byteI + last.byteL
		}
		endByteIdx := startByteIdx
		lineLen := 0
		if line.endChar > line.startChar && line.endChar <= len(chars) {
			lastCh := chars[line.endChar-1]
			endByteIdx = lastCh.byteI + lastCh.byteL
			lineLen = endByteIdx - startByteIdx
		}

		cx := alignOffset + indent
		itemStart := len(allGlyphs)
		itemStartByte := startByteIdx
		var curStyle *TextStyle
		var curSubSup winSubSupKind
		itemX := cx
		itemIsEmoji := false

		flushItem := func(endByte int) {
			gc := len(allGlyphs) - itemStart
			if gc <= 0 {
				return
			}
			var w float64
			for _, gl := range allGlyphs[itemStart : itemStart+gc] {
				w += gl.XAdvance
			}
			c := baseColor
			hasBg := cfg.Style.BgColor.A > 0
			bg := cfg.Style.BgColor
			hasUL := cfg.Style.Underline
			hasST := cfg.Style.Strikethrough
			sw := cfg.Style.StrokeWidth
			sc := cfg.Style.StrokeColor
			itemStyle := cfg.Style
			if curStyle != nil {
				itemStyle = *curStyle
				c = curStyle.Color
				if c.A == 0 {
					c = baseColor
				}
				hasBg = curStyle.BgColor.A > 0
				bg = curStyle.BgColor
				hasUL = curStyle.Underline
				hasST = curStyle.Strikethrough
				sw = curStyle.StrokeWidth
				sc = curStyle.StrokeColor
			}
			yOff := 0.0
			switch curSubSup {
			case winSubSupSub:
				yOff = fontAscent * 0.3
			case winSubSupSup:
				yOff = -fontAscent * 0.35
			}
			items = append(items, Item{
				Style:                  itemStyle,
				Width:                  w,
				X:                      itemX * pixelScale,
				Y:                      (lineY + fontAscent + yOff) * pixelScale,
				Ascent:                 fontAscent * pixelScale,
				Descent:                fontDescent * pixelScale,
				GlyphStart:             itemStart,
				GlyphCount:             gc,
				StartIndex:             itemStartByte,
				Length:                 endByte - itemStartByte,
				Color:                  c,
				UseOriginalColor:       itemIsEmoji,
				UnderlineOffset:        2.0,
				UnderlineThickness:     1.0,
				StrikethroughOffset:    fontAscent * 0.35 * pixelScale,
				StrikethroughThickness: 1.0,
				HasUnderline:           hasUL,
				HasStrikethrough:       hasST,
				HasBgColor:             hasBg,
				BgColor:                bg,
				StrokeWidth:            sw,
				StrokeColor:            sc,
				HasStroke:              sw > 0,
			})
			itemStart = len(allGlyphs)
		}

		for ci := line.startChar; ci < line.endChar; ci++ {
			ch := chars[ci]
			if ch.text == "\n" {
				continue
			}

			// Split item at emoji/non-emoji boundary.
			cp := []rune(ch.text)[0]
			charIsEmoji := isEmojiRune(cp)
			if charIsEmoji != itemIsEmoji && len(allGlyphs) > itemStart {
				flushItem(ch.byteI)
				itemStartByte = ch.byteI
				itemX = cx
				itemIsEmoji = charIsEmoji
			}
			if len(allGlyphs) == itemStart {
				itemIsEmoji = charIsEmoji
			}

			// Split item at style boundary for rich text.
			cw := ch.width + letterSpacing
			if overrides != nil {
				if s, ok := overrides[ch.byteI]; ok {
					adjusted := winApplySubSup(s)
					if curStyle == nil || adjusted != *curStyle {
						flushItem(ch.byteI)
						itemStartByte = ch.byteI
						itemX = cx
						sCopy := adjusted
						curStyle = &sCopy
						curSubSup = winDetectSubSup(s)
						// Re-select font and re-measure with new font.
						ctx.selectFont(adjusted)
					}
					// Always measure with the override font.
					w, _ := ctx.gdi.measureString(ch.text)
					cw = float64(w) + letterSpacing
				}
			}

			allGlyphs = append(allGlyphs, Glyph{
				Index:     uint32(ch.byteI),
				Codepoint: uint32(ch.byteL),
				XAdvance:  cw * pixelScale,
			})

			crIdx := len(charRects)
			charRects = append(charRects, CharRect{
				Rect: Rect{
					X:      float32(cx * pixelScale),
					Y:      float32(lineY * pixelScale),
					Width:  float32(cw * pixelScale),
					Height: float32(lineHeight * pixelScale),
				},
				Index: ch.byteI,
			})
			charRectByIndex[ch.byteI] = crIdx

			attrIdx := len(logAttrs)
			isWS := ch.text == " " || ch.text == "\t"
			prevWS := ci > 0 && (chars[ci-1].text == " " || chars[ci-1].text == "\t")
			logAttrs = append(logAttrs, LogAttr{
				IsCursorPosition: true,
				IsWordStart:      ci == 0 || (prevWS && !isWS),
				IsWordEnd:        isWS && ci > 0 && !prevWS,
			})
			logAttrByIndex[ch.byteI] = attrIdx

			cx += cw
		}

		flushItem(endByteIdx)

		// Re-select base font if overrides changed it.
		if overrides != nil {
			ctx.selectFont(cfg.Style)
			curStyle = nil
			curSubSup = winSubSupNone
		}

		layoutLines = append(layoutLines, Line{
			StartIndex:       startByteIdx,
			Length:           lineLen,
			IsParagraphStart: li == 0,
			Rect: Rect{
				X:      float32(alignOffset * pixelScale),
				Y:      float32(lineY * pixelScale),
				Width:  float32(line.width * pixelScale),
				Height: float32(lineHeight * pixelScale),
			},
		})

		penY += lineHeight + lineSpacing
	}

	// End-of-text log attr.
	endAttrIdx := len(logAttrs)
	logAttrs = append(logAttrs, LogAttr{IsCursorPosition: true})
	logAttrByIndex[len(text)] = endAttrIdx

	totalWidth := 0.0
	for _, line := range lines {
		totalWidth = max(totalWidth, line.width)
	}

	result := Layout{
		Text:            text,
		Items:           items,
		Glyphs:          allGlyphs,
		CharRects:       charRects,
		CharRectByIndex: charRectByIndex,
		Lines:           layoutLines,
		LogAttrs:        logAttrs,
		LogAttrByIndex:  logAttrByIndex,
		Width:           float32(totalWidth * pixelScale),
		Height:          float32(penY * pixelScale),
		VisualWidth:     float32(totalWidth * pixelScale),
		VisualHeight:    float32(max(penY, 0) * pixelScale),
	}
	result.buildPositionCaches()
	return result
}

// buildVerticalLayout produces a vertical (top-to-bottom) layout.
// Each character occupies one row; XAdvance=0, YAdvance=-lineHeight.
func (ctx *Context) buildVerticalLayout(text string, cfg TextConfig,
	overrides map[int]TextStyle) Layout {

	scale := ctx.scaleFactor
	pixelScale := 1.0 / float64(scale)

	tm := ctx.gdi.getTextMetrics()
	fontAscent := float64(tm.TmAscent)
	fontDescent := float64(tm.TmDescent)
	lineHeight := fontAscent + fontDescent + float64(tm.TmExternalLeading)

	baseColor := cfg.Style.Color
	if baseColor.A == 0 {
		baseColor = Color{0, 0, 0, 255}
	}

	var allGlyphs []Glyph
	var charRects []CharRect
	charRectByIndex := make(map[int]int)
	var logAttrs []LogAttr
	logAttrByIndex := make(map[int]int)

	penY := fontAscent // start at first baseline
	clusters := segmentGraphemes(text)

	for _, cl := range clusters {
		if cl.text == "\n" || cl.text == "\r" {
			continue
		}

		// Measure char width for centering.
		charW, _ := ctx.gdi.measureString(cl.text)
		centerX := (lineHeight - float64(charW)) / 2.0

		allGlyphs = append(allGlyphs, Glyph{
			Index:     uint32(cl.byteI),
			Codepoint: uint32(cl.byteL),
			XOffset:   centerX * pixelScale,
			XAdvance:  0,
			YAdvance:  -lineHeight * pixelScale,
		})

		crIdx := len(charRects)
		charRects = append(charRects, CharRect{
			Rect: Rect{
				X:      0,
				Y:      float32((penY - fontAscent) * pixelScale),
				Width:  float32(lineHeight * pixelScale),
				Height: float32(lineHeight * pixelScale),
			},
			Index: cl.byteI,
		})
		charRectByIndex[cl.byteI] = crIdx

		attrIdx := len(logAttrs)
		logAttrs = append(logAttrs, LogAttr{IsCursorPosition: true})
		logAttrByIndex[cl.byteI] = attrIdx

		penY += lineHeight
	}

	// End-of-text attr.
	endIdx := len(logAttrs)
	logAttrs = append(logAttrs, LogAttr{IsCursorPosition: true})
	logAttrByIndex[len(text)] = endIdx

	glyphCount := len(allGlyphs)
	totalH := penY

	var items []Item
	if glyphCount > 0 {
		items = append(items, Item{
			Style:      cfg.Style,
			Width:      lineHeight * pixelScale,
			X:          fontAscent * pixelScale,
			Y:          fontAscent * pixelScale,
			Ascent:     fontAscent * pixelScale,
			Descent:    fontDescent * pixelScale,
			GlyphStart: 0,
			GlyphCount: glyphCount,
			StartIndex: 0,
			Length:     len(text),
			Color:      baseColor,
		})
	}

	lines := []Line{{
		StartIndex: 0,
		Length:     len(text),
		Rect: Rect{
			X: 0, Y: 0,
			Width:  float32(lineHeight * pixelScale),
			Height: float32(totalH * pixelScale),
		},
		IsParagraphStart: true,
	}}

	result := Layout{
		Text:            text,
		Items:           items,
		Glyphs:          allGlyphs,
		CharRects:       charRects,
		CharRectByIndex: charRectByIndex,
		Lines:           lines,
		LogAttrs:        logAttrs,
		LogAttrByIndex:  logAttrByIndex,
		Width:           float32(lineHeight * pixelScale),
		Height:          float32(totalH * pixelScale),
		VisualWidth:     float32(lineHeight * pixelScale),
		VisualHeight:    float32(totalH * pixelScale),
	}
	result.buildPositionCaches()
	return result
}

// winSubSupKind identifies subscript/superscript simulation mode.
type winSubSupKind int

const (
	winSubSupNone winSubSupKind = iota
	winSubSupSub
	winSubSupSup
)

// winDetectSubSup checks if a TextStyle uses subs or sups OpenType features.
func winDetectSubSup(s TextStyle) winSubSupKind {
	if s.Features == nil {
		return winSubSupNone
	}
	for _, f := range s.Features.OpenTypeFeatures {
		if f.Value == 0 {
			continue
		}
		switch f.Tag {
		case "subs":
			return winSubSupSub
		case "sups":
			return winSubSupSup
		}
	}
	return winSubSupNone
}

// winApplySubSup returns a modified style that simulates subs/sups by
// scaling the font size to ~65% of original.
func winApplySubSup(s TextStyle) TextStyle {
	if winDetectSubSup(s) == winSubSupNone {
		return s
	}
	s.FontName = winScaleFontSize(s.FontName, 0.65)
	return s
}

// winScaleFontSize scales the numeric size component of a font name.
func winScaleFontSize(fontName string, factor float64) string {
	parts := strings.Fields(fontName)
	if len(parts) == 0 {
		return fontName
	}
	var sz float64
	if _, err := fmt.Sscanf(parts[len(parts)-1], "%f", &sz); err == nil && sz > 0 {
		parts[len(parts)-1] = fmt.Sprintf("%g", sz*factor)
		return strings.Join(parts, " ")
	}
	return fontName
}
