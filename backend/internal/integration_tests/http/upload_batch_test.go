package http_integration_tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	imgapi "server/internal/api/image"
	"server/internal/image"
	"server/internal/testutil"

	"github.com/stretchr/testify/require"
)

func TestPostImagesBatch(t *testing.T) {
	tests := []struct {
		name       string
		commonTags []string
		fileDatas  []TestFileData
		userID     uint64
		wantStatus int
	}{
		{
			name: "success batch upload",
			fileDatas: []TestFileData{
				{
					image: image.UploadRequest{
						Name: "cat.png",
						Tags: []string{"cat"},
					},
					format:  image.FormatPNG,
					content: testutil.MakeUniqueTestPNG(t, 10, 10, 0),
				},
				{
					image: image.UploadRequest{
						Name: "dog.png",
						Tags: []string{"dog"},
					},
					format:  image.FormatPNG,
					content: testutil.MakeUniqueTestPNG(t, 10, 10, 1),
				},
			},
			commonTags: []string{"animal"},
			userID:     1,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cleanup := setupContext(t)
			defer cleanup()

			ctx.SeedTags("animal", "cat", "dog")

			rec := ctx.UploadImageBatch(tt.fileDatas, tt.commonTags, tt.userID)

			if tt.wantStatus < 400 && tt.wantStatus != 207 {
				var res imgapi.ImagePostBatchResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
					t.Errorf("failed to deserialize batch upload response: %s", rec.Body.String())
				}

				if len(res.Failures) > 0 {
					for _, err := range res.Failures {
						t.Errorf("batch upload error: %s", err)
					}
				}

				ctx.AssertStorageCount(len(tt.fileDatas) * 2)
			}

			if rec.Code != tt.wantStatus {
				t.Fatalf("got status %d, want %d, body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestPostImagesBatch_RejectsFewerFilesThanInMetadata(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	fileDatas := []TestFileData{
		{
			image: image.UploadRequest{
				Name: "cat.png",
			},
			format:  image.FormatPNG,
			content: testutil.MakeUniqueTestPNG(t, 10, 10, 0),
		},
		{
			image: image.UploadRequest{
				Name: "dog.png",
			},
			format:  image.FormatPNG,
			content: testutil.MakeUniqueTestPNG(t, 10, 10, 1),
		},
	}

	failedData := TestFileData{
		image: image.UploadRequest{
			Name: "dog_duplicate.png",
		},
		format:  image.FormatPNG,
		content: testutil.MakeUniqueTestPNG(t, 10, 10, 0),
	}
	appendedDatas := append(fileDatas, failedData)
	appendedDatasJSON, err := json.Marshal(appendedDatas)
	require.NoError(t, err, "failed marshaling")
	appendedDatasJSONstr := string(appendedDatasJSON)

	req := buildBatchUploadRequest(t, "/image/upload/batch", appendedDatasJSONstr, fileDatas)
	req = WithTestUser(req, 1)
	rec := httptest.NewRecorder()
	ctx.Router.ServeHTTP(rec, req)

	ctx.AssertStatus(rec, http.StatusBadRequest)
}

func TestPostImagesBatch_MixedSuccessAndFailure(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()

	fileDatas := []TestFileData{
		{
			image: image.UploadRequest{
				Name: "cat.png",
			},
			format:  image.FormatPNG,
			content: testutil.MakeUniqueTestPNG(t, 10, 10, 0),
		},
		{
			image: image.UploadRequest{
				Name: "dog.png",
			},
			format:  image.FormatPNG,
			content: testutil.MakeUniqueTestPNG(t, 10, 10, 1),
		},
	}

	failedData := TestFileData{
		image: image.UploadRequest{
			Name: "dog_duplicate.png",
		},
		format:  image.FormatPNG,
		content: testutil.MakeUniqueTestPNG(t, 10, 10, 0),
	}

	successCount := len(fileDatas)

	fileDatas = append(fileDatas, failedData)

	rec := ctx.UploadImageBatch(fileDatas, nil, 1)
	ctx.AssertStatus(rec, http.StatusMultiStatus)

	var res imgapi.ImagePostBatchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Errorf("failed to deserialize batch upload response: %s", rec.Body.String())
	}

	query := image.NewQueryBuilder().Prefix("dog_duplicate.png.png").Build()

	ctx.AssertImageCount(query, 0)

	if len(res.Successes) != successCount {
		t.Errorf("expected %d successes , got %d", successCount, len(res.Successes))
	}

	ctx.AssertStorageCount(successCount * 2)
}
