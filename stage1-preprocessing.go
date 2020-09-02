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

	fmt.Printf("          %s\n", string(bytes.ReplaceAll(bytes.ReplaceAll(data, []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))
	stage1Masking(quotesDoubleMask, crnlMask, quotesMask[0], &positions, &index)

	fmt.Println(positions[:index])

	preprocessed := make([]byte, 0, len(data))
	// TODO: Use copy() instead of append()
	if index > 0 {
		preprocessed = append(preprocessed, data[:positions[0]]...)
		for i := range positions[:index-1] {
			if data[positions[i]+1] == byte(quote) {
				preprocessed = append(preprocessed, preprocessedDoubleQuote)
				preprocessed = append(preprocessed, data[positions[i]+2:positions[i+1]]...)
			} else {
				preprocessed = append(preprocessed, data[positions[i]+1:positions[i+1]]...)
			}
		}
		if index - 1 > 0 {
			if data[positions[index-1]+1] == byte(quote) {
				preprocessed = append(preprocessed, preprocessedDoubleQuote)
				preprocessed = append(preprocessed, data[positions[index-1]+2:len(data)]...)
			} else {
				preprocessed = append(preprocessed, data[positions[index-1]+1:len(data)]...)
			}
		}
	}

	fmt.Print(hex.Dump(preprocessed))

	preprocessed = bytes.ReplaceAll(preprocessed, []byte{byte(quote)}, []byte{preprocessedQuote})
	preprocessed = bytes.ReplaceAll(preprocessed, []byte{preprocessedDoubleQuote}, []byte{byte(quote)})

	fmt.Println()
	fmt.Print(hex.Dump(preprocessed))

	//00000000  65 72 74 22 2c 22 50 69  6b 65 22 2c 72 6f 62 0a  |ert","Pike",rob.|
	//00000010  4b 65 6e 6e 79 2c 54 68  6f 6d 70 73 6f 6e 2c 6b  |Kenny,Thompson,k|
	//00000020  65 6e 6e 79 0a 22 52 6f  62 65 72 74 22 2c 22 47  |enny."Robert","G|
	//00000030  72 69 65 73 65 6d 65 72  22 2c 22 67 72 22 69 22  |riesemer","gr"i"|
	//                                                                       ^  ^ ^

	// TODO: Still need to replace 0xa --> 0x1 and 0x2c --> 0x2 (but NOT within quotes)  <-- NASTY
	//       Consider replacing IN QUOTED FIELDS ONLY double quotes + \r\n data during preprocesing ?
	//       Fix strings in columns array as post-procesing step ???
	//       double quotes and \r\n do not occur so often

	//Stage2ParseBuffer()

	// Double Quotes
	// 1. consider replacing double quotes with 0x4: DONE
	// 3. replace 0x22 with 0x3 (preprocessed quote) DONE
	// 4. replace 0x5 with 0x22                      DONE
	//

	// TODO: Make sure empty fields are ok: ,"",

	// Carriage returns
	//
	// Try and only filter out in quoted fields
	//
	// If we can do \r\n --> \n\n second step will filter them out
	//
	// 1. consider replacing \r\n in quoted fields with 0xa: DONE (automatically)
	// 2. replace 0xd with 0xa                             : DONE
	preprocessed = bytes.ReplaceAll(preprocessed, []byte{'\r'}, []byte{'\n'})

	fmt.Println()
	fmt.Print(hex.Dump(preprocessed))
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

func preprocessInPlace(in []byte) {

	// TODO: Return indexes of columns where we need to post-process

	quoted := false

	for i := 0; i < len(in); i++ {
		b := in[i]

		if quoted {
			if b == '"' && i+1 < len(in) && in[i+1] == '"' {
				i += 1
			} else if b == '"' {
				in[i] = preprocessedQuote
				quoted = false
			//} else if b == '\r' && i+1 < len(in) && in[i+1] == '\n' {
			//	i += 1
			//} else {
			}
		} else {
			if b == '"' {
				in[i] = preprocessedQuote
				quoted = true
			} else if b == '\r' { // && i+1 < len(in) && in[i+1] == '\n' {
				in[i] = '\n'
			} else if b == ',' {
				// replace separator with '\2'
				in[i] = preprocessedSeparator
			}
		}
	}

	return
}

func alternativeStage1(data []byte) {

	// preprocess
	//   outside quoted fields
	//     1. replace ',' with preprocessedSeparator: DONE
	//     2. replace '"' with preprocessedQuote    : DONE
	//     3. replace \r\n with \n\n                : DONE

	buf := make([]byte, len(data))
	copy(buf, data)

	preprocessInPlace(buf)

	fmt.Println(hex.Dump(buf))

	simdrecords := Stage2ParseBuffer(buf, 0xa, preprocessedSeparator, preprocessedQuote, nil)

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
		log.Fatalf("alternativeStage1: got %v, want %v", simdrecords, records)
	}
}

func preprocessInPlaceMasks(in []byte, quoted *bool) (quoteMask, separatorMask, carriageReturnMask uint64) {

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
