package pan302plugin

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	"google.golang.org/protobuf/proto"
)

var (
	requestCounter atomic.Uint64
	allocationMu   sync.Mutex
	allocations    = map[uint32][]byte{}
)

func nextID() string {
	return fmt.Sprintf("go-%d", requestCounter.Add(1))
}

func call(request *pb.HostRequest) (*pb.HostResponse, error) {
	encoded, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}
	responseBuffer := make([]byte, 4096)
	for {
		result := hostCall(
			bytesPointer(encoded),
			uint32(len(encoded)),
			bytesPointer(responseBuffer),
			uint32(len(responseBuffer)),
		)
		if result < 0 {
			return nil, fmt.Errorf("host_call failed with code %d", result)
		}
		required := int(result)
		if required > len(responseBuffer) {
			responseBuffer = make([]byte, required)
			continue
		}
		var response pb.HostResponse
		if err := proto.Unmarshal(responseBuffer[:required], &response); err != nil {
			return nil, err
		}
		if errResult, ok := response.Result.(*pb.HostResponse_Error); ok {
			return nil, fmt.Errorf("%s: %s", errResult.Error.Code, errResult.Error.Message)
		}
		return &response, nil
	}
}

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

func Allocate(size uint32) uint32 {
	if size == 0 {
		return 0
	}
	buffer := make([]byte, size)
	ptr := bytesPointer(buffer)
	allocationMu.Lock()
	allocations[ptr] = buffer
	allocationMu.Unlock()
	return ptr
}

func Free(ptr uint32) {
	if ptr == 0 {
		return
	}
	allocationMu.Lock()
	delete(allocations, ptr)
	allocationMu.Unlock()
}

func EncodeResponse(value proto.Message) uint64 {
	encoded, err := proto.Marshal(value)
	if err != nil || len(encoded) == 0 {
		return 0
	}
	ptr := Allocate(uint32(len(encoded)))
	copy(pointerBytes(ptr, uint32(len(encoded))), encoded)
	return uint64(ptr)<<32 | uint64(uint32(len(encoded)))
}

func DecodeRequest(ptr, length uint32, target proto.Message) error {
	if ptr == 0 || length == 0 {
		return fmt.Errorf("empty request")
	}
	return proto.Unmarshal(pointerBytes(ptr, length), target)
}

func bytesPointer(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&data[0])))
}

func pointerBytes(ptr, length uint32) []byte {
	if ptr == 0 || length == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), int(length))
}
