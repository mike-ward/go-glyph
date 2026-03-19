//go:build android

// Package android provides an OpenGL ES 3.0 DrawBackend for
// Android. It accepts an ANativeWindow directly via EGL, making it
// suitable for native Android apps that load Go as a c-shared .so.
package android
