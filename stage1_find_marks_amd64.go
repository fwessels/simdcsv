package simdcsv

const INDEXES_SIZE = 1024

func Stage1FindMarks(msg []byte) (record []string) {

	indexes := [INDEXES_SIZE]uint32{}
	indexes_length := uint64(0)

	odd_ends, prev_iter_inside_quote, quote_bits, error_mask := uint64(0), uint64(0), uint64(0), uint64(0)

	find_marks_in_slice(msg, &indexes, &indexes_length,
		&odd_ends, &prev_iter_inside_quote, &quote_bits, &error_mask)

	record = make([]string, indexes_length/2)
	for i := uint64(0); i < indexes_length - 2; i += 2 {
		pos := uint64(indexes[i])
		record[i/2] = string(msg[pos:pos+uint64(indexes[i+1])])
	}
	return
}