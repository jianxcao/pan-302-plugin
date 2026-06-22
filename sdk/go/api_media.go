package pan302plugin

import (
	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

// Media 命名空间，提供读取宿主媒体服务器配置的接口。
var Media = mediaAPI{}

type mediaAPI struct{}

// ServerConfig 读取宿主配置的 Emby/Jellyfin 连接信息。
func (mediaAPI) ServerConfig() (*pb.MediaServerConfigReadResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_MediaServerConfigRead{MediaServerConfigRead: &pb.MediaServerConfigReadRequest{}},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetMediaServerConfigRead(), nil
}
