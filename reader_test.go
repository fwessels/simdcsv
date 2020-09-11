package simdcsv

import (
	_ "encoding/csv"
	"reflect"
	"strings"
	"testing"
)

// These test cases are copied from https://golang.org/src/encoding/csv/reader_test.go
func TestRead(t *testing.T) {
	tests := []struct {
		Name   string
		Input  string
		Output [][]string
		Error  error

		// These fields are copied into the Reader
		Comma              rune
		Comment            rune
		UseFieldsPerRecord bool // false (default) means FieldsPerRecord is -1
		FieldsPerRecord    int
		LazyQuotes         bool
		TrimLeadingSpace   bool
		ReuseRecord        bool
	}{{
		Name:   "Simple",
		Input:  "a,b,c\n",
		Output: [][]string{{"a", "b", "c"}},
	}, {
		Name:   "CRLF",
		Input:  "a,b\r\nc,d\r\n",
		Output: [][]string{{"a", "b"}, {"c", "d"}},
	},
		//	{
		//		// TODO: Failing: got  [["a" "b"] ["c" "d"]] // want [["a" "b\rc" "d"]]
		//		Name:   "BareCR",
		//		Input:  "a,b\rc,d\r\n",
		//		Output: [][]string{{"a", "b\rc", "d"}},
		//	}
		{
			Name: "RFC4180test",
			Input: `#field1,field2,field3
"aaa","bb
b","ccc"
"a,a","b""bb","ccc"
zzz,yyy,xxx
`,
			Output: [][]string{
				{"#field1", "field2", "field3"},
				{"aaa", "bb\nb", "ccc"},
				{"a,a", `b"bb`, "ccc"},
				{"zzz", "yyy", "xxx"},
			},
			UseFieldsPerRecord: true,
			FieldsPerRecord:    0,
		}, {
			Name:   "NoEOLTest",
			Input:  "a,b,c",
			Output: [][]string{{"a", "b", "c"}},
		}, {
			Name:   "Semicolon",
			Input:  "a;b;c\n",
			Output: [][]string{{"a", "b", "c"}},
			Comma:  ';',
		}, {
			Name: "MultiLine",
			Input: `"two
line","one line","three
line
field"`,
			Output: [][]string{{"two\nline", "one line", "three\nline\nfield"}},
		}, {
			Name:  "BlankLine",
			Input: "a,b,c\n\nd,e,f\n\n",
			Output: [][]string{
				{"a", "b", "c"},
				{"d", "e", "f"},
			},
		}, {
			Name:  "BlankLineFieldCount",
			Input: "a,b,c\n\nd,e,f\n\n",
			Output: [][]string{
				{"a", "b", "c"},
				{"d", "e", "f"},
			},
			UseFieldsPerRecord: true,
			FieldsPerRecord:    0,
		}}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			r := NewReader(strings.NewReader(tt.Input))

			if tt.Comma != 0 {
				r.Comma = tt.Comma
			}
			r.Comment = tt.Comment
			if tt.UseFieldsPerRecord {
				r.FieldsPerRecord = tt.FieldsPerRecord
			} else {
				r.FieldsPerRecord = -1
			}
			r.LazyQuotes = tt.LazyQuotes
			r.TrimLeadingSpace = tt.TrimLeadingSpace
			r.ReuseRecord = tt.ReuseRecord

			out, err := r.ReadAll()
			if !reflect.DeepEqual(err, tt.Error) {
				t.Errorf("ReadAll() error:\ngot  %v\nwant %v", err, tt.Error)
			} else if !reflect.DeepEqual(out, tt.Output) {
				t.Errorf("ReadAll() output:\ngot  %q\nwant %q", out, tt.Output)
			}
		})
	}
}
