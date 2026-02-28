package glyph

// DrawComposition renders IME preedit visual feedback: clause
// underlines and preedit cursor. Call after DrawLayout when
// composition is active.
func (r *Renderer) DrawComposition(layout Layout, x, y float32,
	cs *CompositionState, cursorColor Color) {

	if !cs.IsComposing() {
		return
	}

	// Draw clause underlines.
	clauseRects := cs.GetClauseRects(layout)
	for _, cr := range clauseRects {
		thickness := float32(1.0)
		if cr.Style == ClauseSelected {
			thickness = 2.0
		}
		// ~70% opacity.
		ulColor := Color{
			R: cursorColor.R,
			G: cursorColor.G,
			B: cursorColor.B,
			A: 178,
		}
		for _, rect := range cr.Rects {
			ulY := rect.Y + rect.Height - thickness
			r.backend.DrawFilledRect(Rect{
				X:      rect.X + x,
				Y:      ulY + y,
				Width:  rect.Width,
				Height: thickness,
			}, ulColor)
		}
	}

	// Draw cursor at insertion point within preedit.
	cursorPos := cs.DocumentCursorPos()
	if cp, ok := layout.GetCursorPos(cursorPos); ok {
		dimmed := Color{
			R: cursorColor.R,
			G: cursorColor.G,
			B: cursorColor.B,
			A: 178,
		}
		r.backend.DrawFilledRect(Rect{
			X:      cp.X + x,
			Y:      cp.Y + y,
			Width:  2.0,
			Height: cp.Height,
		}, dimmed)
	}
}

// DrawLayoutWithComposition renders a layout with preedit text.
// Currently draws normally; opacity reduction deferred to future.
func (r *Renderer) DrawLayoutWithComposition(layout Layout, x, y float32,
	cs *CompositionState) {

	r.DrawLayout(layout, x, y)
}
