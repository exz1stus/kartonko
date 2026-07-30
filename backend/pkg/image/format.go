package image

import (
	"fmt"
	"strings"

	"github.com/disintegration/imaging"
)

// Format represents an image format with unified handling for
// MIME types, file extensions, and the imaging library's internal format.
// It wraps imaging.Format and provides string conversion methods.
type Format imaging.Format

// Supported formats
const (
	FormatPNG  Format = Format(imaging.PNG)
	FormatJPEG Format = Format(imaging.JPEG)
	FormatGIF  Format = Format(imaging.GIF)
)

// SupportedFormats is the list of all supported image formats.
var SupportedFormats = []Format{
	FormatPNG,
	FormatJPEG,
	FormatGIF,
}

// FormatFromMIME parses a MIME type (e.g., "image/jpeg") and returns the corresponding Format.
// Returns an error if the MIME type is invalid or unsupported.
func FormatFromMIME(mimeType string) (Format, error) {
	parts := strings.SplitN(mimeType, "/", 2)
	if len(parts) != 2 || parts[0] != "image" {
		return 0, fmt.Errorf("invalid MIME type: %s", mimeType)
	}
	return FormatFromExtension(parts[1])
}

// FormatFromExtension parses a file extension (e.g., "jpg", "png", "jpeg") and returns the corresponding Format.
// The extension can be with or without the leading dot.
// Returns an error if the extension is not supported.
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

// ImagingFormat converts this Format to the imaging library's format.
func (f Format) ImagingFormat() imaging.Format {
	return imaging.Format(f)
}

// Extension returns the canonical file extension for this format (without leading dot).
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

// MIMEType returns the MIME type for this format (e.g., "image/jpeg").
func (f Format) MIMEType() string {
	return "image/" + f.Extension()
}

// String returns the format as a string (alias for Extension for backward compatibility).
func (f Format) String() string {
	return f.Extension()
}

// IsSupported returns true if this format is in the SupportedFormats list.
func (f Format) IsSupported() bool {
	for _, sf := range SupportedFormats {
		if f == sf {
			return true
		}
	}
	return false
}

// ParseFormat parses a format string (extension or MIME type) and returns a Format.
// Accepts formats like "jpg", "jpeg", "png", "gif", "image/jpeg", "image/png", etc.
func ParseFormat(s string) (Format, error) {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	// Check if it's a MIME type
	if strings.HasPrefix(s, "image/") {
		return FormatFromMIME(s)
	}

	// Try as extension
	return FormatFromExtension(s)
}
