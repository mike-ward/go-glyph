//go:build windows

package glyph

import "os"

// registerWindowsFonts registers standard Windows font directories
// with FontConfig. This is a stub — it will be wired into
// NewContext once Windows text shaping is implemented.
func (ctx *Context) registerWindowsFonts() {
	_ = windowsFontDirs() // future: register with FontConfig
}

// windowsFontDirs returns the standard Windows font directories.
func windowsFontDirs() []string {
	dirs := []string{`C:\Windows\Fonts`}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		dirs = append(dirs, local+`\Microsoft\Windows\Fonts`)
	}
	return dirs
}
