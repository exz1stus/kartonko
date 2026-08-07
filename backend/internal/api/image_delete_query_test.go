package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func (c *testContext) deleteImagesByQuery(query string, userID uint64) *httptest.ResponseRecorder {
	c.t.Helper()
	url := "/image"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodDelete, url, http.NoBody)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
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
			query:           `prefix=ima`,
			userID:          1,
			wantStatus:      http.StatusOK,
			expectedDeleted: 2,
		},
		{
			name:            "moderator delete by tag success",
			query:           `tags=["dog"]`,
			userID:          1,
			wantStatus:      http.StatusOK,
			expectedDeleted: 2,
		},
		{
			name:            "moderator delete by user id success",
			query:           `user_id=2`,
			userID:          1,
			wantStatus:      http.StatusOK,
			expectedDeleted: 1,
		},
		{
			name:            "user full ownership delete success",
			query:           `user_id=2`,
			userID:          2,
			wantStatus:      http.StatusOK,
			expectedDeleted: 1,
		},
		{
			name:            "user mixed ownership delete failure",
			query:           `tags=["dog"]`,
			userID:          2,
			wantStatus:      http.StatusForbidden,
			expectedDeleted: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContext(t)

			// Seed test images
			ctx.seedImage(`{"name": "image.png"}`, "image.png", makeTestPNG(t, 5, 5), 1)
			ctx.seedImage(`{"name": "image2.png", "tags": ["dog"]}`, "image2.png", makeTestPNG(t, 7, 5), 1)
			ctx.seedImage(`{"name": "user2_image.png", "tags": ["dog"]}`, "user2_image.png", makeTestPNG(t, 5, 6), 2)

			initialCount64, err := ctx.a.imageService.Count(nil)
			initialCount := int(initialCount64)
			if err != nil {
				t.Fatal("failed to count test initial images count")
			}

			rec := ctx.deleteImagesByQuery(tt.query, tt.userID)

			ctx.assertStatus(rec, tt.wantStatus)

			afterStoreCount, err := ctx.store.Count("")
			if err != nil {
				t.Errorf("failed retrieving store images count: %v", err)
			}

			afterDbCount, err := ctx.a.imageService.Count(nil)
			if err != nil {
				t.Errorf("failed retrieving db images count: %v", err)
			}

			expectedRemaining := initialCount - tt.expectedDeleted

			if afterDbCount != int64(expectedRemaining) {
				t.Errorf("expected %d db rows remaining, got %d", expectedRemaining, afterDbCount)
			}

			expectedStoreRemaining := (initialCount - tt.expectedDeleted) * 2
			if afterStoreCount != expectedStoreRemaining {
				t.Errorf("expected %d stored objects remaining, got %d", expectedStoreRemaining, afterStoreCount)
			}
		})
	}
}
