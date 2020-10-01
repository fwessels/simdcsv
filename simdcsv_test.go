package simdcsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"runtime"
)

func TestLosAngelesParkingCitations(t *testing.T) {

	t.Run("single-line", func(t *testing.T) {
		const test = `4277258042,2016-02-09T00:00:00.000,459,,,NJ,,,KW,CM,RD,"3772 MARTIN LUTHER KING, JR BLVD W",00500,55,80.69B,NO PARKING,73,99999,99999,,,
`
		compareAgainstEncodingCsv(t, []byte(test))
	})

	t.Run("multiple-lines", func(t *testing.T) {
		const test = `1103341116,2015-12-21T00:00:00.000,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,NO EVIDENCE OF REG,50,99999,99999,,,
1103700150,2015-12-21T00:00:00.000,1435,,,CA,201512,,GMC,VN,WH,525 S MAIN ST,1C51,1,4000A1,NO EVIDENCE OF REG,50,99999,99999,,,
1104803000,2015-12-21T00:00:00.000,2055,,,CA,201503,,NISS,PA,BK,200 WORLD WAY,2R2,2,8939,WHITE CURB,58,6439997.9,1802686.4,,,
1104820732,2015-12-26T00:00:00.000,1515,,,CA,,,ACUR,PA,WH,100 WORLD WAY,2F11,2,000,17104h,,6440041.1,1802686.2,,,
1105461453,2015-09-15T00:00:00.000,115,,,CA,200316,,CHEV,PA,BK,GEORGIA ST/OLYMPIC,1FB70,1,8069A,NO STOPPING/STANDING,93,99999,99999,,,
1106226590,2015-09-15T00:00:00.000,19,,,CA,201507,,CHEV,VN,GY,SAN PEDRO S/O BOYD,1A35W,1,4000A1,NO EVIDENCE OF REG,50,99999,99999,,,
1106500452,2015-12-17T00:00:00.000,1710,,,CA,201605,,MAZD,PA,BL,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999,,,
1106500463,2015-12-17T00:00:00.000,1710,,,CA,201602,,TOYO,PA,BK,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999,,,
1106506402,2015-12-22T00:00:00.000,945,,,CA,201605,,CHEV,PA,BR,721 S WESTLAKE,2A75,1,8069AA,NO STOP/STAND AM,93,99999,99999,,,
`
		compareAgainstEncodingCsv(t, []byte(test))
	})

	t.Run("multiple-lines-with-header", func(t *testing.T) {
		const test = `Ticket number,Issue Date,Issue time,Meter Id,Marked Time,RP State Plate,Plate Expiry Date,VIN,Make,Body Style,Color,Location,Route,Agency,Violation code,Violation Description,Fine amount,Latitude,Longitude,Agency Description,Color Description,Body Style Description
1103341116,2015-12-21T00:00:00.000,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,NO EVIDENCE OF REG,50,99999,99999,,,
1103700150,2015-12-21T00:00:00.000,1435,,,CA,201512,,GMC,VN,WH,525 S MAIN ST,1C51,1,4000A1,NO EVIDENCE OF REG,50,99999,99999,,,
1104803000,2015-12-21T00:00:00.000,2055,,,CA,201503,,NISS,PA,BK,200 WORLD WAY,2R2,2,8939,WHITE CURB,58,6439997.9,1802686.4,,,
1104820732,2015-12-26T00:00:00.000,1515,,,CA,,,ACUR,PA,WH,100 WORLD WAY,2F11,2,000,17104h,,6440041.1,1802686.2,,,
1105461453,2015-09-15T00:00:00.000,115,,,CA,200316,,CHEV,PA,BK,GEORGIA ST/OLYMPIC,1FB70,1,8069A,NO STOPPING/STANDING,93,99999,99999,,,
1106226590,2015-09-15T00:00:00.000,19,,,CA,201507,,CHEV,VN,GY,SAN PEDRO S/O BOYD,1A35W,1,4000A1,NO EVIDENCE OF REG,50,99999,99999,,,
1106500452,2015-12-17T00:00:00.000,1710,,,CA,201605,,MAZD,PA,BL,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999,,,
1106500463,2015-12-17T00:00:00.000,1710,,,CA,201602,,TOYO,PA,BK,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999,,,
1106506402,2015-12-22T00:00:00.000,945,,,CA,201605,,CHEV,PA,BR,721 S WESTLAKE,2A75,1,8069AA,NO STOP/STAND AM,93,99999,99999,,,
`
		compareAgainstEncodingCsv(t, []byte(test))
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
4277258042,2016-02-09T00:00:00.000,459,,,NJ,,,KW,CM,RD,"3772 MARTIN LUTHER KING, JR BLVD W",00500,55,80.69B,NO PARKING,73,99999,99999,,,
`
		compareAgainstEncodingCsv(t, []byte(test))
	})

	t.Run("all", func(t *testing.T) {

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
				ln := 1000
				if len(lines) < ln {
					ln = len(lines)
				}

				test := bytes.Join(lines[:ln], []byte{0x0a})

				compareAgainstEncodingCsv(t, test)

				lines = lines[ln:]

				runtime.GC()
			}
		}
	})
}

func TestSimdCsv(t *testing.T) {
	t.Run("parking-citations-100K", func(t *testing.T) {
		testSimdCsv(t, "parking-citations-100K.csv")
	})
	t.Run("worldcitiespop", func(t *testing.T) {
		testSimdCsv(t, "worldcitiespop.csv")
	})
}

func testSimdCsv(t *testing.T, filename string) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("%v", err)
	}

	compareAgainstEncodingCsv(t, buf)
}

func TestFilterEmptyLines(t *testing.T) {
	compareAgainstEncodingCsv(t, []byte("a,b,c\n\nd,e,f\n\n"))
	compareAgainstEncodingCsv(t, []byte("a,b,c\n\n\nd,e,f\n\n"))
	compareAgainstEncodingCsv(t, []byte("a,b,c\n\n\n\nd,e,f\n\n"))
	compareAgainstEncodingCsv(t, []byte("a,b,c\n\n,,\n\n\nd,e,f\n\n"))
}

func compareAgainstEncodingCsv(t *testing.T, test []byte) {

	r := NewReader(bytes.NewReader(test))
	simdrecords, err := r.ReadAll()
	if err != nil {
		log.Fatalf("%v", err)
	}
	records := EncodingCsv(test)

	if !reflect.DeepEqual(simdrecords, records) {
		t.Errorf("compareAgainstEncodingCsv: got: %v want: %v", simdrecords, records)
	}
}

func EncodingCsv(csvData []byte) (records [][]string) {

	r := csv.NewReader(bytes.NewReader(csvData))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return
}

// filter out commented rows before returning to client
func testIgnoreCommentedLines(t *testing.T, csvData []byte) {

	const comment = '#'

	simdr := NewReader(bytes.NewReader(csvData))
	simdrecords, err := simdr.ReadAll()
	if err != nil {
		log.Fatalf("%v", err)
	}
	FilterOutComments(&simdrecords, comment)

	r := csv.NewReader(bytes.NewReader(csvData))
	r.Comment = comment
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(simdrecords, records) {
		t.Errorf("testIgnoreCommentedLines: got: %v want: %v", simdrecords, records)
	}
}

func TestIgnoreCommentedLines(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		testIgnoreCommentedLines(t, []byte("a,b,c\n#hello,good,bye\nd,e,f\n\n"))
	})
	t.Run("first", func(t *testing.T) {
		testIgnoreCommentedLines(t, []byte("#a,b,c\nhello,good,bye\nd,e,f\n\n"))
	})
	t.Run("last", func(t *testing.T) {
		testIgnoreCommentedLines(t, []byte("a,b,c\nd,e,f\n#IGNORED\n"))
	})
	t.Run("multiple", func(t *testing.T){
		testIgnoreCommentedLines(t, []byte("a,b,c\n#A,B,C\nd,e,f\n#g,h,i\n"))
	})
}

func testFieldsPerRecord(t *testing.T, csvData []byte, fieldsPerRecord int) {

	simdr := NewReader(bytes.NewReader(csvData))
	simdrecords, err := simdr.ReadAll()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// create copy of fieldsPerRecord since it may be changed
	fieldsPerRecordSimd := fieldsPerRecord
	errSimd := EnsureFieldsPerRecord(&simdrecords, &fieldsPerRecordSimd)

	r := csv.NewReader(bytes.NewReader(csvData))
	r.FieldsPerRecord = fieldsPerRecord
	records, err := r.ReadAll()

	// are both returning errors, then this test is a pass
	if errSimd != nil && err != nil {
		fmt.Println(errSimd)
		fmt.Println(err)
		return
	}

	if !reflect.DeepEqual(simdrecords, records) {
		t.Errorf("TestFieldsPerRecord: got: %v want: %v", simdrecords, records)
	}
}

func TestEnsureFieldsPerRecord(t *testing.T) {

	t.Run("match", func(t *testing.T) {
		testFieldsPerRecord(t, []byte("a,b,c\nd,e,f\ng,h,i\n"), 3)
	})
	t.Run("fail", func(t *testing.T) {
		testFieldsPerRecord(t, []byte("a,b,c\nd,e,f\ng,h,i\n"), 4)
	})
	t.Run("variable", func(t *testing.T) {
		testFieldsPerRecord(t, []byte("a,b,c\nd,e\ng\n"), -1)
	})
	t.Run("auto-pass", func(t *testing.T) {
		testFieldsPerRecord(t, []byte("a,b,c\nd,e,f\ng,h,i\n"), 0)
	})
	t.Run("auto-fail", func(t *testing.T) {
		testFieldsPerRecord(t, []byte("a,b,c\nd,e\ng,h\n"), 0)
	})
}

func testTrimLeadingSpace(t *testing.T, csvData []byte) {

	simdr := NewReader(bytes.NewReader(csvData))
	simdrecords, err := simdr.ReadAll()
	if err != nil {
		log.Fatalf("%v", err)
	}
	TrimLeadingSpace(&simdrecords)

	r := csv.NewReader(bytes.NewReader(csvData))
	r.TrimLeadingSpace =true
	records, err := r.ReadAll()
	if err != nil {
		log.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(simdrecords, records) {
		t.Errorf("testTrimLeadingSpace: got: %v want: %v", simdrecords, records)
	}
}

func TestTrimLeadingSpace(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		testTrimLeadingSpace(t, []byte("a,b,c\n d, e, f\n"))
	})
	t.Run("tabs", func(t *testing.T) {
		testTrimLeadingSpace(t, []byte("\tg,h,i\n"))
	})
	t.Run("unicode", func(t *testing.T) {
		testTrimLeadingSpace(t, []byte("j,"+string('\u00A0')+"k,l\n"))
	})
}

func TestExample(t *testing.T) {

	// Example based on https://play.golang.org/p/XxthE8qqZtZ

	instr := `first_name,last_name,username
"Rob","Pike",rob
Ken,Thompson,ken
"Robert","Griesemer","gri"
`

	instr = strings.Replace(instr, "\n", "\r\n", 1)
	instr = strings.Replace(instr, `"Rob"`, `"Ro""b"`, 1) // separator in quote field that is disabled
	instr = strings.Replace(instr, `"Pike"`, `"Pi,ke"`, 1) // separator in quote field that is disabled
	instr = strings.Replace(instr, `"Robert"`, "Rob\r\nert", 1)     // carriage return in quoted field followed by newline --> treated as newline
	instr = strings.Replace(instr, `"Griesemer"`, "Gries\remer", 1) // carriage return in quoted field not followed by newline  --> not treated as newline

	buf := make([]byte, 128)
	copy(buf, instr)

	out := bytes.NewBufferString("")

	fmt.Fprintln(out, hex.Dump(buf))
	fmt.Fprintf(out,"         input: %s", string(bytes.ReplaceAll(bytes.ReplaceAll(buf[:64], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))
	fmt.Fprintf(out,"·%s\n", string(bytes.ReplaceAll(bytes.ReplaceAll(buf[64:], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))

	separatorMasksIn := getBitMasks(buf, byte(','))
	quoteMasksIn := getBitMasks(buf, byte('"'))
	carriageReturnMasksIn := getBitMasks(buf, byte('\r'))
	newlineMasksIn := getBitMasks(buf, byte('\n'))

	input1 := stage1Input{quoteMasksIn[0], separatorMasksIn[0], carriageReturnMasksIn[0], quoteMasksIn[1], 0, newlineMasksIn[0], newlineMasksIn[1]}
	output1_0 := stage1Output{}
	preprocessMasksToMasksInverted(&input1, &output1_0)

	input1 = stage1Input{input1.quoteMaskInNext, separatorMasksIn[1], carriageReturnMasksIn[1], 0, input1.quoted, newlineMasksIn[1], 0}
	output1_1 := stage1Output{}
	preprocessMasksToMasksInverted(&input1, &output1_1)

	fmt.Fprintf(out, diffBitmask(
		fmt.Sprintf("     quote-in : %064b·%064b", bits.Reverse64(quoteMasksIn[0]), bits.Reverse64(quoteMasksIn[1])),
		fmt.Sprintf("     quote-out: %064b·%064b", bits.Reverse64(output1_0.quoteMaskOut), bits.Reverse64(output1_1.quoteMaskOut))))

	fmt.Fprintf(out, diffBitmask(
		fmt.Sprintf(" separator-in : %064b·%064b", bits.Reverse64(separatorMasksIn[0]), bits.Reverse64(separatorMasksIn[1])),
		fmt.Sprintf(" separator-out: %064b·%064b", bits.Reverse64(output1_0.separatorMaskOut), bits.Reverse64(output1_1.separatorMaskOut))))

	fmt.Fprintf(out, diffBitmask(
		fmt.Sprintf("  carriage-in : %064b·%064b", bits.Reverse64(carriageReturnMasksIn[0]), bits.Reverse64(carriageReturnMasksIn[1])),
		fmt.Sprintf("  carriage-out: %064b·%064b", bits.Reverse64(output1_0.carriageReturnMaskOut), bits.Reverse64(output1_1.carriageReturnMaskOut))))

	input2 := NewInput()
	input2.quoteMask = output1_0.quoteMaskOut
	input2.separatorMask = output1_0.separatorMaskOut
	input2.delimiterMask = newlineMasksIn[0] | output1_0.carriageReturnMaskOut

	output2 := Output{}
	output2.columns = &[128]uint64{}
	output2.rows = &[128]uint64{}

	Stage2ParseMasks(&input2, 0, &output2)

	input2.quoteMask = output1_1.quoteMaskOut
	input2.separatorMask = output1_1.separatorMaskOut
	input2.delimiterMask = newlineMasksIn[1] | output1_1.carriageReturnMaskOut

	fmt.Fprintf(out, "%s\n", fmt.Sprintf("     delimiter: %064b·%064b", bits.Reverse64(newlineMasksIn[0] | output1_0.carriageReturnMaskOut), bits.Reverse64(input2.delimiterMask)))
	fmt.Fprintf(out, "%s\n", fmt.Sprintf("     separator: %064b·%064b", bits.Reverse64(output1_0.separatorMaskOut), bits.Reverse64(output1_1.separatorMaskOut)))
	fmt.Fprintf(out,"         input: %s", string(bytes.ReplaceAll(bytes.ReplaceAll(buf[:64], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))
	fmt.Fprintf(out,"·%s\n", string(bytes.ReplaceAll(bytes.ReplaceAll(buf[64:], []byte{0xd}, []byte{0x20}), []byte{0xa}, []byte{0x20})))
	fmt.Fprintf(out, "%s\n", fmt.Sprintf("        quotes: %064b·%064b", bits.Reverse64(output1_0.quoteMaskOut), bits.Reverse64(output1_1.quoteMaskOut)))

	Stage2ParseMasks(&input2, 64, &output2)

	fmt.Println("line",  output2.line)

	for line := 0; line < output2.line; line += 2 {
		start, size := output2.rows[line], output2.rows[line+1]
		for column := uint64(0); column < size; column++ {
			fmt.Fprint(out, string(buf[output2.columns[(start+column)*2]:output2.columns[(start+column)*2]+output2.columns[(start+column)*2+1]]))
			fmt.Fprint(out, " ")
		}
		fmt.Fprintln(out)
	}

	fmt.Println(out.String())
}

func BenchmarkSimdCsv(b *testing.B) {
	b.Run("parking-citations-100K", func(b *testing.B){
		benchmarkSimdCsv(b, "parking-citations-100K.csv", 100000)
	})
	b.Run("worldcitiespop", func(b *testing.B){
		benchmarkSimdCsv(b, "worldcitiespop.csv", 100000)
	})
}

func benchmarkSimdCsv(b *testing.B, file string, lines int) {

	data, _ := ioutil.ReadFile(file)

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	input, output := stage1Input{} ,stage1Output{}

	// TODO: Remove synchronization code
	buf := make([]byte, (len(data) + 128)&^127)

	postProc := make([]uint64, 0, len(buf)>>6)

	rows := make([]uint64, 100000 * 2 * 1.5)
	columns := make([]string, len(rows)*20)
	simdrecords := make([][]string, 0, len(rows))

	for i := 0; i < b.N; i++ {

		//  TODO: Remove synchronization code
		copy(buf, data)
		postProc = postProc[:0]
		stage1_preprocess_buffer(buf[:len(data)], uint64(','), &input, &output, &postProc)

		Stage2ParseBufferEx(buf[:len(data)], '\n', ',', '"', &simdrecords, &rows, &columns)
	}
}

func BenchmarkSimdCsvGo(b *testing.B) {
	b.Run("parking-citations-100K", func(b *testing.B){
		benchmarkSimdCsvGo(b, "parking-citations-100K.csv")
	})
	b.Run("worldcitiespop", func(b *testing.B){
		benchmarkSimdCsvGo(b, "worldcitiespop.csv")
	})
}

func benchmarkSimdCsvGo(b *testing.B, file string) {

	data, _ := ioutil.ReadFile(file)

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewReader(data))
		_, err := r.ReadAll()
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
}
