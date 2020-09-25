package simdcsv

//go:noescape
func stages_combined_buffer(buf []byte, separatorChar uint64, input *stage1Input, output *stage1Output, postProc *[]uint64, offset uint64, input2 *Input, output2 *OutputAsm, lastCharIsDelimiter uint64, rows []uint64, columns []string) (processed uint64)

//go:noescape
func _stage2_PRUNE()