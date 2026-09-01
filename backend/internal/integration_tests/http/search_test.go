package http_integration_tests

import (
	"fmt"
	"net/http"
	"testing"

	"server/internal/image"
	"server/internal/testutil"

	"github.com/stretchr/testify/require"
)

func TestGetImagesByQuery(t *testing.T) {
	tests := []struct {
		name          string
		setupImages   func(ctx *TestContext)
		query         string
		userID        uint64
		wantStatus    int
		expectedCount int
		checkResponse func(t *testing.T, images []image.ImageResponse)
	}{
		{
			name: "success - no filters",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByNameAndTags("img1.png", "dog", "cat")
				ctx.SeedImageByName("img2.png")
				ctx.SeedImageByName("img3.png")
			},
			query:         "",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 3,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 3)
			},
		},
		{
			name: "success - prefix filter",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByNameAndTags("cat_image.png", "cat")
				ctx.SeedImageByNameAndTags("dog_image.png", "dog")
				ctx.SeedImageByNameAndTags("bird_image.png", "bird")
			},
			query:         "prefix=cat",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 1,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 1)
				require.Equal(t, "cat_image.png", images[0].Filename)
			},
		},
		{
			name: "success - tag filter",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByNameAndTags("cat_image.png", "cat")
				ctx.SeedImageByNameAndTags("dog_image.png", "dog")
				ctx.SeedImageByNameAndTags("cat_dog_image.png", "cat", "dog")
			},
			query:         `tags=["cat"]`,
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 2)

				for _, img := range images {
					require.Contains(t, img.Tags, "cat")
				}
			},
		},
		{
			name: "success - multiple tags filter (AND)",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByNameAndTags("cat_image.png", "cat")
				ctx.SeedImageByNameAndTags("dog_image.png", "dog")
				ctx.SeedImageByNameAndTags("img.png", "cat", "dog")
			},
			query:         `tags=["cat","dog"]`,
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 1,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 1)
				require.Equal(t, "img.png", images[0].Filename)
			},
		},
		{
			name: "success - user filter",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByNameAndTags("cat_image.png", "cat")
				ctx.SeedImageByNameAndTags("dog_image.png", "dog")
				ctx.SeedImageByNameAndTags("img.png", "cat", "dog")
				ctx.SeedImage(
					image.ImagePostRequest{Name: "user_image.png"},
					testutil.MakeUniqueTestPNG(t, 10, 10, 141),
					2,
				)
			},
			query:         "user_id=1",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 3,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 3)

				for _, img := range images {
					require.Equal(t, uint(1), img.UserID)
				}
			},
		},
		{
			name: "success - username filter",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImage(
					image.ImagePostRequest{Name: "alice_img.png"},
					testutil.MakeUniqueTestPNG(t, 10, 10, 141),
					4,
				)
				ctx.SeedImage(
					image.ImagePostRequest{Name: "bob_img.png"},
					testutil.MakeUniqueTestPNG(t, 10, 10, 142),
					3,
				)
			},
			query:         "username=alice",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 1,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 1)
				require.Equal(t, "alice_img.png", images[0].Filename)
			},
		},
		{
			name: "success - pagination with limit",
			setupImages: func(ctx *TestContext) {
				for i := 1; i <= 5; i++ {
					ctx.SeedImageByNameAndTags(
						fmt.Sprintf("page_img%d.png", i),
						"test",
					)
				}
			},
			query:         "limit=2",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 2)
			},
		},
		{
			name: "success - cursor pagination",
			setupImages: func(ctx *TestContext) {
				for i := 1; i <= 5; i++ {
					ctx.SeedImageByNameAndTags(
						fmt.Sprintf("cursor_img%d.png", i),
						"test",
					)
				}
			},
			query:         "limit=2&cursor=2",
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 2)
			},
		},
		{
			name: "success - combined filters",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByNameAndTags("combo1.png", "cat", "animal")
				ctx.SeedImageByNameAndTags("combo2.png", "dog", "animal")
				ctx.SeedImage(
					image.ImagePostRequest{
						Name: "combo3.png",
						Tags: []string{"cat"},
					},
					testutil.MakeUniqueTestPNG(t, 10, 10, 182),
					2,
				)
			},
			query:         `name=combo&tags=["animal"]&user_id=1`,
			userID:        1,
			wantStatus:    http.StatusOK,
			expectedCount: 2,
			checkResponse: func(t *testing.T, images []image.ImageResponse) {
				require.Len(t, images, 2)

				require.ElementsMatch(t,
					[]string{"combo1.png", "combo2.png"},
					[]string{
						images[0].Filename,
						images[1].Filename,
					},
				)
			},
		},
		{
			name: "public access allowed",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByName("test.png")
			},
			query:         "",
			userID:        0,
			wantStatus:    http.StatusOK,
			expectedCount: 1,
		},
		{
			name: "bad request - invalid tags json",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByName("test.png")
			},
			query:      `tags=[invalid]`,
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid cursor",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByName("test.png")
			},
			query:      "cursor=abc",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid limit",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByName("test.png")
			},
			query:      "limit=xyz",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid user_id",
			setupImages: func(ctx *TestContext) {
				ctx.SeedImageByName("test.png")
			},
			query:      "user_id=abc",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cleanup := setupContext(t)
			defer cleanup()

			tt.setupImages(ctx)

			rec, images := ctx.QueryImagesWithResponse(tt.query, tt.userID)

			ctx.AssertStatus(rec, tt.wantStatus)

			if tt.wantStatus != http.StatusOK {
				return
			}

			require.Len(t, images, tt.expectedCount)

			if tt.checkResponse != nil {
				tt.checkResponse(t, images)
			}
		})
	}
}

func TestGetImagesByQuery_EmptyResult(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	rec, images := ctx.QueryImagesWithResponse("", 1)

	ctx.AssertStatus(rec, http.StatusOK)
	require.Empty(t, images)
}
