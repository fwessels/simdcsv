//+build !noasm !appengine

TEXT ·_find_separator(SB), $0-24

    MOVQ         input+0(FP), DI
	MOVQ         separator+8(FP), AX // get separator
	MOVQ         AX, X6
	VPBROADCASTB X6, Y6

    VMOVDQU    (DI), Y8          // load low 32-bytes
    VMOVDQU    0x20(DI), Y9      // load high 32-bytes

    CALL ·__find_separator(SB)

    VZEROUPPER
	MOVQ        AX, maks+16(FP)
    RET

TEXT ·__find_separator(SB), $0
	VPCMPEQB  Y8, Y6, Y10
	VPCMPEQB  Y9, Y6, Y11

	VPMOVMSKB Y10, AX
	VPMOVMSKB Y11, CX
	SHLQ      $32, CX
	ORQ       CX, AX
    RET
