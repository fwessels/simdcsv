package simdcsv

import (
	"fmt"
	"bytes"
	"math/bits"
	"strings"
	"encoding/hex"
	"encoding/csv"
	"log"
	"reflect"
)

// Substitute values when preprocessing a chunk
// NB 0x0 should be avoided (since trailing bytes
// beyond the end of the buffer are zeroed out)
const preprocessedDelimiter = 0x1
const preprocessedSeparator = 0x2

// This value needs to be unique and is not actually used in the preprocessing phase
// (since opening and closing quotes are eliminated)
const preprocessedQuote = 0x3

const preprocessedDoubleQuote = 0x4

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

func stage1Masking(quotesDoubleMask, crnlMask, quotesMask uint64, positions *[64]uint64, index *uint64) {

	const clearMask = 0xfffffffffffffffe

	quotesDoublePos := bits.TrailingZeros64(quotesDoubleMask)
	crnlPos := bits.TrailingZeros64(crnlMask)
	quotesPos := bits.TrailingZeros64(quotesMask)

	quoted := uint64(0)

	for {
		if quotesDoublePos < crnlPos && quotesDoublePos <= quotesPos {

			if quoted != 0 {
				(*positions)[*index] = uint64(quotesDoublePos)
				*index++
			}

			quotesDoubleMask &= clearMask << quotesDoublePos
			quotesDoublePos = bits.TrailingZeros64(quotesDoubleMask)

			// TODO:
			// 1. Clear corresponding two bits in quotesMask as well (easy)
			// 2. Handle case where double quote is split over two masks (HARD)

		} else if crnlPos < quotesDoublePos && crnlPos < quotesPos {

			if quoted != 0 {
				(*positions)[*index] = uint64(crnlPos)
				*index++
			}

			crnlMask &= clearMask << crnlPos
			crnlPos = bits.TrailingZeros64(crnlMask)

		} else if quotesPos < quotesDoublePos && quotesPos < crnlPos {

			quoted = ^quoted

			quotesMask &= clearMask << quotesPos
			quotesPos = bits.TrailingZeros64(quotesMask)

		} else {
			// we must be done
			break
		}
	}
}

type stage1Input struct {
	quoteMaskIn			 uint64
	separatorMaskIn 	 uint64
	carriageReturnMaskIn uint64
	quoteMaskInNext 	 uint64
	quoted				 bool
}

type stage1Output struct {
	quoteMaskOut 		  uint64
	separatorMaskOut 	  uint64
	carriageReturnMaskOut uint64
}

func preprocessMasksToMasks(input *stage1Input, output *stage1Output) {

	const clearMask = 0xfffffffffffffffe

	separatorPos := bits.TrailingZeros64(input.separatorMaskIn)
	carriageReturnPos := bits.TrailingZeros64(input.carriageReturnMaskIn)
	quotePos := bits.TrailingZeros64(input.quoteMaskIn)

	for {
		if quotePos < separatorPos && quotePos < carriageReturnPos {

			if input.quoted && quotePos == 63 && input.quoteMaskInNext&1 == 1 { // last bit of quote mask and first bit of next quote mask set?
				// clear out both active bit and ...
				input.quoteMaskIn &= clearMask << quotePos
				// first bit of next quote mask
				input.quoteMaskInNext &= ^uint64(1)
			} else if input.quoted && input.quoteMaskIn&(1<<(quotePos+1)) != 0 { // next quote bit is also set (so two adjacent bits) ?
				// clear out both active bit and subsequent bit
				input.quoteMaskIn &= clearMask << (quotePos + 1)
			} else {
				output.quoteMaskOut |= 1 << quotePos
				input.quoted = !input.quoted

				input.quoteMaskIn &= clearMask << quotePos
			}

			quotePos = bits.TrailingZeros64(input.quoteMaskIn)

		} else if separatorPos < quotePos && separatorPos < carriageReturnPos {

			if !input.quoted {
				output.separatorMaskOut |= 1 << separatorPos
			}

			input.separatorMaskIn &= clearMask << separatorPos
			separatorPos = bits.TrailingZeros64(input.separatorMaskIn)

		} else if carriageReturnPos < quotePos && carriageReturnPos < separatorPos {

			if !input.quoted {
				output.carriageReturnMaskOut |= 1 << carriageReturnPos
			}

			input.carriageReturnMaskIn &= clearMask << carriageReturnPos
			carriageReturnPos = bits.TrailingZeros64(input.carriageReturnMaskIn)

		} else {
			// we must be done
			break
		}
	}

	return
}

func preprocessInPlaceMasks(in []byte, quoted *bool) (quoteMask, separatorMask, carriageReturnMask uint64) {

	// TODO: Return indexes of columns where we need to post-process

	for i := 0; i < 64 && i < len(in); i++ {
		b := in[i]

		if *quoted {
			if b == '"' && i+1 < len(in) && in[i+1] == '"' {
				i += 1
			} else if b == '"' {
				quoteMask |= 1 << i // in[i] = preprocessedQuote
				*quoted = false
				//} else if b == '\r' && i+1 < len(in) && in[i+1] == '\n' {
				//	i += 1
				//} else {
			}
		} else {
			if b == '"' {
				quoteMask |= 1 << i // in[i] = preprocessedQuote
				*quoted = true
			} else if b == '\r' { // && i+1 < len(in) && in[i+1] == '\n' {
				carriageReturnMask |= 1 << i // in[i] = '\n'
			} else if b == ',' {
				// replace separator with '\2'
				separatorMask |= 1 << i // in[i] = preprocessedSeparator
			}
		}
	}

	return
}

func clearAndMerge(data []byte, mask, replacement uint64) {

	for i := 0; i < 64 && i < len(data); i++ {
		if mask & (1 << i) == (1 << i) {
			data[i] = byte(replacement)
		}
	}
}

func alternativeStage1Masks(data []byte) {

	buf := make([]byte, len(data))
	copy(buf, data)

	quoted := false
	quoteMask0, separatorMask0, carriageReturnMask0 := preprocessInPlaceMasks(buf, &quoted)
	quoteMask1, separatorMask1, carriageReturnMask1 := preprocessInPlaceMasks(buf[64:], &quoted)

	fmt.Println()
	fmt.Printf("%s", string(bytes.ReplaceAll(bytes.ReplaceAll(buf[:64], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))
	fmt.Printf("·%s\n", string(bytes.ReplaceAll(bytes.ReplaceAll(buf[64:], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))

	fmt.Printf("%064b·%064b\n", bits.Reverse64(quoteMask0), bits.Reverse64(quoteMask1))
	fmt.Printf("%064b·%064b\n", bits.Reverse64(separatorMask0), bits.Reverse64(separatorMask1))
	fmt.Printf("%064b·%064b\n", bits.Reverse64(carriageReturnMask0), bits.Reverse64(carriageReturnMask1))

	clearAndMerge(buf, quoteMask0, preprocessedQuote)
	clearAndMerge(buf[64:], quoteMask1, preprocessedQuote)

	clearAndMerge(buf, separatorMask0, preprocessedSeparator)
	clearAndMerge(buf[64:], separatorMask1, preprocessedSeparator)

	clearAndMerge(buf, carriageReturnMask0, '\n')
	clearAndMerge(buf[64:], carriageReturnMask1, '\n')

	fmt.Print(hex.Dump(buf))

	simdrecords := Stage2ParseBuffer(buf, 0xa, preprocessedSeparator, preprocessedQuote, nil)
	fmt.Println(simdrecords)

	//
	// postprocess
	//   replace "" to " in specific columns
	//   replace \r\n to \n in specific columns
	fmt.Printf("double quotes: `%s`\n", simdrecords[3][2])
	simdrecords[3][2] = strings.ReplaceAll(simdrecords[3][2], "\"\"", "\"")
	fmt.Printf(" carriage ret: `%s`\n", simdrecords[2][1])
	simdrecords[2][1] = strings.ReplaceAll(simdrecords[2][1], "\r\n", "\n")

	fmt.Println()

	fmt.Println("[0]:", simdrecords[0])
	fmt.Println("[1]:", simdrecords[1])
	fmt.Println("[2]:", simdrecords[2])
	fmt.Println("[3]:", simdrecords[3])

	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("encoding/csv: %v", err)
	}

	if !reflect.DeepEqual(simdrecords, records) {
		log.Fatalf("alternativeStage1Masks: got %v, want %v", simdrecords, records)
	}
}
