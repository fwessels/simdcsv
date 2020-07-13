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


TEXT ·handle_masks(SB), 7, $0
	SUBQ    $0x8, SP
	MOVQ    BP, 0(SP)
	LEAQ    0(SP), BP
	MOVQ    0x10(SP), AX
	BSFQ    AX, DX
	MOVQ    0x18(SP), BX
	BSFQ    BX, SI
	BSFQ    AX, DI
	MOVL    $0x40, DI
	CMOVQCS DI, DX
	BSFQ    BX, R8
	CMOVQHI DI, SI
	MOVQ    0x30(SP), R8
	MOVQ    0x28(SP), R9
	MOVQ    0x20(SP), R10

loop:
	CMPQ    SI, DX
	JGE     label1
	INCQ    0(R10)
	CMPQ    DX, $0x40
	SBBQ    R11, R11
	MOVQ    DX, CX
	MOVQ    $-0x2, R12
	SHLQ    CL, R12
	ANDQ    R12, R11
	ANDQ    R11, AX
	BSFQ    AX, DX
	CMOVQEQ DI, DX
	JMP     loop

label1:
	CMPQ SI, $0x40
	JE   done
	MOVQ (R10), R11
	BTL  $0x0, R11
	JB   label2
	CMPQ (R9), $-1
	JNE  label3
	MOVQ SI, (R9)

label3:
	CMPQ    SI, $0x40
	SBBQ    R11, R11
	MOVQ    SI, CX
	MOVQ    $-0x2, R12
	SHLQ    CL, R12
	ANDQ    R11, R12
	ANDQ    R12, BX
	BSFQ    BX, R11
	CMOVQEQ DI, R11
	MOVQ    R11, SI
	JMP     loop

label2:
	CMPQ (R8), $-1
	JNE  label3
	MOVQ SI, (R8)
	JMP  label3

done:
	MOVQ 0(SP), BP
	ADDQ $0x8, SP
	RET
