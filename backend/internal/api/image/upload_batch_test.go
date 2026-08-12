package image_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/internal/image"
	tutil "server/internal/testutil/testing"
)

type TestFileData struct {
	filename    string
	contentType string
	content     []byte
}

func buildBatchUploadRequest(t *testing.T, url string, metadataJSON string, files []TestFileData) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if err := w.WriteField("metadata", metadataJSON); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	for _, file := range files {
		if err := tutil.WriteFilePart(w, "files", file.filename, file.contentType, file.content); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

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
						content:     tutil.MakeTestPNG(t, 5, 10),
					},
					{
						filename:    "dog.png",
						contentType: "image/png",
						content:     tutil.MakeTestPNG(t, 10, 10),
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
						content:     tutil.MakeTestPNG(t, 5, 10),
					},
				}
			},
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cleanup := setupContext(t)
			defer cleanup()

			tags := []string{"animal", "cat", "dog"}
			mod, _ := ctx.UserService.GetByID(1)
			tutil.SeedTestTags(t, ctx.TagService, tags, mod)

			fileDatas := tt.fileDatas(t)

			req := buildBatchUploadRequest(t, "/upload/batch", tt.metadata, fileDatas)
			if tt.userID != 0 {
				req = tutil.WithTestUser(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			ctx.Router.ServeHTTP(rec, req)

			if tt.wantStatus < 400 && tt.wantStatus != 207 {
				var res image.ImagePostBatchResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
					t.Errorf("failed to deserialize batch upload response: %s", rec.Body.String())
				}

				if len(res.Failures) > 0 {
					for _, err := range res.Failures {
						t.Errorf("batch upload error: %s", err)
					}
				}

				ctx.AssertStorageCount(len(fileDatas) * 2)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("got status %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestPostImagesBatch_MixedSuccessAndFailure(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	metadata := `{"data": [{"name":"cat.png"},{"name":"dog.png"},{"name":"cat.png"}]}`

	fileDatas := []TestFileData{
		{
			filename:    "cat.png",
			contentType: "image/png",
			content:     tutil.MakeTestPNG(t, 5, 10),
		},
		{
			filename:    "dog.png",
			contentType: "image/png",
			content:     tutil.MakeTestPNG(t, 10, 10),
		},
	}

	failedData := TestFileData{
		filename:    "dog_duplicate.png",
		contentType: "image/png",
		content:     tutil.MakeTestPNG(t, 10, 10),
	}

	successCount := len(fileDatas)

	fileDatas = append(fileDatas, failedData)

	req := buildBatchUploadRequest(t, "/upload/batch", metadata, fileDatas)
	req = tutil.WithTestUser(req, 1)

	rec := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMultiStatus {
		t.Errorf("got status %d, want %d, body=%s", rec.Code, http.StatusMultiStatus, rec.Body.String())
	}

	var res image.ImagePostBatchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Errorf("failed to deserialize batch upload response: %s", rec.Body.String())
	}

	query := image.NewQueryBuilder().Prefix("dog_duplicate.png.png").Build()

	ctx.AssertImageCount(query, 0)

	if len(res.Successes) != successCount {
		t.Errorf("expected %d successes , got %d", successCount, len(res.Successes))
	}

	ctx.AssertStorageCount(successCount * 2)
}
