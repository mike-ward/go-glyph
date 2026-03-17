//go:build ios

// Package ios provides a Metal DrawBackend for iOS. It accepts a
// CAMetalLayer directly instead of using SDL2, making it suitable
// for native iOS apps that compile Go as a c-archive.
package ios
