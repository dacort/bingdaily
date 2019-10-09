package main

import "testing"

func TestWallpaperMetadata(t *testing.T) {
	bwData, err := getLatestWallpaperMetadata()
	if err != nil {
		t.Error("Couldn't download metadata", err)
	}

	if bwData.ImageURL == "" {
		t.Error("We should have an image URL, but we don't...")
	}
}
