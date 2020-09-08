package simdcsv

import (
	"encoding/csv"
	"encoding/hex"
	"log"
	"math/bits"
	"fmt"
	"bytes"
	"strings"
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

	t.Run("go", func(t *testing.T) {
		testStage1PreprocessMasksToMasksFunc(t, preprocessMasksToMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testStage1PreprocessMasksToMasksFunc(t, stage1_preprocess_test)
	})
}

func testStage1PreprocessMasksToMasksFunc(t *testing.T, f func(input *stage1Input, output *stage1Output)) {

	t.Run("simple", func(t *testing.T) {

		const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            `

		result := testStage1PreprocessMasksToMasks(t, []byte(data), f)

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

		result := testStage1PreprocessMasksToMasks(t, []byte(data), f)

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

		result := testStage1PreprocessMasksToMasks(t, []byte(data), f)

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

func testStage1PreprocessMasksToMasks(t *testing.T, data []byte, f func(input *stage1Input, output *stage1Output)) string {

	//fmt.Println(hex.Dump(data))
	separatorMasksIn := getBitMasks(data, byte(','))
	quoteMasksIn := getBitMasks(data, byte('"'))
	carriageReturnMasksIn := getBitMasks(data, byte('\r'))

	input0 := stage1Input{quoteMasksIn[0], separatorMasksIn[0], carriageReturnMasksIn[0], quoteMasksIn[1], 0}
	output0 := stage1Output{}
	f(&input0, &output0)

	input1 := stage1Input{input0.quoteMaskInNext, separatorMasksIn[1], carriageReturnMasksIn[1], 0, input0.quoted}
	output1 := stage1Output{}
	f(&input1, &output1)

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

func TestStage1MaskingOut(t *testing.T) {

	const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Grie                            semer","gr""i"`

	input, output := stage1Input{} ,stage1Output{}
	debug := [32]byte{}

	buf := []byte(data)
	stage1_preprocess_buffer(buf, &input, &output, &debug)

	out := bytes.NewBufferString("")
	fmt.Fprintf(out, hex.Dump(buf))

	const expected =
`00000000  66 69 72 73 74 5f 6e 61  6d 65 02 6c 61 73 74 5f  |first_name.last_|
00000010  6e 61 6d 65 02 75 73 65  72 6e 61 6d 65 0a 52 52  |name.username.RR|
00000020  6f 62 65 72 74 74 02 03  50 69 2c 65 03 02 72 6f  |obertt..Pi,e..ro|
00000030  62 0a 0a 4b 65 6e 6e 79  02 03 68 6f 0d 0a 73 6f  |b..Kenny..ho..so|
00000040  03 02 6b 65 6e 6e 79 0a  03 52 6f 62 65 72 74 03  |..kenny..Robert.|
00000050  02 03 47 72 69 65 20 20  20 20 20 20 20 20 20 20  |..Grie          |
00000060  20 20 20 20 20 20 20 20  20 20 20 20 20 20 20 20  |                |
00000070  20 20 73 65 6d 65 72 03  02 03 67 72 22 22 69 03  |  semer...gr""i.|
`

	if out.String() != expected {
		t.Errorf("TestStage1MaskingOut: got %v, want %v", out.String(), expected)
	}

	simdrecords := Stage2ParseBuffer(buf, 0xa, preprocessedSeparator, preprocessedQuote, nil)

	//
	// postprocess
	//   replace "" to " in specific columns
	//   replace \r\n to \n in specific columns
	simdrecords[3][2] = strings.ReplaceAll(simdrecords[3][2], "\"\"", "\"")
	simdrecords[2][1] = strings.ReplaceAll(simdrecords[2][1], "\r\n", "\n")

	r := csv.NewReader(bytes.NewReader([]byte(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("encoding/csv: %v", err)
	}

	if !reflect.DeepEqual(simdrecords, records) {
		log.Fatalf("TestStage1MaskingOut: got %v, want %v", simdrecords, records)
	}
}

func TestStage1AlternativeMasks(t *testing.T) {

	const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"`

	fmt.Print(hex.Dump([]byte(data)))
	alternativeStage1Masks([]byte(data))
}

func BenchmarkStage1PreprocessingMasks( b *testing.B) {

	const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Grie                           semer","gr""i"
`

	const repeat = 50
	dataN := make([]byte, 128*repeat)
	dataN = bytes.Repeat([]byte(data), repeat)

	b.SetBytes(int64(len(dataN)))
	b.ReportAllocs()
	b.ResetTimer()

	buf := make([]byte, len(dataN))
	for i := 0; i < b.N; i++ {

		input, output := stage1Input{} ,stage1Output{}
		debug := [32]byte{}

		copy(buf, dataN)
		stage1_preprocess_buffer(buf, &input, &output, &debug)
	}
}
