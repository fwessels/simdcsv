package simdcsv

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

func TestPreprocessDoubleQuotes(t *testing.T) {

	const file = `first_name,last_name,username
"Robert","Pike",rob
Kenny,Thompson,kenny
"Robert","Griesemer","gr""i"
Donald,Du""c
k,don
Dagobert,Duck,dago
`
	fmt.Println(hex.Dump([]byte(file)))

	// 00000000  66 69 72 73 74 5f 6e 61  6d 65 2c 6c 61 73 74 5f  |first_name,last_|
	// 00000010  6e 61 6d 65 2c 75 73 65  72 6e 61 6d 65 0a 22 52  |name,username."R|
	// 00000020  6f 62 65 72 74 22 2c 22  50 69 6b 65 22 2c 72 6f  |obert","Pike",ro|
	// 00000030  62 0a 4b 65 6e 6e 79 2c  54 68 6f 6d 70 73 6f 6e  |b.Kenny,Thompson|
	// 00000040  2c 6b 65 6e 6e 79 0a 22  52 6f 62 65 72 74 22 2c  |,kenny."Robert",|
	// 00000050  22 47 72 69 65 73 65 6d  65 72 22 2c 22 67 72 22  |"Griesemer","gr"|
	// 00000060  22 69 22 0a 44 6f 6e 61  6c 64 2c 44 75 22 22 63  |"i".Donald,Du""c|
	// 00000070  0a 6b 2c 64 6f 6e 0a 44  61 67 6f 62 65 72 74 2c  |.k,don.Dagobert,|
	// 00000080  44 75 63 6b 2c 64 61 67  6f 0a                    |Duck,dago.|

	preprocessed := PreprocessDoubleQuotes([]byte(file))

	fmt.Println(hex.Dump(preprocessed))

	// 00000000  66 69 72 73 74 5f 6e 61  6d 65 00 6c 61 73 74 5f  |first_name.last_|
	// 00000010  6e 61 6d 65 00 75 73 65  72 6e 61 6d 65 0a 52 6f  |name.username.Ro|
	// 00000020  62 65 72 74 00 50 69 6b  65 00 72 6f 62 0a 4b 65  |bert.Pike.rob.Ke|
	// 00000030  6e 6e 79 00 54 68 6f 6d  70 73 6f 6e 00 6b 65 6e  |nny.Thompson.ken|
	// 00000040  6e 79 0a 52 6f 62 65 72  74 00 47 72 69 65 73 65  |ny.Robert.Griese|
	// 00000050  6d 65 72 00 67 72 22 69  0a 44 6f 6e 61 6c 64 2c  |mer.gr"i.Donald,|
	// 00000060  44 75 22 63 0a 6b 00 64  6f 6e 0a 44 61 67 6f 62  |Du"c.k.don.Dagob|
	// 00000070  65 72 74 00 44 75 63 6b  00 64 61 67 6f 0a        |ert.Duck.dago.|

	lines := bytes.Split([]byte(preprocessed), []byte{PreprocessedDelimiter})
	for _, line := range lines {
		fields := bytes.Split([]byte(line), []byte{PreprocessedSeparator})
		for i, field := range fields {
			fmt.Print(string(field))
			if i < len(fields)-1 {
				fmt.Print(",")
			}
		}
		fmt.Println()
	}
}

func TestPreprocessCarriageReturns(t *testing.T) {

	// The Reader converts all \r\n sequences in its input to plain \n,
	// including in multiline field values, so that the returned data does
	// not depend on which line-ending convention an input file uses.

	const data = `first,second,third` + "\r" + "\n" +
		"aap,noot,mies" + "\r" + "\n" +
		"vuur,boom,vis" + "\r" + "\n"

	transformed := PreprocessCarriageReturns([]byte(data))

	if len(transformed) != len(data) - 3 {
		t.Errorf("TestPreprocessCarriageReturns: got: %d want: %d", len(transformed), len(data) - 3)
	}

	const crInQuotedField = `first,second,third` + "\r" + "\n" +
		`aap,"no`+ "\r" + "\n" + `ot",mies` + "\r" + "\n" +
		"vuur,boom,vis" + "\r" + "\n"

	transformed = PreprocessCarriageReturns([]byte(crInQuotedField))

	if len(transformed) != len(crInQuotedField) - 4 {
		t.Errorf("TestPreprocessCarriageReturns: got: %d want: %d", len(transformed), len(crInQuotedField) - 4)
	}
}

func testParseSecondPassUnquoted(t *testing.T, f func(input *Input, offset uint64, output *Output)) {

	const file = `a,bb,,ddd,eeee,,,hhhhh,,,,jjjjjj,,,,,ooooooo,,,,,,uuuuuuuu,,,,,
`
	//fmt.Println(hex.Dump([]byte(file)))

	output, _, _ := ParseSecondPass([]byte(file)[:64], '\n', ',', '"', f)
	expected := []uint64{0, 1, 2, 4, 5, 5, 6, 9, 10, 14, 15, 15, 16, 16, 17, 22, 23, 23, 24, 24, 25, 25, 26, 32, 33, 33, 34, 34, 35, 35, 36, 36, 37, 44, 45, 45, 46, 46, 47, 47, 48, 48, 49, 49, 50, 58, 59, 59, 60, 60, 61, 61, 62, 62, 63, 63}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("testParseSecondPassUnquoted: got: %v want: %v", output, expected)
	}
}

func TestParseSecondPassUnquoted(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		testParseSecondPassUnquoted(t, ParseSecondPassMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testParseSecondPassUnquoted(t, parse_second_pass_test)
	})
}

func testParseSecondPassQuoted(t *testing.T, f func(input *Input, offset uint64, output *Output)) {

	const file = `A,"A",BB,,"DDD","EEEE","",,HHHHH,,,,JJJJJJ,,,,,OOOOOOO,,,,,,UUU
`
	//fmt.Println(hex.Dump([]byte(file)))

	output, _, _ := ParseSecondPass([]byte(file)[:64], '\n', ',', '"', f)
	expected := []uint64{0, 1, 3, 4, 6, 8, 9, 9, 11, 14, 17, 21, 24, 24, 26, 26, 27, 32, 33, 33, 34, 34, 35, 35, 36, 42,
		43, 43, 44, 44, 45, 45, 46, 46, 47, 54, 55, 55, 56, 56, 57, 57, 58, 58, 59, 59, 60, 63}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("testParseSecondPassQuoted: got: %v want: %v", output, expected)
	}
}

func TestParseSecondPassQuoted(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		testParseSecondPassQuoted(t, ParseSecondPassMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testParseSecondPassQuoted(t, parse_second_pass_test)
	})
}

func BenchmarkParseSecondPass(b *testing.B) {

	const file = `a,bb,,ddd,eeee,,,hhhhh,,,,jjjjjj,,,,,ooooooo,,,,,,uuuuuuuu,,,,,
`
	//fmt.Println(hex.Dump([]byte(file)))

	separatorMasks := getBitMasks([]byte(file), byte(','))
	delimiterMasks := getBitMasks([]byte(file), byte('\n'))
	quoteMasks := getBitMasks([]byte(file), byte('"'))

	columns, rows := [128]uint64{}, [128]uint64{}
	columns[0] = 0
	offset := uint64(0)
	input := Input{}
	output := Output{&columns, 1, &rows, 0}

	b.SetBytes(int64(len(file)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		columns[0] = 0
		input.separatorMask = separatorMasks[0]
		input.delimiterMask = delimiterMasks[0]
		input.quoteMask = quoteMasks[0]
		input.quoted = uint64(0)
		output.index = 1
		output.line = 0

		parse_second_pass_test(&input, offset, &output)
	}
}

func testParseSecondPassMultipleMasks(t *testing.T, f func(input *Input, offset uint64, output *Output)) {

	const file = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb,ccccccccccccccccccccccccccccccc,ddddddddddddddddddddddddddddddd
`
	//fmt.Println(hex.Dump([]byte(file)))

	output, _, _ := ParseSecondPass([]byte(file), '\n', ',', '"', f)
	expected := []uint64{0, 0x1f, 0x20, 0x3f, 0x40, 0x5f, 0x60, 0x7f}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("TestParseSecondPassMultipleMasks: got: %v want: %v", output, expected)
	}
}

func TestParseSecondPassMultipleMasks(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		testParseSecondPassMultipleMasks(t, ParseSecondPassMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testParseSecondPassMultipleMasks(t, parse_second_pass_test)
	})
}

func testParseSecondPassMultipleRows(t *testing.T, f func(input *Input, offset uint64, output *Output)) {

	const file = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb,ccccccccccccccccccccccccccccccc,ddddddddddddddddddddddddddddddd
eeeeeeeeeeeeeeeeeeeeeeeeeeeeeee,fffffffffffffffffffffffffffffff,ggggggggggggggggggggggggggggggg,hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh
`
	//fmt.Println(hex.Dump([]byte(file)))

	columns, rows, _ := ParseSecondPass([]byte(file), '\n', ',', '"', f)
	expectedCols := []uint64{0, 0x1f, 0x20, 0x3f, 0x40, 0x5f, 0x60, 0x7f, 0x80, 0x9f, 0xa0, 0xbf, 0xc0, 0xdf, 0xe0, 0xff}
	expectedRows := []uint64{8, 16}

	if !reflect.DeepEqual(columns, expectedCols) {
		t.Errorf("TestParseSecondPassMultipleRows: got: %v want: %v", columns, expectedCols)
	}

	if !reflect.DeepEqual(rows, expectedRows) {
		t.Errorf("TestParseSecondPassMultipleRows: got: %v want: %v", rows, expectedRows)
	}
}

func TestParseSecondPassMultipleRows(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		testParseSecondPassMultipleRows(t, ParseSecondPassMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testParseSecondPassMultipleRows(t, parse_second_pass_test)
	})
}

// Opening quote can only start after either , or delimiter
func testBareQuoteInNonQuotedField(t *testing.T, f func(input *Input, offset uint64, output *Output)) {

	bareQuoteInNonQuotedFields := []struct {
		input    string
		expected uint64
	}{
		{` "aaaa","bbbb"`, 1},
		{`"aaaa", "bbbb"`, 8},
		{`"aaaa"
 "bbbb",`, 8},
	}

	for _, bareQuoteInNonQuotedField := range bareQuoteInNonQuotedFields {

		//r := csv.NewReader(strings.NewReader(bareQuoteInNonQuotedField.input))
		//
		//_, err := r.ReadAll()
		//if err == nil {
		//	log.Fatal("Expected error")
		//} else {
		//	fmt.Printf("%v\n", err)
		//}

		in := [64]byte{}
		copy(in[:], bareQuoteInNonQuotedField.input)
		_, _, errorOffset := ParseSecondPass(in[:], '\n', ',', '"', f)

		if errorOffset != bareQuoteInNonQuotedField.expected {
			t.Errorf("testBareQuoteInNonQuotedField: got: %d want: %d", errorOffset, bareQuoteInNonQuotedField.expected)
		}
	}
}

// Opening quote can only start after either , or delimiter
func TestBareQuoteInNonQuotedField(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		testBareQuoteInNonQuotedField(t, ParseSecondPassMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testBareQuoteInNonQuotedField(t, parse_second_pass_test)
	})
}

func testExtraneousOrMissingQuoteInQuotedField(t *testing.T, f func(input *Input, offset uint64, output *Output)) {

	extraneousOrMissingQuoteInQuotedFields := []struct {
		input    string
		expected uint64
	}{
		{`"aaaa" ,bbbb`, 7},
		{`"aaaa" `+`
"bbbb"`, 7},
		{`"aaaa","bbbb" `, 14},
	}

	for _, extraneousOrMissingQuoteInQuotedField := range extraneousOrMissingQuoteInQuotedFields {

		//r := csv.NewReader(strings.NewReader(extraneousOrMissingQuoteInQuotedField.input))
		//
		//_, err := r.ReadAll()
		//if err == nil {
		//	log.Fatal("Expected error")
		//} else {
		//	fmt.Printf("%v\n", err)
		//}

		in := [64]byte{}
		copy(in[:], extraneousOrMissingQuoteInQuotedField.input)

		// TODO: Fix this hack: make sure we always end with a delimiter
		if in[len(extraneousOrMissingQuoteInQuotedField.input)- 1] != '\n' {
			in[len(extraneousOrMissingQuoteInQuotedField.input)] = '\n'
		}

		_, _, errorOffset := ParseSecondPass(in[:], '\n', ',', '"', f)

		if errorOffset != extraneousOrMissingQuoteInQuotedField.expected {
			t.Errorf("TestExtraneousOrMissingQuoteInQuotedField: got: %d want: %d", errorOffset, extraneousOrMissingQuoteInQuotedField.expected)
		}
	}
}

// Closing quote needs to be followed immediate by either a , or delimiter
func TestExtraneousOrMissingQuoteInQuotedField(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		testExtraneousOrMissingQuoteInQuotedField(t, ParseSecondPassMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testExtraneousOrMissingQuoteInQuotedField(t, parse_second_pass_test)
	})
}

func TestParseBlockSecondPass(t *testing.T) {

	const vector = `1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,"NO EVIDENCE,OF REG",50,99999,99999
1103700150,2015-12-21T00:00:00,1435,,,CA,201512,,GMC,VN,WH,525 S MAIN ST,1C51,1,4000A1,NO EVIDENCE OF REG,50,99999,99999
1104803000,2015-12-21T00:00:00,2055,,,CA,201503,,NISS,PA,BK,200 WORLD WAY,2R2,2,8939,WHITE CURB,58,6439997.9,1802686.4
1104820732,2015-12-26T00:00:00,1515,,,CA,,,ACUR,PA,WH,100 WORLD WAY,2F11,2,000,17104h,,6440041.1,1802686.2
1105461453,2015-09-15T00:00:00,115,,,CA,200316,,CHEV,PA,BK,GEORGIA ST/OLYMPIC,1FB70,1,8069A,NO STOPPING/STANDING,93,99999,99999
1106226590,2015-09-15T00:00:00,19,,,CA,201507,,CHEV,VN,GY,SAN PEDRO S/O BOYD,1A35W,1,4000A1,NO EVIDENCE OF REG,50,99999,99999
1106500452,2015-12-17T00:00:00,1710,,,CA,201605,,MAZD,PA,BL,SUNSET/ALVARADO,00217,1,8070,"PARK IN GRID LOCK ZN",163,99999,99999
1106500463,2015-12-17T00:00:00,1710,,,CA,201602,,TOYO,PA,BK,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999
1106506402,2015-12-22T00:00:00,945,,,CA,201605,,CHEV,PA,BR,721 S WESTLAKE,2A75,1,8069AA,NO STOP/STAND AM,93,99999,99999
1106506413,2015-12-22T00:00:00,1100,,,CA,201701,,NISS,PA,SI,1159 HUNTLEY DR,2A75,1,8069AA,NO STOP/STAND AM,93,99999,99999
`
	buf := []byte(strings.Repeat(vector , 1))
	input := Input{}
	rows := make([]uint64, 20)
	columns := make([]uint64, 20 * len(rows))
	output := OutputAsm{unsafe.Pointer(&columns[0]), 1, unsafe.Pointer(&rows[0]), 0}

	fmt.Println(len(buf))

	parse_block_second_pass(buf, '\n', ',', '"', &input, 0, &output)

	fmt.Println(output.index)
	fmt.Println(output.line)

	lStart := 0
	for l := 0; l < output.line; l++ {
		fmt.Println("line", l)
		for i := lStart; i < int(rows[l]); i += 2 {
			if columns[i] == ^uint64(0) || columns[i+1] == ^uint64(0) {
				break
			}
			fmt.Printf("    %016x - %016x: %s\n", columns[i], columns[i+1], string(buf[columns[i]:columns[i+1]]))
		}
		lStart = int(rows[l])
	}
}

func BenchmarkParseBlockSecondPass(b *testing.B) {

	const file = `aaaa,aaaa,aaaa,aaaa,aaaa,aaaaaa,bbbb,bbbb,bbbb,bbbb,bbbb,bbbbbb,cccc,cccc,cccc,cccc,cccc,cccccc,dddd,dddd,dddd,dddd,dddd,dddddd
eeee,eeee,eeee,eeee,eeee,eeeeee,ffff,ffff,ffff,ffff,ffff,ffffff,gggg,gggg,gggg,gggg,gggg,gggggg,hhhh,hhhh,hhhh,hhhh,hhhh,hhhhhh
`

	buf := []byte(strings.Repeat(file , 1000))
	input := Input{}
	columns := make([]uint64, 128000)
	rows := make([]uint64, 128000)
	output := OutputAsm{unsafe.Pointer(&columns[0]), 1, unsafe.Pointer(&rows[0]), 0}

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		columns[0] = 0
		output.index = 1
		output.line = 0

		parse_block_second_pass(buf, '\n', ',', '"', &input, 0, &output)
	}
}
