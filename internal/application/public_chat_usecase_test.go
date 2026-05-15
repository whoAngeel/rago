package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tmc/langchaingo/schema"
	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

// ── Mocks ─────────────────────────────────────────────────────────────────

type mockGroupRepo struct {
	group                *domain.DocumentGroup
	findErr              error
	documentIDs          []int
	incrementErr         error
	slugExistsVal        bool
	incrementQuotaCalled bool
}

func (m *mockGroupRepo) FindBySlug(_ context.Context, _ string) (*domain.DocumentGroup, error) {
	return m.group, m.findErr
}
func (m *mockGroupRepo) FindDocumentIDs(_ context.Context, _ int) ([]int, error) {
	return m.documentIDs, nil
}
func (m *mockGroupRepo) IncrementAttempts(_ context.Context, _ int) error { return m.incrementErr }
func (m *mockGroupRepo) IncrementQuotaUsed(_ context.Context, _ int) error {
	m.incrementQuotaCalled = true
	return nil
}
func (m *mockGroupRepo) Create(_ context.Context, _ *domain.DocumentGroup) error { return nil }
func (m *mockGroupRepo) FindByID(_ context.Context, _, _ int) (*domain.DocumentGroup, error) {
	return m.group, m.findErr
}
func (m *mockGroupRepo) FindByUserID(_ context.Context, _, _, _ int) ([]*domain.DocumentGroup, int64, error) {
	return nil, 0, nil
}
func (m *mockGroupRepo) Update(_ context.Context, _ *domain.DocumentGroup) error { return nil }
func (m *mockGroupRepo) Delete(_ context.Context, _, _ int) error                { return nil }
func (m *mockGroupRepo) AddDocuments(_ context.Context, _ int, _ []int) error   { return nil }
func (m *mockGroupRepo) RemoveDocument(_ context.Context, _, _ int) error        { return nil }
func (m *mockGroupRepo) SlugExists(_ context.Context, _ string) (bool, error) {
	return m.slugExistsVal, nil
}

type mockUserRepoPublic struct {
	user            *domain.User
	findErr         error
	incrementCalled bool
}

func (m *mockUserRepoPublic) Create(_ context.Context, u *domain.User) (*domain.User, error) {
	return u, nil
}
func (m *mockUserRepoPublic) FindByEmail(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.findErr
}
func (m *mockUserRepoPublic) FindById(_ context.Context, _ int) (*domain.User, error) {
	return m.user, m.findErr
}
func (m *mockUserRepoPublic) IncrementChatQuotaUsed(_ context.Context, _ int) error {
	m.incrementCalled = true
	return nil
}

type mockDocRepo struct {
	doc    *domain.Document
	docErr error
}

func (m *mockDocRepo) FindByID(_ context.Context, _ int) (*domain.Document, error) {
	return m.doc, m.docErr
}
func (m *mockDocRepo) CreateDocument(_ context.Context, d *domain.Document) (*domain.Document, error) {
	return d, nil
}
func (m *mockDocRepo) UpdateDocument(_ context.Context, d *domain.Document) (*domain.Document, error) {
	return d, nil
}
func (m *mockDocRepo) FindDocumentByUserID(_ context.Context, _, _, _ int) ([]*domain.Document, int64, error) {
	return nil, 0, nil
}
func (m *mockDocRepo) UpdateDocumentStatus(_ context.Context, _ int, _ domain.DocumentStatus) error {
	return nil
}
func (m *mockDocRepo) DeleteDocument(_ context.Context, _ int) error { return nil }
func (m *mockDocRepo) FindPendingDocuments(_ context.Context, _ int) ([]*domain.Document, error) {
	return nil, nil
}
func (m *mockDocRepo) CreateProcessingStep(_ context.Context, _ *domain.ProcessingStep) error {
	return nil
}
func (m *mockDocRepo) UpdateProcessingStep(_ context.Context, _, _ int, _, _ string) error {
	return nil
}
func (m *mockDocRepo) FindStepsByDocumentID(_ context.Context, _ int) ([]*domain.ProcessingStep, error) {
	return nil, nil
}
func (m *mockDocRepo) FindByChecksum(_ context.Context, _ int, _ string) (*domain.Document, error) {
	return nil, nil
}
func (m *mockDocRepo) CountDocumentsByUserID(_ context.Context, _ int) (int64, error) {
	return 0, nil
}
func (m *mockDocRepo) FindDocumentsForSelect(_ context.Context, _ int) ([]*domain.Document, error) {
	return nil, nil
}
func (m *mockDocRepo) DeleteProcessingStepsByDocumentID(_ context.Context, _ int) error {
	return nil
}
func (m *mockDocRepo) CountByStatus(_ context.Context, _ int, _ domain.DocumentStatus) (int64, error) {
	return 0, nil
}
func (m *mockDocRepo) SumSizeByUserID(_ context.Context, _ int) (int64, error) {
	return 0, nil
}

type mockVectorStore struct {
	results      []ports.SearchResult
	searchErr    error
	capturedArgs struct {
		userID      int
		documentIDs []int
		limit       int
	}
}

func (m *mockVectorStore) Search(_ context.Context, _ string, _ []float32, userID int, documentIDs []int, limit int) ([]ports.SearchResult, error) {
	m.capturedArgs.userID = userID
	m.capturedArgs.documentIDs = documentIDs
	m.capturedArgs.limit = limit
	return m.results, m.searchErr
}
func (m *mockVectorStore) CreateCollection(_ context.Context, _ string, _ int) error { return nil }
func (m *mockVectorStore) UpsertDocuments(_ context.Context, _ string, _ []schema.Document, _ [][]float32) error {
	return nil
}
func (m *mockVectorStore) GetPointsCount(_ context.Context, _ string) (uint64, error) { return 0, nil }
func (m *mockVectorStore) DeleteCollection(_ context.Context, _ string) error          { return nil }

type mockEmbedder struct{ vec []float32 }

func (m *mockEmbedder) EmbedText(_ context.Context, _ string) ([]float32, error) {
	return m.vec, nil
}
func (m *mockEmbedder) ComputeEmbeddings(_ context.Context, _ []string) ([][]float32, error) {
	return [][]float32{m.vec}, nil
}

type mockLLM struct{ answer string }

func (m *mockLLM) GenerateAnswer(_ context.Context, _ string) (ports.LLMResult, error) {
	return ports.LLMResult{Answer: m.answer, Model: "test-model", InputTokens: 10, OutputTokens: 5}, nil
}
func (m *mockLLM) Stream(_ context.Context, _ string, _ func(string) error) (ports.LLMResult, error) {
	return ports.LLMResult{Answer: m.answer, Model: "test-model", InputTokens: 10, OutputTokens: 5}, nil
}

type mockConfigRepo struct{ val string }

func (m *mockConfigRepo) Get(_ context.Context, _ string) (string, error) { return m.val, nil }
func (m *mockConfigRepo) Set(_ context.Context, _, _ string) error         { return nil }

type mockLogger struct{}

func (m *mockLogger) Info(_ string, _ ...any)    {}
func (m *mockLogger) Warn(_ string, _ ...any)    {}
func (m *mockLogger) Debug(_ string, _ ...any)   {}
func (m *mockLogger) Error(_ string, _ ...any)   {}
func (m *mockLogger) Fatal(_ string, _ ...any)   {}
func (m *mockLogger) With(_ ...any) ports.Logger { return m }

type mockLLMUsageRepo struct {
	created []*domain.LLMUsage
}

func (m *mockLLMUsageRepo) Create(_ context.Context, u *domain.LLMUsage) error {
	m.created = append(m.created, u)
	return nil
}
func (m *mockLLMUsageRepo) SumByGroupID(_ context.Context, _ int) (ports.GroupUsageSummary, error) {
	return ports.GroupUsageSummary{}, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────

func unlimitedUser() *mockUserRepoPublic {
	return &mockUserRepoPublic{
		user: &domain.User{ID: 1, ChatQuota: 0, ChatQuotaUsed: 0},
	}
}

func newTestUsecase(
	groupRepo ports.DocumentGroupRepository,
	userRepo ports.UserRepository,
	docRepo ports.DocumentRepository,
	vs ports.VectorStore,
	llm ports.LLMProvider,
) *application.PublicChatUsecase {
	return application.NewPublicChatUsecase(
		groupRepo,
		userRepo,
		docRepo,
		vs,
		&mockEmbedder{vec: []float32{0.1, 0.2}},
		llm,
		nil,
		&mockConfigRepo{val: "Eres un asistente."},
		nil, // usageRepo: nil es válido, se guarda con guard
		&mockLogger{},
		config.Config{QdrantCollection: "test"},
	)
}

// ── Tests ──────────────────────────────────────────────────────────────────

func TestPublicChat_GroupNotFound(t *testing.T) {
	uc := newTestUsecase(&mockGroupRepo{group: nil}, unlimitedUser(), &mockDocRepo{}, &mockVectorStore{}, &mockLLM{})

	_, err := uc.Chat(context.Background(), "noexiste", "hola")

	if !errors.Is(err, application.ErrGroupInactive) {
		t.Errorf("esperaba ErrGroupInactive, got: %v", err)
	}
}

func TestPublicChat_GroupInactive(t *testing.T) {
	group := &domain.DocumentGroup{ID: 1, Slug: "abc", IsActive: false, DocumentIDs: []int{1}}
	uc := newTestUsecase(&mockGroupRepo{group: group}, unlimitedUser(), &mockDocRepo{}, &mockVectorStore{}, &mockLLM{})

	_, err := uc.Chat(context.Background(), "abc", "hola")

	if !errors.Is(err, application.ErrGroupInactive) {
		t.Errorf("esperaba ErrGroupInactive, got: %v", err)
	}
}

func TestPublicChat_UserQuotaExceeded(t *testing.T) {
	group := &domain.DocumentGroup{
		ID: 1, Slug: "abc", IsActive: true, UserID: 1,
		ChatQuota: 0, ChatQuotaUsed: 0,
		DocumentIDs: []int{1},
	}
	userRepo := &mockUserRepoPublic{
		user: &domain.User{ID: 1, ChatQuota: 10, ChatQuotaUsed: 10},
	}
	uc := newTestUsecase(&mockGroupRepo{group: group}, userRepo, &mockDocRepo{}, &mockVectorStore{}, &mockLLM{})

	_, err := uc.Chat(context.Background(), "abc", "hola")

	if !errors.Is(err, application.ErrQuotaExceeded) {
		t.Errorf("esperaba ErrQuotaExceeded, got: %v", err)
	}
}

func TestPublicChat_GroupCapExceeded(t *testing.T) {
	group := &domain.DocumentGroup{
		ID: 1, Slug: "abc", IsActive: true, UserID: 1,
		ChatQuota: 5, ChatQuotaUsed: 5,
		DocumentIDs: []int{1},
	}
	userRepo := &mockUserRepoPublic{
		user: &domain.User{ID: 1, ChatQuota: 500, ChatQuotaUsed: 10},
	}
	uc := newTestUsecase(&mockGroupRepo{group: group}, userRepo, &mockDocRepo{}, &mockVectorStore{}, &mockLLM{})

	_, err := uc.Chat(context.Background(), "abc", "hola")

	if !errors.Is(err, application.ErrQuotaExceeded) {
		t.Errorf("esperaba ErrQuotaExceeded por cap de grupo, got: %v", err)
	}
}

func TestPublicChat_AdminUnlimited(t *testing.T) {
	group := &domain.DocumentGroup{
		ID: 1, Slug: "abc", IsActive: true, UserID: 1,
		ChatQuota: 0, ChatQuotaUsed: 0,
		DocumentIDs: []int{1},
	}
	userRepo := &mockUserRepoPublic{
		user: &domain.User{ID: 1, ChatQuota: 0, ChatQuotaUsed: 9999},
	}
	vs := &mockVectorStore{results: []ports.SearchResult{
		{Document: schema.Document{PageContent: "contenido"}},
	}}
	uc := newTestUsecase(&mockGroupRepo{group: group}, userRepo, &mockDocRepo{}, vs, &mockLLM{answer: "respuesta"})

	answer, err := uc.Chat(context.Background(), "abc", "hola")

	if err != nil {
		t.Fatalf("admin no debería ser bloqueado, got: %v", err)
	}
	if answer != "respuesta" {
		t.Errorf("esperaba 'respuesta', got: %q", answer)
	}
}

func TestPublicChat_IncrementsBothQuotas(t *testing.T) {
	group := &domain.DocumentGroup{
		ID: 1, Slug: "abc", IsActive: true, UserID: 1,
		ChatQuota: 0, ChatQuotaUsed: 0,
		DocumentIDs: []int{1},
	}
	groupRepo := &mockGroupRepo{group: group}
	userRepo := &mockUserRepoPublic{
		user: &domain.User{ID: 1, ChatQuota: 100, ChatQuotaUsed: 0},
	}
	vs := &mockVectorStore{results: []ports.SearchResult{
		{Document: schema.Document{PageContent: "contenido relevante"}},
	}}
	uc := newTestUsecase(groupRepo, userRepo, &mockDocRepo{}, vs, &mockLLM{answer: "ok"})

	_, err := uc.Chat(context.Background(), "abc", "pregunta")

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if !groupRepo.incrementQuotaCalled {
		t.Error("debería haber incrementado la cuota del grupo")
	}
	if !userRepo.incrementCalled {
		t.Error("debería haber incrementado la cuota del usuario")
	}
}

func TestPublicChat_RecordsTokenUsage(t *testing.T) {
	group := &domain.DocumentGroup{
		ID: 5, Slug: "abc", IsActive: true, UserID: 42,
		ChatQuota: 0, ChatQuotaUsed: 0,
		DocumentIDs: []int{1},
	}
	userRepo := &mockUserRepoPublic{
		user: &domain.User{ID: 42, ChatQuota: 0},
	}
	vs := &mockVectorStore{results: []ports.SearchResult{
		{Document: schema.Document{PageContent: "contenido"}},
	}}
	usageRepo := &mockLLMUsageRepo{}

	uc := application.NewPublicChatUsecase(
		&mockGroupRepo{group: group},
		userRepo,
		&mockDocRepo{},
		vs,
		&mockEmbedder{vec: []float32{0.1}},
		&mockLLM{answer: "respuesta"},
		nil,
		&mockConfigRepo{val: "Eres un asistente."},
		usageRepo,
		&mockLogger{},
		config.Config{QdrantCollection: "test"},
	)

	_, err := uc.Chat(context.Background(), "abc", "pregunta")

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if len(usageRepo.created) != 1 {
		t.Fatalf("esperaba 1 registro de uso, got: %d", len(usageRepo.created))
	}
	rec := usageRepo.created[0]
	if rec.UserID != 42 {
		t.Errorf("user_id: esperaba 42, got %d", rec.UserID)
	}
	if rec.GroupID == nil || *rec.GroupID != 5 {
		t.Errorf("group_id: esperaba 5, got %v", rec.GroupID)
	}
	if rec.Model != "test-model" {
		t.Errorf("model: esperaba 'test-model', got %q", rec.Model)
	}
}

func TestPublicChat_PassesDocumentIDsToSearch(t *testing.T) {
	wantDocIDs := []int{3, 7, 12}
	group := &domain.DocumentGroup{
		ID: 1, Slug: "abc", IsActive: true, UserID: 1,
		ChatQuota: 100, ChatQuotaUsed: 0,
		DocumentIDs: wantDocIDs,
	}
	vs := &mockVectorStore{
		results: []ports.SearchResult{
			{Document: schema.Document{PageContent: "contenido relevante"}},
		},
	}
	uc := newTestUsecase(&mockGroupRepo{group: group}, unlimitedUser(), &mockDocRepo{}, vs, &mockLLM{answer: "respuesta"})

	answer, err := uc.Chat(context.Background(), "abc", "pregunta")

	if err != nil {
		t.Fatalf("no esperaba error, got: %v", err)
	}
	if answer != "respuesta" {
		t.Errorf("esperaba 'respuesta', got: %q", answer)
	}
	if vs.capturedArgs.userID != 0 {
		t.Errorf("userID debería ser 0 para chat público, got: %d", vs.capturedArgs.userID)
	}
	if len(vs.capturedArgs.documentIDs) != len(wantDocIDs) {
		t.Fatalf("esperaba %d documentIDs, got: %d", len(wantDocIDs), len(vs.capturedArgs.documentIDs))
	}
	for i, id := range wantDocIDs {
		if vs.capturedArgs.documentIDs[i] != id {
			t.Errorf("documentID[%d]: esperaba %d, got %d", i, id, vs.capturedArgs.documentIDs[i])
		}
	}
}

func TestPublicChat_EmptySearchResultsCallsLLM(t *testing.T) {
	group := &domain.DocumentGroup{
		ID: 1, Slug: "abc", IsActive: true, UserID: 1,
		ChatQuota: 100, ChatQuotaUsed: 0,
		DocumentIDs: []int{1},
	}
	vs := &mockVectorStore{results: []ports.SearchResult{}}
	llm := &mockLLM{answer: "No tengo información suficiente en tus documentos para responder a esto."}
	uc := newTestUsecase(&mockGroupRepo{group: group}, unlimitedUser(), &mockDocRepo{}, vs, llm)

	answer, err := uc.Chat(context.Background(), "abc", "pregunta")

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if answer == "" {
		t.Error("el LLM debería retornar respuesta incluso sin resultados de búsqueda")
	}
}

func TestPublicChat_SearchLimitIs5(t *testing.T) {
	group := &domain.DocumentGroup{
		ID: 1, Slug: "abc", IsActive: true, UserID: 1,
		ChatQuota: 100, ChatQuotaUsed: 0,
		DocumentIDs: []int{1},
	}
	vs := &mockVectorStore{}
	uc := newTestUsecase(&mockGroupRepo{group: group}, unlimitedUser(), &mockDocRepo{}, vs, &mockLLM{answer: "ok"})

	_, _ = uc.Chat(context.Background(), "abc", "pregunta")

	if vs.capturedArgs.limit != 5 {
		t.Errorf("límite de búsqueda debería ser 5, got: %d", vs.capturedArgs.limit)
	}
}

func TestGetGroupInfo_InactiveGroup(t *testing.T) {
	group := &domain.DocumentGroup{ID: 1, Slug: "abc", IsActive: false}
	uc := newTestUsecase(&mockGroupRepo{group: group}, unlimitedUser(), &mockDocRepo{}, &mockVectorStore{}, &mockLLM{})

	_, err := uc.GetGroupInfo(context.Background(), "abc")

	if !errors.Is(err, application.ErrGroupInactive) {
		t.Errorf("esperaba ErrGroupInactive, got: %v", err)
	}
}

func TestGetGroupInfo_ActiveGroup(t *testing.T) {
	doc := &domain.Document{ID: 5, Filename: "archivo.pdf", UserID: 1}
	group := &domain.DocumentGroup{
		ID: 1, Slug: "abc12345", IsActive: true, AllowDownloads: true,
		Name:        "Mi grupo",
		DocumentIDs: []int{5},
	}
	uc := newTestUsecase(&mockGroupRepo{group: group}, unlimitedUser(), &mockDocRepo{doc: doc}, &mockVectorStore{}, &mockLLM{})

	info, err := uc.GetGroupInfo(context.Background(), "abc12345")

	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if info.Name != "Mi grupo" {
		t.Errorf("esperaba 'Mi grupo', got: %q", info.Name)
	}
	if len(info.Documents) != 1 {
		t.Fatalf("esperaba 1 documento, got: %d", len(info.Documents))
	}
	if info.Documents[0].Filename != "archivo.pdf" {
		t.Errorf("esperaba 'archivo.pdf', got: %q", info.Documents[0].Filename)
	}
}
