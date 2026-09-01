package image_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/textproto"
	"server/internal/api/transaction"
	"server/internal/image"
	"server/internal/tag"
	"server/internal/testutil"

	embeddingMocks "server/internal/embedding/mocks"
	imageMocks "server/internal/image/mocks"
	logMocks "server/internal/log/mocks"

	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func fileHeader(t *testing.T, filename, contentType string, data []byte) *multipart.FileHeader {
	t.Helper()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	form, err := multipart.NewReader(body, writer.Boundary()).ReadForm(int64(body.Len()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, form.RemoveAll()) })
	return form.File["file"][0]
}

func TestUpload_Success(t *testing.T) {
	data := testutil.MakeTestPNG(t, 10, 10)
	hash := image.HashBytes(data)
	format := image.FormatPNG
	name := "success.png"
	tags := []string{"animal", "dog"}
	userID := uint(1)

	imageRepo := imageMocks.NewMockImageRepository(t)
	imageRepo.EXPECT().
		ExistsByHash(image.HashBytes(data)).
		Return(false, nil).Once()
	imageRepo.EXPECT().
		ExistsByName(name).
		Return(false, nil).Once()

	txRepo := imageMocks.NewMockImageRepository(t)
	imageRepo.EXPECT().
		WithTx((*gorm.DB)(nil)).
		Return(txRepo).Once()

	txRepo.EXPECT().
		Create(mock.AnythingOfType("*image.ImageMetadata")).
		RunAndReturn(func(actual *image.ImageMetadata) error {
			actual.ID = 42
			return nil
		}).Once()

	txRepo.EXPECT().
		AttachTags(mock.AnythingOfType("uint"), mock.AnythingOfType("[]tag.Tag")).
		Return(nil).Once()

	logService := logMocks.NewMockLogService(t)
	logService.EXPECT().
		Log((*gorm.DB)(nil), "create", "image", uint(1), mock.AnythingOfType("uint"), nil).
		Return(nil).Once()

	objectsService := imageMocks.NewMockObjectService(t)
	objectsService.EXPECT().
		UploadImage(context.Background(), hash, format, data).
		Return(nil).Once()

	embeddingsService := embeddingMocks.NewMockEmbeddingsService(t)
	embeddingsService.EXPECT().
		Upsert(context.Background(), mock.AnythingOfType("uint"), data).
		Return(nil).Once()

	service := image.NewImageService(imageRepo, logService, objectsService, embeddingsService, transaction.TestRunner{})
	resImg, err := service.Upload(
		context.Background(),
		userID,
		&image.ImagePostRequest{Name: name, Tags: tags},
		format,
		data,
	)

	require.NoError(t, err)
	require.Equal(t, uint(42), resImg.ID)
	require.Equal(t, name, resImg.Filename)
	require.Equal(t, hash, resImg.Hash)
	require.Equal(t, format.String(), resImg.Format)
	require.Equal(t, uint(10), resImg.Width)
	require.Equal(t, uint(10), resImg.Height)
	require.Equal(t, tags, tag.TagsToStrings(resImg.Tags))
}

func TestUpload_RejectsDuplicateName(t *testing.T) {
	name := "duplicate.png"

	repo := imageMocks.NewMockImageRepository(t)
	repo.EXPECT().ExistsByHash(mock.Anything).Return(false, nil).Once()
	repo.EXPECT().ExistsByName(name).Return(true, nil).Once()

	service := image.NewImageService(repo, nil, nil, nil, nil)
	img, err := service.Upload(
		context.Background(),
		1,
		&image.ImagePostRequest{Name: name},
		image.FormatPNG,
		testutil.MakeTestPNG(t, 10, 10),
	)

	require.Nil(t, img)
	require.ErrorContains(t, err, "duplicate name")
}

func TestUpload_RejectsDuplicateHash(t *testing.T) {
	data := testutil.MakeTestPNG(t, 10, 10)

	repo := imageMocks.NewMockImageRepository(t)
	repo.EXPECT().ExistsByHash(image.HashBytes(data)).Return(true, nil).Once()

	service := image.NewImageService(repo, nil, nil, nil, nil)
	img, err := service.Upload(
		context.Background(),
		1,
		&image.ImagePostRequest{Name: "duplicate.png"},
		image.FormatPNG,
		data,
	)

	require.Nil(t, img)
	require.ErrorContains(t, err, "duplicate hash")
}
