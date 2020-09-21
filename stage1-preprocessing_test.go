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

func TestStage1PreprocessMasksToMasks(t *testing.T) {

	t.Run("go", func(t *testing.T) {
		testStage1PreprocessMasksToMasksFunc(t, preprocessMasksToMasksInverted)
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

	t.Run("single-carriage-return", func(t *testing.T) {

		const data = `Rober,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr`+"\r"+`i"                             
first_name,last_name,username123`

		result := testStage1PreprocessMasksToMasks(t, []byte(data), f)

		const expected = `
            Rober,"Pi,e",rob  Kenny,"ho  so",kenny "Robert","Griesemer","gr ·i"                              first_name,last_name,username123
     quote: 0000001000010000000000001000000100000001000000101000000000101000·0100000000000000000000000000000000000000000000000000000000000000
     quote: 0000001000010000000000001000000100000001000000101000000000101000·0100000000000000000000000000000000000000000000000000000000000000
                                                                                                                                             
 separator: 0000010001001000000000010000000010000000000000010000000000010000·0000000000000000000000000000000000000000001000000000100000000000
 separator: 0000010000001000000000010000000010000000000000010000000000010000·0000000000000000000000000000000000000000001000000000100000000000
                     ^                                                                                                                       
        \r: 0000000000000000100000000001000000000000000000000000000000000001·0000000000000000000000000000000000000000000000000000000000000000
        \r: 0000000000000000100000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                       ^                                   ^                                                                 
`

		if result != expected {
			t.Errorf("TestStage1PreprocessMasksToMasks: got %v, want %v", result, expected)
		}
	})

	t.Run("carriage-return-new-line-split-over-masks", func(t *testing.T) {

		const data = `Rober,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr`+"\r\n"+`i"                            
first_name,last_name,username123`

		result := testStage1PreprocessMasksToMasks(t, []byte(data), f)

		const expected = `
            Rober,"Pi,e",rob  Kenny,"ho  so",kenny "Robert","Griesemer","gr · i"                             first_name,last_name,username123
     quote: 0000001000010000000000001000000100000001000000101000000000101000·0010000000000000000000000000000000000000000000000000000000000000
     quote: 0000001000010000000000001000000100000001000000101000000000101000·0010000000000000000000000000000000000000000000000000000000000000
                                                                                                                                             
 separator: 0000010001001000000000010000000010000000000000010000000000010000·0000000000000000000000000000000000000000001000000000100000000000
 separator: 0000010000001000000000010000000010000000000000010000000000010000·0000000000000000000000000000000000000000001000000000100000000000
                     ^                                                                                                                       
        \r: 0000000000000000100000000001000000000000000000000000000000000001·0000000000000000000000000000000000000000000000000000000000000000
        \r: 0000000000000000100000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                       ^                                   ^                                                                 
`

		if result != expected {
			t.Errorf("TestStage1PreprocessMasksToMasks: got %v, want %v", result, expected)
		}
	})

	t.Run("bare-cr-2", func(t *testing.T) {

		const data = "a,b\rc,d\r\n                                                                                                                       "

		result := testStage1PreprocessMasksToMasks(t, []byte(data), f)

		const expected = `
            a,b c,d                                                         ·                                                                
     quote: 0000000000000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
     quote: 0000000000000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                                                                                                                             
 separator: 0100010000000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
 separator: 0100010000000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                                                                                                                             
        \r: 0001000100000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
        \r: 0000000100000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
               ^                                                                                                                             
`
		if result != expected {
			t.Errorf("TestStage1PreprocessMasksToMasks: got %v, want %v", result, expected)
		}
	})

	t.Run("bare-cr-at-end-of-mask", func(t *testing.T) {

		const data = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\rc,d\r\n                                                            "

		result := testStage1PreprocessMasksToMasks(t, []byte(data), f)

		const expected = `
            aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb c·,d                                                              
     quote: 0000000000000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
     quote: 0000000000000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                                                                                                                             
 separator: 0000000000000000000000000000001000000000000000000000000000000000·1000000000000000000000000000000000000000000000000000000000000000
 separator: 0000000000000000000000000000001000000000000000000000000000000000·1000000000000000000000000000000000000000000000000000000000000000
                                                                                                                                             
        \r: 0000000000000000000000000000000000000000000000000000000000000010·0010000000000000000000000000000000000000000000000000000000000000
        \r: 0000000000000000000000000000000000000000000000000000000000000000·0010000000000000000000000000000000000000000000000000000000000000
                                                                          ^                                                                  
`
		if result != expected {
			t.Errorf("TestStage1PreprocessMasksToMasks: got %v, want %v", result, expected)
		}
	})

	t.Run("bare-cr-split-over-masks", func(t *testing.T) {

		const data = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\rc,d\r\n                                                           "

		result := testStage1PreprocessMasksToMasks(t, []byte(data), f)

		const expected = `
            aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb ·c,d                                                             
     quote: 0000000000000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
     quote: 0000000000000000000000000000000000000000000000000000000000000000·0000000000000000000000000000000000000000000000000000000000000000
                                                                                                                                             
 separator: 0000000000000000000000000000000100000000000000000000000000000000·0100000000000000000000000000000000000000000000000000000000000000
 separator: 0000000000000000000000000000000100000000000000000000000000000000·0100000000000000000000000000000000000000000000000000000000000000
                                                                                                                                             
        \r: 0000000000000000000000000000000000000000000000000000000000000001·0001000000000000000000000000000000000000000000000000000000000000
        \r: 0000000000000000000000000000000000000000000000000000000000000000·0001000000000000000000000000000000000000000000000000000000000000
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
	newlineMasksIn := getBitMasks(data, byte('\n'))

	input := stage1Input{quoteMasksIn[0], separatorMasksIn[0], carriageReturnMasksIn[0], quoteMasksIn[1], 0, newlineMasksIn[0], newlineMasksIn[1]}
	output0 := stage1Output{}
	f(&input, &output0)

	input = stage1Input{input.quoteMaskInNext, separatorMasksIn[1], carriageReturnMasksIn[1], 0, input.quoted, newlineMasksIn[1], 0}
	output1 := stage1Output{}
	f(&input, &output1)

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

	buf := []byte(data)

	postProc := Stage1PreprocessBuffer(buf, uint64(','))

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

	simdrecords, parsingError := Stage2ParseBuffer(buf, 0xa, preprocessedSeparator, preprocessedQuote, nil)
	if parsingError {
		t.Errorf("TestStage1MaskingOut: unexpected parsing error")
	}

	for _, ppr := range getPostProcRows(buf, postProc, simdrecords) {
		for r := ppr.start; r < ppr.end; r++ {
			for c := range  simdrecords[r] {
				simdrecords[r][c] = strings.ReplaceAll(simdrecords[r][c], "\"\"", "\"")
				simdrecords[r][c] = strings.ReplaceAll(simdrecords[r][c], "\r\n", "\n")
			}
		}
	}

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
	postProc := make([]uint64, 0, len(buf)>>6)

	for i := 0; i < b.N; i++ {

		copy(buf, dataN)
		postProc = postProc[:0]
		Stage1PreprocessBufferEx(buf, uint64(','), &postProc)
	}
}

func TestTrailingCRs(t *testing.T) {

	for cnt := 1; cnt <= 150; cnt++ {

		// TOOD: Handle special case when length is multiple of 64 bytes
		// we need to look ahead
		if cnt == 63 {
			continue
		} else if cnt == 64+63 {
			continue
		}

		input := strings.Repeat("f", cnt) + "\r"
		output := [][]string{{strings.Repeat("f", cnt)}}

		r := NewReader(strings.NewReader(input))

		out, err := r.ReadAll()
		if err != nil {
			t.Errorf("TestTrailingCR() error:%v", err)
		}
		if !reflect.DeepEqual(out, output) {
			t.Errorf("TestTrailingCR() output:\ngot  %q\nwant %q", out, output)
		}
	}
}

func TestStage1DeterminePostProcRows(t *testing.T) {

	t.Run("none", func(t *testing.T) {
		const data = `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`
		pprows := testStage1DeterminePostProcRows(t, []byte(data))
		expected := []postProcRow{}

		if !reflect.DeepEqual(pprows, expected) {
			log.Fatalf("TestStage1DeterminePostProcRows: got %v, want %v", pprows, expected)
		}
	})

	t.Run("double-quote", func(t *testing.T) {
		const data = `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Grie""semer","gri"
`
		pprows := testStage1DeterminePostProcRows(t, []byte(data))
		expected := []postProcRow{{2,4}}

		if !reflect.DeepEqual(pprows, expected) {
			log.Fatalf("TestStage1DeterminePostProcRows: got %v, want %v", pprows, expected)
		}
	})

	t.Run("quoted-CRLF", func(t *testing.T) {
		const data = `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Grie`+"\r\n"+`semer","gri"
`
		pprows := testStage1DeterminePostProcRows(t, []byte(data))
		expected := []postProcRow{{2,4}}

		if !reflect.DeepEqual(pprows, expected) {
			log.Fatalf("TestStage1DeterminePostProcRows: got %v, want %v", pprows, expected)
		}
	})

	t.Run("multiple-double-quotes", func(t *testing.T) {
		const data = `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Grie""semer","gri"
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
"Rob","Pike",rob
Ken,Thompson,ken
"Ro""bert","Griesemer","gri"
`
		pprows := testStage1DeterminePostProcRows(t, []byte(data))
		expected := []postProcRow{{2,6}, {9, 13}}

		if !reflect.DeepEqual(pprows, expected) {
			log.Fatalf("TestStage1DeterminePostProcRows: got %v, want %v", pprows, expected)
		}
	})

	t.Run("multiple-quoted-CRLFs", func(t *testing.T) {
		const data = `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Rob`+"\r\n"+`ert","Griesemer","gri"
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","g`+"\r\n"+`ri"
`
		pprows := testStage1DeterminePostProcRows(t, []byte(data))
		expected := []postProcRow{{2, 6}, {12, 13}}

		if !reflect.DeepEqual(pprows, expected) {
			log.Fatalf("TestStage1DeterminePostProcRows: got %v, want %v", pprows, expected)
		}
	})

	t.Run("mixed", func(t *testing.T) {
		const data = `first_name,last_name,username
"Rob","Pi`+"\r\n"+`ke",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
"Ro""b","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","g`+"\r\n"+`ri"
`
		pprows := testStage1DeterminePostProcRows(t, []byte(data))
		expected := []postProcRow{{0, 3}, {5, 10}, {12, 13}}

		if !reflect.DeepEqual(pprows, expected) {
			log.Fatalf("TestStage1DeterminePostProcRows: got %v, want %v", pprows, expected)
		}
	})

	t.Run("huge", func(t *testing.T) {
		const header = `first_name,last_name,username` + "\n"
		const first = `"Rob","Pike",rob` + "\n"
		const second = `Ken,Thompson,ken` + "\n"
		const third = `"Robert","Griesemer","gri"` + "\n"

		data := header

		for i := 0; i < 250; i++ {
			if i % 59 == 58 {
				data += strings.ReplaceAll(first, "Pike", `Pi""ke`) + second + third
			} else if i % 97 == 96 {
				data += first + second + strings.ReplaceAll(third, "Griesemer", "Grie\r\nsemer")
			} else {
				data += first + second + third
			}
		}

		pprows := testStage1DeterminePostProcRows(t, []byte(data))
		expected := []postProcRow{{172, 176},  {288, 292}, {351, 355}, {528, 532}, {581, 585}, {704, 708}}

		if !reflect.DeepEqual(pprows, expected) {
			log.Fatalf("TestStage1DeterminePostProcRows: got %v, want %v", pprows, expected)
		}
	})

	t.Run("long-lines", func(t *testing.T) {

		data := ""

		for i := 0; i < 50; i++ {
			if i % 11 == 10 {
				data += strings.Repeat("a", 40) + `,"` + strings.Repeat("b", 20) + `""` + strings.Repeat("b", 20) + `",` + strings.Repeat("c", 40) + "\n"
			} else if i %  17 == 16 {
				data += strings.Repeat("a", 40) + "," + strings.Repeat("b", 40) + `,"` + strings.Repeat("c", 15) + "\r\n" + strings.Repeat("c", 25)+ `"` + "\n"
			} else {
				data += strings.Repeat("a", 40) + "," + strings.Repeat("b", 40) + "," + strings.Repeat("c", 40) + "\n"
			}
		}

		pprows := testStage1DeterminePostProcRows(t, []byte(data))
		expected := []postProcRow{{10, 11}, {16, 18}, {21, 22}, {32, 33}, {33, 35}, {43, 44}}

		if !reflect.DeepEqual(pprows, expected) {
			log.Fatalf("TestStage1DeterminePostProcRows: got %v, want %v", pprows, expected)
		}
	})
}

func testStage1DeterminePostProcRows(t *testing.T, buf []byte) []postProcRow {

	postProc := Stage1PreprocessBuffer(buf, uint64(','))
	simdrecords, parsingError := Stage2ParseBuffer(buf, 0xa, preprocessedSeparator, preprocessedQuote, nil)
	if parsingError {
		t.Errorf("testStage1DeterminePostProcRows: unexpected parsing error")
	}

	pprows := getPostProcRows(buf, postProc, simdrecords)

	// Sanity check: there must be either a double quote or \r\n combination to replace in  all
	for _, ppr := range pprows {
		foundAny := false
		for r := ppr.start; r < ppr.end; r++ {
			//fmt.Println(simdrecords[r-1])
			//fmt.Println(simdrecords[r])
			for c := range  simdrecords[r] {
				foundAny = foundAny || strings.Index(simdrecords[r][c], "\"\"") != -1
				foundAny = foundAny || strings.Index(simdrecords[r][c], "\r\n") != -1
			}
		}
		if !foundAny {
			t.Errorf("testStage1DeterminePostProcRows: sanity check fails: could not find any post processing to do")
		}
	}

	return pprows
}

func testStage1DynamicAllocation(t *testing.T) {

	buf, _ := ioutil.ReadFile("parking-citations-10K-postproc.csv")
	bufSingleInvoc, err := ioutil.ReadFile("parking-citations-10K-postproc.csv")
	if err != nil {
		log.Fatalln(err)
	}

	postProcSingleInvoc := make([]uint64, 0, len(buf)>>6)
	{
		input, output := stage1Input{}, stage1Output{}
		// explicitly invoke stage 1 directly with single call
		processed := stage1_preprocess_buffer(bufSingleInvoc, uint64(','), &input, &output, &postProcSingleInvoc, 0)
		if processed < uint64(len(buf)) {
			t.Errorf("testStage1DynamicAllocation: got %v, want %v", processed, len(buf))
		}
	}

	postProc := make([]uint64, 0, 3) // small allocation, make sure we dynamically grow
	postProc = Stage1PreprocessBufferEx(buf, uint64(','), &postProc)

	if !reflect.DeepEqual(postProc, postProcSingleInvoc) {
		t.Errorf("testStage1DynamicAllocation: got %v, want %v", postProc, postProcSingleInvoc)
	}
}

func TestStage1DynamicAllocation(t *testing.T) {

	t.Run("grow-postproc", func(t *testing.T) {
		testStage1DynamicAllocation(t)
	})
}