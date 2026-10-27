// Package codi provides embedded resources for the codi CLI.
package codi

import "embed"

// EmbeddedFS contains all embedded standards and templates.
//
//go:embed standards templates
//go:embed templates/go/base/quality/.gitignore
//go:embed templates/go/base/quality/.golangci.yml.tmpl
//go:embed templates/python/base/quality/.gitignore
var EmbeddedFS embed.FS
