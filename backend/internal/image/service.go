package image

import (
	"context"
	"fmt"
	"server/internal/api/auth"
	"server/internal/api/transaction"
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
	Upload(ctx context.Context, userID uint, req *ImagePostRequest, format Format, data []byte) (*ImageMetadata, error)

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

	transactions transaction.Runner
}

func NewImageService(
	images ImageRepository,
	logs log.LogService,
	objects ObjectService,
	embeddings embeddingspkg.EmbeddingsService,
	transactions transaction.Runner,
) ImageService {
	return &imageService{
		images,
		logs,
		objects,
		embeddings,
		transactions,
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

	if !auth.CanEdit(user.ID, user.Privilege, img.UserID) {
		return errors.ErrPermissionDenied
	}

	if err := s.transactions.Within(ctx, func(tx *gorm.DB) error {
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
		if !auth.CanEdit(user.ID, user.Privilege, img.UserID) {
			return []error{errors.ErrPermissionDenied}
		}

		ids = append(ids, img.ID)
	}

	if err = s.transactions.Within(ctx, func(tx *gorm.DB) error {
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

func (s *imageService) validateNewImage(img *ImageMetadata) error {
	exists, err := s.images.ExistsByHash(img.Hash)
	if err != nil {
		return fmt.Errorf("check duplicate hash: %w", err)
	}
	if exists {
		return fmt.Errorf("duplicate hash: %w", err)
	}

	exists, err = s.images.ExistsByName(img.Filename)
	if err != nil {
		return fmt.Errorf("check duplicate name: %w", err)
	}
	if exists {
		return fmt.Errorf("duplicate name: %w", err)
	}

	return nil
}

func (s *imageService) Upload(ctx context.Context, userID uint, req *ImagePostRequest, format Format, data []byte) (*ImageMetadata, error) {
	imgWidth, imgHeight, err := GetDimensionsBytes(data)
	if err != nil {
		return nil, fmt.Errorf("error getting image dimensions: %v", err)
	}

	img := ConstructImageMetadata(req.Name, HashBytes(data), req.Tags, format, imgWidth, imgHeight, userID)

	if err := s.validateNewImage(img); err != nil {
		return nil, err
	}

	if err := s.objects.UploadImage(ctx, img.Hash, format, data); err != nil {
		return nil, fmt.Errorf("error uploading image: %w", err)
	}

	if err := s.transactions.Within(ctx, func(tx *gorm.DB) error {
		imgRepoTx := s.images.WithTx(tx)

		// Save tags from upload metadata before clearing
		tagsToAttach := img.Tags
		img.Tags = nil

		if err := imgRepoTx.Create(img); err != nil {
			return fmt.Errorf("error saving the image to database: %w", err)
		}

		if err := imgRepoTx.AttachTags(img.ID, tagsToAttach); err != nil {
			return err
		}
		img.Tags = tagsToAttach

		if err := s.logs.Log(tx, "create", "image", userID, img.ID, nil); err != nil {
			return err
		}

		if err := s.embeddings.Upsert(ctx, img.ID, data); err != nil {
			return err
		}

		return nil
	}); err != nil {
		// _ = s.embeddings.Delete(ctx, img.ID)
		_ = s.objects.DeleteImageObjects(ctx, img.Hash, format.String())
		return nil, err
	}

	return img, nil
}
