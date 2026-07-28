package api

import (
	"net/http"
	"testing"
)

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

			afterStoreCount, err := ctx.store.Count("")
			if err != nil {
				t.Errorf("failed retrieving store images count: %v", err)
			}

			afterDbCount, err := ctx.a.imageService.Count(nil)
			if err != nil {
				t.Errorf("failed retrieving images count: %v", err)
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