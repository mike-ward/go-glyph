//go:build ios

package ios

import "testing"

func TestBatchAppend6_SingleTexture(t *testing.T) {
	var b batch
	v := Vertex{}
	b.append6(1, v, v, v, v)
	b.append6(1, v, v, v, v)
	b.append6(1, v, v, v, v)

	if len(b.verts) != 18 {
		t.Fatalf("verts: got %d, want 18", len(b.verts))
	}
	if len(b.cmds) != 1 {
		t.Fatalf("cmds: got %d, want 1 (coalesced)", len(b.cmds))
	}
	if b.cmds[0].vertCount != 18 {
		t.Errorf("vertCount: got %d, want 18", b.cmds[0].vertCount)
	}
	if b.cmds[0].firstVert != 0 {
		t.Errorf("firstVert: got %d, want 0", b.cmds[0].firstVert)
	}
}

func TestBatchAppend6_AlternatingTextures(t *testing.T) {
	var b batch
	v := Vertex{}
	b.append6(1, v, v, v, v)
	b.append6(2, v, v, v, v)
	b.append6(1, v, v, v, v)

	if len(b.cmds) != 3 {
		t.Fatalf("cmds: got %d, want 3 (no coalescing)", len(b.cmds))
	}
	// Verify firstVert advances correctly.
	for i, want := range []int32{0, 6, 12} {
		if b.cmds[i].firstVert != want {
			t.Errorf("cmds[%d].firstVert: got %d, want %d",
				i, b.cmds[i].firstVert, want)
		}
	}
}

func TestBatchAppend6_CoalesceRunThenNew(t *testing.T) {
	var b batch
	v := Vertex{}
	b.append6(5, v, v, v, v)
	b.append6(5, v, v, v, v)
	b.append6(9, v, v, v, v)

	if len(b.cmds) != 2 {
		t.Fatalf("cmds: got %d, want 2", len(b.cmds))
	}
	if b.cmds[0].textureID != 5 || b.cmds[0].vertCount != 12 {
		t.Errorf("cmd[0]: texID=%d vertCount=%d, want 5/12",
			b.cmds[0].textureID, b.cmds[0].vertCount)
	}
	if b.cmds[1].textureID != 9 || b.cmds[1].vertCount != 6 {
		t.Errorf("cmd[1]: texID=%d vertCount=%d, want 9/6",
			b.cmds[1].textureID, b.cmds[1].vertCount)
	}
}

func TestBatchReset(t *testing.T) {
	var b batch
	v := Vertex{}
	b.append6(1, v, v, v, v)
	b.append6(2, v, v, v, v)
	b.reset()

	if len(b.verts) != 0 {
		t.Errorf("verts after reset: %d", len(b.verts))
	}
	if len(b.cmds) != 0 {
		t.Errorf("cmds after reset: %d", len(b.cmds))
	}
	// Verify capacity retained (no realloc on next frame).
	if cap(b.verts) < 12 {
		t.Errorf("verts cap after reset: %d, want >= 12", cap(b.verts))
	}
}

func TestBatchAppend6_Empty(t *testing.T) {
	var b batch
	if len(b.verts) != 0 || len(b.cmds) != 0 {
		t.Fatal("zero-value batch not empty")
	}
}
