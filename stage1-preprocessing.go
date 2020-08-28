package simdcsv

import (
	"fmt"
	"bytes"
	"math/bits"
	"encoding/hex"
)

// Substitute values when preprocessing a chunk
// NB 0x0 should be avoided (since trailing bytes
// beyond the end of the buffer are zeroed out)
const preprocessedDelimiter = 0x1
const preprocessedSeparator = 0x2

// This value needs to be unique and is not actually used in the preprocessing phase
// (since opening and closing quotes are eliminated)
const preprocessedQuote = 0x3

func preprocessDoubleQuotes(in []byte) (out []byte) {

	// Replace delimiter and separators
	// Remove any surrounding quotes
	// Replace double quotes with single quote

	out = make([]byte, 0, len(in))
	quoted := false

	for i := 0; i < len(in); i++ {
		b := in[i]

		if quoted {
			if b == '"' && i+1 < len(in) && in[i+1] == '"' {
				// replace escaped quote with single quote
				out = append(out, '"')
				// and skip next char
				i += 1
			} else if b == '"' {
				quoted = false
			} else if b == '\r' && i+1 < len(in) && in[i+1] == '\n' {
				out = append(out, '\n') // replace \r\n in quoted fields with single \n (conform to encoding/csv behavior)
				i += 1
			} else {
				out = append(out, b)
			}
		} else {
			if b == '"' {
				quoted = true
			} else if b == '\r' && i+1 < len(in) && in[i+1] == '\n' {
				// replace delimiter with '\1'
				out = append(out, preprocessedDelimiter)
				// and skip next char
				i += 1
			} else if b == '\n' {
				// replace delimiter with '\1'
				out = append(out, preprocessedDelimiter)
			} else if b == ',' {
				// replace separator with '\2'
				out = append(out, preprocessedSeparator)
			} else {
				out = append(out, b)
			}
		}
	}

	return
}

func preprocessCarriageReturns(in []byte) (out []byte) {

	out = bytes.ReplaceAll(in, []byte{'\r', '\n'}, []byte{'\n'})
	return
}

func preprocessStage1(data []byte) {

	quote := '"'

	data = data[0x22:]

	quotesMask := getBitMasks(data[:64], byte(quote))
	quotesNextMask := getBitMasks(data[1:65], byte(quote))
	quotesDoubleMask := quotesMask[0] & quotesNextMask[0]
	fmt.Printf("%064b\n", bits.Reverse64(quotesDoubleMask))

	carriageRetMask := getBitMasks(data[:64], byte('\r'))
	newlineMask := getBitMasks(data[1:65], byte('\n'))
	crnlMask := carriageRetMask[0] & newlineMask[0]
	fmt.Printf("%064b\n", bits.Reverse64(crnlMask))

	positions := [64]uint64{}
	index := uint64(0)

	stage1Masking(quotesDoubleMask, crnlMask, &positions, &index)

	fmt.Println(positions[:index])

	preprocessed := make([]byte, 0, len(data))
	if index > 0 {
		preprocessed = append(preprocessed, data[:positions[0]]...)
	}
	for i := range positions[:index-1] {
		preprocessed = append(preprocessed, data[positions[i]+1:positions[i+1]]...)
	}
	if index - 1 > 0 {
		preprocessed = append(preprocessed, data[positions[index-1]+1:len(data)]...)
	}

	fmt.Print(hex.Dump(preprocessed))
}

func stage1Masking(quotesDoubleMask, crnlMask uint64, positions *[64]uint64, index *uint64) {

	const clearMask = 0xfffffffffffffffe

	quotesDoublePos := bits.TrailingZeros64(quotesDoubleMask)
	crnlPos := bits.TrailingZeros64(crnlMask)

	for {
		if quotesDoublePos < crnlPos {

			(*positions)[*index] = uint64(quotesDoublePos)
			*index++

			quotesDoubleMask &= clearMask << quotesDoublePos
			quotesDoublePos = bits.TrailingZeros64(quotesDoubleMask)

		} else if crnlPos < quotesDoublePos {

			(*positions)[*index] = uint64(crnlPos)
			*index++

			crnlMask &= clearMask << crnlPos
			crnlPos = bits.TrailingZeros64(crnlMask)

		} else {
			// we must be done
			break
		}
	}
}
