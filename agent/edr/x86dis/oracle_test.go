package x86dis

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

// TestOracle compares Decode with the reference decoder over a corpus. It runs
// only when X86DIS_ORACLE names a file of "input<TAB>oracle line" pairs, where
// the oracle line is the reference's answer for the hex input. Every
// difference found this way becomes a golden case before it is fixed.
func TestOracle(t *testing.T) {
	path := os.Getenv("X86DIS_ORACLE")
	if path == "" {
		t.Skip("X86DIS_ORACLE not set")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	total, diffs := 0, 0
	for sc.Scan() {
		input, want, found := strings.Cut(sc.Text(), "\t")
		if !found {
			continue
		}
		total++
		if diffs >= 40 {
			if !quietCheck(input, want) {
				diffs++
			}
			continue
		}
		if !checkAgainst(t, input, want) {
			diffs++
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d inputs, %d differences", total, diffs)
	if diffs > 0 {
		t.Fail()
	}
}

func quietCheck(input, want string) bool {
	buf := make([]byte, len(input)/2)
	for i := range buf {
		buf[i] = unhex(input[2*i])<<4 | unhex(input[2*i+1])
	}
	return answer(Decode(buf)) == strings.TrimSpace(want)
}

func unhex(c byte) byte {
	switch {
	case c <= '9':
		return c - '0'
	case c >= 'a':
		return c - 'a' + 10
	}
	return c - 'A' + 10
}
