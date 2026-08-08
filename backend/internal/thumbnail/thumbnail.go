package thumbnail

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"path/filepath"
	"strings"

	"server/internal/storage"
	imgpkg "server/internal/image"

	"github.com/disintegration/imaging"
)

const imagePrefix = "image/"

// RegenerateThumbnails walks every object under the "image/" prefix in the
// storage backend, regenerates its thumbnail, and re-uploads it to the
// "thumb/" prefix. It is fully storage-driven: the only thing it needs from
// the caller is a storage.Storage implementation.
func RegenerateThumbnails(ctx context.Context, store storage.Storage) error {
	objects, err := store.List(ctx, imagePrefix)
	if err != nil {
		return fmt.Errorf("failed listing image objects: %w", err)
	}

	for _, obj := range objects {
		ext := strings.ToLower(filepath.Ext(obj.Key))
		if ext == "" {
			continue
		}
		format, err := imgpkg.FormatFromExtension(ext)

		if !format.IsSupported() {
			log.Printf("skipping unsupported image object %s", obj.Key)
			continue
		}

		body, err := store.Download(ctx, obj.Key)
		if err != nil {
			log.Printf("failed downloading %s: %v", obj.Key, err)
			continue
		}
		data, err := io.ReadAll(body)
		body.Close()
		if err != nil {
			log.Printf("failed reading %s: %v", obj.Key, err)
			continue
		}

		thumb, err := GenerateThumbnail(data, format)
		if err != nil {
			log.Printf("failed generating thumbnail for %s: %v", obj.Key, err)
			continue
		}

		base := filepath.Base(obj.Key)
		hash := strings.TrimSuffix(base, ext)

		thumbKey := storage.ThumbnailKey(hash, format.String())
		if err := store.Upload(ctx, thumbKey, bytes.NewReader(thumb), "image/jpeg"); err != nil {
			log.Printf("failed uploading thumbnail %s: %v", thumbKey, err)
			continue
		}
	}

	return nil
}

// - PNG: preserves alpha channel
// - GIF: preserves animation
// - JPEG: standard JPEG thumbnail
func GenerateThumbnail(data []byte, format imgpkg.Format) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	const thumbDimension = 320
	thumb := imaging.Fit(img, thumbDimension, thumbDimension, imaging.Lanczos)

	var buf bytes.Buffer
	switch format {
	case imgpkg.FormatGIF:
		return data, nil
	case imgpkg.FormatPNG:
		err = imaging.Encode(&buf, thumb, imaging.PNG, imaging.PNGCompressionLevel(png.DefaultCompression))
	default: // JPEG and others
		err = imaging.Encode(&buf, thumb, imaging.JPEG, imaging.JPEGQuality(75))
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
