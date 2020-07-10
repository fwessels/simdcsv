//+build !noasm !appengine

// chunking_first_pass(buf []byte, separator uint64) (out uint64)
TEXT ·chunking_first_pass(SB), 7, $0

    MOVQ         buf+0(FP), DI
    MOVQ         separator+24(FP), AX // get separator
    MOVQ         AX, X6
    VPBROADCASTB X6, Y6

    VMOVDQU    (DI), Y8          // load low 32-bytes
    VMOVDQU    0x20(DI), Y9      // load high 32-bytes

loop:
    VPCMPEQB  Y8, Y6, Y10
    VPCMPEQB  Y9, Y6, Y11

    VPMOVMSKB Y10, AX
    VPMOVMSKB Y11, CX
    SHLQ      $32, CX
    ORQ       CX, AX

    MOVQ      AX, out+32(FP)
    VZEROUPPER
	RET
