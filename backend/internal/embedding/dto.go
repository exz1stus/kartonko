package embeddings

type UpsertRequest struct {
	ImageID   string    `json:"image_id"`
	Embedding []float32 `json:"embedding"`
}

type SearchRequest struct {
	Embedding      []float32 `json:"embedding"`
	Limit          int       `json:"limit"`
	ScoreThreshold float32   `json:"score_threshold"`
}

type SearchResult struct {
	// Qdrant accepts both integer and UUID point IDs. Use any so the HTTP
	// client can decode the integer IDs used by this application.
	ID      any                    `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
	Dimension int       `json:"dimension"`
}
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}
