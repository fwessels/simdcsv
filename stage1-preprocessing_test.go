package simdcsv

import (
	"testing"
	"bytes"
	"fmt"
	"log"
	"reflect"
	"io/ioutil"
	"encoding/hex"
	"encoding/csv"
)

func TestStage1Preprocessing(t *testing.T) {

	var buffer []byte
	containsDoubleQuotes := true

	delimiter, separator, quote := '\n', ',', '"'
	buf := buffer
	if containsDoubleQuotes {
		buf = preprocessDoubleQuotes(buffer)

		delimiter, separator, quote = preprocessedDelimiter, preprocessedSeparator, preprocessedQuote
	}

	Stage2Parse(buf, delimiter, separator, quote, stage2_parse_test)
}

func testStage1PreprocessDoubleQuotes(t *testing.T, data []byte) {

	preprocessed := preprocessDoubleQuotes(data)

	simdrecords := Stage2ParseBuffer(preprocessed, preprocessedDelimiter, preprocessedSeparator, preprocessedQuote, nil)
	records := EncodingCsv(data)

	if !reflect.DeepEqual(simdrecords, records) {
		t.Errorf("testStage1PreprocessDoubleQuotes: got: %v want: %v", simdrecords, records)
	}
}

func TestStage1PreprocessDoubleQuotes(t *testing.T) {

	t.Run("double-quotes", func(t *testing.T) {

		const data = `first_name,last_name,username
"Robert","Pike",rob
Kenny,Thompson,kenny
"Robert","Griesemer","gr""i"
Donald,"Du""ck",don
Dagobert,Duck,dago
`
		testStage1PreprocessDoubleQuotes(t, []byte(data))
	})

	t.Run("newline-in-quoted-field", func(t *testing.T) {

		const data = `first_name,last_name,username
"Robert","Pike",rob
Kenny,Thompson,kenny
"Robert","Griesemer","gr""i"
Donald,"Du""c` + "\n" + `k",don
Dagobert,Duck,dago
`
		testStage1PreprocessDoubleQuotes(t, []byte(data))
	})
}

