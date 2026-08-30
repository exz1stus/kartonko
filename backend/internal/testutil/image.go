package testutil

import (
	"bytes"
	stdimage "image"
	"image/png"
	"testing"
)

func MakeTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := stdimage.NewRGBA(stdimage.Rect(0, 0, w, h))
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("failed to encode test png: %v", err)
	}
	return buf.Bytes()
}

func MakeUniqueTestPNG(t *testing.T, baseW, baseH int, unique int) []byte {
	t.Helper()
	return MakeTestPNG(t, baseW+unique*10, baseH+unique*10)
}
