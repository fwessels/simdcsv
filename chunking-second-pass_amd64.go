package simdcsv

//go:noescape
func parse_second_pass_test(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int)

//go:noescape
func parse_second_pass()
