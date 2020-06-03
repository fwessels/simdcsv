package simdcsv

import (
	"testing"
)

func TestAmbiguityWithFSM(t *testing.T) {

	// transition table
	//                    | quote comma newline other
	// -------------------|--------------------------
	// R (Record start)   |   Q     F      R      U
	// F (Field start)    |   Q     F      R      U
	// U (Unquoted field) |   -     F      R      U
	// Q (Quoted field)   |   E     Q      Q      Q
	// E (quoted End)     |   Q     F      R      -

	const ambigious = `
       l  i  c  e  ,  "  ,  "  ,  1  6 \n  B  o  b  ,  "  ,  "  ,  1  7
	R  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
	F  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
	U  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
	Q  Q  Q  Q  Q  Q  E  F  Q  Q  Q  Q  Q  Q  Q  Q  Q  E  F  Q  Q  Q  Q
	E  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -`

	// Except for E, all other starting states successfully pass through the chunk.
	// Since the remaining starting states R, F, U, and Q fall into two categories, the
	// chunk is ambiguous.

	const unambiguous = `
       l  i  c  e  ,  " \n  "  ,  1  6 \n  B  o  b  ,  "  M  "  ,  1  7
	R  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
	F  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
	U  U  U  U  U  F  Q  Q  E  F  U  U  R  U  U  U  F  Q  Q  E  F  U  U
	Q  Q  Q  Q  Q  Q  E  R  Q  Q  Q  Q  Q  Q  Q  Q  Q  E  -  -  -  -  -
    E  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -  -`

	// The chunk has an invalid state Q, because other are not
	// allowed after the state E, which is transited from the starting state Q
	// after reading the string `lice,"\n",16\nBob,"`. Thus, all valid starting
	// states are unquoted, and the example chunk is therefore unambiguous.

}