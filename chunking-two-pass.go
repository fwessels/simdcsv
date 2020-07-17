package simdcsv

type chunkInfo struct {
	quotes                int
	positionDelimiterEven int
	positionDelimiterOdd  int
	firstCharIsQuote      bool
	lastCharIsQuote       bool
}

func ChunkTwoPass(buf []byte) (ci chunkInfo) {

	ci.positionDelimiterEven, ci.positionDelimiterOdd = -1, -1

	for i := 0; i < len(buf)-1; i++ {
		b := buf[i]
		if b == '"' {
			if buf[i+1] == '"' {
				i++ // skip second quote
			} else {
				ci.quotes++
			}
		} else if b == '\n' {
			if ci.quotes&1 == 1 {
				if ci.positionDelimiterOdd == -1 {
					ci.positionDelimiterOdd = i
				}
			} else {
				if ci.positionDelimiterEven == -1 {
					ci.positionDelimiterEven = i
				}
			}
		}
	}

	ci.firstCharIsQuote = buf[0] == '"'
	ci.lastCharIsQuote = buf[len(buf)-1] == '"'

	return
}

func ChunkTwoPassAsm(buf []byte) (ci chunkInfo) {

	quoteNextMask := 0
	ci.positionDelimiterEven, ci.positionDelimiterOdd = -1, -1

	chunking_first_pass(buf, '"', 0xa, &quoteNextMask, &ci.quotes, &ci.positionDelimiterEven, &ci.positionDelimiterOdd)

	ci.firstCharIsQuote = buf[0] == '"'
	ci.lastCharIsQuote = buf[len(buf)-1] == '"'

	return
}

func GetWidowSizes(chunkInfos []chunkInfo) (widowSizes []int) {

	widowSizes = make([]int, 0, 100)

	totalQuotes := 0
	prevChunkLastCharIsQuote := false

	for i, ci := range chunkInfos {
		afterFirstDelim := 0
		quoteCorrection := 0
		if i != 0 {
			even, odd := ci.positionDelimiterEven+1, ci.positionDelimiterOdd+1
			if prevChunkLastCharIsQuote && ci.firstCharIsQuote {
				// we have an escaped quote precisely at the split between chunks
				quoteCorrection = 1
				// and swap even and odd (detected exactly oppositve)
				even, odd = odd, even
			}

			if totalQuotes&1 == 0 {
				afterFirstDelim = even
			} else {
				afterFirstDelim = odd
			}
		}
		widowSizes = append(widowSizes, afterFirstDelim)

		totalQuotes += ci.quotes - quoteCorrection
		prevChunkLastCharIsQuote = ci.lastCharIsQuote
	}

	return
}
