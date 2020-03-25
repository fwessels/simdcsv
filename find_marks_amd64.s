//+build !noasm !appengine

// find_marks_in_slice(msg []byte, indexes *[INDEXES_SIZE]uint32, indexes_length, carried, position *uint64) (out uint64)
TEXT ·find_marks_in_slice(SB), 7, $0

	MOVQ         $0x2c, AX
	MOVQ         AX, X6
	VPBROADCASTB X6, Y6
	XORQ         AX, AX

loop:
	MOVQ    msg+0(FP), DI
	ADDQ    AX, DI
	VMOVDQU (DI), Y8      // load low 32-bytes
	VMOVDQU 0x20(DI), Y9  // load high 32-bytes
	ADDQ    $0x40, AX
	PUSHQ   AX

	CALL  ·__find_separator(SB)
    PUSHQ AX        // save separator mask

	MOVQ odd_ends+56(FP), DX
	MOVQ prev_iter_inside_quote+64(FP), CX
	MOVQ quote_bits+72(FP), R8
	MOVQ error_mask+80(FP), R9

	CALL ·__find_quote_mask_and_bits(SB)

	MOVQ    AX, DX  // get quotemask
    CALL ·__find_newline_delimiters(SB)

    XORQ   R10, R10
    TZCNTQ BX, CX
    JCS    skipEOL      // carry is set nothing found
    INCQ   CX
    MOVQ   $1, R10
    SHLQ   CX, R10     //
skipEOL:
    DECQ R10        // mask up and and including end-of-line marker

    POPQ  CX         // separator mask
	ORQ   BX, CX     // merge in
	ANDNQ CX, AX, AX
    ANDQ  R10, AX // clear out bits beyond end-of-line marker

	XORQ    R15, R15
	MOVQ    $1, R14
	CMPB    0x40(DI), $0x22                         // Is first char of next 64-byte a quote?
	CMOVQNE R15, R14
	MOVQ    quote_bits+72(FP), R15; MOVQ (R15), R15
	SHRQ    $1, R14, R15                            // Merge in bit-status of next 64-byte chunk

	MOVQ indexes+24(FP), DI
	MOVQ indexes_length+32(FP), SI; MOVQ (SI), BX // INDEX
	MOVQ carried+40(FP), R11; MOVQ (R11), DX      // CARRIED
	MOVQ position+48(FP), R12; MOVQ (R12), R10    // POSITION
	CALL ·__flatten_bits_incremental(SB)
	MOVQ R10, (R12)                               // POSITION
	MOVQ DX, (R11)                                // CARRIED
	MOVQ BX, (SI)                                 // INDEX

	POPQ AX
	CMPQ AX, msg_len+8(FP)
	JLT  loop

	RET
