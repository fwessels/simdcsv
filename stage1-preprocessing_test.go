package simdcsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"math/bits"
	"reflect"
	"strings"
	"testing"
)

func TestStage1PreprocessMasks(t *testing.T) {

	t.Run("go", func(t *testing.T) {
		testStage1PreprocessMasksFunc(t, preprocessMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testStage1PreprocessMasksFunc(t, stage1_preprocess_test)
	})
}

func testStage1PreprocessMasksFunc(t *testing.T, f func(input *stage1Input, output *stage1Output)) {

	t.Run("simple", func(t *testing.T) {

		const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            `

		result := testStage1PreprocessMasks(t, []byte(data), f)

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
			t.Errorf("testStage1PreprocessMasksFunc: got %v, want %v", result, expected)
		}
	})

	t.Run("double-quotes-at-end-of-mask", func(t *testing.T) {

		const data = `Robe,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            
first_name,last_name,username1234`

		result := testStage1PreprocessMasks(t, []byte(data), f)

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
			t.Errorf("testStage1PreprocessMasksFunc: got %v, want %v", result, expected)
		}
	})

	t.Run("double-quotes-split-over-masks", func(t *testing.T) {

		const data = `Rober,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr""i"                            
first_name,last_name,username123`

		result := testStage1PreprocessMasks(t, []byte(data), f)

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
			t.Errorf("testStage1PreprocessMasksFunc: got %v, want %v", result, expected)
		}
	})

	t.Run("single-carriage-return", func(t *testing.T) {

		const data = `Rober,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr`+"\r"+`i"                             
first_name,last_name,username123`

		result := testStage1PreprocessMasks(t, []byte(data), f)

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
			t.Errorf("testStage1PreprocessMasksFunc: got %v, want %v", result, expected)
		}
	})

	t.Run("carriage-return-new-line-split-over-masks", func(t *testing.T) {

		const data = `Rober,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Griesemer","gr`+"\r\n"+`i"                            
first_name,last_name,username123`

		result := testStage1PreprocessMasks(t, []byte(data), f)

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
			t.Errorf("testStage1PreprocessMasksFunc: got %v, want %v", result, expected)
		}
	})

	t.Run("bare-cr-2", func(t *testing.T) {

		const data = "a,b\rc,d\r\n                                                                                                                       "

		result := testStage1PreprocessMasks(t, []byte(data), f)

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
			t.Errorf("testStage1PreprocessMasksFunc: got %v, want %v", result, expected)
		}
	})

	t.Run("bare-cr-at-end-of-mask", func(t *testing.T) {

		const data = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\rc,d\r\n                                                            "

		result := testStage1PreprocessMasks(t, []byte(data), f)

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
			t.Errorf("testStage1PreprocessMasksFunc: got %v, want %v", result, expected)
		}
	})

	t.Run("bare-cr-split-over-masks", func(t *testing.T) {

		const data = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\rc,d\r\n                                                           "

		result := testStage1PreprocessMasks(t, []byte(data), f)

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
			t.Errorf("testStage1PreprocessMasksFunc: got %v, want %v", result, expected)
		}
	})
}

// Test whether the last two YMM words are correctly masked out (beyond end of buffer)
func TestStage1PartialLoad(t *testing.T) {

	const data = `,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`

	for i := 1; i <= 128; i++ {
		buf := []byte(data[:i])

		masks := allocMasks(buf)
		postProc := make([]uint64, ((len(buf)>>6)+1))
		input, output := stage1Input{}, stage1Output{}

		processed := stage1_preprocess_buffer(buf, ',', &input, &output, &postProc, 0, masks)

		out := ""
		if processed <= 64 {
			out = fmt.Sprintf("%064b", bits.Reverse64(masks[1]))
		} else {
			out = fmt.Sprintf("%064b%064b", bits.Reverse64(masks[1]), bits.Reverse64(masks[3+1]))
		}

		expected := strings.Repeat("1", i) + strings.Repeat("0", int(processed)-i)

		//fmt.Println(out)
		//fmt.Println(expected)

		if out != expected {
			t.Errorf("TestStage1PartialLoad: got %v, want %v", out, expected)
		}
	}
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

func getBitMasks(buf []byte, cmp byte) (masks []uint64) {

	if len(buf)%64 != 0 {
		panic("Input strings should be a multipe of 64")
	}

	masks = make([]uint64, 0)

	for i := 0; i < len(buf); i += 64 {
		mask := uint64(0)
		for b, c := range buf[i : i+64] {
			if c == cmp {
				mask = mask | (1 << b)
			}
		}
		masks = append(masks, mask)
	}
	return
}

func testStage1PreprocessMasks(t *testing.T, data []byte, f func(input *stage1Input, output *stage1Output)) string {

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

	masks, postProc := Stage1PreprocessBuffer(buf, uint64(','))

	out := bytes.NewBufferString("")
	fmt.Fprintf(out,"%064b%064b\n", bits.Reverse64(masks[0]), bits.Reverse64(masks[3+0]))
	fmt.Fprintf(out,"%064b%064b\n", bits.Reverse64(masks[1]), bits.Reverse64(masks[3+1]))
	fmt.Fprintf(out,"%064b%064b\n", bits.Reverse64(masks[2]), bits.Reverse64(masks[3+2]))

	//fmt.Println(out.String())

	const expected =
`00000000000000000000000000000100000000000000000001100000000001000000000100000000000000000000000000000000000000000000000000000000
00000000001000000000100000000000000000100000010000000000100000000100000000000000100000000000000000000000000000000000000010000000
00000000000000000000000000000000000000010000100000000000010000001000000010000001010000000000000000000000000000000000000101000001
`

	if out.String() != expected {
		t.Errorf("TestStage1MaskingOut: got %v, want %v", out.String(), expected)
	}

	simdrecords, parsingError := Stage2ParseBuffer(buf, masks, 0xa, nil)
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

func BenchmarkStage1PreprocessingMasks( b *testing.B) {

	const data = `first_name,last_name,username
RRobertt,"Pi,e",rob` + "\r\n" + `Kenny,"ho` + "\r\n" + `so",kenny
"Robert","Grie                           semer","gr""i"
`

	const repeat = 500
	buf := make([]byte, 128*repeat)
	buf = bytes.Repeat([]byte(data), repeat)

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	postProc := make([]uint64, 0, len(buf)>>6)
	masks := allocMasks(buf)

	for i := 0; i < b.N; i++ {

		postProc = postProc[:0]
		Stage1PreprocessBufferEx(buf, uint64(','), &masks, &postProc)
	}
}

func TestTrailingCRs(t *testing.T) {

	for cnt := 1; cnt <= 1500; cnt++ {

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

	masks, postProc := Stage1PreprocessBuffer(buf, uint64(','))
	simdrecords, parsingError := Stage2ParseBuffer(buf, masks, 0xa, nil)
	if parsingError {
		t.Errorf("testStage1DeterminePostProcRows: unexpected parsing error")
	}

	pprows := getPostProcRows(buf, postProc, simdrecords)

	// Sanity check: there must be either a double quote or \r\n combination to replace in  all
	for _, ppr := range pprows {
		foundAny := false
		for r := ppr.start; r < ppr.end; r++ {
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
	masks := allocMasks(buf)
	{
		input, output := stage1Input{}, stage1Output{}
		// explicitly invoke stage 1 directly with single call
		processed := stage1_preprocess_buffer(bufSingleInvoc, uint64(','), &input, &output, &postProcSingleInvoc, 0, masks)
		if processed < uint64(len(buf)) {
			t.Errorf("testStage1DynamicAllocation: got %v, want %v", processed, len(buf))
		}
	}

	postProc := make([]uint64, 0, 3) // small allocation, make sure we dynamically grow
	masks, postProc = Stage1PreprocessBufferEx(buf, uint64(','), nil, &postProc)

	if !reflect.DeepEqual(postProc, postProcSingleInvoc) {
		t.Errorf("testStage1DynamicAllocation: got %v, want %v", postProc, postProcSingleInvoc)
	}
}

func TestStage1DynamicAllocation(t *testing.T) {

	t.Run("grow-postproc", func(t *testing.T) {
		testStage1DynamicAllocation(t)
	})
}

func TestStage1MasksBounds(t *testing.T) {

	buf, err := ioutil.ReadFile("testdata/parking-citations.csv")
	if err != nil {
		panic(err)
	}

	postProc := make([]uint64, 0, (len(buf)>>6)+1)
	masks := allocMasks(buf)

	{
		input, output := stage1Input{}, stage1Output{}

		processed, masksWritten := stage1_preprocess_buffer(buf, uint64(','), &input, &output, &postProc, 0, masks)

		if processed/64 != masksWritten/3 {
			panic("Sanity check fails: processed/64 != masksWritten/3")
		}

		if processed < uint64(len(buf)) {
			panic("Sanity check fails: processed < uint64(len(buf))")
		}
	}

	postProcLoop := make([]uint64, 0, (len(buf)>>6)+1)
	masksLoop := make([]uint64, 10*3)
	masksIndex := 0

	processed, masksWritten := uint64(0), uint64(0)
	inputStage1 := stage1Input{}
	for {
		outputStage1 := stage1Output{}

		index := processed
		processed, masksWritten = stage1_preprocess_buffer(buf, uint64(','), &inputStage1, &outputStage1, &postProcLoop, index, masksLoop)

		if (processed-index)/64 != masksWritten/3 {
			panic("Sanity check fails: (processed-index)/64 != masksWritten/3")
		}

		if processed >= uint64(len(buf)) {
			break
		}

		if !reflect.DeepEqual(masksLoop[:masksWritten], masks[masksIndex:masksIndex+int(masksWritten)]) {
			t.Errorf("TestStage1MasksBounds: got %v, want %v", masksLoop[:masksWritten], masks[masksIndex:masksIndex+int(masksWritten)])
		}

		masksIndex += int(masksWritten)
	}
}

func TestSimdCsvStreaming(t *testing.T) {

	buf, err := ioutil.ReadFile("testdata/parking-citations-10K.csv")
	if err != nil {
		panic(err)
	}

	postProcStream := make([]uint64, 0, ((len(buf)>>6)+1)*2)

	quoted := uint64(0)

	const chunkSize = 1024*30
	chunks := make([][]byte, 0, 100)
	masks := make([][]uint64, 0, 100)
	headers, trailers := make([]uint64, 0, 100), make([]uint64, 0, 100)
	splitRows := make([][]byte, 0, 100)
	splitRow := make([]byte, 0, 256)

	for offset := 0; offset < len(buf); offset += chunkSize {
		var  chunk []byte
		lastChunk := offset+chunkSize > len(buf)
		if lastChunk {
			chunk = buf[offset:]
		} else {
			chunk = buf[offset:offset+chunkSize]
		}

		for len(chunk) > 0 {
			masksStream := make([]uint64, 100000*3)
			masksStream, postProcStream, quoted = Stage1PreprocessBufferEx(chunk, ',', quoted, &masksStream, &postProcStream)

			header, trailer := uint64(0), uint64(0)

			next := len(masksStream) / 3 * 64
			if next < len(chunk) {
				chunks = append(chunks, chunk[:next])
				chunk = chunk[next:]
			} else {
				if offset > 0 {
					for index := 0; index < len(masksStream); index += 3 {
						hr  := bits.TrailingZeros64(masksStream[index])
						header += uint64(hr)
						if hr < 64 {
							for {
								if (masksStream[index] >> hr) & 1 == 1 {
									hr++
									header++
								} else {
									break
								}
							}
							break
						}
					}
				}

				if !lastChunk {
					for index := 3; index < len(masksStream); index += 3 {
						tr  := bits.LeadingZeros64(masksStream[len(masksStream)-index])
						trailer += uint64(tr)
						if tr < 64 {
							break
						}
					}
				}

				splitRow = append(splitRow, chunk[:header]...)

				chunks = append(chunks, chunk/*[header:len(chunk)-int(trailer)]*/)
				chunk = chunk[:0]
			}

			headers = append(headers, header)
			trailers = append(trailers, trailer)
			masks = append(masks, masksStream)

			splitRows = append(splitRows, splitRow)

			if offset+chunkSize > len(buf)  {
				splitRow = buf[len(buf)-int(trailer):]
			} else {
				splitRow = buf[offset+chunkSize-int(trailer):offset+chunkSize]
			}
		}
	}

	rows := make([]uint64, 100000*30)
	columns := make([]string, len(rows)*20)

	inputStage2, outputStage2 := NewInput(), OutputAsm{}

	simdrecords := make([][]string, 0, 1024)
	line := 0

	for i, chunk := range chunks {
		outputStage2.strData = headers[i] & 0x3f // reinit strData for every chunk (fields do not span chunks)

		skip := headers[i] >> 6
		shift := headers[i] & 0x3f

		masks[i][skip*3+0] &= ^uint64((1 << shift)-1)
		masks[i][skip*3+1] &= ^uint64((1 << shift)-1)
		masks[i][skip*3+2] &= ^uint64((1 << shift)-1)

		skipTz := (trailers[i] >> 6) + 1
		shiftTz := trailers[i] & 0x3f

		masks[i][len(masks[i])-int(skipTz)*3+0] &= uint64((1 << (63-shiftTz))-1)
		masks[i][len(masks[i])-int(skipTz)*3+1] &= uint64((1 << (63-shiftTz))-1)
		masks[i][len(masks[i])-int(skipTz)*3+2] &= uint64((1 << (63-shiftTz))-1)

		Stage2ParseBufferExStreaming(chunk[skip*0x40:len(chunk)-int(trailers[i])], masks[i][skip*3:], '\n', &inputStage2, &outputStage2, &rows, &columns)

		for ; line < outputStage2.line; line += 2 {
			simdrecords = append(simdrecords, columns[rows[line]:rows[line]+rows[line+1]])
		}

		if i < len(splitRows)-1 {
			records := EncodingCsv(splitRows[i+1])
			simdrecords = append(simdrecords, records...)
		}
	}

	columns = columns[:(outputStage2.index)/2]
	rows = rows[:outputStage2.line]

	records := EncodingCsv(buf)


	for i := range records {
		if !reflect.DeepEqual(simdrecords[i], records[i]) {
			fmt.Println(i)
			fmt.Printf("TestSimdCsvStreaming: got %v, want %v\n", simdrecords[i], records[i])
		}
	}

}

func TestStage1MasksLoop(t *testing.T) {

	buf, err := ioutil.ReadFile("testdata/parking-citations-100K.csv")
	if err != nil {
		panic(err)
	}

	postProcLoop := make([]uint64, 0, ((len(buf)>>6)+1)*2)
	masksLoop := make([]uint64, 10000*3)

	processed, masksWritten := uint64(0), uint64(0)
	inputStage1 := stage1Input{}
	outputStage1 := stage1Output{}

	rows := make([]uint64, 100000*30)
	columns := make([]string, len(rows)*20)

	inputStage2, outputStage2 := NewInput(), OutputAsm{}

	for {
		index := processed
		processed, masksWritten = stage1_preprocess_buffer(buf, uint64(','), &inputStage1, &outputStage1, &postProcLoop, index, masksLoop)

		if (processed-index)/64 != masksWritten/3 {
			panic("Sanity check fails: (processed-index)/64 != masksWritten/3")
		}

		/*processed =*/ stage2_parse_masks(buf, masksLoop[:masksWritten], rows, columns, ',', &inputStage2, index, &outputStage2)

		if processed >= uint64(len(buf)) {
			break
		}
	}

	columns = columns[:(outputStage2.index)/2]
	rows = rows[:outputStage2.line]

	simdrecords := make([][]string, 0, 1024)

	for i := 0; i < len(rows); i += 2 {
		simdrecords = append(simdrecords, columns[rows[i]:rows[i]+rows[i+1]])
	}

	records := EncodingCsv(buf)

	if !reflect.DeepEqual(simdrecords, records) {
		t.Errorf("TestStage1MasksLoop: got %v, want %v", simdrecords, records)
	}
}
