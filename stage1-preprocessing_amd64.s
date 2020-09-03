//+build !noasm !appengine

// func stage1_preprocess_test(input *stage1Input, output *stage1Output)
TEXT ·stage1_preprocess_test(SB), 7, $0
	MOVQ input+0(FP), AX
	MOVQ output+8(FP), R10
	CALL ·stage1_preprocess(SB)
	RET

// func stage1_preprocess()
TEXT ·stage1_preprocess(SB), 7, $0
	RET
