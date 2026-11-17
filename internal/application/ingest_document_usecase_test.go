package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
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
