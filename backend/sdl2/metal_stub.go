//go:build (linux || windows) && !android

package sdl2

import "github.com/veandco/go-sdl2/sdl"

// SyncMetalLayer is a no-op on non-macOS platforms.
func SyncMetalLayer(_ *sdl.Renderer) {}
