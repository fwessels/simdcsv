package simdcsv

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestChunkTwoPass(t *testing.T) {

	testCases := []struct {
		input           string
		quotes          int
		posDelimEven    int
		posDelimOdd     int
		lastCharIsQuote bool
	}{
		{`col1,col2,col3
1,2,3
4,5,6`, 0, 14, -1, false},
		{`col1,col2,col3
1,"2",3
4,5,6`, 2, 14, -1, false},
		{`col1,col2,col3
"1,2,3
4",5,6`, 2, 14, 21, false},
		{`1,2,3
4",5,6
7,8,9`, 1, 5, 12, false},
		{`1,2,3",
4,5,6
7,8,9",
10,11,12`, 2, 21, 7, false},
		{`1,2,3",
4,""5,6
7,8,9",
10,11,12`, 2, 23, 7, false},
		{`1,2,3",
4,5,6
7,8,9",
10,11,12
13,14,15
16,17,18`, 2, 21, 7, false},
		{`1,2,3",
4,"5",6
7,8,9",
10,11,12
13,14,15
16,17,18`, 4, 23, 7, false},
		{`1,2,3",
4,"5",6
7,8,9",
10,11,12
13,14,15
16,17,18"`, 4, 23, 7, true},
	}

	for i, tc := range testCases {
		ci := ChunkTwoPass([]byte(tc.input))

		// fmt.Println(ci.quotes, ci.positionDelimiterEven, ci.positionDelimiterOdd, ci.lastCharIsQuote)

		if ci.quotes != tc.quotes {
			t.Errorf("TestChunkTwoPass(%d): got: %d want: %d", i, ci.quotes, tc.quotes)
		}

		if ci.positionDelimiterEven != tc.posDelimEven {
			t.Errorf("TestChunkTwoPass(%d): got: %d want: %d", i, ci.positionDelimiterEven, tc.posDelimEven)
		}

		if ci.positionDelimiterOdd != tc.posDelimOdd {
			t.Errorf("TestChunkTwoPass(%d): got: %d want: %d", i, ci.positionDelimiterOdd, tc.posDelimOdd)
		}

		if ci.lastCharIsQuote != tc.lastCharIsQuote {
			t.Errorf("TestChunkTwoPass(%d): got: %v want: %v", i, ci.lastCharIsQuote, tc.lastCharIsQuote)
		}
	}
}

func testTwoPassChain(t *testing.T, filename string, chunkSize int) {

	sourceOfTruth, _ /*lines*/, _ /*maxLineLength*/ := memoryTrackingCsvParser(filename, int64(chunkSize), false)

	csv, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	chunkInfos := make([]chunkInfo, 0, 100)

	for i := 0; i < len(csv); i += chunkSize {
		end := i + chunkSize
		if end > len(csv) {
			end = len(csv)
		}

		ci := ChunkTwoPass(csv[i:end])
		chunkInfos = append(chunkInfos, ci)
	}

	if len(chunkInfos) != len(sourceOfTruth) {
		t.Errorf("TestChunkTwoPass(sizes differ): got: %v want: %v", len(chunkInfos), len(sourceOfTruth))
		return
	}

	widowSizes := GetWidowSizes(chunkInfos)

	for i, widowSize := range widowSizes {
		if widowSize != int(sourceOfTruth[i].widowSize) {
			t.Errorf("TestChunkTwoPass[%d]: got: %v want: %v", i, widowSize, sourceOfTruth[i].widowSize)
		}
	}
}

func TestTwoPassChain(t *testing.T) {
	testTwoPassChain(t, "test-data/Emails.csv", 256*1024)
}

func TestLastCharIsQuote(t *testing.T) {

	const file = `first_name,last_name,username
"Robert","Pike",rob
Kenny,Thompson,kenny
"Robert","Griesemer","gr""i"
Donald,Du"
"ck,don
Dagobert,Duck,dago
`
	fmt.Println(hex.Dump([]byte(file)))

	// 00000000  66 69 72 73 74 5f 6e 61  6d 65 2c 6c 61 73 74 5f  |first_name,last_|
	// 00000010  6e 61 6d 65 2c 75 73 65  72 6e 61 6d 65 0a 22 52  |name,username."R|
	// 00000020  6f 62 65 72 74 22 2c 22  50 69 6b 65 22 2c 72 6f  |obert","Pike",ro|
	// 00000030  62 0a 4b 65 6e 6e 79 2c  54 68 6f 6d 70 73 6f 6e  |b.Kenny,Thompson|
	// 00000040  2c 6b 65 6e 6e 79 0a 22  52 6f 62 65 72 74 22 2c  |,kenny."Robert",|
	// 00000050  22 47 72 69 65 73 65 6d  65 72 22 2c 22 67 72 22  |"Griesemer","gr"|
	// 00000060  22 69 22 0a 44 6f 6e 61  6c 64 2c 44 75 22 0a 22  |"i".Donald,Du"."|
	// 00000070  63 6b 2c 64 6f 6e 0a 44  61 67 6f 62 65 72 74 2c  |ck,don.Dagobert,|
	// 00000080  44 75 63 6b 2c 64 61 67  6f 0a                    |Duck,dago.|

	chunkInfos := make([]chunkInfo, 0, 100)

	chunkInfos = append(chunkInfos, ChunkTwoPass([]byte(file)[0:0x60]))
	chunkInfos = append(chunkInfos, ChunkTwoPass([]byte(file)[0x60:]))

	widowSizes := GetWidowSizes(chunkInfos)

	expected := []int{0, 4}

	if !reflect.DeepEqual(widowSizes, expected) {
		t.Errorf("TestLastCharIsQuote: got: %v want: %v", widowSizes, expected)
	}

	//
	// Escaped qoute at last two positions of chunk
	//
	fileTrunc := append([]byte(file)[0:0x5e], []byte(file)[0x5f:]...)

	if !(fileTrunc[0x5d] != '"' && fileTrunc[0x5e] == '"' && fileTrunc[0x5f] == '"'  && fileTrunc[0x60] != '"') {
		panic("Unexpected situation")
	}

	chunkInfos = chunkInfos[:0]
	chunkInfos = append(chunkInfos, ChunkTwoPass(fileTrunc[0:0x60]))
	chunkInfos = append(chunkInfos, ChunkTwoPass(fileTrunc[0x60:]))

	widowSizes = GetWidowSizes(chunkInfos)
	expected = []int{0, 3}

	if !reflect.DeepEqual(widowSizes, expected) {
		t.Errorf("TestLastCharIsQuote: got: %v want: %v", widowSizes, expected)
	}

	//
	// Escaped qoute at first two positions of chunk
	//
	fileExtended := append([]byte(file)[0:0x5f], 'r')
	fileExtended = append(fileExtended, []byte(file)[0x5f:]...)
	fmt.Println(hex.Dump(fileExtended))

	if !(fileExtended[0x5f] != '"' && fileExtended[0x60] == '"' && fileExtended[0x61] == '"'  && fileExtended[0x62] != '"') {
		panic("Unexpected situation")
	}

	chunkInfos = chunkInfos[:0]
	chunkInfos = append(chunkInfos, ChunkTwoPass(fileExtended[0:0x60]))
	chunkInfos = append(chunkInfos, ChunkTwoPass(fileExtended[0x60:]))

	widowSizes = GetWidowSizes(chunkInfos)
	expected = []int{0, 5}

	if !reflect.DeepEqual(widowSizes, expected) {
		t.Errorf("TestLastCharIsQuote: got: %v want: %v", widowSizes, expected)
	}
	
	//
	// Both escaped qoutes at last two positions of first chunk
	// and first two positions of second chunk
	//
	chunkInfos = chunkInfos[:0]
	chunkInfos = append(chunkInfos, ChunkTwoPass(fileTrunc[0:0x60]))
	chunkInfos = append(chunkInfos, ChunkTwoPass(fileExtended[0x60:]))

	fmt.Print(hex.Dump(fileTrunc[:0x60]))
	fmt.Println(hex.Dump(fileExtended[0x60:]))

	widowSizes = GetWidowSizes(chunkInfos)
	expected = []int{0, 5}

	if !reflect.DeepEqual(widowSizes, expected) {
		t.Errorf("TestLastCharIsQuote: got: %v want: %v", widowSizes, expected)
	}
}
