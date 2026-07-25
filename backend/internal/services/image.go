package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"server/internal/models"
	"server/internal/repositories"
	"server/internal/storage"
	"server/pkg/image"

	"gorm.io/gorm"
)

type ImageService interface {
	Upload(ctx context.Context, image *models.ImageMetadata) error

	GetByID(id uint) (*models.ImageMetadata, error)
	GetByName(name string) (*models.ImageMetadata, error)
	GetByHash(hash string) (*models.ImageMetadata, error)
	Search(query *models.ImageQuery) ([]models.ImageMetadata, error)

	DeleteByID(ctx context.Context, id uint) error
	DeleteByQuery(ctx context.Context, ids []uint) []error

	Count(query *models.ImageQuery) (int64, error)

	ExistsByHash(hash string) (bool, error)
	ExistsByName(name string) (bool, error)
}

type imageService struct {
	images repositories.ImageRepository
	//users repositories.UserRepository
	//  logs    repositories.LogRepository
	storage storage.Storage

	db *gorm.DB
}

func NewImageService(db *gorm.DB, storage storage.Storage) ImageService {
	return &imageService{
		db:      db,
		storage: storage,
	}
}

func (s *imageService) GetByID(id uint) (*models.ImageMetadata, error) {
	return s.images.GetByID(id)
}

func (s *imageService) GetByName(name string) (*models.ImageMetadata, error) {
	return s.images.GetByName(name)
}

func (s *imageService) GetByHash(hash string) (*models.ImageMetadata, error) {
	return s.images.GetByHash(hash)
}

func (s *imageService) Search(query *models.ImageQuery) ([]models.ImageMetadata, error) {
	return s.images.Search(query)
}

func (s *imageService) Count(query *models.ImageQuery) (int64, error) {
	return s.images.Count(query)
}

func (s *imageService) ExistsByHash(hash string) (bool, error) {
	return s.images.ExistsByHash(hash)
}

func (s *imageService) ExistsByName(name string) (bool, error) {
	return s.images.ExistsByName(name)
}

func (s *imageService) DeleteByID(ctx context.Context, user *models.User, id uint) error {
	//TODO: userService.HasPermission

	img, err := s.images.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed getting image for deletion: %v", err)
	}

	if err := s.storage.Delete(ctx, image.ImageKey(img.Hash, img.Format)); err != nil {
		return fmt.Errorf("failed deleting image: %v", err)
	}

	if err := s.storage.Delete(ctx, image.ThumbnailKey(img.Hash)); err != nil {
		return fmt.Errorf("failed deleting image thumbnail: %v", err)
	}

	return s.images.DeleteByID(id)
}

func (s *imageService) DeleteByQuery(ctx context.Context, user *models.User, query *models.ImageQuery) []error {
	//TODO: userService.HasPermission

	imgs, err := s.images.Search(query)
	var errs []error
	if err != nil {
		errs = append(errs, fmt.Errorf("failed querying for deletion %w", err))
		return errs
	}

	ids := make([]uint, len(imgs))
	for i, img := range imgs {
		ids[i] = img.ID
	}

	if err := s.images.DeleteByIDs(ids); err != nil {
		errs = append(errs, fmt.Errorf("failed deleting images by query %w", err))
		return errs
	}

	for _, img := range imgs {
		if err := s.storage.Delete(ctx, image.ImageKey(img.Hash, img.Format)); err != nil {
			errs = append(errs, fmt.Errorf("failed deleting image: %w", err))
		}

		if err := s.storage.Delete(ctx, image.ThumbnailKey(img.Hash)); err != nil {
			errs = append(errs, fmt.Errorf("failed deleting image thumbnail: %w", err))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (s *imageService) Upload(ctx context.Context, user *models.User, uploadMetadata *ImagePostRequest, fileHeader *multipart.FileHeader) (*models.ImageMetadata, error) {
	imgFormat, err := image.MIMETypeToFormat(fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("image format parsing error: %v", err)
	}

	f, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening uploaded file: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading uploaded file: %v", err)
	}

	imgWidth, imgHeight, err := image.GetDimensionsBytes(data)
	if err != nil {
		return nil, fmt.Errorf("error getting image dimensions: %v", err)
	}

	img := models.ConstructImageMetadata(uploadMetadata.Name, uploadMetadata.Tags, imgFormat, imgWidth, imgHeight, user.ID)
	img.Hash = image.HashBytes(data)

	exists, err := s.images.ExistsByHash(img.Hash)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("duplicate image: %w", err)
	}

	var thumb []byte
	thumb, err = image.GenerateThumbnail(data, "."+img.Format)
	if err != nil {
		return nil, fmt.Errorf("error generating thumbnail: %w", err)
	}

	imageKey := image.ImageKey(img.Hash, img.Format)
	if err = s.storage.Upload(ctx, imageKey, bytes.NewReader(data), "image/"+img.Format); err != nil {
		return nil, fmt.Errorf("error uploading image: %w", err)
	}

	thumbKey := image.ThumbnailKey(img.Hash)
	if err = s.storage.Upload(ctx, thumbKey, bytes.NewReader(thumb), "image/jpeg"); err != nil {
		_ = s.storage.Delete(ctx, imageKey)
		return nil, fmt.Errorf("error uploading thumbnail: %w", err)
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		imgRepoTx := s.images.WithTx(tx)

		if err := imgRepoTx.Create(img); err != nil {
			return fmt.Errorf("error saving the image to database: %w", err)
		}

		if err := imgRepoTx.AttachTags(img, img.Tags); err != nil {
			return err
		}

		//TODO: logservice add

		return nil
	})

	if err != nil {
		_ = s.storage.Delete(ctx, imageKey)
		_ = s.storage.Delete(ctx, thumbKey)
		return nil, err
	}

	return img, nil
}
