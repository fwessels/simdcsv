package simdcsv

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/bits"
	"reflect"
	"strings"
	"testing"
)

func TestChunkingFirstPass(t *testing.T) {

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

	quoteNextMask := 0
	quotes, even, odd := 0, -1, -1
	chunking_first_pass([]byte(file)[0x30:0x70], '"', 0xa, &quoteNextMask, &quotes, &even, &odd)
	fmt.Println(quotes, even, odd)
}

func TestFirstPassBufferValidation(t *testing.T) {

	testCases := []struct {
		buffer   string
		expected int
	}{
		{"", 0},  // test empty buffer
		{" ", 0}, // test buffer that is not a multiple of 64-bytes
		{strings.Repeat(` "`, 32), 32},
		{strings.Repeat(` "`, 64), 64},
	}

	for i, tc := range testCases {
		ci := ChunkTwoPassAvx2([]byte(tc.buffer))

		if ci.quotes != tc.expected {
			t.Errorf("TestFirstPassBufferValidation(%d): got: %d want: %d", i, ci.quotes, tc.expected)
		}
	}
}

func TestFirstPass(t *testing.T) {

	csv, err := ioutil.ReadFile("test-data/Emails.csv")
	if err != nil {
		panic(err)
	}

	for size := 512 * 1024; size <= 512*1024; size *= 2 {
		chunk := csv[0:size]

		ci := ChunkTwoPass(chunk)
		ciAsm := ChunkTwoPassAvx2(chunk)
		fmt.Println(ciAsm)
		if !reflect.DeepEqual(ci, ciAsm) {
			t.Errorf("TestFirstPass: mismatch for asm: %v want: %v", ci, ciAsm)
		}
	}
}

//
//       quoteMask = 64-bit mask of quotes
//     newlineMask = 64-bit mask of new lines
// nextCharIsQuote = bool indicate next char is a quote (first char of next ZMM word)
//
func handleMasks(quoteMask, newlineMask uint64, quoteNextMask, quotes *uint64, even, odd *int) {

	const clearMask = 0xfffffffffffffffe

	quotePos := bits.TrailingZeros64(quoteMask)
	newlinePos := bits.TrailingZeros64(newlineMask)

	for {
		//fmt.Println("  quotePos:", quotePos)
		//fmt.Println("newlinePos:", newlinePos)

		if quotePos < newlinePos {
			// check if we have two consecutive escaped quotes
			if quotePos == 63 && *quoteNextMask&1 == 1 {
				// clear out both active bit and ...
				quoteMask &= clearMask << (quotePos + 1)
				// first bit of next quote mask
				*quoteNextMask &= ^uint64(1)
			} else if quoteMask&(1<<(quotePos+1)) != 0 {
				// clear out both active bit and subsequent bit
				quoteMask &= clearMask << (quotePos + 1)
			} else {
				*quotes += 1
				// clear out active bit
				quoteMask &= clearMask << quotePos
			}
			quotePos = bits.TrailingZeros64(quoteMask)
		} else {

			if newlinePos == 64 {
				break
			}

			if *quotes&1 == 0 {
				if *even == -1 {
					*even = newlinePos
				}
			} else {
				if *odd == -1 {
					*odd = newlinePos
				}
			}

			newlineMask &= clearMask << newlinePos
			newlinePos = bits.TrailingZeros64(newlineMask)
		}
	}
}

func testHandleMasks(t *testing.T, f func(quoteMask, newlineMask uint64, quoteNextMask, quotes *uint64, even, odd *int)) {

	testCases := []struct {
		quoteMask       uint64
		quoteNextMask   uint64
		newlineMask     uint64
		expectedQuotes  uint64
		expectedEven    int
		expectedOdd     int
	}{
		//
		// Generic test cases
		//
		{
			0b00101000, 0,
			0b01000000,
			2, 6, -1,
		},
		{
			0b10001000, 0,
			0b01000000,
			2, -1, 6,
		},
		{
			0b00000010001000, 0,
			0b10000001000000,
			2, 13, 6,
		},
		{
			0b00000000100000, 0,
			0b10000000000100,
			1, 2, 13,
		},
		{
			0b00100000000010, 0,
			0b10000000000100,
			2, 13, 2,
		},
		//
		//
		// Test cases with escaped quotes
		//
		{
			0b11011000, 0,
			0b00000000,
			0, -1, -1,
		},
		{
			0b00010100, 0,
			0b01000000,
			2, 6, -1,
		},
		{
			0b00010110, 0,
			0b01000000,
			1, -1, 6,
		},
		{
			0b0001101000010110, 0,
			0b0100000001000000,
			2, 14, 6,
		},
		//
		// Special cases
		//
		{
			0x5555555555555555, 0,
			0xaaaaaaaaaaaaaaaa,
			32, 3, 1,
		},
		{
			0xaaaaaaaaaaaaaaaa, 0,
			0x5555555555555555,
			32, 0, 2,
		},
		{
			0xaaaaaaaaaaaaaaaa, 1,
			0x5555555555555555,
			31, 0, 2,
		},
		{
			0xffffffffffffffff, 0,
			0x0,
			0, -1, -1,
		},
		{
			0xfffffffffffffffe, 0,
			0x0,
			1, -1, -1,
		},
		{
			0xfffffffffffffffe, 1,
			0x0,
			0, -1, -1,
		},
		{
			0x8000000000000000, 0,
			0x0,
			1, -1, -1,
		},
		{
			0x8000000000000000, 1,
			0x0,
			0, -1, -1,
		},
		{
			0x0, 0,
			0xffffffffffffffff,
			0, 0, -1,
		},
		{
			0x1, 0,
			0xfffffffffffffffe,
			1, -1, 1,
		},
		//
		// Test cases using nextCharIsQuote
		//
		{
			0xa000000000000000, 0,
			0x0,
			2, -1, -1,
		},
		{
			0xa000000000000000, 1,
			0x0,
			1, -1, -1,
		},
		{
			0xaaaaaaaaaaaaaaaa, 1,
			0x5555555555555555,
			31, 0, 2,
		},
		{
			0x5555555555555555, 1,
			0xaaaaaaaaaaaaaaaa,
			32, 3, 1,
		},
	}

	for i, tc := range testCases {
		quotes, even, odd := uint64(0), -1, -1
		f(tc.quoteMask, tc.newlineMask, &tc.quoteNextMask, &quotes, &even, &odd)

		if quotes != tc.expectedQuotes {
			t.Errorf("TestHandleMasks(%d): got: %d want: %d", i, quotes, tc.expectedQuotes)
		}

		if even != tc.expectedEven {
			t.Errorf("TestHandleMasks(%d): got: %d want: %d", i, even, tc.expectedEven)
		}

		if odd != tc.expectedOdd {
			t.Errorf("TestHandleMasks(%d): got: %d want: %d", i, odd, tc.expectedOdd)
		}
	}
}

func TestHandleMasks(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		testHandleMasks(t, handleMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testHandleMasks(t, handleMasksAvx2Test)
	})
}

func testHandleSubsequentMasks(t *testing.T, f func(quoteMask, newlineMask uint64, quoteNextMask, quotes *uint64, even, odd *int)) {

	getMasks := func(str string) (masks []uint64) {

		if len(str)%64 != 0 {
			panic("Input strings should be a multipe of 64")
		}

		masks = make([]uint64, 0)

		for i := 0; i < len(str); i += 64 {
			mask := uint64(0)
			for b, c := range str[i : i+64] {
				if c == '"' {
					mask = mask | (1 << b)
				}
			}
			masks = append(masks, mask)
		}
		return
	}

	testCases := []struct {
		quoteString    string
		expectedQuotes uint64
	}{
		//
		// two subsequent quotes
		{
			strings.Repeat(" ", 62) + `""` + strings.Repeat(" ", 64),
			0,
		},
		{
			strings.Repeat(" ", 63) + `""` + strings.Repeat(" ", 63),
			0,
		},
		{
			strings.Repeat(" ", 64) + `""` + strings.Repeat(" ", 62),
			0,
		},
		//
		// three subsequent quotes
		{
			strings.Repeat(" ", 61) + `"""` + strings.Repeat(" ", 64),
			1,
		},
		{
			strings.Repeat(" ", 62) + `"""` + strings.Repeat(" ", 63),
			1,
		},
		{
			strings.Repeat(" ", 63) + `"""` + strings.Repeat(" ", 62),
			1,
		},
		{
			strings.Repeat(" ", 64) + `"""` + strings.Repeat(" ", 61),
			1,
		},
		//
		// four subsequent quotes
		{
			strings.Repeat(" ", 60) + `""""` + strings.Repeat(" ", 64),
			0,
		},
		{
			strings.Repeat(" ", 61) + `""""` + strings.Repeat(" ", 63),
			0,
		},
		{
			strings.Repeat(" ", 62) + `""""` + strings.Repeat(" ", 62),
			0,
		},
		{
			strings.Repeat(" ", 63) + `""""` + strings.Repeat(" ", 61),
			0,
		},
		{
			strings.Repeat(" ", 64) + `""""` + strings.Repeat(" ", 60),
			0,
		},
		//
		// five subsequent quotes
		{
			strings.Repeat(" ", 59) + `"""""` + strings.Repeat(" ", 64),
			1,
		},
		{
			strings.Repeat(" ", 60) + `"""""` + strings.Repeat(" ", 63),
			1,
		},
		{
			strings.Repeat(" ", 61) + `"""""` + strings.Repeat(" ", 62),
			1,
		},
		{
			strings.Repeat(" ", 62) + `"""""` + strings.Repeat(" ", 61),
			1,
		},
		{
			strings.Repeat(" ", 63) + `"""""` + strings.Repeat(" ", 60),
			1,
		},
		{
			strings.Repeat(" ", 64) + `"""""` + strings.Repeat(" ", 59),
			1,
		},
	}

	for ii, tc := range testCases {

		quoteMasks := getMasks(tc.quoteString)
		// fmt.Printf("%016x %016x\n", quoteMasks[0], quoteMasks[1])

		quoteMask, quoteNextMask := quoteMasks[0], uint64(0)
		quotes, even, odd := uint64(0), -1, -1

		for i := range quoteMasks {
			if i+1 < len(quoteMasks) {
				quoteNextMask = quoteMasks[i+1]
			} else {
				quoteNextMask = 0
			}
			f(quoteMask, 0, &quoteNextMask, &quotes, &even, &odd)
			quoteMask = quoteNextMask
		}

		if quotes != tc.expectedQuotes {
			t.Errorf("TestHandleMasks(%d): got: %d want: %d", ii, quotes, tc.expectedQuotes)
		}
	}
}

func TestHandleSubsequentMasks(t *testing.T) {
	t.Run("go", func(t *testing.T) {
		testHandleSubsequentMasks(t, handleMasks)
	})
	t.Run("avx2", func(t *testing.T) {
		testHandleSubsequentMasks(t, handleMasksAvx2Test)
	})
}

func BenchmarkFirstPassAsm(b *testing.B) {

	csv, err := ioutil.ReadFile("test-data/Emails.csv")
	if err != nil {
		panic(err)
	}

	const chunkSize = 512 * 1024

	b.SetBytes(chunkSize /*int64(len(csv))*/)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			quoteNextMask, quotes, even, odd := 0, 0, -1, -1

			chunking_first_pass(csv[0:chunkSize], '"', 0xa, &quoteNextMask, &quotes, &even, &odd)}
	})
}
