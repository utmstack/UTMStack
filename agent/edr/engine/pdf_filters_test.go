package engine

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"testing"
)

// Optional extraction still consumes real scan resources. A decoder reaching
// the configured expansion cap cannot be downgraded to a complete clean scan
// merely because the reference scanner skips that filter.
func TestPDFOptionalExtractionPreservesResourceLimits(t *testing.T) {
	var encoded []byte
	for i := 0; i < 512; i++ {
		encoded = append(encoded, 129, 'A') // 128 decoded bytes per pair
	}
	encoded = append(encoded, 128)
	data := pdfFixture("/Filter /RunLengthDecode", encoded)
	rules := map[string]string{"never.ndb": "Never:0:*:" + hx("never-present-marker") + "\n"}
	for _, budget := range []string{"file", "scan"} {
		t.Run(budget, func(t *testing.T) {
			s := pipelineState(t, rules)
			if budget == "file" {
				s.maxFile = int64(len(data) + 1024)
			} else {
				s.remaining = int64(len(data) + 1024)
			}
			if sig, err := s.visit(data, 0); sig != "" || !errors.Is(err, ErrLimit) {
				t.Fatalf("scan: %q, %v; want resource limit", sig, err)
			}
		})
	}
}

// A literal-only MSB LZW stream is valid even though it does not compress well.
// Writing the codes directly keeps the boundary fixture independent of the
// production decoder, and crosses the 9-to-10-bit EarlyChange transition.
func pdfLZWLiteralFixture(data []byte, earlyChange int) []byte {
	var out []byte
	var pending uint32
	var pendingBits uint
	width := uint(9)
	emit := func(code uint16) {
		pending = pending<<width | uint32(code)
		pendingBits += width
		for pendingBits >= 8 {
			pendingBits -= 8
			out = append(out, byte(pending>>pendingBits))
			pending &= (1 << pendingBits) - 1
		}
	}
	emit(256)
	hi := 257
	for _, b := range data {
		emit(uint16(b))
		hi++
		if width < 12 && hi+earlyChange == 1<<width {
			width++
		}
	}
	emit(257)
	if pendingBits != 0 {
		out = append(out, byte(pending<<(8-pendingBits)))
	}
	return out
}

// The reference decodes LZW streams with early change 1 whatever
// /EarlyChange says (checked with LZW then Flate chains): a stream written
// with early change 0 is not decoded, with or without the parameter.
func TestPDFLZWAsReference(t *testing.T) {
	e := nativeTestEngine(t)
	raw := append(bytes.Repeat([]byte{'X'}, 300), []byte(testMarker)...)
	for name, tc := range map[string]struct {
		filter string
		early  int
		found  bool
	}{
		"default":         {"/Filter /LZWDecode", 1, true},
		"early change 1":  {"/Filter /LZWDecode /DecodeParms << /EarlyChange 1 >>", 1, true},
		"early change 0":  {"/Filter /LZWDecode /DecodeParms << /EarlyChange 0 >>", 0, false},
		"0, no parameter": {"/Filter /LZWDecode", 0, false},
	} {
		encoded := pdfLZWLiteralFixture(raw, tc.early)
		if bytes.Contains(encoded, []byte(testMarker)) {
			t.Fatal("fixture exposes marker without decoding")
		}
		_, sig, err := e.Scan(context.Background(), pdfFixture(tc.filter, encoded))
		if tc.found != (sig == "Native.Test") || err != nil {
			t.Errorf("%s: (%q, %v), want found=%v", name, sig, err, tc.found)
		}
	}
}

// The reference does not undo predictors (checked with PNG rows that split
// the marker and TIFF differences): a stream is its decoded bytes as they
// are, and its decode parameters are not read.
func TestPDFPredictorsAreNotUndone(t *testing.T) {
	e := nativeTestEngine(t)
	plain := append(append(bytes.Repeat([]byte{'x'}, 10), testMarker...), bytes.Repeat([]byte{'y'}, 25)...)
	var png []byte
	for i := 0; i < len(plain); i += 16 {
		png = append(append(png, 0), plain[i:min(i+16, len(plain))]...)
	}
	tiff := append([]byte(nil), plain...)
	for i := len(tiff) - 1; i > 0; i-- {
		tiff[i] -= tiff[i-1]
	}
	for name, tc := range map[string]struct {
		filter string
		data   []byte
		found  bool
	}{
		"png rows split the marker": {"/Filter /FlateDecode /DecodeParms << /Predictor 12 /Columns 16 >>", zlibFixture(t, png), false},
		"tiff differences":          {"/Filter /FlateDecode /DecodeParms << /Predictor 2 /Columns 64 >>", zlibFixture(t, tiff), false},
		"rows do not fit":           {"/Filter /FlateDecode /DecodeParms << /Predictor 15 /Columns 7 /Colors 3 /BitsPerComponent 4 >>", zlibFixture(t, plain), true},
		"unknown parameters":        {"/Filter /FlateDecode /DecodeParms << /Predictor 99 /Foo 1 >>", zlibFixture(t, plain), true},
		"lzw":                       {"/Filter /LZWDecode /DecodeParms << /Predictor 2 /Columns 64 >>", pdfLZWLiteralFixture(plain, 1), true},
	} {
		_, sig, err := e.Scan(context.Background(), pdfFixture(tc.filter, tc.data))
		if tc.found != (sig == "Native.Test") || err != nil {
			t.Errorf("%s: (%q, %v), want found=%v", name, sig, err, tc.found)
		}
	}
}

// The reference extracts a stream only when it has no filter or one of
// Flate, ASCIIHex, ASCII85, DCT and JBIG2; streams with only LZW,
// RunLength, JPX, CCITT, Crypt or unknown filters are not extracted, even an
// attachment's or a script's. This engine still analyzes them, but what it
// cannot decode there does not make the file incomplete. Bytes after the end
// of Flate data are accepted, as the reference reports them as remaining.
func TestPDFStreamsTheReferenceExtracts(t *testing.T) {
	e := nativeTestEngine(t)
	broken := []byte{0x80, 0x01, 0x02}
	for _, filter := range []string{"/LZWDecode", "/RunLengthDecode", "[/LZWDecode /RunLengthDecode]", "/JPXDecode", "/Crypt"} {
		if clean, sig, err := e.Scan(context.Background(), pdfFixture("/Filter "+filter, broken)); !clean || sig != "" || err != nil {
			t.Errorf("%s: (%v, %q, %v), want clean", filter, clean, sig, err)
		}
	}
	for _, filter := range []string{"/FlateDecode", "[/LZWDecode /FlateDecode]"} {
		if clean, sig, err := e.Scan(context.Background(), pdfFixture("/Filter "+filter, broken)); clean || sig != "" || err == nil {
			t.Errorf("%s: (%v, %q, %v), want incomplete", filter, clean, sig, err)
		}
	}
	trailing := append(zlibFixture(t, []byte(testMarker)), []byte("\r\n")...)
	if _, sig, err := e.Scan(context.Background(), pdfFixture("/Filter /FlateDecode", trailing)); sig != "Native.Test" || err != nil {
		t.Errorf("bytes after the Flate data: (%q, %v)", sig, err)
	}
}

// A corrupt zlib trailer is reported after the decompressor has already
// produced the complete stream. Those bytes still need matching; a damaged
// stream without a hit remains incomplete rather than becoming clean.
func TestPDFMatchesOutputBeforeFlateChecksumFailure(t *testing.T) {
	e := nativeTestEngine(t)
	for _, tc := range []struct {
		name  string
		plain []byte
		want  string
	}{
		{"marker", append(bytes.Repeat([]byte("X"), 32), testMarker...), "Native.Test"},
		{"no marker", bytes.Repeat([]byte("Y"), 64), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded := zlibFixture(t, tc.plain)
			encoded[len(encoded)-1] ^= 0xff
			if bytes.Contains(encoded, []byte(testMarker)) {
				t.Fatal("fixture exposes marker without decoding")
			}
			clean, sig, err := e.Scan(context.Background(), pdfFixture("/Filter /FlateDecode", encoded))
			if tc.want != "" {
				if sig != tc.want || err != nil || clean {
					t.Fatalf("damaged stream: (%v, %q, %v), want detection", clean, sig, err)
				}
			} else if clean || sig != "" || !errors.Is(err, ErrMalformed) {
				t.Fatalf("damaged stream: (%v, %q, %v), want incomplete", clean, sig, err)
			}
		})
	}
}

func TestPDFRecoversFlateAfterMalformedStart(t *testing.T) {
	e := nativeTestEngine(t)
	encoded := zlibFixture(t, []byte(testMarker))
	for name, prefix := range map[string][]byte{
		"stray carriage return": {'\r'},
		"bad first line":        []byte("garbage header\r\n"),
	} {
		t.Run(name, func(t *testing.T) {
			stream := append(append([]byte(nil), prefix...), encoded...)
			if bytes.Contains(stream, []byte(testMarker)) {
				t.Fatal("fixture exposes marker without decoding")
			}
			if clean, sig, err := e.Scan(context.Background(), pdfFixture("/Filter /FlateDecode", stream)); clean || sig != "Native.Test" || err != nil {
				t.Fatalf("recovered stream: (%v, %q, %v)", clean, sig, err)
			}
		})
	}
}

func TestPDFDamagedFlateContinuesLaterFilter(t *testing.T) {
	e := nativeTestEngine(t)
	intermediate := []byte(hex.EncodeToString([]byte(testMarker)) + ">")
	encoded := zlibFixture(t, intermediate)
	encoded[len(encoded)-1] ^= 0xff
	if bytes.Contains(encoded, []byte(testMarker)) {
		t.Fatal("fixture exposes marker without both filters")
	}
	clean, sig, err := e.Scan(context.Background(), pdfFixture("/Filter [/FlateDecode /ASCIIHexDecode]", encoded))
	if clean || sig != "Native.Test" || err != nil {
		t.Fatalf("chained damaged stream: (%v, %q, %v)", clean, sig, err)
	}
}

func TestPDFDoesNotScanDiscardedASCIIHexOutput(t *testing.T) {
	e := nativeTestEngine(t)
	encoded := []byte(hex.EncodeToString([]byte(testMarker)) + "GGGG")
	if bytes.Contains(encoded, []byte(testMarker)) {
		t.Fatal("fixture exposes marker without decoding")
	}
	clean, sig, err := e.Scan(context.Background(), pdfFixture("/Filter /ASCIIHexDecode", encoded))
	if clean || sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("failed non-Flate decoder: (%v, %q, %v)", clean, sig, err)
	}
}
