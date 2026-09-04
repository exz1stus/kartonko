package http_integration_tests

import (
	"fmt"
	"net/http"
	imgapi "server/internal/api/image"
	"server/internal/image"
	"server/internal/storage"
	"server/internal/testutil"
	"strings"
	"testing"
)

func TestPostImage_StorageImageUploadFails_NoDBRowAndStorageImageLeft(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()
	ctx.Storage.(*storage.MockStorage).FailUploadOn = func(key string) error {
		if !strings.Contains(key, "thumb") {
			return fmt.Errorf("simulated storage failure on main image")
		}
		return nil
	}

	rec := ctx.UploadImage(imgapi.ImagePostRequest{Name: "y.png"}, image.FormatPNG.MIMEType(), testutil.MakeTestPNG(t, 5, 5), 1)

	ctx.AssertStatus(rec, http.StatusInternalServerError)

	query := image.NewQueryBuilder().Prefix("y.png").Build()

	ctx.AssertImageCount(query, 0)
	ctx.AssertStorageCount(0)
}

func TestPostImage_StorageThumbUploadFails_NoDBRowAndStorageImageLeft(t *testing.T) {
	ctx, cleanup := setupContext(t)
	defer cleanup()
	ctx.Storage.(*storage.MockStorage).FailUploadOn = func(key string) error {
		if strings.Contains(key, "thumb") {
			return fmt.Errorf("simulated storage failure on thumb")
		}
		return nil
	}

	rec := ctx.UploadImage(imgapi.ImagePostRequest{Name: "y.png"}, image.FormatPNG.MIMEType(), testutil.MakeTestPNG(t, 5, 5), 1)
	ctx.AssertStatus(rec, http.StatusInternalServerError)

	query := image.NewQueryBuilder().Prefix("y.png").Build()

	ctx.AssertImageCount(query, 0)
	ctx.AssertStorageCount(0)
}
