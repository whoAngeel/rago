package parser

import "strings"

// supportedExtensions is the single source of truth for the document formats
// the system accepts: file extension to content type.
var supportedExtensions = map[string]string{
	".txt": "text/plain",
	".md":  "text/markdown",
	".pdf": "application/pdf",
}

// ContentTypeForExtension returns the content type for a file extension
// (case-insensitive, with leading dot). ok is false for unsupported formats.
func ContentTypeForExtension(ext string) (string, bool) {
	ct, ok := supportedExtensions[strings.ToLower(ext)]
	return ct, ok
}

// NewDefaultRegistry returns a registry with a parser for every supported
// content type.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register("text/plain", NewPlainTextAdapter())
	r.Register("text/markdown", NewPlainTextAdapter())
	r.Register("application/pdf", NewPDFParser())
	return r
}
