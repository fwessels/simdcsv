//+build !noasm !appengine

#include "common.h"

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

// See Input struct
#define INPUT_BASE   0x38

// See Output struct
#define COLUMNS_BASE 0x0
#define INDEX_OFFSET 0x8
#define ROWS_BASE    0x10
#define LINE_OFFSET  0x18

#define INPUT_STAGE2_SEPARATOR_MASK 0
#define INPUT_STAGE2_DELIMITER_MASK 8
#define INPUT_STAGE2_QUOTE_MASK     16

#define Y_QUOTE_CHAR  Y5
#define Y_SEPARATOR   Y4
#define Y_CARRIAGE_R  Y3
#define Y_NEWLINE     Y2
#define Y_DELIMITER   Y_NEWLINE

// stages_combined_buffer(buf []byte, separatorChar uint64, input1 *stage1Input, output1 *stage1Output, postProc *[]uint64, offset uint64, input2 *Input, output2 *OutputAsm, lastCharIsDelimiter uint64, rows []uint64, columns []string) (processed uint64)
TEXT ·stages_combined_buffer(SB), 7, $0

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

	MOVQ input2+64(FP), BX
	MOVQ buf+0(FP), AX
	MOVQ AX, INPUT_BASE(BX) // initialize input buffer base pointer

	MOVQ output2+72(FP), BX
	MOVQ rows_base+88(FP), AX
	MOVQ AX, ROWS_BASE(BX)        // initialize rows base pointer
	MOVQ columns_base+112(FP), AX
	MOVQ AX, COLUMNS_BASE(BX)     // initialize columns base pointer

	MOVQ offset+56(FP), DX

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
	MOVQ BX, NEWLINE_MASK_IN_NEXT(SI) // store in next slot, so that it gets copied back

loop:
	//  Check whether there is still enough reserved space in the rows and columns destination buffer
	MOVQ output2+72(FP), BX
	MOVQ INDEX_OFFSET(BX), AX    // load output.index
	SHRQ $1, AX                  // divide by 2 to get number of strings (since we write two words per string)
	ADDQ $64, AX                 // absolute maximum of strings to be potentially written per 64 bytes
	CMPQ AX, columns_len+120(FP)
	JGE  done                    // exit out and make sure more memory is allocated

	MOVQ LINE_OFFSET(BX), AX // load output.line
	ADDQ $64, AX             // absolute maximum of lines to be potentially written per 64 bytes
	CMPQ AX, rows_len+96(FP)
	JGE  done                // exit out and make sure more memory is allocated

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

	MOVQ buf+0(FP), DI

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

	MOVQ buf_len+8(FP), CX
	SUBQ DX, CX
	JLT  skipAddTrailingNewline
	ADD_TRAILING_NEWLINE

skipAddTrailingNewline:
	MOVQ BX, NEWLINE_MASK_IN_NEXT(SI)

	PUSHQ DX
	MOVQ  input1+32(FP), AX
	MOVQ  output1+40(FP), R10
	CALL  ·stage1_preprocess(SB)
	POPQ  DX

	MOVQ output1+40(FP), R10

	MOVQ input2+64(FP), SI

	// delimiter mask
	VPCMPEQB Y8, Y_DELIMITER, Y0
	VPCMPEQB Y9, Y_DELIMITER, Y1
	CREATE_MASK(Y0, Y1, AX, BX)

	// Merge in carriage returns that are treated as newlines (and hence filtered out)
	ORQ CARRIAGE_RETURN_MASK_OUT(R10), BX

	// are we processing the last 64-bytes?
	MOVQ DX, AX
	ADDQ $0x40, AX
	CMPQ AX, buf_len+8(FP)
	JLE  notLastZWord

	// Check if we need to OR in closing delimiter into last delimiter mask
	// We only do this the buffer is not already terminated with a delimiter
	MOVQ lastCharIsDelimiter+80(FP), CX
	CMPQ CX, $1
	JZ   notLastZWord
	MOVQ buf_len+8(FP), CX
	ANDQ $0x3f, CX
	MOVQ $1, AX
	SHLQ CX, AX
	ORQ  AX, BX

notLastZWord:
	MOVQ BX, INPUT_STAGE2_DELIMITER_MASK(SI)

	// Set bitmask for quotes
	MOVQ QUOTE_MASK_OUT(R10), AX
	MOVQ AX, INPUT_STAGE2_QUOTE_MASK(SI)

	// Set bitmask for separators
	MOVQ SEPARATOR_MASK_OUT(R10), AX
	MOVQ AX, INPUT_STAGE2_SEPARATOR_MASK(SI)

	MOVQ offset+56(FP), DI
	MOVQ output2+72(FP), R9

	PUSHQ DX
	MOVQ  input2+64(FP), DX
	CALL  ·stage2_parse(SB)
	POPQ  DX

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
	ADDQ $0x40, offset+56(FP)
	ADDQ $0x40, DX
	CMPQ DX, buf_len+8(FP)
	JLT  loop
	JNZ  done                 // in case we end exactly on a 64-byte boundary, check if we need to add a delimiter

addTrailingDelimiter:
	// simulate a last "trailing" delimiter, but only
	// if the buffer is not already terminated by a delimiter
	MOVQ lastCharIsDelimiter+80(FP), CX
	CMPQ CX, $1
	JZ   done

	MOVQ input2+64(FP), SI
	MOVQ $1, CX            // first bit marks first char is delimiter
	MOVQ CX, 8(SI)
	MOVQ $0, CX
	MOVQ CX, 0(SI)
	MOVQ CX, 16(SI)

	MOVQ offset+56(FP), DI
	MOVQ output2+72(FP), R9

	PUSHQ DX
	MOVQ  input2+64(FP), DX
	CALL  ·stage2_parse(SB)
	POPQ  DX

done:
exit:
	VZEROUPPER
	MOVQ DX, processed+136(FP)
	RET
