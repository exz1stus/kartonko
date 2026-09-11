package http_integration_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	imgapi "server/internal/api/image"
	"server/internal/testutil"

	"github.com/stretchr/testify/require"
)

func TestGetImageByName(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	tags := []string{"cat"}
	ctx.SeedTags(tags...)
	img := ctx.SeedImage(imgapi.ImagePostRequest{
		Name: "test.png",
		Tags: tags,
	}, testutil.MakeUniqueTestPNG(t, 10, 10, 1), 1)

	t.Run("success", func(t *testing.T) {
		rec, resp := ctx.GetImageByName("test.png", 1)

		ctx.AssertStatus(rec, http.StatusOK)

		require.Equal(ctx.T, img.ID, resp.ID)
		require.Equal(ctx.T, "test.png", resp.Filename)
		require.Equal(ctx.T, img.Hash, resp.Hash)
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

	tags := []string{"dog"}
	ctx.SeedTags(tags...)
	img := ctx.SeedImage(imgapi.ImagePostRequest{
		Name: "hash_test.png",
		Tags: tags,
	}, testutil.MakeUniqueTestPNG(t, 10, 10, 2), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/hash/"+img.Hash, nil)
		req = WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusOK)

		var resp imgapi.ImageResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		require.Equal(ctx.T, img.Hash, resp.Hash)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/hash/nonexistenthash", nil)
		req = WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusNotFound)
	})
}

func TestGetImageByID(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	tags := []string{"bird"}
	ctx.SeedTags(tags...)
	img := ctx.SeedImage(imgapi.ImagePostRequest{
		Name: "id_test.png",
		Tags: tags,
	}, testutil.MakeUniqueTestPNG(t, 10, 10, 3), 1)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/image/id/%d", img.ID), nil)
		req = WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusOK)

		var resp imgapi.ImageResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		require.Equal(ctx.T, img.ID, resp.ID)
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/id/notanumber", nil)
		req = WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusBadRequest)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/image/id/999999", nil)
		req = WithTestUser(req, 1)
		rec := httptest.NewRecorder()
		ctx.Router.ServeHTTP(rec, req)

		ctx.AssertStatus(rec, http.StatusNotFound)
	})
}

func TestGetRawImageByName(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	tags := []string{"cat"}
	ctx.SeedTags(tags...)
	ctx.SeedImage(imgapi.ImagePostRequest{
		Name: "raw_test.png",
		Tags: tags,
	}, testutil.MakeUniqueTestPNG(t, 10, 10, 4), 1)

	t.Run("success", func(t *testing.T) {
		rec := ctx.GetRawImage("raw_test.png", 1)

		ctx.AssertStatus(rec, http.StatusOK)

		contentType := rec.Header().Get("Content-Type")
		require.Equal(ctx.T, contentType, "image/png")
		require.Greater(ctx.T, rec.Body.Len(), 0, "expected non-empty image body")
	})

	t.Run("not found", func(t *testing.T) {
		rec := ctx.GetRawImage("nonexistent.png", 1)
		ctx.AssertStatus(rec, http.StatusNotFound)
	})
}

func TestGetRawThumbnailByName(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	tags := []string{"dog"}
	ctx.SeedTags(tags...)
	ctx.SeedImage(imgapi.ImagePostRequest{
		Name: "thumb_test.png",
		Tags: tags,
	}, testutil.MakeUniqueTestPNG(t, 10, 10, 5), 1)

	t.Run("success", func(t *testing.T) {
		rec := ctx.GetThumbnail("thumb_test.png", 1)
		ctx.AssertStatus(rec, http.StatusOK)
		require.Greater(ctx.T, rec.Body.Len(), 0, "expected non-empty thumb body")
	})

	t.Run("not found", func(t *testing.T) {
		rec := ctx.GetThumbnail("nonexistent.png", 1)
		ctx.AssertStatus(rec, http.StatusNotFound)
	})
}
