package simdcsv

func ChunkTwoPass(buf []byte) (quotes, positionDelimiterEven, positionDelimiterOdd int, lastCharIsQuote bool) {

	positionDelimiterEven, positionDelimiterOdd = -1, -1

	for i := 0; i < len(buf)-1; i++ {
		b := buf[i]
		if b == '"' {
			if buf[i+1] == '"' {
				i++ // skip second quote
			} else {
				quotes++
			}
		} else if b == '\n' {
			if quotes&1 == 1 {
				if positionDelimiterOdd == -1 {
					positionDelimiterOdd = i
				}
			} else {
				if positionDelimiterEven == -1 {
					positionDelimiterEven = i
				}
			}
		}
	}

	lastCharIsQuote = buf[len(buf)-1] == '"'

	return
}
