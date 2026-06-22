package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	pan302plugin "github.com/jianxcao/pan-302-plugin/sdk/go"
)

type mediaItemsResponse struct {
	Items []mediaItem `json:"Items"`
}

type mediaItem struct {
	Container    string        `json:"Container"`
	Bitrate      int64         `json:"Bitrate"`
	RunTimeTicks int64         `json:"RunTimeTicks"`
	MediaSources []mediaSource `json:"MediaSources"`
}

type mediaSource struct {
	Container    string        `json:"Container"`
	Bitrate      int64         `json:"Bitrate"`
	RunTimeTicks int64         `json:"RunTimeTicks"`
	MediaStreams []mediaStream `json:"MediaStreams"`
}

type mediaStream struct {
	Type             string  `json:"Type"`
	Codec            string  `json:"Codec"`
	Language         string  `json:"Language"`
	Width            int32   `json:"Width"`
	Height           int32   `json:"Height"`
	BitDepth         int32   `json:"BitDepth"`
	BitRate          int64   `json:"BitRate"`
	Channels         int32   `json:"Channels"`
	IsDefault        bool    `json:"IsDefault"`
	VideoRange       string  `json:"VideoRange"`
	AverageFrameRate float64 `json:"AverageFrameRate"`
	RealFrameRate    float64 `json:"RealFrameRate"`
}

func enrichResourceWithMedia(event *pb.StrmEvent, resource *Resource) {
	if event == nil || event.Strm == nil || resource == nil {
		return
	}
	localPath := strings.TrimSpace(event.Strm.LocalPath)
	if localPath == "" {
		return
	}
	resp, err := readMediaItemsByPath(localPath)
	if err != nil {
		pan302plugin.Logger.Warn("查询媒体服务器信息失败，继续推送基础资源", map[string]string{
			"eventId": event.EventId,
			"path":    localPath,
			"error":   err.Error(),
		})
		return
	}
	if resp == nil || len(resp.Items) == 0 {
		pan302plugin.Logger.Info("媒体服务器未返回媒体流信息，继续推送基础资源", map[string]string{
			"eventId": event.EventId,
			"path":    localPath,
		})
		return
	}
	applyMediaInfo(resource, resp)
}

func readMediaItemsByPath(localPath string) (*mediaItemsResponse, error) {
	cfg, err := pan302plugin.Media.ServerConfig()
	if err != nil {
		return nil, err
	}
	if cfg == nil || strings.TrimSpace(cfg.Url) == "" {
		pan302plugin.Logger.Info("未配置媒体服务器，跳过媒体流补充", nil)
		return nil, nil
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.Url), "/")
	u, err := url.Parse(base + "/Items")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("Recursive", "true")
	q.Set("Fields", "MediaSources")
	q.Set("Path", localPath)
	q.Set("Limit", "1")
	u.RawQuery = q.Encode()

	headers := map[string]*pb.StringList{}
	if cfg.Token != "" {
		headers["X-Emby-Token"] = &pb.StringList{Values: []string{cfg.Token}}
	}
	response, err := pan302plugin.HTTP.Request(&pb.HTTPRequestArgs{
		Method:        http.MethodGet,
		Url:           u.String(),
		Headers:       headers,
		TimeoutMillis: 20000,
	})
	if err != nil {
		return nil, err
	}
	if response.Status < 200 || response.Status >= 300 {
		return nil, fmt.Errorf("media server status %d", response.Status)
	}
	var out mediaItemsResponse
	if err := json.Unmarshal(response.Body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func applyMediaInfo(resource *Resource, resp *mediaItemsResponse) {
	if resource == nil || resp == nil || len(resp.Items) == 0 {
		return
	}
	item := resp.Items[0]
	if resource.Container == "" {
		resource.Container = item.Container
	}
	if resource.Bitrate == 0 {
		resource.Bitrate = item.Bitrate
	}
	if resource.RuntimeTicks == 0 {
		resource.RuntimeTicks = item.RunTimeTicks
	}
	var source *mediaSource
	if len(item.MediaSources) > 0 {
		source = &item.MediaSources[0]
	}
	if source != nil {
		if resource.Container == "" {
			resource.Container = source.Container
		}
		if resource.Bitrate == 0 {
			resource.Bitrate = source.Bitrate
		}
		if resource.RuntimeTicks == 0 {
			resource.RuntimeTicks = source.RunTimeTicks
		}
		video := firstMediaStream(source.MediaStreams, "Video")
		if video != nil {
			applyVideoInfo(resource, video)
		}
		audio := defaultOrFirstMediaStream(source.MediaStreams, "Audio")
		if audio != nil {
			applyAudioInfo(resource, audio)
		}
	}
	if resource.Quality == "" {
		resource.Quality = qualityFromMedia(resource.VideoWidth, resource.VideoHeight, resource.VideoHDR)
	}
}

func applyVideoInfo(resource *Resource, stream *mediaStream) {
	if resource.VideoCodec == "" {
		resource.VideoCodec = strings.ToLower(strings.TrimSpace(stream.Codec))
	}
	if resource.VideoWidth == 0 {
		resource.VideoWidth = stream.Width
	}
	if resource.VideoHeight == 0 {
		resource.VideoHeight = stream.Height
	}
	if resource.VideoResolution == "" {
		resource.VideoResolution = resolutionFromDimensions(stream.Width, stream.Height)
	}
	if resource.VideoHDR == "" {
		resource.VideoHDR = normalizeHDR(stream.VideoRange)
	}
	if resource.VideoBitDepth == 0 {
		resource.VideoBitDepth = stream.BitDepth
	}
	if resource.FPS == 0 {
		resource.FPS = firstPositiveFloat(stream.AverageFrameRate, stream.RealFrameRate)
	}
	if resource.Bitrate == 0 {
		resource.Bitrate = stream.BitRate
	}
}

func applyAudioInfo(resource *Resource, stream *mediaStream) {
	if resource.AudioCodec == "" {
		resource.AudioCodec = strings.ToLower(strings.TrimSpace(stream.Codec))
	}
	if resource.AudioChannels == 0 {
		resource.AudioChannels = stream.Channels
	}
	if resource.AudioLanguage == "" {
		resource.AudioLanguage = strings.TrimSpace(stream.Language)
	}
}

func firstMediaStream(streams []mediaStream, typ string) *mediaStream {
	for i := range streams {
		if strings.EqualFold(streams[i].Type, typ) {
			return &streams[i]
		}
	}
	return nil
}

func defaultOrFirstMediaStream(streams []mediaStream, typ string) *mediaStream {
	var first *mediaStream
	for i := range streams {
		stream := &streams[i]
		if !strings.EqualFold(stream.Type, typ) {
			continue
		}
		if first == nil {
			first = stream
		}
		if stream.IsDefault {
			return stream
		}
	}
	return first
}

func resolutionFromDimensions(width, height int32) string {
	switch {
	case width >= 3800 || height >= 2160:
		return "2160p"
	case width >= 2500 || height >= 1440:
		return "1440p"
	case width >= 1900 || height >= 1080:
		return "1080p"
	case width >= 1200 || height >= 720:
		return "720p"
	default:
		return ""
	}
}

func normalizeHDR(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch {
	case strings.Contains(lower, "dolby") || strings.Contains(lower, "dv"):
		return "DV"
	case strings.Contains(lower, "hdr10+"):
		return "HDR10+"
	case strings.Contains(lower, "hdr"):
		return "HDR"
	default:
		return ""
	}
}

func qualityFromMedia(width, height int32, hdr string) string {
	resolution := resolutionFromDimensions(width, height)
	if resolution == "" && hdr == "" {
		return ""
	}
	if resolution == "" {
		return hdr
	}
	if hdr == "" {
		return resolution
	}
	return fmt.Sprintf("%s %s", resolution, hdr)
}

func firstPositiveFloat(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
