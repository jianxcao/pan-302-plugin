package main

import (
	"testing"
)

func TestResourceFromEvent(t *testing.T) {
	event := StrmEvent{
		EventID: "event-1",
		Event:   "strm.created",
		Strm:    StrmInfo{CloudPath: "/fallback/movie.mkv"},
		File: &FileEvent{
			SHA1:     "abc",
			PreSHA1:  "def",
			Size:     1024,
			Name:     "Show.S02E08.2025.2160p.HDR.WEB-DL.mkv",
			Path:     "/TV/Show.S02E08.2025.2160p.HDR.WEB-DL.mkv",
			PickCode: "pick",
		},
	}
	resource := resourceFromEvent(event)
	assertEqual(t, "abc", resource.SHA1)
	assertEqual(t, "Show.S02E08.2025.2160p.HDR.WEB-DL", resource.Title)
	assertEqual(t, "tv", resource.Type)
	assertEqual(t, 2, resource.Season)
	assertEqual(t, 8, resource.Episode)
	assertEqual(t, "2025", resource.Year)
	assertEqual(t, "2160p HDR WEB-DL", resource.Quality)
}

func TestNormalizeCloudHubPath(t *testing.T) {
	assertEqual(t, "TV/Movie.mkv", normalizeCloudHubPath(`//TV\\Movie.mkv`))
}

func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}
