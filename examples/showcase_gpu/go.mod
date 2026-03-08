module github.com/mike-ward/go-glyph/examples/showcase_gpu

go 1.26

require (
	github.com/mike-ward/go-glyph v1.0.0
	github.com/mike-ward/go-glyph/backend/gpu v1.0.0
	github.com/veandco/go-sdl2 v0.4.40
)

replace (
	github.com/mike-ward/go-glyph => ../..
	github.com/mike-ward/go-glyph/backend/gpu => ../../backend/gpu
)
