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

	MOVQ        AX, out+40(FP)

    MOVQ        indexes+24(FP), DI
    MOVQ        indexes_length+32(FP), SI
    // MOVQ     mask+16(FP), MASK
    // MOVQ     carried+24(FP), R11
    // MOVQ     position+32(FP), R12
    MOVQ        (SI), BX // , INDEX
    MOVQ        $0, DX   // , CARRIED
    MOVQ        $0, R10  // , POSITION
    CALL ·__flatten_bits_incremental(SB)
    // MOVQ     POSITION, (R12)
    // MOVQ     CARRIED, (R11)
    MOVQ        BX /*INDEX*/, (SI)
    RET
