//+build !noasm !appengine

// chunking_first_pass(buf []byte, separator uint64) (out uint64)
TEXT ·chunking_first_pass(SB), 7, $0

    MOVQ         buf+0(FP), DI
    MOVQ         separator+24(FP), AX // get separator
    MOVQ         AX, X6
    VPBROADCASTB X6, Y6
    MOVQ         $0x0a, AX // get new line
    MOVQ         AX, X7
    VPBROADCASTB X7, Y7

    MOVQ         $-1, R8
    MOVQ         $-1, R9
    XORQ         R10, R10

    VMOVDQU    (DI), Y8          // load low 32-bytes
    VMOVDQU    0x20(DI), Y9      // load high 32-bytes

loop:
    // find separator
    VPCMPEQB  Y8, Y6, Y10
    VPCMPEQB  Y9, Y6, Y11

    // find new line
    VPCMPEQB  Y8, Y7, Y12
    VPCMPEQB  Y9, Y7, Y13

    VPMOVMSKB Y12, AX
    VPMOVMSKB Y13, CX
    SHLQ      $32, CX
    ORQ       CX, AX
    POPCNTQ   AX, BX
    ADDQ      BX, R10

    VPMOVMSKB Y10, AX
    VPMOVMSKB Y11, CX
    SHLQ      $32, CX
    ORQ       CX, AX

    MOVQ      AX, out+32(FP)
    MOVQ      R8, positionDelimiterEven+40(FP)
    MOVQ      R9, positionDelimiterOdd+48(FP)
    MOVQ      R10, quotes+56(FP)
    VZEROUPPER
	RET
