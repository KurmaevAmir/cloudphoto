package main

import "strings"

const (
	photosPrefix = "photos/"
	maxKeyLen    = 1024
)

func albumPrefix(album string) string {
	return photosPrefix + album + "/"
}

func photoKey(album, photo string) string {
	return albumPrefix(album) + photo
}

func isPhotoFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg")
}
