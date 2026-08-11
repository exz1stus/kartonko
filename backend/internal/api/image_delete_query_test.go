package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tutil "server/internal/testutil/testing"
)

func deleteImagesByQuery(ctx *tutil.TestContext, query string, userID uint64) *httptest.ResponseRecorder {
	ctx.T.Helper()
	url := "/image"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodDelete, url, http.NoBody)
	if userID != 0 {
		req = tutil.WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec, req)
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
			ctx, cleanup := setupContext(t)
			defer cleanup()

			// Seed test images
			ctx.SeedImage(`{"name": "image.png"}`, "image.png", tutil.MakeTestPNG(t, 5, 5), 1)
			ctx.SeedImage(`{"name": "image2.png", "tags": ["dog"]}`, "image2.png", tutil.MakeTestPNG(t, 7, 5), 1)
			ctx.SeedImage(`{"name": "user2_image.png", "tags": ["dog"]}`, "user2_image.png", tutil.MakeTestPNG(t, 5, 6), 2)

			initialCount64, err := ctx.ImageService.Count(nil)
			initialCount := int(initialCount64)
			if err != nil {
				t.Fatal("failed to count test initial images count")
			}

			rec := deleteImagesByQuery(ctx, tt.query, tt.userID)

			ctx.AssertStatus(rec, tt.wantStatus)

			afterStoreCount, err := ctx.Storage.Count("")
			if err != nil {
				t.Errorf("failed retrieving store images count: %v", err)
			}

			afterDbCount, err := ctx.ImageService.Count(nil)
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
