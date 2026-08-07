package services

import (
	"bytes"
	"context"
	"server/internal/clients/embeddings"
)

type EmbeddingsService interface {
	Search(ctx context.Context, query string, limit uint) ([]embeddings.SearchResult, error)
	SearchByImage(ctx context.Context, id uint, limit uint) ([]embeddings.SearchResult, error)
	SearchByBytes(ctx context.Context, data []byte, limit uint) ([]embeddings.SearchResult, error)
	Upsert(ctx context.Context, id uint, data []byte) error
	UpsertBatch(ctx context.Context, id uint, data [][]byte) []error
	Delete(ctx context.Context, id uint) error
}

type embeddingsService struct {
	objects ObjectService
	client  embeddings.EmbeddingsClient
}

func NewEmbeddingsService(objects ObjectService, client embeddings.EmbeddingsClient) EmbeddingsService {
	return &embeddingsService{
		objects,
		client,
	}
}

func (e *embeddingsService) Delete(ctx context.Context, id uint) error {
	return e.client.Delete(ctx, int(id))
}

func (e *embeddingsService) Search(ctx context.Context, query string, limit uint) ([]embeddings.SearchResult, error) {
	res, err := e.client.EmbedText(ctx, query)
	if err != nil {
		return nil, err
	}

	return e.client.SearchVector(ctx, res.Embedding, int(limit))
}

func (e *embeddingsService) SearchByBytes(ctx context.Context, data []byte, limit uint) ([]embeddings.SearchResult, error) {
	return e.client.SearchImage(ctx, bytes.NewReader(data), int(limit))
}

func (e *embeddingsService) SearchByImage(ctx context.Context, id uint, limit uint) ([]embeddings.SearchResult, error) {
	data, _, err := e.objects.GetRawImage(ctx, id)
	if err != nil {
		return nil, err
	}

	return e.client.SearchImage(ctx, data, int(limit))
}

func (e *embeddingsService) Upsert(ctx context.Context, id uint, data []byte) error {
	res, err := e.client.Embed(ctx, bytes.NewReader(data))
	if err != nil {
		return err
	}

	return e.client.Upsert(ctx, int(id), res.Embedding)
}

func (e *embeddingsService) UpsertBatch(ctx context.Context, id uint, data [][]byte) []error {
	panic("unimplemented")
}
