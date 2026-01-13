package media

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"server/pkg/image"
	"strings"
)

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
		if !image.IsFormatSupported(ext) {
			println("skipping unsupported file", entry.Name())
			continue
		}

		src := filepath.Join(uploadPath, entry.Name())
		name := strings.TrimSuffix(entry.Name(), ext)
		dst := filepath.Join(thumbPath, name+ext)

		if err := image.GenerateThumbnail(src, dst); err != nil {
			log.Printf("thumb failed for %s: %v", entry.Name(), err)
		}
	}
}
