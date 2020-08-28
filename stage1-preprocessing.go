package simdcsv

import (
	"bytes"
)

// Substitute values when preprocessing a chunk
// NB 0x0 should be avoided (since trailing bytes
// beyond the end of the buffer are zeroed out)
const preprocessedDelimiter = 0x1
const preprocessedSeparator = 0x2
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
