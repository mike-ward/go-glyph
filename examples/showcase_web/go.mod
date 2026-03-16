module github.com/mike-ward/go-glyph/examples/showcase_web

go 1.26

require (
	github.com/mike-ward/go-glyph v1.0.0
	github.com/mike-ward/go-glyph/backend/web v1.0.0
	github.com/mike-ward/go-glyph/examples/showcase_sections v1.0.0
)

replace (
	github.com/mike-ward/go-glyph => ../..
	github.com/mike-ward/go-glyph/backend/web => ../../backend/web
	github.com/mike-ward/go-glyph/examples/showcase_sections => ../showcase_sections
)
