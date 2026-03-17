module github.com/mike-ward/go-glyph/examples/showcase_ios

go 1.26

require (
	github.com/mike-ward/go-glyph v1.0.0
	github.com/mike-ward/go-glyph/backend/ios v0.0.0
	github.com/mike-ward/go-glyph/examples/showcase_sections v0.0.0
)

replace (
	github.com/mike-ward/go-glyph => ../..
	github.com/mike-ward/go-glyph/backend/ios => ../../backend/ios
	github.com/mike-ward/go-glyph/examples/showcase_sections => ../showcase_sections
)
