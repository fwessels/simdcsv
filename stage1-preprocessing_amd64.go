package simdcsv

//go:noescape
func stage1_preprocess_buffer(buf []byte, input *stage1Input, output *stage1Output)

//go:noescape
func stage1_preprocess_test(input *stage1Input, output *stage1Output)

//go:noescape
func stage1_preprocess()
