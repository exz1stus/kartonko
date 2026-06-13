package image

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/disintegration/imaging"
)

type ImageError struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

var SupportedFormats = []string{"jpeg", "jpg", "png", "gif"}

const THUMBNAILS_FORMAT = imaging.JPEG

func MIMETypeToFormat(mimeType string) (string, error) {
	parts := strings.SplitN(mimeType, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid MIME type: %s", mimeType)
	}

	return parts[1], nil
}

func imagingExtToString(ext imaging.Format) string {
	switch ext {
	case imaging.JPEG:
		return "jpg"
	case imaging.GIF:
		return "gif"
	case imaging.PNG:
		return "png"
	default:
		return ""
	}
}

func ObjectKey(prefix string, hash string, ext string) string {
	return prefix + "/" + hash + "." + ext
}

func ImageKey(hash string, ext string) string {
	return ObjectKey("image", hash, ext)
}

func ThumbnailKey(hash string) string {
	return ObjectKey("thumb", hash, imagingExtToString(THUMBNAILS_FORMAT))
}

func GetDimensionsBytes(data []byte) (uint, uint, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	return uint(cfg.Width), uint(cfg.Height), err
}

func IsFormatSupported(format string) bool {
	for _, supportedFormat := range SupportedFormats {
		if format == supportedFormat {
			return true
		}
	}
	return false
}

func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

func GenerateThumbnail(data []byte, ext string) ([]byte, error) {
	if ext == ".gif" {
		return data, nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	thumb := imaging.Fit(img, 320, 320, imaging.Lanczos)
	var buf bytes.Buffer
	var opts imaging.EncodeOption = imaging.JPEGQuality(75)
	if ext == ".png" {
		opts = imaging.PNGCompressionLevel(75)
	}
	if err := imaging.Encode(&buf, thumb, THUMBNAILS_FORMAT, opts); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
