package api

import (
	"testing"

	"server/internal/storage"
	"server/internal/user"

	"github.com/gin-gonic/gin"

	testutil "server/internal/testutil"
	tutil "server/internal/testutil/testing"
)

// newTestAPI creates a test context with a fresh database and mock storage for API testing.
// Returns the test context and a cleanup function to truncate test tables.
func newTestAPI(t *testing.T) (*tutil.TestContext, func()) {
	t.Helper()
	db := testutil.MustOpenDB(t)
	storage := storage.NewMockStorage()

	// Create user service first so we can use it in test auth middleware
	userRepo := user.NewUserRepository(db)
	userSvc := user.NewUserService(userRepo)

	// Create test users
	createTestUsers(db, userSvc)

	// Test auth middleware that reads X-Test-User-ID header
	testAuthMiddleware := func(authGroup *gin.RouterGroup) {
		authGroup.Use(tutil.TestAuthMiddleware(userSvc))
	}

	apiInstance := MustInitAPIForTestWithAuth(db, storage, testAuthMiddleware, userSvc)
	tagSvc := apiInstance.TagService()
	imgSvc := apiInstance.ImageService()
	router := apiInstance.Router()

	ctx := tutil.NewTestContext(t, router, userSvc, imgSvc, tagSvc, storage)

	cleanup := func() {
		db.Exec("TRUNCATE TABLE image_tags, image_metadata, tags, users, audit_entries RESTART IDENTITY CASCADE")
	}

	return ctx, cleanup
}

// newTestContext creates a test context (alias for newTestAPI for backward compatibility)
func newTestContext(t *testing.T) (*tutil.TestContext, func()) {
	return newTestAPI(t)
}