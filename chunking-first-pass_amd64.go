package simdcsv

//go:noescape
func chunking_first_pass(buf []byte, separator uint64) (out uint64, positionDelimiterEven, positionDelimiterOdd, quotes int64)

//go:noescape
func handle_masks(quoteMask, newlineMask, lastCharIsQuote uint64, quotes *uint64, even, odd *int)
