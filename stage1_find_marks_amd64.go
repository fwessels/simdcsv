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

	fmt.Println(indexes[:indexes_length])
	for i := uint64(0); i < indexes_length - 2; i += 2 {
		pos := uint64(indexes[i])
		fmt.Println(string(msg[pos:pos+uint64(indexes[i+1])]))
	}
}
