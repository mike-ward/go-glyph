module github.com/mike-ward/go-glyph/examples/demo_sdl2

go 1.25.0

require (
	github.com/mike-ward/go-glyph v0.0.0
	github.com/mike-ward/go-glyph/backend/sdl2 v0.0.0
	github.com/veandco/go-sdl2 v0.4.40
)

replace (
	github.com/mike-ward/go-glyph => ../..
	github.com/mike-ward/go-glyph/backend/sdl2 => ../../backend/sdl2
)
