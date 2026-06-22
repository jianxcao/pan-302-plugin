package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	pan302plugin "github.com/jianxcao/pan-302-plugin/sdk/go"
)

type PluginConfig struct {
	APIKey        string   `json:"api_key"`
	BaseURL       string   `json:"base_url"`
	NodeID        string   `json:"node_id"`
	PublicBaseURL string   `json:"public_base_url,omitempty"`
	BatchSize     int      `json:"batch_size,omitempty"`
	IncludePaths  []string `json:"include_paths,omitempty"`
}

var (
	seasonEpisodePattern = regexp.MustCompile(`(?i)(?:^|[\s._-])S?(\d{1,3})E(\d{1,5})(?:$|[\s._-])`)
	yearPattern          = regexp.MustCompile(`(?i)(?:^|[\s._-])(19\d{2}|20\d{2})(?:$|[\s._-])`)
)

func main() {}

//go:wasmexport pan302_alloc
func pan302Alloc(size uint32) uint32 {
	return pan302plugin.Allocate(size)
}

//go:wasmexport pan302_free
func pan302Free(ptr, _ uint32) {
	pan302plugin.Free(ptr)
}

//go:wasmexport pan302_init
func pan302Init(ptr, length uint32) uint64 {
	var request pb.InitRequest
	if err := pan302plugin.DecodeRequest(ptr, length, &request); err != nil {
		return errorResponse(err)
	}
	pan302plugin.Logger.Info("CloudHub 推送插件已启动", nil)
	return successResponse()
}

//go:wasmexport pan302_on_event
func pan302OnEvent(ptr, length uint32) uint64 {
	var event pb.StrmEvent
	if err := pan302plugin.DecodeRequest(ptr, length, &event); err != nil {
		return errorResponse(err)
	}
	if err := handleEvent(&event); err != nil {
		pan302plugin.Logger.Error("CloudHub 事件处理失败", map[string]string{
			"eventId": event.EventId,
			"error":   err.Error(),
		})
		return errorResponse(err)
	}
	return successResponse()
}

func handleEvent(event *pb.StrmEvent) error {
	switch event.Event {
	case "strm.created", "strm.overwritten", "strm.renamed", "strm.moved", "strm.copied", "strm.deleted":
	default:
		return nil
	}
	configResp, err := pan302plugin.Config.Read()
	if err != nil {
		return fmt.Errorf("读取插件配置: %w", err)
	}
	if configResp.Config == nil {
		return fmt.Errorf("插件配置为空")
	}
	configJSON, err := configResp.Config.MarshalJSON()
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	var config PluginConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	config.BaseURL = strings.TrimSpace(config.BaseURL)
	config.NodeID = strings.TrimSpace(config.NodeID)
	config.APIKey = strings.TrimSpace(config.APIKey)
	if config.BaseURL == "" || config.NodeID == "" || config.APIKey == "" {
		pan302plugin.Logger.Warn("请先配置 CloudHub API URL、Node ID 和 API Key", nil)
		return nil
	}
	if event.File == nil {
		return fmt.Errorf("事件 %s 不包含文件数据库快照", event.EventId)
	}

	cloudPath := ""
	if event.Strm != nil {
		cloudPath = event.Strm.CloudPath
	}
	if cloudPath == "" {
		cloudPath = event.File.Path
	}

	if len(config.IncludePaths) > 0 {
		matched := false
		for _, prefix := range config.IncludePaths {
			if matchPath(cloudPath, prefix) {
				matched = true
				break
			}
		}
		if !matched {
			pan302plugin.Logger.Info("路径未在包含列表中，跳过通知", map[string]string{
				"path":    cloudPath,
				"eventId": event.EventId,
			})
			return nil
		}
	}

	client := NewClient(config.BaseURL, config.NodeID, config.APIKey)
	client.PublicBaseURL = config.PublicBaseURL
	if event.Event == "strm.deleted" {
		sha1 := event.File.Hashes["sha1"]
		if sha1 == "" {
			pan302plugin.Logger.Warn("删除事件缺少 SHA1，已跳过", map[string]string{"eventId": event.EventId})
			return nil
		}
		result, err := client.DeleteOwners([]string{sha1})
		if err != nil {
			return err
		}
		pan302plugin.Logger.Info("CloudHub 删除推送成功", map[string]string{
			"eventId":      event.EventId,
			"deletedOwner": strconv.FormatInt(result.DeletedOwners, 10),
		})
		return nil
	}
	resource := resourceFromEvent(event, config)
	if resource.SHA1 == "" {
		pan302plugin.Logger.Warn("创建事件缺少 SHA1，已跳过", map[string]string{"eventId": event.EventId})
		return nil
	}
	enrichResourceWithMedia(event, &resource)
	applyCloudHubRecognizableName(&resource)
	pan302plugin.Logger.Info("发送 cloudhub 名称", map[string]string{
		"Name":    resource.Name,
		"Quality": resource.Quality,
	})
	batchSize := config.BatchSize
	if batchSize <= 0 || batchSize > 500 {
		batchSize = 500
	}
	result, err := client.PushInBatches([]Resource{resource}, batchSize)
	if err != nil {
		errStr := strings.ToLower(err.Error())
		isConflict := strings.Contains(errStr, "unique constraint") || strings.Contains(errStr, "unique")
		if isConflict && resource.SHA1 != "" {
			pan302plugin.Logger.Warn("推送遇到唯一性冲突，尝试先删除已有归属再重新推送", map[string]string{
				"eventId": event.EventId,
				"sha1":    resource.SHA1,
				"error":   err.Error(),
			})
			if _, delErr := client.DeleteOwners([]string{resource.SHA1}); delErr != nil {
				pan302plugin.Logger.Warn("重试删除归属失败", map[string]string{
					"eventId": event.EventId,
					"sha1":    resource.SHA1,
					"error":   delErr.Error(),
				})
			}
			result, err = client.PushInBatches([]Resource{resource}, batchSize)
		}
		if err != nil {
			pan302plugin.Logger.Error("调用 cloudhub 失败", map[string]string{
				"error": err.Error(),
			})
			return nil
		}
	}
	pan302plugin.Logger.Info("CloudHub 资源推送成功", map[string]string{
		"eventId":  event.EventId,
		"inserted": strconv.FormatInt(result.Inserted, 10),
		"updated":  strconv.FormatInt(result.Updated, 10),
	})
	return nil
}

func resourceFromEvent(event *pb.StrmEvent, cfg PluginConfig) Resource {
	file := event.File
	cloudPath := ""
	if event.Strm != nil {
		cloudPath = event.Strm.CloudPath
	}
	if cloudPath == "" {
		cloudPath = file.Path
	}
	season, episode := parseSeasonEpisode(file.Name)
	mediaType := "movie"
	if season > 0 || episode > 0 {
		mediaType = "tv"
	}
	title := file.Name
	if extension := filepath.Ext(title); extension != "" {
		title = strings.TrimSuffix(title, extension)
	}
	sha1 := file.Hashes["sha1"]
	preSha1 := file.Hashes["presha1"]
	return Resource{
		SHA1:      sha1,
		PreSHA1:   preSha1,
		Size:      strconv.FormatInt(file.Size, 10),
		Name:      file.Name,
		RawName:   file.Name,
		Path:      cloudPath,
		PickCode:  file.PickCode,
		Title:     title,
		Type:      mediaType,
		Season:    season,
		Episode:   episode,
		Quality:   parseQuality(file.Name),
		Year:      parseYear(cloudPath, file.Name),
		Schema:    "cloud_resource.v1",
		OwnerName: cfg.NodeID,
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

func parseYear(path, name string) string {
	for _, value := range []string{name, path} {
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
	return pan302plugin.EncodeResponse(&pb.LifecycleResponse{Ok: true})
}

func errorResponse(err error) uint64 {
	return pan302plugin.EncodeResponse(&pb.LifecycleResponse{Ok: false, Error: err.Error()})
}
