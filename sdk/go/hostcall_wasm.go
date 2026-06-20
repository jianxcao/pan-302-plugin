//go:build wasip1 && wasm

package pan302plugin

//go:wasmimport pan302_v1 host_call
func hostCall(requestPtr, requestLen, responsePtr, responseCapacity uint32) int32
