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

	fmt.Printf(" separator: %064b\n", separatorMasks[0])
	fmt.Printf(" delimiter: %064b\n", delimiterMasks[0])
	fmt.Printf("     quote: %064b\n", quoteMasks[0])

	ParseSecondPassMasks(separatorMasks[0], delimiterMasks[0], quoteMasks[0])
}

func ParseSecondPassMasks(separatorMask, delimiterMask, quoteMask uint64) {

	const clearMask = 0xfffffffffffffffe

	separatorPos := bits.TrailingZeros64(separatorMask)
	delimiterPos := bits.TrailingZeros64(delimiterMask)
	quotePos := bits.TrailingZeros64(quoteMask)

	for {
		if separatorPos < delimiterPos && separatorPos < quotePos {

			fmt.Println("Encountered separator at", separatorPos)
			separatorMask &= clearMask << separatorPos
			separatorPos = bits.TrailingZeros64(separatorMask)

		} else if delimiterPos < separatorPos && delimiterPos < quotePos {

			fmt.Println("Encountered delimiter at", delimiterPos)
			delimiterMask &= clearMask << delimiterPos
			delimiterPos = bits.TrailingZeros64(delimiterMask)

		} else if quotePos < separatorPos && quotePos < delimiterPos {

			fmt.Println("Encountered qoute at", quotePos)
			quoteMask &= clearMask << quotePos
			quotePos = bits.TrailingZeros64(quoteMask)

		} else {
			// we must be done
			break
		}
	}
}
