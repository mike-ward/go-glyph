//go:build android

package android

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

func TestBatchAppend6_TextureIDZero(t *testing.T) {
	var b batch
	v := Vertex{}
	b.append6(0, v, v, v, v)
	b.append6(0, v, v, v, v)

	if len(b.cmds) != 1 {
		t.Fatalf("cmds: got %d, want 1 (coalesced)", len(b.cmds))
	}
	if b.cmds[0].textureID != 0 {
		t.Errorf("textureID: got %d, want 0", b.cmds[0].textureID)
	}
}

func TestBatchAppend6_VertexDataPreserved(t *testing.T) {
	var b batch
	v0 := Vertex{PosX: 1, PosY: 2, R: 10, G: 20, B: 30, A: 40, TexU: 0.5, TexV: 0.75}
	v1 := Vertex{PosX: 3, PosY: 4}
	v2 := Vertex{PosX: 5, PosY: 6}
	v3 := Vertex{PosX: 7, PosY: 8}
	b.append6(1, v0, v1, v2, v3)

	// Expected: v0, v1, v2, v0, v2, v3
	if b.verts[0] != v0 {
		t.Errorf("verts[0]: got %+v, want %+v", b.verts[0], v0)
	}
	if b.verts[1] != v1 {
		t.Errorf("verts[1]: got %+v, want %+v", b.verts[1], v1)
	}
	if b.verts[2] != v2 {
		t.Errorf("verts[2]: got %+v, want %+v", b.verts[2], v2)
	}
	if b.verts[3] != v0 {
		t.Errorf("verts[3]: got %+v, want %+v", b.verts[3], v0)
	}
	if b.verts[4] != v2 {
		t.Errorf("verts[4]: got %+v, want %+v", b.verts[4], v2)
	}
	if b.verts[5] != v3 {
		t.Errorf("verts[5]: got %+v, want %+v", b.verts[5], v3)
	}
}

func TestBatchAppend6_LargeBatch(t *testing.T) {
	var b batch
	v := Vertex{}
	const n = 1000
	for i := 0; i < n; i++ {
		b.append6(1, v, v, v, v)
	}
	if len(b.verts) != n*6 {
		t.Errorf("verts: got %d, want %d", len(b.verts), n*6)
	}
	if len(b.cmds) != 1 {
		t.Errorf("cmds: got %d, want 1 (all same texture)", len(b.cmds))
	}
	if b.cmds[0].vertCount != int32(n*6) {
		t.Errorf("vertCount: got %d, want %d", b.cmds[0].vertCount, n*6)
	}
}

func TestBatchReset_RetainsCapacity(t *testing.T) {
	var b batch
	v := Vertex{}
	for i := 0; i < 100; i++ {
		b.append6(uint64(i%3), v, v, v, v)
	}
	vertCap := cap(b.verts)
	cmdCap := cap(b.cmds)
	b.reset()

	if cap(b.verts) < vertCap {
		t.Errorf("verts cap shrunk: %d < %d", cap(b.verts), vertCap)
	}
	if cap(b.cmds) < cmdCap {
		t.Errorf("cmds cap shrunk: %d < %d", cap(b.cmds), cmdCap)
	}
}

func BenchmarkBatchAppend6(b *testing.B) {
	var bt batch
	v := Vertex{}
	b.ResetTimer()
	for b.Loop() {
		bt.append6(1, v, v, v, v)
		if len(bt.verts) > 60000 {
			bt.reset()
		}
	}
}
