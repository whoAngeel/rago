package application

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type Source struct {
	Content string  `json:"content"`
	Source  string  `json:"source"`
	Score   float32 `json:"score"`
}

type ChatUsecase struct {
	ChatRepo       ports.ChatRepository
	ConfigRepo     ports.SystemConfigRepository
	VectorStore    ports.VectorStore
	Embedder       ports.Embedder
	LLM            ports.LLMProvider
	Logger         ports.Logger
	HistoryLimit   int
	CollectionName string
	ContextLimit   int
}

func NewChatUsecase(
	chatRepo ports.ChatRepository,
	configRepo ports.SystemConfigRepository,
	vectorStore ports.VectorStore,
	embedder ports.Embedder,
	llm ports.LLMProvider,
	logger ports.Logger,
	historyLimit int,
	collectionName string,
	contextLimit int,
) *ChatUsecase {
	return &ChatUsecase{
		ChatRepo:       chatRepo,
		ConfigRepo:     configRepo,
		VectorStore:    vectorStore,
		Embedder:       embedder,
		LLM:            llm,
		Logger:         logger,
		HistoryLimit:   historyLimit,
		CollectionName: collectionName,
		ContextLimit:   contextLimit,
	}
}

const fallbackSystemPrompt = `Eres un asistente experto que responde preguntas basándose ÚNICAMENTE en la sección CONTEXTO proporcionada.
Instrucciones:
1. Usa solo la información en la sección CONTEXTO para responder.
2. Si el CONTEXTO no tiene información suficiente, responde: "No tengo información suficiente en tus documentos para responder a esto."
3. No inventes ni uses conocimiento general.
4. Si mencionas datos, cita las fuentes proporcionadas.`

func (uc *ChatUsecase) SendMessage(
	ctx context.Context,
	userID int,
	sessionID *int,
	question string,
) (answer string, sources []Source, newSessionID int, err error) {
	var session *domain.ChatSession
	isNew := false

	if sessionID == nil {
		session = &domain.ChatSession{UserID: userID}
		if err := uc.ChatRepo.CreateSession(ctx, session); err != nil {
			return "", nil, 0, fmt.Errorf("creating session: %w", err)
		}
		isNew = true
		uc.Logger.Info("New chat session created", "session_id", session.ID, "user_id", userID)
	} else {
		session, err = uc.ChatRepo.GetSession(ctx, *sessionID, userID)
		if err != nil {
			return "", nil, 0, fmt.Errorf("getting session: %w", err)
		}
	}

	systemPrompt, err := uc.ConfigRepo.Get(ctx, "system_prompt")
	if err != nil {
		systemPrompt = fallbackSystemPrompt
		uc.Logger.Warn("Error getting system prompt, using fallback", "error", err)
	}

	var historyStr string
	if !isNew {
		messages, err := uc.ChatRepo.GetMessages(ctx, int(session.ID), uc.HistoryLimit)
		if err != nil {
			return "", nil, int(session.ID), fmt.Errorf("getting messages: %w", err)
		}
		slices.Reverse(messages)
		var parts []string
		for _, m := range messages {
			parts = append(parts, fmt.Sprintf("%s: %s", m.Role, m.Content))
		}
		historyStr = strings.Join(parts, "\n")
	}

	queryVector, err := uc.Embedder.EmbedText(ctx, question)
	if err != nil {
		return "", nil, int(session.ID), fmt.Errorf("embedding: %w", err)
	}

	searchResults, err := uc.VectorStore.Search(ctx, uc.CollectionName, queryVector, userID, uc.ContextLimit)
	if err != nil {
		return "", nil, int(session.ID), fmt.Errorf("searching: %w", err)
	}
	uc.Logger.Info("search completed", "results", len(searchResults))

	var contextStr string
	for _, r := range searchResults {
		src := "desconocido"
		if s, ok := r.Document.Metadata["source"]; ok {
			src = fmt.Sprintf("%v", s)
		}
		sources = append(sources, Source{
			Content: r.Document.PageContent,
			Source:  src,
			Score:   r.Score,
		})
		contextStr += fmt.Sprintf("[Fuente: %s]\n%s\n\n", src, r.Document.PageContent)
	}

	prompt := systemPrompt + "\n\n"

	if len(searchResults) == 0 {
		prompt += "No se encontró información relevante en los documentos.\n\n"
	} else {
		prompt += "CONTEXTO:\n" + contextStr + "\n"
	}

	if historyStr != "" {
		prompt += "HISTORIAL:\n" + historyStr + "\n\n"
	}

	prompt += "PREGUNTA: " + question + "\n\nRESPUESTA:"

	uc.Logger.Info("Generating answer", "prompt_len", len(prompt))
	answer, err = uc.LLM.GenerateAnswer(ctx, prompt)
	if err != nil {
		return "", nil, int(session.ID), fmt.Errorf("generating answer: %w", err)
	}
	uc.Logger.Info("Answer generated", "answer_len", len(answer))

	sourcesJSON, err := json.Marshal(sources)
	if err != nil {
		return "", nil, int(session.ID), fmt.Errorf("marshaling sources: %w", err)
	}
	uc.Logger.Info("Sources JSON", "len", len(sourcesJSON), "sources_count", len(sources))

	userMsg := domain.ChatMessage{
		SessionID: int(session.ID),
		Role:      "user",
		Content:   question,
		Sources:   "[]",
	}
	if err := uc.ChatRepo.CreateMessage(ctx, &userMsg); err != nil {
		return "", nil, int(session.ID), fmt.Errorf("saving user message: %w", err)
	}
	uc.Logger.Info("User message saved", "session_id", session.ID, "user_id", userID, "message", question)

	assistantMsg := domain.ChatMessage{
		SessionID: int(session.ID),
		Role:      "assistant",
		Content:   answer,
		Sources:   datatypes.JSON(sourcesJSON),
	}
	if err := uc.ChatRepo.CreateMessage(ctx, &assistantMsg); err != nil {
		return "", nil, int(session.ID), fmt.Errorf("saving assistant message: %w", err)
	}
	uc.Logger.Info("assistant message saved", "session_id", session.ID, "user_id", userID, "message", answer)

	if isNew {
		title := question
		if len([]rune(title)) > 80 {
			title = string([]rune(title)[:80]) + "..."
		}
		session.Title = title
		if err := uc.ChatRepo.UpdateSessionTitle(ctx, int(session.ID), userID, title); err != nil {
			uc.Logger.Error("Failed to update session title", "error", err)
		}
	}

	newSessionID = int(session.ID)

	return answer, sources, newSessionID, nil
}

func (uc *ChatUsecase) ListSessions(ctx context.Context, userID int) ([]*domain.ChatSession, error) {
	sessions, err := uc.ChatRepo.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting sessions: %w", err)
	}
	return sessions, nil
}

func (uc *ChatUsecase) GetSessionHistory(ctx context.Context, sessionID int, userID int) ([]*domain.ChatMessage, error) {
	_, err := uc.ChatRepo.GetSession(ctx, sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("validating session: %w", err)
	}
	messages, err := uc.ChatRepo.GetAllMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("getting messages: %w", err)
	}
	return messages, nil
}

func (uc *ChatUsecase) DeleteSession(ctx context.Context, sessionID, userID int) error {
	err := uc.ChatRepo.DeleteSession(ctx, sessionID, userID)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

func (uc *ChatUsecase) UpdateSessionTitle(ctx context.Context, sessionID int, userID int, title string) error {
	err := uc.ChatRepo.UpdateSessionTitle(ctx, sessionID, userID, title)
	if err != nil {
		return fmt.Errorf("updating session title: %w", err)
	}
	return nil
}
