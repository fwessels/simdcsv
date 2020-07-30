//+build !noasm !appengine

// find_marks_in_slice(msg []byte, indexes *[INDEXES_SIZE]uint32, indexes_length, indexes_cap, carried, position *uint64) (pmsg, out uint64)
TEXT ·find_marks_in_slice(SB), 7, $0
	RET
