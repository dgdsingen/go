// Deprecated
// bufio.Scanner() 버전. 미완성.
package main

import (
	"bufio"
	"bytes"
	"io"
)

func splitRN() bufio.SplitFunc {
	// \r, \n, \r\n 모두 한번에 처리
	return func(bs []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(bs) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexAny(bs, "\r\n"); i >= 0 {
			// return before chunk of index
			return i + 1, bs[0:i], nil
		}
		// not found
		if atEOF {
			return len(bs), bs, nil
		}
		return 0, nil, nil
	}
}

func parseScanner(dst io.Writer, src io.Reader, prefix string) {
	line := &bytes.Buffer{}
	bprefix := []byte(prefix)

	scanner := bufio.NewScanner(src)
	scanner.Split(splitRN())
	for scanner.Scan() {
		dst.Write(concatBytes(line, bprefix, scanner.Bytes(), bsn))
	}
	if err := scanner.Err(); err != nil {
		// error handling
	}
}
