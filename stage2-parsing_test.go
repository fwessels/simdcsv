package simdcsv

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"io/ioutil"
	"log"
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
	expected := []uint64{0, 1, 2, 2, 5, 0, 6, 3, 10, 4, 15, 0, 16, 0, 17, 5, 23, 0, 24, 0, 25, 0, 26, 6, 33, 0, 34, 0, 35, 0, 36, 0, 37, 7, 45, 0, 46, 0, 47, 0, 48, 0, 49, 0, 50, 8, 59, 0, 60, 0, 61, 0, 62, 0, 63, 0}
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
	expected := []uint64{0, 1, 3, 1, 6, 2, 9, 0, 11, 3, 17, 4, 24, 0, 26, 0, 27, 5, 33, 0, 34, 0, 35, 0, 36, 6, 43, 0, 44, 0, 45, 0, 46, 0, 47, 7, 55, 0, 56, 0, 57, 0, 58, 0, 59, 0, 60, 3}

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
	output := Output{&columns, 1, &rows, 0, 0, 0, 128}

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
	expected := []uint64{0, 0x1f, 0x20, 0x1f, 0x40, 0x1f, 0x60, 0x1f}

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
iiiiiiiiiiiiiiiiiiiiiiiiiiiiiii,jjjjjjjjjjjjjjjjjjjjjjjjjjjjjjj,kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk
lllllllllllllllllllllllllllllll
`
	//fmt.Println(hex.Dump([]byte(file)))

	columns, rows, _ := ParseSecondPass([]byte(file), '\n', ',', '"', f)
	expectedCols := []uint64{0, 0x1f, 0x20, 0x1f, 0x40, 0x1f, 0x60, 0x1f, 0x80, 0x1f, 0xa0, 0x1f, 0xc0, 0x1f, 0xe0, 0x1f, 0x100, 0x1f, 0x120, 0x1f, 0x140, 0x1f, 0x160, 0x1f}
	expectedRows := []uint64{  0, 4, 128,
		                      64, 4, 124,
		                     128, 3, 120,
		                     176, 1, 117}

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
	input := Input{base: uint64(uintptr(unsafe.Pointer(&buf[0])))}
	rows := make([][]string, 20)
	columns := make([]string, 20 * len(rows))
	output := OutputAsm{unsafe.Pointer(&columns[0]), 1, unsafe.Pointer(&rows[0]), 0, uint64(uintptr(unsafe.Pointer(&columns[0]))), 0, uint64(cap(columns))}

	parse_block_second_pass(buf, '\n', ',', '"', &input, 0, &output)

	rows = rows[:output.line/3]
	simdrecords := rows // make([][]string, 0)


	r := csv.NewReader(bytes.NewReader(buf))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(simdrecords, records) {
		t.Errorf("TestParseBlockSecondPass: got %v, want %v", simdrecords, records)
	}
}

func BenchmarkParseBlockSecondPass(b *testing.B) {

	buf, err := ioutil.ReadFile("parking-citations-10K.csv")
	if err != nil {
		log.Fatalln(err)
	}

	input := Input{}
	rows := make([][]string, 10000 + 10)
	columns := make([]string, len(rows)*20)

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		output := OutputAsm{unsafe.Pointer(&columns[0]), 1, unsafe.Pointer(&rows[0]), 0, uint64(uintptr(unsafe.Pointer(&columns[0]))), 0, uint64(cap(columns))}

		parse_block_second_pass(buf, '\n', ',', '"', &input, 0, &output)
	}
}

func BenchmarkParseBlockSecondPass2(b *testing.B) {

	buf, err := ioutil.ReadFile("parking-citations-10K.csv")
	if err != nil {
		log.Fatalln(err)
	}

	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		r := csv.NewReader(bytes.NewReader(buf))
		_, err := r.ReadAll()
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
}