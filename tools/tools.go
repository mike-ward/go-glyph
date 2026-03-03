//go:build tools

// Package tools pins documentation tooling dependencies so
// `go mod tidy` keeps them in go.sum.
package tools

import (
	_ "github.com/princjef/gomarkdoc/cmd/gomarkdoc"
	_ "golang.org/x/pkgsite/cmd/pkgsite"
)
