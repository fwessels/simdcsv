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

	// A chunk is AMBIGUOUS if and only if the remaining valid starting
	// states are all either unquoted states or quoted state

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

func TestAmbiquityWithPatterns(t *testing.T) {

	// q-o pattern
	//                    | quote  other
	// -------------------|-------------
	// R (Record start)   |   Q      Q
	// F (Field start)    |   Q      Q
	// U (Unquoted field) |   -      -
	// Q (Quoted field)   |   E      -
	// E (quoted End)     |   Q      Q

	// o-q pattern
	//                    | other  quote
	// -------------------|-------------
	// R (Record start)   |   U      -
	// F (Field start)    |   U      -
	// U (Unquoted field) |   U      -
	// Q (Quoted field)   |   Q      E
	// E (quoted End)     |   -      -

	// Both q-o and o-q patterns have a crucial property: for all
	// possible input states, the FSM transits into the same output state,
	// after reading an input string following the pattern

	// The chunk is ambiguous if and only if it contains neither
	// q-o pattern strings nor o-q pattern strings

	const ambigious = `
       l  i  c  e  ,  "  ,  "  ,  1  6 \n  B  o  b  ,  "  ,  "  ,  1  7
   q-o .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .
   o-q .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .`

	const unambiguous = `
       l  i  c  e  ,  " \n  "  ,  1  6 \n  B  o  b  ,  "  M  "  ,  1  7
   q-o .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  X  X  .  .  .
   o-q .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .  .`
}

func TestSyntaxErrors(t *testing.T) {

	// transition table
	//                    | quote comma newline other
	// -------------------|--------------------------
	// R (Record start)   |   Q     F      R      U
	// F (Field start)    |   Q     F      R      U
	// U (Unquoted field) |   X     F      R      U
	// Q (Quoted field)   |   E     Q      Q      Q
	// E (quoted End)     |   Q     F      R      X
	// X (Error)          |   X     X      X      X


	const chunk1 = `
	A  l  i  c  e  ,  "  F  "  ,  "  H  i \n  " \n  B  o  b  ,  "  M  "  ,  "  H
	U  U  U  U  U  F  Q  Q  E  F  Q  Q  Q  Q  E  R  U  U  U  F  Q  Q  E  F  Q  Q`


	const chunk2 = `
	e  l  l  o \n  " \n  C  h  r  i  s  ,  M  "  ,  "  b  y  e  " \n  D  a  v  e
	Q  Q  Q  Q  Q  E  R  U  U  U  U  U  F  U  X  X  X  X  X  X  X  X  X  X  X  X`

	const chunk3 = `
	,  "  M  "  ,  "  M  o  r  n  i  n  g  ! \n  " \n
	X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X  X`
}
