package signatures

import (
	"fmt"
	"testing"
)

func TestTextTypeClassifiesLikeReference(t *testing.T) {
	cases := []struct {
		name string
		buf  string
		want FileType
	}{
		{"plain ASCII", "hello", TypeTextASCII},
		{"UTF-8 is ASCII here", "caf\xc3\xa9", TypeTextASCII},
		{"Latin-1 is ASCII here", "caf\xe9 cr\xe8me", TypeTextASCII},
		{"DEL makes it non-ASCII; unit 637f is high", "abc\x7f", TypeTextUTF16BE},
		{"UTF-16LE without mark reads as BE", "h\x00i\x00", TypeTextUTF16BE},
		{"UTF-16LE with mark", "\xff\xfeh\x00i\x00", TypeTextUTF16LE},
		{"UTF-16BE with mark", "\xfe\xff\x00h\x00i", TypeTextUTF16BE},
		{"unmarked non-text unit fails at once", "ab\x00\x01", TypeBinaryData},
		{"first unit is never checked", "\x00\x01ab", TypeTextUTF16BE},
		{"unit FFFE fails", "\xff\xfe\x00a\xfe\xff", TypeBinaryData},
		{"bytes from 0x80 are text: no UTF-16 check", "\xff\xfe\xfe\xff", TypeTextASCII},
		{"marked: bad units never reach half the bytes", "\xff\xfe\x01\x00\x02\x00", TypeTextUTF16LE},
		{"one non-text byte", "\x00", TypeBinaryData},
	}
	for _, c := range cases {
		if got := textType([]byte(c.buf)); got != c.want {
			t.Errorf("%s: textType(%q) = %s, want %s", c.name, c.buf, got, c.want)
		}
	}
}

// tarHeader builds a 512-byte header whose checksum field holds sum, written
// as the usual six octal digits, NUL and space.
func tarHeader(magic string, corrupt bool) []byte {
	h := make([]byte, 512)
	copy(h, "file.txt")
	copy(h[124:], "00000000012\x00")
	copy(h[257:], magic)
	sum := 0
	for i, b := range h {
		if i >= 148 && i < 156 {
			b = ' '
		}
		sum += int(b)
	}
	if corrupt {
		sum++
	}
	copy(h[148:], fmt.Sprintf("%06o\x00 ", sum))
	return h
}

func TestIsTarChecksHeaderSumAndMagic(t *testing.T) {
	if typ, ok := isTar(tarHeader("ustar  \x00", false)); !ok || typ != TypePosixTar {
		t.Fatalf("GNU magic: %s %v, want POSIX_TAR", typ, ok)
	}
	if typ, ok := isTar(tarHeader("ustar\x0000", false)); !ok || typ != TypeOldTar {
		t.Fatalf("POSIX magic: %s %v, want OLD_TAR (the reference compares the GNU spelling)", typ, ok)
	}
	if typ, ok := isTar(tarHeader("", false)); !ok || typ != TypeOldTar {
		t.Fatalf("no magic: %s %v, want OLD_TAR", typ, ok)
	}
	if _, ok := isTar(tarHeader("ustar  \x00", true)); ok {
		t.Fatal("a wrong checksum was accepted")
	}
	if _, ok := isTar(tarHeader("ustar  \x00", false)[:511]); ok {
		t.Fatal("a header shorter than 512 bytes was accepted")
	}
	blank := tarHeader("ustar  \x00", false)
	copy(blank[148:156], "        ")
	if _, ok := isTar(blank); ok {
		t.Fatal("an all-blank checksum field was accepted")
	}
	garbage := tarHeader("ustar  \x00", false)
	garbage[153] = 'x'
	if _, ok := isTar(garbage); ok {
		t.Fatal("a checksum field ending in a non-octal, non-space byte was accepted")
	}
}
