module glyph/examples/demo_sdl2

go 1.25.0

require (
	github.com/veandco/go-sdl2 v0.4.40
	glyph v0.0.0
	glyph/backend/sdl2 v0.0.0
)

replace (
	glyph => ../..
	glyph/backend/sdl2 => ../../backend/sdl2
)
