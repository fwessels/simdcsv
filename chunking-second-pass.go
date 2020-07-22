package simdcsv

import (
	"fmt"
	"math/bits"
)
const PreprocessedDelimiter = 0x0
const PreprocessedSeparator = 0x1

func PreprocessDoubleQuotes(in []byte) (out []byte) {

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
			} else {
				out = append(out, b)
			}
		} else {
			if b == '"' {
				quoted = true
			} else if b == '\r' && i+1 < len(in) && in[i+1] == '\n' {
				// replace delimiter with '\1'
				out = append(out, PreprocessedDelimiter)
				// and skip next char
				i += 1
			} else if b == '\n' {
				// replace delimiter with '\1'
				out = append(out, PreprocessedDelimiter)
			} else if b == ',' {
				// replace separator with '\0'
				out = append(out, PreprocessedSeparator)
			} else {
				out = append(out, b)
			}
		}
	}

	return
}

func SecondPass(buffer []byte) {

	containsDoubleQuotes := true

	delimiter, separator, quote := '\n', ',', '"'
	buf := buffer
	if containsDoubleQuotes {
		buf = PreprocessDoubleQuotes(buffer)

		delimiter, separator, quote = PreprocessedDelimiter, PreprocessedSeparator, 0x02
	}

	ParseSecondPass(buf, delimiter, separator, quote)
}

func ParseSecondPass(buffer []byte, delimiter, separator, quote rune) {

	separatorMasks := getBitMasks([]byte(buffer), byte(separator))
	delimiterMasks := getBitMasks([]byte(buffer), byte(delimiter))
	quoteMasks := getBitMasks([]byte(buffer), byte(quote))

	fmt.Printf(" separator: %064b\n", bits.Reverse64(separatorMasks[0]))
	fmt.Printf(" delimiter: %064b\n", bits.Reverse64(delimiterMasks[0]))
	fmt.Printf("     quote: %064b\n", bits.Reverse64(quoteMasks[0]))

	output := [128]uint64{}
	quoted := uint64(0)
	output[0] = 0
	index := 1

	ParseSecondPassMasks(separatorMasks[0], delimiterMasks[0], quoteMasks[0], &output, &index, &quoted)

	for i := 0; i < index-1; i += 2 {
		if output[i] == ^uint64(0) || output[i+1] == ^uint64(0) {
			break
		}
		fmt.Printf("%016x - %016x: %s\n", output[i], output[i+1], string(buffer[output[i]:output[i+1]]))
	}
}


func ParseSecondPassMasks(separatorMask, delimiterMask, quoteMask uint64, output *[128]uint64, index *int, quoted *uint64) {

	const clearMask = 0xfffffffffffffffe

	separatorPos := bits.TrailingZeros64(separatorMask)
	delimiterPos := bits.TrailingZeros64(delimiterMask)
	quotePos := bits.TrailingZeros64(quoteMask)

	for {
		if separatorPos < delimiterPos && separatorPos < quotePos {

			output[*index] += uint64(separatorPos)
			*index++
			output[*index] += uint64(separatorPos + 1)
			*index++

			separatorMask &= clearMask << separatorPos
			separatorPos = bits.TrailingZeros64(separatorMask)

		} else if delimiterPos < separatorPos && delimiterPos < quotePos {

			output[*index] += uint64(delimiterPos)
			*index++
			output[*index] += uint64(delimiterPos + 1)
			*index++

			delimiterMask &= clearMask << delimiterPos
			delimiterPos = bits.TrailingZeros64(delimiterMask)

		} else if quotePos < separatorPos && quotePos < delimiterPos {

			if *quoted == 0 {
				output[*index-1] += 1
			} else {
				output[*index] -= 1
			}

			*quoted = ^*quoted

			quoteMask &= clearMask << quotePos
			quotePos = bits.TrailingZeros64(quoteMask)

		} else {
			// we must be done
			break
		}
	}
}
