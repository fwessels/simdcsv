package simdcsv

import (
	"fmt"
	"math/bits"
)

const INDEXES_SIZE = 1024

func Stage1FindMarks(msg []byte) {

	indexes := [INDEXES_SIZE]uint32{}
	indexes_length := uint64(0)

	result := find_marks_in_slice(msg, &indexes, &indexes_length)
	fmt.Printf("%064b\n", bits.Reverse64(result))

	pos := uint64(0)
	for _, index := range indexes[:indexes_length] {
		pos += uint64(index)
		fmt.Println(string(msg[pos]))
	}
}
