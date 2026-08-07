package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"server/internal/api/dto"
	"server/internal/models"
	"testing"
)

func buildBatchUploadRequest(t *testing.T, url string, metadataJSON string, files []TestFileData) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if err := w.WriteField("metadata", metadataJSON); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	for _, file := range files {
		writeFilePart(t, w, "files", file)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

//TODO: very large upload

func TestPostImagesBatch(t *testing.T) {
	tests := []struct {
		name       string
		metadata   string
		fileDatas  func(t *testing.T) []TestFileData
		userID     uint64
		wantStatus int
	}{
		{
			name:     "success batch upload",
			metadata: `{"data": [{"name":"cat.png","tags":["cat"]},{"name":"dog.png","tags":["dog"]}], "common_tags": ["animal"]}`,
			fileDatas: func(t *testing.T) []TestFileData {
				return []TestFileData{
					{
						filename:    "cat.png",
						contentType: "image/png",
						content:     makeTestPNG(t, 5, 10),
					},
					{
						filename:    "dog.png",
						contentType: "image/png",
						content:     makeTestPNG(t, 10, 10),
					},
				}
			},
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name:     "fewer files than metadata",
			metadata: `{"data": [{"name":"cat.png"},{"name":"dog.png"}]}`,
			fileDatas: func(t *testing.T) []TestFileData {
				return []TestFileData{
					{
						filename:    "cat.png",
						contentType: "image/png",
						content:     makeTestPNG(t, 5, 10),
					},
				}
			},
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContext(t)

			tags := []string{"animal", "cat", "dog"}
			seedTestTags(t, ctx.a.tagService, tags, ctx.mod)

			fileDatas := tt.fileDatas(t)

			req := buildBatchUploadRequest(t, "/upload/batch", tt.metadata, fileDatas)
			if tt.userID != 0 {
				req = withTestUser(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			ctx.r.ServeHTTP(rec, req)

			if tt.wantStatus < 400 && tt.wantStatus != 207 {
				var res dto.ImagePostBatchResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
					t.Errorf("failed to deserialize batch upload response: %s", rec.Body.String())
				}

				if len(res.Failures) > 0 {
					for _, err := range res.Failures {
						t.Errorf("batch upload error: %s", err)
					}
				}

				ctx.assertStorageCount(len(fileDatas) * 2)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("got status %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestPostImagesBatch_MixedSuccessAndFailure(t *testing.T) {
	ctx := newTestContext(t)

	metadata := `{"data": [{"name":"cat.png"},{"name":"dog.png"},{"name":"cat.png"}]}`

	fileDatas := []TestFileData{
		{
			filename:    "cat.png",
			contentType: "image/png",
			content:     makeTestPNG(t, 5, 10),
		},
		{
			filename:    "dog.png",
			contentType: "image/png",
			content:     makeTestPNG(t, 10, 10),
		},
	}

	failedData := TestFileData{
		filename:    "dog_duplicate.png",
		contentType: "image/png",
		content:     makeTestPNG(t, 10, 10),
	}

	succesCount := len(fileDatas)

	fileDatas = append(fileDatas, failedData)

	req := buildBatchUploadRequest(t, "/upload/batch", metadata, fileDatas)
	req = withTestUser(req, 1)

	rec := httptest.NewRecorder()
	ctx.r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMultiStatus {
		t.Errorf("got status %d, want %d, body=%s", rec.Code, http.StatusMultiStatus, rec.Body.String())
	}

	var res dto.ImagePostBatchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Errorf("failed to deserialize batch upload response: %s", rec.Body.String())
	}

	query := models.NewImageQueryBuilder().
		Prefix("dog_duplicate.png.png").
		Build()

	ctx.assertImageCount(query, 0)

	if len(res.Successes) != succesCount {
		t.Errorf("expected %d successes , got %d", succesCount, len(res.Successes))
	}

	ctx.assertStorageCount(succesCount * 2)
}
