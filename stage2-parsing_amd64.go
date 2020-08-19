package simdcsv

//go:noescape
func _stage2_parse_buffer(buf []byte, lastCharIsDelimiter, delimiterChar, separatorChar, quoteChar uint64, input *Input, offset uint64, output *OutputAsm)

func stage2_parse_buffer(buf []byte, delimiterChar, separatorChar, quoteChar uint64, input *Input, offset uint64, output *OutputAsm) {

	lastCharIsDelimiter := uint64(0)
	if len(buf) > 0 && buf[len(buf)-1] == byte(delimiterChar) {
		lastCharIsDelimiter = 1
	}

	_stage2_parse_buffer(buf, lastCharIsDelimiter, delimiterChar, separatorChar, quoteChar, input, offset, output)
}

//go:noescape
func stage2_parse_test(input *Input, offset uint64, output *Output)

//go:noescape
func stage2_parse()
