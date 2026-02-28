package glyph

import "testing"

func TestInsertText(t *testing.T) {
	r := InsertText("Hello", 5, " World")
	if r.NewText != "Hello World" {
		t.Errorf("NewText = %q, want %q", r.NewText, "Hello World")
	}
	if r.CursorPos != 11 {
		t.Errorf("CursorPos = %d, want 11", r.CursorPos)
	}
}

func TestInsertTextAtStart(t *testing.T) {
	r := InsertText("World", 0, "Hello ")
	if r.NewText != "Hello World" {
		t.Errorf("NewText = %q", r.NewText)
	}
	if r.CursorPos != 6 {
		t.Errorf("CursorPos = %d, want 6", r.CursorPos)
	}
}

func TestDeleteBackward(t *testing.T) {
	l := testLayout()
	r := DeleteBackward("Hello\nWorld", l, 3)
	if r.NewText != "Helo\nWorld" {
		t.Errorf("NewText = %q", r.NewText)
	}
	if r.CursorPos != 2 {
		t.Errorf("CursorPos = %d, want 2", r.CursorPos)
	}
	if r.DeletedText != "l" {
		t.Errorf("DeletedText = %q, want 'l'", r.DeletedText)
	}
}

func TestDeleteBackwardAtStart(t *testing.T) {
	l := testLayout()
	r := DeleteBackward("Hello", l, 0)
	if r.NewText != "Hello" {
		t.Errorf("text changed at start: %q", r.NewText)
	}
}

func TestDeleteForward(t *testing.T) {
	l := testLayout()
	r := DeleteForward("Hello\nWorld", l, 2)
	if r.NewText != "Helo\nWorld" {
		t.Errorf("NewText = %q", r.NewText)
	}
	if r.CursorPos != 2 {
		t.Errorf("CursorPos = %d, want 2", r.CursorPos)
	}
}

func TestDeleteSelection(t *testing.T) {
	r := DeleteSelection("Hello World", 5, 11)
	if r.NewText != "Hello" {
		t.Errorf("NewText = %q, want 'Hello'", r.NewText)
	}
	if r.DeletedText != " World" {
		t.Errorf("DeletedText = %q", r.DeletedText)
	}
	if r.CursorPos != 5 {
		t.Errorf("CursorPos = %d, want 5", r.CursorPos)
	}
}

func TestDeleteSelectionReversed(t *testing.T) {
	r := DeleteSelection("Hello World", 11, 5)
	if r.NewText != "Hello" {
		t.Errorf("reversed: NewText = %q", r.NewText)
	}
}

func TestDeleteSelectionNoSelection(t *testing.T) {
	r := DeleteSelection("Hello", 3, 3)
	if r.NewText != "Hello" {
		t.Errorf("no selection: text changed")
	}
}

func TestInsertReplacingSelection(t *testing.T) {
	r := InsertReplacingSelection("Hello World", 6, 11, "Go")
	if r.NewText != "Hello Go" {
		t.Errorf("NewText = %q, want 'Hello Go'", r.NewText)
	}
	if r.CursorPos != 8 {
		t.Errorf("CursorPos = %d, want 8", r.CursorPos)
	}
	if r.DeletedText != "World" {
		t.Errorf("DeletedText = %q", r.DeletedText)
	}
}

func TestInsertReplacingNoSelection(t *testing.T) {
	r := InsertReplacingSelection("Hello", 5, 5, "!")
	if r.NewText != "Hello!" {
		t.Errorf("NewText = %q", r.NewText)
	}
}

func TestGetSelectedText(t *testing.T) {
	got := GetSelectedText("Hello World", 6, 11)
	if got != "World" {
		t.Errorf("GetSelectedText = %q, want 'World'", got)
	}
}

func TestGetSelectedTextReversed(t *testing.T) {
	got := GetSelectedText("Hello World", 11, 6)
	if got != "World" {
		t.Errorf("reversed GetSelectedText = %q", got)
	}
}

func TestGetSelectedTextEmpty(t *testing.T) {
	got := GetSelectedText("Hello", 3, 3)
	if got != "" {
		t.Errorf("no selection = %q, want empty", got)
	}
}

func TestCutSelection(t *testing.T) {
	cut, result := CutSelection("Hello World", 0, 5)
	if cut != "Hello" {
		t.Errorf("cut = %q, want 'Hello'", cut)
	}
	if result.NewText != " World" {
		t.Errorf("NewText = %q", result.NewText)
	}
}

func TestCutNoSelection(t *testing.T) {
	cut, result := CutSelection("Hello", 3, 3)
	if cut != "" {
		t.Errorf("cut = %q, want empty", cut)
	}
	if result.NewText != "Hello" {
		t.Errorf("text changed")
	}
}

func TestDeleteToWordBoundary(t *testing.T) {
	l := testLayout()
	r := DeleteToWordBoundary("Hello\nWorld", l, 8)
	// word_left from 8 = 6
	if r.CursorPos != 6 {
		t.Errorf("CursorPos = %d, want 6", r.CursorPos)
	}
	if r.DeletedText != "Wo" {
		t.Errorf("DeletedText = %q, want 'Wo'", r.DeletedText)
	}
}

func TestDeleteToLineStart(t *testing.T) {
	l := testLayout()
	r := DeleteToLineStart("Hello\nWorld", l, 8)
	if r.CursorPos != 6 {
		t.Errorf("CursorPos = %d, want 6", r.CursorPos)
	}
}

func TestDeleteToLineEnd(t *testing.T) {
	l := testLayout()
	r := DeleteToLineEnd("Hello\nWorld", l, 7)
	if r.CursorPos != 7 {
		t.Errorf("CursorPos = %d, want 7", r.CursorPos)
	}
}

func TestDeleteToWordEnd(t *testing.T) {
	l := testLayout()
	r := DeleteToWordEnd("Hello\nWorld", l, 0)
	// word_right from 0 = 6
	if r.DeletedText != "Hello\n" {
		t.Errorf("DeletedText = %q, want 'Hello\\n'", r.DeletedText)
	}
}

func TestMutationResultToChange(t *testing.T) {
	m := MutationResult{
		RangeStart:  5,
		RangeEnd:    10,
		DeletedText: "World",
	}
	c := m.ToChange("Go")
	if c.OldText != "World" || c.NewText != "Go" {
		t.Errorf("ToChange: old=%q new=%q", c.OldText, c.NewText)
	}
}

func TestClampIndex(t *testing.T) {
	if got := clampIndex(-5, 10); got != 0 {
		t.Errorf("clampIndex(-5,10) = %d", got)
	}
	if got := clampIndex(15, 10); got != 10 {
		t.Errorf("clampIndex(15,10) = %d", got)
	}
	if got := clampIndex(5, 10); got != 5 {
		t.Errorf("clampIndex(5,10) = %d", got)
	}
}
