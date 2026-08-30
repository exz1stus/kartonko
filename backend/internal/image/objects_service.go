package image

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"server/internal/image/thumbnail"
	"server/internal/storage"
)

type ObjectService interface {
	GetRawImage(ctx context.Context, id uint) (reader io.ReadCloser, img *ImageMetadata, err error)
	GetRawThumbnail(ctx context.Context, id uint) (reader io.ReadCloser, img *ImageMetadata, err error)
	GetRawImageByHash(ctx context.Context, hash string) (reader io.ReadCloser, img *ImageMetadata, err error)
	GetRawThumbnailByHash(ctx context.Context, hash string) (reader io.ReadCloser, img *ImageMetadata, err error)

	// GetRawImageData returns only the image data reader (for embeddings service)
	GetRawImageData(ctx context.Context, id uint) (io.ReadCloser, error)

	DeleteImageObjects(ctx context.Context, hash string, format string) error
	UploadImage(ctx context.Context, hash string, format Format, data []byte) error
}

type objectService struct {
	storage storage.Storage
	images  ImageRepository
}

func NewObjectService(storage storage.Storage, images ImageRepository) ObjectService {
	return &objectService{
		storage,
		images,
	}
}

func (s *objectService) getRaw(
	ctx context.Context,
	img *ImageMetadata,
	keyFunc func(string, string) string,
) (io.ReadCloser, *ImageMetadata, error) {
	format, err := ParseFormat(img.Format)
	if err != nil {
		return nil, nil, err
	}

	key := keyFunc(img.Hash, format.String())

	body, err := s.storage.Download(ctx, key)
	if err != nil {
		return nil, nil, err
	}

	return body, img, nil
}

func (s *objectService) GetRawImage(ctx context.Context, id uint) (reader io.ReadCloser, img *ImageMetadata, err error) {
	img, err = s.images.GetByID(id)
	if err != nil {
		return nil, nil, err
	}
	return s.getRaw(ctx, img, storage.ImageKey)
}

func (s *objectService) GetRawThumbnail(ctx context.Context, id uint) (reader io.ReadCloser, img *ImageMetadata, err error) {
	img, err = s.images.GetByID(id)
	if err != nil {
		return nil, nil, err
	}
	return s.getRaw(ctx, img, storage.ThumbnailKey)
}

func (s *objectService) GetRawImageByHash(ctx context.Context, hash string) (reader io.ReadCloser, img *ImageMetadata, err error) {
	img, err = s.images.GetByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	return s.getRaw(ctx, img, storage.ImageKey)
}

func (s *objectService) GetRawThumbnailByHash(ctx context.Context, hash string) (reader io.ReadCloser, img *ImageMetadata, err error) {
	img, err = s.images.GetByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	return s.getRaw(ctx, img, storage.ThumbnailKey)
}

func (s *objectService) GetRawImageData(ctx context.Context, id uint) (io.ReadCloser, error) {
	img, err := s.images.GetByID(id)
	if err != nil {
		return nil, err
	}
	reader, _, err := s.getRaw(ctx, img, storage.ImageKey)
	return reader, err
}

func (s *objectService) DeleteImageObjects(ctx context.Context, hash string, format string) error {
	if err := s.storage.Delete(ctx, storage.ImageKey(hash, format)); err != nil {
		return fmt.Errorf("failed deleting image: %v", err)
	}

	if err := s.storage.Delete(ctx, storage.ThumbnailKey(hash, format)); err != nil {
		return fmt.Errorf("failed deleting image thumbnail: %v", err)
	}

	return nil
}

func (s *objectService) UploadImage(ctx context.Context, hash string, format Format, data []byte) error {
	thumb, err := thumbnail.GenerateThumbnail(data, format.String())
	if err != nil {
		return fmt.Errorf("error generating thumbnail: %w", err)
	}

	imageKey := storage.ImageKey(hash, format.String())
	if err = s.storage.Upload(ctx, imageKey, bytes.NewReader(data), format.MIMEType()); err != nil {
		return fmt.Errorf("error uploading image: %w", err)
	}

	thumbKey := storage.ThumbnailKey(hash, format.String())
	if err = s.storage.Upload(ctx, thumbKey, bytes.NewReader(thumb), format.MIMEType()); err != nil {
		_ = s.storage.Delete(ctx, imageKey)
		return fmt.Errorf("error uploading thumbnail: %w", err)
	}

	return nil
}
