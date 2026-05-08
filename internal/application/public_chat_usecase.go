package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
)

var ErrQuotaExceeded = errors.New("quota exceeded")
var ErrGroupInactive = errors.New("group not found or inactive")

const publicSearchLimit = 5
const publicScoreThreshold float32 = 0.40

type PublicDocument struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
}

type GroupPublicInfo struct {
	Name           string           `json:"name"`
	Slug           string           `json:"slug"`
	AllowDownloads bool             `json:"allow_downloads"`
	Documents      []PublicDocument `json:"documents"`
}

type PublicChatUsecase struct {
	GroupRepo   ports.DocumentGroupRepository
	UserRepo    ports.UserRepository
	DocRepo     ports.DocumentRepository
	VectorStore ports.VectorStore
	Embedder    ports.Embedder
	LLM         ports.LLMProvider
	BlobStorage ports.BlobStorage
	ConfigRepo  ports.SystemConfigRepository
	Logger      ports.Logger
	config      config.Config
}

func NewPublicChatUsecase(
	groupRepo ports.DocumentGroupRepository,
	userRepo ports.UserRepository,
	docRepo ports.DocumentRepository,
	vectorStore ports.VectorStore,
	embedder ports.Embedder,
	llm ports.LLMProvider,
	blobStorage ports.BlobStorage,
	configRepo ports.SystemConfigRepository,
	logger ports.Logger,
	cfg config.Config,
) *PublicChatUsecase {
	return &PublicChatUsecase{
		GroupRepo:   groupRepo,
		UserRepo:    userRepo,
		DocRepo:     docRepo,
		VectorStore: vectorStore,
		Embedder:    embedder,
		LLM:         llm,
		BlobStorage: blobStorage,
		ConfigRepo:  configRepo,
		Logger:      logger,
		config:      cfg,
	}
}

func (uc *PublicChatUsecase) GetGroupInfo(ctx context.Context, slug string) (*GroupPublicInfo, error) {
	group, err := uc.GroupRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if group == nil || !group.IsActive {
		uc.Logger.Info("public: group not found or inactive", "slug", slug)
		return nil, ErrGroupInactive
	}

	docs := make([]PublicDocument, 0, len(group.DocumentIDs))
	for _, id := range group.DocumentIDs {
		doc, err := uc.DocRepo.FindByID(ctx, id)
		if err != nil {
			uc.Logger.Warn("public: doc not found while building group info", "doc_id", id, "error", err)
			continue
		}
		docs = append(docs, PublicDocument{ID: doc.ID, Filename: doc.Filename})
	}

	uc.Logger.Info("public: group info fetched", "slug", slug, "docs", len(docs))
	return &GroupPublicInfo{
		Name:           group.Name,
		Slug:           group.Slug,
		AllowDownloads: group.AllowDownloads,
		Documents:      docs,
	}, nil
}

func (uc *PublicChatUsecase) Chat(ctx context.Context, slug, message string) (string, error) {
	group, err := uc.GroupRepo.FindBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	if group == nil || !group.IsActive {
		uc.Logger.Info("public: chat blocked — group not found or inactive", "slug", slug)
		return "", ErrGroupInactive
	}

	// Contabilizar intento siempre, incluso si está bloqueado
	_ = uc.GroupRepo.IncrementAttempts(ctx, group.ID)

	user, err := uc.UserRepo.FindById(ctx, group.UserID)
	if err != nil {
		return "", fmt.Errorf("finding group owner: %w", err)
	}

	uc.Logger.Info("public: chat request",
		"slug", slug,
		"group_id", group.ID,
		"user_id", user.ID,
		"user_quota_used", user.ChatQuotaUsed,
		"user_quota_total", user.ChatQuota,
		"group_cap", group.ChatQuota,
		"group_cap_used", group.ChatQuotaUsed,
	)

	// Cuota global del usuario (0 = ilimitado)
	if user.ChatQuota > 0 && user.ChatQuotaUsed >= user.ChatQuota {
		uc.Logger.Info("public: chat blocked — user quota exceeded", "slug", slug, "user_id", user.ID)
		return "", ErrQuotaExceeded
	}

	// Cap por grupo (0 = sin cap)
	if group.ChatQuota > 0 && group.ChatQuotaUsed >= group.ChatQuota {
		uc.Logger.Info("public: chat blocked — group cap exceeded", "slug", slug, "group_id", group.ID)
		return "", ErrQuotaExceeded
	}

	queryVector, err := uc.Embedder.EmbedText(ctx, message)
	if err != nil {
		return "", fmt.Errorf("embedding: %w", err)
	}

	uc.Logger.Info("public: searching vectors", "collection", uc.config.QdrantCollection, "document_ids", group.DocumentIDs, "limit", publicSearchLimit)

	searchResults, err := uc.VectorStore.Search(ctx, uc.config.QdrantCollection, queryVector, 0, group.DocumentIDs, publicSearchLimit)
	if err != nil {
		return "", fmt.Errorf("searching: %w", err)
	}

	uc.Logger.Info("public: search completed", "results", len(searchResults))
	for i, r := range searchResults {
		preview := r.Document.PageContent
		if len(preview) > 120 {
			preview = preview[:120]
		}
		uc.Logger.Debug("public: result chunk", "idx", i, "score", r.Score, "content_len", len(r.Document.PageContent), "preview", preview)
	}

	filtered := searchResults[:0]
	for _, r := range searchResults {
		if r.Score >= publicScoreThreshold {
			filtered = append(filtered, r)
		}
	}
	uc.Logger.Info("public: after threshold filter", "threshold", publicScoreThreshold, "before", len(searchResults), "after", len(filtered))

	lowConfidence := false
	if len(filtered) == 0 && len(searchResults) > 0 {
		top := 2
		if len(searchResults) < top {
			top = len(searchResults)
		}
		filtered = searchResults[:top]
		lowConfidence = true
		uc.Logger.Warn("public: low confidence fallback", "using_top", top, "best_score", searchResults[0].Score)
	}
	searchResults = filtered

	systemPrompt, err := uc.ConfigRepo.Get(ctx, "system_prompt")
	if err != nil {
		systemPrompt = fallbackSystemPrompt
	}

	var contextStr string
	for _, r := range searchResults {
		src := "desconocido"
		if s, ok := r.Document.Metadata["source"]; ok {
			src = fmt.Sprintf("%v", s)
		}
		contextStr += fmt.Sprintf("[Fuente: %s]\n%s\n\n", src, r.Document.PageContent)
	}

	var sb strings.Builder
	sb.WriteString(systemPrompt)
	sb.WriteString("\n\n")
	if contextStr == "" {
		uc.Logger.Warn("public: no search results — LLM will respond with fallback", "slug", slug, "document_ids", group.DocumentIDs)
		sb.WriteString("No se encontró información relevante en los documentos.\n\n")
	} else {
		if lowConfidence {
			sb.WriteString("NOTA: La pregunta es muy general. Usá el contexto disponible para describir brevemente de qué tratan los documentos o qué información contienen. Si el usuario quiere saber algo específico, pedile que reformule la pregunta con más detalle.\n\n")
		}
		sb.WriteString("CONTEXTO:\n")
		sb.WriteString(contextStr)
		sb.WriteString("\n")
	}
	sb.WriteString("PREGUNTA: ")
	sb.WriteString(message)
	sb.WriteString("\n\nRESPUESTA:")

	answer, err := uc.LLM.GenerateAnswer(ctx, sb.String())
	if err != nil {
		return "", fmt.Errorf("generating answer: %w", err)
	}

	_ = uc.GroupRepo.IncrementQuotaUsed(ctx, group.ID)
	_ = uc.UserRepo.IncrementChatQuotaUsed(ctx, user.ID)
	uc.Logger.Info("public: chat answered", "slug", slug, "answer_len", len(answer))

	return answer, nil
}

func (uc *PublicChatUsecase) DownloadDocument(ctx context.Context, slug string, docID int) (io.ReadCloser, string, error) {
	group, err := uc.GroupRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, "", err
	}
	if group == nil || !group.IsActive {
		return nil, "", ErrGroupInactive
	}
	if !group.AllowDownloads {
		return nil, "", errors.New("downloads not allowed for this group")
	}

	belongs := false
	for _, id := range group.DocumentIDs {
		if id == docID {
			belongs = true
			break
		}
	}
	if !belongs {
		uc.Logger.Warn("public: doc does not belong to group", "slug", slug, "doc_id", docID, "group_docs", group.DocumentIDs)
		return nil, "", ErrGroupInactive
	}

	doc, err := uc.DocRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, "", err
	}

	reader, err := uc.BlobStorage.Download(ctx, doc.FilePath)
	if err != nil {
		return nil, "", err
	}

	return reader, doc.Filename, nil
}
