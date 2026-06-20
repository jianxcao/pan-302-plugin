package pan302plugin

import (
	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

// Strm 命名空间，提供本地 STRM 播放文件生成和删除的相关接口
var Strm = strmAPI{}

type strmAPI struct{}

// Write 请求宿主在本地生成 STRM 播放文件
func (strmAPI) Write(req *pb.StrmWriteRequest) (*pb.StrmOperationResult, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_StrmWrite{StrmWrite: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetStrmResult(), nil
}

// Delete 请求宿主删除本地已生成的 STRM 播放文件
func (strmAPI) Delete(req *pb.StrmDeleteRequest) (*pb.StrmOperationResult, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_StrmDelete{StrmDelete: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetStrmResult(), nil
}
