package storage

func ObjectKey(prefix, hash string, format string) string {
	return prefix + "/" + hash + "." + format
}

func ImageKey(hash string, format string) string {
	return ObjectKey("image", hash, format)
}

func ThumbnailKey(hash string, format string) string {
	return ObjectKey("thumb", hash, format)
}