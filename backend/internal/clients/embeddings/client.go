package embeddings

import (
	"context"
	"io"
)

type EmbeddingsClient interface {
	Embed(ctx context.Context, file io.Reader) (*EmbedResponse, error)
	EmbedBatch(ctx context.Context, files []io.Reader) ([]EmbedResponse, error)
	EmbedText(ctx context.Context, text string) (*EmbedResponse, error)

	Upsert(ctx context.Context, imageID int, embedding []float32) error
	Delete(ctx context.Context, imageID int) error

	SearchImage(ctx context.Context, file io.Reader, limit int) ([]SearchResult, error)
	SearchVector(ctx context.Context, embedding []float32, limit int) ([]SearchResult, error)
}
