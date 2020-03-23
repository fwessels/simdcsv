//+build !noasm !appengine

#define MASK    AX
#define INDEX   BX
#define ZEROS   CX
#define CARRIED DX
#define SHIFTS  R8
#define POSITION R10
#define LENGTH     R14 // Lengths
#define QUOTE_MASK R15

TEXT ·_flatten_bits_incremental(SB), $0-48

    MOVQ base_ptr+0(FP), DI
    MOVQ pbase+8(FP), SI
    MOVQ mask+16(FP), MASK
    MOVQ quote_bits+24(FP), R15
    MOVQ carried+32(FP), R11
    MOVQ position+40(FP), R12
    MOVQ (SI), INDEX
    MOVQ (R11), CARRIED
    MOVQ (R12), POSITION
    CALL ·__flatten_bits_incremental(SB)
    MOVQ   POSITION, (R12)
    MOVQ   CARRIED, (R11)
    MOVQ   INDEX, (SI)
    RET


TEXT ·__flatten_bits_incremental(SB), $0
    XORQ SHIFTS, SHIFTS

    // First iteration takes CARRIED into account
    TZCNTQ MASK, ZEROS
    JCS    done        // carry is set if ZEROS == 64

    // Two shifts required because maximum combined shift (63+1) exceeds 6-bits
    SHRQ   $1, MASK
    SHRQ   ZEROS, MASK
    SHRQ   ZEROS, QUOTE_MASK
    INCQ   ZEROS
    ADDQ   ZEROS, SHIFTS
    ADDQ   CARRIED, ZEROS
    ADDQ   $2, INDEX
    MOVQ   ZEROS, LENGTH
    DECQ   LENGTH
    ADDL   POSITION, -8(DI)(INDEX*4)
    ADDL   LENGTH, -4(DI)(INDEX*4)
    ADDQ   ZEROS, POSITION
    XORQ   CARRIED, CARRIED // Reset CARRIED to 0 (since it has been used)
    TESTQ  $1, QUOTE_MASK           // Is there an opening quote?
    JZ     noquote_prologue
    ADDL   $1, (DI)(INDEX*4)        // Adjust next position ...
    SUBL   $2, 4(DI)(INDEX*4)       // ... and next length
noquote_prologue:
    SHRQ   $1, QUOTE_MASK

loop:
    TZCNTQ MASK, ZEROS
    JCS    done        // carry is set if ZEROS == 64

    SHRQ   ZEROS, QUOTE_MASK
    INCQ   ZEROS
    SHRQ   ZEROS, MASK
    ADDQ   ZEROS, SHIFTS
    ADDQ   $2, INDEX
    MOVQ   ZEROS, LENGTH
    DECQ   LENGTH
    ADDL   POSITION, -8(DI)(INDEX*4)
    ADDL   LENGTH, -4(DI)(INDEX*4)
    TESTQ  $1, QUOTE_MASK           // Is there an opening quote?
    JZ     noquote
    ADDL   $1, (DI)(INDEX*4)        // Adjust next position ...
    SUBL   $2, 4(DI)(INDEX*4)       // ... and next length
noquote:
    SHRQ   $1, QUOTE_MASK
    ADDQ   ZEROS, POSITION
    JMP    loop

done:
    MOVQ   $64, R9
    SUBQ   SHIFTS, R9
    ADDQ   R9, CARRIED    // CARRIED += 64 - shifts (remaining empty bits to carry over to next call)
    RET
