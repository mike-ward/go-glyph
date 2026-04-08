//go:build !windows

package glyph

// SetDPIAwareWindows is a no-op on non-Windows platforms. On Windows
// it marks the process as per-monitor DPI aware v2 so windows render
// at native physical resolution. See dpi_windows.go.
func SetDPIAwareWindows() {}
