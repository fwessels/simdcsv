package simdcsv

//go:noescape
func parse_second_pass(separatorMask, delimiterMask, quoteMask, offset uint64, quoted *uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int, scratch1, scratch2 uint64)
