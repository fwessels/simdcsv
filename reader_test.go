package simdcsv

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"
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
		Name:   "TrailingCR",
		Input:  "field1,field2\r",
		Output: [][]string{{"field1", "field2"}},
	}, {
		Name:   "QuotedTrailingCR",
		Input:  "\"field\"\r",
		Output: [][]string{{"field"}},
	}, {
		Name:   "FieldCR",
		Input:  "field\rfield\r",
		Output: [][]string{{"field\rfield"}},
	}, {
		Name:   "FieldCRCR",
		Input:  "field\r\rfield\r\r",
		Output: [][]string{{"field\r\rfield\r"}},
	}, {
		Name:   "FieldCRCRLF",
		Input:  "field\r\r\nfield\r\r\n",
		Output: [][]string{{"field\r"}, {"field\r"}},
	}, {
		Name:   "FieldCRCRLFCRCR",
		Input:  "field\r\r\n\r\rfield\r\r\n\r\r",
		Output: [][]string{{"field\r"}, {"\r\rfield\r"}, {"\r"}},
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
		Name:               "BadFieldCount",
		Input:              "a,b,c\nd,e",
		Error:              &csv.ParseError{StartLine: 2, Line: 2, Err: csv.ErrFieldCount},
		UseFieldsPerRecord: true,
		FieldsPerRecord:    0,
	}, {
		Name:               "BadFieldCount1",
		Input:              `a,b,c`,
		Error:              &csv.ParseError{StartLine: 1, Line: 1, Err: csv.ErrFieldCount},
		UseFieldsPerRecord: true,
		FieldsPerRecord:    2,
	}, {
		Name:  "BadDoubleQuotes",
		Input: `a""b,c`,
		Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 1, Err: csv.ErrBareQuote},
	}, {
		Name:  "BadBareQuote",
		Input: `a "word","b"`,
		Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 2, Err: csv.ErrBareQuote},
	}, {
		Name:  "BadTrailingQuote",
		Input: `"a word",b"`,
		Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 10, Err: csv.ErrBareQuote},
	}, {
		Name:  "ExtraneousQuote",
		Input: `"a "word","b"`,
		Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 3, Err: csv.ErrQuote},
	}, {
		Name:  "StartLine1", // Issue 19019
		Input: "a,\"b\nc\"d,e",
		Error: &csv.ParseError{StartLine: 1, Line: 2, Column: 1, Err: csv.ErrQuote},
	}, {
		Name:  "QuoteWithTrailingCRLF",
		Input: "\"foo\"bar\"\r\n",
		Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 4, Err: csv.ErrQuote},
	}, {
		Name:  "BadComma1",
		Comma: '\n',
		Error: errInvalidDelim,
	}, {
		Name:  "BadComma2",
		Comma: '\r',
		Error: errInvalidDelim,
	}, {
		Name:  "BadComma3",
		Comma: '"',
		Error: errInvalidDelim,
	}, {
		Name:  "BadComma4",
		Comma: utf8.RuneError,
		Error: errInvalidDelim,
	}, {
		Name:    "BadComment1",
		Comment: '\n',
		Error:   errInvalidDelim,
	}, {
		Name:    "BadComment2",
		Comment: '\r',
		Error:   errInvalidDelim,
	}, {
		Name:    "BadComment3",
		Comment: utf8.RuneError,
		Error:   errInvalidDelim,
	}, {
		Name:    "BadCommaComment",
		Comma:   'X',
		Comment: 'X',
		Error:   errInvalidDelim,
	}, {
		Name:  "QuotedTrailingCRCR",
		Input: "\"field\"\r\r",
		Error: &csv.ParseError{StartLine: 1, Line: 1, Column: 6, Err: csv.ErrQuote},
	},
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
