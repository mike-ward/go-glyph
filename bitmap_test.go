package glyph

import (
	"math"
	"testing"
)

func TestCheckAllocationSize(t *testing.T) {
	tests := []struct {
		w, h, ch int
		wantErr  bool
	}{
		{100, 100, 4, false},
		{0, 100, 4, true},
		{100, 0, 4, true},
		{-1, 100, 4, true},
		{100, -1, 4, true},
		{50000, 50000, 4, true}, // exceeds 1GB
	}
	for _, tt := range tests {
		_, err := checkAllocationSize(tt.w, tt.h, tt.ch)
		if (err != nil) != tt.wantErr {
			t.Errorf("checkAllocationSize(%d,%d,%d): err=%v, wantErr=%v",
				tt.w, tt.h, tt.ch, err, tt.wantErr)
		}
	}
}

func TestCheckAllocationSizeValid(t *testing.T) {
	size, err := checkAllocationSize(1024, 1024, 4)
	if err != nil {
		t.Fatal(err)
	}
	if size != 1024*1024*4 {
		t.Errorf("size = %d, want %d", size, 1024*1024*4)
	}
}

func TestScaleBitmapBicubicIdentity(t *testing.T) {
	// 2x2 RGBA bitmap.
	src := []byte{
		255, 0, 0, 255, 0, 255, 0, 255,
		0, 0, 255, 255, 255, 255, 0, 255,
	}

	// Scale to same size.
	dst := ScaleBitmapBicubic(src, 2, 2, 2, 2)
	if dst == nil {
		t.Fatal("ScaleBitmapBicubic returned nil")
	}
	if len(dst) != 2*2*4 {
		t.Fatalf("dst len = %d, want %d", len(dst), 2*2*4)
	}

	// Corners should roughly match originals.
	// Allow tolerance due to Catmull-Rom boundary effects.
	tolerance := byte(50)
	if absDiffByte(dst[0], 255) > tolerance {
		t.Errorf("pixel(0,0) R=%d, want ~255", dst[0])
	}
	if absDiffByte(dst[3], 255) > tolerance {
		t.Errorf("pixel(0,0) A=%d, want ~255", dst[3])
	}
}

func TestScaleBitmapBicubicUpscale(t *testing.T) {
	// 1x1 solid white.
	src := []byte{255, 255, 255, 255}
	dst := ScaleBitmapBicubic(src, 1, 1, 4, 4)
	if dst == nil {
		t.Fatal("nil result")
	}
	if len(dst) != 4*4*4 {
		t.Fatalf("len = %d, want %d", len(dst), 4*4*4)
	}

	// All pixels should be solid white.
	for i := 0; i < len(dst); i += 4 {
		if dst[i] != 255 || dst[i+1] != 255 || dst[i+2] != 255 || dst[i+3] != 255 {
			t.Errorf("pixel at offset %d: RGBA=(%d,%d,%d,%d), want solid white",
				i/4, dst[i], dst[i+1], dst[i+2], dst[i+3])
			break
		}
	}
}

func TestScaleBitmapBicubicDownscale(t *testing.T) {
	// 4x4 solid red.
	src := make([]byte, 4*4*4)
	for i := 0; i < len(src); i += 4 {
		src[i] = 200
		src[i+1] = 50
		src[i+2] = 50
		src[i+3] = 255
	}

	dst := ScaleBitmapBicubic(src, 4, 4, 2, 2)
	if dst == nil {
		t.Fatal("nil result")
	}

	// Downscaled solid color should remain approximately same.
	tolerance := byte(30)
	for i := 0; i < len(dst); i += 4 {
		if absDiffByte(dst[i], 200) > tolerance ||
			absDiffByte(dst[i+1], 50) > tolerance ||
			absDiffByte(dst[i+3], 255) > tolerance {
			t.Errorf("pixel at %d: RGBA=(%d,%d,%d,%d), want ~(200,50,50,255)",
				i/4, dst[i], dst[i+1], dst[i+2], dst[i+3])
			break
		}
	}
}

func TestScaleBitmapBicubicZero(t *testing.T) {
	if ScaleBitmapBicubic(nil, 0, 0, 0, 0) != nil {
		t.Error("expected nil for zero dimensions")
	}
	if ScaleBitmapBicubic([]byte{1, 2, 3, 4}, 1, 1, 0, 0) != nil {
		t.Error("expected nil for zero dst dimensions")
	}
}

func TestCubicHermite(t *testing.T) {
	// At t=0, result should be p1.
	v := cubicHermite(0, 100, 200, 300, 0)
	if math.Abs(float64(v-100)) > 0.001 {
		t.Errorf("cubicHermite(t=0) = %f, want 100", v)
	}
	// At t=1, result should be p2.
	v = cubicHermite(0, 100, 200, 300, 1)
	if math.Abs(float64(v-200)) > 0.001 {
		t.Errorf("cubicHermite(t=1) = %f, want 200", v)
	}
}

func TestClampInt(t *testing.T) {
	if clampInt(-5, 0, 10) != 0 {
		t.Error("clampInt(-5,0,10)")
	}
	if clampInt(15, 0, 10) != 10 {
		t.Error("clampInt(15,0,10)")
	}
	if clampInt(5, 0, 10) != 5 {
		t.Error("clampInt(5,0,10)")
	}
}

func TestClampF32(t *testing.T) {
	if clampF32(-1.5, 0, 1) != 0 {
		t.Error("clampF32 low")
	}
	if clampF32(2.5, 0, 1) != 1 {
		t.Error("clampF32 high")
	}
	if clampF32(0.5, 0, 1) != 0.5 {
		t.Error("clampF32 mid")
	}
}

func TestGetPixelRGBAPremul(t *testing.T) {
	// 1x1 pixel: R=200, G=100, B=50, A=128
	src := []byte{200, 100, 50, 128}
	r, g, b, a := getPixelRGBAPremul(src, 1, 1, 0, 0)

	// Premultiply: f = 128/255 ≈ 0.502
	f := float32(128) / 255.0
	expectR := float32(200) * f
	expectG := float32(100) * f
	expectB := float32(50) * f

	tolerance := float32(0.5)
	if absDiffF32(r, expectR) > tolerance {
		t.Errorf("R = %f, want %f", r, expectR)
	}
	if absDiffF32(g, expectG) > tolerance {
		t.Errorf("G = %f, want %f", g, expectG)
	}
	if absDiffF32(b, expectB) > tolerance {
		t.Errorf("B = %f, want %f", b, expectB)
	}
	if absDiffF32(a, 128) > tolerance {
		t.Errorf("A = %f, want 128", a)
	}
}

func TestGetPixelRGBAPremulClamped(t *testing.T) {
	src := []byte{255, 255, 255, 255}
	// Out of bounds should clamp to edge.
	r, _, _, a := getPixelRGBAPremul(src, 1, 1, 5, 5)
	if r != 255 || a != 255 {
		t.Errorf("clamped pixel: R=%f A=%f, want 255", r, a)
	}
}

func TestGetPixelRGBAPremulEmptyImage(t *testing.T) {
	r, g, b, a := getPixelRGBAPremul(nil, 0, 0, 0, 0)
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Error("expected zero for empty image")
	}
}

func absDiffByte(a, b byte) byte {
	if a > b {
		return a - b
	}
	return b - a
}

func absDiffF32(a, b float32) float32 {
	d := a - b
	if d < 0 {
		return -d
	}
	return d
}
