//+build !noasm !appengine

#define CREATE_MASK(Y1, Y2, R1, R2) \
	VPMOVMSKB Y1, R1                \
	VPMOVMSKB Y2, R2                \
	SHLQ      $32, R2               \
	ORQ       R1, R2

#define INIT_ONCE(R1, P1, LABEL) \
	CMPQ R1, $-1                 \
	JNZ  LABEL                   \
	CMPQ (P1), R1                \
	JZ   LABEL                   \
	ADDQ DX, (P1)                \
LABEL:

// func chunking_first_pass(buf []byte, quoteChar, delimiterChar uint64, quoteNextMask, quotes *int, even, odd *int)
TEXT ·chunking_first_pass(SB), 7, $0

	MOVQ         buf+0(FP), DI
	MOVQ         quoteChar+24(FP), AX     // get character for quote
	MOVQ         AX, X6
	VPBROADCASTB X6, Y6
	MOVQ         delimiterChar+32(FP), AX // get character for delimiter
	MOVQ         AX, X7
	VPBROADCASTB X7, Y7

	MOVQ quoteNextMask+40(FP), R11
	MOVQ quotes+48(FP), R10
	MOVQ even+56(FP), R9
	MOVQ odd+64(FP), R8

	XORQ DX, DX

	VMOVDQU (DI)(DX*1), Y8     // load low 32-bytes
	VMOVDQU 0x20(DI)(DX*1), Y9 // load high 32-bytes

	// detect quotes
	VPCMPEQB Y8, Y6, Y10
	VPCMPEQB Y9, Y6, Y11

    CREATE_MASK(Y10, Y11, AX, CX)
	MOVQ      CX, (R11)

loop:
	// find new line delimiter
	VPCMPEQB Y8, Y7, Y12
	VPCMPEQB Y9, Y7, Y13
    CREATE_MASK(Y12, Y13, CX, BX)

	VMOVDQU 0x40(DI)(DX*1), Y8 // load next low 32-bytes
	VMOVDQU 0x60(DI)(DX*1), Y9 // load next high 32-bytes

	// detect next quotes mask
	VPCMPEQB Y8, Y6, Y10
	VPCMPEQB Y9, Y6, Y11
    CREATE_MASK(Y10, Y11, AX, CX)

    // load previous quote mask and store new one
	MOVQ      (R11), AX
	MOVQ      CX, (R11)

	// cache even and odd positions
	MOVQ (R8), R14
	MOVQ (R9), R15

	PUSHQ DI
	PUSHQ DX
	CALL  ·handleMasksAvx2(SB)
	POPQ  DX
	POPQ  DI

	// check if either even or odd has changed, and add base upon initial change
	INIT_ONCE(R15, R9, skipEven)
	INIT_ONCE(R14, R8, skipOdd)

	ADDQ $0x40, DX
	MOVQ DX, CX
	ADDQ $0x40, CX
	CMPQ CX, buf_len+8(FP)
	JLT  loop

	VZEROUPPER
	RET

//
TEXT ·handleMasksAvx2Test(SB), 7, $0
	MOVQ quoteMask+0(FP), AX
	MOVQ newlineMask+8(FP), BX
	MOVQ quoteNextMask+16(FP), R11
	MOVQ quotes+24(FP), R10
	MOVQ even+32(FP), R9
	MOVQ odd+40(FP), R8
	CALL ·handleMasksAvx2(SB)
	RET

//
TEXT ·handleMasksAvx2(SB), 7, $0
	BSFQ    AX, DX
	BSFQ    BX, SI
	BSFQ    AX, DI
	MOVL    $0x40, DI
	CMOVQEQ DI, DX
	BSFQ    BX, R12
	CMOVQEQ DI, SI

loop:
	CMPQ DX, SI
	JGE  label1
	CMPQ DX, $0x3f
	JNE  label2
	MOVQ (R11), R12
	BTL  $0x0, R12
	JAE  label2
	LEAQ 0x1(DX), CX
	ANDQ $-0x2, R12
	MOVQ R12, (R11)
	CMPQ CX, $0x40
	SBBQ DX, DX
	MOVQ $-0x2, R12
	SHLQ CL, R12
	ANDQ DX, R12
	ANDQ R12, AX

label4:
	BSFQ    AX, DX
	CMOVQEQ DI, DX
	JMP     loop

label2:
	LEAQ  0x1(DX), CX
	CMPQ  CX, $0x40
	SBBQ  R12, R12
	MOVL  $0x1, R13
	SHLQ  CL, R13
	ANDQ  R12, R13
	TESTQ AX, R13
	JE    label3
	MOVQ  $-0x2, DX
	SHLQ  CL, DX
	ANDQ  R12, DX
	ANDQ  DX, AX
	JMP   label4

label3:
	INCQ (R10)
	CMPQ DX, $0x40
	SBBQ R12, R12
	MOVQ DX, CX
	MOVQ $-0x2, R13
	SHLQ CL, R13
	ANDQ R12, R13
	ANDQ R13, AX
	JMP  label4

label1:
	CMPQ SI, $0x40
	JE   done
	MOVQ (R10), R12
	BTL  $0x0, R12
	JB   label5
	CMPQ (R9), $-0x1
	JNE  label6
	MOVQ SI, (R9)

label6:
	CMPQ    SI, $0x40
	SBBQ    R12, R12
	MOVQ    SI, CX
	MOVQ    $-0x2, R13
	SHLQ    CL, R13
	ANDQ    R12, R13
	ANDQ    R13, BX
	BSFQ    BX, R12
	CMOVQEQ DI, R12
	MOVQ    R12, SI
	JMP     loop

label5:
	CMPQ (R8), $-0x1
	JNE  label6
	MOVQ SI, (R8)
	JMP  label6

done:
	RET
