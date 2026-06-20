package pan302plugin

import (
	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

// 插件所支持的网盘能力位预设常量
const (
	CapList        = "list"
	CapLink        = "link"
	CapMkdir       = "mkdir"
	CapDelete      = "delete"
	CapRename      = "rename"
	CapMove        = "move"
	CapCopy        = "copy"
	CapPut         = "put"
	CapRapidUpload = "rapid_upload"
	CapStrm        = "strm"
)

// Driver 命名空间，提供操作宿主网盘驱动的所有相关接口
var Driver = driverAPI{}

type driverAPI struct{}

// List 获取宿主中当前所有已注册的网盘驱动列表
func (driverAPI) List() (*pb.DriverListResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverList{DriverList: &pb.DriverListRequest{}},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverList(), nil
}

// Read 获取指定网盘驱动的详细元数据信息
func (driverAPI) Read(req *pb.DriverReadRequest) (*pb.DriverReadResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverRead{DriverRead: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverRead(), nil
}

// Link 请求解析某个网盘文件的直链 (302 播放地址)
func (driverAPI) Link(req *pb.DriverLinkRequest) (*pb.DriverLinkResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverLink{DriverLink: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverLink(), nil
}

// Mkdir 在网盘驱动中创建文件夹
func (driverAPI) Mkdir(req *pb.DriverMkdirRequest) (*pb.DriverObject, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverMkdir{DriverMkdir: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverMkdir(), nil
}

// Rename 重命名网盘中的文件或文件夹
func (driverAPI) Rename(req *pb.DriverRenameRequest) (*pb.DriverRenameResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverRename{DriverRename: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverRename(), nil
}

// Delete 删除网盘中的文件或文件夹
func (driverAPI) Delete(req *pb.DriverDeleteRequest) (*pb.DriverDeleteResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverDelete{DriverDelete: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverDelete(), nil
}

// Move 移动网盘中的文件或文件夹
func (driverAPI) Move(req *pb.DriverMoveCopyRequest) (*pb.DriverMoveCopyResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverMove{DriverMove: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverMoveCopy(), nil
}

// Copy 复制网盘中的文件或文件夹
func (driverAPI) Copy(req *pb.DriverMoveCopyRequest) (*pb.DriverMoveCopyResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverCopy{DriverCopy: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverMoveCopy(), nil
}

// HasCapability 判断某个网盘是否包含指定能力位
func (driverAPI) HasCapability(driver *pb.DriverInfo, cap string) bool {
	for _, c := range driver.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}
