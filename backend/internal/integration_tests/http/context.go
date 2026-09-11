package http_integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	imgapi "server/internal/api/image"
	"server/internal/image"
	imgpkg "server/internal/image"
	"server/internal/storage"
	"server/internal/tag"
	"server/internal/testutil"
	userpkg "server/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const testAuthHeader = "X-Test-User-ID"

func WithTestUser(req *http.Request, userID uint64) *http.Request {
	req.Header.Set(testAuthHeader, fmt.Sprintf("%d", userID))
	return req
}

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

type TestContext struct {
	T            *testing.T
	UserService  userpkg.UserService
	ImageService imgpkg.ImageService
	TagService   tag.TagService
	Router       *gin.Engine
	Storage      storage.TestStorage

	seedImageCount int
	seededTags     map[string]struct{}
	setupUserID    uint64
}

func NewTestContext(
	t *testing.T,
	router *gin.Engine,
	userSvc userpkg.UserService,
	imgSvc imgpkg.ImageService,
	tagSvc tag.TagService,
	storage storage.TestStorage,
) *TestContext {
	t.Helper()

	return &TestContext{
		T:            t,
		UserService:  userSvc,
		ImageService: imgSvc,
		TagService:   tagSvc,
		Router:       router,
		Storage:      storage,
	}
}

func (c *TestContext) AssertStatus(rec *httptest.ResponseRecorder, wantStatus int) {
	c.T.Helper()
	require.Equalf(c.T, wantStatus, rec.Code, "got status %d, want %d, body=%s", rec.Code, wantStatus, rec.Body.String())
}

func (c *TestContext) AssertStorageCount(expected int) {
	c.T.Helper()
	files, err := c.Storage.List(c.T.Context(), "")
	require.NoErrorf(c.T, err, "failed retrieving store count: %v", err)
	require.Equalf(c.T, expected, len(files), "expected %d stored objects, got %d", expected, len(files))
}

func (c *TestContext) AssertTags(got, expected []string, context string) {
	c.T.Helper()
	require.Equal(c.T, got, expected)
	for _, exp := range expected {
		found := false
		for _, g := range got {
			if g == exp {
				found = true
				break
			}
		}

		require.Equal(c.T, found, true)
	}
}

func (c *TestContext) AssertImageCount(expected int64) {
	c.T.Helper()
	c.AssertImageCountQuery(nil, expected)
}

func (c *TestContext) AssertImageCountQuery(query *imgpkg.Query, expected int64) {
	c.T.Helper()
	count, err := c.ImageService.Count(query)
	require.NoError(c.T, err)
	require.Equal(c.T, count, expected)
}

func buildUploadRequest(t *testing.T, url string, postMetadata imgapi.ImagePostRequest, contentType string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if len(content) > 0 && postMetadata.Name != "" {
		if err := WriteFilePart(w, "file", postMetadata.Name, contentType, content); err != nil {
			t.Fatalf("write file part %v", err)
		}
	}

	if err := w.WriteField("name", postMetadata.Name); err != nil {
		t.Fatalf("failed to write name field: %v", err)
	}

	for _, tag := range postMetadata.Tags {
		if err := w.WriteField("tags", tag); err != nil {
			t.Fatalf("failed to write tag field: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

type TestFileData struct {
	image   image.UploadRequest
	format  image.Format
	content []byte
}

func buildBatchUploadRequest(t *testing.T, url string, metadataJSON string, files []TestFileData) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	if err := w.WriteField("metadata", metadataJSON); err != nil {
		t.Fatalf("failed to write metadata field: %v", err)
	}

	for _, file := range files {
		if err := WriteFilePart(w, "files", file.image.Name, file.format.MIMEType(), file.content); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

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

func (c *TestContext) UploadImage(postMetadata imgapi.ImagePostRequest, mimeType string, content []byte, userID uint64) *httptest.ResponseRecorder {
	c.T.Helper()
	req := buildUploadRequest(c.T, "/image/upload", postMetadata, mimeType, content)
	if userID != 0 {
		req = WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.Router.ServeHTTP(rec, req)
	return rec
}

func (c *TestContext) UploadImageWithResponse(postMetadata imgapi.ImagePostRequest, mimeType string, content []byte, userID uint64) (*httptest.ResponseRecorder, imgapi.ImageResponse) {
	c.T.Helper()
	rec := c.UploadImage(postMetadata, mimeType, content, userID)

	var resp imgapi.ImageResponse
	if rec.Code == http.StatusCreated {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			c.T.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, resp
}

func (c *TestContext) UploadImageBatch(fileDatas []TestFileData, commonTags []string, userID uint64) *httptest.ResponseRecorder {
	c.T.Helper()

	data := make([]image.UploadRequest, 0, len(fileDatas))
	for _, img := range fileDatas {
		data = append(data, img.image)
	}

	metadata := image.UploadBatchRequest{
		Data:       data,
		CommonTags: commonTags,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		c.T.Fatalf("failed to marshall request: %v", err)
	}
	metadataJSONstr := string(metadataJSON)

	req := buildBatchUploadRequest(c.T, "/image/upload/batch", metadataJSONstr, fileDatas)
	if userID != 0 {
		req = WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.Router.ServeHTTP(rec, req)
	return rec
}

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

func (c *TestContext) QueryImagesWithResponse(query string, userID uint64) (*httptest.ResponseRecorder, []imgapi.ImageResponse) {
	c.T.Helper()
	rec := c.QueryImages(query, userID)

	var images []imgapi.ImageResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &images); err != nil {
			c.T.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, images
}

func (c *TestContext) GetImageByName(filename string, userID uint64) (*httptest.ResponseRecorder, imgapi.ImageResponse) {
	c.T.Helper()
	req := httptest.NewRequest(http.MethodGet, "/image/"+filename, nil)
	if userID != 0 {
		req = WithTestUser(req, userID)
	}
	rec := httptest.NewRecorder()
	c.Router.ServeHTTP(rec, req)

	var resp imgapi.ImageResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			c.T.Fatalf("failed to unmarshal response: %v", err)
		}
	}
	return rec, resp
}

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

func (c *TestContext) SeedImage(postMetadata imgapi.ImagePostRequest, content []byte, userID uint64) imgapi.ImageResponse {
	c.T.Helper()
	rec, resp := c.UploadImageWithResponse(postMetadata, image.FormatPNG.MIMEType(), content, userID)
	require.Equal(c.T, rec.Code, http.StatusCreated, "seedImage(%s) failed: status=%d body=%s", postMetadata.Name, rec.Code, rec.Body.String())
	c.seedImageCount++
	return resp
}

func (c *TestContext) SeedImageByName(name string) imgapi.ImageResponse {
	c.T.Helper()
	postMetadata := imgapi.ImagePostRequest{
		Name: name,
	}
	rec, resp := c.UploadImageWithResponse(postMetadata, image.FormatPNG.MIMEType(), testutil.MakeUniqueTestPNG(c.T, 10, 10, c.seedImageCount), c.setupUserID)
	require.Equal(c.T, rec.Code, http.StatusCreated, "seedImage(%s) failed: status=%d body=%s", postMetadata.Name, rec.Code, rec.Body.String())
	c.seedImageCount++
	return resp
}

func (c *TestContext) SeedImageByNameAndTags(name string, tags ...string) imgapi.ImageResponse {
	c.T.Helper()
	newTags := make([]string, 0)
	for _, tag := range tags {
		if _, exists := c.seededTags[tag]; !exists {
			newTags = append(newTags, tag)
			c.seededTags[tag] = struct{}{}
		}
	}

	c.SeedTags(newTags...)

	postMetadata := imgapi.ImagePostRequest{
		Name: name,
		Tags: tags,
	}

	rec, resp := c.UploadImageWithResponse(postMetadata, image.FormatPNG.MIMEType(), testutil.MakeUniqueTestPNG(c.T, 10, 10, c.seedImageCount), c.setupUserID)
	require.Equal(c.T, rec.Code, http.StatusCreated, "seedImage(%s) failed: status=%d body=%s", postMetadata.Name, rec.Code, rec.Body.String())
	c.seedImageCount++
	return resp
}

func (c *TestContext) SeedTags(tags ...string) {
	c.T.Helper()
	for _, tn := range tags {
		exists, err := c.TagService.Exists(tn)
		if err != nil {
			c.T.Fatalf("failed to check tag %s: %v", tn, err)
		}
		if exists {
			continue
		}
		_, err = c.TagService.Create(context.Background(), tn, 1) //uses uid 1 - moderator user
		if err != nil {
			c.T.Fatalf("failed to seed tag %s: %v", tn, err)
		}
	}
}
