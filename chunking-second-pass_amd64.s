// func parse_second_pass(separatorMask, delimiterMask, quoteMask, offset uint64, quoted *uint64, columns *[128]uint64, index *int, rows *[128]uint64, line *int, scratch1, scratch2 uint64)
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
	MOVQ    quoted+32(FP), R11
	MOVQ    rows+56(FP), R12
	MOVQ    line+64(FP), R13
	MOVQ    columns+40(FP), R14
	MOVQ    offset+24(FP), R15
	MOVQ    index+48(FP), AX
	JMP     loop

label2:
	MOVQ quoted+32(FP), CX
	MOVQ rows+56(FP), R11
	MOVQ CX, R11
	MOVQ R12, R13
	MOVQ rows+56(FP), R12

loop:
	CMPQ    BX, DI
	JGE     label1
	CMPQ    BX, R9
	JGE     label1
	TESTB   AL, (R14)
	MOVQ    R13, CX
	MOVQ    (AX), R13
	MOVQ    (R14)(R13*8), R12
	LEAQ    (BX)(R15*1), R11
	ADDQ    R11, R12
	MOVQ    R12, (R14)(R13*8)
	MOVQ    (AX), R12
	LEAQ    0x1(R12), R13
	MOVQ    R13, (AX)
	MOVQ    0x8(R14)(R12*8), R13
	LEAQ    (R13)(R11*1), R11
	LEAQ    0x1(R11), R11
	MOVQ    R11, 0x8(R14)(R12*8)
	INCQ    (AX)
	CMPQ    BX, $0x40
	SBBQ    R11, R11
	MOVQ    CX, R12
	MOVQ    BX, CX
	MOVQ    $-0x2, R13
	SHLQ    CL, R13
	ANDQ    R11, R13
	ANDQ    R13, DX
	BSFQ    DX, BX
	CMOVQEQ R10, BX
	JMP     label2

label1:
	MOVQ    DX, scratch1+72(FP)
	CMPQ    DI, BX
	JGE     label3
	CMPQ    DI, R9
	JGE     label3
	TESTB   AL, (R14)
	MOVQ    R11, CX
	MOVQ    (AX), R11
	MOVQ    (R14)(R11*8), DX
	ADDQ    DI, R15
	ADDQ    R15, DX
	MOVQ    DX, (R14)(R11*8)
	MOVQ    (AX), DX
	INCQ    DX
	MOVQ    DX, (AX)
	TESTB   AL, (R12)
	MOVQ    (R13), R11
	MOVQ    DX, (R12)(R11*8)
	INCQ    (R13)
	MOVQ    (AX), DX
	MOVQ    (R14)(DX*8), R11
	LEAQ    (R11)(R15*1), R11
	LEAQ    0x1(R11), R11
	MOVQ    R11, (R14)(DX*8)
	INCQ    (AX)
	CMPQ    DI, $0x40
	SBBQ    DX, DX
	MOVQ    CX, R11
	MOVQ    DI, CX
	MOVQ    $-0x2, R15
	SHLQ    CL, R15
	ANDQ    DX, R15
	ANDQ    R15, SI
	BSFQ    SI, DX
	CMOVQEQ R10, DX

label5:
	MOVQ R13, R12
	MOVQ offset+24(FP), R15
	MOVQ DX, DI
	MOVQ scratch1+72(FP), DX
	JMP  label2

label3:
	CMPQ  R9, BX
	JGE   done
	CMPQ  R9, DI
	JGE   done
	CMPQ  (R11), $0x0
	JNE   label4
	TESTB AL, (R14)
	MOVQ  R13, CX
	MOVQ  (AX), R13
	LEAQ  -0x1(R13), R12
	MOVQ  -0x8(R14)(R13*8), R12
	INCQ  R12
	MOVQ  R12, -0x8(R14)(R13*8)

label6:
	MOVQ    (R11), R12
	NOTQ    R12
	MOVQ    R12, (R11)
	MOVQ    DI, scratch2+80(FP)
	CMPQ    R9, $0x40
	SBBQ    R15, R15
	MOVQ    CX, R13
	MOVQ    R9, CX
	MOVQ    $-0x2, DI
	SHLQ    CL, DI
	ANDQ    R15, DI
	ANDQ    DI, R8
	BSFQ    R8, DI
	CMOVQEQ R10, DI
	MOVQ    rows+56(FP), R12
	MOVQ    scratch2+80(FP), CX
	MOVQ    DI, R9
	JMP     label5

label4:
	TESTB AL, (R14)
	MOVQ  R13, CX
	MOVQ  (AX), R13
	MOVQ  (R14)(R13*8), R12
	DECQ  R12
	MOVQ  R12, (R14)(R13*8)
	JMP   label6

done:
	RET
