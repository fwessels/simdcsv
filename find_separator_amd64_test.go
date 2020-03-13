package simdcsv

import (
	"testing"
	"strings"
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

func TestFindSeparatorTab(t *testing.T) {

	testCases := []struct {
		input       string
		expected    uint64
	}{
		{strings.Repeat("\t", 64), ^uint64(0x0)},
		{"column1\tcolumn2\tcolumn3\tcolumn4\tcolumn5\tcolumn6\tcolumn7\tcolumn8\t", 0x8080808080808080},
	}

	for i, tc := range testCases {

		mask := find_separator([]byte(tc.input), 0x9)

		if mask != tc.expected {
			t.Errorf("TestFindSeparatorTab(%d): got: 0x%x want: 0x%x", i, mask, tc.expected)
		}
	}
}