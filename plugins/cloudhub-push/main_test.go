package main

import (
	"encoding/json"
	"testing"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

func TestNormalizeCloudHubPath(t *testing.T) {
	assertEqual(t, "TV/Movie.mkv", normalizeCloudHubPath(`//TV\\Movie.mkv`))
}

func TestResourceFromMediaEvent(t *testing.T) {
	event := pb.MediaEvent{
		EventId: "event-1",
		Event:   "media.item.added",
		Item: &pb.MediaItemInfo{
			Name:              "姜黎非与大师兄相认",
			SeriesName:        "千香",
			ProductionYear:    2026,
			IndexNumber:       13,
			ParentIndexNumber: 1,
			Container:         "mkv",
			Bitrate:           2066995,
			Width:             3840,
			Height:            2160,
			RuntimeTicks:      27380040000,
			VideoCodec:        "hevc",
			Fps:               23.976,
		},
		Resource: &pb.FileSnapshot{
			Hashes: map[string]string{"sha1": "abc", "presha1": "should-not-send"},
			Size:   707430281,
			Name:   "千香 - S01E13 - 第 13 集.mkv",
			Path:   "/disk/media/tv/china/千香 (2026)/Season 1/千香 - S01E13 - 第 13 集.mkv",
		},
	}

	resource := resourceFromMediaEvent(&event, PluginConfig{NodeID: "test-node"})
	assertEqual(t, "abc", resource.SHA1)
	assertEqual(t, "千香", resource.Title)
	assertEqual(t, "千香 - S01E13 - 第 13 集.mkv", resource.Name)
	assertEqual(t, "tv", resource.Type)
	assertEqual(t, 1, resource.Season)
	assertEqual(t, 13, resource.Episode)
	assertEqual(t, "2026", resource.Year)
	assertEqual(t, "/disk/media/tv/china/千香 (2026)/Season 1/千香 - S01E13 - 第 13 集.mkv", resource.Path)
	assertEqual(t, "mkv", resource.Container)
	assertEqual(t, int32(3840), resource.VideoWidth)
	assertEqual(t, int32(2160), resource.VideoHeight)
	assertEqual(t, int64(27380040000), resource.RuntimeTicks)
	assertEqual(t, int64(2066995), resource.Bitrate)
	assertEqual(t, "hevc", resource.VideoCodec)
	assertEqual(t, 23.976, resource.FPS)
	assertEqual(t, "test-node", resource.OwnerName)

	payload, err := json.Marshal(resource)
	if err != nil {
		t.Fatal(err)
	}
	if containsJSONField(payload, "pre_sha1") {
		t.Fatalf("pre_sha1 should not be sent: %s", payload)
	}
	if containsJSONField(payload, "raw_name") {
		t.Fatalf("raw_name should not be sent: %s", payload)
	}
}

func TestHandleMediaEventSkipsMissingResource(t *testing.T) {
	event := pb.MediaEvent{
		EventId: "event-1",
		Event:   "media.item.added",
	}

	if err := handleMediaEvent(&event); err != nil {
		t.Fatalf("expected missing media resource to be skipped, got %v", err)
	}
}

func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func containsJSONField(payload []byte, field string) bool {
	var value map[string]any
	if err := json.Unmarshal(payload, &value); err != nil {
		return false
	}
	_, ok := value[field]
	return ok
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		path     string
		prefix   string
		expected bool
	}{
		{"/Movies/Inception.strm", "/Movies", true},
		{"/Movies/Inception.strm", "/Movies/", true},
		{"/Movies", "/Movies", true},
		{"/Movies", "/Movies/", true},
		{"/Movies-Backup/Inception.strm", "/Movies", false},
		{"/TV/Show.mkv", "/Movies", false},
		{"/Movies/SciFi/Interstellar.strm", "/Movies/SciFi", true},
		{"/Movies/SciFi/Interstellar.strm", "/Movies/SciFi/", true},
		{"/Movies/Inception.strm", "Movies", true},
		{"Movies/Inception.strm", "/Movies", true},
		{"Movies", "Movies", true},
		{"/Movies", "Movies", true},
	}

	for _, tt := range tests {
		actual := matchPath(tt.path, tt.prefix)
		if actual != tt.expected {
			t.Errorf("matchPath(%q, %q) = %v; expected %v", tt.path, tt.prefix, actual, tt.expected)
		}
	}
}
