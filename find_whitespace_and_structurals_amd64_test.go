package simdcsv

import (
	"testing"
)

func testFindWhitespaceAndStructurals(t *testing.T, f func([]byte, *uint64, *uint64)) {

	testCases := []struct {
		input          string
		expected_ws    uint64
		expected_strls uint64
	}{
		{`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`, 0x0, 0x0},
		{` aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`, 0x1, 0x0},
		{`:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`, 0x0, 0x1},
		{` :aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`, 0x1, 0x2},
		{`: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`, 0x2, 0x1},
		{`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa `, 0x8000000000000000, 0x0},
		{`aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:`, 0x0, 0x8000000000000000},
		{`a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a `, 0xaaaaaaaaaaaaaaaa, 0x0},
		{` a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a a`, 0x5555555555555555, 0x0},
		{`a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:`, 0x0, 0xaaaaaaaaaaaaaaaa},
		{`:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a:a`, 0x0, 0x5555555555555555},
		{`                                                                `, 0xffffffffffffffff, 0x0},
		{`{                                                               `, 0xfffffffffffffffe, 0x1},
		{`}                                                               `, 0xfffffffffffffffe, 0x1},
		{`"                                                               `, 0xfffffffffffffffe, 0x0},
		{`::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::`, 0x0, 0xffffffffffffffff},
		{`{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{{`, 0x0, 0xffffffffffffffff},
		{`}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}}`, 0x0, 0xffffffffffffffff},
		{`  :                                                             `, 0xfffffffffffffffb, 0x4},
		{`    :                                                           `, 0xffffffffffffffef, 0x10},
		{`      :     :      :          :             :                  :`, 0x7fffefffbff7efbf, 0x8000100040081040},
	}

	for i, tc := range testCases {
		whitespace := uint64(0)
		structurals := uint64(0)

		f([]byte(tc.input), &whitespace, &structurals)

		if whitespace != tc.expected_ws {
			t.Errorf("testFindWhitespaceAndStructurals(%d): got: 0x%x want: 0x%x", i, whitespace, tc.expected_ws)
		}

		if structurals != tc.expected_strls {
			t.Errorf("testFindWhitespaceAndStructurals(%d): got: 0x%x want: 0x%x", i, structurals, tc.expected_strls)
		}
	}
}

func TestFindWhitespaceAndStructurals(t *testing.T) {
	t.Run("avx2", func(t *testing.T) {
		testFindWhitespaceAndStructurals(t, find_whitespace_and_structurals)
	})
}
