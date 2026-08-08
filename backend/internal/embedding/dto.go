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
	ID      string                 `json:"id"`
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
