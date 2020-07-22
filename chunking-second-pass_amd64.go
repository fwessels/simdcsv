package simdcsv

//go:noescape
func parse_second_pass(separatorMask, delimiterMask, quoteMask uint64, output *[128]uint64, index *int, quoted *uint64)
