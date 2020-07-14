package simdcsv

import (
	"encoding/hex"
	"fmt"
	_ "io/ioutil"
	"math/bits"
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

	out, even, odd, quotes := chunking_first_pass([]byte(file)[0x30:0x70], 0xa)
	fmt.Printf("%064b\n", out)
	fmt.Println(even, odd, quotes)
}

func handleMasks(quoteMask, newlineMask uint64, quotes *uint64, even, odd *int) {

	const clearMask = 0xfffffffffffffffe

	quotePos := bits.TrailingZeros64(quoteMask)
	newlinePos := bits.TrailingZeros64(newlineMask)

	for {
		//fmt.Println("  quotePos:", quotePos)
		//fmt.Println("newlinePos:", newlinePos)

		if quotePos < newlinePos {
			*quotes += 1
			// clear out active bit
			quoteMask &= clearMask << quotePos
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

func TestHandleMasks(t *testing.T) {

	testCases := []struct {
		quoteMask      uint64
		newlineMask    uint64
		expectedQuotes uint64
		expectedEven   int
		expectedOdd    int
	}{
		{
			0b00101000,
			0b01000000, 2, 6, -1,
		},
		{
			0b10001000,
			0b01000000, 2, -1, 6,
		},
		{
			0b00000010001000,
			0b10000001000000, 2, 13, 6,
		},
		{
			0b00000000100000,
			0b10000000000100, 1, 2, 13,
		},
		{
			0b00100000000010,
			0b10000000000100, 2, 13, 2,
		},
	}

	for i, tc := range testCases {
		quotes, even, odd := uint64(0), -1, -1
		handleMasks(tc.quoteMask, tc.newlineMask, &quotes, &even, &odd)

		quotesAsm, evenAsm, oddAsm := uint64(0), -1, -1
		handle_masks(tc.quoteMask, tc.newlineMask, &quotesAsm, &evenAsm, &oddAsm)

		if quotes != quotesAsm {
			t.Errorf("TestHandleMasks(%d): mismatch for asm: %d want: %d", i, quotes, quotesAsm)
		}
		if even != evenAsm {
			t.Errorf("TestHandleMasks(%d): mismatch for asm: %d want: %d", i, even, evenAsm)
		}
		if odd != oddAsm {
			t.Errorf("TestHandleMasks(%d): mismatch for asm: %d want: %d", i, odd, oddAsm)
		}

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
