package engine

import (
	"strings"
	"testing"
)

func TestNormalizeTextMatchesReferenceVectors(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Hello,\tWORLD!\r\n", "hello, world! "},
		{"  A  ", " a "},
		{"a\x00b", "ab"},
		{"a \x01 b", "a b"},
		{"Caf\xc3\xa9 Cr\xe8me", "caf crme"},
		{"\xff\xfeH\x00i\x00 \x00A\x00\x41\x04", "hi aa"},
		{"\x00H\x00I", "hi"},
		{"@[\\]^_\x60{\x7c}~\x7f", "@[\\]^_\x60{\x7c}~\x7f"},
		{"A\x0bB\x0cC", "a b c"},
		{"\xef\xbb\xbf@ECHO OFF\r\nDEL C:\\*.* /Q\r\n", "@echo off del c:\\*.* /q "},
	}
	for _, c := range cases {
		if got := string(normalizeText([]byte(c.in), 0)); got != c.want {
			t.Errorf("normalizeText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeTextResetsWhitespaceAtEachChunk(t *testing.T) {
	const chunk = 1000
	in := strings.Repeat("x", chunk-1) + "  y"
	want := strings.Repeat("x", chunk-1) + "  y"
	if got := string(normalizeText([]byte(in), chunk)); got != want {
		t.Fatalf("a whitespace run across a chunk boundary: got %q", got[chunk-3:])
	}
	if got := string(normalizeText([]byte(in), 0)); got != strings.Repeat("x", chunk-1)+" y" {
		t.Fatalf("without chunks: got %q", got[chunk-3:])
	}
}
