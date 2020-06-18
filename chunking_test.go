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
	"strings"
	"testing"
)

func TestChunkWorker(t *testing.T) {
	blob, _ := ioutil.ReadFile("parking-citations-2M.csv")

	for s := 14; s < 16; s++ {
		ChunkBlob(blob, 1<<s)
	}
}

func parseCsv(filename string) {

	memmap, err := mmap.Open(filename)
	if err != nil {
		log.Fatalf("%v", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("%v", err)
	}

	r := csv.NewReader(file)

	const mapWindow = 1024 * 1024
	buf, addrBase := make([]byte, mapWindow), int64(0x0)
	memmap.ReadAt(buf, addrBase)

	addr, lines := 0, 0
	for {
		if addr >= len(buf)/2 {
			addrBase += mapWindow / 2
			fmt.Printf("Remapping at %08x\n", addrBase)
			memmap.ReadAt(buf, addrBase)
			addr -= mapWindow / 2
		}
		record, err := r.Read()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(record)
		length := 0
		for _, f := range record {
			length += len(f) + 1
		}
		addr += length

		for fudge := 1; ; fudge += 1 {
			if buf[addr-1] == 0x0a {
				break
			}
			if buf[addr-1-fudge] == 0x0a {
				addr -= fudge
				break
			}
			if buf[addr-1+fudge] == 0x0a {
				addr += fudge
				break
			}

			if fudge >= length/2 {
				log.Fatalf("Unable to find newline: %d", fudge)
			}
		}

		lines += 1

		// fmt.Print(hex.Dump(buf[addr-8 : addr+8]))

		// if lines > 100000 {
		// 	break
		// }
	}

	fmt.Println(lines)
}

func TestVerifyChunking(t *testing.T) {

	parseCsv("Parking_Violations_Issued_-_Fiscal_Year_2017.csv")
}
