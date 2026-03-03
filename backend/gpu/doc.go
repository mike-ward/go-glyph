// Package gpu provides a raw OpenGL 3.3 [glyph.DrawBackend] backed by SDL2.
// It renders directly to an OpenGL context obtained from an SDL_Window,
// bypassing SDL2's own renderer.
//
// Create a backend with [New], then pass it to glyph.NewRenderer each frame:
//
//	b, err := gpu.New(sdlWindowPtr, dpiScale)
//	renderer := glyph.NewRenderer(b, ctx)
//
//	// Per-frame loop:
//	b.BeginFrame()
//	renderer.DrawLayout(layout, x, y)
//	b.EndFrame(0, 0, 0, 1, logW, logH)
package gpu
