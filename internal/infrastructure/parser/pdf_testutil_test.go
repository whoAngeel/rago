package parser

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// requireTools skips the test when any of the binaries is missing.
func requireTools(t *testing.T, bins ...string) {
	t.Helper()
	for _, b := range bins {
		if _, err := exec.LookPath(b); err != nil {
			t.Skipf("%s not on PATH", b)
		}
	}
}

func pdfEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, "(", `\(`, ")", `\)`)
	return r.Replace(s)
}

// buildPDF returns a minimal valid PDF with one Helvetica text line per page;
// an empty string yields a page with an empty content stream (a stand-in for
// a scanned page).
func buildPDF(fontSize int, pages ...string) []byte {
	var objs []string
	pageRefs := make([]string, len(pages))
	for i := range pages {
		pageRefs[i] = fmt.Sprintf("%d 0 R", 4+2*i)
	}
	objs = append(objs,
		"<< /Type /Catalog /Pages 2 0 R >>",
		fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(pageRefs, " "), len(pages)),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	)
	for i, text := range pages {
		stream := ""
		if text != "" {
			stream = fmt.Sprintf("BT /F1 %d Tf 72 700 Td (%s) Tj ET", fontSize, pdfEscape(text))
		}
		objs = append(objs,
			fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>", 5+2*i),
			fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
		)
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objs))
	for i, o := range objs {
		offsets[i] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, off := range offsets {
		fmt.Fprintf(&buf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, xref)
	return buf.Bytes()
}
