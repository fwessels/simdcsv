package simdcsv

//go:noescape
func chunking_first_pass(buf []byte, separator uint64) (out uint64, positionDelimiterEven, positionDelimiterOdd, quotes int64)

//go:noescape
func handle_masks(quoteMask, newlineMask uint64, even, odd *int, quotes *uint64)
