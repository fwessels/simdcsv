//+build !noasm !appengine

#define CREATE_MASK(Y1, Y2, R1, R2) \
	VPMOVMSKB Y1, R1  \
	VPMOVMSKB Y2, R2  \
	SHLQ      $32, R2 \
	ORQ       R1, R2

// func parse_block_second_pass()
TEXT ·parse_block_second_pass(SB), 7, $0

	MOVQ         delimiterChar+24(FP), AX // get character for delimiter
	MOVQ         AX, X4
	VPBROADCASTB X4, Y4
	MOVQ         separatorChar+32(FP), AX // get character for separator
	MOVQ         AX, X5
	VPBROADCASTB X5, Y5
	MOVQ         quoteChar+40(FP), AX     // get character for quote
	MOVQ         AX, X6
	VPBROADCASTB X6, Y6

	XORQ DX, DX

    // Check whether it is necessary to adjust pointer for first string element
    MOVQ output+64(FP), R9
    MOVQ (R9), R9 // columnns pointer
    CMPQ (R9), $0
    JNZ  skip
    MOVQ buf+0(FP), DI
    MOVQ DI, (R9)
skip:

loop:
	MOVQ buf+0(FP), DI
	MOVQ input+48(FP), SI

	VMOVDQU (DI)(DX*1), Y8     // load low 32-bytes
	VMOVDQU 0x20(DI)(DX*1), Y9 // load high 32-bytes

	VPCMPEQB Y8, Y4, Y10
	VPCMPEQB Y9, Y4, Y11
	CREATE_MASK(Y10, Y11, AX, CX)
	MOVQ     CX, 8(SI)

	VPCMPEQB Y8, Y5, Y10
	VPCMPEQB Y9, Y5, Y11
	CREATE_MASK(Y10, Y11, AX, CX)
	MOVQ     CX, 0(SI)

	VPCMPEQB Y8, Y6, Y10
	VPCMPEQB Y9, Y6, Y11
	CREATE_MASK(Y10, Y11, AX, CX)
	MOVQ     CX, 16(SI)

	MOVQ offset+56(FP), DI
	MOVQ output+64(FP), R9

	PUSHQ DX
	MOVQ  input+48(FP), DX
	CALL  ·parse_second_pass(SB)
	POPQ  DX

	ADDQ $0x40, offset+56(FP)
	ADDQ $0x40, DX
	CMPQ DX, buf_len+8(FP)
	JLT  loop

	VZEROUPPER
	RET

// func parse_second_pass_test(input *Input, offset uint64, output *Output)
TEXT ·parse_second_pass_test(SB), 7, $0
	MOVQ input+0(FP), DX
	MOVQ offset+8(FP), DI
	MOVQ output+16(FP), R9
	CALL ·parse_second_pass(SB)
	RET

// func parse_second_pass()
TEXT ·parse_second_pass(SB), 7, $0
	MOVL    $0x40, BX
	MOVQ    (DX), SI
	BSFQ    SI, SI
	CMOVQEQ BX, SI
	MOVQ    0x8(DX), R8
	BSFQ    R8, R8
	CMOVQEQ BX, R8
	MOVQ    0x10(DX), R10
	BSFQ    R10, R10
	CMOVQEQ BX, R10

label1:
	CMPQ  SI, R8
	JGE   label4
	CMPQ  SI, R10
	JGE   label4
	CMPQ  0x18(DX), $0x0
	JNE   label3
	MOVQ  0x28(DX), R11
	TESTQ R11, R11
	JBE   label2
	INCQ  R11
	LEAQ  (SI)(DI*1), R12
	CMPQ  R11, R12
	JE    label2
	CMPQ  0x30(DX), $0x0
	JNE   label2
	MOVQ  R12, 0x30(DX)

label2:
	MOVQ  $0x0, 0x28(DX)
	MOVQ  (R9), R11
	TESTB AL, (R11)
	MOVQ  0x38(DX), R12
	MOVQ  0x8(R9), AX
	MOVQ  (R11)(AX*8), R13
	ADDQ  R13, R12
	ADDQ  SI, R12
	ADDQ  DI, R12
	LEAQ  -0x1(AX), R13
	MOVQ  -0x8(R11)(AX*8), R13
	SUBQ  R13, R12
	MOVQ  R12, (R11)(AX*8)
	MOVQ  0x8(R9), R11
	LEAQ  0x1(R11), AX
	MOVQ  AX, 0x8(R9)
	MOVQ  (R9), R12
	TESTB AL, (R12)
	MOVQ  0x38(DX), R13
	MOVQ  0x8(R12)(R11*8), R14
	ADDQ  R14, R13
	ADDQ  SI, R13
	LEAQ  (R13)(DI*1), R13
	LEAQ  0x1(R13), R13
	MOVQ  R13, 0x8(R12)(R11*8)
	INCQ  0x8(R9)
	LEAQ  (SI)(DI*1), R11
	MOVQ  R11, 0x20(DX)

label3:
	MOVQ    (DX), R11
	CMPQ    SI, $0x40
	SBBQ    R12, R12
	MOVQ    SI, CX
	MOVQ    $-0x2, R13
	SHLQ    CL, R13
	ANDQ    R12, R13
	ANDQ    R11, R13
	BSFQ    R13, SI
	MOVQ    R13, (DX)
	CMOVQEQ BX, SI
	JMP     label1

label4:
	CMPQ  R8, SI
	JGE   label7
	CMPQ  R8, R10
	JGE   label7
	CMPQ  0x18(DX), $0x0
	JNE   label6
	MOVQ  0x28(DX), R11
	TESTQ R11, R11
	JBE   label5
	INCQ  R11
	LEAQ  (R8)(DI*1), R12
	CMPQ  R11, R12
	JE    label5
	CMPQ  0x30(DX), $0x0
	JNE   label5
	MOVQ  R12, 0x30(DX)

label5:
	MOVQ  $0x0, 0x28(DX)
	MOVQ  (R9), R11
	TESTB AL, (R11)
	MOVQ  0x38(DX), R12
	MOVQ  0x8(R9), AX
	MOVQ  (R11)(AX*8), R13
	ADDQ  R13, R12
	ADDQ  R8, R12
	ADDQ  DI, R12
	LEAQ  -0x1(AX), R13
	MOVQ  -0x8(R11)(AX*8), R13
	SUBQ  R13, R12
	MOVQ  R12, (R11)(AX*8)
	MOVQ  0x8(R9), R11
	INCQ  R11
	MOVQ  R11, 0x8(R9)
	MOVQ  0x10(R9), R12
	TESTB AL, (R12)
	MOVQ  0x18(R9), AX
	MOVQ  R11, (R12)(AX*8)
	INCQ  0x18(R9)
	MOVQ  (R9), R11
	TESTB AL, (R11)
	MOVQ  0x38(DX), R12
	MOVQ  0x8(R9), AX
	MOVQ  (R11)(AX*8), R13
	ADDQ  R13, R12
	ADDQ  R8, R12
	LEAQ  (R12)(DI*1), R12
	LEAQ  0x1(R12), R12
	MOVQ  R12, (R11)(AX*8)
	INCQ  0x8(R9)
	LEAQ  (R8)(DI*1), R11
	MOVQ  R11, 0x20(DX)

label6:
	MOVQ    0x8(DX), R11
	CMPQ    R8, $0x40
	SBBQ    R12, R12
	MOVQ    R8, CX
	MOVQ    $-0x2, R13
	SHLQ    CL, R13
	ANDQ    R12, R13
	ANDQ    R11, R13
	BSFQ    R13, R8
	MOVQ    R13, 0x8(DX)
	CMOVQEQ BX, R8
	JMP     label1

label7:
	CMPQ R10, SI
	JGE  label11
	CMPQ R10, R8
	JGE  label11
	CMPQ 0x18(DX), $0x0
	JNE  label10
	MOVQ 0x20(DX), R11
	INCQ R11
	LEAQ (R10)(DI*1), R12
	CMPQ R11, R12
	JE   label8
	CMPQ 0x30(DX), $0x0
	JNE  label8
	MOVQ R12, 0x30(DX)

label8:
	MOVQ  (R9), R11
	TESTB AL, (R11)
	MOVQ  0x8(R9), R12
	LEAQ  -0x1(R12), AX
	MOVQ  -0x8(R11)(R12*8), R13
	INCQ  R13
	MOVQ  R13, -0x8(R11)(R12*8)

label9:
	MOVQ    0x18(DX), R11
	NOTQ    R11
	MOVQ    R11, 0x18(DX)
	MOVQ    0x10(DX), R11
	CMPQ    R10, $0x40
	SBBQ    R12, R12
	MOVQ    R10, CX
	MOVQ    $-0x2, R13
	SHLQ    CL, R13
	ANDQ    R13, R12
	ANDQ    R11, R12
	BSFQ    R12, R11
	MOVQ    R12, 0x10(DX)
	CMOVQEQ BX, R11
	MOVQ    R11, R10
	JMP     label1

label10:
	MOVQ  (R9), R11
	TESTB AL, (R11)
	MOVQ  0x8(R9), AX
	MOVQ  (R11)(AX*8), R12
	DECQ  R12
	MOVQ  R12, (R11)(AX*8)
	LEAQ  (R10)(DI*1), R11
	MOVQ  R11, 0x28(DX)
	JMP   label9

label11:
	RET
