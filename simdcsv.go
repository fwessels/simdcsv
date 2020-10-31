package simdcsv

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Below is the same interface definition from encoding/csv

// A Reader reads records from a CSV-encoded file.
//
// As returned by NewReader, a Reader expects input conforming to RFC 4180.
// The exported fields can be changed to customize the details before the
// first call to Read or ReadAll.
//
// The Reader converts all \r\n sequences in its input to plain \n,
// including in multiline field values, so that the returned data does
// not depend on which line-ending convention an input file uses.
type Reader struct {
	// Comma is the field delimiter.
	// It is set to comma (',') by NewReader.
	// Comma must be a valid rune and must not be \r, \n,
	// or the Unicode replacement character (0xFFFD).
	Comma rune

	// Comment, if not 0, is the comment character. Lines beginning with the
	// Comment character without preceding whitespace are ignored.
	// With leading whitespace the Comment character becomes part of the
	// field, even if TrimLeadingSpace is true.
	// Comment must be a valid rune and must not be \r, \n,
	// or the Unicode replacement character (0xFFFD).
	// It must also not be equal to Comma.
	Comment rune

	// FieldsPerRecord is the number of expected fields per record.
	// If FieldsPerRecord is positive, Read requires each record to
	// have the given number of fields. If FieldsPerRecord is 0, Read sets it to
	// the number of fields in the first record, so that future records must
	// have the same field count. If FieldsPerRecord is negative, no check is
	// made and records may have a variable number of fields.
	FieldsPerRecord int

	// If LazyQuotes is true, a quote may appear in an unquoted field and a
	// non-doubled quote may appear in a quoted field.
	LazyQuotes bool

	// If TrimLeadingSpace is true, leading white space in a field is ignored.
	// This is done even if the field delimiter, Comma, is white space.
	TrimLeadingSpace bool

	ReuseRecord bool   // Deprecated: Unused by simdcsv.
	TrailingComma bool // Deprecated: No longer used.

	r *bufio.Reader
}

var errInvalidDelim = errors.New("csv: invalid field or comment delimiter")

func validDelim(r rune) bool {
	return r != 0 && r != '"' && r != '\r' && r != '\n' && utf8.ValidRune(r) && r != utf8.RuneError
}

// NewReader returns a new Reader that reads from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		Comma: ',',
		r:     bufio.NewReader(r),
	}
}

// Each record is a slice of fields.
// A successful call returns err == nil, not err == io.EOF. Because ReadAll is
// defined to read until EOF, it does not treat end of file as an error to be
// reported.
func (r *Reader) ReadAll() ([][]string, error) {

	fallback := func(ioReader io.Reader) ([][]string, error) {
		rCsv := csv.NewReader(ioReader)
		rCsv.LazyQuotes = r.LazyQuotes
		rCsv.TrimLeadingSpace = r.TrimLeadingSpace
		rCsv.Comment = r.Comment
		rCsv.Comma = r.Comma
		rCsv.FieldsPerRecord = r.FieldsPerRecord
		rCsv.ReuseRecord = r.ReuseRecord
		return rCsv.ReadAll()
	}

	if r.Comma == r.Comment || !validDelim(r.Comma) || (r.Comment != 0 && !validDelim(r.Comment)) {
		return nil, errInvalidDelim
	}

	if r.LazyQuotes ||
		r.Comma != 0 && r.Comma > unicode.MaxLatin1 ||
		r.Comment != 0 && r.Comment > unicode.MaxLatin1 {
		return fallback(r.r)
	}

	buf, err := r.r.ReadBytes(0)
	if err != nil && err != io.EOF {
		return nil, err
	}

	masks, postProc := Stage1PreprocessBuffer(buf, uint64(r.Comma))

	records, parseError := Stage2ParseBuffer(buf, masks, '\n', nil)
	if parseError {
		return fallback(bytes.NewReader(buf))
	}

	for _, ppr := range getPostProcRows(buf, postProc, records) {
		for r := ppr.start; r < ppr.end; r++ {
			for c := range records[r] {
				records[r][c] = strings.ReplaceAll(records[r][c], "\"\"", "\"")
				records[r][c] = strings.ReplaceAll(records[r][c], "\r\n", "\n")
			}
		}
	}

	if r.Comment != 0 {
		FilterOutComments(&records, byte(r.Comment))
	}
	if r.TrimLeadingSpace {
		TrimLeadingSpace(&records)
	}

	// create copy of fieldsPerRecord since it may be changed
	fieldsPerRecord := r.FieldsPerRecord
	if errSimd := EnsureFieldsPerRecord(&records, &fieldsPerRecord); errSimd != nil {
		return fallback(bytes.NewReader(buf))
	}

	if len(records) == 0 {
		return nil, nil
	} else {
		return records, nil
	}
}

type ChunkInfo struct {
	chunk    []byte
	masks    []uint64
	postProc []uint64
	header   uint64
	trailer  uint64
	splitRow []byte
	lastChunk bool
}

type ReadOutput struct {
	records [][]string
	err     error
}

// ReadAllStreaming reads all the remaining records from r.
// Each record is a slice of fields.
// A successful call returns err == nil, not err == io.EOF. Because ReadAll is
// defined to read until EOF, it does not treat end of file as an error to be
// reported.
func (r *Reader) ReadAllStreaming(out chan ReadOutput) {

	fallback := func(ioReader io.Reader) ReadOutput {
		rCsv := csv.NewReader(ioReader)
		rCsv.LazyQuotes = r.LazyQuotes
		rCsv.TrimLeadingSpace = r.TrimLeadingSpace
		rCsv.Comment = r.Comment
		rCsv.Comma = r.Comma
		rCsv.FieldsPerRecord = r.FieldsPerRecord
		rCsv.ReuseRecord = r.ReuseRecord
		rcds, err := rCsv.ReadAll()
		return ReadOutput{rcds, err}
	}

	if r.Comma == r.Comment || !validDelim(r.Comma) || (r.Comment != 0 && !validDelim(r.Comment)) {
		go func() {
			out <- ReadOutput{nil, errInvalidDelim}
			close(out)
		}()
		return
	}

	if r.LazyQuotes ||
		r.Comma != 0 && r.Comma > unicode.MaxLatin1 ||
		r.Comment != 0 && r.Comment > unicode.MaxLatin1 {
		go func() {
			out <- fallback(r.r)
			close(out)
		}()
		return
	}

	buf, err := r.r.ReadBytes(0)
	if err != nil && err != io.EOF {
		go func() {
			out <- ReadOutput{nil, err}
			close(out)
		}()
		return
	}

	chunkSize := 1024 * 256

	// round chunkSize to next multiple of 64
	chunkSize = (chunkSize + 63) &^ 63
	masksSize := ((chunkSize >> 6) + 2) * 3 // add 2 extra slots as safety for masks

	chunks := make(chan ChunkInfo)
	splitRow := make([]byte, 0, 256)

	go func() {

		quoted := uint64(0) // initialized quoted state to unquoted

		for offset := 0; offset < len(buf); offset += chunkSize {
			var chunk []byte
			lastChunk := offset+chunkSize >= len(buf)
			if lastChunk {
				chunk = buf[offset:]
			} else {
				chunk = buf[offset : offset+chunkSize]
			}

			// TODO: Use memory pool
			postProcStream := make([]uint64, 0, ((chunkSize>>6)+1)*2)
			masksStream := make([]uint64, masksSize)
			masksStream, postProcStream, quoted = Stage1PreprocessBufferEx(chunk, uint64(r.Comma), quoted, &masksStream, &postProcStream)

			header, trailer := uint64(0), uint64(0)

			if offset > 0 {
				for index := 0; index < len(masksStream); index += 3 {
					hr := bits.TrailingZeros64(masksStream[index])
					header += uint64(hr)
					if hr < 64 {
						// upon finding the first delimiter bit, we can break out
						// (since any adjacent delimiter bits, whether representing a newline or a carriage return,
						//  are treated as empty lines anyways)
						break
					}
				}
				if header == uint64(len(masksStream))/3*64 {
					// we have not found a newline delimiter, so set
					// header to size of chunk (meaning we will be skipping simd processing)
					header = uint64(len(chunk))
				}
			}

			if !lastChunk && header < uint64(len(chunk)) {
				for index := 3; index < len(masksStream); index += 3 {
					tr := bits.LeadingZeros64(masksStream[len(masksStream)-index])
					trailer += uint64(tr)
					if tr < 64 {
						break
					}
				}
			}

			splitRow = append(splitRow, chunk[:header]...)

			if header < uint64(len(chunk)) {
				chunks <- ChunkInfo{chunk, masksStream, postProcStream, header, trailer, splitRow, lastChunk}
			}  else {
				chunks <- ChunkInfo{nil, nil, nil, 0, 0, splitRow, lastChunk}
			}

			splitRow = make([]byte, 0, 128)
			splitRow = append(splitRow, chunk[len(chunk)-int(trailer):]...)
		}
		close(chunks)
	}()

	go func() {

		for chunkInfo := range chunks {

			simdrecords := make([][]string, 0, 1024)

			if len(chunkInfo.splitRow) > 0 { // first append the row split between chunks
				records := EncodingCsv(chunkInfo.splitRow)
				simdrecords = append(simdrecords, records...)
			}

			if chunkInfo.chunk != nil {
				rows := make([]uint64, chunkSize/256*3*1)
				columns := make([]string, len(rows)*20)
				inputStage2, outputStage2 := NewInput(), OutputAsm{}

				outputStage2.strData = chunkInfo.header & 0x3f // reinit strData for every chunk (fields do not span chunks)

				skip := chunkInfo.header >> 6
				shift := chunkInfo.header & 0x3f

				chunkInfo.masks[skip*3+0] &= ^uint64((1 << shift) - 1)
				chunkInfo.masks[skip*3+1] &= ^uint64((1 << shift) - 1)
				chunkInfo.masks[skip*3+2] &= ^uint64((1 << shift) - 1)

				skipTz := (chunkInfo.trailer >> 6) + 1
				shiftTz := chunkInfo.trailer & 0x3f

				chunkInfo.masks[len(chunkInfo.masks)-int(skipTz)*3+0] <<= shiftTz
				chunkInfo.masks[len(chunkInfo.masks)-int(skipTz)*3+1] <<= shiftTz
				chunkInfo.masks[len(chunkInfo.masks)-int(skipTz)*3+2] <<= shiftTz
				chunkInfo.masks[len(chunkInfo.masks)-int(skipTz)*3+0] >>= shiftTz
				chunkInfo.masks[len(chunkInfo.masks)-int(skipTz)*3+1] >>= shiftTz
				chunkInfo.masks[len(chunkInfo.masks)-int(skipTz)*3+2] >>= shiftTz

				var parsingError bool
				rows, columns, parsingError = Stage2ParseBufferExStreaming(chunkInfo.chunk[skip*0x40:len(chunkInfo.chunk)-int(chunkInfo.trailer)], chunkInfo.masks[skip*3:], '\n', &inputStage2, &outputStage2, &rows, &columns)
				if parsingError {
					out <- fallback(bytes.NewReader(buf))
					break
				}

				columns = columns[:(outputStage2.index)/2]
				rows = rows[:outputStage2.line]

				for line := 0; line < outputStage2.line; line += 2 {
					simdrecords = append(simdrecords, columns[rows[line]:rows[line]+rows[line+1]])
				}

				// TODO: Check whether post processing array is in sync (with offset into buffer)
				if chunkInfo.postProc != nil {
//					for _, ppr := range getPostProcRows(chunkInfo.chunk[skip*0x40:len(chunkInfo.chunk)-int(chunkInfo.trailer)], chunkInfo.postProc, simdrecords) {
						for r := 0/*ppr.start*/; r < len(simdrecords)/*ppr.end*/; r++ {
							for c := range simdrecords[r] {
								simdrecords[r][c] = strings.ReplaceAll(simdrecords[r][c], "\"\"", "\"")
								simdrecords[r][c] = strings.ReplaceAll(simdrecords[r][c], "\r\n", "\n")
							}
						}
//					}
				}

			}


			if r.Comment != 0 {
				FilterOutComments(&simdrecords, byte(r.Comment))
			}
			if r.TrimLeadingSpace {
				TrimLeadingSpace(&simdrecords)
			}

			// create copy of fieldsPerRecord since it may be changed
			fieldsPerRecord := r.FieldsPerRecord
			if errSimd := EnsureFieldsPerRecord(&simdrecords, &fieldsPerRecord); errSimd != nil {
				out <- fallback(bytes.NewReader(buf))
				break
			}

			out <- ReadOutput{simdrecords, nil}

			if !chunkInfo.lastChunk {
				splitRow = chunkInfo.splitRow
			}
		}

		close(out)
	}()

	return
}

func (r *Reader) ReadAll2() ([][]string, error) {

	out := make(chan ReadOutput)
	r.ReadAllStreaming(out)

	records := make([][]string, 0)
	for rcrds := range out {
		if rcrds.err != nil {
			// drain channel
			for _ = range out{}
			return nil, rcrds.err
		}
		records = append(records, rcrds.records...)
	}

	if len(records) == 0 {
		return nil, nil
	} else {
		return records, nil
	}
}

func FilterOutComments(records *[][]string, comment byte) {

	// iterate in reverse so as to prevent starting over when removing element
	for i := len(*records) - 1; i >= 0; i-- {
		record := (*records)[i]
		if len(record) > 0 && len(record[0]) > 0 && record[0][0] == comment {
			*records = append((*records)[:i], (*records)[i+1:len(*records)]...)
		}
	}
}

func EnsureFieldsPerRecord(records *[][]string, fieldsPerRecord *int) error {

	if *fieldsPerRecord == 0 {
		if len(*records) > 0 {
			*fieldsPerRecord = len((*records)[0])
		}
	}

	if *fieldsPerRecord > 0 {
		for i, record := range *records {
			if len(record) != *fieldsPerRecord {
				*records = nil
				return errors.New(fmt.Sprintf("record on line %d: wrong number of fields", i+1))
			}
		}
	}
	return nil
}

func TrimLeadingSpace(records *[][]string) {

	for i := 0; i < len(*records); i++ {
		for j := range (*records)[i] {
			(*records)[i][j] = strings.TrimLeftFunc((*records)[i][j], func(r rune) bool {
				return unicode.IsSpace(r)
			})
		}
	}
}

func allocMasks(buf []byte) []uint64 {
	return make([]uint64, ((len(buf)>>6)+4)*3)
}
