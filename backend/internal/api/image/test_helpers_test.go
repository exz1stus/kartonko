package image_test

import (
	"testing"

	rootapi "server/internal/api"
	embeddings "server/internal/embedding"
	"server/internal/storage"
	testutil "server/internal/testutil"
	tutil "server/internal/testutil/testing"
	"server/internal/user"

	"gorm.io/gorm"
)

func setupContext(t *testing.T) (*tutil.TestContext, func()) {
	t.Helper()

	db := testutil.MustOpenDB(t)
	testStorage := storage.NewMockTestStorage()
	builder := rootapi.NewApiBuilder(db, testStorage, embeddings.NewMockEmbeddingsClient()).
		InitRepos().
		InitServices().
		InitHandlers("test-secret")

	createTestUsers(t, db, builder.UserService)
	server := builder.Build()
	server.InitRoutes(tutil.TestAuthMiddleware(builder.UserService))

	ctx := tutil.NewTestContext(
		t,
		server.Router(),
		builder.UserService,
		builder.ImageService(),
		builder.TagService(),
		testStorage,
	)

	return ctx, func() { testutil.CleanTestDB(db) }
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
			t.Fatalf("failed creating test user %s: %v", u.Username, err)
		}
	}
}
