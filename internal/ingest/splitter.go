package ingest

import "github.com/tmc/langchaingo/textsplitter"

func NewRecursiveTextSplitter(chunkSize, chunkOverlap int) *textsplitter.RecursiveCharacter {
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(chunkSize),
		textsplitter.WithChunkOverlap(chunkOverlap),
	)
	return &splitter
}
