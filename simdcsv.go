package simdcsv

import (
	"fmt"
)

func stage1(line []byte) []uint32 {

	return []uint32{0xcf, 10, 1, 19, 1, 4, 1, 0, 1, 0, 1, 2, 1, 6, 1, 0, 1, 4, 1, 2, 1, 2, 1, 15, 1, 5, 1, 1, 1, 6, 1, 18, 1, 2, 1, 5, 1, 5,
		2, 10, 1, 19, 1, 4, 1, 0, 1, 0, 1, 2, 1, 6, 1, 0, 1, 3, 1, 2, 1, 2, 1, 13, 1, 4, 1, 1, 1, 6, 1, 18, 1, 2, 1, 5, 1, 5}
}

func stage2(buf []byte, incrs []uint32) {

	const columns = 19

	offset, end := uint64(0), uint64(0)
	for i := 0; i < len(incrs); i += 2 {

		offset += uint64(incrs[i])
		end = offset + uint64(incrs[i+1])
		if end > offset {
			fmt.Print("'"+string(buf[offset:end])+"'", " ")
			offset = end
		}

		if i%(columns*2) == 18*2 {
			fmt.Println()
		}
	}
}
