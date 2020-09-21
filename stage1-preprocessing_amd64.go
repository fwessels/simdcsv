package simdcsv

//go:noescape
func stage1_preprocess_buffer(buf []byte, separatorChar uint64, input *stage1Input, output *stage1Output, postProc *[]uint64, offset uint64) (processed uint64)

//go:noescape
func stage1_preprocess_test(input *stage1Input, output *stage1Output)

//go:noescape
func stage1_preprocess()

func Stage1PreprocessBuffer(buf []byte, separatorChar uint64) ([]uint64) {

	return Stage1PreprocessBufferEx(buf, separatorChar, nil)
}

func Stage1PreprocessBufferEx(buf []byte, separatorChar uint64, postProc *[]uint64) ([]uint64) {

	if postProc == nil {
		_postProc := make([]uint64, 0, 128)
		postProc = &_postProc
	}

	processed :=uint64(0)
	for {
		input := stage1Input{}
		output := stage1Output{}
		processed = stage1_preprocess_buffer(buf, uint64(','), &input, &output, postProc, processed)

		if processed >= uint64(len(buf)) {
			break
		}

		// Check if we need to grow the slice for keeping track of the lines to post process
		if len(*postProc) >= cap(*postProc)/2 {
			_postProc := make([]uint64, len(*postProc), cap(*postProc)*2)
			copy(_postProc, (*postProc)[:])
			postProc = &_postProc
		}
	}

	return *postProc
}
