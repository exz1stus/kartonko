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

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync()
}

func GenerateThumbnail(srcPath, dstPath string) error {
	img, err := imaging.Open(srcPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	ext := filepath.Ext(dstPath)
	thumb := imaging.Fit(img, 320, 320, imaging.Lanczos)

	if ext == ".gif" {
		return CopyFile(srcPath, dstPath)
	}

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
		return fmt.Errorf("Failed deleting image %s: %w", path, err)
	}

	return nil
}

func DeleteImagesWithName(name string, path string) error {
	files, err := filepath.Glob(filepath.Join(path, name+"*"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("image %s not found in path %s", name, path)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("Failed deleting image %s: %w", file, err)
		}
	}

	return nil
}
