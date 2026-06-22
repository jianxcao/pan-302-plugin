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

func TestResourceFromEventDoesNotUseOldCloudPathForQuality(t *testing.T) {
	event := pb.StrmEvent{
		EventId: "event-1",
		Event:   "strm.renamed",
		Strm: &pb.StrmEventInfo{
			CloudPath:    "/TV/炽夏 (2026)/Season 1/炽夏 - S01E01 - 第 1 集.mp4",
			OldCloudPath: "/TV/炽夏 (2026)/Season 1/Never.Ending.Summer.2026.S01E01.2160p.WEB-DL.AAC.H.265.mp4",
		},
		File: &pb.FileSnapshot{
			Hashes: map[string]string{"sha1": "abc"},
			Size:   1024,
			Name:   "炽夏 - S01E01 - 第 1 集.mp4",
			Path:   "/TV/炽夏 (2026)/Season 1/炽夏 - S01E01 - 第 1 集.mp4",
		},
	}
	resource := resourceFromEvent(&event, PluginConfig{NodeID: "test-node"})
	assertEqual(t, "", resource.Quality)
	applyCloudHubRecognizableName(&resource)
	assertEqual(t, "炽夏 S01E01.mp4", resource.Name)
}

func TestNormalizeCloudHubPath(t *testing.T) {
	assertEqual(t, "TV/Movie.mkv", normalizeCloudHubPath(`//TV\\Movie.mkv`))
}

func TestHandleEventSkipsMissingFileSnapshot(t *testing.T) {
	event := pb.StrmEvent{
		EventId: "event-1",
		Event:   "strm.created",
	}

	if err := handleEvent(&event); err != nil {
		t.Fatalf("expected missing file snapshot to be skipped, got %v", err)
	}
}

func TestApplyMediaInfo(t *testing.T) {
	resource := Resource{}
	applyMediaInfo(&resource, &mediaItemsResponse{
		Items: []mediaItem{{
			Container:    "mkv",
			Bitrate:      5802806,
			RunTimeTicks: 27540490000,
			MediaSources: []mediaSource{{
				Container:    "mkv",
				Bitrate:      5802806,
				RunTimeTicks: 27540490000,
				MediaStreams: []mediaStream{{
					Type:       "Video",
					Codec:      "hevc",
					Width:      3840,
					Height:     1608,
					BitDepth:   10,
					BitRate:    5802806,
					VideoRange: "HDR 10",
				}, {
					Type:      "Audio",
					Codec:     "aac",
					Channels:  2,
					Language:  "eng",
					IsDefault: false,
				}, {
					Type:      "Audio",
					Codec:     "eac3",
					Channels:  6,
					Language:  "chi",
					IsDefault: true,
				}},
			}},
		}},
	})
	assertEqual(t, "mkv", resource.Container)
	assertEqual(t, "hevc", resource.VideoCodec)
	assertEqual(t, "2160p", resource.VideoResolution)
	assertEqual(t, int32(3840), resource.VideoWidth)
	assertEqual(t, int32(1608), resource.VideoHeight)
	assertEqual(t, "HDR", resource.VideoHDR)
	assertEqual(t, int32(10), resource.VideoBitDepth)
	assertEqual(t, "eac3", resource.AudioCodec)
	assertEqual(t, int32(6), resource.AudioChannels)
	assertEqual(t, "chi", resource.AudioLanguage)
	assertEqual(t, int64(27540490000), resource.RuntimeTicks)
	assertEqual(t, int64(5802806), resource.Bitrate)
	assertEqual(t, "2160p HDR", resource.Quality)
}

func TestApplyMediaInfoKeepsResourceWhenEmpty(t *testing.T) {
	resource := Resource{Name: "Movie.mkv", Quality: "1080p"}
	applyMediaInfo(&resource, nil)
	applyMediaInfo(&resource, &mediaItemsResponse{})
	assertEqual(t, "Movie.mkv", resource.Name)
	assertEqual(t, "1080p", resource.Quality)
}

func TestApplyCloudHubRecognizableNameMovie(t *testing.T) {
	resource := Resource{
		Name:            "Inception.2010.mkv",
		Path:            "/Movies/Inception (2010)/Inception.2010.mkv",
		Title:           "Inception.2010",
		Type:            "movie",
		Year:            "2010",
		VideoResolution: "2160p",
		FPS:             23.976,
		VideoHDR:        "HDR",
		VideoCodec:      "hevc",
		AudioCodec:      "eac3",
	}
	applyCloudHubRecognizableName(&resource)
	assertEqual(t, "Inception (2010) - 2160p 23.976fps HDR HEVC EAC3.mkv", resource.Name)
	assertEqual(t, "Inception", resource.Title)
}

func TestApplyCloudHubRecognizableNameTV(t *testing.T) {
	resource := Resource{
		Name:            "问心 - S01E04 - 第 4 集.mkv",
		Path:            "/TV/问心 (2023)/Season 2/问心 - S01E04 - 第 4 集.mkv",
		Title:           "问心 - S01E04 - 第 4 集",
		Type:            "tv",
		Season:          1,
		Episode:         4,
		VideoResolution: "2160p",
		FPS:             25,
		VideoHDR:        "HDR",
		VideoCodec:      "hevc",
		AudioCodec:      "eac3",
	}
	applyCloudHubRecognizableName(&resource)
	assertEqual(t, "问心 S01E04 - 2160p 25fps HDR HEVC.mkv", resource.Name)
	assertEqual(t, "问心", resource.Title)
}

func TestApplyCloudHubRecognizableNameTVUsesChinesePathTitle(t *testing.T) {
	resource := Resource{
		Name:    "Never.Ending.Summer.2026.S01E01.2160p.WEB-DL.AAC.H.265-Hiiveweb.mp4",
		Path:    "/test/media/tv/china/炽夏 (2026)/Season 1/炽夏 - S01E01 - 第 1 集.mp4",
		Title:   "Never.Ending.Summer.2026.S01E01.2160p.WEB-DL.AAC.H.265-Hiiveweb",
		Type:    "tv",
		Season:  1,
		Episode: 1,
		Quality: "2160p WEB-DL",
		Year:    "2026",
	}
	applyCloudHubRecognizableName(&resource)
	assertEqual(t, "炽夏", resource.Title)
	assertEqual(t, "炽夏 S01E01 - 2160p.mp4", resource.Name)
}

func assertEqual[T comparable](t *testing.T, expected, actual T) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
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
