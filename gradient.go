package glyph

// GradientStop defines a color at a normalized position (0.0–1.0).
type GradientStop struct {
	Color    Color
	Position float32
}

// GradientConfig defines an N-stop gradient for text rendering.
// Stops must be sorted by position in ascending order.
type GradientConfig struct {
	Stops     []GradientStop
	Direction GradientDirection
}

// GradientColorAt samples the gradient at normalized position t.
func GradientColorAt(stops []GradientStop, t float32) Color {
	if len(stops) == 0 {
		return Color{0, 0, 0, 255}
	}
	if len(stops) == 1 || t <= stops[0].Position {
		return stops[0].Color
	}
	last := stops[len(stops)-1]
	if t >= last.Position {
		return last.Color
	}
	for i := 0; i < len(stops)-1; i++ {
		if t >= stops[i].Position && t <= stops[i+1].Position {
			span := stops[i+1].Position - stops[i].Position
			if span <= 0 {
				return stops[i].Color
			}
			localT := (t - stops[i].Position) / span
			return LerpColor(stops[i].Color, stops[i+1].Color, localT)
		}
	}
	return last.Color
}
