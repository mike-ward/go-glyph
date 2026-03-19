module github.com/mike-ward/go-glyph/examples/showcase_android

go 1.26

require (
	github.com/mike-ward/go-glyph v1.0.0
	github.com/mike-ward/go-glyph/backend/android v0.0.0
	github.com/mike-ward/go-glyph/examples/showcase_sections v0.0.0
)

require github.com/rivo/uniseg v0.2.0 // indirect

replace (
	github.com/mike-ward/go-glyph => ../..
	github.com/mike-ward/go-glyph/backend/android => ../../backend/android
	github.com/mike-ward/go-glyph/examples/showcase_sections => ../showcase_sections
)
