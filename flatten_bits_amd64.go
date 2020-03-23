package simdcsv

import "unsafe"

const INDEX_SIZE = 1024

//go:noescape
func __flatten_bits_incremental()

//go:noescape
func _flatten_bits_incremental(base_ptr, pbase unsafe.Pointer, mask, quote_bits uint64, carried unsafe.Pointer, position unsafe.Pointer)

func flatten_bits_incremental(base *[INDEX_SIZE]uint32, base_index *int, mask, quote_bits uint64, carried *int, position *uint64) {
	_flatten_bits_incremental(unsafe.Pointer(&(*base)[0]), unsafe.Pointer(base_index), mask, quote_bits, unsafe.Pointer(carried), unsafe.Pointer(position))
}
