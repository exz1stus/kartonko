package api

import (
	"net/http"
	"net/http/httptest"
	"server/internal/storage"
	"testing"
)

func buildDeleteByQueryRequest(t *testing.T, url, queryString string) *http.Request {
	t.Helper()

	resourceUrl := url + "?" + queryString
	req := httptest.NewRequest(http.MethodDelete, resourceUrl, http.NoBody)
	return req
}

func TestDeleteImageByQuery(t *testing.T) {
	tests := []struct {
		name            string
		query           string
		userID          uint64
		wantStatus      int
		expectedDeleted int
	}{
		{
			name:            "moderator delete by name success",
			query:           `name=ima`,
			userID:          1,
			wantStatus:      http.StatusOK,
			expectedDeleted: 2, // image.png, image2.png (both start with "ima")
		},
		{
			name:            "moderator delete by tag success",
			query:           `tags=["dog"]`,
			userID:          1,
			wantStatus:      http.StatusOK,
			expectedDeleted: 2, // image2.png, user2_image.png (both have tag "dog")
		},
		{
			name:            "moderator delete by user id success",
			query:           `user_id=2`,
			userID:          1,
			wantStatus:      http.StatusOK,
			expectedDeleted: 1, // user2_image.png (owned by user 2)
		},
		{
			name:            "user full ownership delete success",
			query:           `user_id=2`,
			userID:          2,
			wantStatus:      http.StatusOK,
			expectedDeleted: 1, // user2_image.png (owned by user 2)
		},
		{
			name:            "user mixed ownership delete failure",
			query:           `tags=["dog"]`,
			userID:          2,
			wantStatus:      http.StatusForbidden,
			expectedDeleted: 0, // should fail, no images deleted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAPI(t)
			r := newTestRouter(a)
			store := a.storage.(storage.TestStorage)

			mod, err := a.userService.GetByID(1)
			if err != nil {
				t.Fatalf("failed getting moderator user")
			}
			tags := []string{"dog"}
			seedTestTags(t, a.tagService, tags, mod)

			seedImageUsingRequest(t, r, `{"name": "image.png"}`, "image.png", makeTestPNG(t, 5, 5), 1)
			seedImageUsingRequest(t, r, `{"name": "image2.png", "tags": ["dog"]}`, "image2.png", makeTestPNG(t, 7, 5), 1)
			seedImageUsingRequest(t, r, `{"name": "user2_image.png", "tags": ["dog"]}`, "user2_image.png", makeTestPNG(t, 5, 6), 2)

			initialCount64, err := a.imageService.Count(nil)
			initialCount := int(initialCount64)
			if err != nil {
				t.Fatal("failed to count test initial images count")
			}

			req := buildDeleteByQueryRequest(t, "/image", tt.query)

			if tt.userID != 0 {
				req = withTestUser(req, tt.userID)
			}

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			afterStoreCount, err := store.Count("")
			if err != nil {
				t.Errorf("failed retrieving store images count")
			}

			afterDbCount, err := a.imageService.Count(nil)
			if err != nil {
				t.Errorf("failed retrieving db images count")
			}

			expectedRemaining := initialCount - tt.expectedDeleted

			if afterDbCount != int64(expectedRemaining) {
				t.Errorf("expected %d db rows remaining, got %d", expectedRemaining, afterDbCount)
			}

			expectedStoreRemaining := (initialCount - tt.expectedDeleted) * 2 // image + thumb
			if afterStoreCount != expectedStoreRemaining {
				t.Errorf("expected %d stored objects remaining, got %d", expectedStoreRemaining, afterStoreCount)
			}
		})
	}
}
