package image

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	embeddingspkg "server/internal/embedding"
	"server/internal/errors"
	"server/internal/log"
	userpkg "server/internal/user"

	"gorm.io/gorm"
)

func toServiceError[T any](val T, err error) (T, error) {
	if err == nil {
		return val, nil
	}
	var zero T
	if err == gorm.ErrRecordNotFound {
		return zero, errors.ErrNotFound
	}
	return zero, err
}

type ImageService interface {
	Upload(ctx context.Context, user *userpkg.User, uploadMetadata *ImagePostRequest, fileHeader *multipart.FileHeader) (*ImageMetadata, error)

	GetByID(id uint) (*ImageMetadata, error)
	GetByName(name string) (*ImageMetadata, error)
	GetByHash(hash string) (*ImageMetadata, error)
	Search(query *Query) ([]ImageMetadata, error)

	DeleteByID(ctx context.Context, user *userpkg.User, id uint) error
	DeleteByName(ctx context.Context, user *userpkg.User, name string) error
	DeleteByQuery(ctx context.Context, user *userpkg.User, query *Query) []error

	Count(query *Query) (int64, error)

	ExistsByHash(hash string) (bool, error)
	ExistsByName(name string) (bool, error)
}

type imageService struct {
	images     ImageRepository
	logs       log.LogService
	objects    ObjectService
	embeddings embeddingspkg.EmbeddingsService

	db *gorm.DB
}

func NewImageService(db *gorm.DB, images ImageRepository, logs log.LogService, objects ObjectService, embeddings embeddingspkg.EmbeddingsService) ImageService {
	return &imageService{
		images,
		logs,
		objects,
		embeddings,
		db,
	}
}

func (s *imageService) GetByID(id uint) (*ImageMetadata, error) {
	return toServiceError(s.images.GetByID(id))
}

func (s *imageService) GetByName(name string) (*ImageMetadata, error) {
	return toServiceError(s.images.GetByName(name))
}

func (s *imageService) GetByHash(hash string) (*ImageMetadata, error) {
	return toServiceError(s.images.GetByHash(hash))
}

func (s *imageService) Search(query *Query) ([]ImageMetadata, error) {
	return s.images.Search(query)
}

func (s *imageService) Count(query *Query) (int64, error) {
	return s.images.Count(query)
}

func (s *imageService) ExistsByHash(hash string) (bool, error) {
	return s.images.ExistsByHash(hash)
}

func (s *imageService) ExistsByName(name string) (bool, error) {
	return s.images.ExistsByName(name)
}

func (s *imageService) DeleteByID(ctx context.Context, user *userpkg.User, id uint) error {
	if user == nil {
		return fmt.Errorf("deleting image: received nil user")
	}
	img, err := s.images.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed getting image for deletion: %v", err)
	}

	allowed := user.Privilege == userpkg.Moderator || user.ID == img.UserID
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

	format, err := ParseFormat(img.Format)
	if err != nil {
		return fmt.Errorf("failed to parse image format: %w", err)
	}

	if err := s.embeddings.Delete(ctx, img.ID); err != nil {
		return err
	}

	if err := s.objects.DeleteImageObjects(ctx, img.Hash, format.String()); err != nil {
		return fmt.Errorf("failed deleting image: %v", err)
	}

	return nil
}

func (s *imageService) DeleteByQuery(ctx context.Context, user *userpkg.User, query *Query) []error {
	if user == nil {
		return []error{fmt.Errorf("deleting image: received nil user")}
	}
	imgs, err := s.images.Search(query)
	var errs []error
	if err != nil {
		errs = append(errs, fmt.Errorf("failed querying for deletion %w", err))
		return errs
	}

	ids := make([]uint, 0, len(imgs))
	for _, img := range imgs {
		if !(user.Privilege == userpkg.Moderator || user.ID == img.UserID) {
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
		format, err := ParseFormat(img.Format)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse image format for %s: %w", img.Hash, err))
			continue
		}
		if err := s.objects.DeleteImageObjects(ctx, img.Hash, format.String()); err != nil {
			errs = append(errs, fmt.Errorf("failed deleting image: %w", err))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (s *imageService) DeleteByName(ctx context.Context, user *userpkg.User, name string) error {
	img, err := s.GetByName(name)
	if err != nil {
		return err
	}

	return s.DeleteByID(ctx, user, img.ID)
}

func (s *imageService) Upload(ctx context.Context, user *userpkg.User, uploadMetadata *ImagePostRequest, fileHeader *multipart.FileHeader) (*ImageMetadata, error) {
	format, err := FormatFromMIME(fileHeader.Header.Get("Content-Type"))
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

	imgWidth, imgHeight, err := GetDimensionsBytes(data)
	if err != nil {
		return nil, fmt.Errorf("error getting image dimensions: %v", err)
	}

	img := ConstructImageMetadata(uploadMetadata.Name, uploadMetadata.Tags, format.String(), imgWidth, imgHeight, user.ID)
	img.Hash = HashBytes(data)

	exists, err := s.images.ExistsByHash(img.Hash)
	if err != nil {
		return nil, fmt.Errorf("check duplicate: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("duplicate image: %w", err)
	}

	if err := s.objects.UploadImage(ctx, img.Hash, format.String(), data); err != nil {
		return nil, fmt.Errorf("error uploading image: %w", err)
	}

	if err := s.embeddings.Upsert(ctx, img.ID, data); err != nil {
		return nil, err
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
		_ = s.embeddings.Delete(ctx, img.ID)
		_ = s.objects.DeleteImageObjects(ctx, img.Hash, format.String())
		return nil, err
	}

	return img, nil}