//+build !noasm !appengine

// See Input struct
#define INPUT_BASE   0x38

// See Output struct
#define COLUMNS_BASE 0x0
#define INDEX_OFFSET 0x8
#define ROWS_BASE    0x10
#define LINE_OFFSET  0x18

#define CREATE_MASK(Y1, Y2, R1, R2) \
	VPMOVMSKB Y1, R1  \
	VPMOVMSKB Y2, R2  \
	SHLQ      $32, R2 \
	ORQ       R1, R2

#define MASK_TRAILING_BYTES(MAX, Y) \
	LEAQ    MASKTABLE<>(SB), AX \
	MOVQ    $MAX, BX            \
	SUBQ    CX, BX              \
	VMOVDQU (AX)(BX*1), Y10     \ // Load mask
	VPAND   Y10, Y, Y           \ // Mask message

DATA MASKTABLE<>+0x000(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x008(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x010(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x018(SB)/8, $0x00ffffffffffffff
DATA MASKTABLE<>+0x020(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x028(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x030(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x038(SB)/8, $0x0000000000000000
GLOBL MASKTABLE<>(SB), 8, $64

// func _stage2_parse_buffer()
TEXT ·_stage2_parse_buffer(SB), 7, $0

	MOVQ         delimiterChar+80(FP), AX // get character for delimiter
	MOVQ         AX, X4
	VPBROADCASTB X4, Y4
	MOVQ         separatorChar+88(FP), AX // get character for separator
	MOVQ         AX, X5
	VPBROADCASTB X5, Y5
	MOVQ         quoteChar+96(FP), AX     // get character for quote
	MOVQ         AX, X6
	VPBROADCASTB X6, Y6

	MOVQ input+104(FP), BX
	MOVQ buf+0(FP), AX
	MOVQ AX, INPUT_BASE(BX) // initialize input buffer base pointer

	MOVQ output+120(FP), BX
	MOVQ rows_base+32(FP), AX
	MOVQ AX, ROWS_BASE(BX)       // initialize rows base pointer
	MOVQ columns_base+56(FP), AX
	MOVQ AX, COLUMNS_BASE(BX)    // initialize columns base pointer

	MOVQ offset+112(FP), DX

loop:
	//  Check whether there is still enough reserved space in the rows and columns destination buffer
	MOVQ output+120(FP), BX
	MOVQ INDEX_OFFSET(BX), AX   // load output.index
	SHRQ $1, AX                 // divide by 2 to get number of strings (since we write two words per string)
	ADDQ $64, AX                // absolute maximum of strings to be potentially written per 64 bytes
	CMPQ AX, columns_len+64(FP)
	JGE  done                   // exit out and make sure more memory is allocated

	MOVQ LINE_OFFSET(BX), AX // load output.line
	ADDQ $64, AX             // absolute maximum of lines to be potentially written per 64 bytes
	CMPQ AX, rows_len+40(FP)
	JGE  done                // exit out and make sure more memory is allocated

	MOVQ buf+0(FP), DI
	MOVQ input+104(FP), SI

	// do we need to do a partial load?
	MOVQ DX, CX
	ADDQ $0x40, CX
	CMPQ CX, buf_len+8(FP)
	JGT  partialLoad

	VMOVDQU (DI)(DX*1), Y8     // load low 32-bytes
	VMOVDQU 0x20(DI)(DX*1), Y9 // load high 32-bytes

joinAfterPartialLoad:
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

	MOVQ offset+112(FP), DI
	MOVQ output+120(FP), R9

	PUSHQ DX
	MOVQ  input+104(FP), DX
	CALL  ·stage2_parse(SB)
	POPQ  DX

	ADDQ $0x40, offset+112(FP)
	ADDQ $0x40, DX
	CMPQ DX, buf_len+8(FP)
	JLT  loop
	JNZ  done                  // in case we end exactly on a 64-byte boundary, check if we need to add a delimiter

addTrailingDelimiter:
	// simulate a last "trailing" delimiter, but only
	// if the buffer is not already terminated by a delimiter
	MOVQ lastCharIsDelimiter+24(FP), CX
	CMPQ CX, $1
	JZ   done

	MOVQ input+104(FP), SI
	MOVQ $1, CX            // first bit marks first char is delimiter
	MOVQ CX, 8(SI)
	MOVQ $0, CX
	MOVQ CX, 0(SI)
	MOVQ CX, 16(SI)

	MOVQ input+104(FP), DX
	MOVQ offset+112(FP), DI
	MOVQ output+120(FP), R9
	CALL ·stage2_parse(SB)

done:
	VZEROUPPER
	MOVQ DX, processed+128(FP)
	RET

partialLoad:
	// do a partial load and mask out bytes after the end of the message with whitespace
	VMOVDQU (DI)(DX*1), Y8 // always load low 32-bytes

	MOVQ buf_len+8(FP), CX
	ANDQ $0x3f, CX
	CMPQ CX, $0x20
	JGE  maskingHigh

	// perform masking on low 32-bytes
	MASK_TRAILING_BYTES(0x1f, Y8)
	VPXOR Y9, Y9, Y9           // clear upper 32-bytes
	JMP   joinAfterPartialLoad

maskingHigh:
	// perform masking on high 32-bytes
	VMOVDQU 0x20(DI)(DX*1), Y9   // load high 32-bytes
	MASK_TRAILING_BYTES(0x3f, Y9)
	JMP     joinAfterPartialLoad

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
	JGE   label5
	CMPQ  SI, R10
	JGE   label5
	CMPQ  0x18(DX), $0x0
	JNE   label4
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
	MOVQ  0x8(R9), R11
	LEAQ  0x1(R11), AX
	MOVQ  AX, 0x8(R9)
	MOVQ  (R9), R12
	TESTB AL, (R12)
	MOVQ  0x28(R9), R13
	MOVQ  SI, CX
	SUBQ  R13, SI
	ADDQ  DI, SI
	SUBQ  0x20(R9), SI
	MOVQ  SI, 0x8(R12)(R11*8)
	INCQ  0x8(R9)
	LEAQ  (CX)(DI*1), SI
	LEAQ  0x1(SI), SI
	MOVQ  SI, 0x20(R9)
	MOVQ  $0x0, 0x28(R9)
	LEAQ  (CX)(DI*1), SI
	MOVQ  SI, 0x20(DX)

label3:
	MOVQ    (DX), SI
	CMPQ    CX, $0x40
	SBBQ    R11, R11
	MOVQ    $-0x2, R12
	SHLQ    CL, R12
	ANDQ    R11, R12
	ANDQ    SI, R12
	BSFQ    R12, SI
	MOVQ    R12, (DX)
	CMOVQEQ BX, SI
	JMP     label1

label4:
	MOVQ SI, CX
	JMP  label3

label5:
	CMPQ  R8, SI
	JGE   label11
	CMPQ  R8, R10
	JGE   label11
	CMPQ  0x18(DX), $0x0
	JNE   label10
	MOVQ  0x28(DX), R11
	TESTQ R11, R11
	JBE   label6
	INCQ  R11
	LEAQ  (R8)(DI*1), R12
	CMPQ  R12, R11
	JE    label6
	CMPQ  0x30(DX), $0x0
	JNE   label6
	MOVQ  R12, 0x30(DX)

label6:
	MOVQ  $0x0, 0x28(DX)
	MOVQ  (R9), R11
	TESTB AL, (R11)
	MOVQ  0x8(R9), AX
	MOVQ  0x38(DX), R12
	ADDQ  0x20(R9), R12
	MOVQ  R12, (R11)(AX*8)
	MOVQ  0x8(R9), R11
	LEAQ  0x1(R11), AX
	MOVQ  AX, 0x8(R9)
	MOVQ  (R9), R12
	TESTB AL, (R12)
	MOVQ  0x28(R9), R13
	MOVQ  R8, CX
	SUBQ  R13, R8
	ADDQ  DI, R8
	SUBQ  0x20(R9), R8
	MOVQ  R8, 0x8(R12)(R11*8)
	MOVQ  0x8(R9), AX
	LEAQ  0x1(AX), R8
	MOVQ  R8, 0x8(R9)
	LEAQ  (CX)(DI*1), R11
	LEAQ  0x1(R11), R11
	MOVQ  R11, 0x20(R9)
	MOVQ  $0x0, 0x28(R9)
	LEAQ  (CX)(DI*1), R11
	SHRQ  $0x1, R8
	MOVQ  0x30(R9), R12
	SUBQ  R12, R8
	CMPQ  R8, $0x1
	JNE   label9
	MOVQ  (R9), R8
	TESTB AL, (R8)
	MOVQ  (R8)(AX*8), R8
	TESTQ R8, R8
	JNE   label9

label7:
	MOVQ 0x8(R9), R8
	SHRQ $0x1, R8
	MOVQ R8, 0x30(R9)
	MOVQ R11, 0x20(DX)

label8:
	MOVQ    0x8(DX), R8
	CMPQ    CX, $0x40
	SBBQ    R11, R11
	MOVQ    $-0x2, R12
	SHLQ    CL, R12
	ANDQ    R11, R12
	ANDQ    R8, R12
	BSFQ    R12, R8
	MOVQ    R12, 0x8(DX)
	CMOVQEQ BX, R8
	JMP     label1

label9:
	MOVQ  0x10(R9), R8
	TESTB AL, (R8)
	MOVQ  0x18(R9), AX
	MOVQ  R12, (R8)(AX*8)
	MOVQ  0x18(R9), R8
	LEAQ  0x1(R8), AX
	MOVQ  AX, 0x18(R9)
	MOVQ  0x10(R9), R12
	TESTB AL, (R12)
	MOVQ  0x8(R9), R13
	SHRQ  $0x1, R13
	SUBQ  0x30(R9), R13
	MOVQ  R13, 0x8(R12)(R8*8)
	INCQ  0x18(R9)
	JMP   label7

label10:
	MOVQ R8, CX
	JMP  label8

label11:
	CMPQ R10, SI
	JGE  label15
	CMPQ R10, R8
	JGE  label15
	CMPQ 0x18(DX), $0x0
	JNE  label14
	MOVQ 0x20(DX), R11
	INCQ R11
	LEAQ (R10)(DI*1), R12
	CMPQ R11, R12
	JE   label12
	CMPQ 0x30(DX), $0x0
	JNE  label12
	MOVQ R12, 0x30(DX)

label12:
	INCQ 0x20(R9)

label13:
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

label14:
	INCQ 0x28(R9)
	LEAQ (R10)(DI*1), R11
	MOVQ R11, 0x28(DX)
	JMP  label13

label15:
	RET
