package simdcsv

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"golang.org/x/exp/mmap"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestChunkWorker(t *testing.T) {
	blob, _ := ioutil.ReadFile("parking-citations-2M.csv")

	for s := 14; s < 16; s++ {
		ChunkBlob(blob, 1<<s)
	}
}

func dumpWithAddr(buf []byte, addr int64) {
	d := hex.Dump(buf)
	lines := strings.Split(d, "\n")

	for i, l := range lines {
		l = strings.ReplaceAll(l, fmt.Sprintf("%08x", i*16), fmt.Sprintf("%08x", int(addr)+i*16))
		fmt.Println(strings.ReplaceAll(l, " 0a ", "<0a>"))
	}
}

const splitSize = 1 << 16

func parseCsv(filename string, dump bool) (chunks []chunkResult) {

	chunks = make([]chunkResult, 0)
	chunks = append(chunks, chunkResult{widowSize: 0})

	memmap, err := mmap.Open(filename)
	defer memmap.Close()
	if err != nil {
		log.Fatalf("%v", err)
	}

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Fatalf("%v", err)
	}

	r := csv.NewReader(file)

	const mapWindow = 1024 * 1024
	buf, addrBase := make([]byte, mapWindow), int64(0x0)
	memmap.ReadAt(buf, addrBase)

	addr, prev_addr, lines := int64(0), int64(0), 0
	assumeHasWidow := false

	for {
		if addr-addrBase >= int64(len(buf)/2) {
			addrBase += mapWindow / 2
			// fmt.Printf("Remapping at %08x\n", addrBase)
			memmap.ReadAt(buf, addrBase)
		}
		record, err := r.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(record)
		length := int64(0)
		for _, f := range record {
			length += int64(len(f)) + 1
		}
		prev_addr, addr = addr, addr+length

		for fudge := int64(1); ; fudge += 1 {
			if buf[addr-addrBase-1] == 0x0a {
				break
			}
			if buf[addr-addrBase-1-fudge] == 0x0a {
				addr -= fudge
				break
			}
			if buf[addr-addrBase-1+fudge] == 0x0a {
				addr += fudge
				break
			}

			if fudge >= length/2 {
				log.Fatalf("Unable to find newline: %d", fudge)
			}
		}

		if (addr-1)&(splitSize-1) == splitSize-1 {
			//
			// Delimiter is last character of the chunk, next chunk has no way of
			// knowing that it exactly start with a new line, so next chunk has to
			// assume that it starts with a widow regardless.
			//
			// Likewise, if this is true, then we have no orphan (since we land
			// precisely on a delimiter at the end of the chunk.)
			//
			assumeHasWidow = true
		} else if assumeHasWidow || (prev_addr & ^(splitSize-1)) < ((addr-1) & ^(splitSize-1)) {
			chunkBase := addr & ^(splitSize - 1)

			prevOrphanSize := uint64(0)
			if !assumeHasWidow { // orphan size of previous block is 0 is we assume we start with a widow
				prevOrphanSize = uint64(chunkBase - prev_addr)
			}
			if len(chunks) > 0 {
				chunks[len(chunks)-1].orphanSize = prevOrphanSize
			}

			widowSize := uint64(addr - chunkBase)
			chunks = append(chunks, chunkResult{part: len(chunks), widowSize: widowSize})

			if dump {
				start := ((chunkBase - addrBase) & ^0xf) - (((int64(prevOrphanSize) >> 4) + 1) << 4)
				end := ((chunkBase - addrBase) & ^0xf) + (((int64(widowSize) >> 4) + 1) << 4)

				fmt.Println("part:", chunks[len(chunks)-1].part)
				dumpWithAddr(buf[start:end], chunkBase-(((int64(prevOrphanSize)>>4)+1)<<4))
				fmt.Println()
			}
			assumeHasWidow = false
		}
		lines += 1
	}

	// Write last orphan_size
	if len(chunks) > 0 {
		chunks[len(chunks)-1].orphanSize = uint64(addr - prev_addr)
	}

	fmt.Println(lines)
	fmt.Println(len(chunks))

	return
}

func TestVerifyChunking(t *testing.T) {

	const filename = "Parking_Violations_Issued_-_Fiscal_Year_2017.csv"
	sourceOfTruth := parseCsv(filename, false)

	memmap, err := mmap.Open(filename)
	defer memmap.Close()
	if err != nil {
		log.Fatalf("%v", err)
	}

	buf := make([]byte, splitSize)
	for i := 0; i < len(sourceOfTruth); i++ {
		memmap.ReadAt(buf, int64(i*splitSize))

		result := deriveChunkResult(chunkInput{i, buf})
		if !reflect.DeepEqual(result, sourceOfTruth[i]) {
			r := result
			r.status = sourceOfTruth[i].status
			if !reflect.DeepEqual(r, sourceOfTruth[i]) {
				fmt.Println(" truth:", sourceOfTruth[i])
				fmt.Println("result:", result)
			}
		}
	}
}
