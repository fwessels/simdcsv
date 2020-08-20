package simdcsv

import (
	"unsafe"
	"log"
)

//go:noescape
func _stage2_parse_buffer(buf []byte, lastCharIsDelimiter, delimiterChar, separatorChar, quoteChar uint64, input *Input, offset uint64, output *OutputAsm)

func stage2_parse_buffer(buf []byte, delimiterChar, separatorChar, quoteChar uint64, input *Input, offset uint64, output *OutputAsm) {

	lastCharIsDelimiter := uint64(0)
	if len(buf) > 0 && buf[len(buf)-1] == byte(delimiterChar) {
		lastCharIsDelimiter = 1
	}

	_stage2_parse_buffer(buf, lastCharIsDelimiter, delimiterChar, separatorChar, quoteChar, input, offset, output)
}

// Perform CSV parsing on a buffer
//
// `records` may be passed in, if non-nil it will be reused
// and grown accordingly
func Stage2ParseBuffer(buf []byte, delimiterChar, separatorChar, quoteChar uint64, records *[][]string) [][]string {

	r, _, _ := Stage2ParseBufferEx(buf, delimiterChar, separatorChar, quoteChar, records, nil, nil)
	return r
}

// Same as above, but allow reuse of `rows` and `columns` slices as well
func Stage2ParseBufferEx(buf []byte, delimiterChar, separatorChar, quoteChar uint64, records *[][]string, rows *[]uint64, columns *[]string) ([][]string, []uint64, []string) {

	if records == nil {
		_records := make([][]string, 0, 1024)
		records = &_records
	}
	if rows == nil {
		_rows := make([]uint64, 15024)
		rows = &_rows
	}
	if columns == nil {
		_columns := make([]string, len(*rows)*10)
		columns = &_columns
	}

	input := Input{base: unsafe.Pointer(&buf[0])}
	output := OutputAsm{columns: unsafe.Pointer(&(*columns)[0]), rows: unsafe.Pointer(&(*rows)[0])}

	// TODO: pass in columns and rows as slices -- have assembly put pointer  in output struct
	stage2_parse_buffer(buf, delimiterChar, separatorChar, quoteChar, &input, 0, &output)

	if output.index >= 2 {
		// Sanity check -- we must not point beyond the end of the buffer
		if peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(output.index-2)*8) - uint64(uintptr(unsafe.Pointer(&buf[0]))) +
			peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(output.index-1)*8) > uint64(len(buf)) {
			log.Fatalf("ERORR: Pointing past end of buffer")
		}
	}

	*columns = (*columns)[:(output.index)/2]
	*rows = (*rows)[:output.line]

	*records = (*records)[:0]
	start := 0
	for _, row := range *rows {
		*records = append(*records, (*columns)[start:start+int(row)])
		start += int(row)
	}

	return *records, *rows, *columns
}

//go:noescape
func stage2_parse_test(input *Input, offset uint64, output *Output)

//go:noescape
func stage2_parse()
