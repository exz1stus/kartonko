package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"server/internal/api/dto"
	"server/internal/auth"
	"server/internal/errors"
	"server/internal/models"
	"server/internal/repositories"

	"server/internal/storage"
	"server/pkg/image"

	"gorm.io/gorm"
)

type ImageService interface {
	Upload(ctx context.Context, user *models.User, uploadMetadata *dto.ImagePostRequest, fileHeader *multipart.FileHeader) (*models.ImageMetadata, error)

	GetByID(id uint) (*models.ImageMetadata, error)
	GetByName(name string) (*models.ImageMetadata, error)
	GetByHash(hash string) (*models.ImageMetadata, error)
	Search(query *models.ImageQuery) ([]models.ImageMetadata, error)

	DeleteByID(ctx context.Context, user *models.User, id uint) error
	DeleteByName(ctx context.Context, user *models.User, name string) error
	DeleteByQuery(ctx context.Context, user *models.User, query *models.ImageQuery) []error

	Count(query *models.ImageQuery) (int64, error)

	ExistsByHash(hash string) (bool, error)
	ExistsByName(name string) (bool, error)
}

type imageService struct {
	images  repositories.ImageRepository
	logs    LogService
	storage storage.Storage

	db *gorm.DB
}

func NewImageService(db *gorm.DB, images repositories.ImageRepository, logs LogService, storage storage.Storage) ImageService {
	return &imageService{
		images,
		logs,
		storage,
		db,
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
	if user == nil {
		return fmt.Errorf("deleting image: received nil user")
	}
	img, err := s.images.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed getting image for deletion: %v", err)
	}

	allowed := auth.CanEdit(user, img)
	if !allowed {
		return errors.ErrPermissionDenied
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		imagesRepoTx := s.images.WithTx(tx)
		if err := imagesRepoTx.DeleteByID(id); err != nil {
			return err
		}
		if err := s.logs.Log(tx, "delete", "image", user.ID, img.ID, nil); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	format, err := image.ParseFormat(img.Format)
	if err != nil {
		return fmt.Errorf("failed to parse image format: %w", err)
	}

	if err := s.storage.Delete(ctx, image.ImageKey(img.Hash, format)); err != nil {
		return fmt.Errorf("failed deleting image: %v", err)
	}

	if err := s.storage.Delete(ctx, image.ThumbnailKey(img.Hash, format)); err != nil {
		return fmt.Errorf("failed deleting image thumbnail: %v", err)
	}

	return nil
}

func (s *imageService) DeleteByQuery(ctx context.Context, user *models.User, query *models.ImageQuery) []error {
	if user == nil {
		fmt.Errorf("deleting image: recieved nil user")
	}
	imgs, err := s.images.Search(query)
	var errs []error
	if err != nil {
		errs = append(errs, fmt.Errorf("failed querying for deletion %w", err))
		return errs
	}

	ids := make([]uint, 0, len(imgs))
	for _, img := range imgs {
		if !auth.CanEdit(user, &img) {
			return []error{errors.ErrPermissionDenied}
		}

		ids = append(ids, img.ID)
	}

	if err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		imagesRepoTx := s.images.WithTx(tx)
		if err := imagesRepoTx.DeleteByIDs(ids); err != nil {
			return fmt.Errorf("failed deleting images by query %w", err)
		}

		for _, id := range ids {
			if err := s.logs.Log(tx, "delete", "image", user.ID, id, nil); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		errs = append(errs, err)
		return errs
	}

	for _, img := range imgs {
		format, err := image.ParseFormat(img.Format)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse image format for %s: %w", img.Hash, err))
			continue
		}
		if err := s.storage.Delete(ctx, image.ImageKey(img.Hash, format)); err != nil {
			errs = append(errs, fmt.Errorf("failed deleting image: %w", err))
		}

		if err := s.storage.Delete(ctx, image.ThumbnailKey(img.Hash, format)); err != nil {
			errs = append(errs, fmt.Errorf("failed deleting image thumbnail: %w", err))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (s *imageService) DeleteByName(ctx context.Context, user *models.User, name string) error {
	img, err := s.GetByName(name)
	if err != nil {
		return err
	}

	return s.DeleteByID(ctx, user, img.ID)
}

func (s *imageService) Upload(ctx context.Context, user *models.User, uploadMetadata *dto.ImagePostRequest, fileHeader *multipart.FileHeader) (*models.ImageMetadata, error) {
	format, err := image.FormatFromMIME(fileHeader.Header.Get("Content-Type"))
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

	img := models.ConstructImageMetadata(uploadMetadata.Name, uploadMetadata.Tags, format.String(), imgWidth, imgHeight, user.ID)
	img.Hash = image.HashBytes(data)

	exists, err := s.images.ExistsByHash(img.Hash)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("duplicate image: %w", err)
	}

	var thumb []byte
	thumb, err = image.GenerateThumbnail(data, format)
	if err != nil {
		return nil, fmt.Errorf("error generating thumbnail: %w", err)
	}

	imageKey := image.ImageKey(img.Hash, format)
	if err = s.storage.Upload(ctx, imageKey, bytes.NewReader(data), format.MIMEType()); err != nil {
		return nil, fmt.Errorf("error uploading image: %w", err)
	}

	thumbKey := image.ThumbnailKey(img.Hash, format)
	if err = s.storage.Upload(ctx, thumbKey, bytes.NewReader(thumb), "image/jpeg"); err != nil {
		_ = s.storage.Delete(ctx, imageKey)
		return nil, fmt.Errorf("error uploading thumbnail: %w", err)
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		imgRepoTx := s.images.WithTx(tx)

		// Save tags from upload metadata before clearing
		tagsToAttach := img.Tags
		img.Tags = nil

		if err := imgRepoTx.Create(img); err != nil {
			return fmt.Errorf("error saving the image to database: %w", err)
		}

		if err := imgRepoTx.AttachTags(img, tagsToAttach); err != nil {
			return err
		}

		if err := s.logs.Log(tx, "create", "image", user.ID, img.ID, nil); err != nil {
			return err
		}

		return nil
	}); err != nil {
		_ = s.storage.Delete(ctx, imageKey)
		_ = s.storage.Delete(ctx, thumbKey)
		return nil, err
	}

	return img, nil
}
