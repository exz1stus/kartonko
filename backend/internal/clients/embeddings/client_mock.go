package embeddings

import (
	"context"
	"io"
	"sync"
)

// MockEmbeddingsClient is a test double for EmbeddingsClient
type MockEmbeddingsClient struct {
	mu sync.RWMutex

	// Store embeddings by image ID
	embeddings map[int][]float32

	// Search results to return
	searchResults []SearchResult

	// Error to return on next call
	err error

	// Track calls
	embedCalls    []io.Reader
	upsertCalls   []struct{ id int; emb []float32 }
	deleteCalls   []int
	searchCalls   []struct{ emb []float32; limit int }
	searchImgCalls []io.Reader
}

func NewMockEmbeddingsClient() *MockEmbeddingsClient {
	return &MockEmbeddingsClient{
		embeddings:    make(map[int][]float32),
		searchResults: []SearchResult{},
	}
}

func (m *MockEmbeddingsClient) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

func (m *MockEmbeddingsClient) SetSearchResults(results []SearchResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchResults = results
}

func (m *MockEmbeddingsClient) Embed(ctx context.Context, file io.Reader) (*EmbedResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	// Return a fixed dummy embedding
	emb := make([]float32, 512)
	for i := range emb {
		emb[i] = 0.1
	}
	m.embedCalls = append(m.embedCalls, file)
	return &EmbedResponse{Embedding: emb, Dimension: 512}, nil
}

func (m *MockEmbeddingsClient) EmbedBatch(ctx context.Context, files []io.Reader) ([]EmbedResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	var results []EmbedResponse
	for range files {
		emb := make([]float32, 512)
		for i := range emb {
			emb[i] = 0.1
		}
		results = append(results, EmbedResponse{Embedding: emb, Dimension: 512})
	}
	return results, nil
}

func (m *MockEmbeddingsClient) EmbedText(ctx context.Context, text string) (*EmbedResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	emb := make([]float32, 512)
	for i := range emb {
		emb[i] = 0.1
	}
	return &EmbedResponse{Embedding: emb, Dimension: 512}, nil
}

func (m *MockEmbeddingsClient) Upsert(ctx context.Context, imageID int, embedding []float32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return m.err
	}

	m.embeddings[imageID] = embedding
	m.upsertCalls = append(m.upsertCalls, struct{ id int; emb []float32 }{imageID, embedding})
	return nil
}

func (m *MockEmbeddingsClient) Delete(ctx context.Context, imageID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return m.err
	}

	delete(m.embeddings, imageID)
	m.deleteCalls = append(m.deleteCalls, imageID)
	return nil
}

func (m *MockEmbeddingsClient) SearchImage(ctx context.Context, file io.Reader, limit int) ([]SearchResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	m.searchImgCalls = append(m.searchImgCalls, file)
	return m.searchResults, nil
}

func (m *MockEmbeddingsClient) SearchVector(ctx context.Context, embedding []float32, limit int) ([]SearchResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	m.searchCalls = append(m.searchCalls, struct{ emb []float32; limit int }{embedding, limit})
	return m.searchResults, nil
}

// GetEmbedding retrieves stored embedding for verification
func (m *MockEmbeddingsClient) GetEmbedding(imageID int) ([]float32, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	emb, ok := m.embeddings[imageID]
	return emb, ok
}

// GetCalls returns recorded calls for verification
func (m *MockEmbeddingsClient) GetCalls() ([]io.Reader, []struct{ id int; emb []float32 }, []int, []struct{ emb []float32; limit int }, []io.Reader) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.embedCalls, m.upsertCalls, m.deleteCalls, m.searchCalls, m.searchImgCalls
}

var _ EmbeddingsClient = (*MockEmbeddingsClient)(nil)