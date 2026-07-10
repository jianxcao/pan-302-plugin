package main

import (
	"encoding/json"
	"testing"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

type recordingCloudHubClient struct {
	deletedSHA1s []string
}

func (c *recordingCloudHubClient) PushInBatches([]Resource, int) (*PushResponse, error) {
	return &PushResponse{State: true}, nil
}

func (c *recordingCloudHubClient) DeleteOwners(sha1s []string) (*DeleteOwnersResponse, error) {
	c.deletedSHA1s = append([]string(nil), sha1s...)
	return &DeleteOwnersResponse{State: true, DeletedOwners: int64(len(sha1s))}, nil
}

func TestShouldHandleResourceReady(t *testing.T) {
	assertEqual(t, true, shouldHandleResourceReady("strm", false))
	assertEqual(t, false, shouldHandleResourceReady("media", false))
	assertEqual(t, false, shouldHandleResourceReady("strm", true))
	assertEqual(t, true, shouldHandleResourceReady("media", true))
}

func TestResourceFromReadyEvent(t *testing.T) {
	event := &pb.ResourceReadyEvent{
		EventId: "strm-1",
		Event:   "strm.created",
		Strm:    &pb.StrmEventInfo{CloudPath: "/tv/千香 (2026)/千香.S01E13.2160p.mkv"},
		File: &pb.FileSnapshot{
			Id:       "file-1",
			Name:     "千香.S01E13.2160p.mkv",
			Path:     "/tv/千香 (2026)/千香.S01E13.2160p.mkv",
			Size:     707430281,
			Hashes:   map[string]string{"sha1": "abc"},
			PickCode: "pick-1",
		},
		Media: &pb.MediaItemInfo{
			SeriesName: "千香", ProductionYear: 2026, ParentIndexNumber: 1, IndexNumber: 13,
			Width: 3840, Height: 2160, Container: "mkv", VideoCodec: "hevc", Fps: 23.976,
		},
	}

	resource := resourceFromReadyEvent(event, PluginConfig{NodeID: "test-node"})
	assertEqual(t, "abc", resource.SHA1)
	assertEqual(t, "file-1", resource.FileId)
	assertEqual(t, "707430281", resource.Size)
	assertEqual(t, "千香.S01E13.2160p.mkv", resource.Name)
	assertEqual(t, "/tv/千香 (2026)/千香.S01E13.2160p.mkv", resource.Path)
	assertEqual(t, "pick-1", resource.PickCode)
	assertEqual(t, "tv", resource.Type)
	assertEqual(t, 1, resource.Season)
	assertEqual(t, 13, resource.Episode)
	assertEqual(t, "2160p", resource.Quality)
	assertEqual(t, "2026", resource.Year)
	assertEqual(t, "test-node", resource.OwnerName)
	assertEqual(t, "千香", resource.Title)
	assertEqual(t, int32(3840), resource.VideoWidth)
	assertEqual(t, "hevc", resource.VideoCodec)
	payload, err := json.Marshal(resource)
	if err != nil {
		t.Fatal(err)
	}
	if containsJSONField(payload, "pre_sha1") || containsJSONField(payload, "raw_name") {
		t.Fatalf("unexpected legacy fields: %s", payload)
	}
}

func TestHandleStrmDeletedUsesSnapshotSHA1(t *testing.T) {
	client := &recordingCloudHubClient{}
	event := &pb.StrmEvent{
		EventId: "strm-delete-1",
		Event:   "strm.deleted",
		Strm:    &pb.StrmEventInfo{CloudPath: "/movies/movie.mkv"},
		File:    &pb.FileSnapshot{Hashes: map[string]string{"sha1": "sha1-before-delete"}},
	}

	if err := handleStrmEventWithConfig(event, PluginConfig{}, client); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, 1, len(client.deletedSHA1s))
	assertEqual(t, "sha1-before-delete", client.deletedSHA1s[0])
}

func TestHandleStrmDeletedWithoutSHA1SkipsDelete(t *testing.T) {
	client := &recordingCloudHubClient{}
	event := &pb.StrmEvent{
		EventId: "strm-delete-dir",
		Event:   "strm.deleted",
		Strm:    &pb.StrmEventInfo{CloudPath: "/movies/season"},
		File:    &pb.FileSnapshot{Path: "/movies/season"},
	}

	if err := handleStrmEventWithConfig(event, PluginConfig{}, client); err != nil {
		t.Fatal(err)
	}
	assertEqual(t, 0, len(client.deletedSHA1s))
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
