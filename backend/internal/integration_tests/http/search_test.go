package http_integration_tests

// import (
// 	"fmt"
// 	"net/http"
// 	"testing"

// 	"server/internal/image"
// 	"server/internal/testutil"
// )

// func TestGetImagesByQuery(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		setupImages   func(ctx *TestContext)
// 		query         string
// 		userID        uint64
// 		wantStatus    int
// 		expectedCount int
// 		checkResponse func(t *testing.T, images []image.ImageResponse)
// 	}{
// 		{
// 			name: "success - no filters",
// 			setupImages: func(ctx *TestContext) {
// 				ctx.SeedImage(`{"name":"img1.png","tags":["cat"]}`, "img1.png", testutil.MakeUniqueTestPNG(t, 10, 10, 100), 1)
// 				ctx.SeedImage(`{"name":"img2.png","tags":["dog"]}`, "img2.png", testutil.MakeUniqueTestPNG(t, 10, 10, 101), 1)
// 				ctx.SeedImage(`{"name":"img3.png","tags":["cat","dog"]}`, "img3.png", testutil.MakeUniqueTestPNG(t, 10, 10, 102), 2)
// 			},
// 			query:         "",
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 3,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 3 {
// 					t.Errorf("expected 3 images, got %d", len(images))
// 				}
// 			},
// 		},
// 		{
// 			name: "success - prefix filter",
// 			setupImages: func(ctx *TestContext) {
// 				ctx.SeedImage(`{"name":"cat_image.png","tags":["cat"]}`, "cat_image.png", testutil.MakeUniqueTestPNG(t, 10, 10, 110), 1)
// 				ctx.SeedImage(`{"name":"dog_image.png","tags":["dog"]}`, "dog_image.png", testutil.MakeUniqueTestPNG(t, 10, 10, 111), 1)
// 				ctx.SeedImage(`{"name":"bird_image.png","tags":["bird"]}`, "bird_image.png", testutil.MakeUniqueTestPNG(t, 10, 10, 112), 1)
// 			},
// 			query:         "prefix=cat",
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 1,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 1 {
// 					t.Errorf("expected 1 image, got %d", len(images))
// 				}
// 				if images[0].Filename != "cat_image.png" {
// 					t.Errorf("expected cat_image.png, got %s", images[0].Filename)
// 				}
// 			},
// 		},
// 		{
// 			name: "success - tag filter",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"img1.png","tags":["cat"]}`, "img1.png", testutil.MakeUniqueTestPNG(t, 10, 10, 120), 1)
// 				ctx.SeedImage(`{"name":"img2.png","tags":["dog"]}`, "img2.png", testutil.MakeUniqueTestPNG(t, 10, 10, 121), 1)
// 				ctx.SeedImage(`{"name":"img3.png","tags":["cat","dog"]}`, "img3.png", testutil.MakeUniqueTestPNG(t, 10, 10, 122), 1)
// 			},
// 			query:         `tags=["cat"]`,
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 2,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 2 {
// 					t.Errorf("expected 2 images with cat tag, got %d", len(images))
// 				}
// 				for _, img := range images {
// 					hasCat := false
// 					for _, tag := range img.Tags {
// 						if tag == "cat" {
// 							hasCat = true
// 							break
// 						}
// 					}
// 					if !hasCat {
// 						t.Errorf("image %s should have cat tag", img.Filename)
// 					}
// 				}
// 			},
// 		},
// 		{
// 			name: "success - multiple tags filter (AND)",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"img1.png","tags":["cat"]}`, "img1.png", testutil.MakeUniqueTestPNG(t, 10, 10, 130), 1)
// 				ctx.SeedImage(`{"name":"img2.png","tags":["dog"]}`, "img2.png", testutil.MakeUniqueTestPNG(t, 10, 10, 131), 1)
// 				ctx.SeedImage(`{"name":"img3.png","tags":["cat","dog"]}`, "img3.png", testutil.MakeUniqueTestPNG(t, 10, 10, 132), 1)
// 			},
// 			query:         `tags=["cat","dog"]`,
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 1,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 1 {
// 					t.Errorf("expected 1 image with both cat and dog tags, got %d", len(images))
// 				}
// 				if images[0].Filename != "img3.png" {
// 					t.Errorf("expected img3.png, got %s", images[0].Filename)
// 				}
// 			},
// 		},
// 		{
// 			name: "success - user filter",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"user1_img.png","tags":["cat"]}`, "user1_img.png", testutil.MakeUniqueTestPNG(t, 10, 10, 140), 1)
// 				ctx.SeedImage(`{"name":"user2_img.png","tags":["dog"]}`, "user2_img.png", testutil.MakeUniqueTestPNG(t, 10, 10, 141), 2)
// 				ctx.SeedImage(`{"name":"user1_img2.png","tags":["bird"]}`, "user1_img2.png", testutil.MakeUniqueTestPNG(t, 10, 10, 142), 1)
// 			},
// 			query:         "user_id=1",
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 2,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 2 {
// 					t.Errorf("expected 2 images for user 1, got %d", len(images))
// 				}
// 				for _, img := range images {
// 					if img.UserID != 1 {
// 						t.Errorf("expected user_id 1, got %d for image %s", img.UserID, img.Filename)
// 					}
// 				}
// 			},
// 		},
// 		{
// 			name: "success - username filter",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"alice_img.png","tags":["cat"]}`, "alice_img.png", testutil.MakeUniqueTestPNG(t, 10, 10, 150), 4)
// 				ctx.SeedImage(`{"name":"bob_img.png","tags":["dog"]}`, "bob_img.png", testutil.MakeUniqueTestPNG(t, 10, 10, 151), 3)
// 			},
// 			query:         "username=alice",
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 1,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 1 {
// 					t.Errorf("expected 1 image for alice, got %d", len(images))
// 				}
// 				if images[0].Filename != "alice_img.png" {
// 					t.Errorf("expected alice_img.png, got %s", images[0].Filename)
// 				}
// 			},
// 		},
// 		{
// 			name: "success - pagination with limit",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				for i := 1; i <= 5; i++ {
// 					ctx.SeedImage(fmt.Sprintf(`{"name":"page_img%d.png","tags":["test"]}`, i), fmt.Sprintf("page_img%d.png", i), testutil.MakeUniqueTestPNG(t, 10, 10, 160+i), 1)
// 				}
// 			},
// 			query:         "limit=2",
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 2,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 2 {
// 					t.Errorf("expected 2 images with limit=2, got %d", len(images))
// 				}
// 			},
// 		},
// 		{
// 			name: "success - cursor pagination",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				for i := 1; i <= 5; i++ {
// 					ctx.SeedImage(fmt.Sprintf(`{"name":"cursor_img%d.png","tags":["test"]}`, i), fmt.Sprintf("cursor_img%d.png", i), testutil.MakeUniqueTestPNG(t, 10, 10, 170+i), 1)
// 				}
// 			},
// 			query:         "limit=2&cursor=2",
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 2,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 2 {
// 					t.Errorf("expected 2 images with cursor=2&limit=2, got %d", len(images))
// 				}
// 			},
// 		},
// 		{
// 			name: "success - combined filters",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"combo1.png","tags":["cat","animal"]}`, "combo1.png", testutil.MakeUniqueTestPNG(t, 10, 10, 180), 1)
// 				ctx.SeedImage(`{"name":"combo2.png","tags":["dog","animal"]}`, "combo2.png", testutil.MakeUniqueTestPNG(t, 10, 10, 181), 1)
// 				ctx.SeedImage(`{"name":"combo3.png","tags":["cat"]}`, "combo3.png", testutil.MakeUniqueTestPNG(t, 10, 10, 182), 2)
// 			},
// 			query:         `name=combo&tags=["animal"]&user_id=1`,
// 			userID:        1,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 2,
// 			checkResponse: func(t *testing.T, images []image.ImageResponse) {
// 				if len(images) != 2 {
// 					t.Errorf("expected 2 images, got %d", len(images))
// 				}
// 			},
// 		},
// 		{
// 			name: "public access allowed",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"test.png","tags":[]}`, "test.png", testutil.MakeUniqueTestPNG(t, 10, 10, 190), 1)
// 			},
// 			query:         "",
// 			userID:        0,
// 			wantStatus:    http.StatusOK,
// 			expectedCount: 1,
// 		},
// 		{
// 			name: "bad request - invalid tags json",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"test.png","tags":[]}`, "test.png", testutil.MakeUniqueTestPNG(t, 10, 10, 191), 1)
// 			},
// 			query:      `tags=[invalid]`,
// 			userID:     1,
// 			wantStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "bad request - invalid cursor",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"test.png","tags":[]}`, "test.png", testutil.MakeUniqueTestPNG(t, 10, 10, 192), 1)
// 			},
// 			query:      "cursor=abc",
// 			userID:     1,
// 			wantStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "bad request - invalid limit",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"test.png","tags":[]}`, "test.png", testutil.MakeUniqueTestPNG(t, 10, 10, 193), 1)
// 			},
// 			query:      "limit=xyz",
// 			userID:     1,
// 			wantStatus: http.StatusBadRequest,
// 		},
// 		{
// 			name: "bad request - invalid user_id",
// 			setupImages: func(ctx *testutil.TestContext) {
// 				ctx.SeedImage(`{"name":"test.png","tags":[]}`, "test.png", testutil.MakeUniqueTestPNG(t, 10, 10, 194), 1)
// 			},
// 			query:      "user_id=abc",
// 			userID:     1,
// 			wantStatus: http.StatusBadRequest,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctx, cleanup := setupContext(t)
// 			defer cleanup()
// 			tt.setupImages(ctx)

// 			rec, images := ctx.QueryImagesWithResponse(tt.query, tt.userID)

// 			ctx.AssertStatus(rec, tt.wantStatus)

// 			if tt.wantStatus == http.StatusOK {
// 				if len(images) != tt.expectedCount {
// 					t.Errorf("expected %d images, got %d", tt.expectedCount, len(images))
// 				}

// 				if tt.checkResponse != nil {
// 					tt.checkResponse(t, images)
// 				}
// 			}
// 		})
// 	}
// }

// func TestGetImagesByQuery_EmptyResult(t *testing.T) {
// 	ctx, cleanup := setupContext(t)
// 	defer cleanup()

// 	rec, images := ctx.QueryImagesWithResponse("", 1)

// 	ctx.AssertStatus(rec, http.StatusOK)

// 	if len(images) != 0 {
// 		t.Errorf("expected 0 images, got %d", len(images))
// 	}
// }
