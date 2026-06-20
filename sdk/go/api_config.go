package pan302plugin

import (
	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

// Config 命名空间，提供插件配置数据读取和更新的相关接口
var Config = configAPI{}

type configAPI struct{}

// Read 读取当前插件的持久化配置数据
func (configAPI) Read() (*pb.ConfigReadResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_ConfigRead{ConfigRead: &pb.ConfigReadRequest{}},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetConfigRead(), nil
}

// Write 更新当前插件的持久化配置数据
func (configAPI) Write(req *pb.ConfigWriteRequest) (*pb.ConfigWriteResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_ConfigWrite{ConfigWrite: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetConfigWrite(), nil
}
