//go:build !wasip1 || !wasm

package pan302plugin

func hostCall(_, _, _, _ uint32) int32 {
	return -4
}
