package simdcsv

import (
	"testing"
)

func TestStage1Preprocessing(t *testing.T) {

	var buffer []byte
	containsDoubleQuotes := true

	delimiter, separator, quote := '\n', ',', '"'
	buf := buffer
	if containsDoubleQuotes {
		buf = preprocessDoubleQuotes(buffer)

		delimiter, separator, quote = PreprocessedDelimiter, PreprocessedSeparator, 0x02
	}

	Stage2Parse(buf, delimiter, separator, quote, stage2_parse_test)
}
