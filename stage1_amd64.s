//+build !noasm !appengine

#define CREATE_MASK(_Y1, _Y2, _R1, _R2) \
	VPMOVMSKB _Y1, _R1 \
	VPMOVMSKB _Y2, _R2 \
	SHLQ      $32, _R2 \
	ORQ       _R1, _R2

#define MASK_TRAILING_BYTES(MAX, SCRATCH1, SCRATCH2, RPOS, Y_SCRATCH, _Y) \
	LEAQ    MASKTABLE<>(SB), SCRATCH1         \
	MOVQ    $MAX, SCRATCH2                    \
	SUBQ    RPOS, SCRATCH2                    \
	VMOVDQU (SCRATCH1)(SCRATCH2*1), Y_SCRATCH \ // Load mask
	VPAND   Y_SCRATCH, _Y, _Y                 // Mask message

DATA MASKTABLE<>+0x000(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x008(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x010(SB)/8, $0xffffffffffffffff
DATA MASKTABLE<>+0x018(SB)/8, $0x00ffffffffffffff
DATA MASKTABLE<>+0x020(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x028(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x030(SB)/8, $0x0000000000000000
DATA MASKTABLE<>+0x038(SB)/8, $0x0000000000000000
GLOBL MASKTABLE<>(SB), 8, $64

#define ADD_TRAILING_NEWLINE \
	MOVQ $1, AX \
	SHLQ CX, AX \ // only lower 6 bits are taken into account, which is good for current and next YMM words
	ORQ  AX, BX

// See stage1Input struct
#define QUOTE_MASK_IN           0
#define SEPARATOR_MASK_IN       8
#define CARRIAGE_RETURN_MASK_IN 16
#define QUOTE_MASK_IN_NEXT      24
#define QUOTED                  32
#define NEWLINE_MASK_IN         40
#define NEWLINE_MASK_IN_NEXT    48

// See stage1Output struct
#define QUOTE_MASK_OUT            0
#define SEPARATOR_MASK_OUT        8
#define CARRIAGE_RETURN_MASK_OUT  16
#define NEEDS_POST_PROCESSING_OUT 24

// Offsets for  masks slice
#define MASKS_NEWLINE_OFFSET   0
#define MASKS_SEPARATOR_OFFSET 8
#define MASKS_QUOTE_OFFSET     16
#define MASKS_ELEM_SIZE        24

#define Y_QUOTE_CHAR  Y5
#define Y_SEPARATOR   Y4
#define Y_CARRIAGE_R  Y3
#define Y_NEWLINE     Y2

// func stage1_preprocess_buffer(buf []byte, separatorChar uint64, input1 *stage1Input, output1 *stage1Output, postProc *[]uint64, offset uint64, masks []uint64) (processed, masksWritten uint64)
TEXT ·stage1_preprocess_buffer(SB), 7, $0

	MOVQ         $0x0a, AX                // character for newline
	MOVQ         AX, X2
	VPBROADCASTB X2, Y_NEWLINE
	MOVQ         $0x0d, AX                // character for carriage return
	MOVQ         AX, X3
	VPBROADCASTB X3, Y_CARRIAGE_R
	MOVQ         separatorChar+24(FP), AX // get character for separator
	MOVQ         AX, X4
	VPBROADCASTB X4, Y_SEPARATOR
	MOVQ         $0x22, AX                // character for quote
	MOVQ         AX, X5
	VPBROADCASTB X5, Y_QUOTE_CHAR

	MOVQ buf+0(FP), DI
	MOVQ offset+56(FP), DX
	MOVQ masks_base+64(FP), R11
	MOVQ $6, R12                // initialize indexing reg at 6, so we can compare to length of slice

	MOVQ DX, CX
	ADDQ $0x40, CX
	CMPQ CX, buf_len+8(FP)
	JLE  fullLoadPrologue
	MOVQ buf_len+8(FP), BX
	CALL ·partialLoad(SB)
	JMP  skipFullLoadPrologue

fullLoadPrologue:
	VMOVDQU (DI)(DX*1), Y6     // load low 32-bytes
	VMOVDQU 0x20(DI)(DX*1), Y7 // load high 32-bytes

skipFullLoadPrologue:
	MOVQ input1+32(FP), SI

	// quote mask
	VPCMPEQB Y6, Y_QUOTE_CHAR, Y0
	VPCMPEQB Y7, Y_QUOTE_CHAR, Y1
	CREATE_MASK(Y0, Y1, AX, CX)
	MOVQ     CX, QUOTE_MASK_IN_NEXT(SI) // store in next slot, so that it gets copied back

	// newline
	VPCMPEQB Y6, Y_NEWLINE, Y0
	VPCMPEQB Y7, Y_NEWLINE, Y1
	CREATE_MASK(Y0, Y1, AX, BX)

	MOVQ buf_len+8(FP), CX
	CMPQ CX, $64
	JGE  skipAddTrailingNewlinePrologue
	ADD_TRAILING_NEWLINE

skipAddTrailingNewlinePrologue:
	MOVQ BX, NEWLINE_MASK_IN_NEXT(SI)                           // store in next slot, so that it gets copied back
	MOVQ BX, MASKS_NEWLINE_OFFSET-MASKS_ELEM_SIZE*2(R11)(R12*8)

loop:
	VMOVDQU Y6, Y8 // get low 32-bytes
	VMOVDQU Y7, Y9 // get high 32-bytes

	MOVQ input1+32(FP), SI

	// copy next masks to current slot (for quote mask and newline mask)
	MOVQ QUOTE_MASK_IN_NEXT(SI), CX
	MOVQ CX, QUOTE_MASK_IN(SI)
	MOVQ NEWLINE_MASK_IN_NEXT(SI), CX
	MOVQ CX, NEWLINE_MASK_IN(SI)

	// separator mask
	VPCMPEQB Y8, Y_SEPARATOR, Y0
	VPCMPEQB Y9, Y_SEPARATOR, Y1
	CREATE_MASK(Y0, Y1, AX, CX)
	MOVQ     CX, SEPARATOR_MASK_IN(SI)

	// carriage return
	VPCMPEQB Y8, Y_CARRIAGE_R, Y0
	VPCMPEQB Y9, Y_CARRIAGE_R, Y1
	CREATE_MASK(Y0, Y1, AX, CX)
	MOVQ     CX, CARRIAGE_RETURN_MASK_IN(SI)

	// do we need to do a partial load?
	MOVQ DX, CX
	ADDQ $0x80, CX
	CMPQ CX, buf_len+8(FP)
	JLE  fullLoad
	MOVQ buf_len+8(FP), BX
	CALL ·partialLoad(SB)
	JMP  skipFullLoad

fullLoad:
	// load next pair of YMM words
	VMOVDQU 0x40(DI)(DX*1), Y6 // load low 32-bytes of next pair
	VMOVDQU 0x60(DI)(DX*1), Y7 // load high 32-bytes of next pair

skipFullLoad:
	VPCMPEQB Y6, Y_QUOTE_CHAR, Y0
	VPCMPEQB Y7, Y_QUOTE_CHAR, Y1
	CREATE_MASK(Y0, Y1, AX, CX)
	MOVQ     CX, QUOTE_MASK_IN_NEXT(SI)

	// quote mask next for next YMM word
	VPCMPEQB Y6, Y_NEWLINE, Y0
	VPCMPEQB Y7, Y_NEWLINE, Y1
	CREATE_MASK(Y0, Y1, AX, BX)

	// Write unaltered newline mask into next slot already
	MOVQ BX, MASKS_NEWLINE_OFFSET-MASKS_ELEM_SIZE(R11)(R12*8)

	MOVQ buf_len+8(FP), CX
	SUBQ DX, CX
	JLT  skipAddTrailingNewline
	ADD_TRAILING_NEWLINE

skipAddTrailingNewline:
	MOVQ BX, NEWLINE_MASK_IN_NEXT(SI)

	PUSHQ R12
	PUSHQ R11
	PUSHQ DI
	PUSHQ DX
	MOVQ  input1+32(FP), AX
	MOVQ  output1+40(FP), R10
	CALL  ·stage1_preprocess(SB)
	POPQ  DX
	POPQ  DI
	POPQ  R11
	POPQ  R12

	MOVQ output1+40(FP), R10

	// write out masks to slice
	MOVQ QUOTE_MASK_OUT(R10), AX
	MOVQ AX, MASKS_QUOTE_OFFSET-MASKS_ELEM_SIZE*2(R11)(R12*8)
	MOVQ SEPARATOR_MASK_OUT(R10), AX
	MOVQ AX, MASKS_SEPARATOR_OFFSET-MASKS_ELEM_SIZE*2(R11)(R12*8)
	MOVQ CARRIAGE_RETURN_MASK_OUT(R10), AX
	ORQ  AX, MASKS_NEWLINE_OFFSET-MASKS_ELEM_SIZE*2(R11)(R12*8)
	ADDQ $3, R12

	MOVQ output1+40(FP), R10
	CMPQ NEEDS_POST_PROCESSING_OUT(R10), $1
	JNZ  unmodified

	MOVQ postProc+48(FP), AX
	MOVQ 0(AX), BX
	MOVQ 8(AX), CX
	MOVQ DX, (BX)(CX*8)
	INCQ 8(AX)
	INCQ CX
	ADDQ $0x40, DX
	CMPQ CX, 16(AX)          // slice is full?
	JGE  exit
	SUBQ $0x40, DX

unmodified:
	CMPQ R12, masks_len+72(FP) // still space in masks slice?
	JGE  exit

	ADDQ $0x40, DX
	CMPQ DX, buf_len+8(FP)
	JLT  loop

exit:
	VZEROUPPER
	MOVQ DX, processed+88(FP)
	SUBQ $6, R12
	MOVQ R12, masksWritten+96(FP)
	RET

// CX = base for loading
// BX = buf_len+8(FP)
TEXT ·partialLoad(SB), 7, $0
	VPXOR Y6, Y6, Y6 // clear lower 32-bytes
	VPXOR Y7, Y7, Y7 // clear upper 32-bytes

	SUBQ $0x40, CX

	// check whether we need to load at all?
	CMPQ CX, BX
	JGT  partialLoadDone

	// do a partial load and mask out bytes after the end of the message with whitespace
	VMOVDQU (DI)(CX*1), Y6 // always load low 32-bytes

	ANDQ $0x3f, BX
	CMPQ BX, $0x20
	JGE  maskingHigh

	// perform masking on low 32-bytes
	MASK_TRAILING_BYTES(0x1f, AX, CX, BX, Y0, Y6)
	RET

maskingHigh:
	VMOVDQU 0x20(DI)(CX*1), Y7 // load high 32-bytes
	MASK_TRAILING_BYTES(0x3f, AX, CX, BX, Y0, Y7)

partialLoadDone:
	RET
