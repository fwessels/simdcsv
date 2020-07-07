package simdcsv

import (
	_ "fmt"
	"io/ioutil"
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

func TestTwoPassChain(t *testing.T) {

	sourceOfTruth, _ /*lines*/, _ /*maxLineLength*/ := memoryTrackingCsvParser("test-data/Emails.csv", 1024*1024, false)

	csv, err := ioutil.ReadFile("test-data/Emails.csv")
	if err != nil {
		panic(err)
	}

	chunkInfos := make([]chunkInfo, 0, 100)

	for i := 0; i < len(csv); i += 1024 * 1024 {
		end := i + 1024*1024
		if end > len(csv) {
			end = len(csv)
		}

		ci := ChunkTwoPass(csv[i:end])
		chunkInfos = append(chunkInfos, ci)
	}

	totalQuotes := 0
	for i, ci := range chunkInfos {
		afterFirstDelim := ci.positionDelimiterEven + 1
		if i == 0 {
			afterFirstDelim = 0
		} else if totalQuotes&1 == 1 {
			afterFirstDelim = ci.positionDelimiterOdd + 1
		}

		if uint64(afterFirstDelim) != sourceOfTruth[i].widowSize {
			t.Errorf("TestChunkTwoPass: got: %v want: %v", afterFirstDelim, sourceOfTruth[i].widowSize)
		}

		totalQuotes += ci.quotes
	}
}
