package simdcsv

//go:noescape
func parse_block_second_pass(buf []byte, delimiterChar, separatorChar, quoteChar uint64, input *Input, offset uint64, columns *[10240]uint64, index *int, rows *[640]uint64, line *int)

//go:noescape
func parse_second_pass_test(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int)

//go:noescape
func parse_second_pass()
