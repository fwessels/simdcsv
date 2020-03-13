package simdcsv

import (
	"testing"
)

func TestFindSeparator(t *testing.T) {

	testCases := []struct {
		input    string
		expected uint64
	}{
		{`                                                                `, 0x0},
		{`,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,,`, ^uint64(0x0)},
		{`, , , , , , , , , , , , , , , , , , , , , , , , , , , , , , , , `, 0x5555555555555555},
		{` , , , , , , , , , , , , , , , , , , , , , , , , , , , , , , , ,`, 0xaaaaaaaaaaaaaaaa},
		{`,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  ,,  `, 0x3333333333333333},
	}

	for i, tc := range testCases {

		mask := find_separator([]byte(tc.input), ',')

		if mask != tc.expected {
			t.Errorf("TestFindSeparator(%d): got: 0x%x want: 0x%x", i, mask, tc.expected)
		}
	}
}