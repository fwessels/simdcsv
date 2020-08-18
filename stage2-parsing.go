package simdcsv

import (
	"bytes"
	_ "fmt"
	"math/bits"
	"unsafe"
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

func PreprocessCarriageReturns(in []byte) (out []byte) {

	out = bytes.ReplaceAll(in, []byte{'\r', '\n'}, []byte{'\n'})
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

	Stage2Parse(buf, delimiter, separator, quote, stage2_parse_test)
}

func Stage2Parse(buffer []byte, delimiter, separator, quote rune,
	f func(input *Input, offset uint64, output *Output)) ([]uint64, []uint64, uint64) {

	separatorMasks := getBitMasks([]byte(buffer), byte(separator))
	delimiterMasks := getBitMasks([]byte(buffer), byte(delimiter))
	quoteMasks := getBitMasks([]byte(buffer), byte(quote))

	//fmt.Printf(" separator: %064b\n", bits.Reverse64(separatorMasks[0]))
	//fmt.Printf(" delimiter: %064b\n", bits.Reverse64(delimiterMasks[0]))
	//fmt.Printf("     quote: %064b\n", bits.Reverse64(quoteMasks[0]))

	columns, rows := [128]uint64{}, [128]uint64{}
	columns[0] = 0
	offset := uint64(0)
	input := Input{lastSeparatorOrDelimiter: ^uint64(0)}
	output := Output{&columns, 1, &rows, 0, 0, 0, 128}

	for maskIndex := 0; maskIndex < len(separatorMasks); maskIndex++ {
		input.separatorMask = separatorMasks[maskIndex]
		input.delimiterMask = delimiterMasks[maskIndex]
		input.quoteMask = quoteMasks[maskIndex]

		f(&input, offset, &output)
		offset += 0x40
	}

	return columns[:output.index-1], rows[:output.line], input.errorOffset
}

type Input struct {
	separatorMask            uint64
	delimiterMask            uint64
	quoteMask                uint64
	quoted                   uint64
	lastSeparatorOrDelimiter uint64
	lastClosingQuote         uint64
	errorOffset				 uint64
	base					 uint64
}

type Output struct {
	columns  *[128]uint64
	index    int
	rows     *[128]uint64
	line     int
	col_base uint64
	col_prev uint64
	col_cap  uint64
}

// Equivalent for invoking from Assembly
type OutputAsm struct {
	columns  unsafe.Pointer
	index    int
	rows     unsafe.Pointer
	line     int
	col_base uint64
	col_prev uint64
	col_cap  uint64
}

func Stage2ParseMasks(input *Input, offset uint64, output *Output) {

	const clearMask = 0xfffffffffffffffe

	separatorPos := bits.TrailingZeros64(input.separatorMask)
	delimiterPos := bits.TrailingZeros64(input.delimiterMask)
	quotePos := bits.TrailingZeros64(input.quoteMask)

	for {
		if separatorPos < delimiterPos && separatorPos < quotePos {

			if input.quoted == 0 {
				// verify that last closing quote is immediately followed by either a separator or delimiter
				if  input.lastClosingQuote > 0 &&
					input.lastClosingQuote + 1 != uint64(separatorPos) + offset {
					if input.errorOffset == 0 {
						input.errorOffset = uint64(separatorPos) + offset // mark first error position
					}
				}
				input.lastClosingQuote = 0

				output.columns[output.index] = (input.base + output.columns[output.index] + uint64(separatorPos) + offset) - output.columns[output.index-1] // size of previous element
				output.index++
				output.columns[output.index] = input.base + output.columns[output.index] + uint64(separatorPos) + offset + 1 // start of next element
				output.index++

				input.lastSeparatorOrDelimiter = uint64(separatorPos) + offset
			}

			input.separatorMask &= clearMask << separatorPos
			separatorPos = bits.TrailingZeros64(input.separatorMask)

		} else if delimiterPos < separatorPos && delimiterPos < quotePos {

			if input.quoted == 0 {
				// verify that last closing quote is immediately followed by either a separator or delimiter
				if  input.lastClosingQuote > 0 &&
					input.lastClosingQuote + 1 != uint64(delimiterPos) + offset {
					if input.errorOffset == 0 {
						input.errorOffset = uint64(delimiterPos) + offset // mark first error position
					}
				}
				input.lastClosingQuote = 0

				output.columns[output.index] = (input.base + output.columns[output.index] + uint64(delimiterPos) + offset) - output.columns[output.index-1]
				output.index++

				if uint64(output.index) / 2 - output.col_prev == 1 && // we just have a line with a single element
					output.columns[output.index-1] == 0 {			  // and its length is zero (implying empty line)
					// prevent empty lines from being written
				} else {
					// write out length for a new row
					output.rows[output.line] = uint64(output.index) / 2 - output.col_prev // length of element
					output.line++
				}

				output.col_prev = uint64(output.index) / 2	// keep current index for next round

				output.columns[output.index] = input.base  + output.columns[output.index] + uint64(delimiterPos) + offset + 1 // start of next element
				output.index++

				input.lastSeparatorOrDelimiter = uint64(delimiterPos) + offset
			}

			input.delimiterMask &= clearMask << delimiterPos
			delimiterPos = bits.TrailingZeros64(input.delimiterMask)

		} else if quotePos < separatorPos && quotePos < delimiterPos {

			if input.quoted == 0 {
				// check that this opening quote is preceded by either a separator or delimiter
				if input.lastSeparatorOrDelimiter + 1 != uint64(quotePos) + offset {
					if input.errorOffset == 0 {
						input.errorOffset = uint64(quotePos) + offset
					}
				}
				output.columns[output.index-1] += 1
			} else {
				output.columns[output.index] -= 1
				input.lastClosingQuote = uint64(quotePos) + offset // record position of last closing quote
			}

			input.quoted = ^input.quoted

			input.quoteMask &= clearMask << quotePos
			quotePos = bits.TrailingZeros64(input.quoteMask)

		} else {
			// we must be done
			break
		}
	}
}

