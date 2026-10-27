// Package codi provides embedded resources for the codi CLI.
package codi

import "embed"

// EmbeddedFS contains all embedded standards and templates.
//
//go:embed standards templates
var EmbeddedFS embed.FS
