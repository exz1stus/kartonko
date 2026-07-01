package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"server/internal/models"
	"testing"
)

// Posting batch of images
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
			a, store := newTestAPI(t)
			r := newTestRouter(a)

			tags := []string{"animal", "cat", "dog"}
			seedTestTags(t, a.models.Tags, tags)

			fileDatas := tt.fileDatas(t)

			req := buildBatchUploadRequest(t, "/upload/batch", tt.metadata, fileDatas)
			if tt.userID != 0 {
				req = withTestUser(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if tt.wantStatus < 400 && tt.wantStatus != 207 {
				var res ImageBatchResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
					t.Errorf("failed to deserialize batch upload response: %s", rec.Body.String())
				}

				if len(res.Failures) > 0 {
					for _, err := range res.Failures {
						t.Errorf("batch upload error: %s", err)
					}
				}
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantStatus == http.StatusOK && store.Count() != len(fileDatas)*2 {
				t.Errorf("expected %d stored objects (image+thumb), got %d", len(fileDatas)*2, store.Count())
			}
		})
	}
}

func TestPostImagesBatch_MixedSuccessAndFailure(t *testing.T) {
	a, store := newTestAPI(t)
	r := newTestRouter(a)

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
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMultiStatus {
		t.Errorf("got status %d, want %d, body=%s", rec.Code, http.StatusMultiStatus, rec.Body.String())
	}

	var res ImageBatchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Errorf("failed to deserialize batch upload response: %s", rec.Body.String())
	}

	var count int64
	a.models.Images.Db.Model(&models.ImageMetadata{}).Where("filename = ?", "dog_duplicate.png.png").Count(&count)
	if count != 0 {
		t.Errorf("expected no failed images in db")
	}

	if len(res.Successes) != succesCount {
		t.Errorf("expected %d successes , got %d", succesCount, len(res.Successes))
	}

	if store.Count() != succesCount*2 {
		t.Errorf("expected %d stored objects (image+thumb), got %d", succesCount*2, store.Count())
	}

}
