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

func DriverList() (*pb.DriverListResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverList{DriverList: &pb.DriverListRequest{}},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverList(), nil
}

func DriverRead(req *pb.DriverReadRequest) (*pb.DriverReadResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverRead{DriverRead: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverRead(), nil
}

func DriverLink(req *pb.DriverLinkRequest) (*pb.DriverLinkResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverLink{DriverLink: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverLink(), nil
}

func DriverMkdir(req *pb.DriverMkdirRequest) (*pb.DriverObject, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverMkdir{DriverMkdir: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverMkdir(), nil
}

func DriverRename(req *pb.DriverRenameRequest) (*pb.DriverRenameResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverRename{DriverRename: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverRename(), nil
}

func DriverDelete(req *pb.DriverDeleteRequest) (*pb.DriverDeleteResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverDelete{DriverDelete: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverDelete(), nil
}

func DriverMove(req *pb.DriverMoveCopyRequest) (*pb.DriverMoveCopyResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverMove{DriverMove: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverMoveCopy(), nil
}

func DriverCopy(req *pb.DriverMoveCopyRequest) (*pb.DriverMoveCopyResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_DriverCopy{DriverCopy: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetDriverMoveCopy(), nil
}

func StrmWrite(req *pb.StrmWriteRequest) (*pb.StrmOperationResult, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_StrmWrite{StrmWrite: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetStrmResult(), nil
}

func StrmDelete(req *pb.StrmDeleteRequest) (*pb.StrmOperationResult, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_StrmDelete{StrmDelete: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetStrmResult(), nil
}

func ConfigRead() (*pb.ConfigReadResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_ConfigRead{ConfigRead: &pb.ConfigReadRequest{}},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetConfigRead(), nil
}

func ConfigWrite(req *pb.ConfigWriteRequest) (*pb.ConfigWriteResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_ConfigWrite{ConfigWrite: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetConfigWrite(), nil
}

func Log(level, message string, fields map[string]string) {
	_, _ = call(&pb.HostRequest{
		Id: nextID(),
		Request: &pb.HostRequest_LogWrite{
			LogWrite: &pb.LogWriteRequest{
				Level:   level,
				Message: message,
				Fields:  fields,
			},
		},
	})
}

func RequestHTTP(request *pb.HTTPRequestArgs) (*pb.HTTPResponseData, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_HttpRequest{HttpRequest: request},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetHttpResponse(), nil
}

func HasCapability(driver *pb.DriverInfo, cap string) bool {
	for _, c := range driver.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}
