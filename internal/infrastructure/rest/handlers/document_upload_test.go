package handlers

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

type fakeUploader struct {
	calls       int
	filename    string
	contentType string
	body        []byte
}

func (f *fakeUploader) Upload(_ context.Context, _ int, filename string, r io.Reader, _ int64, contentType string) (*domain.Document, error) {
	f.calls++
	f.filename = filename
	f.contentType = contentType
	f.body, _ = io.ReadAll(r)
	return &domain.Document{ID: 1, Status: domain.DocumentStatus("pending")}, nil
}

type nopLogger struct{}

func (nopLogger) Debug(string, ...any)       {}
func (nopLogger) Info(string, ...any)        {}
func (nopLogger) Warn(string, ...any)        {}
func (nopLogger) Error(string, ...any)       {}
func (nopLogger) Fatal(string, ...any)       {}
func (n nopLogger) With(...any) ports.Logger { return n }

func postUpload(t *testing.T, filename, headerContentType string, body []byte) (*httptest.ResponseRecorder, *fakeUploader) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", headerContentType)
	part, err := mw.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(body)
	mw.Close()

	fake := &fakeUploader{}
	handler := &DocumentHandler{uploader: fake, logger: nopLogger{}, config: config.Config{MaxUploadSize: 1 << 20}}
	r := gin.New()
	r.POST("/documents", handler.Upload)

	req := httptest.NewRequest(http.MethodPost, "/documents", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w, fake
}

func TestUpload_RejectsUnsupportedExtension(t *testing.T) {
	w, fake := postUpload(t, "sheet.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", []byte("PK"))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	if fake.calls != 0 {
		t.Errorf("usecase called %d times, want 0", fake.calls)
	}
}

func TestUpload_RejectsPDFWithoutMagicBytes(t *testing.T) {
	for name, body := range map[string][]byte{
		"wrong bytes": []byte("<html>not a pdf</html>"),
		"too short":   []byte("%PDF"),
		"empty":       nil,
	} {
		t.Run(name, func(t *testing.T) {
			w, fake := postUpload(t, "doc.pdf", "application/pdf", body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", w.Code)
			}
			if fake.calls != 0 {
				t.Errorf("usecase called %d times, want 0", fake.calls)
			}
		})
	}
}

func TestUpload_PDFIgnoresBrowserHeaderAndKeepsAllBytes(t *testing.T) {
	body := []byte("%PDF-1.7\nrest of the file")
	w, fake := postUpload(t, "doc.pdf", "text/plain", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body %s", w.Code, w.Body.String())
	}
	if fake.contentType != "application/pdf" {
		t.Errorf("contentType = %q, want application/pdf", fake.contentType)
	}
	if !bytes.Equal(fake.body, body) {
		t.Errorf("usecase received %q, want full unmodified %q", fake.body, body)
	}
}

func TestUpload_MarkdownContentTypeFromExtension(t *testing.T) {
	w, fake := postUpload(t, "notes.md", "application/octet-stream", []byte("# hi"))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if fake.contentType != "text/markdown" {
		t.Errorf("contentType = %q, want text/markdown", fake.contentType)
	}
}

func TestUpload_TxtStillAccepted(t *testing.T) {
	w, fake := postUpload(t, "a.txt", "text/plain; charset=utf-8", []byte("hello"))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if fake.contentType != "text/plain" || string(fake.body) != "hello" {
		t.Errorf("got contentType=%q body=%q", fake.contentType, fake.body)
	}
}
