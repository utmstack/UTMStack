package platform

import (
	"bufio"
	"io"
	"strings"
)

// readJournalLines reads newline-delimited lines from r and calls handle for
// each one, with no line-length ceiling. bufio.Scanner caps a token at
// MaxScanTokenSize (64KB by default) and, on a line over that, stops reading
// the whole stream — every entry still behind it is lost too, not just the
// long one. bufio.Reader.ReadString has no such ceiling: it keeps growing its
// own buffer until it finds '\n', so no journal entry is ever dropped for its
// length (ING-F6). Returns nil on a clean EOF, or the first non-EOF read
// error.
func readJournalLines(r io.Reader, handle func(line string)) error {
	reader := bufio.NewReader(r)
	for {
		line, readErr := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line != "" {
			handle(line)
		}
		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return readErr
		}
	}
}
