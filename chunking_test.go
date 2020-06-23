package simdcsv

import (
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/fwessels/kaggle-go"
	"golang.org/x/exp/mmap"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"text/tabwriter"
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

func memoryTrackingCsvParser(filename string, splitSize int64, dump bool) (chunks []chunkResult) {

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
	buf := make([]byte, mapWindow)
	addrBase := int64(-mapWindow / 2) // make sure we trigger a memory map
	endOfMem := 0

	addr, prev_addr, lines := int64(0), int64(0), 0
	assumeHasWidow := false

	for {
		if addr-addrBase >= mapWindow/2 {
			addrBase += mapWindow / 2
			// fmt.Printf("Remapping at %08x\n", addrBase)
			n, err := memmap.ReadAt(buf, addrBase)
			if err != nil && !errors.Is(err, io.EOF) {
				log.Fatalf("memmap failed: %v", err)
			}
			endOfMem = int(addrBase) + n
			if n < len(buf) {
				buf = buf[:n]
			}
		}
		record, err := r.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(record)
		length := int64(len(record) - 1) // nr of commas minus 1
		for _, f := range record {
			length += int64(len(f))
		}
		prev_addr, addr = addr, addr+length

		for fudge := int64(1); ; {
			if buf[addr-addrBase-1] == 0x0a {
				break
			}
			if buf[addr-addrBase-1+fudge] == 0x0a {
				addr += fudge
				break
			}
			fudge += 1
			if addr+fudge >= int64(endOfMem) {
				addr += fudge
				break
			}

			if fudge >= length/2 {
				fmt.Println(record)

				chunkBase := addr & ^(splitSize - 1)
				start := ((chunkBase - addrBase) & ^0xf) - 0x100
				end := ((chunkBase - addrBase) & ^0xf) + 0x100

				dumpWithAddr(buf[start:end], chunkBase-0x100)
				fmt.Println()

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

	return
}

func testVerifyChunking(t *testing.T, filename string) {

	fi, err := os.Stat(filename)
	if err != nil {
		log.Fatalf("%v", err)
	}
	filesize := fi.Size()

	for shift := 14; shift < 20; shift++ {

		splitSize := int64(1 << shift)
		fmt.Println("Testing for splitSize", splitSize)
		sourceOfTruth := memoryTrackingCsvParser(filename, splitSize, false)

		memmap, err := mmap.Open(filename)
		defer memmap.Close()
		if err != nil {
			log.Fatalf("%v", err)
		}

		buf := make([]byte, splitSize)
		for i := 0; i < len(sourceOfTruth); i++ {
			n, err := memmap.ReadAt(buf, int64(i)*splitSize)
			if err != nil && !errors.Is(err, io.EOF) {
				log.Fatalf("memmap failed: %v", err)
			}

			result := deriveChunkResult(chunkInput{i, buf[:n]})
			if !reflect.DeepEqual(result, sourceOfTruth[i]) {
				r := result
				r.status = sourceOfTruth[i].status
				if !reflect.DeepEqual(r, sourceOfTruth[i]) {
					t.Errorf("TestVerifyChunking mismatch: got %v, want %v", result, sourceOfTruth[i])
				}
			}
		}

		if filesize < splitSize {
			break // no point in continuing testing
		}
	}
}

func TestVerifyChunking(t *testing.T) {

	testVerifyChunking(t, "./test-data/country_wise_latest.csv")
	testVerifyChunking(t, "./test-data/covid_19_clean_complete.csv")
	testVerifyChunking(t, "./test-data/day_wise.csv")
	testVerifyChunking(t, "./test-data/full_grouped.csv")
	testVerifyChunking(t, "./test-data/usa_county_wise.csv")
	testVerifyChunking(t, "./test-data/worldometer_data.csv")
}

func TestVerifyThruKaggle(t *testing.T) {

	entries := kaggle.ListByVotesPopularity(10*1024*1024, 20*1024*1024, 100)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	for _, r := range entries {
		for _, f := range r {
			fmt.Fprint(w, f+"\t")
		}
		fmt.Fprintln(w)
	}
	w.Flush()

	const downloadPath = "./test-data/"
	for index, r := range entries {

		if index < 6 {
			continue
		}

		if index >= 7 {
			break
		}

		if err := os.RemoveAll(downloadPath); err != nil {
			t.Errorf("Error while removing directory: %v", nil)
		}

		dataset := r[0]
		fmt.Println("Downloading", dataset)
		kaggle.Download(r[0], downloadPath)

		files, err := ioutil.ReadDir(downloadPath)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			extension := filepath.Ext(file.Name())
			if extension == ".csv" {

				if blackListed(dataset + "/" + file.Name()) {
					continue
				}

				fmt.Println(dataset + "/" + file.Name())
				testVerifyChunking(t, downloadPath+file.Name())
			}
		}

	}
}

func blackListed(filename string) bool {

	list := []string{
		// uses 0x0d as delimiter instead of 0x0a
		"AnalyzeBoston/crimes-in-boston/offense_codes.csv",
		//
		// parse error on line 4, column 40: bare " in non-quoted-field
		//                                         v        v
		// 2016-03-14 12:52:21,Jeep_Grand_Cherokee_"Overland",privat,Angebot,9800,test,suv,2004,automatik,163,grand,125000,8,diesel,jeep,,2016-03-14 00:00:00,0,90480,2016-04-05 12:47:46
		"orgesleka/used-cars-database/autos.csv",
		//
		// record on line 878: wrong number of fields
		"NUFORC/ufo-sightings/complete.csv",
	}

	for _, l := range list {
		if filename == l {
			return true
		}
	}

	return false
}
