package parser

import (
	"strings"

	"github.com/whoAngeel/rago/internal/core/ports"
)

// Version identifies the extraction and cleaning logic; it is part of the parse
// cache key. Change it whenever extraction or cleaning output changes (for
// example edits to PDF text extraction, page cleaning or the OCR prompt).
const Version = "2"

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
func NewDefaultRegistry(ocr ports.OCR, pdfOpts ...PDFOption) *Registry {
	r := NewRegistry()
	r.Register("text/plain", NewPlainTextAdapter())
	r.Register("text/markdown", NewPlainTextAdapter())
	r.Register("application/pdf", NewPDFParser(ocr, pdfOpts...))
	return r
}
