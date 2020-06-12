package simdcsv

import (
	"io/ioutil"
	"testing"
)

func TestChunkWorker(t *testing.T) {
	blob, _ := ioutil.ReadFile("parking-citations-2M.csv")

	for s := 14; s < 16; s++ {
		ChunkBlob(blob, 1 << s)
	}
}
