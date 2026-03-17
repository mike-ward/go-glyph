package glyph

// testLayout builds a simple Layout with known geometry for testing
// query functions. Two lines: "Hello" (indices 0..5) and "World"
// (indices 6..11). Each char 10px wide, 20px tall.
func testLayout() Layout {
	charRects := []CharRect{
		{Rect: Rect{X: 0, Y: 0, Width: 10, Height: 20}, Index: 0},   // H
		{Rect: Rect{X: 10, Y: 0, Width: 10, Height: 20}, Index: 1},  // e
		{Rect: Rect{X: 20, Y: 0, Width: 10, Height: 20}, Index: 2},  // l
		{Rect: Rect{X: 30, Y: 0, Width: 10, Height: 20}, Index: 3},  // l
		{Rect: Rect{X: 40, Y: 0, Width: 10, Height: 20}, Index: 4},  // o
		{Rect: Rect{X: 0, Y: 20, Width: 10, Height: 20}, Index: 6},  // W
		{Rect: Rect{X: 10, Y: 20, Width: 10, Height: 20}, Index: 7}, // o
		{Rect: Rect{X: 20, Y: 20, Width: 10, Height: 20}, Index: 8}, // r
		{Rect: Rect{X: 30, Y: 20, Width: 10, Height: 20}, Index: 9}, // l
		{Rect: Rect{X: 40, Y: 20, Width: 10, Height: 20}, Index: 10}, // d
	}
	charRectByIndex := map[int]int{
		0: 0, 1: 1, 2: 2, 3: 3, 4: 4,
		6: 5, 7: 6, 8: 7, 9: 8, 10: 9,
	}
	lines := []Line{
		{StartIndex: 0, Length: 5, Rect: Rect{X: 0, Y: 0, Width: 50, Height: 20}},
		{StartIndex: 6, Length: 5, Rect: Rect{X: 0, Y: 20, Width: 50, Height: 20}},
	}
	logAttrs := []LogAttr{
		{IsCursorPosition: true, IsWordStart: true},  // 0: H
		{IsCursorPosition: true},                      // 1: e
		{IsCursorPosition: true},                      // 2: l
		{IsCursorPosition: true},                      // 3: l
		{IsCursorPosition: true},                      // 4: o
		{IsCursorPosition: true, IsWordEnd: true},     // 5: \n
		{IsCursorPosition: true, IsWordStart: true},   // 6: W
		{IsCursorPosition: true},                      // 7: o
		{IsCursorPosition: true},                      // 8: r
		{IsCursorPosition: true},                      // 9: l
		{IsCursorPosition: true},                      // 10: d
		{IsCursorPosition: true, IsWordEnd: true},     // 11: end
	}
	logAttrByIndex := map[int]int{
		0: 0, 1: 1, 2: 2, 3: 3, 4: 4, 5: 5,
		6: 6, 7: 7, 8: 8, 9: 9, 10: 10, 11: 11,
	}
	return Layout{
		Text:            "Hello\nWorld",
		CharRects:       charRects,
		CharRectByIndex: charRectByIndex,
		Lines:           lines,
		LogAttrs:        logAttrs,
		LogAttrByIndex:  logAttrByIndex,
		Width:           50,
		Height:          40,
	}
}
