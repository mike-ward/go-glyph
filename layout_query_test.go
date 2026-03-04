package glyph

import "testing"

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
		{IsCursorPosition: true, IsWordStart: true},   // 0: H
		{IsCursorPosition: true},                       // 1: e
		{IsCursorPosition: true},                       // 2: l
		{IsCursorPosition: true},                       // 3: l
		{IsCursorPosition: true},                       // 4: o
		{IsCursorPosition: true, IsWordEnd: true},      // 5: \n
		{IsCursorPosition: true, IsWordStart: true},    // 6: W
		{IsCursorPosition: true},                       // 7: o
		{IsCursorPosition: true},                       // 8: r
		{IsCursorPosition: true},                       // 9: l
		{IsCursorPosition: true},                       // 10: d
		{IsCursorPosition: true, IsWordEnd: true},      // 11: end
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

func TestHitTest(t *testing.T) {
	l := testLayout()
	idx := l.HitTest(15, 5) // middle of 'e' on line 1
	if idx != 1 {
		t.Errorf("HitTest(15,5) = %d, want 1", idx)
	}
}

func TestHitTestMiss(t *testing.T) {
	l := testLayout()
	idx := l.HitTest(200, 200)
	if idx != -1 {
		t.Errorf("HitTest(200,200) = %d, want -1", idx)
	}
}

func TestHitTestRect(t *testing.T) {
	l := testLayout()
	r, ok := l.HitTestRect(5, 5)
	if !ok {
		t.Fatal("HitTestRect: expected ok")
	}
	if r.X != 0 || r.Width != 10 {
		t.Errorf("HitTestRect = %+v, want X=0 Width=10", r)
	}
}

func TestGetCharRect(t *testing.T) {
	l := testLayout()
	r, ok := l.GetCharRect(2)
	if !ok {
		t.Fatal("GetCharRect(2) not found")
	}
	if r.X != 20 {
		t.Errorf("GetCharRect(2).X = %f, want 20", r.X)
	}
}

func TestGetCharRectMissing(t *testing.T) {
	l := testLayout()
	_, ok := l.GetCharRect(5) // newline, no rect
	if ok {
		t.Error("GetCharRect(5) should not exist")
	}
}

func TestGetClosestOffset(t *testing.T) {
	l := testLayout()
	// Click in middle of char at index 3
	idx := l.GetClosestOffset(35, 10)
	if idx != 3 {
		t.Errorf("GetClosestOffset(35,10) = %d, want 3", idx)
	}
}

func TestGetClosestOffsetPastLine(t *testing.T) {
	l := testLayout()
	// Click past end of line 1
	idx := l.GetClosestOffset(200, 10)
	if idx != 5 {
		t.Errorf("GetClosestOffset(200,10) = %d, want 5", idx)
	}
}

func TestGetSelectionRects(t *testing.T) {
	l := testLayout()
	rects := l.GetSelectionRects(1, 4) // "ell"
	if len(rects) != 1 {
		t.Fatalf("GetSelectionRects(1,4) returned %d rects, want 1", len(rects))
	}
	if rects[0].Width != 30 {
		t.Errorf("selection width = %f, want 30", rects[0].Width)
	}
}

func TestGetSelectionRectsMultiLine(t *testing.T) {
	l := testLayout()
	rects := l.GetSelectionRects(2, 8) // cross-line selection
	if len(rects) != 2 {
		t.Fatalf("cross-line selection: %d rects, want 2", len(rects))
	}
}

func TestGetCursorPos(t *testing.T) {
	l := testLayout()
	pos, ok := l.GetCursorPos(0)
	if !ok {
		t.Fatal("GetCursorPos(0) not found")
	}
	if pos.X != 0 || pos.Height != 20 {
		t.Errorf("GetCursorPos(0) = %+v", pos)
	}
}

func TestGetCursorPosLineEnd(t *testing.T) {
	l := testLayout()
	pos, ok := l.GetCursorPos(5) // end of line 1
	if !ok {
		t.Fatal("GetCursorPos(5) not found")
	}
	if pos.X != 50 { // line width
		t.Errorf("cursor at line end: X=%f, want 50", pos.X)
	}
}

func TestGetValidCursorPositions(t *testing.T) {
	l := testLayout()
	positions := l.GetValidCursorPositions()
	if len(positions) == 0 {
		t.Fatal("no valid cursor positions")
	}
	// Should be sorted
	for i := 1; i < len(positions); i++ {
		if positions[i] <= positions[i-1] {
			t.Fatal("positions not sorted")
		}
	}
}

func TestMoveCursorLeft(t *testing.T) {
	l := testLayout()
	if got := l.MoveCursorLeft(3); got != 2 {
		t.Errorf("MoveCursorLeft(3) = %d, want 2", got)
	}
	if got := l.MoveCursorLeft(0); got != 0 {
		t.Errorf("MoveCursorLeft(0) = %d, want 0", got)
	}
}

func TestMoveCursorRight(t *testing.T) {
	l := testLayout()
	if got := l.MoveCursorRight(3); got != 4 {
		t.Errorf("MoveCursorRight(3) = %d, want 4", got)
	}
}

func TestMoveCursorWordLeft(t *testing.T) {
	l := testLayout()
	if got := l.MoveCursorWordLeft(8); got != 6 {
		t.Errorf("MoveCursorWordLeft(8) = %d, want 6", got)
	}
}

func TestMoveCursorWordRight(t *testing.T) {
	l := testLayout()
	if got := l.MoveCursorWordRight(0); got != 6 {
		t.Errorf("MoveCursorWordRight(0) = %d, want 6", got)
	}
}

func TestMoveCursorLineStart(t *testing.T) {
	l := testLayout()
	if got := l.MoveCursorLineStart(8); got != 6 {
		t.Errorf("MoveCursorLineStart(8) = %d, want 6", got)
	}
}

func TestMoveCursorLineEnd(t *testing.T) {
	l := testLayout()
	if got := l.MoveCursorLineEnd(2); got != 5 {
		t.Errorf("MoveCursorLineEnd(2) = %d, want 5", got)
	}
}

func TestMoveCursorLineEndSoftWrap(t *testing.T) {
	// Simulate soft-wrap: line 0 ends where line 1 starts.
	l := testLayout()
	l.Lines[0].Length = 6 // lineEnd = 6 = line 1 StartIndex
	l.Lines[1].StartIndex = 6
	// Cursor at boundary (col 0 of line 1) → end of line 1.
	if got := l.MoveCursorLineEnd(6); got != 11 {
		t.Errorf("MoveCursorLineEnd(6) = %d, want 11", got)
	}
}

func TestMoveCursorLineStartSoftWrap(t *testing.T) {
	l := testLayout()
	l.Lines[0].Length = 6
	l.Lines[1].StartIndex = 6
	// Cursor at boundary (col 0 of line 1) → start of line 1.
	if got := l.MoveCursorLineStart(6); got != 6 {
		t.Errorf("MoveCursorLineStart(6) = %d, want 6", got)
	}
}

func TestMoveCursorUp(t *testing.T) {
	l := testLayout()
	// From index 8 (line 2, x=25), move up should land on line 1
	got := l.MoveCursorUp(8, -1)
	if got < 0 || got > 5 {
		t.Errorf("MoveCursorUp(8) = %d, expected 0..5", got)
	}
}

func TestMoveCursorDown(t *testing.T) {
	l := testLayout()
	got := l.MoveCursorDown(2, -1)
	if got < 6 || got > 11 {
		t.Errorf("MoveCursorDown(2) = %d, expected 6..11", got)
	}
}

func TestMoveCursorUpFirstLine(t *testing.T) {
	l := testLayout()
	if got := l.MoveCursorUp(2, -1); got != 2 {
		t.Errorf("MoveCursorUp on first line = %d, want 2", got)
	}
}

func TestMoveCursorDownLastLine(t *testing.T) {
	l := testLayout()
	if got := l.MoveCursorDown(8, -1); got != 8 {
		t.Errorf("MoveCursorDown on last line = %d, want 8", got)
	}
}

func TestGetWordAtIndex(t *testing.T) {
	l := testLayout()
	start, end := l.GetWordAtIndex(2)
	if start != 0 || end != 5 {
		t.Errorf("GetWordAtIndex(2) = (%d, %d), want (0, 5)", start, end)
	}
}

func TestGetParagraphAtIndex(t *testing.T) {
	l := testLayout()
	text := "First paragraph\n\nSecond paragraph"
	start, end := l.GetParagraphAtIndex(5, text)
	if start != 0 || end != 15 {
		t.Errorf("GetParagraphAtIndex(5) = (%d, %d), want (0, 15)", start, end)
	}
}

func TestGetParagraphAtIndexSecond(t *testing.T) {
	l := testLayout()
	text := "First\n\nSecond"
	start, end := l.GetParagraphAtIndex(8, text)
	if start != 7 || end != len(text) {
		t.Errorf("GetParagraphAtIndex(8) = (%d, %d), want (7, %d)",
			start, end, len(text))
	}
}

func TestEmptyLayoutDefaults(t *testing.T) {
	l := Layout{}
	if got := l.HitTest(0, 0); got != -1 {
		t.Errorf("empty HitTest = %d, want -1", got)
	}
	if got := l.GetClosestOffset(0, 0); got != 0 {
		t.Errorf("empty GetClosestOffset = %d, want 0", got)
	}
}
