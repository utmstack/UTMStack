package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"unicode/utf16"
)

func uuTestEncode(data []byte) []byte {
	out := []byte("begin 644 attachment.bin\n")
	for len(data) != 0 {
		n := min(45, len(data))
		line := data[:n]
		out = append(out, byte(n)+' ')
		for len(line) != 0 {
			var a, b, c byte
			a = line[0]
			if len(line) > 1 {
				b = line[1]
			}
			if len(line) > 2 {
				c = line[2]
			}
			out = append(out, ' '+a>>2, ' '+((a&3)<<4|b>>4), ' '+((b&15)<<2|c>>6), ' '+(c&63))
			line = line[min(3, len(line)):]
		}
		out = append(out, '\n')
		data = data[n:]
	}
	return append(out, []byte("`\nend\n")...)
}

func TestUUEncodedContainers(t *testing.T) {
	e := nativeTestEngine(t)
	// A conventional short historical-uuencode vector decodes to Cat.
	known := []byte("begin 644 cat.txt\n#0V%T\n`\nend\n")
	decoded, _, err := (&scanState{ctx: context.Background()}).uuHistoricalBlock(known, len("begin 644 cat.txt\n"), 32)
	if string(decoded) != "Cat" || err != nil {
		t.Fatalf("published example decoded as %q: %v", decoded, err)
	}
	if clean, sig, err := e.Scan(context.Background(), known); !clean || sig != "" || err != nil {
		t.Fatalf("known sample: %v %q %v", clean, sig, err)
	}
	inner := gzipFixture(t, []byte(testMarker))
	historical := uuTestEncode(inner)
	base64Envelope := []byte("begin-base64 644 attachment.bin\n" + base64.StdEncoding.EncodeToString(inner) + "\n====\n")
	mail := append([]byte("MIME-Version: 1.0\r\nContent-Type: text/plain\r\n\r\nAttached below:\r\n"), historical...)
	for name, data := range map[string][]byte{"historical": historical, "base64": base64Envelope, "mail": mail} {
		t.Run(name, func(t *testing.T) {
			if bytes.Contains(data, []byte(testMarker)) {
				t.Fatal("fixture exposes marker without decoding")
			}
			clean, sig, err := e.Scan(context.Background(), data)
			if clean || sig != "Native.Test" || err != nil {
				t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
}

func TestUUEncodedIncomplete(t *testing.T) {
	e := nativeTestEngine(t)
	for name, data := range map[string][]byte{
		"invalid mode":   []byte("begin 999 file\n#0V%T\n`\nend\n"),
		"missing end":    []byte("begin 644 file\n#0V%T\n`\n"),
		"bad alphabet":   []byte("begin 644 file\n#0~%T\n`\nend\n"),
		"missing base64": []byte("begin-base64 644 file\nQ2F0\n"),
	} {
		t.Run(name, func(t *testing.T) {
			clean, sig, err := e.Scan(context.Background(), data)
			if clean || sig != "" || !errors.Is(err, ErrMalformed) {
				t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("ordinary text")); !clean || sig != "" || err != nil {
		t.Fatal(clean, sig, err)
	}
}

// Ordinary text whose lines start with "begin " (PowerShell begin blocks, SQL
// transactions, prose) is not a uuencode envelope and must scan clean, not
// incomplete. AMSI fail-closed would otherwise block these scripts.
func TestBeginLinesAreNotMalformedUUEncode(t *testing.T) {
	e := nativeTestEngine(t)
	ps := "param([int]$x)\nbegin { $n = 0 }\nprocess { $n++ }\nend { $n }\n"
	utf16le := func(s string) []byte {
		var out []byte
		for _, u := range utf16.Encode([]rune(s)) {
			out = append(out, byte(u), byte(u>>8))
		}
		return out
	}
	for name, input := range map[string][]byte{
		"powershell-ascii":   []byte(ps),
		"powershell-utf16le": utf16le(ps),
		"sql":                []byte("begin transaction;\nupdate t set a = 1;\ncommit;\n"),
		"prose":              []byte("Notes\nbegin with the setup steps below\n"),
		"base64-prose":       []byte("begin-base64 is the header keyword\n"),
		"short-number":       []byte("begin 12 items\n"),
	} {
		t.Run(name, func(t *testing.T) {
			clean, sig, err := e.Scan(context.Background(), input)
			if err != nil || !clean || sig != "" {
				t.Fatalf("benign text: clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
}

// A recognized header (a numeric mode then a name) still commits to decoding:
// a broken body or a non-octal mode stays incomplete, and a valid envelope
// after ordinary begin lines is still decoded.
func TestUUEncodeHeaderRecognition(t *testing.T) {
	e := nativeTestEngine(t)
	for name, data := range map[string][]byte{
		"broken body":     []byte("begin 644 file\nthis is not uuencoded\n"),
		"empty name":      []byte("begin 644 \n#0V%T\n`\nend\n"),
		"four digit mode": []byte("begin 0644 file\n#0V%T\n"),
		"base64 no name":  []byte("begin-base64 644 \nQ2F0\n====\n"),
	} {
		t.Run(name, func(t *testing.T) {
			clean, sig, err := e.Scan(context.Background(), data)
			if clean || sig != "" || !errors.Is(err, ErrMalformed) {
				t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
	hidden := append([]byte("begin {\n  $x = 1\n}\n"), uuTestEncode(gzipFixture(t, []byte(testMarker)))...)
	if clean, sig, err := e.Scan(context.Background(), hidden); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("envelope after a begin block: clean=%v sig=%q err=%v", clean, sig, err)
	}
}
