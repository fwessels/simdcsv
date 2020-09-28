package simdcsv

//go:noescape
func stage1_preprocess_buffer(buf []byte, separatorChar uint64, input *stage1Input, output *stage1Output, postProc *[]uint64, offset uint64) (processed uint64)

//go:noescape
func stage1_preprocess_test(input *stage1Input, output *stage1Output)

//go:noescape
func stage1_preprocess()

//go:noescape
func partialLoad()

func Stage1PreprocessBuffer(buf []byte, separatorChar uint64) ([]uint64) {

	return Stage1PreprocessBufferEx(buf, separatorChar, nil)
}

func Stage1PreprocessBufferEx(buf []byte, separatorChar uint64, postProc *[]uint64) ([]uint64) {

	if postProc == nil {
		_postProc := make([]uint64, 0, 128)
		postProc = &_postProc
	}

	processed, quoted :=uint64(0), uint64(0)
	for {
		inputStage1, outputStage1 := stage1Input{}, stage1Output{}
		inputStage1.quoted = quoted

		processed = stage1_preprocess_buffer(buf, separatorChar, &inputStage1, &outputStage1, postProc, processed)

		if processed >= uint64(len(buf)) {
			break
		}

		// Check if we need to grow the slice for keeping track of the lines to post process
		if len(*postProc) >= cap(*postProc)/2 {
			_postProc := make([]uint64, len(*postProc), cap(*postProc)*2)
			copy(_postProc, (*postProc)[:])
			postProc = &_postProc
		}

		quoted = inputStage1.quoted
	}

	return *postProc
}
