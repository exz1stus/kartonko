package image

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	"image/png"

	"github.com/disintegration/imaging"
)

func ObjectKey(prefix, hash string, format Format) string {
	return prefix + "/" + hash + "." + format.Extension()
}

func ImageKey(hash string, format Format) string {
	return ObjectKey("image", hash, format)
}

func ThumbnailKey(hash string, format Format) string {
	return ObjectKey("thumb", hash, format)
}

func GetDimensionsBytes(data []byte) (uint, uint, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	return uint(cfg.Width), uint(cfg.Height), err
}

func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

// GenerateThumbnail generates a thumbnail preserving the source format:
// - PNG: preserves alpha channel
// - GIF: preserves animation (resizes all frames)
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
		// Decode all frames, resize each, re-encode as animated GIF
		return generateGIFThumbnail(data, thumbDimension)
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

// generateGIFThumbnail decodes all frames of a GIF, resizes them, and re-encodes as animated GIF.
func generateGIFThumbnail(data []byte, maxDim int) ([]byte, error) {
	reader := bytes.NewReader(data)
	gifImg, err := gif.DecodeAll(reader)
	if err != nil {
		return nil, err
	}

	// Resize each frame
	for i, frame := range gifImg.Image {
		bounds := frame.Bounds()
		// Calculate scale factor
		scale := float64(maxDim) / float64(max(bounds.Dx(), bounds.Dy()))
		if scale >= 1 {
			continue // already small enough
		}
		newW := int(float64(bounds.Dx()) * scale)
		newH := int(float64(bounds.Dy()) * scale)
		resized := imaging.Resize(frame, newW, newH, imaging.Lanczos)
		// Create new paletted image at correct position
		newFrame := image.NewPaletted(image.Rect(0, 0, newW, newH), frame.Palette)
		draw.Draw(newFrame, newFrame.Bounds(), resized, image.Point{}, draw.Src)
		gifImg.Image[i] = newFrame
	}
	gifImg.Config.Width = gifImg.Image[0].Bounds().Dx()
	gifImg.Config.Height = gifImg.Image[0].Bounds().Dy()

	var buf bytes.Buffer
	err = gif.EncodeAll(&buf, gifImg)
	return buf.Bytes(), err
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
