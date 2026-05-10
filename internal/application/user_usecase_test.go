package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
)

// mockUserRepo reutiliza los mismos mocks de public_chat_usecase_test.go
// pero necesitamos uno propio para controlar FindById

type stubUserRepo struct {
	user    *domain.User
	findErr error
}

func (m *stubUserRepo) Create(_ context.Context, u *domain.User) (*domain.User, error) {
	return u, nil
}
func (m *stubUserRepo) FindByEmail(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.findErr
}
func (m *stubUserRepo) FindById(_ context.Context, _ int) (*domain.User, error) {
	return m.user, m.findErr
}
func (m *stubUserRepo) IncrementChatQuotaUsed(_ context.Context, _ int) error { return nil }

type stubDocRepo struct {
	count    int64
	countErr error
}

func (m *stubDocRepo) FindByID(_ context.Context, _ int) (*domain.Document, error)       { return nil, nil }
func (m *stubDocRepo) CreateDocument(_ context.Context, d *domain.Document) (*domain.Document, error) {
	return d, nil
}
func (m *stubDocRepo) UpdateDocument(_ context.Context, d *domain.Document) (*domain.Document, error) {
	return d, nil
}
func (m *stubDocRepo) FindDocumentByUserID(_ context.Context, _, _, _ int) ([]*domain.Document, int64, error) {
	return nil, 0, nil
}
func (m *stubDocRepo) UpdateDocumentStatus(_ context.Context, _ int, _ domain.DocumentStatus) error {
	return nil
}
func (m *stubDocRepo) DeleteDocument(_ context.Context, _ int) error      { return nil }
func (m *stubDocRepo) FindPendingDocuments(_ context.Context, _ int) ([]*domain.Document, error) {
	return nil, nil
}
func (m *stubDocRepo) CreateProcessingStep(_ context.Context, _ *domain.ProcessingStep) error {
	return nil
}
func (m *stubDocRepo) UpdateProcessingStep(_ context.Context, _, _ int, _, _ string) error {
	return nil
}
func (m *stubDocRepo) FindStepsByDocumentID(_ context.Context, _ int) ([]*domain.ProcessingStep, error) {
	return nil, nil
}
func (m *stubDocRepo) FindByChecksum(_ context.Context, _ int, _ string) (*domain.Document, error) {
	return nil, nil
}
func (m *stubDocRepo) CountDocumentsByUserID(_ context.Context, _ int) (int64, error) {
	return m.count, m.countErr
}
func (m *stubDocRepo) FindDocumentsForSelect(_ context.Context, _ int) ([]*domain.Document, error) {
	return nil, nil
}

// ── Tests ──────────────────────────────────────────────────────────────────

func TestGetProfile_ReturnsProfileWithCount(t *testing.T) {
	user := &domain.User{ID: 1, Name: "Juan", Email: "juan@test.com", RoleID: 3, MaxDocuments: 10}
	uc := application.NewUserUsecase(&stubUserRepo{user: user}, &stubDocRepo{count: 7})

	profile, err := uc.GetProfile(context.Background(), 1)

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if profile.MaxDocuments != 10 {
		t.Errorf("max_documents: esperaba 10, got %d", profile.MaxDocuments)
	}
	if profile.DocumentCount != 7 {
		t.Errorf("document_count: esperaba 7, got %d", profile.DocumentCount)
	}
	if profile.Role != "viewer" {
		t.Errorf("role: esperaba 'viewer', got %q", profile.Role)
	}
}

func TestGetProfile_AdminUnlimited(t *testing.T) {
	user := &domain.User{ID: 2, Name: "Admin", Email: "admin@test.com", RoleID: 1, MaxDocuments: 0}
	uc := application.NewUserUsecase(&stubUserRepo{user: user}, &stubDocRepo{count: 42})

	profile, err := uc.GetProfile(context.Background(), 2)

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if profile.MaxDocuments != 0 {
		t.Errorf("admin debería tener max_documents=0 (ilimitado), got %d", profile.MaxDocuments)
	}
	if profile.DocumentCount != 42 {
		t.Errorf("document_count: esperaba 42, got %d", profile.DocumentCount)
	}
	if profile.Role != "admin" {
		t.Errorf("role: esperaba 'admin', got %q", profile.Role)
	}
}

func TestGetProfile_UserNotFound(t *testing.T) {
	uc := application.NewUserUsecase(
		&stubUserRepo{findErr: errors.New("not found")},
		&stubDocRepo{},
	)

	_, err := uc.GetProfile(context.Background(), 99)

	if err == nil {
		t.Fatal("esperaba error cuando el usuario no existe")
	}
}

func TestGetProfile_CountError(t *testing.T) {
	user := &domain.User{ID: 1, RoleID: 3, MaxDocuments: 10}
	uc := application.NewUserUsecase(
		&stubUserRepo{user: user},
		&stubDocRepo{countErr: errors.New("db error")},
	)

	_, err := uc.GetProfile(context.Background(), 1)

	if err == nil {
		t.Fatal("esperaba error cuando falla el conteo")
	}
}
