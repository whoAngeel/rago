package application_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

func TestDeleteDocument_RejectsDocumentOfAnotherUser(t *testing.T) {
	repo := &mockDocRepo{doc: &domain.Document{UserID: 1, FilePath: "1/5/manual.pdf"}}
	// Storage and vector dependencies are nil on purpose: reaching them would
	// panic, which proves nothing is deleted before the owner check.
	uc := application.NewIngestDocumentUsecase(repo, nil, nil, nil)

	err := uc.DeleteDocument(context.Background(), 5, 2)

	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("DeleteDocument by non-owner = %v; want ErrNotFound", err)
	}
}

func TestDeleteDocument_MissingDocumentIsNotFound(t *testing.T) {
	repo := &mockDocRepo{docErr: errors.New("record not found")}
	uc := application.NewIngestDocumentUsecase(repo, nil, nil, nil)

	err := uc.DeleteDocument(context.Background(), 5, 2)

	if !errors.Is(err, application.ErrNotFound) {
		t.Fatalf("DeleteDocument of missing document = %v; want ErrNotFound", err)
	}
}

type fakeBlobs struct {
	exists  bool
	uploads int
	deleted []string
}

func (b *fakeBlobs) Upload(_ context.Context, name string, _ io.Reader) (string, error) {
	b.uploads++
	return name, nil
}
func (b *fakeBlobs) Download(_ context.Context, _ string) (io.ReadCloser, error) { return nil, nil }
func (b *fakeBlobs) Delete(_ context.Context, p string) error {
	b.deleted = append(b.deleted, p)
	return nil
}
func (b *fakeBlobs) Exists(_ context.Context, _ string) (bool, error) { return b.exists, nil }

func uploadWithExisting(t *testing.T, existing *domain.Document) (*domain.Document, error, *mockDocRepo, *fakeBlobs) {
	t.Helper()
	repo := &mockDocRepo{byChecksum: existing}
	blobs := &fakeBlobs{}
	users := &mockUserRepoPublic{user: &domain.User{}}
	uc := application.NewIngestDocumentUsecase(repo, users, blobs, nil)
	doc, err := uc.Upload(context.Background(), 1, "a.txt", strings.NewReader("data"), 4, "text/plain")
	return doc, err, repo, blobs
}

func TestUpload_DuplicateCompletedReturnsExistingWithError(t *testing.T) {
	sum := "abc"
	existing := &domain.Document{ID: 7, Status: domain.StatusCompleted, Checksum: &sum, FilePath: "blobs/abc"}

	doc, err, repo, blobs := uploadWithExisting(t, existing)

	if !errors.Is(err, application.ErrDuplicateDocument) {
		t.Fatalf("err = %v; want ErrDuplicateDocument", err)
	}
	if doc == nil || doc.ID != 7 {
		t.Errorf("doc = %v; want the existing document", doc)
	}
	if len(repo.requeued) != 0 || blobs.uploads != 0 {
		t.Errorf("requeued = %v, uploads = %d; want none", repo.requeued, blobs.uploads)
	}
}

func TestUpload_DuplicateFailedIsRequeuedAndMissingBlobRestored(t *testing.T) {
	sum := "abc"
	existing := &domain.Document{ID: 7, Status: domain.StatusFailed, RetryCount: 3, ErrorMessage: "boom", Checksum: &sum, FilePath: "blobs/abc"}

	doc, err, repo, blobs := uploadWithExisting(t, existing)

	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if doc.Status != domain.StatusPending || doc.RetryCount != 0 || doc.ErrorMessage != "" {
		t.Errorf("doc = %+v; want pending, retry 0, no error", doc)
	}
	if len(repo.requeued) != 1 || repo.requeued[0] != 7 {
		t.Errorf("requeued = %v; want [7]", repo.requeued)
	}
	// The fake storage reports the blob as missing, so it is stored again.
	if blobs.uploads != 1 {
		t.Errorf("uploads = %d; want 1", blobs.uploads)
	}
}

func TestDeleteDocument_KeepsBlobSharedWithAnotherDocument(t *testing.T) {
	repo := &mockDocRepo{doc: &domain.Document{ID: 5, UserID: 1, FilePath: "blobs/abc"}, othersByPath: 1}
	blobs := &fakeBlobs{}
	ingest := application.NewIngestUsecase(&rtVectorStore{}, nil, rtLogger{}, config.Config{})
	uc := application.NewIngestDocumentUsecase(repo, nil, blobs, ingest)

	if err := uc.DeleteDocument(context.Background(), 5, 1); err != nil {
		t.Fatal(err)
	}
	if len(blobs.deleted) != 0 {
		t.Errorf("deleted = %v; want the shared blob kept", blobs.deleted)
	}
}

func TestDeleteDocument_DeletesBlobWhenUnreferenced(t *testing.T) {
	repo := &mockDocRepo{doc: &domain.Document{ID: 5, UserID: 1, FilePath: "blobs/abc"}}
	blobs := &fakeBlobs{}
	ingest := application.NewIngestUsecase(&rtVectorStore{}, nil, rtLogger{}, config.Config{})
	uc := application.NewIngestDocumentUsecase(repo, nil, blobs, ingest)

	if err := uc.DeleteDocument(context.Background(), 5, 1); err != nil {
		t.Fatal(err)
	}
	if len(blobs.deleted) != 1 || blobs.deleted[0] != "blobs/abc" {
		t.Errorf("deleted = %v; want [blobs/abc]", blobs.deleted)
	}
}
