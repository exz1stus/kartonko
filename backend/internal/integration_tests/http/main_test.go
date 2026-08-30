package http_integration_tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"server/internal/api"
	"server/internal/database"
	embeddings "server/internal/embedding"
	"server/internal/integration_tests"
	"server/internal/storage"
	"server/internal/user"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupContext(t *testing.T) (*TestContext, func()) {
	t.Helper()

	db, err := database.InitGorm(postgres.Open(integration_tests.SharedDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to init gorm: %s", err.Error())
	}
	testStorage := storage.NewMockTestStorage()
	app := api.NewApp(db, testStorage, embeddings.NewMockEmbeddingsClient())
	api := api.NewAPI(app, "test-secret")
	api.InitRoutes(TestAuthMiddleware(app.UserService))

	createTestUsers(t, db, app.UserService)

	ctx := NewTestContext(
		t,
		api.Router(),
		app.UserService,
		app.ImageService,
		app.TagService,
		testStorage,
	)

	return ctx, func() { integration_tests.CleanTestDB(db) }
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

func TestMain(m *testing.M) {
	ctx := context.Background()
	env := integration_tests.FindEnvTestFile()
	fmt.Printf("Test env: %s\n", env)
	os.Setenv("ENV_FILE", env)

	pgContainer := integration_tests.MustRunPostgreSQLContainer(ctx)
	defer pgContainer.Terminate(ctx)

	os.Exit(m.Run())
}
