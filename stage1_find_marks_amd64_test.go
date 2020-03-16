package simdcsv

import (
	"fmt"
	"testing"
	"strings"
	"math/bits"
)

func TestStage1(t *testing.T) {

	vectors := []string{
		"1103341116,2015-12-21T00:00:00,1251,,,CA,200304,,HOND,PA,GY,13147 WELBY WAY,01521,1,4000A1,NO EVIDENCE OF REG,50,99999,99999    ",
		"          1                   1    111  1      11    1  1  1               1     1 1      1                  1  1     1         ",
		"1103700150,2015-12-21T00:00:00,1435,,,CA,201512,,GMC,VN,WH,525 S MAIN ST,1C51,1,4000A1,NO EVIDENCE OF REG,50,99999,99999        ",
		"          1                   1    111  1      11   1  1  1             1    1 1      1                  1  1     1             ",
		"1104803000,2015-12-21T00:00:00,2055,,,CA,201503,,NISS,PA,BK,200 WORLD WAY,2R2,2,8939,WHITE CURB,58,6439997.9,1802686.4          ",
		"          1                   1    111  1      11    1  1  1             1   1 1    1          1  1         1                   ",
		"1104820732,2015-12-26T00:00:00,1515,,,CA,,,ACUR,PA,WH,100 WORLD WAY,2F11,2,000,17104h,,6440041.1,1802686.2                      ",
		"          1                   1    111  111    1  1  1             1    1 1   1      11         1                               ",
		"1105461453,2015-09-15T00:00:00,115,,,CA,200316,,CHEV,PA,BK,GEORGIA ST/OLYMPIC,1FB70,1,8069A,NO STOPPING/STANDING,93,99999,99999 ",
		"          1                   1   111  1      11    1  1  1                  1     1 1     1                    1  1     1      ",
		"1106226590,2015-09-15T00:00:00,19,,,CA,201507,,CHEV,VN,GY,SAN PEDRO S/O BOYD,1A35W,1,4000A1,NO EVIDENCE OF REG,50,99999,99999   ",
		"          1                   1  111  1      11    1  1  1                  1     1 1      1                  1  1     1        ",
		"1106500452,2015-12-17T00:00:00,1710,,,CA,201605,,MAZD,PA,BL,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999   ",
		"          1                   1    111  1      11    1  1  1               1     1 1    1                    1   1     1        ",
		"1106500463,2015-12-17T00:00:00,1710,,,CA,201602,,TOYO,PA,BK,SUNSET/ALVARADO,00217,1,8070,PARK IN GRID LOCK ZN,163,99999,99999   ",
		"          1                   1    111  1      11    1  1  1               1     1 1    1                    1   1     1        ",
		"1106506402,2015-12-22T00:00:00,945,,,CA,201605,,CHEV,PA,BR,721 S WESTLAKE,2A75,1,8069AA,NO STOP/STAND AM,93,99999,99999         ",
		"          1                   1   111  1      11    1  1  1              1    1 1      1                1  1     1              ",
		"1106506413,2015-12-22T00:00:00,1100,,,CA,201701,,NISS,PA,SI,1159 HUNTLEY DR,2A75,1,8069AA,NO STOP/STAND AM,93,99999,99999       ",
		"          1                   1    111  1      11    1  1  1               1    1 1      1                1  1     1            ",
	}

	for i := 0; i < len(vectors); i += 2 {

		v := vectors[i]
		m1 := find_separator([]byte(v), ',')
		m2 := find_separator([]byte(v[64:]), ',')

		mask := fmt.Sprintf("%064b%064b", bits.Reverse64(m1), bits.Reverse64(m2))
		mask = strings.ReplaceAll(mask, "0", " ")

		if mask != vectors[i+1] {
			t.Errorf("TestStage1(%d): got: %s want: %s", i, mask, vectors[i+1])
		}
	}
}