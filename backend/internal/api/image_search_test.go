package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/storage"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestGetImagesByQuery(t *testing.T) {
	testTags := []string{"cat", "dog", "bird", "animal", "test"}

	tests := []struct {
		name          string
		setupImages   func(t *testing.T, r *gin.Engine, a *api)
		query         string
		userID        uint64
		wantStatus    int
		expectedCount int
		checkResponse func(t *testing.T, images []dto.ImageResponse)
	}{
		{
			name: "success - no filters",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"img1.png","tags":["cat"]}`, "img1.png", makeUniqueTestPNG(t, 10, 10, 100), 1)
				seedImage(t, r, `{"name":"img2.png","tags":["dog"]}`, "img2.png", makeUniqueTestPNG(t, 10, 10, 101), 1)
				seedImage(t, r, `{"name":"img3.png","tags":["cat","dog"]}`, "img3.png", makeUniqueTestPNG(t, 10, 10, 102), 2)
			},
			query:         "",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 3,
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 3 {
					t.Errorf("expected 3 images, got %d", len(images))
				}
			},
		},
		{
			name: "success - prefix filter",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"cat_image.png","tags":["cat"]}`, "cat_image.png", makeUniqueTestPNG(t, 10, 10, 110), 1)
				seedImage(t, r, `{"name":"dog_image.png","tags":["dog"]}`, "dog_image.png", makeUniqueTestPNG(t, 10, 10, 111), 1)
				seedImage(t, r, `{"name":"bird_image.png","tags":["bird"]}`, "bird_image.png", makeUniqueTestPNG(t, 10, 10, 112), 1)
			},
			query:         "name=cat",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 1,
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 1 {
					t.Errorf("expected 1 image, got %d", len(images))
				}
				if images[0].Filename != "cat_image.png" {
					t.Errorf("expected cat_image.png, got %s", images[0].Filename)
				}
			},
		},
		{
			name: "success - tag filter",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"img1.png","tags":["cat"]}`, "img1.png", makeUniqueTestPNG(t, 10, 10, 120), 1)
				seedImage(t, r, `{"name":"img2.png","tags":["dog"]}`, "img2.png", makeUniqueTestPNG(t, 10, 10, 121), 1)
				seedImage(t, r, `{"name":"img3.png","tags":["cat","dog"]}`, "img3.png", makeUniqueTestPNG(t, 10, 10, 122), 1)
			},
			query:         `tags=["cat"]`,
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2, // img1.png and img3.png have cat tag
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 2 {
					t.Errorf("expected 2 images with cat tag, got %d", len(images))
				}
				for _, img := range images {
					hasCat := false
					for _, tag := range img.Tags {
						if tag == "cat" {
							hasCat = true
							break
						}
					}
					if !hasCat {
						t.Errorf("image %s should have cat tag", img.Filename)
					}
				}
			},
		},
		{
			name: "success - multiple tags filter (AND)",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"img1.png","tags":["cat"]}`, "img1.png", makeUniqueTestPNG(t, 10, 10, 130), 1)
				seedImage(t, r, `{"name":"img2.png","tags":["dog"]}`, "img2.png", makeUniqueTestPNG(t, 10, 10, 131), 1)
				seedImage(t, r, `{"name":"img3.png","tags":["cat","dog"]}`, "img3.png", makeUniqueTestPNG(t, 10, 10, 132), 1)
			},
			query:         `tags=["cat","dog"]`,
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 1, // only img3.png has both tags
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 1 {
					t.Errorf("expected 1 image with both cat and dog tags, got %d", len(images))
				}
				if images[0].Filename != "img3.png" {
					t.Errorf("expected img3.png, got %s", images[0].Filename)
				}
			},
		},
		{
			name: "success - user filter",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"user1_img.png","tags":["cat"]}`, "user1_img.png", makeUniqueTestPNG(t, 10, 10, 140), 1)
				seedImage(t, r, `{"name":"user2_img.png","tags":["dog"]}`, "user2_img.png", makeUniqueTestPNG(t, 10, 10, 141), 2)
				seedImage(t, r, `{"name":"user1_img2.png","tags":["bird"]}`, "user1_img2.png", makeUniqueTestPNG(t, 10, 10, 142), 1)
			},
			query:         "user_id=1",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2, // user 1 owns 2 images
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 2 {
					t.Errorf("expected 2 images for user 1, got %d", len(images))
				}
				for _, img := range images {
					if img.UserID != 1 {
						t.Errorf("expected user_id 1, got %d for image %s", img.UserID, img.Filename)
					}
				}
			},
		},
		{
			name: "success - username filter",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				// Users are: mod(1), alice(2), bob(3)
				seedImage(t, r, `{"name":"alice_img.png","tags":["cat"]}`, "alice_img.png", makeUniqueTestPNG(t, 10, 10, 150), 2)
				seedImage(t, r, `{"name":"bob_img.png","tags":["dog"]}`, "bob_img.png", makeUniqueTestPNG(t, 10, 10, 151), 3)
			},
			query:         "username=alice",
			userID:        1, // moderator can query by username
			wantStatus:    http.StatusOK,
			expectedCount: 1,
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 1 {
					t.Errorf("expected 1 image for alice, got %d", len(images))
				}
				if images[0].Filename != "alice_img.png" {
					t.Errorf("expected alice_img.png, got %s", images[0].Filename)
				}
			},
		},
		{
			name: "success - pagination with cursor and limit",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				for i := 1; i <= 5; i++ {
					seedImage(t, r, fmt.Sprintf(`{"name":"page_img%d.png","tags":["test"]}`, i), fmt.Sprintf("page_img%d.png", i), makeUniqueTestPNG(t, 10, 10, 160+i), 1)
				}
			},
			query:         "limit=2",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2,
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 2 {
					t.Errorf("expected 2 images with limit=2, got %d", len(images))
				}
			},
		},
		{
			name: "success - cursor pagination",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				for i := 1; i <= 5; i++ {
					seedImage(t, r, fmt.Sprintf(`{"name":"cursor_img%d.png","tags":["test"]}`, i), fmt.Sprintf("cursor_img%d.png", i), makeUniqueTestPNG(t, 10, 10, 170+i), 1)
				}
			},
			query:         "limit=2&cursor=2",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2, // should get items 3 and 4 (0-indexed, cursor=2 skips first 2)
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 2 {
					t.Errorf("expected 2 images with cursor=2&limit=2, got %d", len(images))
				}
			},
		},
		{
			name: "success - combined filters",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"combo1.png","tags":["cat","animal"]}`, "combo1.png", makeUniqueTestPNG(t, 10, 10, 180), 1)
				seedImage(t, r, `{"name":"combo2.png","tags":["dog","animal"]}`, "combo2.png", makeUniqueTestPNG(t, 10, 10, 181), 1)
				seedImage(t, r, `{"name":"combo3.png","tags":["cat"]}`, "combo3.png", makeUniqueTestPNG(t, 10, 10, 182), 2)
			},
			query:         "name=combo&tags=[\"animal\"]&user_id=1",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2, // combo1 and combo2 (owned by user 1, have animal tag, name starts with combo)
			checkResponse: func(t *testing.T, images []dto.ImageResponse) {
				if len(images) != 2 {
					t.Errorf("expected 2 images, got %d", len(images))
				}
			},
		},
		{
			name: "public access allowed",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"test.png","tags":[]}`, "test.png", makeUniqueTestPNG(t, 10, 10, 190), 1)
			},
			query:         "",
			userID:        0, // no auth - public access allowed
			wantStatus:    http.StatusOK,
			expectedCount: 1,
		},
		{
			name: "bad request - invalid tags json",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"test.png","tags":[]}`, "test.png", makeUniqueTestPNG(t, 10, 10, 191), 1)
			},
			query:      `tags=[invalid]`,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid cursor",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"test.png","tags":[]}`, "test.png", makeUniqueTestPNG(t, 10, 10, 192), 1)
			},
			query:      "cursor=abc",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid limit",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"test.png","tags":[]}`, "test.png", makeUniqueTestPNG(t, 10, 10, 193), 1)
			},
			query:      "limit=xyz",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid user_id",
			setupImages: func(t *testing.T, r *gin.Engine, a *api) {
				seedImage(t, r, `{"name":"test.png","tags":[]}`, "test.png", makeUniqueTestPNG(t, 10, 10, 194), 1)
			},
			query:      "user_id=abc",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAPI(t)
			r := newTestRouter(a)

			// Clean database before each subtest
			if err := a.CleanTestDB(); err != nil {
				t.Fatalf("failed to clean test db: %v", err)
			}
			a.storage.(*storage.MockStorage).Reset()

			// Re-seed users FIRST (they were truncated)
			users := []models.User{
				{Model: gorm.Model{ID: 1}, Username: "mod", Privilege: models.Moderator, Email: "mod@test.local", ProviderID: "test-provider-1"},
				{Model: gorm.Model{ID: 2}, Username: "alice", Privilege: models.Unprivileged, Email: "alice@test.local", ProviderID: "test-provider-2"},
				{Model: gorm.Model{ID: 3}, Username: "bob", Privilege: models.Unprivileged, Email: "bob@test.local", ProviderID: "test-provider-3"},
			}
			seedTestUsers(t, a.userService, users)

			// Seed required tags
			mod, err := a.userService.GetByID(1)
			if err != nil {
				t.Fatalf("failed getting moderator user")
			}
			seedTestTags(t, a.tagService, testTags, mod)

			// Setup test images
			tt.setupImages(t, r, a)

			// Build request
			url := "/image"
			if tt.query != "" {
				url += "?" + tt.query
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tt.userID != 0 {
				req = withTestUser(req, tt.userID)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
				return
			}

			if tt.wantStatus == http.StatusOK {
				var images []dto.ImageResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &images); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if len(images) != tt.expectedCount {
					t.Errorf("expected %d images, got %d", tt.expectedCount, len(images))
				}

				if tt.checkResponse != nil {
					tt.checkResponse(t, images)
				}
			}
		})
	}
}

func TestGetImagesByQuery_EmptyResult(t *testing.T) {
	a := newTestAPI(t)
	r := newTestRouter(a)

	// No images in DB
	req := httptest.NewRequest(http.MethodGet, "/image", nil)
	req = withTestUser(req, 1)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		return
	}

	var images []dto.ImageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &images); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(images) != 0 {
		t.Errorf("expected 0 images, got %d", len(images))
	}
}
