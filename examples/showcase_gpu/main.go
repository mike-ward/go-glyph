// Command showcase_gpu is a comprehensive feature gallery for the
// glyph library using an SDL2 window with raw Metal backend.
// Scroll with mouse wheel or Page Up/Down, Home/End keys.
package main

import (
	"fmt"
	"math"
	"runtime"
	"unsafe"

	"glyph"
	"glyph/backend/gpu"

	"github.com/veandco/go-sdl2/sdl"
)

func init() { runtime.LockOSThread() }

const (
	screenW    = 1000
	screenH    = 800
	margin     = 40
	sectionGap = 30
)

// Dark theme palette.
var (
	bgColor   = gc(20, 20, 25, 255)
	textColor = gc(220, 220, 225, 255)
	dimColor  = gc(140, 140, 150, 255)
	accent    = gc(100, 160, 255, 255)
	warm      = gc(255, 140, 80, 255)
	cool      = gc(80, 180, 255, 255)
	divider   = gc(50, 50, 60, 255)
	highlight = gc(255, 220, 80, 255)
	codeGreen = gc(160, 220, 140, 255)
)

func gc(r, g, b, a uint8) glyph.Color {
	return glyph.Color{R: r, G: g, B: b, A: a}
}

type section struct {
	title  string
	height float32
	draw   func(a *app, x, y, w float32)
}

type app struct {
	window  *sdl.Window
	backend *gpu.Backend
	ts      *glyph.TextSystem
	sects   []section
	scrollY float32
	frame   int
}

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	win, err := sdl.CreateWindow("go_glyph showcase (Metal)",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		screenW, screenH,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE|
			sdl.WINDOW_ALLOW_HIGHDPI|gpu.WindowFlag())
	if err != nil {
		panic(err)
	}
	defer win.Destroy()

	physW, _ := gpu.WindowDrawableSize(unsafe.Pointer(win))
	winW, _ := win.GetSize()
	dpi := float32(1)
	if winW > 0 {
		dpi = float32(physW) / float32(winW)
	}

	be, err := gpu.New(unsafe.Pointer(win), dpi)
	if err != nil {
		panic(err)
	}
	defer be.Destroy()

	ts, err := glyph.NewTextSystem(be)
	if err != nil {
		panic(err)
	}
	defer ts.Free()

	a := &app{window: win, backend: be, ts: ts}
	a.buildSections()

	sdl.AddEventWatchFunc(func(ev sdl.Event, _ interface{}) bool {
		if we, ok := ev.(*sdl.WindowEvent); ok {
			if we.Event == sdl.WINDOWEVENT_EXPOSED ||
				we.Event == sdl.WINDOWEVENT_SIZE_CHANGED {
				a.render()
			}
		}
		return true
	}, nil)

	for {
		for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
			switch e := ev.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.MouseWheelEvent:
				a.scrollY -= float32(e.Y) * 40
				a.clampScroll()
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					a.handleKey(e.Keysym.Sym)
				}
			}
		}
		a.render()
	}
}

func (a *app) handleKey(sym sdl.Keycode) {
	_, wh := a.window.GetSize()
	switch sym {
	case sdl.K_HOME:
		a.scrollY = 0
	case sdl.K_END:
		a.scrollY = a.totalHeight() - float32(wh)
	case sdl.K_PAGEUP:
		a.scrollY -= float32(wh) * 0.8
	case sdl.K_PAGEDOWN:
		a.scrollY += float32(wh) * 0.8
	case sdl.K_UP:
		a.scrollY -= 40
	case sdl.K_DOWN:
		a.scrollY += 40
	}
	a.clampScroll()
}

func (a *app) totalHeight() float32 {
	h := float32(20)
	for _, s := range a.sects {
		h += s.height + sectionGap
	}
	return h
}

func (a *app) clampScroll() {
	_, wh := a.window.GetSize()
	max := a.totalHeight() - float32(wh)
	if max < 0 {
		max = 0
	}
	if a.scrollY > max {
		a.scrollY = max
	}
	if a.scrollY < 0 {
		a.scrollY = 0
	}
}

func (a *app) render() {
	a.backend.BeginFrame()
	a.drawSections()
	a.ts.Commit()
	w, h := a.window.GetSize()
	a.backend.EndFrame(
		float32(bgColor.R)/255, float32(bgColor.G)/255,
		float32(bgColor.B)/255, 1.0,
		int(w), int(h))
	a.frame++
}

func (a *app) drawSections() {
	ww, wh := a.window.GetSize()
	cw := float32(ww) - margin*2
	y := float32(20) - a.scrollY

	for i := range a.sects {
		s := &a.sects[i]

		// Cull above viewport.
		if y+s.height < 0 {
			y += s.height + sectionGap
			continue
		}
		// Stop below viewport.
		if y > float32(wh) {
			break
		}

		// Divider line between sections.
		if i > 0 {
			a.backend.DrawFilledRect(glyph.Rect{
				X: margin, Y: y - sectionGap/2,
				Width: cw, Height: 1,
			}, divider)
		}

		// Section header.
		_ = a.ts.DrawText(margin, y, s.title, glyph.TextConfig{
			Style: glyph.TextStyle{
				FontName:      "Sans 11",
				Typeface:      glyph.TypefaceBold,
				Color:         accent,
				LetterSpacing: 2,
			},
		})

		// Section content.
		s.draw(a, margin, y+30, cw)

		y += s.height + sectionGap
	}
}

// ----- sections -----

func (a *app) buildSections() {
	a.sects = []section{
		{"ℹ️ INTRO", 100, drawIntro},
		{"ℹ️ TYPOGRAPHY", 200, drawTypography},
		{"ℹ️ DECORATIONS", 110, drawDecorations},
		{"ℹ️ TEXT STROKE", 150, drawStroke},
		{"LAYOUT", 220, drawLayout},
		{"RICH TEXT", 60, drawRichText},
		{"PANGO MARKUP", 60, drawMarkup},
		{"GRADIENTS", 160, drawGradients},
		{"INTERNATIONALIZATION", 250, drawI18n},
		{"OPENTYPE FEATURES", 210, drawOpenType},
		{"LETTER SPACING", 140, drawSpacing},
		{"FONT SIZES", 220, drawSizes},
		{"ROTATED TEXT", 180, drawRotated},
		{"VERTICAL TEXT", 260, drawVertical},
		{"TEXT ON PATH", 250, drawPathText},
		{"SKEWED TEXT", 140, drawSkewed},
	}
}

// --- Section 1: Intro ---

func drawIntro(a *app, x, y, w float32) {
	grad := &glyph.GradientConfig{
		Direction: glyph.GradientHorizontal,
		Stops: []glyph.GradientStop{
			{Color: gc(100, 160, 255, 255), Position: 0.0},
			{Color: gc(200, 120, 255, 255), Position: 0.5},
			{Color: gc(255, 100, 160, 255), Position: 1.0},
		},
	}
	_ = a.ts.DrawText(x, y, "go_glyph", glyph.TextConfig{
		Style:    glyph.TextStyle{FontName: "Sans Bold 48", Color: textColor},
		Gradient: grad,
	})
	_ = a.ts.DrawText(x, y+58, "GPU-accelerated text rendering for Go", glyph.TextConfig{
		Style: glyph.TextStyle{FontName: "Sans 16", Color: dimColor},
	})
}

// --- Section 2: Typography ---

func drawTypography(a *app, x, y, w float32) {
	dy := float32(0)
	families := [][2]string{
		{"Sans (default)", "Sans 18"},
		{"Serif", "Serif 18"},
		{"Monospace", "Monospace 18"},
	}
	for _, f := range families {
		_ = a.ts.DrawText(x, y+dy, f[0], glyph.TextConfig{
			Style: glyph.TextStyle{FontName: f[1], Color: textColor},
		})
		dy += 28
	}
	dy += 8
	faces := []struct {
		label string
		tf    glyph.Typeface
	}{
		{"Bold", glyph.TypefaceBold},
		{"Italic", glyph.TypefaceItalic},
		{"Bold Italic", glyph.TypefaceBoldItalic},
	}
	for _, f := range faces {
		_ = a.ts.DrawText(x, y+dy, f.label, glyph.TextConfig{
			Style: glyph.TextStyle{
				FontName: "Sans 18", Color: textColor, Typeface: f.tf,
			},
		})
		dy += 28
	}
}

// --- Section 3: Decorations ---

func drawDecorations(a *app, x, y, w float32) {
	_ = a.ts.DrawText(x, y, "Underlined text", glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName: "Sans 18", Color: textColor, Underline: true,
		},
	})
	_ = a.ts.DrawText(x, y+30, "Strikethrough text", glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName: "Sans 18", Color: textColor, Strikethrough: true,
		},
	})
	_ = a.ts.DrawText(x, y+60, "Highlighted text", glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName: "Sans 18",
			Color:    gc(30, 30, 30, 255),
			BgColor:  highlight,
		},
	})
}

// --- Section 4: Text Stroke ---

func drawStroke(a *app, x, y, w float32) {
	// Hollow: transparent fill, visible stroke.
	_ = a.ts.DrawText(x, y, "Hollow Text", glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName:    "Sans Bold 32",
			Color:       gc(0, 0, 0, 0),
			StrokeWidth: 2.0,
			StrokeColor: textColor,
		},
	})
	// Outlined: colored fill + white stroke.
	_ = a.ts.DrawText(x, y+45, "Outlined Text", glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName:    "Sans Bold 32",
			Color:       accent,
			StrokeWidth: 1.5,
			StrokeColor: gc(255, 255, 255, 255),
		},
	})
	// Neon stroke.
	_ = a.ts.DrawText(x, y+90, "Neon Stroke", glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName:    "Sans Bold 32",
			Color:       gc(0, 0, 0, 0),
			StrokeWidth: 2.5,
			StrokeColor: gc(0, 255, 180, 255),
		},
	})
}

// --- Section 5: Layout ---

func drawLayout(a *app, x, y, w float32) {
	wrapW := float32(400)
	if w < 450 {
		wrapW = w - 20
	}

	// Background rect to show wrap boundary.
	a.backend.DrawFilledRect(glyph.Rect{
		X: x, Y: y, Width: wrapW, Height: 72,
	}, gc(30, 30, 38, 255))
	_ = a.ts.DrawText(x+4, y+4,
		"This paragraph demonstrates word wrapping within a "+
			"constrained width. The layout engine automatically "+
			"breaks lines at word boundaries.",
		glyph.TextConfig{
			Style: glyph.TextStyle{FontName: "Sans 14", Color: textColor},
			Block: glyph.BlockStyle{Wrap: glyph.WrapWord, Width: wrapW - 8},
		})

	// Alignment.
	dy := float32(88)
	aligns := []struct {
		label string
		a     glyph.Alignment
	}{
		{"Left aligned", glyph.AlignLeft},
		{"Center aligned", glyph.AlignCenter},
		{"Right aligned", glyph.AlignRight},
	}
	for _, al := range aligns {
		_ = a.ts.DrawText(x, y+dy, al.label, glyph.TextConfig{
			Style: glyph.TextStyle{FontName: "Sans 14", Color: textColor},
			Block: glyph.BlockStyle{Width: wrapW, Align: al.a},
		})
		dy += 22
	}

	// Hanging indent.
	dy += 10
	_ = a.ts.DrawText(x+20, y+dy,
		"1. This is a numbered item with a hanging indent. "+
			"Continuation lines align to the indent, not the number.",
		glyph.TextConfig{
			Style: glyph.TextStyle{FontName: "Sans 14", Color: textColor},
			Block: glyph.BlockStyle{
				Wrap: glyph.WrapWord, Width: wrapW - 20,
				Indent: -20,
			},
		})
}

// --- Section 6: Rich Text ---

func drawRichText(a *app, x, y, w float32) {
	rt := glyph.RichText{
		Runs: []glyph.StyleRun{
			{Text: "Rich text: ", Style: glyph.TextStyle{
				FontName: "Sans 18", Color: textColor,
			}},
			{Text: "bold", Style: glyph.TextStyle{
				FontName: "Sans 18", Color: warm, Typeface: glyph.TypefaceBold,
			}},
			{Text: ", ", Style: glyph.TextStyle{
				FontName: "Sans 18", Color: textColor,
			}},
			{Text: "italic", Style: glyph.TextStyle{
				FontName: "Sans 18", Color: cool, Typeface: glyph.TypefaceItalic,
			}},
			{Text: ", ", Style: glyph.TextStyle{
				FontName: "Sans 18", Color: textColor,
			}},
			{Text: "underlined", Style: glyph.TextStyle{
				FontName: "Sans 18", Color: gc(140, 255, 140, 255),
				Underline: true,
			}},
			{Text: ", and ", Style: glyph.TextStyle{
				FontName: "Sans 18", Color: textColor,
			}},
			{Text: "monospace", Style: glyph.TextStyle{
				FontName: "Monospace 16", Color: codeGreen,
				BgColor: gc(40, 40, 50, 255),
			}},
		},
	}
	layout, err := a.ts.LayoutRichText(rt, glyph.TextConfig{})
	if err == nil {
		a.ts.DrawLayout(layout, x, y)
	}
}

// --- Section 7: Pango Markup ---

func drawMarkup(a *app, x, y, w float32) {
	markup := `<b>Bold</b>, <i>italic</i>, ` +
		`<span foreground="#ff9966">orange</span>, ` +
		`<span size="x-large">large</span>, ` +
		`<u>underline</u>, ` +
		`<span font_family="monospace">mono</span>`
	_ = a.ts.DrawText(x, y, markup, glyph.TextConfig{
		Style:     glyph.TextStyle{FontName: "Sans 16", Color: textColor},
		UseMarkup: true,
	})
}

// --- Section 8: Gradients ---

func drawGradients(a *app, x, y, w float32) {
	rainbow := &glyph.GradientConfig{
		Direction: glyph.GradientHorizontal,
		Stops: []glyph.GradientStop{
			{Color: gc(255, 0, 0, 255), Position: 0.0},
			{Color: gc(255, 165, 0, 255), Position: 0.2},
			{Color: gc(255, 255, 0, 255), Position: 0.4},
			{Color: gc(0, 200, 0, 255), Position: 0.6},
			{Color: gc(0, 100, 255, 255), Position: 0.8},
			{Color: gc(160, 0, 255, 255), Position: 1.0},
		},
	}
	_ = a.ts.DrawText(x, y, "Horizontal Rainbow Gradient", glyph.TextConfig{
		Style:    glyph.TextStyle{FontName: "Sans Bold 28", Color: textColor},
		Gradient: rainbow,
	})

	vertical := &glyph.GradientConfig{
		Direction: glyph.GradientVertical,
		Stops: []glyph.GradientStop{
			{Color: gc(255, 100, 50, 255), Position: 0.0},
			{Color: gc(50, 150, 255, 255), Position: 1.0},
		},
	}
	_ = a.ts.DrawText(x, y+45, "Vertical Warm to Cool", glyph.TextConfig{
		Style:    glyph.TextStyle{FontName: "Sans Bold 28", Color: textColor},
		Gradient: vertical,
	})

	_ = a.ts.DrawText(x, y+90, "Gradient + Stroke", glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName:    "Sans Bold 28",
			Color:       textColor,
			StrokeWidth: 1.5,
			StrokeColor: gc(255, 255, 255, 180),
		},
		Gradient: rainbow,
	})
}

// --- Section 9: Internationalization ---

func drawI18n(a *app, x, y, w float32) {
	lines := []struct {
		label, text string
		c           glyph.Color
	}{
		{"Chinese", "\u4f60\u597d\u4e16\u754c\uff01", accent},
		{"Japanese", "\u3053\u3093\u306b\u3061\u306f\u4e16\u754c", cool},
		{"Korean", "\uc548\ub155\ud558\uc138\uc694", warm},
		{"Arabic (RTL)", "\u0645\u0631\u062d\u0628\u0627 \u0628\u0627\u0644\u0639\u0627\u0644\u0645",
			gc(200, 160, 80, 255)},
		{"Hebrew (RTL)", "\u05e9\u05dc\u05d5\u05dd \u05e2\u05d5\u05dc\u05dd",
			gc(160, 200, 80, 255)},
		{"Cyrillic", "\u041f\u0440\u0438\u0432\u0435\u0442, \u043c\u0438\u0440!",
			gc(180, 140, 200, 255)},
		{"Greek", "\u0393\u03b5\u03b9\u03b1 \u03c3\u03bf\u03c5 \u039a\u03cc\u03c3\u03bc\u03b5!",
			gc(140, 200, 180, 255)},
		{"Emoji",
			"\U0001F680 \U0001F30D \U0001F525 \U0001F44D " +
				"\U0001F600 \u2764\ufe0f \U0001F308 \U0001F389",
			textColor},
	}
	dy := float32(0)
	for _, l := range lines {
		_ = a.ts.DrawText(x, y+dy, l.label, glyph.TextConfig{
			Style: glyph.TextStyle{FontName: "Sans 12", Color: dimColor},
		})
		_ = a.ts.DrawText(x+130, y+dy, l.text, glyph.TextConfig{
			Style: glyph.TextStyle{FontName: "Sans 20", Color: l.c},
		})
		dy += 30
	}
}

// --- Section 10: OpenType Features ---

func drawOpenType(a *app, x, y, w float32) {
	// Note: feature availability depends on installed fonts.
	feats := []struct {
		label, text string
		features    []glyph.FontFeature
	}{
		{"Ligatures (dlig)", "ff fi fl ffi ffl ct st",
			[]glyph.FontFeature{{Tag: "dlig", Value: 1}}},
		{"Small Caps (smcp)", "Small Capitals Text",
			[]glyph.FontFeature{{Tag: "smcp", Value: 1}}},
		{"Old-style Figures (onum)", "0123456789",
			[]glyph.FontFeature{{Tag: "onum", Value: 1}}},
		{"Tabular Figures (tnum)", "0123456789",
			[]glyph.FontFeature{{Tag: "tnum", Value: 1}}},
	}
	dy := float32(0)
	for _, f := range feats {
		_ = a.ts.DrawText(x, y+dy, f.label, glyph.TextConfig{
			Style: glyph.TextStyle{FontName: "Sans 12", Color: dimColor},
		})
		dy += 18
		_ = a.ts.DrawText(x, y+dy, f.text, glyph.TextConfig{
			Style: glyph.TextStyle{
				FontName: "Serif 20", Color: textColor,
				Features: &glyph.FontFeatures{
					OpenTypeFeatures: f.features,
				},
			},
		})
		dy += 34
	}
}

// --- Section 11: Letter Spacing ---

func drawSpacing(a *app, x, y, w float32) {
	spacings := []struct {
		label   string
		spacing float32
	}{
		{"Tight (-1.5pt)", -1.5},
		{"Normal (0pt)", 0},
		{"Wide (3pt)", 3.0},
		{"Extra wide (8pt)", 8.0},
	}
	dy := float32(0)
	for _, s := range spacings {
		_ = a.ts.DrawText(x, y+dy, s.label, glyph.TextConfig{
			Style: glyph.TextStyle{
				FontName:      "Sans 18",
				Color:         textColor,
				LetterSpacing: s.spacing,
			},
		})
		dy += 30
	}
}

// --- Section 12: Font Sizes ---

func drawSizes(a *app, x, y, w float32) {
	sizes := []int{10, 12, 14, 18, 24, 32, 48}
	dy := float32(0)
	for _, sz := range sizes {
		_ = a.ts.DrawText(x, y+dy,
			fmt.Sprintf("%dpt The quick brown fox", sz),
			glyph.TextConfig{
				Style: glyph.TextStyle{
					FontName: fmt.Sprintf("Sans %d", sz),
					Color:    textColor,
				},
			})
		dy += float32(sz) + 10
	}
}

// --- Section 13: Rotated Text ---

func drawRotated(a *app, x, y, w float32) {
	// Animated spinner.
	angle := float32(a.frame%360) * math.Pi / 180.0
	layout, err := a.ts.LayoutText("Spinning!", glyph.TextConfig{
		Style: glyph.TextStyle{FontName: "Sans Bold 24", Color: warm},
	})
	if err == nil {
		a.ts.DrawLayoutRotated(layout, x+80, y+60, angle)
	}

	// Static samples at fixed angles.
	rx := x + 250
	for _, deg := range []float32{15, 30, 45, -15} {
		rad := deg * math.Pi / 180.0
		l2, err := a.ts.LayoutText(
			fmt.Sprintf("%.0f\u00b0", deg),
			glyph.TextConfig{
				Style: glyph.TextStyle{FontName: "Sans 16", Color: accent},
			})
		if err == nil {
			a.ts.DrawLayoutRotated(l2, rx, y+60, rad)
		}
		rx += 80
	}
}

// --- Section 14: Vertical Text ---

func drawVertical(a *app, x, y, w float32) {
	texts := []struct {
		text string
		c    glyph.Color
	}{
		{"\u7e26\u66f8\u304d\u30c6\u30b9\u30c8", cool},
		{"\u5782\u76f4\u6587\u5b57\u6d4b\u8bd5", warm},
		{"\ud55c\uad6d\uc5b4\ud14c\uc2a4\ud2b8", gc(180, 140, 200, 255)},
	}
	dx := float32(0)
	for _, t := range texts {
		_ = a.ts.DrawText(x+dx, y, t.text, glyph.TextConfig{
			Style:       glyph.TextStyle{FontName: "Sans 22", Color: t.c},
			Orientation: glyph.OrientationVertical,
		})
		dx += 40
	}
}

// --- Section 15: Text on Path ---

func drawPathText(a *app, x, y, w float32) {
	text := "Text flowing along a circular path!"
	cfg := glyph.TextConfig{
		Style: glyph.TextStyle{FontName: "Sans Bold 18", Color: accent},
	}
	layout, err := a.ts.LayoutText(text, cfg)
	if err != nil {
		return
	}
	positions := layout.GlyphPositions()
	if len(positions) == 0 {
		return
	}

	cx := x + minf(w*0.35, 300)
	cy := y + 120
	radius := float32(150)

	var totalAdv float32
	for _, p := range positions {
		totalAdv += p.Advance
	}

	// Arc-length parameterization: arc span from text width,
	// not a fixed constant. Matches VGlyph algorithm.
	arcSpan := totalAdv / radius
	startAngle := -arcSpan / 2

	placements := make([]glyph.GlyphPlacement, len(layout.Glyphs))
	// Default: offscreen.
	for i := range placements {
		placements[i] = glyph.GlyphPlacement{X: -9999, Y: -9999}
	}

	cumAdv := float32(0)
	for _, p := range positions {
		mid := cumAdv + p.Advance*0.5
		theta := startAngle + mid/radius

		tangent := theta + math.Pi/2

		arcX := cx + radius*float32(math.Cos(float64(theta)))
		arcY := cy + radius*float32(math.Sin(float64(theta)))

		halfAdv := p.Advance * 0.5
		gx := arcX - halfAdv*float32(math.Cos(float64(tangent)))
		gy := arcY - halfAdv*float32(math.Sin(float64(tangent)))

		placements[p.Index] = glyph.GlyphPlacement{
			X: gx, Y: gy, Angle: tangent,
		}
		cumAdv += p.Advance
	}

	a.ts.DrawLayoutPlaced(layout, placements)
}

// --- Section 16: Skewed Text ---

func drawSkewed(a *app, x, y, w float32) {
	l1, err := a.ts.LayoutText("Skewed Text", glyph.TextConfig{
		Style: glyph.TextStyle{FontName: "Sans Bold 28", Color: warm},
	})
	if err == nil {
		a.ts.DrawLayoutTransformed(l1, x, y, glyph.AffineSkew(-0.3, 0))
	}

	l2, err := a.ts.LayoutText("Reverse Skew", glyph.TextConfig{
		Style: glyph.TextStyle{FontName: "Sans Bold 28", Color: cool},
	})
	if err == nil {
		a.ts.DrawLayoutTransformed(l2, x, y+50, glyph.AffineSkew(0.3, 0))
	}

	l3, err := a.ts.LayoutText("Skew + Gradient", glyph.TextConfig{
		Style: glyph.TextStyle{FontName: "Sans Bold 28", Color: textColor},
	})
	if err == nil {
		grad := &glyph.GradientConfig{
			Direction: glyph.GradientHorizontal,
			Stops: []glyph.GradientStop{
				{Color: gc(255, 100, 100, 255), Position: 0},
				{Color: gc(100, 100, 255, 255), Position: 1},
			},
		}
		a.ts.Renderer().DrawLayoutTransformedWithGradient(
			l3, x, y+100, glyph.AffineSkew(-0.2, 0), grad)
	}
}

func minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
