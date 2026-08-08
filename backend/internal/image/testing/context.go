package testing

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	stdimage "image"
	_ "image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"server/internal/api"
	imgpkg "server/internal/image"
	"server/internal/storage"
	"server/internal/tag"
	userpkg "server/internal/user"
	"server/internal/testutil"
)

const testAuthHeader = "X-Test-User-ID"

// WithTestUser adds a test user ID header to a request
func WithTestUser(req *http.Request, userID uint64) *http.Request {
	req.Header.Set(testAuthHeader, fmt.Sprintf("%d", userID))
	return req
}

// TestAuthMiddleware creates a Gin middleware for test authentication
func TestAuthMiddleware(userSvc userpkg.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.GetHeader(testAuthHeader)
		if idStr == "" {
			c.Next()
			return
		}
		var id uint64
		fmt.Sscanf(idStr, "%d", &id)
		usr, err := userSvc.GetByID(uint(id))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("failed parsing test user: %v", err)})
			c.Abort()
			return
		}
		c.Set("user", usr)
		c.Next()
	}
}

// TestContext holds common test dependencies for image tests
type TestContext struct {
	T            *testing.T
	UserService  userpkg.UserService
	ImageService imgpkg.ImageService
	TagService   tag.TagService
	Router       *gin.Engine
	Storage      storage.Storage
}

// NewTestContext creates a fresh test context with seeded data
func NewTestContext(t *testing.T, router *gin.Engine, userSvc userpkg.UserService, imgSvc imgpkg.ImageService, tagSvc tag.TagService, storage storage.Storage) *TestContext {
	t.Helper()

	// Add auth middleware
	router.Use(TestAuthMiddleware(userSvc))

	mod, err := userSvc.GetByID(1)
	if err != nil {
		t.Fatalf("failed getting moderator user: %v", err)
	}
	SeedTestTags(t, tagSvc, []string{"animal", "cat", "dog", "bird", "test"}, mod)

	return &TestContext{
		T:            t,
		UserService:  userSvc,
		ImageService: imgSvc,
		TagService:   tagSvc,
		Router:       router,
		Storage:      storage,
	}
}

// AssertStatus checks response status code
func (c *TestContext) AssertStatus(rec *httptest.ResponseRecorder, wantStatus int) {
	c.T.Helper()
	if rec.Code != wantStatus {
		c.T.Errorf("got status %d, want %d, body=%s", rec.Code, wantStatus, rec.Body.String())
	}
}

// AssertStorageCount checks storage object count
func (c *TestContext) AssertStorageCount(expected int) {
	c.T.Helper()
	files, err := c.Storage.List(c.T.Context(), "")
	if err != nil {
		c.T.Errorf("failed retrieving store count: %v", err)
	}
	if len(files) != expected {
		c.T.Errorf("expected %d stored objects, got %d", expected, len(files))
	}
}

// AssertTags checks tags match expected (order-insensitive)
func AssertTags(t *testing.T, got, expected []string, context string) {
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

// AssertImageCount checks number of images in DB matching a query
func (c *TestContext) AssertImageCount(query *imgpkg.Query, expected int64) {
	c.T.Helper()
	count, err := c.ImageService.Count(query)
	if err != nil {
		c.T.Errorf("failed counting images: %v", err)
		return
	}
	if count != expected {
		c.T.Errorf("expected %d images, got %d", expected, count)
	}
}

// BuildUploadRequest builds a multipart upload request
func BuildUploadRequest(t *testing.T, url, metadataJSON, filename, contentType string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if err := w.WriteField("metadata", metadataJSON); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	if filename != "" && len(content) > 0 {
		if err := WriteFilePart(w, "file", filename, contentType, content); err != nil {
			t.Fatalf("write file part %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

// WriteFilePart writes a file part to multipart writer
func WriteFilePart(w *multipart.Writer, fieldName, filename, contentType string, content []byte) error {
	if fieldName == "" || filename == "" || len(content) == 0 {
		return fmt.Errorf("bad file data provided")
	}
	part, err := w.CreatePart(map[string][]string{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename)},
		"Content-Type":        {contentType},
	})
	if err != nil {
		return fmt.Errorf("failed to create form part: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		return fmt.Errorf("failed to write part content: %v", err)
	}
	return nil
}

// UploadImage uploads an image and returns the response recorder
func (c *TestContext) UploadImage(metadata, filename string, content []byte, userID uint64) *httptest.ResponseRecorder {
	c.T.Helper()
	req := BuildUploadRequest(c.T, "/upload", metadata, filename, "image/png", content)
	if userID != 0 {
		req = WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.Router.ServeHTTP(rec, req)
	return rec
}

// UploadImageWithResponse uploads and returns both recorder and parsed response
func (c *TestContext) UploadImageWithResponse(metadata, filename string, content []byte, userID uint64) (*httptest.ResponseRecorder, imgpkg.ImageResponse) {
	c.T.Helper()
	rec := c.UploadImage(metadata, filename, content, userID)

	var resp imgpkg.ImageResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			c.T.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, resp
}

// QueryImages performs a GET /image request with query string
func (c *TestContext) QueryImages(query string, userID uint64) *httptest.ResponseRecorder {
	c.T.Helper()
	url := "/image"
	if query != "" {
		url += "?" + query
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	if userID != 0 {
		req = WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.Router.ServeHTTP(rec, req)
	return rec
}

// QueryImagesWithResponse performs query and returns parsed response
func (c *TestContext) QueryImagesWithResponse(query string, userID uint64) (*httptest.ResponseRecorder, []imgpkg.ImageResponse) {
	c.T.Helper()
	rec := c.QueryImages(query, userID)

	var images []imgpkg.ImageResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &images); err != nil {
			c.T.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, images
}

// GetImageByName fetches an image by name
func (c *TestContext) GetImageByName(filename string, userID uint64) (*httptest.ResponseRecorder, imgpkg.ImageResponse) {
	c.T.Helper()
	req := httptest.NewRequest(http.MethodGet, "/image/"+filename, nil)
	if userID != 0 {
		req = WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.Router.ServeHTTP(rec, req)

	var resp imgpkg.ImageResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			c.T.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, resp
}

// GetRawImage fetches raw image data
func (c *TestContext) GetRawImage(filename string, userID uint64) *httptest.ResponseRecorder {
	c.T.Helper()
	req := httptest.NewRequest(http.MethodGet, "/image/raw/"+filename, nil)
	if userID != 0 {
		req = WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.Router.ServeHTTP(rec, req)
	return rec
}

// GetThumbnail fetches thumbnail
func (c *TestContext) GetThumbnail(filename string, userID uint64) *httptest.ResponseRecorder {
	c.T.Helper()
	req := httptest.NewRequest(http.MethodGet, "/image/thumb/"+filename, nil)
	if userID != 0 {
		req = WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.Router.ServeHTTP(rec, req)
	return rec
}

// SeedImage uploads an image for test setup (panics on failure)
func (c *TestContext) SeedImage(metadata, filename string, content []byte, userID uint64) imgpkg.ImageResponse {
	c.T.Helper()
	rec, resp := c.UploadImageWithResponse(metadata, filename, content, userID)
	if rec.Code != http.StatusOK {
		c.T.Fatalf("seedImage(%s) failed: status=%d body=%s", filename, rec.Code, rec.Body.String())
	}
	return resp
}

// SeedTestTags creates tags if they don't exist
func SeedTestTags(t *testing.T, svc tag.TagService, tags []string, usr *userpkg.User) {
	t.Helper()
	for _, tn := range tags {
		exists, err := svc.Exists(tn)
		if err != nil {
			t.Fatalf("failed to check tag %s: %v", tn, err)
		}
		if exists {
			continue
		}
		_, err = svc.Create(context.Background(), tn, usr)
		if err != nil {
			t.Fatalf("failed to seed tag %s: %v", tn, err)
		}
	}
}

// MakeTestPNG creates a simple test PNG of given dimensions
func MakeTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := stdimage.NewRGBA(stdimage.Rect(0, 0, w, h))
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("failed to encode test png: %v", err)
	}
	return buf.Bytes()
}

// MakeUniqueTestPNG creates a unique test PNG by varying dimensions
func MakeUniqueTestPNG(t *testing.T, baseW, baseH int, unique int) []byte {
	t.Helper()
	return MakeTestPNG(t, baseW+unique*10, baseH+unique*10)
}

// HashBytes returns SHA256 hex of data
func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

// NewTestAPI creates a test API instance with all services wired
func NewTestAPI(t *testing.T) (*TestContext, func()) {
	t.Helper()
	db := testutil.MustOpenDB(t)
	storage := storage.NewMockStorage()

	apiInstance := api.MustInitAPIForTest(db, storage)
	userSvc := apiInstance.UserService()
	tagSvc := apiInstance.TagService()
	imgSvc := apiInstance.ImageService()
	router := apiInstance.Router()

	ctx := NewTestContext(t, router, userSvc, imgSvc, tagSvc, storage)

	cleanup := func() {
		db.Exec("TRUNCATE TABLE image_tags, image_metadata, tags, users, audit_entries RESTART IDENTITY CASCADE")
	}

	return ctx, cleanup
}