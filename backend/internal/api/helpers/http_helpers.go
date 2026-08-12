package helpers

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"server/internal/api/auth"
	"server/internal/errors"
	"server/internal/user"
	"strconv"

	"github.com/gin-gonic/gin"
)

func WriteFilePart(w *multipart.Writer, fieldName string, filename string, contentType string, content []byte) error {
	if fieldName == "" || filename == "" || len(content) == 0 {
		return fmt.Errorf("bad file data provided")
	}

	part, err := w.CreatePart(textproto.MIMEHeader{
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

func WrapBadRequest(err error) error {
	if err == nil {
		return nil
	}
	if err == errors.ErrBadRequest || err == errors.ErrUnauthorized || err == errors.ErrNotFound || err == errors.ErrPermissionDenied {
		return err
	}
	return fmt.Errorf("%w: %v", errors.ErrBadRequest, err)
}

func wrapUnauthorized(err error) error {
	if err == nil {
		return nil
	}
	if err == errors.ErrUnauthorized || err == errors.ErrBadRequest || err == errors.ErrNotFound || err == errors.ErrPermissionDenied {
		return err
	}
	return fmt.Errorf("%w: %v", errors.ErrUnauthorized, err)
}

func RespondJSON(c *gin.Context, status int, data any) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(status, data)
}

func RespondCreated(c *gin.Context, data any) {
	RespondJSON(c, http.StatusCreated, data)
}

func RespondMultiStatus(c *gin.Context, data any) {
	RespondJSON(c, http.StatusMultiStatus, data)
}

func WithUser(c *gin.Context, fn func(*user.User) error) {
	user, err := auth.GetUserFromContext(c)
	if err != nil {
		errors.RespondError(c, wrapUnauthorized(err))
		return
	}
	if err := fn(user); err != nil {
		errors.RespondError(c, err)
	}
}

func WithQuery[Q any](c *gin.Context, parseFn func(*gin.Context) (Q, error), fn func(Q) error) {
	query, err := parseFn(c)
	if err != nil {
		errors.RespondError(c, WrapBadRequest(err))
		return
	}
	if err := fn(query); err != nil {
		errors.RespondError(c, err)
	}
}

func WithUserAndQuery[Q any](c *gin.Context, parseFn func(*gin.Context) (Q, error), fn func(*user.User, Q) error) {
	user, err := auth.GetUserFromContext(c)
	if err != nil {
		errors.RespondError(c, wrapUnauthorized(err))
		return
	}
	query, err := parseFn(c)
	if err != nil {
		errors.RespondError(c, WrapBadRequest(err))
		return
	}
	if err := fn(user, query); err != nil {
		errors.RespondError(c, err)
	}
}

func GetCursorLimit(c *gin.Context) (cursor, limit int, err error) {
	cursorStr := c.Query("cursor")
	limitStr := c.Query("limit")

	cursor = 0
	limit = 50

	if cursorStr != "" {
		cursor, err = strconv.Atoi(cursorStr)
		if err != nil {
			return 0, 0, WrapBadRequest(fmt.Errorf("invalid cursor parameter: %w", err))
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return 0, 0, WrapBadRequest(fmt.Errorf("invalid limit parameter: %w", err))
		}
	}
	return cursor, limit, nil
}

func GetJSON[T any](c *gin.Context, dst *T, requiredFields ...string) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		return WrapBadRequest(err)
	}
	v := reflect.ValueOf(dst).Elem()
	for _, field := range requiredFields {
		f := v.FieldByName(field)
		if f.IsValid() && f.Kind() == reflect.String && f.String() == "" {
			return WrapBadRequest(fmt.Errorf("field %s is required", field))
		}
	}
	return nil
}

func HandleGet[T any](c *gin.Context, fetch func() (T, error), toResponse func(T) any) {
	item, err := fetch()
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	RespondJSON(c, http.StatusOK, toResponse(item))
}

func HandleList[T any](c *gin.Context, fetch func(cursor, limit int) ([]T, error), toResponse func(T) any) {
	cursor, limit, err := GetCursorLimit(c)
	if err != nil {
		errors.RespondError(c, err)
		return
	}

	items, err := fetch(cursor, limit)
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	resp := make([]any, len(items))
	for i, item := range items {
		resp[i] = toResponse(item)
	}
	RespondJSON(c, http.StatusOK, gin.H{"items": resp})
}

func HandleStream(c *gin.Context, fetch func(ctx context.Context) (io.ReadCloser, any, error), contentType func(any) string) {
	body, meta, err := fetch(c.Request.Context())
	if err != nil {
		errors.RespondError(c, err)
		return
	}
	defer body.Close()

	if contentType != nil && meta != nil {
		c.Header("Content-Type", contentType(meta))
	}
	io.Copy(c.Writer, body)
}

func HandleBatch[T any, R any](
	c *gin.Context,
	items []T,
	process func(T) (R, error),
	toResponse func(R) any,
	created bool,
) {
	var successes []R
	var failures []struct {
		Key   string `json:"key"`
		Error string `json:"error"`
	}

	for _, item := range items {
		res, err := process(item)
		if err != nil {
			key := extractKey(item)
			failures = append(failures, struct {
				Key   string `json:"key"`
				Error string `json:"error"`
			}{Key: key, Error: err.Error()})
			continue
		}
		successes = append(successes, res)
	}

	resp := struct {
		Successes []R `json:"successes"`
		Failures  []struct {
			Key   string `json:"key"`
			Error string `json:"error"`
		} `json:"failures,omitempty"`
	}{
		Successes: successes,
		Failures:  failures,
	}

	if len(failures) > 0 && len(successes) == 0 {
		RespondJSON(c, http.StatusBadRequest, resp)
		return
	}

	if len(failures) > 0 {
		RespondMultiStatus(c, resp)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	RespondJSON(c, status, resp)
}

func extractKey(item any) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	for _, fieldName := range []string{"Name", "ID", "Hash", "Key", "Filename"} {
		f := v.FieldByName(fieldName)
		if f.IsValid() && f.CanInterface() {
			return fmt.Sprintf("%v", f.Interface())
		}
	}
	return ""
}
