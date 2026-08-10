package image

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/jpeg"
)

func GetDimensionsBytes(data []byte) (uint, uint, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	return uint(cfg.Width), uint(cfg.Height), err
}

func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}
