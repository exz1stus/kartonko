package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"server/internal/models"
	"server/internal/repositories"
	"server/internal/storage"
	"server/pkg/image"
)

type ObjectService interface {
	GetRawImage(ctx context.Context, id uint) (reader io.ReadCloser, img *models.ImageMetadata, err error)
	GetRawThumbnail(ctx context.Context, id uint) (reader io.ReadCloser, img *models.ImageMetadata, err error)
	GetRawImageByHash(ctx context.Context, hash string) (reader io.ReadCloser, img *models.ImageMetadata, err error)
	GetRawThumbnailByHash(ctx context.Context, hash string) (reader io.ReadCloser, img *models.ImageMetadata, err error)

	DeleteImageObjects(ctx context.Context, hash string, format image.Format) error
	UploadImage(ctx context.Context, hash string, format image.Format, data []byte) error
}

type objectService struct {
	storage storage.Storage
	images  repositories.ImageRepository
}

func NewObjectService(storage storage.Storage, images repositories.ImageRepository) ObjectService {
	return &objectService{
		storage,
		images,
	}
}

func (s *objectService) getRaw(
	ctx context.Context,
	img *models.ImageMetadata,
	keyFunc func(string, image.Format) string,
) (io.ReadCloser, *models.ImageMetadata, error) {
	format, err := image.ParseFormat(img.Format)
	if err != nil {
		return nil, nil, err
	}

	key := keyFunc(img.Hash, format)

	body, err := s.storage.Download(ctx, key)
	if err != nil {
		return nil, nil, err
	}

	return body, img, nil
}

func (s *objectService) GetRawImage(ctx context.Context, id uint) (reader io.ReadCloser, img *models.ImageMetadata, err error) {
	img, err = s.images.GetByID(id)
	if err != nil {
		return nil, nil, err
	}
	return s.getRaw(ctx, img, image.ImageKey)
}

func (s *objectService) GetRawThumbnail(ctx context.Context, id uint) (reader io.ReadCloser, img *models.ImageMetadata, err error) {
	img, err = s.images.GetByID(id)
	if err != nil {
		return nil, nil, err
	}
	return s.getRaw(ctx, img, image.ThumbnailKey)
}

func (s *objectService) GetRawImageByHash(ctx context.Context, hash string) (reader io.ReadCloser, img *models.ImageMetadata, err error) {
	img, err = s.images.GetByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	return s.getRaw(ctx, img, image.ImageKey)
}

func (s *objectService) GetRawThumbnailByHash(ctx context.Context, hash string) (reader io.ReadCloser, img *models.ImageMetadata, err error) {
	img, err = s.images.GetByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	return s.getRaw(ctx, img, image.ThumbnailKey)
}

func (s *objectService) DeleteImageObjects(ctx context.Context, hash string, format image.Format) error {
	if err := s.storage.Delete(ctx, image.ImageKey(hash, format)); err != nil {
		return fmt.Errorf("failed deleting image: %v", err)
	}

	if err := s.storage.Delete(ctx, image.ThumbnailKey(hash, format)); err != nil {
		return fmt.Errorf("failed deleting image thumbnail: %v", err)
	}

	return nil
}

func (s *objectService) UploadImage(ctx context.Context, hash string, format image.Format, data []byte) error {
	thumb, err := image.GenerateThumbnail(data, format)
	if err != nil {
		return fmt.Errorf("error generating thumbnail: %w", err)
	}

	imageKey := image.ImageKey(hash, format)
	if err = s.storage.Upload(ctx, imageKey, bytes.NewReader(data), format.MIMEType()); err != nil {
		return fmt.Errorf("error uploading image: %w", err)
	}

	thumbKey := image.ThumbnailKey(hash, format)
	if err = s.storage.Upload(ctx, thumbKey, bytes.NewReader(thumb), format.MIMEType()); err != nil {
		_ = s.storage.Delete(ctx, imageKey)
		return fmt.Errorf("error uploading thumbnail: %w", err)
	}

	return nil
}
