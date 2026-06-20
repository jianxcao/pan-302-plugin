package pan302plugin

import (
	"fmt"
	"sync/atomic"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	"google.golang.org/protobuf/proto"
)

var requestCounter atomic.Uint64

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
