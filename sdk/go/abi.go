package pan302plugin

import (
	"fmt"
	"sync"
	"unsafe"

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
