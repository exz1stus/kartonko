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
	"server/internal/env"
	"server/internal/models"
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

func makeFileDataFromMetadata(t *testing.T, metadataJSON string, content func(t *testing.T) []byte) []TestFileData {
	var metadata []ImageMetadata
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

func buildBatchUploadRequest(t *testing.T, url string, metadataJSON string, files []TestFileData) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if err := w.WriteField("metadata", metadataJSON); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	for _, file := range files {
		if file.filename != "" {
			writeFilePart(t, w, "files", file)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func buildUploadRequest(t *testing.T, url, metadataJSON, filename, contentType string, content []byte) *http.Request {
	fileData := TestFileData{filename, contentType, content}

	return buildUploadRequestFileData(t, url, metadataJSON, fileData)
}

func buildUploadRequestFileData(t *testing.T, url string, metadataJSON string, file TestFileData) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if err := w.WriteField("metadata", metadataJSON); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	if file.filename != "" {
		writeFilePart(t, w, "file", file)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func seedImageUsingRequest(t *testing.T, api *api, metadata string, filename string, userID uint64) {
	t.Helper()

	r := newTestRouter(api)

	req := buildUploadRequest(t, "/upload", metadata, filename, "image/png", makeTestPNG(t, 5, 5))
	req = withTestUser(req, userID)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("seedImage(%s) failed: status=%d body=%s", filename, rec.Code, rec.Body.String())
	}
}

func seedTestTags(t *testing.T, model *models.TagModel, tags []string) {
	t.Helper()

	for _, tag := range tags {
		_, err := model.CreateTag(tag)
		if err != nil {
			t.Fatalf("failed to seed tag %s: %v", tag, err)
		}
	}
}

func seedTestUsers(t *testing.T, model *models.UserModel, users []models.User) {
	t.Helper()

	for _, u := range users {
		_, err := model.CreateUser(&u)
		if err != nil {
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
		user, err := api.models.Users.GetUserById(id)
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
	r.GET("/images", a.GetImagesByQuery)
	r.DELETE("/image/:name", a.DeleteImageByName)
	r.DELETE("/images", a.DeleteImagesByQuery)
	r.POST("/upload", a.PostImage)
	r.POST("/upload/batch", a.PostImagesBatch)

	return r
}

func newTestAPI(t *testing.T) (*api, *storage.MockStorage) {
	t.Helper()

	gormModels, err := models.InitGorm(postgres.Open(sharedDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to init gorm: %v", err)
	}
	storage := storage.NewMockStorage()
	api := &api{models: gormModels, storage: storage, jwtSecret: env.GetEnvString("JWT_SECRET")}
	api.router = newTestRouter(api)

	users := []models.User{
		{Model: gorm.Model{ID: 1}, Username: "mod", Privileage: models.Moderator, Email: "mod@test.local", ProviderID: "test-provider-1"},
		{Model: gorm.Model{ID: 2}, Username: "alice", Privileage: models.Unprivileaged, Email: "alice@test.local", ProviderID: "test-provider-2"},
		{Model: gorm.Model{ID: 3}, Username: "bob", Privileage: models.Unprivileaged, Email: "bob@test.local", ProviderID: "test-provider-3"},
	}

	seedTestUsers(t, gormModels.Users, users)

	// Postgres persists across tests in the same container, so each test
	// must clean up after itself. Truncate everything before each test
	// rather than dropping/recreating the schema (much faster).
	t.Cleanup(func() {
		gormModels.Images.Db.Exec("TRUNCATE TABLE image_tags, image_metadata, tags, users, audit_entries RESTART IDENTITY CASCADE")
	})

	return api, storage
}
