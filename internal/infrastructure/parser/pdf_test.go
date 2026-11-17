package parser

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/tmc/langchaingo/schema"
)

type fakeOCR struct {
	mu    sync.Mutex
	calls [][]byte
	text  string
	err   error
}

func (f *fakeOCR) Recognize(_ context.Context, png []byte) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, png)
	return f.text, f.err
}

var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

const (
	pageOne   = "hold for 3 seconds at 455.1 MHz and then release the button"
	pageThree = "the display shows 3 440x1 440 pixels on the main monitor"
)

func parsePDF(t *testing.T, ctx context.Context, p *PDFParser, data []byte) ([]schema.Document, error) {
	t.Helper()
	return p.Parse(ctx, bytes.NewReader(data), "application/pdf")
}

func TestPDFParser_TextPDF(t *testing.T) {
	requireTools(t, "pdftotext", "pdftoppm")
	docs, err := parsePDF(t, context.Background(), NewPDFParser(nil), buildPDF(12, pageOne, pageThree))
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("got %d docs; want 2", len(docs))
	}
	if docs[0].PageContent != pageOne {
		t.Errorf("page 1 = %q; want %q", docs[0].PageContent, pageOne)
	}
	for i, d := range docs {
		if d.Metadata["source"] != "text" {
			t.Errorf("doc %d source = %v; want text", i, d.Metadata["source"])
		}
		if d.Metadata["page_number"] != i+1 {
			t.Errorf("doc %d page_number = %v; want %d", i, d.Metadata["page_number"], i+1)
		}
		if d.Metadata["page_count"] != 2 {
			t.Errorf("doc %d page_count = %v; want 2", i, d.Metadata["page_count"])
		}
	}
}

func TestPDFParser_MixedPDF(t *testing.T) {
	requireTools(t, "pdftotext", "pdftoppm")
	mixed := buildPDF(12, pageOne, "", pageThree)

	t.Run("blank page goes through OCR", func(t *testing.T) {
		ocr := &fakeOCR{text: "text recovered by the ocr engine"}
		docs, err := parsePDF(t, context.Background(), NewPDFParser(ocr), mixed)
		if err != nil {
			t.Fatal(err)
		}
		if len(ocr.calls) != 1 {
			t.Fatalf("ocr called %d times; want 1", len(ocr.calls))
		}
		if !bytes.HasPrefix(ocr.calls[0], pngSignature) {
			t.Errorf("ocr input does not start with the PNG signature")
		}
		if len(docs) != 3 {
			t.Fatalf("got %d docs; want 3", len(docs))
		}
		for i, d := range docs {
			if d.Metadata["page_number"] != i+1 {
				t.Errorf("doc %d page_number = %v; want %d", i, d.Metadata["page_number"], i+1)
			}
			if d.Metadata["page_count"] != 3 {
				t.Errorf("doc %d page_count = %v; want 3", i, d.Metadata["page_count"])
			}
		}
		if docs[1].Metadata["source"] != "ocr" || docs[1].PageContent != ocr.text {
			t.Errorf("page 2 = %q (%v); want ocr text", docs[1].PageContent, docs[1].Metadata["source"])
		}
		if docs[0].Metadata["source"] != "text" || docs[2].Metadata["source"] != "text" {
			t.Errorf("pages 1 and 3 should have source text")
		}
	})
}

func TestPDFParser_AlwaysOCRUsesEngineOnTextPages(t *testing.T) {
	requireTools(t, "pdftotext", "pdftoppm")
	always := &fakeOCR{text: "text transcribed by the always engine"}
	p := NewPDFParser(nil, WithAlwaysOCR(always))
	docs, err := parsePDF(t, context.Background(), p, buildPDF(12, pageOne, pageThree))
	if err != nil {
		t.Fatal(err)
	}
	if len(always.calls) != 2 {
		t.Fatalf("always engine called %d times; want 2", len(always.calls))
	}
	if len(docs) != 2 {
		t.Fatalf("got %d docs; want 2", len(docs))
	}
	for i, d := range docs {
		if d.PageContent != always.text {
			t.Errorf("doc %d = %q; want %q", i, d.PageContent, always.text)
		}
		if d.Metadata["source"] != "ocr" {
			t.Errorf("doc %d source = %v; want ocr", i, d.Metadata["source"])
		}
		if d.Metadata["page_number"] != i+1 {
			t.Errorf("doc %d page_number = %v; want %d", i, d.Metadata["page_number"], i+1)
		}
	}
}

func TestPDFParser_AlwaysOCRErrorKeepsTextLayer(t *testing.T) {
	requireTools(t, "pdftotext", "pdftoppm")
	always := &fakeOCR{err: errors.New("provider down")}
	p := NewPDFParser(nil, WithAlwaysOCR(always))
	docs, err := parsePDF(t, context.Background(), p, buildPDF(12, pageOne, pageThree))
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("got %d docs; want 2", len(docs))
	}
	want := []string{pageOne, pageThree}
	for i, d := range docs {
		if d.PageContent != want[i] {
			t.Errorf("doc %d = %q; want %q", i, d.PageContent, want[i])
		}
		if d.Metadata["source"] != "text" {
			t.Errorf("doc %d source = %v; want text", i, d.Metadata["source"])
		}
	}
}
