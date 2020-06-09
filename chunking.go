package simdcsv

import (
	"sync"
	"runtime"
	"fmt"
)

type chunkInput struct {
	start uint64
	end   uint64
}

type chunkResult struct {
	widowSize  uint64
	orphanSize uint64
	ambiguous  bool
	quoted     bool
}

func chunkWorker(chunks <-chan chunkInput, results chan<- chunkResult) {

	for _ = range chunks {

		results <- chunkResult{0, 0, false, false}
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
		for start := uint64(0); ; start += chunkSize {

			end := start+chunkSize
			if end > uint64(len(blob)) {
				end = uint64(len(blob))
			}

			chunks <- chunkInput{start, end}

			if end >= uint64(len(blob)) {
				break
			}
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
		fmt.Println(r)
	}
}
