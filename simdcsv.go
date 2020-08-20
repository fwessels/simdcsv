package simdcsv

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

func ReadAll(buf []byte) (records [][]string) {
	records = Stage2ParseBuffer(buf, '\n', ',', '"', nil)
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

func EnsureFieldsPerRecord(records *[][]string, fieldsPerRecord *int) error {

	if *fieldsPerRecord == 0 {
		if len(*records) > 0 {
			*fieldsPerRecord = len((*records)[0])
		}
	}

	if *fieldsPerRecord > 0 {
		for i, record := range *records {
			if len(record) != *fieldsPerRecord {
				*records = nil
				return errors.New(fmt.Sprintf("record on line %d: wrong number of fields", i+1))
			}
		}
	}
	return nil
}

func TrimLeadingSpace(records *[][]string) {

	for i := 0; i < len(*records); i++ {
		for j := range (*records)[i] {
			(*records)[i][j] = strings.TrimLeftFunc((*records)[i][j], func(r rune) bool {
				return unicode.IsSpace(r)
			})
		}
	}
}
