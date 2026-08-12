package image_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tutil "server/internal/testutil/testing"
)

func buildDeleteRequest(t *testing.T, url, filename string) *http.Request {
	t.Helper()
	resourceUrl := url + "/" + filename
	req := httptest.NewRequest(http.MethodDelete, resourceUrl, http.NoBody)
	return req
}

func TestDeleteImage(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		userID     uint64
		wantStatus int
	}{
		{
			"delete success",
			"image.png",
			2,
			http.StatusOK,
		},
		{
			"delete success moderator user but not an owner",
			"image.png",
			1,
			http.StatusOK,
		},
		{
			"delete forbidden user is not an owner or a moderator",
			"image.png",
			3,
			http.StatusForbidden,
		},
		{
			"delete failure user unauthorized",
			"image.png",
			0,
			http.StatusUnauthorized,
		},
		{
			"delete failure image doesn't exist",
			"image2.png",
			2,
			http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cleanup := setupContext(t)
			defer cleanup()

			// Seed the image
			ctx.SeedImage(`{"name": "image.png"}`, "image.png", tutil.MakeTestPNG(t, 5, 5), 2)

			storeCount, err := ctx.Storage.List(t.Context(), "")
			if err != nil {
				t.Errorf("failed retrieving store images count: %v", err)
			}

			dbCount, err := ctx.ImageService.Count(nil)
			if err != nil {
				t.Errorf("failed retrieving images count: %v", err)
			}

			req := buildDeleteRequest(t, "/image", tt.filename)
			if tt.userID != 0 {
				req = tutil.WithTestUser(req, tt.userID)
			}
			rec := httptest.NewRecorder()
			ctx.Router.ServeHTTP(rec, req)

			ctx.AssertStatus(rec, tt.wantStatus)

			if tt.wantStatus != http.StatusOK {
				ctx.AssertStorageCount(len(storeCount))
				ctx.AssertImageCount(nil, dbCount)

				return
			}

			ctx.AssertStorageCount(0)
			ctx.AssertImageCount(nil, 0)
		})
	}
}
