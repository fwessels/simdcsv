//+build !noasm !appengine

// find_double_quotes(mask uint64)
TEXT ·find_double_quotes(SB), 7, $0
    MOVQ mask+0(FP), AX
    MOVQ indices+8(FP), R8
    XORQ  R9, R9
    XORQ  DX, DX
loop:
    TZCNTQ AX, CX
    JCS    done     // carry is set if CX == 64
    ADDQ   CX, DX
    SHRQ   CX, AX
    MOVQ   AX, BX
    ANDQ   $3, BX
    CMPQ   BX, $3   // Two consequetive bits set?
    JNZ    skip
    MOVD   DX, (R8)(R9*4)
    INCQ   R9
skip:
    SHRQ   $2, AX
    ADDQ   $2, DX
    JMP    loop

done:
    MOVQ R9, entries+32(FP)
    RET
