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

func TestParseSecondPass(t *testing.T) {

	const file = `a,bb,,ddd,eeee,,,hhhhh,,,,jjjjjj,,,,,ooooooo,,,,,,uuuuuuuu,,,,,
`
	fmt.Println(hex.Dump([]byte(file)))

	ParseSecondPass([]byte(file)[:64], '\n', ',', '"')
}

func TestParseSecondPassQuoted(t *testing.T) {

	const file = `A,"A",BB,,"DDD","EEEE","",,HHHHH,,,,JJJJJJ,,,,,OOOOOOO,,,,,,UUU
`
	fmt.Println(hex.Dump([]byte(file)))

	ParseSecondPass([]byte(file)[:64], '\n', ',', '"')
}

func BenchmarkParseSecondPass(b *testing.B) {

	const file = `a,bb,,ddd,eeee,,,hhhhh,,,,jjjjjj,,,,,ooooooo,,,,,,uuuuuuuu,,,,,
`
	//fmt.Println(hex.Dump([]byte(file)))

	separatorMasks := getBitMasks([]byte(file), byte(','))
	delimiterMasks := getBitMasks([]byte(file), byte('\n'))
	quoteMasks := getBitMasks([]byte(file), byte('"'))

	output := [128]uint64{}
	quoted := uint64(0)
	output[0] = 0
	index := 1

	b.SetBytes(int64(len(file)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		quoted = uint64(0)
		output[0] = 0
		index = 1

		parse_second_pass(separatorMasks[0], delimiterMasks[0], quoteMasks[0], &output, &index, &quoted)
	}
}

func TestParseSecondPassMultipleMasks(t *testing.T) {

	const file = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb,ccccccccccccccccccccccccccccccc,ddddddddddddddddddddddddddddddd
`
	fmt.Println(hex.Dump([]byte(file)))

	output, _ := ParseSecondPass([]byte(file), '\n', ',', '"')
	expected := []uint64{0, 0x1f, 0x20, 0x3f, 0x40, 0x5f, 0x60, 0x7f}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("TestParseSecondPassMultipleMasks: got: %v want: %v", output, expected)
	}
}

func TestParseSecondPassMultipleRows(t *testing.T) {

	const file = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa,bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb,ccccccccccccccccccccccccccccccc,ddddddddddddddddddddddddddddddd
eeeeeeeeeeeeeeeeeeeeeeeeeeeeeee,fffffffffffffffffffffffffffffff,ggggggggggggggggggggggggggggggg,hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh
`
	fmt.Println(hex.Dump([]byte(file)))

	columns, rows := ParseSecondPass([]byte(file), '\n', ',', '"')
	expectedCols := []uint64{0, 0x1f, 0x20, 0x3f, 0x40, 0x5f, 0x60, 0x7f, 0x80, 0x9f, 0xa0, 0xbf, 0xc0, 0xdf, 0xe0, 0xff}
	expectedRows := []uint64{8, 16}

	if !reflect.DeepEqual(columns, expectedCols) {
		t.Errorf("TestParseSecondPassMultipleRows: got: %v want: %v", columns, expectedCols)
	}

	if !reflect.DeepEqual(rows, expectedRows) {
		t.Errorf("TestParseSecondPassMultipleRows: got: %v want: %v", rows, expectedRows)
	}

	fmt.Println(expectedRows)
}
