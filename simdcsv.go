package simdcsv

import (
	"unsafe"
)

func ReadAll(buf []byte) (rows [][]string) {

	input := Input{base: uint64(uintptr(unsafe.Pointer(&buf[0])))}
	rows = make([][]string, 10000 + 100)
	columns := make([]string, len(rows)*100)
	output := OutputAsm{unsafe.Pointer(&columns[0]), 1, unsafe.Pointer(&rows[0]), 0, uint64(uintptr(unsafe.Pointer(&columns[0]))), 0, uint64(cap(columns))}

	stage2_parse_buffer(buf, '\n', ',', '"', &input, 0, &output)

	rows = rows[:output.line/3]

	return
}
