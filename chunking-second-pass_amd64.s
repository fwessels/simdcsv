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

	MOVQ offset+56(FP), R9
	MOVQ columns+64(FP), R13
	MOVQ index+72(FP), R12
	MOVQ rows+80(FP), DI
	MOVQ line+88(FP), R11

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

// func parse_second_pass(input *Input, offset uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int)
TEXT ·parse_second_pass_test(SB), 7, $0
	MOVQ input+0(FP), DX
	MOVQ offset+8(FP), R9
	MOVQ columns+16(FP), R13
	MOVQ index+24(FP), R12
	MOVQ rows+32(FP), DI
	MOVQ line+40(FP), R11
	CALL ·parse_second_pass(SB)
	RET

// func parse_second_pass()
TEXT ·parse_second_pass(SB), 7, $0
	MOVL    $0x40, BX
	MOVQ    0(DX), SI
	BSFQ    SI, SI
	CMOVQEQ BX, SI
	MOVQ    0x8(DX), R8
	BSFQ    R8, R8
	CMOVQEQ BX, R8
	MOVQ    0x10(DX), R10
	BSFQ    R10, R10
	CMOVQEQ BX, R10
	JMP     label1

loop:
	MOVQ AX, R11

label1:
	CMPQ    SI, R8
	JGE     label2
	CMPQ    SI, R10
	JGE     label2
	TESTB   AL, 0(R13)
	MOVQ    0(R12), AX
	MOVQ    0(R13)(AX*8), R14
	LEAQ    0(SI)(R9*1), R15
	ADDQ    R15, R14
	MOVQ    R14, 0(R13)(AX*8)
	MOVQ    0(R12), R14
	LEAQ    0x1(R14), AX
	MOVQ    AX, 0(R12)
	MOVQ    R11, AX
	MOVQ    0x8(R13)(R14*8), R11
	LEAQ    0(R11)(R15*1), R11
	LEAQ    0x1(R11), R11
	MOVQ    R11, 0x8(R13)(R14*8)
	INCQ    0(R12)
	MOVQ    0(DX), R11
	CMPQ    SI, $0x40
	SBBQ    R14, R14
	MOVQ    SI, CX
	MOVQ    $-0x2, R15
	SHLQ    CL, R15
	ANDQ    R14, R15
	ANDQ    R11, R15
	BSFQ    R15, SI
	MOVQ    R15, 0(DX)
	CMOVQEQ BX, SI
	JMP     loop

label2:
	CMPQ    R8, SI
	JGE     label3
	CMPQ    R8, R10
	JGE     label3
	TESTB   AL, 0(R13)
	MOVQ    0(R12), AX
	MOVQ    0(R13)(AX*8), R14
	LEAQ    0(R8)(R9*1), R15
	ADDQ    R15, R14
	MOVQ    R14, 0(R13)(AX*8)
	MOVQ    0(R12), R14
	INCQ    R14
	MOVQ    R14, 0(R12)
	TESTB   AL, 0(DI)
	MOVQ    0(R11), AX
	MOVQ    R14, 0(DI)(AX*8)
	INCQ    0(R11)
	MOVQ    0(R12), AX
	MOVQ    0(R13)(AX*8), R14
	LEAQ    0(R14)(R15*1), R14
	LEAQ    0x1(R14), R14
	MOVQ    R14, 0(R13)(AX*8)
	INCQ    0(R12)
	MOVQ    0x8(DX), R14
	CMPQ    R8, $0x40
	SBBQ    R15, R15
	MOVQ    R8, CX
	MOVQ    R11, AX
	MOVQ    $-0x2, R11
	SHLQ    CL, R11
	ANDQ    R15, R11
	ANDQ    R14, R11
	BSFQ    R11, R8
	MOVQ    R11, 0x8(DX)
	CMOVQEQ BX, R8
	JMP     loop

label3:
	CMPQ  R10, SI
	JGE   done
	CMPQ  R10, R8
	JGE   done
	CMPQ  0x18(DX), $0x0
	JNE   label4
	TESTB AL, 0(R13)
	MOVQ  0(R12), R14
	LEAQ  -0x1(R14), AX
	MOVQ  -0x8(R13)(R14*8), R15
	INCQ  R15
	MOVQ  R15, -0x8(R13)(R14*8)

label5:
	MOVQ    0x18(DX), R14
	NOTQ    R14
	MOVQ    R14, 0x18(DX)
	MOVQ    0x10(DX), R14
	CMPQ    R10, $0x40
	SBBQ    R15, R15
	MOVQ    R10, CX
	MOVQ    R11, AX
	MOVQ    $-0x2, R11
	SHLQ    CL, R11
	ANDQ    R11, R15
	ANDQ    R14, R15
	BSFQ    R15, R11
	MOVQ    R15, 0x10(DX)
	CMOVQEQ BX, R11
	MOVQ    R11, R10
	JMP     loop

label4:
	TESTB AL, 0(R13)
	MOVQ  0(R12), AX
	MOVQ  0(R13)(AX*8), R14
	DECQ  R14
	MOVQ  R14, 0(R13)(AX*8)
	JMP   label5

done:
	RET
