package simdcsv

func PreprocessDoubleQuotes(in []byte) (out []byte) {

	// Replace separator with '\0'
	// Remove surrounding quotes
	// Replace double quotes with single quote

	out = make([]byte, 0, len(in))
	quoted := false

	for i := 0; i < len(in); i++ {
		b := in[i]

		if quoted {
			if b == '"' && i+1 < len(in) && in[i+1] == '"' {
				// replace escaped quote with single quote
				out = append(out, '"')
			} else if b == '"' {
				quoted = false
			} else {
				out = append(out, b)
			}
		} else {
			if b == '"' {
				quoted = true
			} else if b == ',' {
				// replace separator with '\0'
				out = append(out, 0x0)
			} else {
				out = append(out, b)
			}
		}
	}

	return
}

