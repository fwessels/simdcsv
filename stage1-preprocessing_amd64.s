//+build !noasm !appengine

// func stage1_preprocess_test(input *stage1Input, output *stage1Output)
TEXT ·stage1_preprocess_test(SB), 7, $0
	MOVQ input+0(FP), AX
	MOVQ output+8(FP), R10
	CALL ·stage1_preprocess(SB)
	RET

// func stage1_preprocess()
TEXT ·stage1_preprocess(SB), 7, $0
	MOVQ    0x8(AX), DX
	BSFQ    DX, BX
	MOVQ    0x10(AX), SI
	BSFQ    SI, DI
	MOVQ    (AX), R8
	BSFQ    R8, R9
	BSFQ    DX, DX
	MOVL    $0x40, DX
	CMOVQEQ DX, BX
	BSFQ    SI, SI
	CMOVQEQ DX, DI
	BSFQ    R8, SI
	CMOVQEQ DX, R9
	MOVQ    R10, SI
	JMP     label2

label1:
	MOVQ R8, R9

label2:
	CMPQ R9, BX
	JGE  label8
	CMPQ R9, DI
	JGE  label8
	MOVQ 0x20(AX), R8
	CMPQ R8, $0x0
	JE   label5
	CMPQ R9, $0x3f
	JNE  label7
	MOVQ 0x18(AX), R10
	ANDQ $0x1, R10
	CMPQ R10, $0x1
	JNE  label4
	MOVQ R9, CX
	MOVQ $-0x2, R8
	SHLQ CL, R8
	ANDQ R8, (AX)
	ANDQ $-0x2, 0x18(AX)

label3:
	MOVQ    (AX), R8
	BSFQ    R8, R8
	CMOVQEQ DX, R8
	JMP     label1

label4:
	CMPQ R8, $0x0

label5:
	JE    label6
	MOVQ  (AX), R8
	LEAQ  0x1(R9), CX
	CMPQ  CX, $0x40
	SBBQ  R10, R10
	MOVL  $0x1, R11
	SHLQ  CL, R11
	ANDQ  R10, R11
	TESTQ R8, R11
	JE    label6
	MOVQ  $-0x2, R9
	SHLQ  CL, R9
	ANDQ  R10, R9
	ANDQ  R8, R9
	MOVQ  R9, (AX)
	JMP   label3

label6:
	TESTB AL, (SI)
	CMPQ  R9, $0x40
	SBBQ  R8, R8
	MOVQ  R9, CX
	MOVL  $0x1, R10
	SHLQ  CL, R10
	ANDQ  R8, R10
	ORQ   R10, (SI)
	MOVQ  0x20(AX), R10
	NOTQ  R10
	MOVQ  R10, 0x20(AX)
	MOVQ  $-0x2, R10
	SHLQ  CL, R10
	ANDQ  R10, R8
	ANDQ  R8, (AX)
	JMP   label3

label7:
	CMPQ R8, $0x0
	JMP  label5

label8:
	CMPQ  BX, R9
	JGE   label12
	CMPQ  BX, DI
	JGE   label12
	CMPQ  0x20(AX), $0x0
	JNE   label11
	TESTB AL, (SI)
	CMPQ  BX, $0x40
	SBBQ  R8, R8
	MOVQ  BX, CX
	MOVL  $0x1, R10
	SHLQ  CL, R10
	ANDQ  R8, R10
	ORQ   R10, 0x8(SI)

label9:
	MOVQ    0x8(AX), R8
	CMPQ    CX, $0x40
	SBBQ    R10, R10
	MOVQ    $-0x2, R11
	SHLQ    CL, R11
	ANDQ    R10, R11
	ANDQ    R8, R11
	BSFQ    R11, BX
	MOVQ    R11, 0x8(AX)
	CMOVQEQ DX, BX

label10:
	MOVQ R9, R8
	JMP  label1

label11:
	MOVQ BX, CX
	JMP  label9

label12:
	CMPQ  DI, R9
	JGE   label15
	CMPQ  DI, BX
	JGE   label15
	CMPQ  0x20(AX), $0x0
	JNE   label14
	TESTB AL, (SI)
	CMPQ  DI, $0x40
	SBBQ  R8, R8
	MOVQ  DI, CX
	MOVL  $0x1, R10
	SHLQ  CL, R10
	ANDQ  R8, R10
	ORQ   R10, 0x10(SI)

label13:
	MOVQ    0x10(AX), R8
	CMPQ    CX, $0x40
	SBBQ    R10, R10
	MOVQ    $-0x2, R11
	SHLQ    CL, R11
	ANDQ    R10, R11
	ANDQ    R8, R11
	BSFQ    R11, R8
	MOVQ    R11, 0x10(AX)
	CMOVQEQ DX, R8
	MOVQ    R8, DI
	JMP     label10

label14:
	MOVQ DI, CX
	JMP  label13

label15:
	RET
