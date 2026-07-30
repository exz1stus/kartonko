package image

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"strings"

	"github.com/disintegration/imaging"
)

var SupportedFormats = []string{"jpeg", "jpg", "png", "gif"}

const THUMBNAILS_FORMAT = imaging.JPEG

func MIMETypeToFormat(mimeType string) (string, error) {
	parts := strings.SplitN(mimeType, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid MIME type: %s", mimeType)
	}

	return parts[1], nil
}

func ObjectKey(prefix string, hash string, format string) string {
	return prefix + "/" + hash + "." + format
}

func ImageKey(hash string, format string) string {
	return ObjectKey("image", hash, format)
}

func ThumbnailKey(hash string, format string) string {
	return ObjectKey("thumb", hash, format)
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

func GenerateThumbnail(data []byte, format imaging.Format) ([]byte, error) {
	if format == imaging.GIF {
		return data, nil
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	const thumbDimension = 320
	thumb := imaging.Fit(img, thumbDimension, thumbDimension, imaging.Lanczos)
	var buf bytes.Buffer
	switch format {
	case imaging.JPEG:
		err = imaging.Encode(&buf, thumb, imaging.JPEG, imaging.JPEGQuality(75))
	case imaging.PNG:
		err = imaging.Encode(&buf, thumb, imaging.PNG, imaging.PNGCompressionLevel(png.DefaultCompression))
	default:
		err = imaging.Encode(&buf, thumb, THUMBNAILS_FORMAT)
	}

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
