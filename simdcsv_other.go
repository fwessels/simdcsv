// +build !amd64 appengine !gc noasm

package simdcsv

// SupportedCPU will return whether the CPU is supported.
func SupportedCPU() bool {
	return false
}

func stage1PreprocessBufferEx(buf []byte, separatorChar, quoted uint64, masks *[]uint64, postProc *[]uint64) ([]uint64, []uint64, uint64) {
	return nil, nil, 0
}

func stage2ParseBufferExStreaming(buf []byte, masks []uint64, delimiterChar uint64, inputStage2 *inputStage2, outputStage2 *outputAsm, rows *[]uint64, columns *[]string) ([]uint64, []string, bool) {
	return nil, nil, false
}
