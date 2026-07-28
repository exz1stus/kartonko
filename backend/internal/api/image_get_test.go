package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"server/internal/api/dto"
	"testing"
)

func TestGetImageByName(t *testing.T) {
	ctx := newTestContext(t)

	img := ctx.seedImage(`{"name":"test.png","tags":["cat"]}`, "test.png", makeUniqueTestPNG(t, 10, 10, 1), 1)

	t.Run("success", func(t *testing.T) {
		rec, resp := ctx.getImageByName("test.png", 1)

		ctx.assertStatus(rec, http.StatusOK)

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

	t.Run("not found", func(t *testing.T) {
		rec, _ := ctx.getImageByName("nonexistent.png", 1)
		ctx.assertStatus(rec, http.StatusNotFound)
	})

	t.Run("public access allowed", func(t *testing.T) {
		rec, _ := ctx.getImageByName("test.png", 0)
		ctx.assertStatus(rec, http.StatusOK)
	})
}

func TestGetImageByHash(t *testing.T) {
	ctx := newTestContext(t)

	img := ctx.seedImage(`{"name":"hash_test.png","tags":["dog"]}`, "hash_test.png", makeUniqueTestPNG(t, 10, 10, 2), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/hash/"+img.Hash, nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.r.ServeHTTP(rec, req)

		ctx.assertStatus(rec, http.StatusOK)

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
		ctx.r.ServeHTTP(rec, req)

		ctx.assertStatus(rec, http.StatusNotFound)
	})
}

func TestGetImageByID(t *testing.T) {
	ctx := newTestContext(t)

	img := ctx.seedImage(`{"name":"id_test.png","tags":["bird"]}`, "id_test.png", makeUniqueTestPNG(t, 10, 10, 3), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/image/id/%d", img.ID), nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.r.ServeHTTP(rec, req)

		ctx.assertStatus(rec, http.StatusOK)

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
		ctx.r.ServeHTTP(rec, req)

		ctx.assertStatus(rec, http.StatusBadRequest)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/id/999999", nil)
		req = withTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.r.ServeHTTP(rec, req)

		ctx.assertStatus(rec, http.StatusNotFound)
	})
}

func TestGetRawImageByName(t *testing.T) {
	ctx := newTestContext(t)

	_ = ctx.seedImage(`{"name":"raw_test.png","tags":["cat"]}`, "raw_test.png", makeUniqueTestPNG(t, 10, 10, 4), 1)

	t.Run("success", func(t *testing.T) {
		rec := ctx.getRawImage("raw_test.png", 1)

		ctx.assertStatus(rec, http.StatusOK)

		contentType := rec.Header().Get("Content-Type")
		if contentType != "image/png" {
			t.Errorf("expected content-type image/png, got %s", contentType)
		}

		if rec.Body.Len() == 0 {
			t.Error("expected non-empty image body")
		}
	})

	t.Run("not found", func(t *testing.T) {
		rec := ctx.getRawImage("nonexistent.png", 1)
		ctx.assertStatus(rec, http.StatusNotFound)
	})
}

func TestGetRawThumbnailByName(t *testing.T) {
	ctx := newTestContext(t)

	_ = ctx.seedImage(`{"name":"thumb_test.png","tags":["dog"]}`, "thumb_test.png", makeUniqueTestPNG(t, 10, 10, 5), 1)

	t.Run("success", func(t *testing.T) {
		rec := ctx.getThumbnail("thumb_test.png", 1)

		ctx.assertStatus(rec, http.StatusOK)

		contentType := rec.Header().Get("Content-Type")
		if contentType != "image/jpeg" {
			t.Errorf("expected content-type image/jpeg, got %s", contentType)
		}

		if rec.Body.Len() == 0 {
			t.Error("expected non-empty thumbnail body")
		}
	})

	t.Run("not found", func(t *testing.T) {
		rec := ctx.getThumbnail("nonexistent.png", 1)
		ctx.assertStatus(rec, http.StatusNotFound)
	})
}