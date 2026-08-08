package image

import (
	"fmt"
	"strings"

	"github.com/disintegration/imaging"
)

type Format imaging.Format

const (
	FormatPNG  Format = Format(imaging.PNG)
	FormatJPEG Format = Format(imaging.JPEG)
	FormatGIF  Format = Format(imaging.GIF)
)

var SupportedFormats = []Format{
	FormatPNG,
	FormatJPEG,
	FormatGIF,
}

func FormatFromMIME(mimeType string) (Format, error) {
	parts := strings.SplitN(mimeType, "/", 2)
	if len(parts) != 2 || parts[0] != "image" {
		return 0, fmt.Errorf("invalid MIME type: %s", mimeType)
	}
	return FormatFromExtension(parts[1])
}

func FormatFromExtension(ext string) (Format, error) {
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)

	switch ext {
	case "jpg", "jpeg":
		return FormatJPEG, nil
	case "png":
		return FormatPNG, nil
	case "gif":
		return FormatGIF, nil
	default:
		return 0, fmt.Errorf("unsupported extension: %s", ext)
	}
}

func (f Format) ImagingFormat() imaging.Format {
	return imaging.Format(f)
}

// without leading dot
func (f Format) Extension() string {
	switch f {
	case FormatJPEG:
		return "jpg"
	case FormatPNG:
		return "png"
	case FormatGIF:
		return "gif"
	default:
		return "bin"
	}
}

func (f Format) MIMEType() string {
	return "image/" + f.Extension()
}

func (f Format) String() string {
	return f.Extension()
}

func (f Format) IsSupported() bool {
	for _, sf := range SupportedFormats {
		if f == sf {
			return true
		}
	}
	return false
}

func ParseFormat(s string) (Format, error) {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if strings.HasPrefix(s, "image/") {
		return FormatFromMIME(s)
	}

	return FormatFromExtension(s)
}
