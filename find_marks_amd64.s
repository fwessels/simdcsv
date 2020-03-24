//+build !noasm !appengine

// find_marks_in_slice(msg []byte, indexes *[INDEXES_SIZE]uint32, indexes_length *uint64) (out uint64)
TEXT ·find_marks_in_slice(SB), 7, $0

    MOVQ         msg+0(FP), DI
	MOVQ         $0x2c, AX
	MOVQ         AX, X6
	VPBROADCASTB X6, Y6

    VMOVDQU    (DI), Y8          // load low 32-bytes
    VMOVDQU    0x20(DI), Y9      // load high 32-bytes

    CALL ·__find_separator(SB)
    PUSHQ       AX

    MOVQ odd_ends+40(FP), DX
    MOVQ prev_iter_inside_quote+48(FP), CX
    MOVQ quote_bits+56(FP), R8
    MOVQ error_mask+64(FP), R9

    CALL ·__find_quote_mask_and_bits(SB)
    PUSHQ       AX // quote_mask

    MOVQ        indexes_length+32(FP), SI
    // MOVQ     mask+16(FP), MASK
    // MOVQ     carried+24(FP), R11
    // MOVQ     position+32(FP), R12
    MOVQ        (SI), BX // , INDEX
    MOVQ        $0, DX   // , CARRIED
    MOVQ        $0, R10  // , POSITION

    POPQ        CX // quote_mask
    POPQ        AX // separator mask
    ANDNQ       AX, CX, AX

    MOVQ        quote_bits+56(FP), R15
    MOVQ        (R15), R15
    MOVQ        $0, R14
    CMPB        0x40(DI), $0x22  // Is first char of next 64-byte a quote?
    JNZ         skip
    MOVQ        $1,  R14
skip:
    SHRQ        $1, R14, R15

    MOVQ        indexes+24(FP), DI
    CALL ·__flatten_bits_incremental(SB)
    // MOVQ     POSITION, (R12)
    // MOVQ     CARRIED, (R11)
    MOVQ        BX /*INDEX*/, (SI)
    RET
