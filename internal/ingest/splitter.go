package ingest

import (
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type Splitter struct {
	splitter textsplitter.RecursiveCharacter
}

func NewRecursiveTextSplitter(chunkSize, chunkOverlap int) *Splitter {
	return &Splitter{
		splitter: textsplitter.NewRecursiveCharacter(
			textsplitter.WithChunkSize(chunkSize),
			textsplitter.WithChunkOverlap(chunkOverlap),
		),
	}
}

func (s *Splitter) SplitDocuments(docs []schema.Document) ([]schema.Document, error) {
	return textsplitter.SplitDocuments(s.splitter, docs)
}
