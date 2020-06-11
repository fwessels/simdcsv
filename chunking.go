package simdcsv

import (
	"bytes"
	"fmt"
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
	quoted     bool
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

const (
	HasNoQuotes chunkStatus = iota
	Unambigous
	Ambigous
)

func (s chunkStatus) String() string {
	return [...]string{"HasNoQuotes", "Unambigous", "Ambigous"}[s]
}

func chunkWorker(chunks <-chan chunkInput, results chan<- chunkResult) {

	for in := range chunks {

		prefixSize := PREFIX_SIZE
		if len(in.chunk) < prefixSize {
			prefixSize = len(in.chunk)
		}

		// has no quotes    | unquoted
		// unambiguous      | unquoted
		// unambiguous      | quoted
		// ambiguous        | unquoted
		// ambiguous        | quoted

		chunkStatus, quoted := HasNoQuotes, false
		if bytes.ContainsRune(in.chunk[:prefixSize], '"') {
			if determineAmbiguity(in.chunk[:prefixSize]) {
				chunkStatus = Ambigous
			} else {
				chunkStatus = Unambigous
			}
		}

		results <- chunkResult{in.part, 0, 0, chunkStatus, quoted}
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
