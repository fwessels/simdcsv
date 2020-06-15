package simdcsv

import (
	"io/ioutil"
	"testing"
)

func TestCsv(t *testing.T) {

	buf, _ := ioutil.ReadFile("parking-citations-10.csv")

	incrs := stage1(buf)
	stage2(buf, incrs)
}

func TestCsvSpaces(t *testing.T) {

	// benchmark approach with spaces
}

func TestCsvEscapes(t *testing.T) {

	// benchmark approach with escaped fields
}
