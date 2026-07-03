package main

import (
	"encoding/json"
	"fmt"
	"strings"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	pan302plugin "github.com/jianxcao/pan-302-plugin/sdk/go"
)

type Client struct {
	BaseURL       string
	NodeID        string
	APIKey        string
	PublicBaseURL string
}

type Resource struct {
	SHA1            string  `json:"sha1"`
	FileId          string  `json:"file_id"`
	Size            string  `json:"size,omitempty"`
	Name            string  `json:"name,omitempty"`
	Path            string  `json:"path,omitempty"`
	PickCode        string  `json:"pick_code,omitempty"`
	Title           string  `json:"title,omitempty"`
	TMDBID          string  `json:"tmdb_id,omitempty"`
	Type            string  `json:"type,omitempty"`
	Season          int     `json:"season,omitempty"`
	Episode         int     `json:"episode,omitempty"`
	StartSeason     int     `json:"start_season,omitempty"`
	StartEpisode    int     `json:"start_episode,omitempty"`
	IsFullSeason    bool    `json:"is_full_season,omitempty"`
	Quality         string  `json:"quality,omitempty"`
	Container       string  `json:"container,omitempty"`
	VideoCodec      string  `json:"video_codec,omitempty"`
	VideoResolution string  `json:"video_resolution,omitempty"`
	VideoWidth      int32   `json:"video_width,omitempty"`
	VideoHeight     int32   `json:"video_height,omitempty"`
	VideoHDR        string  `json:"video_hdr,omitempty"`
	VideoBitDepth   int32   `json:"video_bit_depth,omitempty"`
	FPS             float64 `json:"fps,omitempty"`
	AudioCodec      string  `json:"audio_codec,omitempty"`
	AudioChannels   int32   `json:"audio_channels,omitempty"`
	AudioLanguage   string  `json:"audio_language,omitempty"`
	RuntimeTicks    int64   `json:"runtime_ticks,omitempty"`
	Bitrate         int64   `json:"bitrate,omitempty"`
	Year            string  `json:"year,omitempty"`
	Category        string  `json:"category,omitempty"`
	Actors          string  `json:"actors,omitempty"`
	//  "schema": "cloud_resource.v1",
	Schema    string `json:"Schema,omitempty"`
	OwnerName string `json:"ownerName,omitempty"`
}

type PushResponse struct {
	State               bool  `json:"state"`
	Inserted            int64 `json:"inserted"`
	Updated             int64 `json:"updated"`
	Unchanged           int64 `json:"unchanged"`
	OwnerAdded          int64 `json:"owner_added"`
	OwnerChanged        int64 `json:"owner_changed"`
	EventCandidates     int64 `json:"event_candidates"`
	EventsRecorded      int64 `json:"events_recorded"`
	BroadcastSuppressed bool  `json:"broadcast_suppressed"`
}

type DeleteOwnersResponse struct {
	State            bool     `json:"state"`
	DeletedOwners    int64    `json:"deleted_owners"`
	DeletedResources int64    `json:"deleted_resources"`
	SHA1s            []string `json:"sha1s"`
}

func NewClient(baseURL, nodeID, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		NodeID:  nodeID,
		APIKey:  apiKey,
	}
}

func (c *Client) Push(resources []Resource) (*PushResponse, error) {
	items := make([]Resource, 0, len(resources))
	for _, item := range resources {
		item.SHA1 = strings.ToUpper(strings.TrimSpace(item.SHA1))
		item.Path = normalizeCloudHubPath(item.Path)
		if item.SHA1 == "" {
			continue
		}
		if item.Type == "" {
			item.Type = "movie"
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return &PushResponse{State: true}, nil
	}
	var result PushResponse
	if err := c.postJSON("/v1/push", items, &result); err != nil {
		return nil, err
	}
	if !result.State {
		return nil, fmt.Errorf("CloudHub push 返回 state=false")
	}
	return &result, nil
}

func (c *Client) PushInBatches(resources []Resource, batchSize int) (*PushResponse, error) {
	if batchSize <= 0 {
		batchSize = 500
	}
	total := &PushResponse{State: true}
	for start := 0; start < len(resources); start += batchSize {
		end := min(start+batchSize, len(resources))
		result, err := c.Push(resources[start:end])
		if err != nil {
			return nil, err
		}
		total.Inserted += result.Inserted
		total.Updated += result.Updated
		total.Unchanged += result.Unchanged
		total.OwnerAdded += result.OwnerAdded
		total.OwnerChanged += result.OwnerChanged
		total.EventCandidates += result.EventCandidates
		total.EventsRecorded += result.EventsRecorded
		total.BroadcastSuppressed = total.BroadcastSuppressed || result.BroadcastSuppressed
	}
	return total, nil
}

func (c *Client) DeleteOwners(sha1s []string) (*DeleteOwnersResponse, error) {
	seen := map[string]struct{}{}
	cleaned := make([]string, 0, len(sha1s))
	for _, sha1 := range sha1s {
		sha1 = strings.ToUpper(strings.TrimSpace(sha1))
		if sha1 == "" {
			continue
		}
		if _, exists := seen[sha1]; exists {
			continue
		}
		seen[sha1] = struct{}{}
		cleaned = append(cleaned, sha1)
	}
	if len(cleaned) == 0 {
		return &DeleteOwnersResponse{State: true}, nil
	}
	var result DeleteOwnersResponse
	if err := c.postJSON("/v1/owners/delete", map[string]any{"sha1s": cleaned}, &result); err != nil {
		return nil, err
	}
	if !result.State {
		return nil, fmt.Errorf("CloudHub delete 返回 state=false")
	}
	return &result, nil
}

func (c *Client) postJSON(path string, body any, target any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	headers := map[string]*pb.StringList{
		"content-type": {Values: []string{"application/json"}},
		"node-id":      {Values: []string{c.NodeID}},
		"x-api-key":    {Values: []string{c.APIKey}},
	}
	if c.PublicBaseURL != "" {
		headers["node-public-url"] = &pb.StringList{Values: []string{strings.TrimRight(c.PublicBaseURL, "/")}}
	}
	response, err := pan302plugin.HTTP.Request(&pb.HTTPRequestArgs{
		Method:        "POST",
		Url:           c.BaseURL + path,
		Headers:       headers,
		Body:          encoded,
		TimeoutMillis: 20000,
	})
	if err != nil {
		return err
	}
	if response.Status < 200 || response.Status >= 300 {
		return fmt.Errorf("CloudHub %s 请求失败: HTTP %d: %s", path, response.Status, string(response.Body))
	}
	if len(response.Body) == 0 {
		return nil
	}
	return json.Unmarshal(response.Body, target)
}

func normalizeCloudHubPath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	for strings.Contains(value, "//") {
		value = strings.ReplaceAll(value, "//", "/")
	}
	return strings.Trim(value, "/")
}
