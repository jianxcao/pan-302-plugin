package main

import (
	"testing"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

func TestResourceFromEvent(t *testing.T) {
	event := pb.StrmEvent{
		EventId: "event-1",
		Event:   "strm.created",
		Strm:    &pb.StrmEventInfo{CloudPath: "/fallback/movie.mkv"},
		File: &pb.FileSnapshot{
			Hashes: map[string]string{
				"sha1":    "abc",
				"presha1": "def",
			},
			Size:     1024,
			Name:     "Show.S02E08.2025.2160p.HDR.WEB-DL.mkv",
			Path:     "/TV/Show.S02E08.2025.2160p.HDR.WEB-DL.mkv",
			PickCode: "pick",
		},
	}
	cfg := PluginConfig{
		NodeID: "test-node",
	}
	resource := resourceFromEvent(&event, cfg)
	assertEqual(t, "abc", resource.SHA1)
	assertEqual(t, "Show.S02E08.2025.2160p.HDR.WEB-DL", resource.Title)
	assertEqual(t, "tv", resource.Type)
	assertEqual(t, 2, resource.Season)
	assertEqual(t, 8, resource.Episode)
	assertEqual(t, "2025", resource.Year)
	assertEqual(t, "2160p HDR WEB-DL", resource.Quality)
	assertEqual(t, "test-node", resource.OwnerName)
	assertEqual(t, "cloud_resource.v1", resource.Schema)
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
