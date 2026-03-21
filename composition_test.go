package glyph

import "testing"

func TestCompositionStateLifecycle(t *testing.T) {
	cs := NewCompositionState()
	if cs.IsComposing() {
		t.Error("should not be composing initially")
	}

	cs.Start(10)
	if !cs.IsComposing() {
		t.Error("should be composing after Start")
	}
	if cs.PreeditStart != 10 {
		t.Errorf("PreeditStart = %d, want 10", cs.PreeditStart)
	}

	cs.SetMarkedText("abc", 2)
	if cs.PreeditText != "abc" {
		t.Errorf("PreeditText = %q", cs.PreeditText)
	}
	if cs.CursorOffset != 2 {
		t.Errorf("CursorOffset = %d", cs.CursorOffset)
	}

	result := cs.Commit()
	if result != "abc" {
		t.Errorf("Commit = %q, want 'abc'", result)
	}
	if cs.IsComposing() {
		t.Error("should not be composing after Commit")
	}
}

func TestCompositionStateReset(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(5)
	cs.SetMarkedText("test", 3)
	cs.Reset()

	if cs.IsComposing() {
		t.Error("should not be composing after Reset")
	}
	if cs.PreeditText != "" {
		t.Errorf("PreeditText = %q after Reset", cs.PreeditText)
	}
}

func TestCompositionDocumentCursorPos(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(10)
	cs.SetMarkedText("hello", 3)
	if got := cs.DocumentCursorPos(); got != 13 {
		t.Errorf("DocumentCursorPos = %d, want 13", got)
	}
}

func TestCompositionPreeditEnd(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(10)
	cs.SetMarkedText("hello", 5)
	if got := cs.PreeditEnd(); got != 15 {
		t.Errorf("PreeditEnd = %d, want 15", got)
	}
}

func TestCompositionBoundsNotComposing(t *testing.T) {
	cs := NewCompositionState()
	_, ok := cs.CompositionBounds(Layout{})
	if ok {
		t.Error("should not have bounds when not composing")
	}
}

func TestCompositionClauses(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(0)
	cs.SetMarkedText("test", 4)

	cs.HandleClause(0, 2, 0) // raw
	cs.HandleClause(2, 2, 2) // selected

	if len(cs.Clauses) != 2 {
		t.Fatalf("clause count = %d, want 2", len(cs.Clauses))
	}
	if cs.Clauses[0].Style != ClauseRaw {
		t.Errorf("clause 0 style = %d", cs.Clauses[0].Style)
	}
	if cs.Clauses[1].Style != ClauseSelected {
		t.Errorf("clause 1 style = %d", cs.Clauses[1].Style)
	}
}

func TestCompositionClearClauses(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(0)
	cs.HandleClause(0, 5, 1)
	cs.ClearClauses()
	if len(cs.Clauses) != 0 {
		t.Errorf("clauses not cleared: len=%d", len(cs.Clauses))
	}
	if cs.SelectedClause != -1 {
		t.Errorf("SelectedClause = %d, want -1", cs.SelectedClause)
	}
}

func TestCompositionHandleMarkedText(t *testing.T) {
	cs := NewCompositionState()
	cs.HandleMarkedText("hello", 3, 10)
	if !cs.IsComposing() {
		t.Error("should be composing after HandleMarkedText")
	}
	if cs.PreeditText != "hello" {
		t.Errorf("PreeditText = %q", cs.PreeditText)
	}
}

func TestCompositionHandleInsertText(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(5)
	cs.SetMarkedText("test", 4)

	result := cs.HandleInsertText("final")
	if result != "final" {
		t.Errorf("HandleInsertText = %q, want 'final'", result)
	}
	if cs.IsComposing() {
		t.Error("should not be composing after HandleInsertText")
	}
}

func TestCompositionHandleUnmarkText(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(0)
	cs.SetMarkedText("abc", 3)
	cs.HandleUnmarkText()
	if cs.IsComposing() {
		t.Error("should not be composing after HandleUnmarkText")
	}
}

func TestCompositionHandleClauseInvalid(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(0)
	cs.HandleClause(-1, 5, 0)
	cs.HandleClause(0, -1, 0)
	if len(cs.Clauses) != 0 {
		t.Error("invalid clauses should be rejected")
	}
}

func TestDeadKeyLifecycle(t *testing.T) {
	dks := DeadKeyState{}
	if dks.HasPending {
		t.Error("should not have pending initially")
	}

	dks.StartDeadKey('`', 5)
	if !dks.HasPending {
		t.Error("should have pending after StartDeadKey")
	}

	result, combined := dks.TryCombine('e')
	if !combined {
		t.Error("grave + e should combine")
	}
	if result != "\u00e8" { // è
		t.Errorf("result = %q, want è", result)
	}
	if dks.HasPending {
		t.Error("pending should be cleared after combine")
	}
}

func TestDeadKeyInvalidCombination(t *testing.T) {
	dks := DeadKeyState{}
	dks.StartDeadKey('`', 0)

	result, combined := dks.TryCombine('x')
	if combined {
		t.Error("grave + x should not combine")
	}
	if result != "`x" {
		t.Errorf("result = %q, want '`x'", result)
	}
}

func TestDeadKeyClear(t *testing.T) {
	dks := DeadKeyState{}
	dks.StartDeadKey('^', 0)
	dks.Clear()
	if dks.HasPending {
		t.Error("should not have pending after Clear")
	}
}

func TestDeadKeyNoPending(t *testing.T) {
	dks := DeadKeyState{}
	result, combined := dks.TryCombine('e')
	if combined || result != "" {
		t.Error("no pending should return empty")
	}
}

func TestIsDeadKey(t *testing.T) {
	deadKeys := []rune{'`', '\'', '^', '~', '"', ':', ','}
	for _, r := range deadKeys {
		if !IsDeadKey(r) {
			t.Errorf("IsDeadKey(%q) = false", r)
		}
	}
	if IsDeadKey('a') {
		t.Error("'a' should not be dead key")
	}
}

func TestCombineDeadKeyAllAccents(t *testing.T) {
	tests := []struct {
		dead, base rune
		want       rune
	}{
		{'`', 'a', 0x00E0},  // à
		{'\'', 'e', 0x00E9}, // é
		{'^', 'i', 0x00EE},  // î
		{'~', 'n', 0x00F1},  // ñ
		{'"', 'u', 0x00FC},  // ü
		{':', 'o', 0x00F6},  // ö
		{',', 'c', 0x00E7},  // ç
		{',', 'C', 0x00C7},  // Ç
		{'`', 'A', 0x00C0},  // À
		{'~', 'O', 0x00D5},  // Õ
	}
	for _, tc := range tests {
		got, ok := combineDeadKey(tc.dead, tc.base)
		if !ok {
			t.Errorf("combineDeadKey(%q, %q) failed", tc.dead, tc.base)
			continue
		}
		if got != tc.want {
			t.Errorf("combineDeadKey(%q, %q) = %U, want %U",
				tc.dead, tc.base, got, tc.want)
		}
	}
}

func TestGetClauseRectsNoClauses(t *testing.T) {
	cs := NewCompositionState()
	cs.Start(0)
	cs.SetMarkedText("hi", 2)

	l := testLayout()
	rects := cs.GetClauseRects(l)
	// No explicit clauses, but preedit exists: should return
	// single raw clause rect.
	if len(rects) != 1 {
		t.Fatalf("clause rects = %d, want 1", len(rects))
	}
	if rects[0].Style != ClauseRaw {
		t.Errorf("style = %d, want ClauseRaw", rects[0].Style)
	}
}
