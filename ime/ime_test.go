package ime

import "testing"

func TestStubBridge(t *testing.T) {
	called := false
	b := NewBridge(Callbacks{
		OnMarkedText: func(string, int) { called = true },
		OnInsertText: func(string) { called = true },
		OnUnmarkText: func() { called = true },
	})

	// Stub bridge should be inactive.
	if b.IsActive() {
		t.Error("stub bridge should not be active")
	}

	// These should not panic.
	b.Enable(0, 0, 100, 20)
	b.Disable()
	b.SetCursorRect(10, 10, 2, 20)

	if called {
		t.Error("stub should not invoke callbacks")
	}
}

func TestStubBridgeType(t *testing.T) {
	b := NewBridge(Callbacks{})
	if _, ok := b.(stubBridge); !ok {
		t.Error("expected stubBridge on this platform")
	}
}
