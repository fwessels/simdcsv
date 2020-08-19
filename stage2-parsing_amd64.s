//+build !noasm !appengine

#define CREATE_MASK(Y1, Y2, R1, R2) \
	VPMOVMSKB Y1, R1  \
	VPMOVMSKB Y2, R2  \
	SHLQ      $32, R2 \
	ORQ       R1, R2

// func _stage2_parse_buffer()
TEXT ·_stage2_parse_buffer(SB), 7, $0

	MOVQ         delimiterChar+32(FP), AX // get character for delimiter
	MOVQ         AX, X4
	VPBROADCASTB X4, Y4
	MOVQ         separatorChar+40(FP), AX // get character for separator
	MOVQ         AX, X5
	VPBROADCASTB X5, Y5
	MOVQ         quoteChar+48(FP), AX     // get character for quote
	MOVQ         AX, X6
	VPBROADCASTB X6, Y6

	XORQ DX, DX

	// Check whether it is necessary to adjust pointer for first string element
	MOVQ output+72(FP), R9
	MOVQ (R9), R9          // columnns pointer
	CMPQ (R9), $0
	JNZ  loop              // skip setting first element
	MOVQ buf+0(FP), DI
	MOVQ DI, (R9)

loop:
	MOVQ buf+0(FP), DI
	MOVQ input+56(FP), SI

	VMOVDQU (DI)(DX*1), Y8     // load low 32-bytes
	VMOVDQU 0x20(DI)(DX*1), Y9 // load high 32-bytes

	// delimiter mask
	VPCMPEQB Y8, Y4, Y10
	VPCMPEQB Y9, Y4, Y11
	CREATE_MASK(Y10, Y11, AX, BX)

	// are we processing the last 64-bytes?
	MOVQ DX, AX
	ADDQ $0x40, AX
	CMPQ AX, buf_len+8(FP)
	JLE  notLastZWord

	// Check if we need to OR in closing delimiter into last delimiter mask
	// We only do this the buffer is not already terminated with a delimiter
	MOVQ lastCharIsDelimiter+24(FP), CX
	CMPQ CX, $1
	JZ   notLastZWord
	MOVQ buf_len+8(FP), CX
	ANDQ $0x3f, CX
	MOVQ $1, AX
	SHLQ CX, AX
	ORQ  AX, BX

notLastZWord:
	MOVQ BX, 8(SI)

	// separator mask
	VPCMPEQB Y8, Y5, Y10
	VPCMPEQB Y9, Y5, Y11
	CREATE_MASK(Y10, Y11, AX, CX)
	MOVQ     CX, 0(SI)

	// quote mask
	VPCMPEQB Y8, Y6, Y10
	VPCMPEQB Y9, Y6, Y11
	CREATE_MASK(Y10, Y11, AX, CX)
	MOVQ     CX, 16(SI)

	MOVQ offset+64(FP), DI
	MOVQ output+72(FP), R9

	PUSHQ DX
	MOVQ  input+56(FP), DX
	CALL  ·stage2_parse(SB)
	POPQ  DX

	ADDQ $0x40, offset+64(FP)
	ADDQ $0x40, DX
	CMPQ DX, buf_len+8(FP)
	JLT  loop
	JZ   addTrailingDelimiter // in case we end exactly on a 64-byte boundary, check if we need to add a delimiter

	VZEROUPPER
	RET

addTrailingDelimiter:
	// simulate a last "trailing" delimiter, but only
	// if the buffer is not already terminated by a delimiter
	MOVQ lastCharIsDelimiter+24(FP), CX
	CMPQ CX, $1
	JZ   done

	MOVQ input+56(FP), SI
	MOVQ $1, CX           // first bit marks first char is delimiter
	MOVQ CX, 8(SI)
	MOVQ $0, CX
	MOVQ CX, 0(SI)
	MOVQ CX, 16(SI)

	MOVQ input+56(FP), DX
	MOVQ offset+64(FP), DI
	MOVQ output+72(FP), R9
	CALL ·stage2_parse(SB)

done:
	VZEROUPPER
	RET

// func stage2_parse_test(input *Input, offset uint64, output *Output)
TEXT ·stage2_parse_test(SB), 7, $0
	MOVQ input+0(FP), DX
	MOVQ offset+8(FP), DI
	MOVQ output+16(FP), R9
	CALL ·stage2_parse(SB)
	RET

// func stage2_parse()
TEXT ·stage2_parse(SB), 7, $0
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
	CMPQ  R12, R11
	JE    label2
	CMPQ  0x30(DX), $0x0
	JNE   label2
	MOVQ  R12, 0x30(DX)

label2:
	MOVQ  $0x0, 0x28(DX)
	MOVQ  (R9), R11
	TESTB AL, (R11)
	MOVQ  0x8(R9), AX
	MOVQ  0x38(DX), R12
	ADDQ  0x20(R9), R12
	MOVQ  R12, (R11)(AX*8)
	MOVQ  0x8(R9), AX
	LEAQ  0x1(AX), R11
	MOVQ  R11, 0x8(R9)
	MOVQ  (R9), R12
	TESTB AL, (R12)
	MOVQ  0x38(DX), R13
	SUBQ  0x28(R9), R13
	ADDQ  SI, R13
	ADDQ  DI, R13
	MOVQ  (R12)(AX*8), R14
	SUBQ  R14, R13
	MOVQ  R13, 0x8(R12)(AX*8)
	INCQ  0x8(R9)
	LEAQ  (SI)(DI*1), R11
	LEAQ  0x1(R11), R11
	MOVQ  R11, 0x20(R9)
	MOVQ  $0x0, 0x28(R9)
	LEAQ  (SI)(DI*1), R11
	MOVQ  R11, 0x20(DX)

label3:
	MOVQ    (DX), R11
	CMPQ    SI, $0x40
	SBBQ    R12, R12
	MOVQ    SI, CX
	MOVQ    $-0x2, R13
	SHLQ    CL, R13
	ANDQ    R13, R12
	ANDQ    R11, R12
	BSFQ    R12, SI
	MOVQ    R12, (DX)
	CMOVQEQ BX, SI
	JMP     label1

label4:
	CMPQ  R8, SI
	JGE   label9
	CMPQ  R8, R10
	JGE   label9
	CMPQ  0x18(DX), $0x0
	JNE   label7
	MOVQ  0x28(DX), R11
	TESTQ R11, R11
	JBE   label5
	INCQ  R11
	LEAQ  (R8)(DI*1), R12
	CMPQ  R12, R11
	JE    label5
	CMPQ  0x30(DX), $0x0
	JNE   label5
	MOVQ  R12, 0x30(DX)

label5:
	MOVQ  $0x0, 0x28(DX)
	MOVQ  (R9), R11
	TESTB AL, (R11)
	MOVQ  0x8(R9), AX
	MOVQ  0x38(DX), R12
	ADDQ  0x20(R9), R12
	MOVQ  R12, (R11)(AX*8)
	MOVQ  0x8(R9), AX
	LEAQ  0x1(AX), R11
	MOVQ  R11, 0x8(R9)
	MOVQ  (R9), R12
	TESTB AL, (R12)
	MOVQ  0x38(DX), R13
	SUBQ  0x28(R9), R13
	ADDQ  R8, R13
	ADDQ  DI, R13
	MOVQ  (R12)(AX*8), R14
	SUBQ  R14, R13
	MOVQ  R13, 0x8(R12)(AX*8)
	MOVQ  0x8(R9), AX
	LEAQ  0x1(AX), R11
	MOVQ  R11, 0x8(R9)
	LEAQ  (R8)(DI*1), R12
	LEAQ  0x1(R12), R12
	MOVQ  R12, 0x20(R9)
	MOVQ  $0x0, 0x28(R9)
	LEAQ  (R8)(DI*1), R12
	SHRQ  $0x1, R11
	SUBQ  0x30(R9), R11
	CMPQ  R11, $0x1
	JNE   label8
	MOVQ  (R9), R13
	TESTB AL, (R13)
	MOVQ  (R13)(AX*8), R13
	TESTQ R13, R13
	JNE   label8

label6:
	MOVQ 0x8(R9), R11
	SHRQ $0x1, R11
	MOVQ R11, 0x30(R9)
	MOVQ R12, 0x20(DX)

label7:
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

label8:
	MOVQ  0x10(R9), R13
	TESTB AL, (R13)
	MOVQ  0x18(R9), AX
	MOVQ  R11, (R13)(AX*8)
	INCQ  0x18(R9)
	JMP   label6

label9:
	CMPQ R10, SI
	JGE  label13
	CMPQ R10, R8
	JGE  label13
	CMPQ 0x18(DX), $0x0
	JNE  label12
	MOVQ 0x20(DX), R11
	INCQ R11
	LEAQ (R10)(DI*1), R12
	CMPQ R11, R12
	JE   label10
	CMPQ 0x30(DX), $0x0
	JNE  label10
	MOVQ R12, 0x30(DX)

label10:
	INCQ 0x20(R9)

label11:
	MOVQ    0x18(DX), R11
	NOTQ    R11
	MOVQ    R11, 0x18(DX)
	MOVQ    0x10(DX), R11
	CMPQ    R10, $0x40
	SBBQ    R12, R12
	MOVQ    R10, CX
	MOVQ    $-0x2, R13
	SHLQ    CL, R13
	ANDQ    R12, R13
	ANDQ    R11, R13
	BSFQ    R13, R11
	MOVQ    R13, 0x10(DX)
	CMOVQEQ BX, R11
	MOVQ    R11, R10
	JMP     label1

label12:
	INCQ 0x28(R9)
	LEAQ (R10)(DI*1), R11
	MOVQ R11, 0x28(DX)
	JMP  label11

label13:
	RET
