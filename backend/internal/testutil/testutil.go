package testutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"server/internal/image"
	"server/internal/log"
	"server/internal/tag"
	"server/internal/user"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var sharedDSN string

// findEnvTestFile locates the .env.test file for test configuration
func findEnvTestFile() string {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, ".env.test")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		dir = filepath.Dir(dir)
	}
	return ".env.test"
}

// TestMain sets up the test database container.
// Call this from each feature's *_test.go: func TestMain(m *testing.M) { testutil.TestMain(m) }
func TestMain(m *testing.M) {
	ctx := context.Background()

	os.Setenv("ENV_FILE", findEnvTestFile())
	// Note: env.Init() should be called by the test if needed

	pgContainer, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}
	defer pgContainer.Terminate(ctx)

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}
	sharedDSN = dsn

	os.Exit(m.Run())
}

// SharedDSN returns the PostgreSQL DSN for the test container.
func SharedDSN() string {
	return sharedDSN
}

// MustOpenDB opens a GORM connection to the test database and runs AutoMigrate.
func MustOpenDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open(sharedDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Run AutoMigrate to create tables
	err = db.AutoMigrate(
		&image.ImageMetadata{},
		&tag.Tag{},
		&user.User{},
		&log.AuditEntry{},
	)
	if err != nil {
		t.Fatalf("failed to auto migrate database: %v", err)
	}

	return db
}