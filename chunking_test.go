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

	f, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer f.Close()

	sbr := NewSingleByteReader(f)
	if sbr == nil {
		log.Fatalf("Failed to create SingleByteReader")
	}

	r := csv.NewReader(sbr)

	addr, prev_addr, lines := int64(0), int64(0), 0
	assumeHasWidow := false

	for {
		_ /*record*/, err := r.Read()

		if err == io.EOF {
			// For a single chunk, write last line as orphan
			if len(chunks) == 1 {
				chunks[0].orphanSize = uint64(sbr.GetIndex() - prev_addr)
			}
			break
		} else if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(record)
		prev_addr, addr = addr, sbr.GetIndex()

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

			// if dump {
			// 	start := ((chunkBase - addrBase) & ^0xf) - (((int64(prevOrphanSize) >> 4) + 1) << 4)
			// 	end := ((chunkBase - addrBase) & ^0xf) + (((int64(widowSize) >> 4) + 1) << 4)
			//
			// 	fmt.Println("part:", chunks[len(chunks)-1].part)
			// 	dumpWithAddr(buf[start:end], chunkBase-(((int64(prevOrphanSize)>>4)+1)<<4))
			// 	fmt.Println()
			// }
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
		// mismatch: got {1121 3549 5325 0}, want {1121 11583 1320 0}
		"mrisdal/fake-news/fake.csv",
		//
		// parse error on line 4, column 40: bare " in non-quoted-field
		//                                         v        v
		// 2016-03-14 12:52:21,Jeep_Grand_Cherokee_"Overland",privat,Angebot,9800,test,suv,2004,automatik,163,grand,125000,8,diesel,jeep,,2016-03-14 00:00:00,0,90480,2016-04-05 12:47:46
		"orgesleka/used-cars-database/autos.csv",
		//
		// record on line 878: wrong number of fields
		"NUFORC/ufo-sightings/complete.csv",
		//
		"shivamb/real-or-fake-fake-jobposting-prediction/fake_job_postings.csv",
		//
		// mismatch: got {0 0 4 0}, want {0 0 2401 0},
		"kaggle/hillary-clinton-emails/Emails.csv",
		//
		"udacity/armenian-online-job-postings/online-job-postings.csv",
		//
		// parse error on line 1, column 1: bare " in non-quoted-field
		"theworldbank/health-nutrition-and-population-statistics/data.csv",
		//
		// mismatch: got {3 88 346 0}, want {3 259 346 0}
		"airbnb/boston/listings.csv",
		"airbnb/boston/reviews.csv",
		//
		// mismatch: got {5588 2 29 0}, want {5588 2834 1541 0}
		"adhab/jobposts/data job posts.csv",
		//
		// mismatch: got {28 112 122 0}, want {28 112 198 0}
		"codename007/funding-successful-projects/test.csv",
		"codename007/funding-successful-projects/train.csv",
		//
		// mismatch: got {1116 12 252 0}, want {1116 152 252 0}
		"harriken/emoji-sentiment/__readervswriter.csv",
		"harriken/emoji-sentiment/comments2emoji_frequency_matrix_cleaned.csv",
		"harriken/emoji-sentiment/ijstable.csv",
		//
		"ryanxjhan/cbc-news-coronavirus-articles-march-26/news.csv",
		//
		// mismatch: got {4167 2139 237 0}, want {4167 1994 0 0}
		"PromptCloudHQ/us-jobs-on-monstercom/monster_com-job_sample.csv",
		"PromptCloudHQ/innerwear-data-from-victorias-secret-and-others/us_topshop_com.csv",
		//
		// record on line 7: wrong number of fields
		"bls/consumer-price-index/cu.area.csv",
		// record on line 8: wrong number of fields
		"bls/consumer-price-index/cu.item.csv",
		// record on line 2: wrong number of fields
		"bls/consumer-price-index/cu.series.csv",
		//
		// parse error on line 3677, column 18: extraneous or missing " in quoted-field
		"orgesleka/android-apps/apps.csv",
		//
		// mismatch: got {2840 1415 2751 0}, want {2840 1415 19135 0}
		"PromptCloudHQ/jobs-on-naukricom/naukri_com-job_sample.csv",
		//
		// mismatch: got {3452 124 4189 0}, want {3452 2015 258 0}
		"PromptCloudHQ/jobs-on-naukricom/dice_com-job_us_sample.csv",
		"PromptCloudHQ/us-technology-jobs-on-dicecom/dice_com-job_us_sample.csv",
	}

	for _, l := range list {
		if filename == l {
			return true
		}
	}

	return false
}
