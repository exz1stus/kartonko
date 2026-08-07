package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func buildDeleteRequest(t *testing.T, url, filename string) *http.Request {
	t.Helper()
	resourceUrl := url + "/" + filename
	req := httptest.NewRequest(http.MethodDelete, resourceUrl, http.NoBody)
	return req
}

func (c *testContext) deleteImageByName(filename string, userID uint64) *httptest.ResponseRecorder {
	c.t.Helper()
	req := buildDeleteRequest(c.t, "/image", filename)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
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
			ctx := newTestContext(t)

			// Seed the image
			ctx.seedImage(`{"name": "image.png"}`, "image.png", makeTestPNG(t, 5, 5), 2)

			storeCount, err := ctx.store.Count("")
			if err != nil {
				t.Errorf("failed retrieving store images count: %v", err)
			}

			dbCount, err := ctx.a.imageService.Count(nil)
			if err != nil {
				t.Errorf("failed retrieving images count: %v", err)
			}

			rec := ctx.deleteImageByName(tt.filename, tt.userID)

			ctx.assertStatus(rec, tt.wantStatus)

			if tt.wantStatus != http.StatusOK {
				ctx.assertStorageCount(storeCount)
				ctx.assertImageCount(nil, dbCount)

				return
			}

			ctx.assertStorageCount(0)
			ctx.assertImageCount(nil, 0)
		})
	}
}
