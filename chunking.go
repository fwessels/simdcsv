package simdcsv

import (
	"encoding/csv"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"runtime"
	"sync"
)

type chunkInput struct {
	part  int
	chunk []byte
}

type chunkResult struct {
	part       int
	widowSize  uint64
	orphanSize uint64
	status     chunkStatus
}

const PREFIX_SIZE = 64 * 1024

func detectQoPattern(input []byte) bool {

	for i, q := range input[:len(input)-1] {
		if q == '"' {
			o := input[i+1]
			if o != '"' && o != ',' && o != '\n' {
				return true
			}
		}
	}
	return false
}

func detectOqPattern(input []byte) bool {

	for i, q := range input[1:] {
		if q == '"' {
			o := input[i]
			if o != '"' && o != ',' && o != '\n' {
				return true
			}
		}
	}
	return false
}

func determineAmbiguity(prefixChunk []byte) (ambiguous bool) {

	hasQo := detectQoPattern(prefixChunk)
	hasOq := detectOqPattern(prefixChunk)
	ambiguous = hasQo == false && hasOq == false

	return
}

type chunkStatus int

// Determination of start of first complete line was
// definitive or not
const (
	Unambigous chunkStatus = iota
	Ambigous
)

func (s chunkStatus) String() string {
	return [...]string{"Unambigous", "Ambigous"}[s]
}

//
// TODO: Move back to _test file (once fast parsing is in place)
//
type SingleByteReader struct {
	r io.Reader
	i int64 // current reading index
}

func (m *SingleByteReader) Read(b []byte) (n int, err error) {
	n, err = m.r.Read(b[:1])
	m.i += int64(n)
	return
}

func (m *SingleByteReader) GetIndex() int64 {
	return m.i
}

func NewSingleByteReader(r io.Reader) *SingleByteReader {

	br := bufio.NewReader(r)

	return &SingleByteReader{br, 0}
}

//
// TODO: Move back to _test file (once fast parsing is in place)
//
func augmentedFsmRune(r rune, state int32) int32 {

	// transition table
	//                    | quote comma newline other
	// -------------------|--------------------------
	// R (Record start)   |   Q     F      R      U
	// F (Field start)    |   Q     F      R      U
	// U (Unquoted field) |   X     F      R      U
	// Q (Quoted field)   |   E     Q      Q      Q
	// E (quoted End)     |   Q     F      R      X
	// X (Error)          |   X     X      X      X

	switch r {
	case '"':
		switch state {
		case 'U', 'X':
			state = 'X'
		case 'Q':
			state = 'E'
		default:
			state = 'Q'
		}

	case ',', '\n':
		switch state {
		case 'Q', 'X':
			break // unchanged
		default:
			if r == ',' {
				state = 'F'
			} else {
				state = 'R'
			}
		}

	default:
		switch state {
		case 'Q':
			break
		case 'E', 'X':
			state = 'X'
		default:
			state = 'U'
		}
	}
	return state
}

func determineQuotedOrUnquoted(prefix []byte) (quoted, ambiguous bool) {

	endStates := make(map[int32]bool)

	initialStates := []int32{'R', 'F', 'U', 'Q', 'E'}

	for _, state := range initialStates {
		for _, r := range prefix {
			state = augmentedFsmRune(rune(r), state)
		}

		if state != 'X' {
			endStates[state] = true
		}
	}

	ambiguous = len(endStates) != 1
	if !ambiguous {
		quoted = endStates['Q'] == true
	}
	return
}

func deriveChunkResult(in chunkInput) chunkResult {

	prefixSize := PREFIX_SIZE
	if len(in.chunk) < prefixSize {
		prefixSize = len(in.chunk)
	}

	chunkStatus := Unambigous
	if bytes.ContainsRune(in.chunk[:prefixSize], '"') {
		if determineAmbiguity(in.chunk[:prefixSize]) {
			chunkStatus = Ambigous
		}
	}

	widowSize := uint64(0)
	if in.part > 0 {
		i := 0

		quoted, _ := determineQuotedOrUnquoted(in.chunk[:PREFIX_SIZE])

		if quoted {
			for ; i < len(in.chunk); i++ {
				if in.chunk[i] == '"' {
					// Is there an escaped quote?
					if i+1 < len(in.chunk) && in.chunk[i+1] == '"' {
						// If so, advance one extra in order to skip
						// next character (which is a quote)
						i++
						widowSize++
					} else {
						break
					}
				}
				widowSize++
			}
		}

		// we are now at the end (closing) quote
		state := 'Q'
		for ; i < len(in.chunk); i++ {
			widowSize++
			state = augmentedFsmRune(rune(in.chunk[i]), state)
			if state != 'Q' && in.chunk[i] == '\n' {
				break
			}
		}
	}

	orphanSize := uint64(findOrphanSize(in.chunk[widowSize:]))

	return chunkResult{in.part, widowSize, orphanSize, chunkStatus}
}

// Find the orphan size by, starting from the widowSize offset,
// to iterate through the CSV content until we hit the end of the buffer
//
// TODO: Using SingleByteReader and encoding/csv for now
//       Should be replaced with accelerated/SIMD code
//
func findOrphanSize(buf []byte) int {

	sbr := NewSingleByteReader(strings.NewReader(string(buf)))
	if sbr == nil {
		return 0
	}

	r := csv.NewReader(sbr)

	addr := int64(0)

	for {
		_ /*record*/, err := r.Read()

		if err == io.EOF {
			// For a single chunk, write last line as orphan
			break
		} else if err != nil {
			break
		}

		addr = sbr.GetIndex()
	}

	return len(buf) - int(addr)
}

func chunkWorker(chunks <-chan chunkInput, results chan<- chunkResult) {

	for in := range chunks {
		results <- deriveChunkResult(in)
	}
}

func ChunkBlob(blob []byte, chunkSize uint64) {

	var wg sync.WaitGroup
	chunks := make(chan chunkInput)
	results := make(chan chunkResult)

	// Start one go routine per CPU
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			chunkWorker(chunks, results)
		}()
	}

	// Push chunks onto input channel
	go func() {
		for part, start := 0, uint64(0); ; start += chunkSize {

			end := start + chunkSize
			if end > uint64(len(blob)) {
				end = uint64(len(blob))
			}

			chunks <- chunkInput{part, blob[start:end]}

			if end >= uint64(len(blob)) {
				break
			}

			part++
		}

		// Close input channel
		close(chunks)
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results) // Close output channel
	}()

	for r := range results {
		fmt.Println(r, r.status.String())
	}
}
