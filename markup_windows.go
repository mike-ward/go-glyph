//go:build windows

package glyph

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// parsePangoMarkup parses a Pango markup string into StyleRuns.
func parsePangoMarkup(text string, baseStyle TextStyle) ([]StyleRun, error) {
	wrapped := "<root>" + text + "</root>"
	decoder := xml.NewDecoder(strings.NewReader(wrapped))

	var runs []StyleRun
	styleStack := []TextStyle{baseStyle}

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return runs, fmt.Errorf("glyph: markup parse error: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			cur := styleStack[len(styleStack)-1]
			switch t.Name.Local {
			case "b":
				cur = markupApplyBold(cur)
			case "i":
				cur = markupApplyItalic(cur)
			case "u":
				cur.Underline = true
			case "s":
				cur.Strikethrough = true
			case "span":
				cur = markupApplySpan(cur, t.Attr)
			}
			styleStack = append(styleStack, cur)
		case xml.EndElement:
			if len(styleStack) > 1 {
				styleStack = styleStack[:len(styleStack)-1]
			}
		case xml.CharData:
			if s := string(t); len(s) > 0 {
				runs = append(runs, StyleRun{
					Text:  s,
					Style: styleStack[len(styleStack)-1],
				})
			}
		}
	}

	if len(runs) == 0 {
		return []StyleRun{{Text: text, Style: baseStyle}}, nil
	}
	return runs, nil
}

func markupApplyBold(s TextStyle) TextStyle {
	if s.Typeface == TypefaceItalic || s.Typeface == TypefaceBoldItalic {
		s.Typeface = TypefaceBoldItalic
	} else {
		s.Typeface = TypefaceBold
	}
	return s
}

func markupApplyItalic(s TextStyle) TextStyle {
	if s.Typeface == TypefaceBold || s.Typeface == TypefaceBoldItalic {
		s.Typeface = TypefaceBoldItalic
	} else {
		s.Typeface = TypefaceItalic
	}
	return s
}

func markupApplySpan(s TextStyle, attrs []xml.Attr) TextStyle {
	for _, attr := range attrs {
		switch attr.Name.Local {
		case "foreground", "fgcolor", "color":
			if c, ok := markupParseHexColor(attr.Value); ok {
				s.Color = c
			}
		case "font_family", "face":
			s.FontName = markupReplaceFontFamily(s.FontName, attr.Value)
		case "size":
			if sz, ok := markupNamedSize(attr.Value); ok {
				s.FontName = markupReplaceFontSize(s.FontName, sz)
			} else if v, err := strconv.Atoi(attr.Value); err == nil && v > 0 {
				s.FontName = markupReplaceFontSize(s.FontName, float32(v)/1024.0)
			}
		case "weight":
			if attr.Value == "bold" {
				s = markupApplyBold(s)
			}
		case "style":
			if attr.Value == "italic" || attr.Value == "oblique" {
				s = markupApplyItalic(s)
			}
		case "underline":
			s.Underline = attr.Value == "single" || attr.Value == "true"
		case "strikethrough":
			s.Strikethrough = attr.Value == "true"
		}
	}
	return s
}

func markupParseHexColor(s string) (Color, bool) {
	s = strings.TrimPrefix(s, "#")
	switch len(s) {
	case 3: // #RGB → #RRGGBB
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	case 4: // #RGBA → #RRGGBBAA
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2], s[3], s[3]})
	case 6, 8:
		// Already correct length.
	default:
		return Color{}, false
	}
	r, err1 := strconv.ParseUint(s[0:2], 16, 8)
	g, err2 := strconv.ParseUint(s[2:4], 16, 8)
	b, err3 := strconv.ParseUint(s[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return Color{}, false
	}
	a := uint64(255)
	if len(s) == 8 {
		var err error
		a, err = strconv.ParseUint(s[6:8], 16, 8)
		if err != nil {
			return Color{}, false
		}
	}
	return Color{R: byte(r), G: byte(g), B: byte(b), A: byte(a)}, true
}

func markupNamedSize(name string) (float32, bool) {
	switch strings.ToLower(name) {
	case "xx-small":
		return 6.9, true
	case "x-small":
		return 8.3, true
	case "small":
		return 10, true
	case "medium":
		return 12, true
	case "large":
		return 14.4, true
	case "x-large":
		return 17.3, true
	case "xx-large":
		return 20.7, true
	}
	return 0, false
}

func markupReplaceFontFamily(fontName, newFamily string) string {
	parts := strings.Fields(fontName)
	if len(parts) > 0 {
		if _, err := fmt.Sscanf(parts[len(parts)-1], "%f", new(float32)); err == nil {
			return newFamily + " " + parts[len(parts)-1]
		}
	}
	return newFamily
}

func markupReplaceFontSize(fontName string, newSize float32) string {
	parts := strings.Fields(fontName)
	end := len(parts)
	if end > 0 {
		if _, err := fmt.Sscanf(parts[end-1], "%f", new(float32)); err == nil {
			end--
		}
	}
	family := strings.Join(parts[:end], " ")
	if family == "" {
		family = "Sans"
	}
	return fmt.Sprintf("%s %g", family, newSize)
}
