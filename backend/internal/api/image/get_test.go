package image_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"server/internal/image"
	tutil "server/internal/testutil/testing"
)

func TestGetImageByName(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	img := ctx.SeedImage(`{"name":"test.png","tags":["cat"]}`, "test.png", tutil.MakeUniqueTestPNG(t, 10, 10, 1), 1)

	t.Run("success", func(t *testing.T) {
		rec, resp := ctx.GetImageByName("test.png", 1)

		ctx.AssertStatus(rec, http.StatusOK)

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
		rec, _ := ctx.GetImageByName("nonexistent.png", 1)
		ctx.AssertStatus(rec, http.StatusNotFound)
	})

	t.Run("public access allowed", func(t *testing.T) {
		rec, _ := ctx.GetImageByName("test.png", 0)
		ctx.AssertStatus(rec, http.StatusOK)
	})
}

func TestGetImageByHash(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	img := ctx.SeedImage(`{"name":"hash_test.png","tags":["dog"]}`, "hash_test.png", tutil.MakeUniqueTestPNG(t, 10, 10, 2), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/hash/"+img.Hash, nil)
		req = tutil.WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusOK)

		var resp image.ImageResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if resp.Hash != img.Hash {
			t.Errorf("expected hash %s, got %s", img.Hash, resp.Hash)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/hash/nonexistenthash", nil)
		req = tutil.WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusNotFound)
	})
}

func TestGetImageByID(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	img := ctx.SeedImage(`{"name":"id_test.png","tags":["bird"]}`, "id_test.png", tutil.MakeUniqueTestPNG(t, 10, 10, 3), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/image/id/%d", img.ID), nil)
		req = tutil.WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusOK)

		var resp image.ImageResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if resp.ID != img.ID {
			t.Errorf("expected ID %d, got %d", img.ID, resp.ID)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/id/notanumber", nil)
		req = tutil.WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusBadRequest)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/id/999999", nil)
		req = tutil.WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusNotFound)
	})
}

func TestGetRawImageByName(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	_ = ctx.SeedImage(`{"name":"raw_test.png","tags":["cat"]}`, "raw_test.png", tutil.MakeUniqueTestPNG(t, 10, 10, 4), 1)

	t.Run("success", func(t *testing.T) {
		rec := ctx.GetRawImage("raw_test.png", 1)

		ctx.AssertStatus(rec, http.StatusOK)

		contentType := rec.Header().Get("Content-Type")
		if contentType != "image/png" {
			t.Errorf("expected content-type image/png, got %s", contentType)
		}

		if rec.Body.Len() == 0 {
			t.Error("expected non-empty image body")
		}
	})

	t.Run("not found", func(t *testing.T) {
		rec := ctx.GetRawImage("nonexistent.png", 1)
		ctx.AssertStatus(rec, http.StatusNotFound)
	})
}

func TestGetRawThumbnailByName(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	_ = ctx.SeedImage(`{"name":"thumb_test.png","tags":["dog"]}`, "thumb_test.png", tutil.MakeUniqueTestPNG(t, 10, 10, 5), 1)

	t.Run("success", func(t *testing.T) {
		rec := ctx.GetThumbnail("thumb_test.png", 1)

		ctx.AssertStatus(rec, http.StatusOK)

		if rec.Body.Len() == 0 {
			t.Error("expected non-empty thumbnail body")
		}
	})

	t.Run("not found", func(t *testing.T) {
		rec := ctx.GetThumbnail("nonexistent.png", 1)
		ctx.AssertStatus(rec, http.StatusNotFound)
	})
}
