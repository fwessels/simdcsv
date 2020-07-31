package simdcsv

const INDEXES_SIZE = 1024*10

func Stage1FindMarks(msg []byte) (record []string) {

	indexes := [INDEXES_SIZE]uint32{}
	indexes_length := uint64(0)

	record = make([]string, indexes_length/2)
	for i := uint64(0); i < indexes_length; i += 2 {
		pos := uint64(indexes[i])
		record[i/2] = string(msg[pos:pos+uint64(indexes[i+1])])
	}
	return
}
