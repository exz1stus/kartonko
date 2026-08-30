package repository_integration_tests

import (
	"context"
	"fmt"
	"os"
	"server/internal/database"
	"server/internal/integration_tests"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupPostgreContext(t *testing.T) (*RepoTestContext, func()) {
	db, err := database.InitGorm(postgres.Open(integration_tests.SharedDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to init gorm: %s", err.Error())
	}
	return &RepoTestContext{db}, func() { integration_tests.CleanTestDB(db) }
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
