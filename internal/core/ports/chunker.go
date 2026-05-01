package ports

type Chunker interface {
	Chunk(text string) ([]string, error)
}
