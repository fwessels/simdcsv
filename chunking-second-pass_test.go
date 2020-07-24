package simdcsv

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
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

func testParseSecondPassUnquoted(t *testing.T, f func(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int)) {

	const file = `a,bb,,ddd,eeee,,,hhhhh,,,,jjjjjj,,,,,ooooooo,,,,,,uuuuuuuu,,,,,
`
	//fmt.Println(hex.Dump([]byte(file)))

	output, _ := ParseSecondPass([]byte(file)[:64], '\n', ',', '"', f)
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

func testParseSecondPassQuoted(t *testing.T, f func(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int)) {

	const file = `A,"A",BB,,"DDD","EEEE","",,HHHHH,,,,JJJJJJ,,,,,OOOOOOO,,,,,,UUU
`
	//fmt.Println(hex.Dump([]byte(file)))

	output, _ := ParseSecondPass([]byte(file)[:64], '\n', ',', '"', f)
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
	index, line := 1, 0
	offset := uint64(0)
	input := Input{}

	b.SetBytes(int64(len(file)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		columns[0] = 0
		index = 1
		line = 0
		input.separatorMask = separatorMasks[0]
		input.delimiterMask = delimiterMasks[0]
		input.quoteMask = quoteMasks[0]
		input.quoted = uint64(0)

		parse_second_pass_test(&input, offset, &columns, &index, &rows, &line)
	}
}

func testParseSecondPassMultipleMasks(t *testing.T, f func(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int)) {

	const file = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb,ccccccccccccccccccccccccccccccc,ddddddddddddddddddddddddddddddd
`
	//fmt.Println(hex.Dump([]byte(file)))

	output, _ := ParseSecondPass([]byte(file), '\n', ',', '"', f)
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

func testParseSecondPassMultipleRows(t *testing.T, f func(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int)) {

	const file = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb,ccccccccccccccccccccccccccccccc,ddddddddddddddddddddddddddddddd
eeeeeeeeeeeeeeeeeeeeeeeeeeeeeee,fffffffffffffffffffffffffffffff,ggggggggggggggggggggggggggggggg,hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh
`
	//fmt.Println(hex.Dump([]byte(file)))

	columns, rows := ParseSecondPass([]byte(file), '\n', ',', '"', f)
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
