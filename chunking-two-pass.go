package simdcsv

type chunkInfo struct {
	quotes                int
	positionDelimiterEven int
	positionDelimiterOdd  int
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

	ci.lastCharIsQuote = buf[len(buf)-1] == '"'

	return
}
