package api

import (
	"testing"

	embeddings "server/internal/embedding"
	embeddingspkg "server/internal/embedding"
	"server/internal/image"
	"server/internal/log"
	"server/internal/storage"
	"server/internal/tag"
	"server/internal/user"

	"gorm.io/gorm"

	testutil "server/internal/testutil"
	tutil "server/internal/testutil/testing"
)

func mustInitTestAPI(t *testing.T, storage storage.TestStorage) *api {
	t.Helper()
	db := testutil.MustOpenDB(t)

	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo)

	createTestUsers(t, db, userService)

	embeddingsClient := embeddings.NewMockEmbeddingsClient()

	imageRepo := image.NewImageRepository(db)
	tagRepo := tag.NewTagRepository(db)
	logRepo := log.NewLogRepository(db)

	logService := log.NewLogService(logRepo)
	objectService := image.NewObjectService(storage, imageRepo)
	embeddingsService := embeddingspkg.NewEmbeddingsService(objectService, embeddingsClient)

	imageService := image.NewImageService(db, imageRepo, logService, objectService, embeddingsService)
	tagService := tag.NewTagService(db, tagRepo, logService)

	api := &api{
		imageService:      imageService,
		userService:       userService,
		tagService:        tagService,
		logService:        logService,
		objectService:     objectService,
		embeddingsService: embeddingsService,

		storage:   storage,
		jwtSecret: "test-secret",
		db:        db,
	}

	api.initRoutes(tutil.TestAuthMiddleware(userService))

	return api
}

func createTestUsers(t *testing.T, db *gorm.DB, userService user.UserService) {
	t.Helper()

	users := []*user.User{
		{Username: "moderator", Email: "moderator@test.com", Privilege: user.Moderator, Provider: "test", ProviderID: "moderator"},
		{Username: "user1", Email: "user1@test.com", Privilege: user.Unprivileged, Provider: "test", ProviderID: "user1"},
		{Username: "user2", Email: "user2@test.com", Privilege: user.Unprivileged, Provider: "test", ProviderID: "user2"},
		{Username: "alice", Email: "alice@test.com", Privilege: user.Unprivileged, Provider: "test", ProviderID: "alice"},
		{Username: "user3", Email: "user3@test.com", Privilege: user.Unprivileged, Provider: "test", ProviderID: "user3"},
	}

	for _, u := range users {
		if err := userService.Create(u); err != nil {
			t.Fatalf("failed to create test user %s: %v", u.Username, err)
		}
	}
}

func setupContext(t *testing.T) (*tutil.TestContext, func()) {
	testStorage := storage.NewMockTestStorage()
	api := mustInitTestAPI(t, testStorage)
	ctx := tutil.NewTestContext(t, api.router, api.userService, api.imageService, api.tagService, testStorage)

	cleanup := func() {
		testutil.CleanTestDB(api.db)
	}

	return ctx, cleanup
}
