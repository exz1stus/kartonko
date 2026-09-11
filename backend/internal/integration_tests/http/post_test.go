package http_integration_tests

import (
	"encoding/json"
	"net/http"
	imgapi "server/internal/api/image"
	"server/internal/image"
	"server/internal/tag"
	"server/internal/testutil"
	"testing"
)

func TestPostImage(t *testing.T) {
	tests := []struct {
		name       string
		filename   string
		tags       []string
		content    []byte
		mimeType   string
		userID     uint64
		wantStatus int
	}{
		{
			name:       "success valid png",
			filename:   "cat.png",
			tags:       []string{"animal", "cat", "brown"},
			content:    testutil.MakeTestPNG(t, 10, 10),
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing name in metadata",
			filename:   "",
			content:    testutil.MakeTestPNG(t, 5, 5),
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported mime type",
			filename:   "doc.pdf",
			content:    []byte("%PDF-1.4"),
			mimeType:   "application/pdf",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "corrupt image bytes with valid mime",
			filename:   "broken.png",
			content:    []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00},
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "unauthenticated request",
			filename:   "noauth.png",
			content:    testutil.MakeTestPNG(t, 5, 5),
			mimeType:   "image/png",
			userID:     0,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "no file attached",
			filename:   "nofile.png",
			content:    nil,
			mimeType:   "",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "tags are not added",
			filename:   "notallowedtags.png",
			tags:       []string{"notallowedtag"},
			content:    testutil.MakeTestPNG(t, 5, 5),
			mimeType:   "image/png",
			userID:     1,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cleanup := setupContext(t)
			defer cleanup()
			ctx.SeedTags("animal", "cat", "brown")

			postMetadata := imgapi.ImagePostRequest{Name: tt.filename, Tags: tt.tags}
			rec := ctx.UploadImage(postMetadata, tt.mimeType, tt.content, tt.userID)
			ctx.AssertStatus(rec, tt.wantStatus)

			if tt.wantStatus == http.StatusOK {
				ctx.AssertImageCount(1)
				ctx.AssertStorageCount(2)
			}
		})
	}
}

func TestPostImage_TagsAdded(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	ctx.SeedTags("animal", "cat")

	rec := ctx.UploadImage(
		imgapi.ImagePostRequest{Name: "tagged.png", Tags: []string{"animal", "cat"}},
		image.FormatPNG.MIMEType(),
		testutil.MakeTestPNG(t, 10, 10),
		1,
	)

	ctx.AssertStatus(rec, http.StatusCreated)

	// Verify response contains tags
	var resp imgapi.ImageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	ctx.AssertTags(resp.Tags, []string{"animal", "cat"}, "response")

	// Verify tags are persisted in database
	img, err := ctx.ImageService.GetByName("tagged.png")
	if err != nil {
		t.Fatalf("failed to get image from DB: %v", err)
	}
	dbTags := tag.TagsToStrings(img.Tags)
	ctx.AssertTags(dbTags, []string{"animal", "cat"}, "database")

	ctx.AssertImageCount(1)
	ctx.AssertStorageCount(2)
}
