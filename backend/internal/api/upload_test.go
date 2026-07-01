package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"server/internal/models"
	"strings"
	"testing"
)

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
			content:    func(t *testing.T) []byte { return makeTestPNG(t, 10, 10) },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing name in metadata",
			metadata:   `{"name":"","tags":[]}`,
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
			metadata:   `{"name":"doc.pdf","tags":[]}`,
			filename:   "doc.pdf",
			content:    func(t *testing.T) []byte { return []byte("%PDF-1.4") },
			mimeType:   "application/pdf",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "corrupt image bytes with valid mime",
			metadata:   `{"name":"broken.png","tags":[]}`,
			filename:   "broken.png",
			content:    func(t *testing.T) []byte { return []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00} },
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "unauthenticated request",
			metadata:   `{"name":"noauth.png","tags":[]}`,
			filename:   "noauth.png",
			content:    func(t *testing.T) []byte { return makeTestPNG(t, 5, 5) },
			mimeType:   "image/png",
			userID:     0,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "no file attached",
			metadata:   `{"name":"nofile.png","tags":[]}`,
			filename:   "", // triggers no file part
			content:    func(t *testing.T) []byte { return nil },
			mimeType:   "",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "tags are not added",
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
			a, store := newTestAPI(t)
			r := newTestRouter(a)

			tags := []string{"animal"}
			seedTestTags(t, a.models.Tags, tags)

			req := buildUploadRequest(t, "/upload", tt.metadata, tt.filename, tt.mimeType, tt.content(t))
			if tt.userID != 0 {
				req = withTestUser(req, tt.userID)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus == http.StatusOK && store.Count() != 2 {
				t.Errorf("expected 2 stored objects (image+thumb), got %d", store.Count())
			}
		})
	}
}

func TestPostImage_DuplicateName(t *testing.T) {
	a, store := newTestAPI(t)

	r := newTestRouter(a)
	req1 := buildUploadRequest(t, "/upload", `{"name":"dup.png","tags":[]}`, "dup.png", "image/png", makeTestPNG(t, 5, 5))
	req1 = withTestUser(req1, 1)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	req2 := buildUploadRequest(t, "/upload", `{"name":"dup.png","tags":[]}`, "dup.png", "image/png", makeTestPNG(t, 5, 5))
	req2 = withTestUser(req2, 1)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for duplicate name, got %d: %s", rec2.Code, rec2.Body.String())
	}

	// only the original upload's objects should exist
	if store.Count() != 2 {
		t.Errorf("expected no new storage objects from rejected duplicate, got %d", store.Count())
	}
}

func TestPostImage_DuplicateHash_DifferentName(t *testing.T) {
	a, store := newTestAPI(t)
	content := makeTestPNG(t, 7, 7)

	r := newTestRouter(a)

	req1 := buildUploadRequest(t, "/upload", `{"name":"first.png","tags":[]}`, "first.png", "image/png", content)
	req1 = withTestUser(req1, 1)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first upload failed: %d %s", rec1.Code, rec1.Body.String())
	}

	req2 := buildUploadRequest(t, "/upload", `{"name":"second.png","tags":[]}`, "second.png", "image/png", content)
	req2 = withTestUser(req2, 1)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusInternalServerError {
		t.Fatalf("expected duplicate-hash upload to fail, got %d: %s", rec2.Code, rec2.Body.String())
	}
	if store.Count() != 2 {
		t.Errorf("expected only first upload's objects to exist, got %d", store.Count())
	}
}

func TestPostImage_StorageUploadFails_NoDBRowAndStorageImageLeft(t *testing.T) {
	a, store := newTestAPI(t)
	store.FailUploadOn = func(key string) error {
		if !strings.Contains(key, "thumb") {
			return fmt.Errorf("simulated storage failure on main image")
		}
		return nil
	}

	r := newTestRouter(a)
	req := buildUploadRequest(t, "/upload", `{"name":"y.png","tags":[]}`, "y.png", "image/png", makeTestPNG(t, 5, 5))
	req = withTestUser(req, 1)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	var count int64
	a.models.Images.Db.Model(&models.ImageMetadata{}).Where("filename = ?", "y.png").Count(&count)
	if count != 0 {
		t.Errorf("expected no DB row after rollback, found %d", count)
	}
	if store.Count() != 0 {
		t.Errorf("expected no storage objects, found %d", store.Count())
	}
}

func TestPostImage_StorageThumbUploadFails_NoDBRowAndStorageImageLeft(t *testing.T) {
	a, store := newTestAPI(t)
	store.FailUploadOn = func(key string) error {
		if strings.Contains(key, "thumb") {
			return fmt.Errorf("simulated storage failure on thumb")
		}
		return nil
	}

	r := newTestRouter(a)
	req := buildUploadRequest(t, "/upload", `{"name":"y.png","tags":[]}`, "y.png", "image/png", makeTestPNG(t, 5, 5))
	req = withTestUser(req, 1)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}

	var count int64
	a.models.Images.Db.Model(&models.ImageMetadata{}).Where("filename = ?", "y.png").Count(&count)
	if count != 0 {
		t.Errorf("expected no DB row after rollback, found %d", count)
	}
	if store.Count() != 0 {
		t.Errorf("expected no storage objects, found %d", store.Count())
	}
}
