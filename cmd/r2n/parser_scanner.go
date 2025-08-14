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
	scanner := bufio.NewScanner(src)
	scanner.Split(splitRN())

	bprefix := []byte(prefix)
	bn := []byte{'\n'}

	for scanner.Scan() {
		line := scanner.Bytes()
		// add logics: remove \r, ...
		dst.Write(bprefix)
		dst.Write(line)
		dst.Write(bn)
	}
	if err := scanner.Err(); err != nil {
		// error handling
	}
}
