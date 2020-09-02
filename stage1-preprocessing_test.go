package simdcsv

import (
	"testing"
	"bytes"
	"reflect"
	"fmt"
	"encoding/hex"
)

func testStage1PreprocessDoubleQuotes(t *testing.T, data []byte) {

	preprocessed := preprocessDoubleQuotes(data)

	simdrecords := Stage2ParseBuffer(preprocessed, preprocessedDelimiter, preprocessedSeparator, preprocessedQuote, nil)
	records := EncodingCsv(data)

	if !reflect.DeepEqual(simdrecords, records) {
		t.Errorf("testStage1PreprocessDoubleQuotes: got: %v want: %v", simdrecords, records)
	}
}

func TestStage1PreprocessDoubleQuotes(t *testing.T) {

	const first = `first_name,last_name,username
"Robert","Pike",rob
Kenny,Thompson,kenny
"Robert","Griesemer","gr""i"
Donald,"Du""c`

	const second = `k",don
Dagobert,Duck,dago
`
	t.Run("double-quotes", func(t *testing.T) {

		const data = first + second
		testStage1PreprocessDoubleQuotes(t, []byte(data))
	})

	t.Run("newline-in-quoted-field", func(t *testing.T) {

		const data = first + "\n" + second
		testStage1PreprocessDoubleQuotes(t, []byte(data))
	})

	t.Run("carriage-return-in-quoted-field", func(t *testing.T) {

		const data = first + "\r\n" + second
		testStage1PreprocessDoubleQuotes(t, []byte(data))
	})
}

func TestStage1PreprocessMasksToMasks(t *testing.T) {
	t.Run("simple", func(t *testing.T) {

		const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            `

		testStage1PreprocessMasksToMasks(t, []byte(data))
	})

	t.Run("double-quotes-at-end-of-mask", func(t *testing.T) {

		const data = `Robe,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            
first_name,last_name,username1234`

		testStage1PreprocessMasksToMasks(t, []byte(data))
	})

	t.Run("double-quotes-split-over-masks", func(t *testing.T) {

		const data = `Rober,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            
first_name,last_name,username123`

		testStage1PreprocessMasksToMasks(t, []byte(data))
	})
}

func testStage1PreprocessMasksToMasks(t *testing.T, data []byte) {

	fmt.Println(hex.Dump(data))
	separatorMasksIn := getBitMasks(data, byte(','))
	quoteMasksIn := getBitMasks(data, byte('"'))
	carriageReturnMasksIn := getBitMasks(data, byte('\r'))

	quoteMaskNew := quoteMasksIn[1]
	quoted := false
	quoteMaskOut0, separatorMaskOut0, carriageReturnMaskOut0 := preprocessMasksToMasks(quoteMasksIn[0], separatorMasksIn[0], carriageReturnMasksIn[0], &quoteMaskNew, &quoted)

	quoteMasksIn1 := quoteMaskNew
	quoteMaskNew = 0
	quoteMaskOut1, separatorMaskOut1, carriageReturnMaskOut1 := preprocessMasksToMasks(quoteMasksIn1, separatorMasksIn[1], carriageReturnMasksIn[1], &quoteMaskNew, &quoted)

	fmt.Println()
	fmt.Printf("            %s", string(bytes.ReplaceAll(bytes.ReplaceAll(data[:64], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))
	fmt.Printf("·%s\n", string(bytes.ReplaceAll(bytes.ReplaceAll(data[64:], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))

	fmt.Printf("     quote: %064b·%064b\n", bits.Reverse64(quoteMasksIn[0]), bits.Reverse64(quoteMasksIn[1]))
	fmt.Printf("     quote: %064b·%064b\n", bits.Reverse64(quoteMaskOut0), bits.Reverse64(quoteMaskOut1))
	fmt.Println()
	fmt.Printf(" separator: %064b·%064b\n", bits.Reverse64(separatorMasksIn[0]), bits.Reverse64(separatorMasksIn[1]))
	fmt.Printf(" separator: %064b·%064b\n", bits.Reverse64(separatorMaskOut0), bits.Reverse64(separatorMaskOut1))
	fmt.Println()
	fmt.Printf("        \\r: %064b·%064b\n", bits.Reverse64(carriageReturnMasksIn[0]), bits.Reverse64(carriageReturnMasksIn[1]))
	fmt.Printf("        \\r: %064b·%064b\n", bits.Reverse64(carriageReturnMaskOut0), bits.Reverse64(carriageReturnMaskOut1))
}

func TestStage1AlternativeMasks(t *testing.T) {

	const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"`

	fmt.Print(hex.Dump([]byte(data)))
	alternativeStage1Masks([]byte(data))
}