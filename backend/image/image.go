package image

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
)

var SupportedFormats = []string{"jpeg", "png"}

func MIMETypeToFormat(mimeType string) (string, error) {
	parts := strings.SplitN(mimeType, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid MIME type: %s", mimeType)
	}

	return parts[1], nil
}

func IsFormatSupported(format string) bool {
	for _, supportedFormat := range SupportedFormats {
		if format == supportedFormat {
			return true
		}
	}
	return false
}

func HashFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.ReadAll(io.TeeReader(src, hasher)); err != nil {
		return "", err
	}
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	return hash, nil
}
