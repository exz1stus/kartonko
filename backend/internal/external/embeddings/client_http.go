package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type httpEmbeddingClient struct {
	baseURL string
	client  *http.Client
}

func (e *httpEmbeddingClient) do(req *http.Request, out any) error {
	res, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"embedding service returned %d: %s",
			res.StatusCode,
			string(body),
		)
	}

	if out == nil {
		return nil
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("failed decoding response: %w", err)
	}

	return nil
}

func (h *httpEmbeddingClient) Delete(ctx context.Context, imageID int) error {
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("%s/vectors/%d", h.baseURL, imageID),
		nil,
	)
	if err != nil {
		return err
	}

	return h.do(req, nil)
}

func (h *httpEmbeddingClient) Embed(ctx context.Context, file io.Reader) (*EmbedResponse, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	part, err := w.CreateFormFile("file", "file")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/embedding", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	out := &EmbedResponse{}
	err = h.do(req, out)

	return out, err
}

func (h *httpEmbeddingClient) EmbedBatch(ctx context.Context, files []io.Reader) ([]EmbedResponse, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	for _, file := range files {
		part, err := w.CreateFormFile("files", "file")
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(part, file); err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/embedding/batch", body)
	if err != nil {
		return nil, err
	}

	var out []EmbedResponse
	err = h.do(req, out)

	return out, err
}

func (h *httpEmbeddingClient) SearchImage(ctx context.Context, file io.Reader, limit int) ([]SearchResult, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	part, err := w.CreateFormFile("file", "file")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/search/image", h.baseURL),
		body,
	)
	if err != nil {
		return nil, err
	}

	out := &SearchResponse{}
	if err := h.do(req, out); err != nil {
		return nil, err
	}

	return out.Results, nil
}

func (h *httpEmbeddingClient) SearchVector(ctx context.Context, embedding []float32, limit int) ([]SearchResult, error) {
	body, err := json.Marshal(embedding)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/search/vector", h.baseURL),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	out := &SearchResponse{}
	if err := h.do(req, out); err != nil {
		return nil, err
	}

	return out.Results, nil
}

func (h *httpEmbeddingClient) Upsert(ctx context.Context, imageID int, embedding []float32) error {
	body, err := json.Marshal(embedding)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s/vectors/%d", h.baseURL, imageID),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	return h.do(req, nil)
}

func (h *httpEmbeddingClient) EmbedText(ctx context.Context, text string) (*EmbedResponse, error) {
	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/embedding/text", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	out := &EmbedResponse{}
	err = h.do(req, out)

	return out, err
}

func NewHTTPEmbeddingClient() EmbeddingsClient {
	return &httpEmbeddingClient{
		baseURL: fmt.Sprintf("embeddings:%s/api/v1", os.Getenv("EMBEDDINGS_PORT")),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
