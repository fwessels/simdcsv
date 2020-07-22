
// func parse_second_pass(separatorMask, delimiterMask, quoteMask uint64, output *[128]uint64, index *int, quoted *uint64)
TEXT ·parse_second_pass(SB), 7, $0
	MOVQ    separatorMask+0(FP), DX
	BSFQ    DX, BX
	MOVQ    delimiterMask+8(FP), SI
	BSFQ    SI, DI
	MOVQ    quoteMask+16(FP), R8
	BSFQ    R8, R9
	BSFQ    DX, R10
	MOVL    $0x40, R10
	CMOVQEQ R10, BX
	BSFQ    SI, R11
	CMOVQEQ R10, DI
	BSFQ    R8, R11
	CMOVQEQ R10, R9
	MOVQ    output+24(FP), R12
	MOVQ    index+32(FP), R13
	MOVQ    quoted+40(FP), R11

loop:
	CMPQ    BX, DI
	JGE     label1
	CMPQ    BX, R9
	JGE     label1
	TESTB   AL, (R12)
	MOVQ    (R13), AX
	MOVQ    (R12)(AX*8), R14
	ADDQ    BX, R14
	MOVQ    R14, (R12)(AX*8)
	MOVQ    (R13), R14
	LEAQ    0x1(R14), AX
	MOVQ    AX, (R13)
	MOVQ    0x8(R12)(R14*8), R15
	LEAQ    (R15)(BX*1), R15
	LEAQ    0x1(R15), R15
	MOVQ    R15, 0x8(R12)(R14*8)
	INCQ    (R13)
	CMPQ    BX, $0x40
	SBBQ    R14, R14
	MOVQ    BX, CX
	MOVQ    $-0x2, R15
	SHLQ    CL, R15
	ANDQ    R14, R15
	ANDQ    R15, DX
	BSFQ    DX, BX
	CMOVQEQ R10, BX
	JMP     loop

label1:
	CMPQ    DI, BX
	JGE     label3
	CMPQ    DI, R9
	JGE     label3
	TESTB   AL, (R12)
	MOVQ    (R13), AX
	MOVQ    (R12)(AX*8), R14
	ADDQ    DI, R14
	MOVQ    R14, (R12)(AX*8)
	MOVQ    (R13), R14
	LEAQ    0x1(R14), AX
	MOVQ    AX, (R13)
	MOVQ    0x8(R12)(R14*8), R15
	LEAQ    (R15)(DI*1), R15
	LEAQ    0x1(R15), R15
	MOVQ    R15, 0x8(R12)(R14*8)
	INCQ    (R13)
	CMPQ    DI, $0x40
	SBBQ    R14, R14
	MOVQ    DI, CX
	MOVQ    $-0x2, R15
	SHLQ    CL, R15
	ANDQ    R14, R15
	ANDQ    R15, SI
	BSFQ    SI, DI
	CMOVQEQ R10, DI
	JMP     loop

label3:
	CMPQ  R9, BX
	JGE   done
	CMPQ  R9, DI
	JGE   done
	CMPQ  (R11), $0x0
	JNE   label4
	TESTB AL, (R12)
	MOVQ  (R13), R14
	LEAQ  -0x1(R14), AX
	MOVQ  -0x8(R12)(R14*8), R15
	INCQ  R15
	MOVQ  R15, -0x8(R12)(R14*8)

label5:
	MOVQ    (R11), R14
	NOTQ    R14
	MOVQ    R14, (R11)
	CMPQ    R9, $0x40
	SBBQ    R14, R14
	MOVQ    R9, CX
	MOVQ    $-0x2, R15
	SHLQ    CL, R15
	ANDQ    R15, R14
	ANDQ    R14, R8
	BSFQ    R8, R14
	CMOVQEQ R10, R14
	MOVQ    R14, R9
	JMP     loop

label4:
	TESTB AL, (R12)
	MOVQ  (R13), AX
	MOVQ  (R12)(AX*8), R14
	DECQ  R14
	MOVQ  R14, (R12)(AX*8)
	JMP   label5

done:
	RET
