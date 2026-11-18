package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// pdfWithHiddenObject places object 2 inside object 1's binary stream body
// (or at top level for the control) and points the cross-reference table at
// object 2's own offset, as xref-driven readers resolve it.
func pdfWithHiddenObject(t *testing.T, hidden bool, prefix, header string) []byte {
	t.Helper()
	compressed := zlibFixture(t, []byte(testMarker))
	obj2 := header + fmt.Sprintf("\n<< /Length %d /Filter /FlateDecode >>\nstream\n", len(compressed)) + string(compressed) + "\nendstream\nendobj\n"
	var b bytes.Buffer
	b.WriteString("%PDF-1.7\n")
	off1 := b.Len()
	var off2 int
	if hidden {
		inner := prefix + obj2
		fmt.Fprintf(&b, "1 0 obj\n<< /Length %d >>\nstream\n", len(inner))
		off2 = b.Len() + len(prefix)
		b.WriteString(inner)
		b.WriteString("\nendstream\nendobj\n")
	} else {
		b.WriteString("1 0 obj\n<< /Length 1 >>\nstream\n\x01\nendstream\nendobj\n")
		off2 = b.Len()
		b.WriteString(obj2)
	}
	off3 := b.Len()
	b.WriteString("3 0 obj\n<< /Type /Catalog /OpenAction 2 0 R >>\nendobj\n")
	xref := b.Len()
	b.WriteString("xref\n0 4\n0000000000 65535 f \n")
	for _, off := range []int{off1, off2, off3} {
		fmt.Fprintf(&b, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size 4 /Root 3 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	return b.Bytes()
}

// An indirect object placed inside another object's binary stream body, and
// referenced by the cross-reference table at its own offset, is a real object
// for xref-driven and reconstructing readers. Its filtered payload is decoded.
func TestPDFObjectHiddenInsideStreamIsDecoded(t *testing.T) {
	e := nativeTestEngine(t)
	if clean, sig, err := e.Scan(context.Background(), pdfWithHiddenObject(t, false, "", "2 0 obj")); clean || sig != "Native.Test" || err != nil {
		t.Fatalf("control not detected: clean=%v sig=%q err=%v", clean, sig, err)
	}
	for name, tc := range map[string]struct{ prefix, header string }{
		"binary prefix":      {"\x01", "2 0 obj"},
		"digit prefix":       {"\x017", "2 0 obj"},
		"line breaks":        {"\x01\r\n", "2\r\n0\r\nobj"},
		"comment separators": {"\x01\n", "2 0 % generation\nobj"},
		"zero-padded tokens": {"\x01", "0002 0000000 obj"},
	} {
		t.Run(name, func(t *testing.T) {
			data := pdfWithHiddenObject(t, true, tc.prefix, tc.header)
			if clean, sig, err := e.Scan(context.Background(), data); clean || sig != "Native.Test" || err != nil {
				t.Fatalf("clean=%v sig=%q err=%v; want the hidden object decoded", clean, sig, err)
			}
		})
	}
}

// Objects can also hide in a comment or in another object's string, where the
// sequential pass sees only undecoded text.
func TestPDFObjectHiddenInCommentOrStringIsDecoded(t *testing.T) {
	e := nativeTestEngine(t)
	hidden := fmt.Sprintf("2 0 obj <%x> endobj", testMarker)
	for name, data := range map[string]string{
		"comment": "%PDF-1.7\n% " + hidden + "\n1 0 obj\n(ordinary)\nendobj\n%%EOF\n",
		"string":  "%PDF-1.7\n1 0 obj\n(" + hidden + ")\nendobj\n%%EOF\n",
	} {
		t.Run(name, func(t *testing.T) {
			if clean, sig, err := e.Scan(context.Background(), []byte(data)); clean || sig != "Native.Test" || err != nil {
				t.Fatalf("clean=%v sig=%q err=%v; want the hidden string decoded", clean, sig, err)
			}
		})
	}
}

// Embedded headers are parsed where they sit, so one parse can run on to the
// end of the document. Many headers in front of a long unterminated comment,
// or of a stream whose end is far away, must not make the pass quadratic:
// together these parses examine at most the document's size.
func TestPDFEmbeddedObjectParsingIsLinear(t *testing.T) {
	e := nativeTestEngine(t)
	// Without a bound the cost grows with headers x tail: tens of seconds here.
	const headers, streams, tail = 6000, 2000, 4 << 20
	var comment bytes.Buffer
	comment.WriteString("%PDF-1.7\n1 0 obj\n(ordinary)\nendobj\n%")
	for i := 0; i < headers; i++ {
		comment.WriteString(" 2 0 obj %")
	}
	comment.Write(bytes.Repeat([]byte("A"), tail))
	comment.WriteString("\n%%EOF\n")
	var stream bytes.Buffer
	// Each "endstreamX" is a keyword match that is not a stream end.
	body := strings.Repeat("2 0 obj << >> stream\n", headers) + strings.Repeat("endstreamX", tail/10)
	fmt.Fprintf(&stream, "%%PDF-1.7\n1 0 obj\n<< /Length %d >>\nstream\n\x01%s\nendstream\nendobj\n%%%%EOF\n", len(body)+1, body)
	// A declared length that ends inside a long white-space run makes the
	// parser skip that run before it falls back to searching for endstream.
	// Every stream whose length points there, embedded or not, repeats it.
	padded := func(embedded bool) []byte {
		const object = "2 0 obj << /Length 9 0 R >> stream\nendstream endobj\n"
		var b bytes.Buffer
		b.WriteString("%PDF-1.7\n9 0 obj\n")
		lengthAt := b.Len()
		b.WriteString("0000000000 endobj\n")
		if embedded {
			b.WriteString("1 0 obj\n<< /Length 0000000000 >>\nstream\n\x01")
		}
		first := b.Len() + strings.Index(object, "stream\n") + len("stream\n")
		b.WriteString(strings.Repeat(object, streams))
		padding := b.Len()
		b.WriteString(strings.Repeat(" ", tail) + "X")
		if embedded {
			b.WriteString("\nendstream\nendobj")
		}
		b.WriteString("\n%%EOF\n")
		data := b.Bytes()
		copy(data[lengthAt:], fmt.Sprintf("%010d", padding-first))
		if embedded {
			at := bytes.Index(data, []byte("/Length 0000000000")) + len("/Length ")
			copy(data[at:], fmt.Sprintf("%010d", padding+tail+1-(at+len("0000000000 >>\nstream\n"))))
		}
		return data
	}
	// The padding cases' lengths do not end at endstream, so the end found
	// at endstream is used, as the reference calculates it: they are clean.
	for name, tc := range map[string]struct {
		data       []byte
		incomplete bool
	}{
		"unterminated comment": {comment.Bytes(), true},
		"distant endstream":    {stream.Bytes(), true},
		"embedded padding":     {padded(true), false},
		"sequential padding":   {padded(false), false},
	} {
		t.Run(name, func(t *testing.T) {
			start := time.Now()
			clean, sig, err := e.Scan(context.Background(), tc.data)
			if elapsed := time.Since(start); clean == tc.incomplete || sig != "" || (err != nil) != tc.incomplete || elapsed > 10*time.Second {
				t.Fatalf("clean=%v sig=%q err=%v elapsed=%v; want a bounded pass, incomplete %v", clean, sig, err, elapsed, tc.incomplete)
			}
		})
	}
}

func TestPDFEmbeddedObjectHeadersNeverClean(t *testing.T) {
	e := nativeTestEngine(t)
	// A header that is not followed by a parseable object cannot be resolved
	// the way a reader following a cross-reference offset would.
	if clean, sig, err := e.Scan(context.Background(), pdfFixture("", []byte("\x012 0 obj\n<< /A )\n"))); clean || sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("clean=%v sig=%q err=%v; want malformed", clean, sig, err)
	}
	// Words that only contain "obj" are not object headers.
	benign := pdfFixture("", []byte("var obj = 1; x.obj; (1 obj) /obj 2 obj3"))
	if clean, sig, err := e.Scan(context.Background(), benign); !clean || sig != "" || err != nil {
		t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
	}
	// A benign hidden object is parsed without making the document incomplete.
	if clean, sig, err := e.Scan(context.Background(), pdfFixture("", []byte("\x012 0 obj\n(hello)\nendobj"))); !clean || sig != "" || err != nil {
		t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
	}
}
