package simdcsv

import "unsafe"

//go:noescape
func __find_whitespace_and_structurals()

//go:noescape
func _find_whitespace_and_structurals(input, whitespace, structurals unsafe.Pointer)

func find_whitespace_and_structurals(buf []byte, whitespace, structurals *uint64) {
	_find_whitespace_and_structurals(unsafe.Pointer(&buf[0]), unsafe.Pointer(whitespace), unsafe.Pointer(structurals))
}
