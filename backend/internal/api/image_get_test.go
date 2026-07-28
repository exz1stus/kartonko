package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"server/internal/api/dto"
	"testing"

	"github.com/gin-gonic/gin"
)

// makeUniqueTestPNG creates a test PNG with unique dimensions to avoid duplicate hash errors
func makeUniqueTestPNG(t *testing.T, baseW, baseH int, unique int) []byte {
	t.Helper()
	// Use a larger range to ensure uniqueness
	return makeTestPNG(t, baseW+unique*10, baseH+unique*10)
}

// seedImage uploads an image and returns the response
func seedImage(t *testing.T, r *gin.Engine, metadata, filename string, content []byte, userID uint64) dto.ImageResponse {
	t.Helper()
	req := buildUploadRequest(t, "/upload", metadata, filename, "image/png", content)
	req = withTestUser(req, userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("seedImage(%s) failed: status=%d body=%s", filename, rec.Code, rec.Body.String())
	}

	var resp dto.ImageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return resp
}

func TestGetImageByName(t *testing.T) {
	a := newTestAPI(t)
	r := newTestRouter(a)

	// Seed required tags
	mod, err := a.userService.GetByID(1)
	if err != nil {
		t.Fatalf("failed getting moderator user")
	}
	seedTestTags(t, a.tagService, []string{"cat", "dog", "bird"}, mod)

	// Upload test image
	img := seedImage(t, r, `{"name":"test.png","tags":["cat"]}`, "test.png", makeUniqueTestPNG(t, 10, 10, 1), 1)

	// Test successful retrieval
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/test.png", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			return
		}

		var resp dto.ImageResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if resp.ID != img.ID {
			t.Errorf("expected ID %d, got %d", img.ID, resp.ID)
		}
		if resp.Filename != "test.png" {
			t.Errorf("expected filename test.png, got %s", resp.Filename)
		}
		if resp.Hash != img.Hash {
			t.Errorf("expected hash %s, got %s", img.Hash, resp.Hash)
		}
	})

	// Test not found
	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/nonexistent.png", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// Test unauthorized access - GET endpoints are public, so they should succeed even without auth
	t.Run("public access allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/test.png", nil)
		// No user header - should still work for public GET endpoints
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 for public access, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestGetImageByHash(t *testing.T) {
	a := newTestAPI(t)
	r := newTestRouter(a)

	// Seed required tags
	mod, err := a.userService.GetByID(1)
	if err != nil {
		t.Fatalf("failed getting moderator user")
	}
	seedTestTags(t, a.tagService, []string{"cat", "dog", "bird"}, mod)

	img := seedImage(t, r, `{"name":"hash_test.png","tags":["dog"]}`, "hash_test.png", makeUniqueTestPNG(t, 10, 10, 2), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/hash/"+img.Hash, nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			return
		}

		var resp dto.ImageResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if resp.Hash != img.Hash {
			t.Errorf("expected hash %s, got %s", img.Hash, resp.Hash)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/hash/nonexistenthash", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestGetImageByID(t *testing.T) {
	a := newTestAPI(t)
	r := newTestRouter(a)

	// Seed required tags
	mod, err := a.userService.GetByID(1)
	if err != nil {
		t.Fatalf("failed getting moderator user")
	}
	seedTestTags(t, a.tagService, []string{"cat", "dog", "bird"}, mod)

	img := seedImage(t, r, `{"name":"id_test.png","tags":["bird"]}`, "id_test.png", makeUniqueTestPNG(t, 10, 10, 3), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/image/id/%d", img.ID), nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			return
		}

		var resp dto.ImageResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if resp.ID != img.ID {
			t.Errorf("expected ID %d, got %d", img.ID, resp.ID)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/id/notanumber", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/id/999999", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestGetRawImageByName(t *testing.T) {
	a := newTestAPI(t)
	r := newTestRouter(a)

	// Seed required tags
	mod, err := a.userService.GetByID(1)
	if err != nil {
		t.Fatalf("failed getting moderator user")
	}
	seedTestTags(t, a.tagService, []string{"cat", "dog"}, mod)

	_ = seedImage(t, r, `{"name":"raw_test.png","tags":["cat"]}`, "raw_test.png", makeUniqueTestPNG(t, 10, 10, 4), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/raw/raw_test.png", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			return
		}

		// Verify content type
		contentType := rec.Header().Get("Content-Type")
		if contentType != "image/png" {
			t.Errorf("expected content-type image/png, got %s", contentType)
		}

		// Verify content length matches
		if rec.Body.Len() == 0 {
			t.Error("expected non-empty image body")
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/raw/nonexistent.png", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestGetRawThumbnailByName(t *testing.T) {
	a := newTestAPI(t)
	r := newTestRouter(a)

	// Seed required tags
	mod, err := a.userService.GetByID(1)
	if err != nil {
		t.Fatalf("failed getting moderator user")
	}
	seedTestTags(t, a.tagService, []string{"cat", "dog"}, mod)

	_ = seedImage(t, r, `{"name":"thumb_test.png","tags":["dog"]}`, "thumb_test.png", makeUniqueTestPNG(t, 10, 10, 5), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/thumb/thumb_test.png", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			return
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "image/jpeg" {
			t.Errorf("expected content-type image/jpeg, got %s", contentType)
		}

		if rec.Body.Len() == 0 {
			t.Error("expected non-empty thumbnail body")
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/thumb/nonexistent.png", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}
