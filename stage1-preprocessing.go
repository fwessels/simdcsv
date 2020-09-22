package simdcsv

import (
	"math/bits"
	"reflect"
	"unsafe"
)

// Substitute values when preprocessing a chunk
// NB 0x0 should be avoided (since trailing bytes
// beyond the end of the buffer are zeroed out)
const preprocessedSeparator = 0x2
const preprocessedQuote = 0x3

type stage1Input struct {
	quoteMaskIn          uint64
	separatorMaskIn      uint64
	carriageReturnMaskIn uint64
	quoteMaskInNext      uint64
	quoted               uint64
	newlineMaskIn        uint64
	newlineMaskInNext    uint64
}

type stage1Output struct {
	quoteMaskOut          uint64
	separatorMaskOut      uint64
	carriageReturnMaskOut uint64
	needsPostProcessing   uint64
}

func preprocessMasksToMasksInverted(input *stage1Input, output *stage1Output) {

	const clearMask = 0xfffffffffffffffe

	separatorMaskIn := input.separatorMaskIn
	carriageReturnMaskIn := input.carriageReturnMaskIn
	quoteMaskIn := input.quoteMaskIn

	separatorPos := bits.TrailingZeros64(separatorMaskIn)
	carriageReturnPos := bits.TrailingZeros64(carriageReturnMaskIn)
	quotePos := bits.TrailingZeros64(quoteMaskIn)

	output.quoteMaskOut = quoteMaskIn                   // copy quote mask to output
	output.separatorMaskOut = separatorMaskIn           // copy separator mask to output
	output.carriageReturnMaskOut = carriageReturnMaskIn // copy carriage return mask to output
	output.needsPostProcessing = 0                      // flag to indicate whether post-processing is need for these masks

	for {
		if quotePos < separatorPos && quotePos < carriageReturnPos {

			if input.quoted != 0 && quotePos == 63 && input.quoteMaskInNext&1 == 1 { // last bit of quote mask and first bit of next quote mask set?
				// clear out both active bit and ...
				quoteMaskIn &= clearMask << quotePos
				output.quoteMaskOut &= ^(uint64(1) << quotePos) // mask out quote
				output.needsPostProcessing = 1                  // post-processing is required for double quotes
				// first bit of next quote mask
				input.quoteMaskInNext &= ^uint64(1)
			} else if input.quoted != 0 && quoteMaskIn&(1<<(quotePos+1)) != 0 { // next quote bit is also set (so two adjacent bits) ?
				// clear out both active bit and subsequent bit
				quoteMaskIn &= clearMask << (quotePos + 1)
				output.quoteMaskOut &= ^(uint64(3) << quotePos) // mask out two quotes
				output.needsPostProcessing = 1                  // post-processing is required for double quotes
			} else {
				input.quoted = ^input.quoted

				quoteMaskIn &= clearMask << quotePos
			}

			quotePos = bits.TrailingZeros64(quoteMaskIn)

		} else if separatorPos < quotePos && separatorPos < carriageReturnPos {

			if input.quoted != 0 {
				output.separatorMaskOut &= ^(uint64(1) << separatorPos) // mask out separator bit in quoted field
			}

			separatorMaskIn &= clearMask << separatorPos
			separatorPos = bits.TrailingZeros64(separatorMaskIn)

		} else if carriageReturnPos < quotePos && carriageReturnPos < separatorPos {

			if input.quoted != 0 {
				output.carriageReturnMaskOut &= ^(uint64(1) << carriageReturnPos) // mask out carriage return bit in quoted field
				output.needsPostProcessing = 1                                    // post-processing is required for carriage returns in quoted fields
			} else {
				if carriageReturnPos == 63 && input.newlineMaskInNext&1 == 0 {
					output.carriageReturnMaskOut &= ^(uint64(1) << carriageReturnPos) // mask out carriage return for replacement without following newline
				} else if input.newlineMaskIn&(uint64(1)<<(carriageReturnPos+1)) == 0 {
					output.carriageReturnMaskOut &= ^(uint64(1) << carriageReturnPos) // mask out carriage return bit in quoted field
				}
			}

			carriageReturnMaskIn &= clearMask << carriageReturnPos
			carriageReturnPos = bits.TrailingZeros64(carriageReturnMaskIn)

		} else {
			// we must be done
			break
		}
	}

	return
}

type postProcRow struct {
	start int
	end  int
}

//
// Determine which rows and columns need post processing
// This is  need to replace both "" to " as well as
// \r\n to \n in specific columns
func getPostProcRows(buf []byte, postProc []uint64, simdrecords [][]string) (ppRows []postProcRow) {

	// TODO: Crude implementation, make more refined/granular

	sliceptr := func(slc []byte) uintptr {
		return (*reflect.SliceHeader)(unsafe.Pointer(&slc)).Data
	}
	stringptr := func (s string) uintptr {
		return (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	}

	ppRows = make([]postProcRow, 0, 128)

	row, pbuf := 0, sliceptr(buf)
	for ipp, pp := range postProc {

		if ipp < len(postProc) - 1 && pp == postProc[ipp+1]  {
			continue // if offset occurs multiple times, process only last one
		}

		// find start row to process
		for row < len(simdrecords) && uint64(stringptr(simdrecords[row][0])-pbuf) < pp {
			row++
		}

		ppr := postProcRow{}
		if row > 0 {
			ppr.start = row-1
		}

		// find end row to process
		for row < len(simdrecords) && uint64(stringptr(simdrecords[row][0])-pbuf) < pp+64 {
			row++
		}
		ppr.end = row

		ppRows = append(ppRows, ppr)
	}
	return
}
