package simdcsv

//go:noescape
func find_marks_in_slice(msg []byte, indexes *[INDEXES_SIZE]uint32, indexes_length *uint64) (out uint64)
