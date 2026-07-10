package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	sdk "github.com/jianxcao/pan-302-plugin/sdk/go"
)

type PluginConfig struct {
	APIKey             string   `json:"api_key"`
	BaseURL            string   `json:"base_url"`
	NodeID             string   `json:"node_id"`
	PublicBaseURL      string   `json:"public_base_url,omitempty"`
	BatchSize          int      `json:"batch_size,omitempty"`
	IncludePaths       []string `json:"include_paths,omitempty"`
	UseMediaAddedEvent bool     `json:"use_media_added_event,omitempty"`
}

type cloudHubClient interface {
	PushInBatches(resources []Resource, batchSize int) (*PushResponse, error)
	DeleteOwners(sha1s []string) (*DeleteOwnersResponse, error)
}

var (
	seasonEpisodePattern = regexp.MustCompile(`(?i)(?:^|[\s._-])S?(\d{1,3})E(\d{1,5})(?:$|[\s._-])`)
	yearPattern          = regexp.MustCompile(`(?:^|[^0-9])(19\d{2}|20\d{2})(?:$|[^0-9])`)
	eventRouter          = sdk.NewEventRouter().
				OnStrm(handleStrmEvent).
				OnResourceReady(handleResourceReadyEvent)
)

func main() {}

//go:wasmexport pan302_alloc
func pan302Alloc(size uint32) uint32 {
	return sdk.Allocate(size)
}

//go:wasmexport pan302_free
func pan302Free(ptr, _ uint32) {
	sdk.Free(ptr)
}

//go:wasmexport pan302_init
func pan302Init(ptr, length uint32) uint64 {
	var request pb.InitRequest
	if err := sdk.DecodeRequest(ptr, length, &request); err != nil {
		return errorResponse(err)
	}
	sdk.Logger.Info("CloudHub 推送插件已启动", nil)
	return successResponse()
}

//go:wasmexport pan302_on_event
func pan302OnEvent(ptr, length uint32) uint64 {
	eventID, err := eventRouter.Dispatch(ptr, length)
	if err != nil {
		sdk.Logger.Error("CloudHub 事件处理失败", map[string]string{"eventId": eventID, "error": err.Error()})
		return errorResponse(err)
	}
	sdk.Logger.Debug("CloudHub 事件处理完成", map[string]string{"eventId": eventID})
	return successResponse()
}

func handleStrmEvent(event *pb.StrmEvent) error {
	sdk.Logger.Debug("CloudHub 收到 STRM 事件", map[string]string{"eventId": event.EventId, "event": event.Event})
	sdk.Logger.Debug("CloudHub 收到 STRM 事件 FILE", event.File)
	config, ok, err := readPluginConfig()
	if err != nil {
		return err
	}
	if !ok {
		sdk.Logger.Debug("CloudHub 配置未就绪，跳过 STRM 事件", map[string]string{"eventId": event.EventId, "event": event.Event})
		return nil
	}
	if event.Event != "strm.deleted" {
		sdk.Logger.Debug("CloudHub 跳过非删除 STRM 事件", map[string]string{"eventId": event.EventId, "event": event.Event})
		return nil
	}
	client := NewClient(config.BaseURL, config.NodeID, config.APIKey)
	client.PublicBaseURL = config.PublicBaseURL
	return handleStrmEventWithConfig(event, config, client)
}

func handleResourceReadyEvent(event *pb.ResourceReadyEvent) error {
	sdk.Logger.Debug("CloudHub 收到 Resource Ready 事件", event)
	config, ok, err := readPluginConfig()
	if err != nil || !ok {
		return err
	}
	if !shouldHandleResourceReady(event.Source, config.UseMediaAddedEvent) {
		sdk.Logger.Debug("CloudHub 按插件来源开关跳过 Resource Ready 事件", map[string]string{
			"eventId": event.EventId, "source": event.Source,
			"useMediaAddedEvent": strconv.FormatBool(config.UseMediaAddedEvent),
		})
		return nil
	}
	if event.File == nil {
		sdk.Logger.Warn("Resource Ready 事件缺少文件快照，已跳过", map[string]string{"eventId": event.EventId, "source": event.Source})
		return nil
	}
	cloudPath := event.File.Path
	if event.Strm != nil && event.Strm.CloudPath != "" {
		cloudPath = event.Strm.CloudPath
	}
	if !includePathMatched(config, cloudPath, event.EventId) {
		return nil
	}
	client := NewClient(config.BaseURL, config.NodeID, config.APIKey)
	client.PublicBaseURL = config.PublicBaseURL
	resource := resourceFromReadyEvent(event, config)
	if resource.SHA1 == "" {
		sdk.Logger.Warn("资源就绪事件缺少 SHA1，已跳过", map[string]string{"eventId": event.EventId})
		return nil
	}
	return pushResource(client, resource, config, event.EventId)
}

func shouldHandleResourceReady(source string, useMediaAddedEvent bool) bool {
	// 该开关只决定 CloudHub 插件消费哪一种 Resource Ready 来源。
	// 主程序仍会发布全部 STRM、Media 和 Resource Ready 事件，供其他插件独立订阅。
	if useMediaAddedEvent {
		return source == "media"
	}
	return source == "strm"
}

func handleStrmEventWithConfig(event *pb.StrmEvent, config PluginConfig, client cloudHubClient) error {
	if event.File == nil {
		sdk.Logger.Warn("STRM 事件缺少文件快照，已跳过", map[string]string{"eventId": event.EventId})
		return nil
	}
	cloudPath := event.File.Path
	if event.Strm != nil && event.Strm.CloudPath != "" {
		cloudPath = event.Strm.CloudPath
	}
	if !includePathMatched(config, cloudPath, event.EventId) {
		return nil
	}
	if event.Event == "strm.deleted" {
		sha1 := event.File.Hashes["sha1"]
		if sha1 == "" {
			sdk.Logger.Debug("STRM 删除事件缺少 SHA1，已跳过", map[string]string{"eventId": event.EventId, "path": cloudPath})
			return nil
		}
		result, err := client.DeleteOwners([]string{sha1})
		if err != nil {
			return err
		}
		sdk.Logger.Info("CloudHub STRM 删除推送成功", map[string]string{
			"eventId":      event.EventId,
			"deletedOwner": strconv.FormatInt(result.DeletedOwners, 10),
		})
		return nil
	}
	return nil
}

func readPluginConfig() (PluginConfig, bool, error) {
	configResp, err := sdk.Config.Read()
	if err != nil {
		return PluginConfig{}, false, fmt.Errorf("读取插件配置: %w", err)
	}
	return parsePluginConfig(configResp.Config)
}

func parsePluginConfig(value interface{ MarshalJSON() ([]byte, error) }) (PluginConfig, bool, error) {
	var config PluginConfig
	if value == nil {
		return config, false, fmt.Errorf("插件配置为空")
	}
	configJSON, err := value.MarshalJSON()
	if err != nil {
		return config, false, fmt.Errorf("marshal config: %w", err)
	}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return config, false, fmt.Errorf("unmarshal config: %w", err)
	}
	config.BaseURL = strings.TrimSpace(config.BaseURL)
	config.NodeID = strings.TrimSpace(config.NodeID)
	config.APIKey = strings.TrimSpace(config.APIKey)
	if config.BaseURL == "" || config.NodeID == "" || config.APIKey == "" {
		sdk.Logger.Warn("请先配置 CloudHub API URL、Node ID 和 API Key", nil)
		return config, false, nil
	}
	return config, true, nil
}

func includePathMatched(config PluginConfig, cloudPath, eventID string) bool {
	if len(config.IncludePaths) == 0 {
		return true
	}
	for _, prefix := range config.IncludePaths {
		if matchPath(cloudPath, prefix) {
			return true
		}
	}
	sdk.Logger.Info("路径未在包含列表中，跳过通知", map[string]string{
		"path":    cloudPath,
		"eventId": eventID,
	})
	return false
}

func pushResource(client cloudHubClient, resource Resource, config PluginConfig, eventID string) error {
	sdk.Logger.Info("发送 cloudhub 名称", resource)
	batchSize := config.BatchSize
	if batchSize <= 0 || batchSize > 500 {
		batchSize = 500
	}
	result, err := client.PushInBatches([]Resource{resource}, batchSize)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		isConflict := strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "unique")
		if isConflict && resource.SHA1 != "" {
			sdk.Logger.Warn("推送遇到唯一性冲突，尝试先删除已有归属再重新推送", map[string]string{
				"eventId": eventID,
				"sha1":    resource.SHA1,
				"error":   err.Error(),
			})
			if _, delErr := client.DeleteOwners([]string{resource.SHA1}); delErr != nil {
				sdk.Logger.Warn("重试删除归属失败", map[string]string{
					"eventId": eventID,
					"sha1":    resource.SHA1,
					"error":   delErr.Error(),
				})
			}
			result, err = client.PushInBatches([]Resource{resource}, batchSize)
		}
		if err != nil {
			sdk.Logger.Error("调用 cloudhub 失败", map[string]string{
				"error": err.Error(),
			})
			return nil
		}
	}
	sdk.Logger.Info("CloudHub 资源推送成功", map[string]string{
		"eventId":  eventID,
		"inserted": strconv.FormatInt(result.Inserted, 10),
		"updated":  strconv.FormatInt(result.Updated, 10),
	})
	return nil
}

func resourceFromReadyEvent(event *pb.ResourceReadyEvent, cfg PluginConfig) Resource {
	file := event.File
	cloudPath := file.Path
	if event.Strm != nil && event.Strm.CloudPath != "" {
		cloudPath = event.Strm.CloudPath
	}
	season, episode := parseSeasonEpisode(file.Name)
	mediaType := "movie"
	if season > 0 || episode > 0 {
		mediaType = "tv"
	}
	resource := Resource{
		SHA1:      file.Hashes["sha1"],
		FileId:    file.Id,
		Size:      strconv.FormatInt(file.Size, 10),
		Name:      file.Name,
		Path:      cloudPath,
		PickCode:  file.PickCode,
		Type:      mediaType,
		Season:    season,
		Episode:   episode,
		Quality:   parseQuality(file.Name),
		Year:      parseYear(cloudPath, file.Name),
		Schema:    "cloud_resource.v1",
		OwnerName: cfg.NodeID,
	}
	applyMediaInfo(&resource, event.Media)
	return resource
}

func applyMediaInfo(resource *Resource, item *pb.MediaItemInfo) {
	if resource == nil || item == nil {
		return
	}
	resource.Title = mediaTitle(item)
	resource.Container = item.Container
	resource.VideoWidth = item.Width
	resource.VideoHeight = item.Height
	resource.RuntimeTicks = item.RuntimeTicks
	resource.Bitrate = item.Bitrate
	resource.VideoCodec = item.VideoCodec
	resource.FPS = item.Fps
	resource.VideoHDR = item.VideoRange
	if item.ProductionYear > 0 {
		resource.Year = strconv.Itoa(int(item.ProductionYear))
	}
	if resource.Season == 0 && resource.Episode == 0 {
		resource.Season = int(item.ParentIndexNumber)
		resource.Episode = int(item.IndexNumber)
		if resource.Season > 0 || resource.Episode > 0 {
			resource.Type = "tv"
		}
	}
	if resource.Quality == "" && (item.Width > 0 || item.Height > 0) {
		resource.Quality = resolutionFromDimensions(item.Width, item.Height)
	}
	if resource.Quality != "" && item.VideoRange != "" {
		lowerQ := strings.ToLower(resource.Quality)
		if !containsAny(lowerQ, "hdr", "sdr", "dv", "dovi") {
			resource.Quality = resource.Quality + " " + item.VideoRange
		}
	}
}

func mediaTitle(item *pb.MediaItemInfo) string {
	if item == nil {
		return ""
	}
	if strings.TrimSpace(item.SeriesName) != "" {
		return strings.TrimSpace(item.SeriesName)
	}
	return strings.TrimSpace(item.Name)
}

func resolutionFromDimensions(width, height int32) string {
	switch {
	case width >= 3800 || height >= 2160:
		return "2160p"
	case width >= 2560 || height >= 1440:
		return "1440p"
	case width >= 1920 || height >= 1080:
		return "1080p"
	case width >= 1280 || height >= 720:
		return "720p"
	default:
		return ""
	}
}

func parseSeasonEpisode(name string) (int, int) {
	matches := seasonEpisodePattern.FindStringSubmatch(name)
	if len(matches) != 3 {
		return 0, 0
	}
	season, _ := strconv.Atoi(matches[1])
	episode, _ := strconv.Atoi(matches[2])
	return season, episode
}

func parseYear(values ...string) string {
	for _, value := range values {
		matches := yearPattern.FindStringSubmatch(value)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}

func parseQuality(name string) string {
	lower := strings.ToLower(name)
	var values []string
	switch {
	case containsAny(lower, "2160p", "4k", "uhd", "3840x2160"):
		values = append(values, "2160p")
	case containsAny(lower, "1440p", "2k", "2560x1440"):
		values = append(values, "1440p")
	case containsAny(lower, "1080p", "fhd", "1920x1080"):
		values = append(values, "1080p")
	case containsAny(lower, "720p", "1280x720"):
		values = append(values, "720p")
	}
	if containsAny(lower, "hdr10+", "hdr10", "hdr", "10bit") {
		values = append(values, "HDR")
	}
	if containsAny(lower, "dovi", "dolby vision", ".dv.", " dv ") {
		values = append(values, "DV")
	}
	if strings.Contains(lower, "remux") {
		values = append(values, "REMUX")
	}
	if containsAny(lower, "web-dl", "webdl", "webrip") {
		values = append(values, "WEB-DL")
	}
	if containsAny(lower, "bluray", "bdrip", "bdr") {
		values = append(values, "BluRay")
	}
	return strings.Join(values, " ")
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func matchPath(path string, prefix string) bool {
	p := "/" + strings.Trim(path, "/")
	pf := "/" + strings.Trim(prefix, "/")
	p = strings.TrimSuffix(p, "/")
	pf = strings.TrimSuffix(pf, "/")
	if p == pf {
		return true
	}
	return strings.HasPrefix(p, pf+"/")
}

func successResponse() uint64 {
	return sdk.EncodeResponse(&pb.LifecycleResponse{Ok: true})
}

func errorResponse(err error) uint64 {
	return sdk.EncodeResponse(&pb.LifecycleResponse{Ok: false, Error: err.Error()})
}
