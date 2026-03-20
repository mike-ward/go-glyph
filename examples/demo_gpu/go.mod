module github.com/mike-ward/go-glyph/examples/demo_gpu

go 1.26

require (
	github.com/mike-ward/go-glyph v1.4.1
	github.com/veandco/go-sdl2 v0.4.40
)

replace github.com/mike-ward/go-glyph => ../..

require github.com/rivo/uniseg v0.2.0 // indirect
