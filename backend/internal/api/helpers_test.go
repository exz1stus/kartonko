package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"runtime"
	"server/internal/api/dto"
	"server/internal/database"
	"server/internal/env"
	"server/internal/models"
	"server/internal/repositories"
	"server/internal/services"
	"server/internal/storage"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var sharedDSN string

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

func TestMain(m *testing.M) {
	ctx := context.Background()

	os.Setenv("ENV_FILE", findEnvTestFile())
	env.Init()

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

type TestFileData struct {
	filename    string
	contentType string
	content     []byte
}

func writeFilePart(t *testing.T, w *multipart.Writer, fieldName string, file TestFileData) {
	t.Helper()
	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, file.filename)},
		"Content-Type":        {file.contentType},
	})
	if err != nil {
		t.Fatalf("failed to create form part: %v", err)
	}
	if _, err := part.Write(file.content); err != nil {
		t.Fatalf("failed to write part content: %v", err)
	}
}

func makeTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("failed to encode test png : %v", err)
	}
	return buf.Bytes()
}

// makeUniqueTestPNG creates a test PNG with unique dimensions to avoid duplicate hash errors
func makeUniqueTestPNG(t *testing.T, baseW, baseH int, unique int) []byte {
	t.Helper()
	return makeTestPNG(t, baseW+unique*10, baseH+unique*10)
}

func makeFileDataFromMetadata(t *testing.T, metadataJSON string, content func(t *testing.T) []byte) []TestFileData {
	var metadata []dto.ImagePostRequest
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		t.Fatalf("failed to unmarshall test metadata json : %v", err)
	}

	fileData := make([]TestFileData, len(metadata))

	for i, meta := range metadata {
		splittedStr := strings.Split(meta.Name, ".")
		if len(splittedStr) <= 1 {
			t.Fatalf("failed to extract mime type from image name: ")
		}

		extractedFormat := splittedStr[1]

		fileData[i] = TestFileData{
			filename:    meta.Name,
			contentType: fmt.Sprintf("image/%s", extractedFormat),
			content:     content(t),
		}
	}

	return fileData
}

func seedImageUsingRequest(t *testing.T, r *gin.Engine, metadata string, filename string, content []byte, userID uint64) {
	t.Helper()

	req := buildUploadRequest(t, "/upload", metadata, filename, "image/png", content)
	req = withTestUser(req, userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("seedImage(%s) failed: status=%d body=%s", filename, rec.Code, rec.Body.String())
	}
}

func seedTestTags(t *testing.T, service services.TagService, tags []string, user *models.User) {
	t.Helper()

	for _, tag := range tags {
		_, err := service.Create(context.Background(), tag, user)
		if err != nil {
			t.Fatalf("failed to seed tag %s: %v", tag, err)
		}
	}
}

func seedTestUsers(t *testing.T, service services.UserService, users []models.User) {
	t.Helper()

	for _, u := range users {
		if err := service.Create(&u); err != nil {
			t.Fatalf("failed to seed user %s: %v", u.Username, err)
		}
	}
}

const testUserHeader = "X-Test-User-ID"

func withTestUser(req *http.Request, userID uint64) *http.Request {
	req.Header.Set(testUserHeader, fmt.Sprintf("%d", userID))
	return req
}

func testAuthMiddleware(api *api) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.GetHeader(testUserHeader)
		if idStr == "" {
			c.Next()
			return
		}
		var id uint64
		fmt.Sscanf(idStr, "%d", &id)
		user, err := api.userService.GetByID(uint(id))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("failed parsing test user from context: %v", err)})
			c.Abort()
		}
		c.Set("user", user)
		c.Next()
	}
}

func newTestRouter(a *api) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(testAuthMiddleware(a))

	r.GET("/image/:name", a.GetImageByName)
	r.GET("/image/hash/:hash", a.GetImageByHash)
	r.GET("/image/id/:id", a.GetImageByID)
	r.GET("/image/raw/:name", a.GetRawImageByName)
	r.GET("/image/thumb/:name", a.GetRawThumbnailByName)
	r.GET("/image", a.GetImagesByQuery)
	r.DELETE("/image/:name", a.DeleteImageByName)
	r.DELETE("/image", a.DeleteImagesByQuery)
	r.POST("/upload", a.PostImage)
	r.POST("/upload/batch", a.PostImagesBatch)

	return r
}

func newTestAPI(t *testing.T) *api {
	t.Helper()

	db, err := database.InitGorm(postgres.Open(sharedDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to init gorm: %v", err)
	}
	storage := storage.NewMockStorage()

	imageRepo := repositories.NewImageRepository(db)
	userRepo := repositories.NewUserRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	logRepo := repositories.NewLogRepository(db)

	logService := services.NewLogService(logRepo)
	userService := services.NewUserService(userRepo, imageRepo)
	imageService := services.NewImageService(db, imageRepo, logService, storage)
	tagService := services.NewTagService(db, tagRepo, logService)

	api := &api{
		imageService: imageService,
		userService:  userService,
		tagService:   tagService,
		logService:   logService,

		storage:   storage,
		jwtSecret: env.GetEnvString("JWT_SECRET"),
		db:        db,
	}

	api.router = newTestRouter(api)

	users := []models.User{
		{Model: gorm.Model{ID: 1}, Username: "mod", Privilege: models.Moderator, Email: "mod@test.local", ProviderID: "test-provider-1"},
		{Model: gorm.Model{ID: 2}, Username: "alice", Privilege: models.Unprivileged, Email: "alice@test.local", ProviderID: "test-provider-2"},
		{Model: gorm.Model{ID: 3}, Username: "bob", Privilege: models.Unprivileged, Email: "bob@test.local", ProviderID: "test-provider-3"},
	}

	seedTestUsers(t, api.userService, users)

	// Postgres persists across tests in the same container
	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE image_tags, image_metadata, tags, users, audit_entries RESTART IDENTITY CASCADE")
	})

	return api
}

// ============================================================================
// Common Test Contexts & Helpers
// ============================================================================

// testContext holds common test dependencies for all test types
type testContext struct {
	t       *testing.T
	a       *api
	r       *gin.Engine
	store   *storage.MockStorage
	mod     *models.User
}

// newTestContext creates a fresh test context with seeded tags
func newTestContext(t *testing.T) *testContext {
	t.Helper()
	a := newTestAPI(t)
	r := newTestRouter(a)
	store := a.storage.(*storage.MockStorage)

	mod, err := a.userService.GetByID(1)
	if err != nil {
		t.Fatalf("failed getting moderator user")
	}
	seedTestTags(t, a.tagService, []string{"animal", "cat", "dog", "bird", "test"}, mod)

	return &testContext{t: t, a: a, r: r, store: store, mod: mod}
}

// assertStatus checks response status
func (c *testContext) assertStatus(rec *httptest.ResponseRecorder, wantStatus int) {
	c.t.Helper()
	if rec.Code != wantStatus {
		c.t.Errorf("got status %d, want %d, body=%s", rec.Code, wantStatus, rec.Body.String())
	}
}

// assertStorageCount checks storage object count
func (c *testContext) assertStorageCount(expected int) {
	c.t.Helper()
	count, err := c.store.Count("")
	if err != nil {
		c.t.Errorf("failed retrieving store images count: %v", err)
	}
	if count != expected {
		c.t.Errorf("expected %d stored objects, got %d", expected, count)
	}
}

// assertTags checks tags match expected (order-insensitive)
func assertTags(t *testing.T, got, expected []string, context string) {
	t.Helper()
	if len(got) != len(expected) {
		t.Errorf("%s: expected %d tags, got %d: %v", context, len(expected), len(got), got)
		return
	}
	for _, exp := range expected {
		found := false
		for _, g := range got {
			if g == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: expected tag %q, got %v", context, exp, got)
		}
	}
}

// assertImageCount checks number of images in DB matching a query
func (c *testContext) assertImageCount(query *models.ImageQuery, expected int64) {
	c.t.Helper()
	count, err := c.a.imageService.Count(query)
	if err != nil {
		c.t.Errorf("failed counting images: %v", err)
		return
	}
	if count != expected {
		c.t.Errorf("expected %d images, got %d", expected, count)
	}
}

// uploadImage uploads an image and returns the response recorder
func (c *testContext) uploadImage(metadata, filename string, content []byte, userID uint64) *httptest.ResponseRecorder {
	c.t.Helper()
	req := buildUploadRequest(c.t, "/upload", metadata, filename, "image/png", content)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
}

// uploadImageWithResponse uploads and returns both recorder and parsed response
func (c *testContext) uploadImageWithResponse(metadata, filename string, content []byte, userID uint64) (*httptest.ResponseRecorder, dto.ImageResponse) {
	c.t.Helper()
	rec := c.uploadImage(metadata, filename, content, userID)

	var resp dto.ImageResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			c.t.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, resp
}

// queryImages performs a GET /image request with query string
func (c *testContext) queryImages(query string, userID uint64) *httptest.ResponseRecorder {
	c.t.Helper()
	url := "/image"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
}

// queryImagesWithResponse performs query and returns parsed response
func (c *testContext) queryImagesWithResponse(query string, userID uint64) (*httptest.ResponseRecorder, []dto.ImageResponse) {
	c.t.Helper()
	rec := c.queryImages(query, userID)

	var images []dto.ImageResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &images); err != nil {
			c.t.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, images
}

// getImageByName fetches an image by name
func (c *testContext) getImageByName(filename string, userID uint64) (*httptest.ResponseRecorder, dto.ImageResponse) {
	c.t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/image/"+filename, nil)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)

	var resp dto.ImageResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			c.t.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, resp
}

// getRawImage fetches raw image data
func (c *testContext) getRawImage(filename string, userID uint64) *httptest.ResponseRecorder {
	c.t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/image/raw/"+filename, nil)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
}

// getThumbnail fetches thumbnail
func (c *testContext) getThumbnail(filename string, userID uint64) *httptest.ResponseRecorder {
	c.t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/image/thumb/"+filename, nil)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
}

// deleteImageByName deletes a single image
func (c *testContext) deleteImageByName(filename string, userID uint64) *httptest.ResponseRecorder {
	c.t.Helper()
	req := buildDeleteRequest(c.t, "/image", filename)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
}

// deleteImagesByQuery deletes images by query
func (c *testContext) deleteImagesByQuery(query string, userID uint64) *httptest.ResponseRecorder {
	c.t.Helper()
	url := "/image"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodDelete, url, http.NoBody)
	if userID != 0 {
		req = withTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.r.ServeHTTP(rec, req)
	return rec
}

// seedImage uploads an image for test setup (panics on failure)
func (c *testContext) seedImage(metadata, filename string, content []byte, userID uint64) dto.ImageResponse {
	c.t.Helper()
	rec, resp := c.uploadImageWithResponse(metadata, filename, content, userID)
	if rec.Code != http.StatusOK {
		c.t.Fatalf("seedImage(%s) failed: status=%d body=%s", filename, rec.Code, rec.Body.String())
	}
	return resp
}

// buildDeleteRequest creates a DELETE request for a single image
func buildDeleteRequest(t *testing.T, url, filename string) *http.Request {
	t.Helper()
	resourceUrl := url + "/" + filename
	req := httptest.NewRequest(http.MethodDelete, resourceUrl, http.NoBody)
	return req
}
