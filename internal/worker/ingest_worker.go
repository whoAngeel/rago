package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/whoAngeel/rago/internal/application"
	"github.com/whoAngeel/rago/internal/core/domain"
	"github.com/whoAngeel/rago/internal/core/ports"
	"github.com/whoAngeel/rago/internal/infrastructure/config"
	"github.com/whoAngeel/rago/internal/infrastructure/logger"
)

type IngestWorker struct {
	DocRepo      ports.DocumentRepository
	BlobStorage  ports.BlobStorage
	ParserRegistry ports.ParserRegistry
	Chunker      ports.Chunker
	Embedder     ports.Embedder
	IngestUC     *application.IngestUsecase
	PollInterval time.Duration
	Concurrency  int
	MaxRetries   int
	spotCh       chan struct{}
	processed    atomic.Int64
	wg           sync.WaitGroup
	config       config.Config
	logger       ports.Logger
}

func NewIngestWorker(
	docRepo ports.DocumentRepository,
	blobStorage ports.BlobStorage,
	parserRegistry ports.ParserRegistry,
	chunker ports.Chunker,
	embedder ports.Embedder,
	ingestUC *application.IngestUsecase,
	pollInterval time.Duration,
	concurrency int,
	maxRetries int,
	config config.Config,
) *IngestWorker {
	log := logger.New(config.Env).With()
	return &IngestWorker{
		DocRepo:        docRepo,
		BlobStorage:    blobStorage,
		ParserRegistry: parserRegistry,
		Chunker:        chunker,
		Embedder:       embedder,
		IngestUC:       ingestUC,
		PollInterval:   pollInterval,
		Concurrency:    concurrency,
		MaxRetries:     maxRetries,
		spotCh:         make(chan struct{}, concurrency),
		config:         config,
		logger:         log,
	}
}

func (w *IngestWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.PollInterval)
	defer ticker.Stop()

	w.logger.Info("IngestWorker started", "poll_interval", w.PollInterval, "concurrency", w.Concurrency)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("IngestWorker stopping, waiting for active jobs...")
			w.wg.Wait()
			w.logger.Info("IngestWorker stopped", "total_processed", w.processed.Load())
			return
		case <-ticker.C:
			docs, err := w.DocRepo.FindPendingDocuments(ctx, w.Concurrency)

			if err != nil {
				w.logger.Error("Polling error", "error", err)
				continue
			}

			if len(docs) == 0 {
				continue
			}

			for _, doc := range docs {
				w.spotCh <- struct{}{}
				w.wg.Add(1)
				go func(d *domain.Document) {
					defer w.wg.Done()
					defer func() { <-w.spotCh }()
					w.processDocument(ctx, d)
				}(doc)
			}

		}
	}

}

func (w *IngestWorker) processDocument(ctx context.Context, doc *domain.Document) {
	log := w.logger.With("doc_id", doc.ID, "filename", doc.Filename)

	now := time.Now()
	doc.Status = domain.StatusProcessing
	doc.ProcessingStartedAt = &now
	if _, err := w.DocRepo.UpdateDocument(ctx, doc); err != nil {
		log.Error("failed to update document status", "error", err)
		return
	}

	stepID, start := w.startStep(ctx, doc.ID, "download")
	reader, err := w.BlobStorage.Download(ctx, doc.FilePath)
	if err != nil {
		w.finishStep(ctx, stepID, start, err)
		w.handleDocumentError(ctx, doc, fmt.Errorf("download: %w", err))
		return
	}
	defer reader.Close()

	rawBytes, err := io.ReadAll(reader)
	if err != nil {
		w.finishStep(ctx, stepID, start, err)
		w.handleDocumentError(ctx, doc, fmt.Errorf("reading file: %w", err))
		return
	}
	w.finishStep(ctx, stepID, start, nil)

	stepID, start = w.startStep(ctx, doc.ID, "parse")
	parser, err := w.ParserRegistry.Resolve(doc.ContentType)
	if err != nil {
		w.finishStep(ctx, stepID, start, err)
		w.handleDocumentError(ctx, doc, fmt.Errorf("parser resolve: %w", err))
		return
	}
	parsedDocs, err := parser.Parse(ctx, bytes.NewReader(rawBytes), doc.ContentType)
	w.finishStep(ctx, stepID, start, err)
	if err != nil {
		w.handleDocumentError(ctx, doc, fmt.Errorf("parsing: %w", err))
		return
	}

	var chunks []string
	for _, parsedDoc := range parsedDocs {
		if parsedDoc.Metadata["chunk_type"] == "structured" {
			chunks = append(chunks, parsedDoc.PageContent)
		} else {
			stepID, start = w.startStep(ctx, doc.ID, "chunk")
			docChunks, err := w.Chunker.Chunk(parsedDoc.PageContent)
			w.finishStep(ctx, stepID, start, err)
			if err != nil {
				w.handleDocumentError(ctx, doc, fmt.Errorf("chunking: %w", err))
				return
			}
			chunks = append(chunks, docChunks...)
		}
	}

	if len(chunks) == 0 {
		w.handleDocumentError(ctx, doc, fmt.Errorf("no chunks produced"))
		return
	}

	err = w.IngestUC.Execute(ctx, doc, nil, chunks, func(stepName string, start time.Time, err error) {
		stepID, _ := w.startStep(ctx, doc.ID, stepName)
		w.finishStep(ctx, stepID, start, err)
	})
	if err != nil {
		w.handleDocumentError(ctx, doc, fmt.Errorf("ingest: %w", err))
		return
	}

	doc.Status = domain.StatusCompleted
	doc.ErrorMessage = ""
	if _, err := w.DocRepo.UpdateDocument(ctx, doc); err != nil {
		log.Error("failed to mark document completed", "error", err)
		return
	}
	w.processed.Add(1)
	log.Info("document processed successfully")
}

func (w *IngestWorker) startStep(ctx context.Context, docID int, stepName string) (int, time.Time) {
	start := time.Now()
	step := &domain.ProcessingStep{
		DocumentID: docID,
		StepName:   stepName,
		Status:     "started",
	}
	if err := w.DocRepo.CreateProcessingStep(ctx, step); err != nil {
		w.logger.Warn("failed to create processing step", "step", stepName, "doc_id", docID, "error", err)
		return 0, start
	}
	return step.ID, start
}

func (w *IngestWorker) finishStep(ctx context.Context, stepID int, start time.Time, err error) {
	if stepID == 0 {
		return
	}
	duration := int(time.Since(start).Milliseconds())
	status := "completed"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}
	w.DocRepo.UpdateProcessingStep(ctx, stepID, duration, status, errMsg)
}

func (w *IngestWorker) handleDocumentError(ctx context.Context, doc *domain.Document, err error) {
	log := w.logger.With("doc_id", doc.ID, "filename", doc.Filename)
	doc.ErrorMessage = err.Error()
	doc.RetryCount++

	if doc.RetryCount >= w.MaxRetries {
		doc.Status = domain.StatusFailed
		log.Error("document permanently failed", "error", err, "retry_count", doc.RetryCount)
	} else {
		doc.Status = domain.StatusPending
		log.Warn("document failed, will retry", "error", err, "retry_count", doc.RetryCount)
	}

	if _, updateErr := w.DocRepo.UpdateDocument(ctx, doc); updateErr != nil {
		log.Error("failed to update document error state", "original_error", err, "update_error", updateErr)
	}
}

func (w *IngestWorker) Stop() {

}
