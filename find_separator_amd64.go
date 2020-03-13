package simdcsv

import "unsafe"

//go:noescape
func __find_separator()

//go:noescape
func _find_separator(input unsafe.Pointer, separator uint64) (mask uint64)

func find_separator(buf []byte, separator byte) uint64 {
	return _find_separator(unsafe.Pointer(&buf[0]), uint64(separator))
}
