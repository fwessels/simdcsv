//+build !noasm !appengine

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

// func _stage2_parse_masks()
TEXT ·_stage2_parse_masks(SB), 7, $0

	MOVQ input2+104(FP), BX
	MOVQ buf_base+0(FP), AX
	MOVQ AX, INPUT_BASE(BX) // initialize input buffer base pointer

	MOVQ output2+120(FP), BX
	MOVQ rows_base+56(FP), AX
	MOVQ AX, ROWS_BASE(BX)       // initialize rows base pointer
	MOVQ columns_base+80(FP), AX
	MOVQ AX, COLUMNS_BASE(BX)    // initialize columns base pointer

	MOVQ offset+112(FP), DX
	MOVQ masks_base+24(FP), DI

	MOVQ  DX, BX
	SHRQ  $6, BX
	IMULQ $24, BX
	ADDQ  BX, DI

loop:
	//  Check whether there is still enough reserved space in the rows and columns destination buffer
	MOVQ output2+120(FP), BX
	MOVQ INDEX_OFFSET(BX), AX   // load output.index
	SHRQ $1, AX                 // divide by 2 to get number of strings (since we write two words per string)
	ADDQ $64, AX                // absolute maximum of strings to be potentially written per 64 bytes
	CMPQ AX, columns_len+88(FP)
	JGE  done                   // exit out and make sure more memory is allocated

	MOVQ LINE_OFFSET(BX), AX // load output.line
	ADDQ $64, AX             // absolute maximum of lines to be potentially written per 64 bytes
	CMPQ AX, rows_len+64(FP)
	JGE  done                // exit out and make sure more memory is allocated

	MOVQ input2+104(FP), SI

	MOVQ (DI), BX

	// are we processing the last 64-bytes?
	MOVQ DX, AX
	ADDQ $0x40, AX
	CMPQ AX, buf_len+8(FP)
	JLE  notLastZWord

	// Check if we need to OR in closing delimiter into last delimiter mask
	// We only do this the buffer is not already terminated with a delimiter
	MOVQ lastCharIsDelimiter+48(FP), CX
	CMPQ CX, $1
	JZ   notLastZWord
	MOVQ buf_len+8(FP), CX
	ANDQ $0x3f, CX
	MOVQ $1, AX
	SHLQ CX, AX
	ORQ  AX, BX

notLastZWord:
	MOVQ BX, INPUT_STAGE2_DELIMITER_MASK(SI)

	// separator mask
	MOVQ 8(DI), CX
	MOVQ CX, INPUT_STAGE2_SEPARATOR_MASK(SI)

	// quote mask
	MOVQ 16(DI), CX
	MOVQ CX, INPUT_STAGE2_QUOTE_MASK(SI)
	ADDQ $24, DI

	PUSHQ DI
	PUSHQ DX
	MOVQ  offset+112(FP), DI
	MOVQ  output2+120(FP), R9
	MOVQ  input2+104(FP), DX
	CALL  ·stage2_parse(SB)
	POPQ  DX
	POPQ  DI

	ADDQ $0x40, offset+112(FP)
	ADDQ $0x40, DX
	CMPQ DX, buf_len+8(FP)
	JLT  loop
	JNZ  done                  // in case we end exactly on a 64-byte boundary, check if we need to add a delimiter

addTrailingDelimiter:
	// simulate a last "trailing" delimiter, but only
	// if the buffer is not already terminated by a delimiter
	MOVQ lastCharIsDelimiter+48(FP), CX
	CMPQ CX, $1
	JZ   done

	MOVQ input2+104(FP), SI
	MOVQ $1, CX                              // first bit marks first char is delimiter
	MOVQ CX, INPUT_STAGE2_DELIMITER_MASK(SI)
	MOVQ $0, CX
	MOVQ CX, INPUT_STAGE2_SEPARATOR_MASK(SI)
	MOVQ CX, INPUT_STAGE2_QUOTE_MASK(SI)

	MOVQ offset+112(FP), DI
	MOVQ output2+120(FP), R9

	PUSHQ DX
	MOVQ  input2+104(FP), DX
	CALL  ·stage2_parse(SB)
	POPQ  DX

done:
	VZEROUPPER
	MOVQ DX, processed+128(FP)
	RET
