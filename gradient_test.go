package glyph

import "testing"

func TestLerpColorEndpoints(t *testing.T) {
	a := Color{0, 0, 0, 255}
	b := Color{255, 255, 255, 255}

	c0 := LerpColor(a, b, 0.0)
	if c0.R != 0 || c0.G != 0 || c0.B != 0 || c0.A != 255 {
		t.Errorf("lerp t=0: %+v", c0)
	}
	c1 := LerpColor(a, b, 1.0)
	if c1.R != 255 || c1.G != 255 || c1.B != 255 || c1.A != 255 {
		t.Errorf("lerp t=1: %+v", c1)
	}
}

func TestLerpColorMidpoint(t *testing.T) {
	a := Color{0, 0, 0, 0}
	b := Color{200, 100, 50, 250}
	c := LerpColor(a, b, 0.5)
	if c.R != 100 || c.G != 50 || c.B != 25 || c.A != 125 {
		t.Errorf("lerp mid: %+v", c)
	}
}

func TestLerpColorClampsT(t *testing.T) {
	a := Color{10, 20, 30, 40}
	b := Color{110, 120, 130, 140}

	cNeg := LerpColor(a, b, -5.0)
	if cNeg.R != a.R || cNeg.G != a.G {
		t.Errorf("lerp t<0: %+v", cNeg)
	}
	cOver := LerpColor(a, b, 10.0)
	if cOver.R != b.R || cOver.G != b.G {
		t.Errorf("lerp t>1: %+v", cOver)
	}
}

func TestGradientColorAtEmptyStops(t *testing.T) {
	c := GradientColorAt(nil, 0.5)
	if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Errorf("empty stops: %+v", c)
	}
}

func TestGradientColorAtSingleStop(t *testing.T) {
	stops := []GradientStop{{Color: Color{100, 150, 200, 255}, Position: 0.5}}
	c0 := GradientColorAt(stops, 0.0)
	if c0.R != 100 {
		t.Errorf("single stop t=0: R=%d", c0.R)
	}
	c1 := GradientColorAt(stops, 1.0)
	if c1.R != 100 {
		t.Errorf("single stop t=1: R=%d", c1.R)
	}
}

func TestGradientColorAtTwoStops(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{0, 0, 0, 255}, Position: 0.0},
		{Color: Color{200, 100, 50, 255}, Position: 1.0},
	}
	c := GradientColorAt(stops, 0.5)
	if c.R != 100 || c.G != 50 || c.B != 25 {
		t.Errorf("two stops mid: %+v", c)
	}
}

func TestGradientColorAtFourStops(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{255, 0, 0, 255}, Position: 0.0},
		{Color: Color{0, 255, 0, 255}, Position: 0.33},
		{Color: Color{0, 0, 255, 255}, Position: 0.66},
		{Color: Color{255, 255, 255, 255}, Position: 1.0},
	}
	c0 := GradientColorAt(stops, 0.0)
	if c0.R != 255 || c0.G != 0 {
		t.Errorf("4-stop t=0: %+v", c0)
	}
	c1 := GradientColorAt(stops, 0.33)
	if c1.R != 0 || c1.G != 255 {
		t.Errorf("4-stop t=0.33: %+v", c1)
	}
	c3 := GradientColorAt(stops, 1.0)
	if c3.R != 255 || c3.G != 255 || c3.B != 255 {
		t.Errorf("4-stop t=1: %+v", c3)
	}
}

func TestGradientColorAtBeforeFirstStop(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{50, 100, 150, 200}, Position: 0.3},
		{Color: Color{200, 200, 200, 255}, Position: 0.8},
	}
	c := GradientColorAt(stops, 0.0)
	if c.R != 50 || c.G != 100 {
		t.Errorf("before first: %+v", c)
	}
}

func TestGradientColorAtAfterLastStop(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{50, 100, 150, 200}, Position: 0.2},
		{Color: Color{200, 200, 200, 255}, Position: 0.7},
	}
	c := GradientColorAt(stops, 1.0)
	if c.R != 200 || c.A != 255 {
		t.Errorf("after last: %+v", c)
	}
}

func TestGradientColorAtCoincidentPositions(t *testing.T) {
	stops := []GradientStop{
		{Color: Color{255, 0, 0, 255}, Position: 0.5},
		{Color: Color{0, 0, 255, 255}, Position: 0.5},
	}
	c := GradientColorAt(stops, 0.5)
	if c.R != 255 || c.B != 0 {
		t.Errorf("coincident: %+v", c)
	}
}
