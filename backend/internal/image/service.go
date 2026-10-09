package image

import (
	"context"
	"fmt"
	"server/internal/api/transaction"
	embeddingspkg "server/internal/embedding"
	"server/internal/errors"
	"server/internal/log"
	"server/internal/tag"
	userpkg "server/internal/user"
	"strconv"

	"gorm.io/gorm"
)

type ImageService interface {
	Upload(ctx context.Context, userID uint, req UploadRequest, format Format, data []byte) (*ImageMetadata, error)

	Get(ctx context.Context, id uint) (*ImageMetadata, error)
	GetByName(ctx context.Context, name string) (*ImageMetadata, error)
	GetByHash(ctx context.Context, hash string) (*ImageMetadata, error)

	Update(ctx context.Context, userID uint, imageID uint, image *ImagePatchRequest) (*ImageMetadata, error)

	Delete(ctx context.Context, userID uint, id uint) error
	DeleteByName(ctx context.Context, userID uint, name string) error
	DeleteByQuery(ctx context.Context, userID uint, query *Query) []error

	Search(ctx context.Context, query *Query) ([]ImageMetadata, error)
	Count(ctx context.Context, query *Query) (int64, error)

	ExistsByHash(ctx context.Context, hash string) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

type imageService struct {
	images     ImageRepository
	users      userpkg.UserRepository
	logs       log.LogService
	objects    ObjectService
	embeddings embeddingspkg.EmbeddingsService

	transactions transaction.Runner
}

func NewImageService(
	images ImageRepository,
	users userpkg.UserRepository,
	logs log.LogService,
	objects ObjectService,
	embeddings embeddingspkg.EmbeddingsService,
	transactions transaction.Runner,
) ImageService {
	return &imageService{
		images,
		users,
		logs,
		objects,
		embeddings,
		transactions,
	}
}

func (s *imageService) Get(ctx context.Context, id uint) (*ImageMetadata, error) {
	return s.images.Get(ctx, id)
}

func (s *imageService) GetByName(ctx context.Context, name string) (*ImageMetadata, error) {
	return s.images.GetByName(ctx, name)
}

func (s *imageService) GetByHash(ctx context.Context, hash string) (*ImageMetadata, error) {
	return s.images.GetByHash(ctx, hash)
}

func (s *imageService) Update(ctx context.Context, userID uint, imageID uint, req *ImagePatchRequest) (*ImageMetadata, error) {
	img, err := s.images.Get(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if err := s.checkUserPermission(ctx, img.UserID, userID); err != nil {
		return nil, err
	}

	if req.Filename != nil {
		img.Filename = *req.Filename
	}

	if req.Tags != nil {
		tags := tag.TagsByNames(req.Tags)
		img.Tags = tags
	}

	if err := s.images.Update(ctx, img); err != nil {
		return nil, err
	}

	return img, err
}

func (s *imageService) Search(ctx context.Context, query *Query) ([]ImageMetadata, error) {
	if !query.Semantic || query.Prefix == "" {
		return s.images.Search(ctx, query)
	}

	// Fetch enough candidates to apply relational filters locally and then apply
	// cursor pagination without changing the filename search semantics.
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	results, err := s.embeddings.Search(ctx, query.Prefix, uint(limit+query.Cursor))
	if err != nil {
		return nil, fmt.Errorf("failed to search embeddings: %w", err)
	}

	imgs := make([]ImageMetadata, 0, limit)
	skipped := 0
	for _, result := range results {
		id, err := parseSearchResultID(result)
		if err != nil {
			continue
		}
		img, err := s.Get(ctx, id)
		if err != nil {
			continue
		}
		if !matchesSemanticFilters(img, query) {
			continue
		}
		if skipped < query.Cursor {
			skipped++
			continue
		}
		imgs = append(imgs, *img)
		if len(imgs) == limit {
			break
		}
	}
	return imgs, nil
}

func parseSearchResultID(result embeddingspkg.SearchResult) (uint, error) {
	if id, ok := result.Payload["image_id"]; ok {
		return parseImageID(id)
	}
	return parseImageID(result.ID)
}

func parseImageID(value any) (uint, error) {
	switch id := value.(type) {
	case float64:
		if id < 0 || id != float64(uint(id)) {
			return 0, fmt.Errorf("invalid image ID %v", value)
		}
		return uint(id), nil
	case float32:
		if id < 0 || id != float32(uint(id)) {
			return 0, fmt.Errorf("invalid image ID %v", value)
		}
		return uint(id), nil
	case int:
		if id < 0 {
			return 0, fmt.Errorf("invalid image ID %d", id)
		}
		return uint(id), nil
	case uint:
		return id, nil
	case string:
		parsed, err := strconv.ParseUint(id, 10, 0)
		if err != nil {
			return 0, fmt.Errorf("invalid image ID %q: %w", id, err)
		}
		return uint(parsed), nil
	default:
		return 0, fmt.Errorf("invalid image ID type %T", value)
	}
}

func matchesSemanticFilters(img *ImageMetadata, query *Query) bool {
	if query.User != nil && img.UserID != query.User.ID {
		return false
	}
	if len(query.Tags) == 0 {
		return true
	}

	imageTags := make(map[string]struct{}, len(img.Tags))
	for _, tag := range img.Tags {
		imageTags[tag.Name] = struct{}{}
	}
	for _, wanted := range query.Tags {
		if _, ok := imageTags[wanted]; !ok {
			return false
		}
	}
	return true
}

func (s *imageService) Count(ctx context.Context, query *Query) (int64, error) {
	return s.images.Count(ctx, query)
}

func (s *imageService) ExistsByHash(ctx context.Context, hash string) (bool, error) {
	return s.images.ExistsByHash(ctx, hash)
}

func (s *imageService) ExistsByName(ctx context.Context, name string) (bool, error) {
	return s.images.ExistsByName(ctx, name)
}

func (s *imageService) checkUserPermission(ctx context.Context, imageOwnerID, userID uint) error {
	user, err := s.users.Get(ctx, userID)
	if err != nil {
		return err
	}

	if !userpkg.CanEdit(user.ID, user.Privilege, imageOwnerID) {
		return errors.ErrPermissionDenied
	}

	return nil
}

func (s *imageService) Delete(ctx context.Context, userID, id uint) error {
	img, err := s.images.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.checkUserPermission(ctx, img.UserID, userID); err != nil {
		return err
	}

	if err := s.transactions.Within(ctx, func(tx *gorm.DB) error {
		imagesRepoTx := s.images.WithTx(tx)
		if err := imagesRepoTx.Delete(ctx, id); err != nil {
			return err
		}
		if err := s.logs.Log(tx, "delete", "image", userID, img.ID, nil); err != nil {
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
		return fmt.Errorf("failed deleting image: %w", err)
	}

	return nil
}

func (s *imageService) checkUserPermissionForMany(ctx context.Context, images []ImageMetadata, userID uint) []error {
	user, err := s.users.Get(ctx, userID)
	if err != nil {
		return []error{err}
	}

	var errs []error
	for _, image := range images {
		if !userpkg.CanEdit(user.ID, user.Privilege, image.UserID) {
			errs = append(errs, errors.WrapPermissionDenied(
				fmt.Errorf("user ID: %d lacks permission for image ID: %d", user.ID, image.ID),
			))
		}
	}

	return errs
}

func (s *imageService) DeleteByQuery(ctx context.Context, userID uint, query *Query) []error {
	imgs, err := s.images.Search(ctx, query)
	var errs []error
	if err != nil {
		errs = append(errs, fmt.Errorf("failed querying for deletion %w", err))
		return errs
	}

	if errs := s.checkUserPermissionForMany(ctx, imgs, userID); len(errs) > 0 {
		return errs
	}

	ids := make([]uint, 0, len(imgs))
	for _, img := range imgs {
		ids = append(ids, img.ID)
	}

	if err = s.transactions.Within(ctx, func(tx *gorm.DB) error {
		imagesRepoTx := s.images.WithTx(tx)
		if err := imagesRepoTx.DeleteByIDs(ctx, ids); err != nil {
			return fmt.Errorf("failed deleting images by query %w", err)
		}

		for _, id := range ids {
			if err := s.logs.Log(tx, "delete", "image", userID, id, nil); err != nil {
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
		// TODO: delete embeddings
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (s *imageService) DeleteByName(ctx context.Context, userID uint, name string) error {
	img, err := s.GetByName(ctx, name)
	if err != nil {
		return err
	}

	return s.Delete(ctx, userID, img.ID)
}

func (s *imageService) validateNewImage(ctx context.Context, img *ImageMetadata) error {
	exists, err := s.images.ExistsByHash(ctx, img.Hash)
	if err != nil {
		return fmt.Errorf("check duplicate hash: %w", err)
	}
	if exists {
		return fmt.Errorf("duplicate hash")
	}

	exists, err = s.images.ExistsByName(ctx, img.Filename)
	if err != nil {
		return fmt.Errorf("check duplicate name: %w", err)
	}
	if exists {
		return fmt.Errorf("duplicate name %s", img.Filename)
	}

	return nil
}

func (s *imageService) Upload(ctx context.Context, userID uint, req UploadRequest, format Format, data []byte) (*ImageMetadata, error) {
	imgWidth, imgHeight, err := GetDimensionsBytes(data)
	if err != nil {
		return nil, fmt.Errorf("error getting image dimensions: %w", err)
	}

	img := ConstructImageMetadata(req.Name, HashBytes(data), req.Tags, format, imgWidth, imgHeight, userID)

	if err := s.validateNewImage(ctx, img); err != nil {
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

		if err := imgRepoTx.Create(ctx, img); err != nil {
			return fmt.Errorf("error saving the image to database: %w", err)
		}

		if err := imgRepoTx.AttachTags(ctx, img.ID, tagsToAttach); err != nil {
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
