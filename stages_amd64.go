package simdcsv

import (
	"log"
)

//go:noescape
func stage1_preprocess_buffer(buf []byte, separatorChar uint64, input1 *stage1Input, output1 *stage1Output, postProc *[]uint64, offset uint64) (processed uint64)

//go:noescape
func stage1_preprocess_test(input *stage1Input, output *stage1Output)

//go:noescape
func stage1_preprocess()

//go:noescape
func partialLoad()

func Stage1PreprocessBuffer(buf []byte, separatorChar uint64) ([]uint64) {

	return Stage1PreprocessBufferEx(buf, separatorChar, nil)
}

func Stage1PreprocessBufferEx(buf []byte, separatorChar uint64, postProc *[]uint64) ([]uint64) {

	if postProc == nil {
		_postProc := make([]uint64, 0, 128)
		postProc = &_postProc
	}

	processed, quoted :=uint64(0), uint64(0)
	for {
		inputStage1, outputStage1 := stage1Input{}, stage1Output{}
		inputStage1.quoted = quoted

		processed = stage1_preprocess_buffer(buf, separatorChar, &inputStage1, &outputStage1, postProc, processed)

		if processed >= uint64(len(buf)) {
			break
		}

		// Check if we need to grow the slice for keeping track of the lines to post process
		if len(*postProc) >= cap(*postProc)/2 {
			_postProc := make([]uint64, len(*postProc), cap(*postProc)*2)
			copy(_postProc, (*postProc)[:])
			postProc = &_postProc
		}

		quoted = inputStage1.quoted
	}

	return *postProc
}

//go:noescape
func _stage2_parse_buffer(buf []byte, lastCharIsDelimiter uint64, rows []uint64, columns []string, delimiterChar, separatorChar, quoteChar uint64, input2 *Input, offset uint64, output2 *OutputAsm) (processed uint64)

//go:noescape
func _stage2_parse_masks(buf []byte, masks []uint64, lastCharIsDelimiter uint64, rows []uint64, columns []string, input2 *Input, offset uint64, output2 *OutputAsm) (processed uint64)

func stage2_parse_buffer(buf []byte, rows []uint64, columns []string, delimiterChar, separatorChar, quoteChar uint64, input *Input, offset uint64, output *OutputAsm) (processed uint64) {

	lastCharIsDelimiter := uint64(0)
	if len(buf) > 0 && buf[len(buf)-1] == byte(delimiterChar) {
		lastCharIsDelimiter = 1
	}

	processed = _stage2_parse_buffer(buf, lastCharIsDelimiter, rows, columns, delimiterChar, separatorChar, quoteChar, input, offset, output)
	return
}

func stage2_parse_masks(buf []byte, masks []uint64, rows []uint64, columns []string, delimiterChar uint64, input *Input, offset uint64, output *OutputAsm) (processed uint64) {

	lastCharIsDelimiter := uint64(0)
	if len(buf) > 0 && buf[len(buf)-1] == byte(delimiterChar) {
		lastCharIsDelimiter = 1
	}

	processed = _stage2_parse_masks(buf, masks, lastCharIsDelimiter, rows, columns, input, offset, output)
	return
}

// Perform CSV parsing on a buffer
//
// `records` may be passed in, if non-nil it will be reused
// and grown accordingly
func Stage2ParseBuffer(buf []byte, delimiterChar, separatorChar, quoteChar uint64, records *[][]string) ([][]string, bool) {

	r, _, _, parseError := Stage2ParseBufferEx(buf, delimiterChar, separatorChar, quoteChar, records, nil, nil)
	return r, parseError
}

// Same as above, but allow reuse of `rows` and `columns` slices as well
func Stage2ParseBufferEx(buf []byte, delimiterChar, separatorChar, quoteChar uint64, records *[][]string, rows *[]uint64, columns *[]string) ([][]string, []uint64, []string, /*parsingError*/ bool) {

	errorOut := func() ([][]string, []uint64, []string, /*parsingError*/ bool) {
		*columns = (*columns)[:0]
		*rows = (*rows)[:0]
		return *records, *rows, *columns, true
	}

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
	for {
		processed := stage2_parse_buffer(buf, *rows, *columns, delimiterChar, separatorChar, quoteChar, &inputStage2, offset, &outputStage2)
		if inputStage2.errorOffset != 0 {
			return errorOut()
		}
		if int(processed) >= len(buf) {
			break
		}

		// Sanity check
		if offset == processed {
			log.Fatalf("failed to process anything")
		}
		offset = processed

		// Check whether we need to double columns slice capacity
		if outputStage2.index / 2 >= cap(*columns) / 2 {
			_columns := make([]string, cap(*columns)*2)
			copy(_columns, (*columns)[:outputStage2.index/2])
			columns = &_columns
		}

		// Check whether we need to double rows slice capacity
		if outputStage2.line >= cap(*rows) / 2 {
			_rows := make([]uint64, cap(*rows)*2)
			copy(_rows, (*rows)[:outputStage2.line])
			rows = &_rows
		}
	}

	// Is the final quoted field not closed?
	if inputStage2.quoted != 0 {
		return errorOut()
	}

	//if outputStage2.index >= 2 {
	//	// Sanity check -- we must not point beyond the end of the buffer
	//	if peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-2)*8) != 0 &&
	//		peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-2)*8) - uint64(uintptr(unsafe.Pointer(&buf[0]))) +
	//			peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-1)*8) > uint64(len(buf)) {
	//		log.Fatalf("ERROR: Pointing past end of buffer")
	//	}
	//}

	*columns = (*columns)[:(outputStage2.index)/2]
	*rows = (*rows)[:outputStage2.line]

	for i := 0; i < len(*rows); i += 2 {
		*records = append(*records, (*columns)[(*rows)[i]:(*rows)[i]+(*rows)[i+1]])
	}

	return *records, *rows, *columns, false
}

//go:noescape
func stage2_parse_test(input *Input, offset uint64, output *Output)

//go:noescape
func stage2_parse()

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

	delimiterChar := '\n'
	lastCharIsDelimiter := uint64(0)
	if len(buf) > 0 && buf[len(buf)-1] == byte(delimiterChar) {
		lastCharIsDelimiter = 1
	}

	inputStage2, outputStage2 := NewInput(), OutputAsm{}

	offset := uint64(0)
	processed, quoted := uint64(0), uint64(0)
	for {
		inputStage1, outputStage1 := stage1Input{}, stage1Output{}
		inputStage1.quoted = quoted

		processed = stages_combined_buffer(buf, separatorChar, &inputStage1, &outputStage1, postProc, offset, &inputStage2, &outputStage2, lastCharIsDelimiter, *rows, *columns)
		if inputStage2.errorOffset != 0 {
			return errorOut()
		}
		if int(processed) >= len(buf) {
			break
		}

		// Sanity check
		if offset == processed {
			log.Fatalf("failed to process anything")
		}
		offset = processed

		// Check whether we need to double columns slice capacity
		if outputStage2.index / 2 >= cap(*columns) / 2 {
			_columns := make([]string, cap(*columns)*2)
			copy(_columns, (*columns)[:outputStage2.index/2])
			columns = &_columns
		}

		// Check whether we need to double rows slice capacity
		if outputStage2.line >= cap(*rows) / 2 {
			_rows := make([]uint64, cap(*rows)*2)
			copy(_rows, (*rows)[:outputStage2.line])
			rows = &_rows
		}

		// Check if we need to grow the slice for keeping track of the lines to post process
		if len(*postProc) >= cap(*postProc)/2 {
			_postProc := make([]uint64, len(*postProc), cap(*postProc)*2)
			copy(_postProc, (*postProc)[:])
			postProc = &_postProc
		}

		quoted = inputStage1.quoted
	}

	// Is the final quoted field not closed?
	if inputStage2.quoted != 0 {
		return errorOut()
	}

	//if outputStage2.index >= 2 {
	//	// Sanity check -- we must not point beyond the end of the buffer
	//	if peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-2)*8) != 0 &&
	//		peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-2)*8) - uint64(uintptr(unsafe.Pointer(&buf[0]))) +
	//			peek(uintptr(unsafe.Pointer(&(*columns)[0])), uint64(outputStage2.index-1)*8) > uint64(len(buf)) {
	//		log.Fatalf("ERROR: Pointing past end of buffer")
	//	}
	//}

	*columns = (*columns)[:(outputStage2.index)/2]
	*rows = (*rows)[:outputStage2.line]

	for i := 0; i < len(*rows); i += 2 {
		*records = append(*records, (*columns)[(*rows)[i]:(*rows)[i]+(*rows)[i+1]])
	}

	return *records, *rows, *columns, *postProc, false
}
