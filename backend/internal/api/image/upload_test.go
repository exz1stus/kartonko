package image_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"server/internal/image"
	"server/internal/storage"
	"server/internal/tag"

	tutil "server/internal/testutil/testing"
)

func buildUploadRequest(t *testing.T, url, metadataJSON, filename, contentType string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if err := w.WriteField("metadata", metadataJSON); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	if filename != "" && len(content) > 0 {
		if err := tutil.WriteFilePart(w, "file", filename, contentType, content); err != nil {
			t.Fatalf("write file part %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestPostImage(t *testing.T) {
	tests := []struct {
		name       string
		metadata   string
		filename   string
		content    func(t *testing.T) []byte
		mimeType   string
		userID     uint64
		wantStatus int
	}{
		{
			name:       "success valid png",
			metadata:   `{"name":"cat.png","tags":["animal"]}`,
			filename:   "cat.png",
			content:    func(t *testing.T) []byte { return tutil.MakeTestPNG(t, 10, 10) },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing name in metadata",
			metadata:   `{"name":""}`,
			filename:   "x.png",
			content:    func(t *testing.T) []byte { return tutil.MakeTestPNG(t, 5, 5) },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid metadata JSON",
			metadata:   `{not json`,
			filename:   "x.png",
			content:    func(t *testing.T) []byte { return tutil.MakeTestPNG(t, 5, 5) },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported mime type",
			metadata:   `{"name":"doc.pdf"}`,
			filename:   "doc.pdf",
			content:    func(t *testing.T) []byte { return []byte("%PDF-1.4") },
			mimeType:   "application/pdf",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "corrupt image bytes with valid mime",
			metadata:   `{"name":"broken.png"}`,
			filename:   "broken.png",
			content:    func(t *testing.T) []byte { return []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00} },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "unauthenticated request",
			metadata:   `{"name":"noauth.png"}`,
			filename:   "noauth.png",
			content:    func(t *testing.T) []byte { return tutil.MakeTestPNG(t, 5, 5) },
			mimeType:   "image/png",
			userID:     0,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "no file attached",
			metadata:   `{"name":"nofile.png"}`,
			filename:   "",
			content:    func(t *testing.T) []byte { return nil },
			mimeType:   "",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "tags are not added (invalid JSON array)",
			metadata:   `{"name":"nofile.png","tags":[some_tag1, some_tag2]}`,
			filename:   "notallowedtags.png",
			content:    func(t *testing.T) []byte { return tutil.MakeTestPNG(t, 5, 5) },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cleanup := setupContext(t)
			defer cleanup()

			req := buildUploadRequest(t, "/upload", tt.metadata, tt.filename, tt.mimeType, tt.content(t))
			if tt.userID != 0 {
				req = tutil.WithTestUser(req, tt.userID)
			}
			rec := httptest.NewRecorder()
			ctx.Router.ServeHTTP(rec, req)

			ctx.AssertStatus(rec, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				ctx.AssertImageCount(nil, 1)
				ctx.AssertStorageCount(2)
			}
		})
	}
}

func TestPostImage_TagsAdded(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	metadata := `{"name":"tagged.png","tags":["animal","cat"]}`
	rec := ctx.UploadImage(metadata, "tagged.png", tutil.MakeTestPNG(t, 10, 10), 1)

	ctx.AssertStatus(rec, http.StatusOK)

	// Verify response contains tags
	var resp image.ImageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	tutil.AssertTags(t, resp.Tags, []string{"animal", "cat"}, "response")

	// Verify tags are persisted in database
	img, err := ctx.ImageService.GetByName("tagged.png")
	if err != nil {
		t.Fatalf("failed to get image from DB: %v", err)
	}
	dbTags := tag.TagsToStrings(img.Tags)
	tutil.AssertTags(t, dbTags, []string{"animal", "cat"}, "database")

	ctx.AssertImageCount(nil, 1)
	ctx.AssertStorageCount(2)
}

func TestPostImage_DuplicateName(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	req1 := buildUploadRequest(t, "/upload", `{"name":"dup.png"}`, "dup.png", "image/png", tutil.MakeTestPNG(t, 5, 5))
	req1 = tutil.WithTestUser(req1, 1)
	rec1 := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec1, req1)

	req2 := buildUploadRequest(t, "/upload", `{"name":"dup.png"}`, "dup.png", "image/png", tutil.MakeTestPNG(t, 5, 5))
	req2 = tutil.WithTestUser(req2, 1)
	rec2 := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for duplicate name, got %d: %s", rec2.Code, rec2.Body.String())
	}

	ctx.AssertImageCount(nil, 1)
	ctx.AssertStorageCount(2)
}

func TestPostImage_DuplicateHash_DifferentName(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()
	content := tutil.MakeTestPNG(t, 7, 7)

	req1 := buildUploadRequest(t, "/upload", `{"name":"first.png"}`, "first.png", "image/png", content)
	req1 = tutil.WithTestUser(req1, 1)
	rec1 := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first upload failed: %d %s", rec1.Code, rec1.Body.String())
	}

	req2 := buildUploadRequest(t, "/upload", `{"name":"second.png"}`, "second.png", "image/png", content)
	req2 = tutil.WithTestUser(req2, 1)
	rec2 := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("expected duplicate-hash upload to fail, got %d: %s", rec2.Code, rec2.Body.String())
	}

	ctx.AssertImageCount(nil, 1)
	ctx.AssertStorageCount(2)
}

func TestPostImage_StorageUploadFails_NoDBRowAndStorageImageLeft(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()
	ctx.Storage.(*storage.MockStorage).FailUploadOn = func(key string) error {
		if !strings.Contains(key, "thumb") {
			return fmt.Errorf("simulated storage failure on main image")
		}
		return nil
	}

	req := buildUploadRequest(t, "/upload", `{"name":"y.png"}`, "y.png", "image/png", tutil.MakeTestPNG(t, 5, 5))
	req = tutil.WithTestUser(req, 1)
	rec := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}

	query := image.NewQueryBuilder().Prefix("dog_duplicate.png.png").Build()

	ctx.AssertImageCount(query, 0)
	ctx.AssertStorageCount(0)
}

func TestPostImage_StorageThumbUploadFails_NoDBRowAndStorageImageLeft(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()
	ctx.Storage.(*storage.MockStorage).FailUploadOn = func(key string) error {
		if strings.Contains(key, "thumb") {
			return fmt.Errorf("simulated storage failure on thumb")
		}
		return nil
	}

	req := buildUploadRequest(t, "/upload", `{"name":"y.png"}`, "y.png", "image/png", tutil.MakeTestPNG(t, 5, 5))
	req = tutil.WithTestUser(req, 1)
	rec := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}

	query := image.NewQueryBuilder().Prefix("dog_duplicate.png.png").Build()

	ctx.AssertImageCount(query, 0)
	ctx.AssertStorageCount(0)
}
