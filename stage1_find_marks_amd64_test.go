package simdcsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"bytes"
	"math/bits"
	"reflect"
	"strings"
	"io/ioutil"
	"testing"
)

func TestStage1FindMarksUnaligned(t *testing.T) {
	test := strings.Repeat(
`1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,NO EVIDENCE OF REG,50,99999,99999
1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,"NO EVIDENCE OF REG",50,99999,99999
`, 100)

	record := Stage1FindMarks([]byte(test))

	want := []string{"1103341116", "2015-12-21T00:00:00", "1251", "", "", "CA", "200304", "", "HOND", "PA", "GY", "13147 WELBY WAY", "01521", "1", "4000A1", "NO EVIDENCE OF REG", "50", "99999", "99999"}

	for i := 0; i < len(record); i += len(want) {
		if !reflect.DeepEqual(record[i:i+len(want)], want) {
			t.Errorf("TestStage1FindMarksUnaligned(%d): got: %v want: %v", i, record, want)
		}
	}
}

func TestStage1LosAngelesParkingCitations(t *testing.T) {

	t.Run("fail", func(t *testing.T) {
		const test = `4277258042,2016-02-09T00:00:00.000,459,,,NJ,,,KW,CM,RD,"3772 MARTIN LUTHER KING, JR BLVD W",00500,55,80.69B,NO PARKING,73,99999,99999,,,
`
		testStage1LosAngelesParkingCitations(t, []byte(test))
	})
	t.Run("quoted-lines", func(t *testing.T) {
		const test = `4272958045,2015-12-31T00:00:00.000,847,,,CA,201503,,JAGU,PA,BL,"3749 MARTIN LUTHER KING, JR BLVD",57B,56,5204A-,DISPLAY OF TABS,25,6459025.9,1827359.3,,,
4248811976,2015-01-12T00:00:00.000,541,,,CA,201507,,CHEV,PA,WT,"107 S,ARBOLES COVRT",00503,56,22500E,BLOCKING DRIVEWAY,68,6475910.9,1729065.4,,,
4275646756,2016-01-19T00:00:00.000,1037,,,NY,,,CADI,PA,BK,"641, CALHOUN AVE",378R1,53,80.69BS,NO PARK/STREET CLEAN,73,99999,99999,,,
4276086533,2016-02-04T00:00:00.000,1121,,1013,CA,,,VOLK,PA,SL,"31,00 7TH ST W",00463,54,80.69C,PARKED OVER TIME LIM,58,99999,99999,,,
4277212796,2016-02-17T00:00:00.000,1602,,1140,GA,201610,,CHEV,PA,BK,"130, ELECTRIC AVE",908R,51,80.69C,PARKED OVER TIME LIM,58,99999,99999,,,
4277882641,2016-02-23T00:00:00.000,719,,,CA,6,,HOND,PA,GN,"18,2 MAIN ST S",00656,56,80.69AA+,NO STOP/STAND,93,99999,99999,,,
4276685420,2016-02-25T00:00:00.000,812,,,FL,,,CHRY,PA,BK,"3281, PERLITA AVE",00674,56,80.69BS,NO PARK/STREET CLEAN,73,99999,99999,,,
4277393536,2016-03-08T00:00:00.000,2247,,,CA,201603,,MITS,PA,MR,"1579 KING, JR BLVD",00500,55,22500E,BLOCKING DRIVEWAY,68,99999,99999,,,
4280358482,2016-04-07T00:00:00.000,857,,,CA,201606,,UNK,MH,WT,",1931 WEST AVENUE 30",00673,56,80.69BS,NO PARK/STREET CLEAN,73,99999,99999,,,
4281118855,2016-04-17T00:00:00.000,1544,",5",,CA,201703,,FORD,VN,GN,330 SOUTH HAMEL ROAD,00401,54,80.58L,PREFERENTIAL PARKING,68,6446091.5,1849240.9,,,
4251090233,2015-01-14T00:00:00.000,1138,,,CA,201504,,FORD,PU,BL,"772, LANKERSHIM BLVD",378R1,53,80.69BS,NO PARK/STREET CLEAN,73,99999,99999,,,
4284911094,2016-06-21T00:00:00.000,1520,,,CA,6,,KIA,PA,BK,"3171, OLYMPIC BLVD",00456,54,80.70,NO STOPPING/ANTI-GRI,163,99999,99999,,,
`
		testStage1LosAngelesParkingCitations(t, []byte(test))
	})

	t.Run("all", func(t *testing.T) {

		t.Skip()

		if testing.Short() {
			t.Skip("skipping... too long")
		}

		if buf, err := ioutil.ReadFile("testdata/parking-citations.csv"); err != nil {
			log.Fatalf("%v", err)
		} else {
			buf := bytes.ReplaceAll(buf, []byte{0x0d}, []byte{})
			lines := bytes.Split(buf, []byte("\n"))
			lines = lines[1:]
			for len(lines) > 0 {
				ln := 100
				if len(lines) < ln {
					ln = len(lines)
				}

				test := bytes.Join(lines[:ln], []byte{0x0a})
				test = append(test, []byte{0x0a}...)

				testStage1LosAngelesParkingCitations(t, test)

				lines = lines[ln:]
			}
		}
	})
}

func testStage1LosAngelesParkingCitations(t *testing.T, test []byte) {

	record := Stage1FindMarks([]byte(test))
	records := GolangCsvParser([]byte(test))

	for i := 0; i < len(record); i += 22 {
		if !reflect.DeepEqual(record[i:i+22], records[i/22]) {
			t.Errorf("TestStage1FindMarksUnaligned(%d): got: %v want: %v", i, record[i:i+22], records[i/22])
		}
	}
}

func GolangCsvParser(csvData []byte) (records [][]string) {

	records = make([][]string, 0, 1000)

	r := csv.NewReader(bytes.NewReader(csvData))

	for {
		if record, err := r.Read(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			records = append(records, record)
		}
	}
	return
}

func TestStage1FindMarksAligned128(t *testing.T) {

	vectors := []string{
		"1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,NO EVIDENCE OF REG,50,99999,99999\n   ",
		"1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,NO EVIDENCE OF REG,50,99999,99999\na,b",
		"1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,\"NO EVIDENCE OF REG\",50,99999,99999\n ",
		"1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,\"NO EVIDENCE,OF REG\",50,99999,99999\n ",
		"1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,\"NO EVIDENCE, OF REG\",50,99999,99999\n",
	}

	want := []string{"1103341116", "2015-12-21T00:00:00", "1251", "", "", "CA", "200304", "", "HOND", "PA", "GY", "13147 WELBY WAY", "01521", "1", "4000A1", "NO EVIDENCE OF REG", "50", "99999", "99999"}

	for i := 0; i <= 2; i++ {
		record := Stage1FindMarks([]byte(vectors[i]))
		if !reflect.DeepEqual(record, want) {
			t.Errorf("TestStage1FindMarks(%d): got: %v want: %v", i, record, want)
		}
	}

	// Reflect (ignored) comma in expected results
	want[15] = "NO EVIDENCE,OF REG"
	record := Stage1FindMarks([]byte(vectors[len(vectors)-2]))
	if !reflect.DeepEqual(record, want) {
		t.Errorf("TestStage1FindMarks: got: %v want: %v", record, want)
	}

	want[15] = "NO EVIDENCE, OF REG"
	record = Stage1FindMarks([]byte(vectors[len(vectors)-1]))
	if !reflect.DeepEqual(record, want) {
		t.Errorf("TestStage1FindMarks: got: %v want: %v", record, want)
	}
}

func TestStage1FindMarksMergeInNextQuoteBit(t *testing.T) {

	for repetition := 2; repetition < 64+16; repetition++ {

		start := fmt.Sprintf("%s,", strings.Repeat("A", repetition))
		rest := `"NO EVIDENCE,OF REG",50,99999,99999` + "\n"

		vector := start + rest
		vector += strings.Repeat(" ", 128-len(vector))

		record := Stage1FindMarks([]byte(vector))
		//	fmt.Println(record)

		want := []string{"", "NO EVIDENCE,OF REG", "50", "99999", "99999"}
		want[0] = strings.Repeat("A", repetition)
		if !reflect.DeepEqual(record, want) {
			t.Errorf("TestStage1FindMarksMergeInNextQuoteBit: got: %v want: %v", record, want)
			panic("exit ")
		}
	}
}

func BenchmarkStage1Golang(b *testing.B) {

	vector :=
`1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,NO EVIDENCE OF REG,50,99999,99999
1103700150,2015-12-21T00:00:00,1435,,,CA,201512,,GMC,VN,WH,525 S MAIN ST,1C51,1,4000A1,NO EVIDENCE OF REG,50,99999,99999
1104803000,2015-12-21T00:00:00,2055,,,CA,201503,,NISS,PA,BK,200 WORLD WAY,2R2,2,8939,WHITE CURB,58,6439997.9,1802686.4
1104820732,2015-12-26T00:00:00,1515,,,CA,,,ACUR,PA,WH,100 WORLD WAY,2F11,2,000,17104h,,6440041.1,1802686.2
1105461453,2015-09-15T00:00:00,115,,,CA,200316,,CHEV,PA,BK,GEORGIA ST/OLYMPIC,1FB70,1,8069A,NO STOPPING/STANDING,93,99999,99999
1106226590,2015-09-15T00:00:00,19,,,CA,201507,,CHEV,VN,GY,SAN PEDRO S/O BOYD,1A35W,1,4000A1,NO EVIDENCE OF REG,50,99999,99999
1106500452,2015-12-17T00:00:00,1710,,,CA,201605,,MAZD,PA,BL,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999
1106500463,2015-12-17T00:00:00,1710,,,CA,201602,,TOYO,PA,BK,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999
1106506402,2015-12-22T00:00:00,945,,,CA,201605,,CHEV,PA,BR,721 S WESTLAKE,2A75,1,8069AA,NO STOP/STAND AM,93,99999,99999
1106506413,2015-12-22T00:00:00,1100,,,CA,201701,,NISS,PA,SI,1159 HUNTLEY DR,2A75,1,8069AA,NO STOP/STAND AM,93,99999,99999
`

	b.SetBytes(int64(len(vector)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		r := csv.NewReader(strings.NewReader(vector))

		for {
			_, err := r.Read()
			if err == io.EOF {
				break
			}
		}
	}
}

