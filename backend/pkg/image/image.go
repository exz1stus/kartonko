package image

import (
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

var SupportedFormats = []string{".jpeg", ".jpg", ".png", ".gif"}

func MIMETypeToFormat(mimeType string) (string, error) {
	parts := strings.SplitN(mimeType, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid MIME type: %s", mimeType)
	}

	return parts[1], nil
}

func GetDimensions(file *multipart.FileHeader) (uint, uint, error) {
	f, err := file.Open()
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}

	return uint(cfg.Width), uint(cfg.Height), nil
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

func GenerateThumbnail(srcPath, dstPath string) error {
	img, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	ext := filepath.Ext(dstPath)
	thumb := imaging.Fit(img, 320, 320, imaging.Lanczos)

	opts := imaging.JPEGQuality(75)

	if ext == ".png" {
		opts = imaging.PNGCompressionLevel(75)
	}

	return imaging.Save(
		thumb,
		dstPath,
		opts,
	)
}

func DeleteImage(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("Failed deleting image %s: %w", err)
	}

	return nil
}
