package image

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
)

func GetDimensionsBytes(data []byte) (uint, uint, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	return uint(cfg.Width), uint(cfg.Height), err
}

func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

// GenerateThumbnail creates a thumbnail from image data
// - PNG: preserves alpha channel
// - GIF: preserves animation
// - JPEG: standard JPEG thumbnail
func GenerateThumbnail(data []byte, format Format) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	const thumbDimension = 320
	thumb := imaging.Fit(img, thumbDimension, thumbDimension, imaging.Lanczos)

	var buf bytes.Buffer
	switch format {
	case FormatGIF:
		return data, nil
	case FormatPNG:
		err = imaging.Encode(&buf, thumb, imaging.PNG, imaging.PNGCompressionLevel(png.DefaultCompression))
	default: // JPEG and others
		err = imaging.Encode(&buf, thumb, imaging.JPEG, imaging.JPEGQuality(75))
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
