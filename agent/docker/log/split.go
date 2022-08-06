package log

import (
	"bufio"
	"bytes"
)

// splitWithLF is a split function for a bufio.Scanner that splits data with LF.
// It implements bufio.SplitFunc.
func splitWithLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i != -1 {
		// If data contains LF, return data until LF
		return i + 1, data[:i+1], nil
	}

	if atEOF {
		// data does not contain LF, and reached EOF.
		return len(data), data, nil
	}

	// Request more data until we reach LF.
	return 0, nil, nil
}

var _ bufio.SplitFunc = splitWithLF
