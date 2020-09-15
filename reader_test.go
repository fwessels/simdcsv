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
	}, {
		Name:   "BareCR",
		Input:  "a,b\rc,d\r\n",
		Output: [][]string{{"a", "b\rc", "d"}},
	}, {
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
	}, {
		Name:             "TrimSpace",
		Input:            " a,  b,   c\n",
		Output:           [][]string{{"a", "b", "c"}},
		TrimLeadingSpace: true,
	}, {
		Name:   "LeadingSpace",
		Input:  " a,  b,   c\n",
		Output: [][]string{{" a", "  b", "   c"}},
	}, {
		Name:    "Comment",
		Input:   "#1,2,3\na,b,c\n#comment",
		Output:  [][]string{{"a", "b", "c"}},
		Comment: '#',
	}, {
		Name:   "NoComment",
		Input:  "#1,2,3\na,b,c",
		Output: [][]string{{"#1", "2", "3"}, {"a", "b", "c"}},
	}, {
		Name:       "LazyQuotes",
		Input:      `a "word","1"2",a","b`,
		Output:     [][]string{{`a "word"`, `1"2`, `a"`, `b`}},
		LazyQuotes: true,
	}, {
		Name:       "BareQuotes",
		Input:      `a "word","1"2",a"`,
		Output:     [][]string{{`a "word"`, `1"2`, `a"`}},
		LazyQuotes: true,
	}, {
		Name:       "BareDoubleQuotes",
		Input:      `a""b,c`,
		Output:     [][]string{{`a""b`, `c`}},
		LazyQuotes: true,
	}, {
		Name:   "FieldCount",
		Input:  "a,b,c\nd,e",
		Output: [][]string{{"a", "b", "c"}, {"d", "e"}},
	}, {
		Name:   "TrailingCommaEOF",
		Input:  "a,b,c,",
		Output: [][]string{{"a", "b", "c", ""}},
	}, {
		Name:   "TrailingCommaEOL",
		Input:  "a,b,c,\n",
		Output: [][]string{{"a", "b", "c", ""}},
	}, {
		Name:             "TrailingCommaSpaceEOF",
		Input:            "a,b,c, ",
		Output:           [][]string{{"a", "b", "c", ""}},
		TrimLeadingSpace: true,
	}, {
		Name:             "TrailingCommaSpaceEOL",
		Input:            "a,b,c, \n",
		Output:           [][]string{{"a", "b", "c", ""}},
		TrimLeadingSpace: true,
	}, {
		Name:             "TrailingCommaLine3",
		Input:            "a,b,c\nd,e,f\ng,hi,",
		Output:           [][]string{{"a", "b", "c"}, {"d", "e", "f"}, {"g", "hi", ""}},
		TrimLeadingSpace: true,
	}, {
		Name:   "NotTrailingComma3",
		Input:  "a,b,c, \n",
		Output: [][]string{{"a", "b", "c", " "}},
	}, {
		Name: "CommaFieldTest",
		Input: `x,y,z,w
x,y,z,
x,y,,
x,,,
,,,
"x","y","z","w"
"x","y","z",""
"x","y","",""
"x","","",""
"","","",""
`,
		Output: [][]string{
			{"x", "y", "z", "w"},
			{"x", "y", "z", ""},
			{"x", "y", "", ""},
			{"x", "", "", ""},
			{"", "", "", ""},
			{"x", "y", "z", "w"},
			{"x", "y", "z", ""},
			{"x", "y", "", ""},
			{"x", "", "", ""},
			{"", "", "", ""},
		},
	}, {
		Name:  "TrailingCommaIneffective1",
		Input: "a,b,\nc,d,e",
		Output: [][]string{
			{"a", "b", ""},
			{"c", "d", "e"},
		},
		TrimLeadingSpace: true,
	}, {
		Name:  "ReadAllReuseRecord",
		Input: "a,b\nc,d",
		Output: [][]string{
			{"a", "b"},
			{"c", "d"},
		},
		ReuseRecord: true,
	}, {
		Name:  "CRLFInQuotedField", // Issue 21201
		Input: "A,\"Hello\r\nHi\",B\r\n",
		Output: [][]string{
			{"A", "Hello\nHi", "B"},
		},
	}, {
		Name:   "BinaryBlobField", // Issue 19410
		Input:  "x09\x41\xb4\x1c,aktau",
		Output: [][]string{{"x09A\xb4\x1c", "aktau"}},
	}, {
		Name:             "NonASCIICommaAndComment",
		Input:            "a£b,c£ \td,e\n€ comment\n",
		Output:           [][]string{{"a", "b,c", "d,e"}},
		TrimLeadingSpace: true,
		Comma:            '£',
		Comment:          '€',
	}, {
		Name:    "NonASCIICommaAndCommentWithQuotes",
		Input:   "a€\"  b,\"€ c\nλ comment\n",
		Output:  [][]string{{"a", "  b,", " c"}},
		Comma:   '€',
		Comment: 'λ',
	}, {
		// λ and θ start with the same byte.
		// This tests that the parser doesn't confuse such characters.
		Name:    "NonASCIICommaConfusion",
		Input:   "\"abθcd\"λefθgh",
		Output:  [][]string{{"abθcd", "efθgh"}},
		Comma:   'λ',
		Comment: '€',
	}, {
		Name:    "NonASCIICommentConfusion",
		Input:   "λ\nλ\nθ\nλ\n",
		Output:  [][]string{{"λ"}, {"λ"}, {"λ"}},
		Comment: 'θ',
	}, {
		Name:   "QuotedFieldMultipleLF",
		Input:  "\"\n\n\n\n\"",
		Output: [][]string{{"\n\n\n\n"}},
	}, {
		Name:  "MultipleCRLF",
		Input: "\r\n\r\n\r\n\r\n",
	}, {
		// The implementation may read each line in several chunks if it doesn't fit entirely
		// in the read buffer, so we should test the code to handle that condition.
		Name:    "HugeLines",
		Input:   strings.Repeat("#ignore\n", 10000) + strings.Repeat("@", 5000) + "," + strings.Repeat("*", 5000),
		Output:  [][]string{{strings.Repeat("@", 5000), strings.Repeat("*", 5000)}},
		Comment: '#',
	}, {
		Name:       "LazyQuoteWithTrailingCRLF",
		Input:      "\"foo\"bar\"\r\n",
		Output:     [][]string{{`foo"bar`}},
		LazyQuotes: true,
	}, {
		Name:   "DoubleQuoteWithTrailingCRLF",
		Input:  "\"foo\"\"bar\"\r\n",
		Output: [][]string{{`foo"bar`}},
	}, {
		Name:   "EvenQuotes",
		Input:  `""""""""`,
		Output: [][]string{{`"""`}},
	}, {
		Name:       "LazyOddQuotes",
		Input:      `"""""""`,
		Output:     [][]string{{`"""`}},
		LazyQuotes: true,
	}, {
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
