package glyph

import (
	"math"
	"testing"
)

const transformEpsilon = float32(0.0001)

func near(a, b float32) bool {
	return float32(math.Abs(float64(a-b))) < transformEpsilon
}

func TestAffineIdentity(t *testing.T) {
	tr := AffineIdentity()
	if tr.XX != 1 || tr.XY != 0 || tr.YX != 0 || tr.YY != 1 || tr.X0 != 0 || tr.Y0 != 0 {
		t.Errorf("identity fields: %+v", tr)
	}
	x, y := tr.Apply(3.5, -2.0)
	if !near(x, 3.5) || !near(y, -2.0) {
		t.Errorf("identity apply: got (%v, %v), want (3.5, -2.0)", x, y)
	}
}

func TestAffineRotationQuarterTurn(t *testing.T) {
	tr := AffineRotation(float32(math.Pi) * 0.5)
	x, y := tr.Apply(1.0, 0.0)
	if !near(x, 0.0) || !near(y, 1.0) {
		t.Errorf("rotation: got (%v, %v), want (0, 1)", x, y)
	}
}

func TestAffineTranslation(t *testing.T) {
	tr := AffineTranslation(5.0, -2.0)
	x, y := tr.Apply(3.0, 4.0)
	if !near(x, 8.0) || !near(y, 2.0) {
		t.Errorf("translation: got (%v, %v), want (8, 2)", x, y)
	}
}

func TestAffineSkew(t *testing.T) {
	tr := AffineSkew(0.5, -0.25)
	x, y := tr.Apply(4.0, 2.0)
	if !near(x, 5.0) || !near(y, 1.0) {
		t.Errorf("skew: got (%v, %v), want (5, 1)", x, y)
	}
}
