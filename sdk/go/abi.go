package pan302plugin

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

var (
	allocationMu sync.Mutex
	allocations  = map[uint32][]byte{}
)

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

func DecodeEventRequest(ptr, length uint32) (proto.Message, error) {
	if ptr == 0 || length == 0 {
		return nil, fmt.Errorf("empty request")
	}
	payload := append([]byte(nil), pointerBytes(ptr, length)...)
	return decodeEventPayload(payload)
}

func decodeEventPayload(payload []byte) (proto.Message, error) {
	eventName, err := eventNameFromPayload(payload)
	if err != nil {
		return nil, err
	}
	var event proto.Message
	switch {
	case strings.HasPrefix(eventName, "strm."):
		event = &pb.StrmEvent{}
	case strings.HasPrefix(eventName, "media."):
		event = &pb.MediaEvent{}
	case strings.HasPrefix(eventName, "resource."):
		event = &pb.ResourceReadyEvent{}
	default:
		return nil, fmt.Errorf("unsupported event %q", eventName)
	}
	if err := proto.Unmarshal(payload, event); err != nil {
		return nil, fmt.Errorf("decode %s: %w", eventName, err)
	}
	return event, nil
}

func eventNameFromPayload(payload []byte) (string, error) {
	for len(payload) > 0 {
		number, wireType, tagLength := protowire.ConsumeTag(payload)
		if tagLength < 0 {
			return "", protowire.ParseError(tagLength)
		}
		payload = payload[tagLength:]
		if number == 2 && wireType == protowire.BytesType {
			value, valueLength := protowire.ConsumeBytes(payload)
			if valueLength < 0 {
				return "", protowire.ParseError(valueLength)
			}
			if len(value) == 0 {
				return "", fmt.Errorf("event name is empty")
			}
			return string(value), nil
		}
		valueLength := protowire.ConsumeFieldValue(number, wireType, payload)
		if valueLength < 0 {
			return "", protowire.ParseError(valueLength)
		}
		payload = payload[valueLength:]
	}
	return "", fmt.Errorf("event name is missing")
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
