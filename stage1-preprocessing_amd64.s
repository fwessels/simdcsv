//+build !noasm !appengine

#define UNPACK_BITMASK(_R1, _XR1, _YR1) \
	\ // source: https://stackoverflow.com/a/24242696
	VMOVQ        _R1, _XR1       \
	VPBROADCASTD _XR1, _YR1      \
	VPSHUFB      Y14, _YR1, _YR1 \
	VPANDN       Y15, _YR1, _YR1 \
	VPCMPEQB     Y13, _YR1, _YR1 \

// func stage1_preprocess_buffer(buf []byte, input *stage1Input, output *stage1Output)
TEXT ·stage1_preprocess_buffer(SB), 7, $0

	XORQ DX, DX

loop:
	PUSHQ DX
	MOVQ  input+24(FP), AX
	MOVQ  output+32(FP), R10
	CALL  ·stage1_preprocess(SB)
	POPQ  DX

	LEAQ    ANDMASK<>(SB), AX
	VMOVDQU (AX), Y15
	LEAQ    SHUFMASK<>(SB), AX
	VMOVDQU (AX), Y14
	VPXOR   Y13, Y13, Y13
	MOVQ         $0x2, AX // preprocessedSeparator
	MOVQ         AX, X12
	VPBROADCASTB X12, Y12
	MOVQ         $0x3, AX // preprocessedQuote
	MOVQ         AX, X11
	VPBROADCASTB X11, Y11
	MOVQ         $0x0a, AX // new line
	MOVQ         AX, X10
	VPBROADCASTB X10, Y10

	MOVQ    buf+0(FP), DI
	VMOVDQU (DI)(DX*1), Y8     // load low 32-bytes
	VMOVDQU 0x20(DI)(DX*1), Y9 // load high 32-bytes

    // Replace quotes
	MOVQ output+32(FP), R10
	MOVQ 0x0(R10), AX
    UNPACK_BITMASK(AX, X0, Y0)
    SHRQ $32, AX
    UNPACK_BITMASK(AX, X1, Y1)
	VPBLENDVB Y0, Y11, Y8, Y8
	VPBLENDVB Y1, Y11, Y9, Y9

    // Replace separators
	MOVQ output+32(FP), R10
	MOVQ 0x8(R10), AX
    UNPACK_BITMASK(AX, X0, Y0)
    SHRQ $32, AX
    UNPACK_BITMASK(AX, X1, Y1)
	VPBLENDVB Y0, Y12, Y8, Y8
	VPBLENDVB Y1, Y12, Y9, Y9

    // Replace carriage returns
	MOVQ output+32(FP), R10
	MOVQ 0x10(R10), AX
    UNPACK_BITMASK(AX, X0, Y0)
    SHRQ $32, AX
    UNPACK_BITMASK(AX, X1, Y1)
	VPBLENDVB Y0, Y10, Y8, Y8
	VPBLENDVB Y1, Y10, Y9, Y9

MOVQ debug+40(FP), AX
VMOVDQU Y0, (AX)

	VMOVDQU Y8, (DI)(DX*1)
	VMOVDQU Y9, 0x20(DI)(DX*1)

//	ADDQ $0x40, DX
//	CMPQ DX, $0x40*10
//	JLT  loop

	RET

DATA SHUFMASK<>+0x000(SB)/8, $0x0000000000000000
DATA SHUFMASK<>+0x008(SB)/8, $0x0101010101010101
DATA SHUFMASK<>+0x010(SB)/8, $0x0202020202020202
DATA SHUFMASK<>+0x018(SB)/8, $0x0303030303030303
GLOBL SHUFMASK<>(SB), 8, $32

DATA ANDMASK<>+0x000(SB)/8, $0x8040201008040201
DATA ANDMASK<>+0x008(SB)/8, $0x8040201008040201
DATA ANDMASK<>+0x010(SB)/8, $0x8040201008040201
DATA ANDMASK<>+0x018(SB)/8, $0x8040201008040201
GLOBL ANDMASK<>(SB), 8, $32


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
