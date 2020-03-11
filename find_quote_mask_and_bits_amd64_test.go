package simdcsv

import (
	"testing"
)

func testFindQuoteMaskAndBits(t *testing.T, f func([]byte, uint64, *uint64, *uint64, *uint64) uint64) {

	testCases := []struct {
		inputOE      uint64 // odd_ends
		input        string
		expected     uint64
		expectedQB   uint64 // quote_bits
		expectedPIIQ uint64 // prev_iter_inside_quote
		expectedEM   uint64 // error_mask
	}{
		{0x0, `  ""                                                            `, 0x4, 0xc, 0 ,0},
		{0x0, `  "-"                                                           `, 0xc, 0x14, 0 ,0},
		{0x0, `  "--"                                                          `, 0x1c, 0x24, 0 ,0},
		{0x0, `  "---"                                                         `, 0x3c, 0x44, 0 ,0},
		{0x0, `  "-------------"                                               `, 0xfffc, 0x10004, 0 ,0},
		{0x0, `  "---------------------------------------"                     `, 0x3fffffffffc, 0x40000000004, 0 ,0},
		{0x0, `"--------------------------------------------------------------"`, 0x7fffffffffffffff, 0x8000000000000001, 0 ,0},

		// quote is not closed --> prev_iter_inside_quote should be set
		{0x0, `                                                            "---`, 0xf000000000000000, 0x1000000000000000, ^uint64(0) ,0},
		{0x0, `                                                            "", `, 0x1000000000000000, 0x3000000000000000, 0 ,0},
		{0x0, `                                                            "-",`, 0x3000000000000000, 0x5000000000000000, 0 ,0},
		{0x0, `                                                            "--"`, 0x7000000000000000, 0x9000000000000000, 0 ,0},
		{0x0, `                                                            "---`, 0xf000000000000000, 0x1000000000000000, ^uint64(0),0},

		// test previous mask ending in backslash
		{0x1, `"                                                               `, 0x0, 0x0, 0x0,0x0},
		{0x1, `"""                                                             `, 0x2, 0x6, 0x0 ,0x0},
		{0x0, `"                                                               `, 0xffffffffffffffff, 0x1, ^uint64(0),0x0},
		{0x0, `"""                                                             `, 0xfffffffffffffffd, 0x7, ^uint64(0), 0x0},

		// test invalid chars (< 0x20) that are enclosed in quotes
		{0x0, `"` + string([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}) + ` "                             `, 0x3ffffffff, 0x400000001, 0, 0x1fffffffe},
		{0x0, `"` + string([]byte{0, 32, 1, 32, 2, 32, 3, 32, 4, 32, 5, 32, 6, 32, 7, 32, 8, 32, 9, 32, 10, 32, 11, 32, 12, 32, 13, 32, 14, 32, 15, 32, 16, 32, 17, 32, 18, 32, 19, 32, 20, 32, 21, 32, 22, 32, 23, 32, 24, 32, 25, 32, 26, 32, 27, 32, 28, 32, 29, 32, 31}) + ` "`, 0x7fffffffffffffff, 0x8000000000000001, 0, 0x2aaaaaaaaaaaaaaa},
		{0x0, `" ` + string([]byte{0, 32, 1, 32, 2, 32, 3, 32, 4, 32, 5, 32, 6, 32, 7, 32, 8, 32, 9, 32, 10, 32, 11, 32, 12, 32, 13, 32, 14, 32, 15, 32, 16, 32, 17, 32, 18, 32, 19, 32, 20, 32, 21, 32, 22, 32, 23, 32, 24, 32, 25, 32, 26, 32, 27, 32, 28, 32, 29, 32, 31}) + `"`, 0x7fffffffffffffff, 0x8000000000000001, 0, 0x5555555555555554},
	}

	for i, tc := range testCases {

		prev_iter_inside_quote, quote_bits, error_mask := uint64(0), uint64(0), uint64(0)

		mask := f([]byte(tc.input), tc.inputOE, &prev_iter_inside_quote, &quote_bits, &error_mask)

		if mask != tc.expected {
			t.Errorf("testFindQuoteMaskAndBits(%d): got: 0x%x want: 0x%x", i, mask, tc.expected)
		}

		if quote_bits != tc.expectedQB {
			t.Errorf("testFindQuoteMaskAndBits(%d): got quote_bits: 0x%x want: 0x%x", i, quote_bits, tc.expectedQB)
		}

		if prev_iter_inside_quote != tc.expectedPIIQ {
			t.Errorf("testFindQuoteMaskAndBits(%d): got prev_iter_inside_quote: 0x%x want: 0x%x", i, prev_iter_inside_quote, tc.expectedPIIQ)
		}

		if error_mask != tc.expectedEM {
			t.Errorf("testFindQuoteMaskAndBits(%d): got error_mask: 0x%x want: 0x%x", i, error_mask, tc.expectedEM)
		}
	}

	testCasesPIIQ := []struct {
		inputPIIQ    uint64
		input        string
		expectedPIIQ uint64
	}{
		// prev_iter_inside_quote state remains unchanged
		{ uint64(0), `----------------------------------------------------------------`, uint64(0)},
		{ ^uint64(0), `----------------------------------------------------------------`, ^uint64(0)},

		// prev_iter_inside_quote state remains flips
		{ uint64(0), `---------------------------"------------------------------------`, ^uint64(0)},
		{ ^uint64(0), `---------------------------"------------------------------------`, uint64(0)},

		// prev_iter_inside_quote state remains flips twice (thus unchanged)
		{ uint64(0), `----------------"------------------------"----------------------`, uint64(0)},
		{ ^uint64(0), `----------------"------------------------"----------------------`, ^uint64(0)},
	}

	for i, tc := range testCasesPIIQ {

		prev_iter_inside_quote, quote_bits, error_mask := tc.inputPIIQ, uint64(0), uint64(0)

		f([]byte(tc.input), 0, &prev_iter_inside_quote, &quote_bits, &error_mask)

		if prev_iter_inside_quote != tc.expectedPIIQ {
			t.Errorf("testFindQuoteMaskAndBits(%d): got prev_iter_inside_quote: 0x%x want: 0x%x", i, prev_iter_inside_quote, tc.expectedPIIQ)
		}
	}
}

func TestFindQuoteMaskAndBits(t *testing.T) {
	t.Run("avx2", func(t *testing.T) {
		testFindQuoteMaskAndBits(t, find_quote_mask_and_bits)
	})
}
