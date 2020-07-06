package simdcsv

import (
	"fmt"
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
		quotes, positionDelimiterEven, positionDelimiterOdd, lastCharIsQuote := ChunkTwoPass([]byte(tc.input))

		fmt.Println(quotes, positionDelimiterEven, positionDelimiterOdd, lastCharIsQuote)

		if quotes != tc.quotes {
			t.Errorf("TestChunkTwoPass(%d): got: %d want: %d", i, quotes, tc.quotes)
		}

		if positionDelimiterEven != tc.posDelimEven {
			t.Errorf("TestChunkTwoPass(%d): got: %d want: %d", i, positionDelimiterEven, tc.posDelimEven)
		}

		if positionDelimiterOdd != tc.posDelimOdd {
			t.Errorf("TestChunkTwoPass(%d): got: %d want: %d", i, positionDelimiterOdd, tc.posDelimOdd)
		}

		if lastCharIsQuote != tc.lastCharIsQuote {
			t.Errorf("TestChunkTwoPass(%d): got: %v want: %v", i, lastCharIsQuote, tc.lastCharIsQuote)
		}
	}
}
