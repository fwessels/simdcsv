package simdcsv

//go:noescape
func parse_block_second_pass(buf []byte, delimiterChar, separatorChar, quoteChar uint64, input *Input, offset uint64, output *OutputBig)

//go:noescape
func parse_second_pass_test(input *Input, offset uint64, output *Output)

//go:noescape
func parse_second_pass()
