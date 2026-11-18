package engine

import (
	"bytes"
	"os"
	"testing"
)

func TestPDFSimpleObjectFlagsBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		kind       byte
		stream     bool
		flags      uint32
		known      bool
	}{
		{"scalar", "\n123\n", 'i', false, 0, true},
		{"catalog", "\n<< /Type /Catalog >>\n", 'd', false, 2, true},
		{"page", "\n<< /Type /Page >>\n", 'd', false, 2 | 1<<23, true},
		{"stream", "\n<< /Length 1 >>\nstream\nx\nendstream\n", 'd', true, 3, true},
		{"stream-with-name-bytes", "\n<< /Length 6 >>\nstream\n/Page\nendstream\n", 'd', true, 3, true},
		{"unproved-stream", "\n<< /Length 1 >>\nstream\nx\nendstream\n", 'd', false, 0, false},
		{"duplicate-stream-delimiter", "\n<< /Length 20 >>\nstream\nstream\nx\nendstream\n", 'd', true, 0, false},
		{"flate-filter", "\n<< /Filter /FlateDecode >>\n", 'd', false, 2 | 1<<16 | 1<<5, true},
		{"hex-filter-stream", "\n<< /Length 3 /Filter /ASCIIHexDecode >>\nstream\n41>\nendstream\n", 'd', true, 3 | 1<<16 | 1<<3, true},
		{"unknown-filter", "\n<< /Filter /MadeUp >>\n", 'd', false, 0, false},
		{"filter-array", "\n<< /Filter [/FlateDecode] >>\n", 'd', false, 2 | 1<<16 | 1<<5, true},
		{"duplicate-filter", "\n<< /Filter /FlateDecode /FlateDecode >>\n", 'd', false, 0, false},
		{"missing-filter", "\n<< /Filter >>\n", 'd', false, 0, false},
		{"indirect-flag", "\n<< /OpenAction 2 0 R >>\n", 'd', false, 0, false},
		{"escaped-name", "\n<< /Ty#70e /Catalog >>\n", 'd', false, 0, false},
		{"nested-dictionary", "\n<< /A << /Type /Catalog >> >>\n", 'd', false, 2, true},
		{"string-containing-name", "\n<< /A (/Page) >>\n", 'd', false, 2 | 1<<23, true},
		{"scalar-with-dictionary", "\n(<< /Page >>)\n", 's', false, 2 | 1<<23, true},
		{"wrapped-string-dictionary", "\n(abc << /Page >> xyz)\n", 's', false, 2 | 1<<23, true},
		{"string-dictionary-before-ordinary", "\n(<< /Page >>) << /Type /Catalog >>\n", 's', false, 2 | 1<<23, true},
		{"ordinary-after-plain-string", "\n(plain) << /Type /Page >>\n", 's', false, 2 | 1<<23, true},
		{"later-page-does-not-override-first-dictionary", "\n(<< /Catalog >>) << /Type /Page >>\n", 's', false, 2, true},
		{"first-of-two-dictionaries-in-string", "\n(<< /Page >> << /Catalog >>) << /Type /Page >>\n", 's', false, 2 | 1<<23, true},
		{"second-literal-has-first-dictionary", "\n(alpha)(<< /Page >>)\n", 's', false, 2 | 1<<23, true},
		{"first-of-two-literals-wins", "\n(<< /Page >>) (<< /Catalog >>)\n", 's', false, 2 | 1<<23, true},
		{"escaped-string-dictionary", "\n(<< \\/Page >>)\n", 's', false, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flags, known := pdfSimpleObjectFlags([]byte(tc.body), tc.kind, tc.stream)
			if flags != tc.flags || known != tc.known {
				t.Fatalf("flags=%#x known=%t; want %#x %t", flags, known, tc.flags, tc.known)
			}
		})
	}
}

func TestPDFComplexObjectInvalidatesWholeFlagSnapshot(t *testing.T) {
	data := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n2 0 obj\n<< /OpenAction 1 0 R >>\nendobj\n%%EOF\n")
	first := bytes.Index(data, []byte("obj")) + len("obj")
	second := bytes.Index(data, []byte("2 0 obj")) + len("2 0 obj")
	firstEnd := bytes.Index(data[first:], []byte("endobj")) + first
	secondEnd := bytes.Index(data[second:], []byte("endobj")) + second
	ps := &pdfScan{objects: 2, hookObjects: []pdfHookObject{
		{ref: pdfRef{object: 1}, offset: first, end: firstEnd, kind: 'd'},
		{ref: pdfRef{object: 2}, offset: second, end: secondEnd, kind: 'd'},
	}}
	state := ps.bytecodeState(data)
	if !state.ObjectsKnown || len(state.Objects) != 2 || state.Objects[0].FlagsKnown || state.Objects[1].FlagsKnown {
		t.Fatalf("unsafe PDF flag snapshot: %+v", state)
	}
}

func TestPDFMultipleLiteralObjectsReachFlagHooks(t *testing.T) {
	for _, tc := range []struct {
		name, body, hook, want string
	}{
		{"second-literal", "(alpha)(<< /Page >>)", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"separated-second-literal", "(alpha) (<< /Page >>)", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"first-dictionary-wins", "(<< /Page >>) (<< /Catalog >>)", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"catalog-first", "(<< /Catalog >> << /Page >>)", "pdf-flags-catalog.cbc", "UTM.PDF.FlagsCatalog"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hook, err := os.ReadFile("testdata/" + tc.hook)
			if err != nil {
				t.Fatal(err)
			}
			pdf := []byte("%PDF-1.7\n1 0 obj\n" + tc.body + "\nendobj\n%%EOF\n")
			state := pipelineState(t, map[string]string{tc.hook: string(hook)})
			if name, err := state.visit(pdf, 0); name != tc.want || err != nil {
				t.Fatalf("PDF hook: %q, %v; want %q", name, err, tc.want)
			}
		})
	}
}
