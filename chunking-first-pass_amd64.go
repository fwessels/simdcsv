package simdcsv

//go:noescape
func chunking_first_pass(buf []byte, quoteChar, delimiterChar uint64, quotes *int, even, odd *int)

//go:noescape
func handle_masks_test(quoteMask, newlineMask uint64, quoteNextMask, quotes *uint64, even, odd *int)

//go:noescape
func handle_masks()
