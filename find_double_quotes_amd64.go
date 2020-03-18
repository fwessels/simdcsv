package simdcsv

//go:noescape
func find_double_quotes(mask uint64, indices []uint32) (entries uint64)
