module glyph/examples/showcase_gpu

go 1.25.0

require (
	github.com/veandco/go-sdl2 v0.4.40
	glyph v0.0.0
	glyph/backend/gpu v0.0.0
)

replace (
	glyph => ../..
	glyph/backend/gpu => ../../backend/gpu
)
