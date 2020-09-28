package simdcsv

import (
	"unsafe"
	"log"
)

//go:noescape
func stages_combined_buffer(buf []byte, separatorChar uint64, input1 *stage1Input, output1 *stage1Output, postProc *[]uint64, offset uint64, input2 *Input, output2 *OutputAsm, lastCharIsDelimiter uint64, rows []uint64, columns []string) (processed uint64)

func StagesCombined(buf []byte, separatorChar uint64, records *[][]string) ([][]string, []uint64, /*parsingError*/ bool) {

	r, _, _, postProc, parseError := StagesCombinedEx(buf, separatorChar, records, nil, nil, nil)
	return r, postProc, parseError
}

func StagesCombinedEx(buf []byte, separatorChar uint64, records *[][]string, rows *[]uint64, columns *[]string, postProc *[]uint64) ([][]string, []uint64, []string, []uint64, /*parsingError*/ bool) {

	errorOut := func() ([][]string, []uint64, []string, []uint64, /*parsingError*/ bool) {
		*columns = (*columns)[:0]
		*rows = (*rows)[:0]
		return *records, *rows, *columns, *postProc, true
	}

	inputStage1, outputStage1 := stage1Input{}, stage1Output{}
	if postProc == nil {
		_postProc := make([]uint64, 0, 128*128*2)
		postProc = &_postProc
	}

	*postProc = (*postProc)[:0]

	if rows == nil {
		_rows := make([]uint64, 1024) // do not reserve less than 128
		rows = &_rows
	}
	if columns == nil {
		_columns := make([]string, 10240)
		columns = &_columns
	}

	// for repeat calls the actual lengths may have been reduced, so set arrays to maximum size
	*rows = (*rows)[:cap(*rows)]
	*columns = (*columns)[:cap(*columns)]

	if records == nil {
		_records := make([][]string, 0, 1024)
		records = &_records
	}

	*records = (*records)[:0]

	inputStage2, outputStage2 := NewInput(), OutputAsm{}

	offset := uint64(0)

	delimiterChar := '\n'
	lastCharIsDelimiter := uint64(0)
	if len(buf) > 0 && buf[len(buf)-1] == byte(delimiterChar) {
		lastCharIsDelimiter = 1
	}

	stages_combined_buffer(buf, separatorChar, &inputStage1, &outputStage1, postProc, offset, &inputStage2, &outputStage2, lastCharIsDelimiter, *rows, *columns)

	// Is the final quoted field not closed?
	if inputStage2.quoted != 0 {
		return errorOut()
	}

	if outputStage2.index >= 2 {
		// Sanity check -- we must not point beyond the end of the buffer
		if peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-2)*8) != 0 &&
			peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-2)*8) - uint64(uintptr(unsafe.Pointer(&buf[0]))) +
				peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-1)*8) > uint64(len(buf)) {
			log.Fatalf("ERROR: Pointing past end of buffer")
		}
	}

	*columns = (*columns)[:(outputStage2.index)/2]
	*rows = (*rows)[:outputStage2.line]

	for i := 0; i < len(*rows); i += 2 {
		*records = append(*records, (*columns)[(*rows)[i]:(*rows)[i]+(*rows)[i+1]])
	}

	return *records, *rows, *columns, *postProc, false
}
