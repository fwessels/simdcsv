package simdcsv

//go:noescape
func stage2_parse_buffer(buf []byte, delimiterChar, separatorChar, quoteChar uint64, input *Input, offset uint64, output *OutputAsm)

//go:noescape
func stage2_parse_test(input *Input, offset uint64, output *Output)

//go:noescape
func stage2_parse()
