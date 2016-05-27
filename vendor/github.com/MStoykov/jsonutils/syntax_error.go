package jsonutils

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

const (
	newline  = '\n'
	emptyMsg = `in no json`
)

// SyntaxError wraps around json.SyntaxError to provide better context by it's
// Error().The original error is available through the method Original
type SyntaxError struct {
	original *json.SyntaxError
	msg      string
}

func (s *SyntaxError) Error() string {
	return s.msg
}

// Original returns the original json.SyntaxError
func (s *SyntaxError) Original() *json.SyntaxError {
	return s.original
}

// NewSyntaxError returns a SyntaxError wrapping around json.SyntaxError.
// The provided jsonContents should be the input that produced the error
// the contextSize is the number of lines around the to be printed as well
//
// Notice: in the message all the tabs are replaced by a single space
func NewSyntaxError(original *json.SyntaxError, jsonContents []byte, contextSize int64) *SyntaxError {
	if len(jsonContents) == 0 {
		return &SyntaxError{
			original: original,
			// replace the tabs so that the offset is correct
			msg: strings.Join(
				[]string{"with no json contents got :", original.Error()}, ""),
		}
	}

	context, offsetOnLine, nextLineOffset := getContextAroundOffset(jsonContents, original.Offset, contextSize)

	var errorShowingLineBuffer = make([][]byte, 0, 9)
	errorShowingLineBuffer = append(errorShowingLineBuffer, replaceTabsWithSpace(context[:nextLineOffset]))
	errorShowingLineBuffer = append(errorShowingLineBuffer, []byte{newline})
	if offsetOnLine > 2 {
		errorShowingLineBuffer = append(errorShowingLineBuffer, bytes.Repeat([]byte{'-'}, int(offsetOnLine-2)))
	}
	var lineNumber = bytes.Count(jsonContents[:original.Offset], []byte{newline}) + 1
	errorShowingLineBuffer = append(errorShowingLineBuffer, strconv.AppendInt([]byte("^ on line "), int64(lineNumber), 10))
	errorShowingLineBuffer = append(errorShowingLineBuffer, strconv.AppendInt([]byte(" column "), offsetOnLine, 10))
	errorShowingLineBuffer = append(errorShowingLineBuffer, []byte(" got :"))
	errorShowingLineBuffer = append(errorShowingLineBuffer, []byte(original.Error()))
	errorShowingLineBuffer = append(errorShowingLineBuffer, replaceTabsWithSpace(context[nextLineOffset:]))
	var msg = bytes.Join(errorShowingLineBuffer, []byte{})

	return &SyntaxError{
		original: original,
		// replace the tabs so that the offset is correct
		msg: string(msg),
	}
}

func replaceTabsWithSpace(input []byte) []byte {
	return bytes.Replace(input, []byte{'\t'}, []byte{' '}, -1)
}

// for the provided contexts, offset it returns lines of context,
// the offset on the line of the error and the offset of the '\n'
// of the line on which ther error is.
// All the new offsets are for the returned context.
func getContextAroundOffset(contents []byte, offset, lines int64) (context []byte, onLineOffset, lineOffset int64) {
	// get the end of the line previous to the one the error is on
	var start = getOffsetXLineBack(contents[:offset], 1)
	onLineOffset = offset - start
	// calculate the correct start of the context
	start = getOffsetXLineBack(contents[:start], lines)
	// get the end of the line the error is on
	var end = getOffsetXLineForward(contents[offset:], 1) + offset
	// correct lineOffset for the context
	lineOffset = end - start
	end = getOffsetXLineForward(contents[min(len(contents), int(end+1)):], lines) + end
	context = contents[start:min(len(contents), int(end+1))]
	return
}

func getOffsetXLineBack(contents []byte, lines int64) (offset int64) {
	for offset = int64(len(contents)); offset != -1 && lines > 0; lines-- {
		offset = int64(bytes.LastIndexByte(contents[:offset], newline))
	}
	if offset == -1 {
		offset = 0
	}

	return
}

func getOffsetXLineForward(contents []byte, lines int64) (offset int64) {
	lines--
	for offset = int64(bytes.IndexByte(contents[offset:], newline)); offset != -1 && lines > 0; lines-- {
		offset += int64(bytes.IndexByte(contents[offset+1:], newline)) + 1
	}

	if offset == -1 {
		offset = int64(len(contents))
	}
	return
}

func min(l, r int) int {
	if l > r {
		return r
	}
	return l
}
