package simdcsv

import (
	"encoding/hex"
	"log"
	"math/bits"
	"fmt"
	"bytes"
	"reflect"
	"testing"
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

		result := testStage1PreprocessMasksToMasks(t, []byte(data))

		const expected = `
            first_name,last_name,username RRobertt,"Pi,e",rob  Kenny,"ho  so·",kenny "Robert","Griesemer","gr""i"                            
     quote: 0000000000000000000000000000000000000001000010000000000001000000·1000000010000001010000000001010011010000000000000000000000000000
     quote: 0000000000000000000000000000000000000001000010000000000001000000·1000000010000001010000000001010000010000000000000000000000000000
                                                                                                             ^^                              
 separator: 0000000000100000000010000000000000000010001001000000000010000000·0100000000000000100000000000100000000000000000000000000000000000
 separator: 0000000000100000000010000000000000000010000001000000000010000000·0100000000000000100000000000100000000000000000000000000000000000
                                                      ^                                                                                      
        \r: 0000000000000000000000000000000000000000000000000100000000001000·0000000000000000000000000000000000000000000000000000000000000000
        \r: 0000000000000000000000000000000000000000000000000100000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                                                        ^                                                                    
`

		if result != expected {
			t.Errorf("TestStage1PreprocessMasksToMasks: got %v, want %v", result, expected)
		}
	})

	t.Run("double-quotes-at-end-of-mask", func(t *testing.T) {

		const data = `Robe,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            
first_name,last_name,username1234`

		result := testStage1PreprocessMasksToMasks(t, []byte(data))

		const expected = `
            Robe,"Pi,e",rob  Kenny,"ho  so",kenny "Robert","Griesemer","gr""·i"                             first_name,last_name,username1234
     quote: 0000010000100000000000010000001000000010000001010000000001010011·0100000000000000000000000000000000000000000000000000000000000000
     quote: 0000010000100000000000010000001000000010000001010000000001010000·0100000000000000000000000000000000000000000000000000000000000000
                                                                          ^^                                                                 
 separator: 0000100010010000000000100000000100000000000000100000000000100000·0000000000000000000000000000000000000000010000000001000000000000
 separator: 0000100000010000000000100000000100000000000000100000000000100000·0000000000000000000000000000000000000000010000000001000000000000
                    ^                                                                                                                        
        \r: 0000000000000001000000000010000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
        \r: 0000000000000001000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                      ^                                                                                                      
`

		if result != expected {
			t.Errorf("TestStage1PreprocessMasksToMasks: got %v, want %v", result, expected)
		}
	})

	t.Run("double-quotes-split-over-masks", func(t *testing.T) {

		const data = `Rober,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            
first_name,last_name,username123`

		result := testStage1PreprocessMasksToMasks(t, []byte(data))

		const expected = `
            Rober,"Pi,e",rob  Kenny,"ho  so",kenny "Robert","Griesemer","gr"·"i"                             first_name,last_name,username123
     quote: 0000001000010000000000001000000100000001000000101000000000101001·1010000000000000000000000000000000000000000000000000000000000000
     quote: 0000001000010000000000001000000100000001000000101000000000101000·0010000000000000000000000000000000000000000000000000000000000000
                                                                           ^ ^                                                               
 separator: 0000010001001000000000010000000010000000000000010000000000010000·0000000000000000000000000000000000000000001000000000100000000000
 separator: 0000010000001000000000010000000010000000000000010000000000010000·0000000000000000000000000000000000000000001000000000100000000000
                     ^                                                                                                                       
        \r: 0000000000000000100000000001000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
        \r: 0000000000000000100000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                       ^                                                                                                     
`

		if result != expected {
			t.Errorf("TestStage1PreprocessMasksToMasks: got %v, want %v", result, expected)
		}
	})
}

func diffBitmask(diff1, diff2 string) (diff string) {
	if len(diff1) != len(diff2) {
		log.Fatalf("sizes don't match")
	}


	for i := range diff1 {
		if diff1[i] != diff2[i] {
			diff += "^"
		} else {
			diff += " "
		}
	}

	return diff1 + "\n" + diff2 + "\n" + diff + "\n"
}

func testStage1PreprocessMasksToMasks(t *testing.T, data []byte) string {

	//fmt.Println(hex.Dump(data))
	separatorMasksIn := getBitMasks(data, byte(','))
	quoteMasksIn := getBitMasks(data, byte('"'))
	carriageReturnMasksIn := getBitMasks(data, byte('\r'))

	input0 := stage1Input{quoteMasksIn[0], separatorMasksIn[0], carriageReturnMasksIn[0], quoteMasksIn[1], 0}
	output0 := stage1Output{}
	preprocessMasksToMasks(&input0, &output0)

	input1 := stage1Input{input0.quoteMaskInNext, separatorMasksIn[1], carriageReturnMasksIn[1], 0, input0.quoted}
	output1 := stage1Output{}
	preprocessMasksToMasks(&input1, &output1)

	out := bytes.NewBufferString("")

	fmt.Fprintln(out)
	fmt.Fprintf(out,"            %s", string(bytes.ReplaceAll(bytes.ReplaceAll(data[:64], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))
	fmt.Fprintf(out,"·%s\n", string(bytes.ReplaceAll(bytes.ReplaceAll(data[64:], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))

	fmt.Fprintf(out, diffBitmask(
		fmt.Sprintf("     quote: %064b·%064b", bits.Reverse64(quoteMasksIn[0]), bits.Reverse64(quoteMasksIn[1])),
		fmt.Sprintf("     quote: %064b·%064b", bits.Reverse64(output0.quoteMaskOut), bits.Reverse64(output1.quoteMaskOut))))

	fmt.Fprintf(out, diffBitmask(
		fmt.Sprintf(" separator: %064b·%064b", bits.Reverse64(separatorMasksIn[0]), bits.Reverse64(separatorMasksIn[1])),
		fmt.Sprintf(" separator: %064b·%064b", bits.Reverse64(output0.separatorMaskOut), bits.Reverse64(output1.separatorMaskOut))))

	fmt.Fprintf(out, diffBitmask(
		fmt.Sprintf("        \\r: %064b·%064b", bits.Reverse64(carriageReturnMasksIn[0]), bits.Reverse64(carriageReturnMasksIn[1])),
		fmt.Sprintf("        \\r: %064b·%064b", bits.Reverse64(output0.carriageReturnMaskOut), bits.Reverse64(output1.carriageReturnMaskOut))))

	return out.String()
}

func TestStage1AlternativeMasks(t *testing.T) {

	const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"`

	fmt.Print(hex.Dump([]byte(data)))
	alternativeStage1Masks([]byte(data))
}