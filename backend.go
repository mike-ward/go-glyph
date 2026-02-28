package glyph

// TextureID is an opaque handle to a GPU texture managed by a DrawBackend.
type TextureID uint64

// DrawBackend abstracts the GPU rendering backend. Implementations
// provide texture management and drawing primitives. The Renderer
// calls only these methods — never a specific backend directly.
type DrawBackend interface {
	// NewTexture allocates a new RGBA texture of the given size.
	NewTexture(width, height int) TextureID

	// UpdateTexture uploads RGBA pixel data to an existing texture.
	// data must be width*height*4 bytes.
	UpdateTexture(id TextureID, data []byte)

	// DeleteTexture releases a texture.
	DeleteTexture(id TextureID)

	// DrawTexturedQuad draws a textured rectangle with color tinting.
	DrawTexturedQuad(id TextureID, src, dst Rect, c Color)

	// DrawFilledRect draws an untextured filled rectangle.
	DrawFilledRect(dst Rect, c Color)

	// DrawTexturedQuadTransformed draws a textured quad with an
	// affine transform applied.
	DrawTexturedQuadTransformed(id TextureID, src, dst Rect, c Color, t AffineTransform)

	// DPIScale returns the display DPI scale factor.
	DPIScale() float32
}
