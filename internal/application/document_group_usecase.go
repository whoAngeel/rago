package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
)

var (
	ErrGroupNotFound      = errors.New("group not found")
	ErrInvalidDocument    = errors.New("invalid document: must be completed and owned by user")
	ErrSlugCollision      = errors.New("slug collision after max retries")
)

type DocumentGroupUsecase struct {
	GroupRepo ports.DocumentGroupRepository
	DocRepo   ports.DocumentRepository
}

func NewDocumentGroupUsecase(
	groupRepo ports.DocumentGroupRepository,
	docRepo ports.DocumentRepository,
) *DocumentGroupUsecase {
	return &DocumentGroupUsecase{GroupRepo: groupRepo, DocRepo: docRepo}
}

func (uc *DocumentGroupUsecase) Create(ctx context.Context, userID int, name string, documentIDs []int, allowDownloads bool) (*domain.DocumentGroup, error) {
	if err := uc.validateDocuments(ctx, userID, documentIDs); err != nil {
		return nil, err
	}

	slug, err := uc.generateUniqueSlug(ctx)
	if err != nil {
		return nil, err
	}

	group := &domain.DocumentGroup{
		UserID:         userID,
		Name:           name,
		Slug:           slug,
		IsActive:       true,
		AllowDownloads: allowDownloads,
	}
	if err := uc.GroupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	if len(documentIDs) > 0 {
		if err := uc.GroupRepo.AddDocuments(ctx, group.ID, documentIDs); err != nil {
			return nil, err
		}
		group.DocumentIDs = documentIDs
	}

	return group, nil
}

func (uc *DocumentGroupUsecase) List(ctx context.Context, userID, page, limit int) ([]*domain.DocumentGroup, int64, error) {
	return uc.GroupRepo.FindByUserID(ctx, userID, page, limit)
}

func (uc *DocumentGroupUsecase) Get(ctx context.Context, id, userID int) (*domain.DocumentGroup, error) {
	group, err := uc.GroupRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}
	return group, nil
}

func (uc *DocumentGroupUsecase) Update(ctx context.Context, id, userID int, name string, isActive, allowDownloads bool) error {
	group, err := uc.GroupRepo.FindByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	group.Name = name
	group.IsActive = isActive
	group.AllowDownloads = allowDownloads
	return uc.GroupRepo.Update(ctx, group)
}

func (uc *DocumentGroupUsecase) Delete(ctx context.Context, id, userID int) error {
	group, err := uc.GroupRepo.FindByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	return uc.GroupRepo.Delete(ctx, id, userID)
}

func (uc *DocumentGroupUsecase) AddDocuments(ctx context.Context, groupID, userID int, documentIDs []int) error {
	group, err := uc.GroupRepo.FindByID(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	if err := uc.validateDocuments(ctx, userID, documentIDs); err != nil {
		return err
	}
	return uc.GroupRepo.AddDocuments(ctx, groupID, documentIDs)
}

func (uc *DocumentGroupUsecase) RemoveDocument(ctx context.Context, groupID, documentID, userID int) error {
	group, err := uc.GroupRepo.FindByID(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}
	if err := uc.GroupRepo.RemoveDocument(ctx, groupID, documentID); err != nil {
		return err
	}
	// Auto-deactivate si el grupo queda sin documentos
	remaining, err := uc.GroupRepo.FindDocumentIDs(ctx, groupID)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		group.IsActive = false
		return uc.GroupRepo.Update(ctx, group)
	}
	return nil
}

func (uc *DocumentGroupUsecase) validateDocuments(ctx context.Context, userID int, documentIDs []int) error {
	for _, docID := range documentIDs {
		doc, err := uc.DocRepo.FindByID(ctx, docID)
		if err != nil {
			return fmt.Errorf("%w: document %d", ErrInvalidDocument, docID)
		}
		if doc.UserID != userID || doc.Status != domain.StatusCompleted {
			return fmt.Errorf("%w: document %d", ErrInvalidDocument, docID)
		}
	}
	return nil
}

func (uc *DocumentGroupUsecase) generateUniqueSlug(ctx context.Context) (string, error) {
	for range 5 {
		slug, err := domain.GenerateSlug()
		if err != nil {
			return "", err
		}
		exists, err := uc.GroupRepo.SlugExists(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
	}
	return "", ErrSlugCollision
}
