package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/storage"
	"strings"
	"testing"
)

func buildUploadRequest(t *testing.T, url, metadataJSON, filename, contentType string, content []byte) *http.Request {
	fileData := TestFileData{filename, contentType, content}
	return buildUploadRequestFileData(t, url, metadataJSON, fileData)
}

func buildUploadRequestFileData(t *testing.T, url string, metadataJSON string, file TestFileData) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if err := w.WriteField("metadata", metadataJSON); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	if file.filename != "" {
		writeFilePart(t, w, "file", file)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

// uploadTestCase represents a single test case for image upload
type uploadTestCase struct {
	name       string
	metadata   string
	filename   string
	content    func(t *testing.T) []byte
	mimeType   string
	userID     uint64
	wantStatus int
}

func newTestUploadContext(t *testing.T) *testContext {
	t.Helper()
	a := newTestAPI(t)
	r := newTestRouter(a)
	store := a.storage.(*storage.MockStorage)

	mod, err := a.userService.GetByID(1)
	if err != nil {
		t.Fatalf("failed getting moderator user")
	}
	seedTestTags(t, a.tagService, []string{"animal", "cat", "dog"}, mod)

	return &testContext{t: t, a: a, r: r, store: store, mod: mod}
}

func (c *testContext) upload(tt uploadTestCase) *httptest.ResponseRecorder {
	req := buildUploadRequest(c.t, "/upload", tt.metadata, tt.filename, tt.mimeType, tt.content(c.t))
	if tt.userID != 0 {
		req = withTestUser(req, tt.userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
}

func TestPostImage(t *testing.T) {
	tests := []uploadTestCase{
		{
			name:       "success valid png",
			metadata:   `{"name":"cat.png","tags":["animal"]}`,
			filename:   "cat.png",
			content:    func(t *testing.T) []byte { return makeTestPNG(t, 10, 10) },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing name in metadata",
			metadata:   `{"name":""}`,
			filename:   "x.png",
			content:    func(t *testing.T) []byte { return makeTestPNG(t, 5, 5) },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid metadata JSON",
			metadata:   `{not json`,
			filename:   "x.png",
			content:    func(t *testing.T) []byte { return makeTestPNG(t, 5, 5) },
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
			content:    func(t *testing.T) []byte { return makeTestPNG(t, 5, 5) },
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
			content:    func(t *testing.T) []byte { return makeTestPNG(t, 5, 5) },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestUploadContext(t)
			rec := ctx.upload(tt)

			ctx.assertStatus(rec, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				ctx.assertStorageCount(2)
			}
		})
	}
}

func TestPostImage_TagsAdded(t *testing.T) {
	ctx := newTestUploadContext(t)

	metadata := `{"name":"tagged.png","tags":["animal","cat"]}`
	rec := ctx.upload(uploadTestCase{
		metadata:   metadata,
		filename:   "tagged.png",
		content:    func(t *testing.T) []byte { return makeTestPNG(t, 10, 10) },
		mimeType:   "image/png",
		userID:     1,
		wantStatus: http.StatusOK,
	})

	ctx.assertStatus(rec, http.StatusOK)

	// Verify response contains tags
	var resp dto.ImageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	assertTags(t, resp.Tags, []string{"animal", "cat"}, "response")

	// Verify tags are persisted in database
	img, err := ctx.a.imageService.GetByName("tagged.png")
	if err != nil {
		t.Fatalf("failed to get image from DB: %v", err)
	}
	dbTags := models.TagsToStrings(img.Tags)
	assertTags(t, dbTags, []string{"animal", "cat"}, "database")

	ctx.assertStorageCount(2)
}

func TestPostImage_DuplicateName(t *testing.T) {
	ctx := newTestUploadContext(t)

	req1 := buildUploadRequest(t, "/upload", `{"name":"dup.png"}`, "dup.png", "image/png", makeTestPNG(t, 5, 5))
	req1 = withTestUser(req1, 1)
	rec1 := httptest.NewRecorder()
	ctx.r.ServeHTTP(rec1, req1)

	req2 := buildUploadRequest(t, "/upload", `{"name":"dup.png"}`, "dup.png", "image/png", makeTestPNG(t, 5, 5))
	req2 = withTestUser(req2, 1)
	rec2 := httptest.NewRecorder()
	ctx.r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for duplicate name, got %d: %s", rec2.Code, rec2.Body.String())
	}

	ctx.assertStorageCount(2)
}

func TestPostImage_DuplicateHash_DifferentName(t *testing.T) {
	ctx := newTestUploadContext(t)
	content := makeTestPNG(t, 7, 7)

	req1 := buildUploadRequest(t, "/upload", `{"name":"first.png"}`, "first.png", "image/png", content)
	req1 = withTestUser(req1, 1)
	rec1 := httptest.NewRecorder()
	ctx.r.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first upload failed: %d %s", rec1.Code, rec1.Body.String())
	}

	req2 := buildUploadRequest(t, "/upload", `{"name":"second.png"}`, "second.png", "image/png", content)
	req2 = withTestUser(req2, 1)
	rec2 := httptest.NewRecorder()
	ctx.r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("expected duplicate-hash upload to fail, got %d: %s", rec2.Code, rec2.Body.String())
	}

	ctx.assertStorageCount(2)
}

func TestPostImage_StorageUploadFails_NoDBRowAndStorageImageLeft(t *testing.T) {
	ctx := newTestUploadContext(t)
	ctx.store.FailUploadOn = func(key string) error {
		if !strings.Contains(key, "thumb") {
			return fmt.Errorf("simulated storage failure on main image")
		}
		return nil
	}

	req := buildUploadRequest(t, "/upload", `{"name":"y.png"}`, "y.png", "image/png", makeTestPNG(t, 5, 5))
	req = withTestUser(req, 1)
	rec := httptest.NewRecorder()
	ctx.r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}

	query := models.NewImageQueryBuilder().Prefix("dog_duplicate.png.png").Build()

	ctx.assertImageCount(query, 0)
	ctx.assertStorageCount(0)
}

func TestPostImage_StorageThumbUploadFails_NoDBRowAndStorageImageLeft(t *testing.T) {
	ctx := newTestUploadContext(t)
	ctx.store.FailUploadOn = func(key string) error {
		if strings.Contains(key, "thumb") {
			return fmt.Errorf("simulated storage failure on thumb")
		}
		return nil
	}

	req := buildUploadRequest(t, "/upload", `{"name":"y.png"}`, "y.png", "image/png", makeTestPNG(t, 5, 5))
	req = withTestUser(req, 1)
	rec := httptest.NewRecorder()
	ctx.r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}

	query := models.NewImageQueryBuilder().Prefix("dog_duplicate.png.png").Build()

	ctx.assertImageCount(query, 0)
	ctx.assertStorageCount(0)
}
