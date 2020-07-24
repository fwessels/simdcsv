package simdcsv

import (
	_ "fmt"
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

	ParseSecondPass(buf, delimiter, separator, quote, parse_second_pass_test)
}

func ParseSecondPass(buffer []byte, delimiter, separator, quote rune,
	f func(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int)) ([]uint64, []uint64) {

	separatorMasks := getBitMasks([]byte(buffer), byte(separator))
	delimiterMasks := getBitMasks([]byte(buffer), byte(delimiter))
	quoteMasks := getBitMasks([]byte(buffer), byte(quote))

	//fmt.Printf(" separator: %064b\n", bits.Reverse64(separatorMasks[0]))
	//fmt.Printf(" delimiter: %064b\n", bits.Reverse64(delimiterMasks[0]))
	//fmt.Printf("     quote: %064b\n", bits.Reverse64(quoteMasks[0]))

	columns, rows := [128]uint64{}, [128]uint64{}
	columns[0] = 0
	index, line := 1, 0
	offset := uint64(0)
	input := Input{}

	for maskIndex := 0; maskIndex < len(separatorMasks); maskIndex++ {
		input.separatorMask = separatorMasks[maskIndex]
		input.delimiterMask = delimiterMasks[maskIndex]
		input.quoteMask = quoteMasks[maskIndex]

		f(&input, offset, &columns, &index, &rows, &line)
		offset += 0x40
	}

	//for i := 0; i < index-1; i += 2 {
	//	if columns[i] == ^uint64(0) || columns[i+1] == ^uint64(0) {
	//		break
	//	}
	//	fmt.Printf("%016x - %016x: %s\n", columns[i], columns[i+1], string(buffer[columns[i]:columns[i+1]]))
	//}
	//
	//for l := 0; l < line; l++ {
	//	fmt.Println(rows[l])
	//}

	return columns[:index-1], rows[:line]
}

type Input struct {
	separatorMask uint64
	delimiterMask uint64
	quoteMask     uint64
	quoted        uint64
}

func ParseSecondPassMasks(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int) {

	const clearMask = 0xfffffffffffffffe

	separatorPos := bits.TrailingZeros64(input.separatorMask)
	delimiterPos := bits.TrailingZeros64(input.delimiterMask)
	quotePos := bits.TrailingZeros64(input.quoteMask)

	for {
		if separatorPos < delimiterPos && separatorPos < quotePos {

			columns[*index] += uint64(separatorPos) + offset
			*index++
			columns[*index] += uint64(separatorPos) + offset + 1
			*index++

			input.separatorMask &= clearMask << separatorPos
			separatorPos = bits.TrailingZeros64(input.separatorMask)

		} else if delimiterPos < separatorPos && delimiterPos < quotePos {

			columns[*index] += uint64(delimiterPos) + offset
			*index++
			rows[*line] = uint64(*index)
			*line++
			columns[*index] += uint64(delimiterPos)  + offset + 1
			*index++

			input.delimiterMask &= clearMask << delimiterPos
			delimiterPos = bits.TrailingZeros64(input.delimiterMask)

		} else if quotePos < separatorPos && quotePos < delimiterPos {

			if (*input).quoted == 0 {
				columns[*index-1] += 1
			} else {
				columns[*index] -= 1
			}

			(*input).quoted = ^(*input).quoted

			input.quoteMask &= clearMask << quotePos
			quotePos = bits.TrailingZeros64(input.quoteMask)

		} else {
			// we must be done
			break
		}
	}
}

