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
			"delete succes moderator user but not an owner",
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
			"delete failure user unathorized",
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
			a, store := newTestAPI(t)
			r := newTestRouter(a)

			seedImageUsingRequest(t, r, `{"name": "image.png"}`, "image.png", makeTestPNG(t, 5, 5), 2)

			storeCount := store.Count()
			dbCount, err := a.models.Images.GetImageCount()

			if err != nil {
				t.Errorf("failed retrieving images count")
			}

			req := buildDeleteRequest(t, "/image", tt.filename)

			if tt.userID != 0 {
				req = withTestUser(req, tt.userID)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			afterStoreCount := store.Count()
			afterDbCount, err := a.models.Images.GetImageCount()

			if err != nil {
				t.Errorf("failed retrieving images count")
			}

			if tt.wantStatus != http.StatusOK {
				if afterStoreCount != storeCount {
					t.Errorf("expected %d stored objects (image and thumb), got %d", storeCount, afterStoreCount)
				}

				if afterDbCount != dbCount {
					t.Errorf("expected %d db row, got %d", dbCount, afterDbCount)
				}

				return
			}

			if afterStoreCount != 0 {
				t.Errorf("expected 0 stored objects (image and thumb deleted), got %d", afterStoreCount)
			}

			if afterDbCount != 0 {
				t.Errorf("expected 0 db rows, got %d", afterDbCount)
			}
		})
	}
}
