package http_integration_tests

import (
	"net/http"
	"net/http/httptest"
	imgapi "server/internal/api/image"
	"server/internal/testutil"
	"testing"
)

func deleteImagesByQuery(ctx *TestContext, query string, userID uint64) *httptest.ResponseRecorder {
	ctx.T.Helper()
	url := "/image"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodDelete, url, http.NoBody)
	if userID != 0 {
		req = WithTestUser(req, userID)
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

			ctx.SeedTags("dog")
			images := []struct {
				request imgapi.ImagePostRequest
				userID  uint64
			}{
				{
					request: imgapi.ImagePostRequest{
						Name: "image.png",
					},
					userID: 1,
				},
				{
					request: imgapi.ImagePostRequest{
						Name: "image2.png",
						Tags: []string{"dog"},
					},
					userID: 1,
				},
				{
					request: imgapi.ImagePostRequest{
						Name: "user2_image.png",
						Tags: []string{"dog"},
					},
					userID: 2,
				},
			}

			for i, img := range images {
				ctx.SeedImage(img.request, testutil.MakeUniqueTestPNG(ctx.T, 10, 10, i), img.userID)
			}

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
