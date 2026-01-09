package image

import (
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
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
	img, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	thumb := imaging.Fit(img, 320, 320, imaging.Lanczos)

	return imaging.Save(
		thumb,
		dstPath,
		imaging.JPEGQuality(75),
	)
}

func RegenerateThumbnails(uploadPath, thumbPath string) {
	println("regenerating thumbnails...")
	_ = os.RemoveAll(thumbPath)
	_ = os.MkdirAll(thumbPath, 0755)

	entries, err := os.ReadDir(uploadPath)
	if err != nil {
		fmt.Printf("failed to read uploads dir: %v", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !IsFormatSupported(ext) {
			println("skipping unsupported file", entry.Name())
			continue
		}

		src := filepath.Join(uploadPath, entry.Name())
		name := strings.TrimSuffix(entry.Name(), ext)
		dst := filepath.Join(thumbPath, name+".jpg")

		if err := GenerateThumbnail(src, dst); err != nil {
			log.Printf("thumb failed for %s: %v", entry.Name(), err)
		}
	}
}
