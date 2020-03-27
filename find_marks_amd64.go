package simdcsv

//go:noescape
func find_marks_in_slice(msg []byte, indexes *[INDEXES_SIZE]uint32, indexes_length *uint64, indexes_cap uint64, carried, position *uint64,
					     prev_iter_inside_quote, quote_bits, error_mask *uint64) (pmsg, endofline, out uint64)
