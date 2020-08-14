package simdcsv

import (
	"errors"
	"fmt"
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

func FilterOutComments(records *[][]string, comment byte) {

	// iterate in reverse so as to prevent starting over when removing element
	for i := len(*records)-1; i >= 0; i-- {
		record := (*records)[i]
		if len(record) > 0 && len(record[0]) > 0 && record[0][0] == comment {
			*records = append((*records)[:i], (*records)[i+1:len(*records)]...)
		}
	}
}

func EnsureFieldsPerRecord(records [][]string, fieldsPerRecord int) error {

	if fieldsPerRecord > 0 {
		for i, record := range records {
			if len(record) != fieldsPerRecord {
				return errors.New(fmt.Sprintf("record on line %d: wrong number of fields", i+1))
			}
		}
	}
	return nil
}
