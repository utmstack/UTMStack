package engine

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func rtfEscapedText(value string, unicode bool) string {
	var b strings.Builder
	for _, c := range []byte(value) {
		if unicode {
			fmt.Fprintf(&b, `\u%d `, c)
		} else {
			fmt.Fprintf(&b, `\'%02x`, c)
		}
	}
	return b.String()
}

func TestRTFTextEncodedDetection(t *testing.T) {
	for _, unicode := range []bool{false, true} {
		for _, positive := range []bool{false, true} {
			text, want := "A harmless quarterly invoice.", ""
			if positive {
				text, want = formatCoverageMarker, "RTF.Text"
			}
			data := []byte(`{\rtf1\ansi\uc0 ` + rtfEscapedText(text, unicode) + `}`)
			if bytes.Contains(data, []byte(text)) {
				t.Fatal("fixture leaks the unencoded text")
			}
			expectScan(t, pipelineState(t, map[string]string{"text.ndb": "RTF.Text:0:*:" + hx(formatCoverageMarker) + "\n"}), data, want, "")
		}
	}
}

func TestRTFTextControlSemantics(t *testing.T) {
	for _, c := range []struct{ name, rtf, want string }{
		{"formatting", `{\rtf1\ansi A{\b B}\i0 C}`, "ABC"},
		{"escapes", `{\rtf1 a\\b\{c\}\'21}`, `a\b{c}!`},
		{"hex case", `{\rtf1 \'4a\'4B}`, "JK"},
		{"line breaks", "{\\rtf1 A\r\nB\\par C\\line D\\tab E\\cell F\\row G}", "AB\r\nC\r\nD\tE\tF\r\nG"},
		{"escaped line breaks", "{\\rtf1 A\\\r\nB\\\nC}", "A\r\nB\r\nC"},
		{"delimiter", `{\rtf1\b  AB}`, " AB"},
		{"unicode default fallback", `{\rtf1 \u65?\u66\'3f\u67\{D}`, "ABCD"},
		{"unicode control fallback", `{\rtf1\uc2\u65\par ?B}`, "AB"},
		{"unicode bin fallback", `{\rtf1\uc1\u65\bin3 \{}B}`, "AB"},
		{"fallback scopes", `{\rtf1\uc1\u65?{\uc0\u66 }\u67?}`, "ABC"},
		{"fallback stops before group", `{\rtf1\u65{\b B}C}`, "ABC"},
		{"fallback ignores lexical newline", "{\\rtf1\\u65\r\n?B}", "AB"},
		{"unicode signed surrogate", `{\rtf1\uc1\u-10179?\u-8704?}`, "😀"},
		{"unicode signed boundaries", `{\rtf1\uc0\u32767\u-32768\u-1\u0}`, "\u7fff\u8000\uffff\x00"},
		{"windows1252", `{\rtf1\ansi\ansicpg1252 Caf\'e9 \'80}`, "Café €"},
		{"font1252", `{\rtf1\ansi\deff0{\fonttbl{\f0\fcharset0 Arial;}{\f1\fcharset1\cpg1252 Times;}}\f1 Caf\'e9}`, "Café"},
		{"unicode independent of legacy font", `{\rtf1\ansicpg932\uc0\u65\u66}`, "AB"},
		{"unicode symbols", `{\rtf1 A\~B\-C\_D\emdash E\endash F\lquote G\rquote H\ldblquote I\rdblquote J\bullet K}`, "A\u00a0B\u00adC\u2011D—E–F‘G’H“I”J•K"},
		{"unicode joining controls", `{\rtf1 A\zwj B\zwnj C\zwbo D\zwnbo E\ltrmark F\rtlmark G}`, "A\u200dB\u200cC\u200bD\u2060E\u200eF\u200fG"},
		{"starred destination", `{\rtf1 A{\*\unknown \'58\'59}B}`, "AB"},
		{"skipped nested binary", `{\rtf1 A{\pict\bin4 {}\\}B}`, "AB"},
		{"field result", `{\rtf1 {\field{\*\fldinst SECRET}{\fldrslt actual text}}}`, "actual text"},
		{"trailing whitespace", "{\\rtf1 okay}\r\n\t ", "okay"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := decodeRTFText(context.Background(), []byte(c.rtf), 4096)
			if err != nil || string(got) != c.want {
				t.Fatalf("text=%q err=%v, want %q", got, err, c.want)
			}
		})
	}
}

func TestRTFTextNoMetadataOrSyntheticFileMatches(t *testing.T) {
	marker := rtfEscapedText(formatCoverageMarker, false)
	for _, dest := range []string{"fonttbl", "stylesheet", "colortbl", "info", "listtable", "fldinst", "pict", "objclass", "generator", `*\arbitrary`} {
		t.Run(dest, func(t *testing.T) {
			data := []byte(`{\rtf1\ansi{\` + dest + ` ` + marker + `}ordinary body text}`)
			expectScan(t, pipelineState(t, map[string]string{"text.ndb": "RTF.Text:0:*:" + hx(formatCoverageMarker) + "\n"}), data, "", "")
		})
	}
	// Logical text must preserve separators and literal escape spelling. It
	// must not concatenate through paragraph/tab control words to fake a match.
	half := len(formatCoverageMarker) / 2
	for _, separator := range []string{`\par `, `\line `, `\tab `, `\cell `, `\\`, `\{`, "\\\n", `\zwj `} {
		data := []byte(`{\rtf1 ` + rtfEscapedText(formatCoverageMarker[:half], false) + separator + rtfEscapedText(formatCoverageMarker[half:], false) + `}`)
		expectScan(t, pipelineState(t, map[string]string{"text.ndb": "RTF.Text:0:*:" + hx(formatCoverageMarker) + "\n"}), data, "", "")
	}
	for _, separate := range []string{"header", "footer", "footnote", "annotation", "shptxt"} {
		data := []byte(`{\rtf1 ` + rtfEscapedText(formatCoverageMarker[:half], false) + `{\` + separate + ` ` + rtfEscapedText(formatCoverageMarker[half:], false) + `}}`)
		expectScan(t, pipelineState(t, map[string]string{"text.ndb": "RTF.Text:0:*:" + hx(formatCoverageMarker) + "\n"}), data, "", "")
		data = []byte(`{\rtf1 {\` + separate + ` ` + marker + `}harmless body}`)
		expectScan(t, pipelineState(t, map[string]string{"text.ndb": "RTF.Text:0:*:" + hx(formatCoverageMarker) + "\n"}), data, "RTF.Text", "")
	}
	doc := []byte(`{\rtf1 ` + marker + `}`)
	sum := md5.Sum([]byte(formatCoverageMarker))
	for name, rules := range map[string]map[string]string{
		"hash": {"text.hdb": fmt.Sprintf("%x:%d:RTF.NotAFile\n", sum, len(formatCoverageMarker))},
		"EOF":  {"text.ndb": "RTF.NotAnEOF:0:EOF-" + fmt.Sprint(len(formatCoverageMarker)) + ":" + hx(formatCoverageMarker) + "\n"},
	} {
		t.Run(name, func(t *testing.T) { expectScan(t, pipelineState(t, rules), doc, "", "") })
	}
	// A text representation belongs to the RTF object. Its container remains
	// the surrounding ZIP, rather than a fabricated plain-text child file.
	rules := map[string]string{"text.ldb": "RTF.InZip;Engine:81-255,Target:0,Container:CL_TYPE_ZIP;0;" + hx(formatCoverageMarker) + "\n"}
	expectScan(t, pipelineState(t, rules), doc, "", "")
	expectScan(t, pipelineState(t, rules), zipFixture(t, doc), "RTF.InZip", "")
}

func TestRTFTextMalformedAndUnsupported(t *testing.T) {
	for _, rtf := range []string{
		`{\rtf1`, `{\rtext harmless}`, `{\rtf11 text}`, `{\rtf1 text`, `{\rtf1 text}}`, `{\rtf1 text}junk`,
		`{\rtf1 \'x1}`, `{\rtf1 \'1}`, `{\rtf1 \}`, `{\rtf1 \bin}`, `{\rtf1 \bin-1 x}`, `{\rtf1 \bin9 x}`,
		`{\rtf1 \uc-1 x}`, `{\rtf1 \uc32768 x}`, `{\rtf1 \uc x}`, `{\rtf1 \u x}`, `{\rtf1 \u32768?}`,
		`{\rtf1 \u-32769?}`, `{\rtf1\uc0\u-10179x}`, `{\rtf1\uc0\u-8704}`, `{\rtf1 \b2147483648 x}`,
		`{\rtf1 \deff-1 text}`, `{\rtf1\ansicpg0 text}`, `{\rtf1{\fonttbl{\f0\fcharset-1 Arial;}}text}`, `{\rtf1 \'81}`,
		`{\rtf1 \` + strings.Repeat("a", 33) + ` text}`,
	} {
		if _, err := decodeRTFText(context.Background(), []byte(rtf), 4096); !errors.Is(err, ErrMalformed) {
			t.Fatalf("%q: %v, want malformed", rtf, err)
		}
	}
	for _, rtf := range []string{
		`{\rtf1\ansicpg932 ASCII}`, `{\rtf1\mac ASCII}`, `{\rtf1{\fonttbl{\f0\fcharset2 Symbol;}}\f0 ABC}`,
		`{\rtf1{\fonttbl{\f0 Arial;}}\f1 ABC}`, `{\rtf1{\upr{ANSI}{\*\ud Unicode}}}`, `{\rtf1\fromhtml1 body}`,
		`{\rtf1 A\bin3 xxxB}`, `{\rtf1{\*\ud Unicode}}`, `{\rtf1 A\chpgn B}`,
	} {
		if _, err := decodeRTFText(context.Background(), []byte(rtf), 4096); !errors.Is(err, ErrUnsupported) {
			t.Fatalf("%q: %v, want unsupported", rtf, err)
		}
	}
}

func TestRTFTextBudgetsCancellationAndSingleView(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := decodeRTFText(ctx, []byte(`{\rtf1 ordinary text}`), 100); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	during := &rtfCancelContext{Context: context.Background()}
	if text, err := decodeRTFText(during, []byte(`{\rtf1 `+strings.Repeat("A", 20000)+`}`), 20000); !errors.Is(err, context.Canceled) || len(text) > 4096 {
		t.Fatalf("mid-walk cancellation text=%d err=%v", len(text), err)
	}
	for _, doc := range []string{`{\rtf1\uc0\u8364}`, `{\rtf1 ordinary text}`} {
		b, err := decodeRTFText(context.Background(), []byte(doc), 2)
		if !errors.Is(err, ErrLimit) || len(b) > 2 || cap(b) > 2 {
			t.Fatalf("bounded output len=%d cap=%d err=%v", len(b), cap(b), err)
		}
	}
	deep := []byte(`{\rtf1 ` + strings.Repeat("{", maxRTFTextGroups) + strings.Repeat("}", maxRTFTextGroups) + `}`)
	if _, err := decodeRTFText(context.Background(), deep, 4096); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	var fonts strings.Builder
	fonts.WriteString(`{\rtf1{\fonttbl`)
	for i := 0; i <= maxRTFTextFonts; i++ {
		fmt.Fprintf(&fonts, `{\f%d Arial;}`, i)
	}
	fonts.WriteString(`}body}`)
	if _, err := decodeRTFText(context.Background(), []byte(fonts.String()), 4096); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	data := []byte(`{\rtf1 ordinary ` + strings.Repeat(`{\b text}`, 1000) + `}`)
	view, err := decodeRTFText(context.Background(), data, int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	rules := map[string]string{"dummy.ndb": "RTF.Absent:0:*:" + hx(formatCoverageMarker) + "\n"}
	s := pipelineState(t, rules)
	s.remaining = int64(len(data) + len(view))
	if sig, err := s.visit(data, 0); sig != "" || err != nil || s.remaining != 0 || s.files != 2 {
		t.Fatalf("single view/charge: sig=%q err=%v remaining=%d files=%d", sig, err, s.remaining, s.files)
	}
	for _, kind := range []string{"bytes", "files", "depth"} {
		s := pipelineState(t, rules)
		switch kind {
		case "bytes":
			s.remaining = int64(len(data) + len(view) - 1)
		case "files":
			s.maxFiles = 1
		case "depth":
			s.maxDepth = 0
		}
		expectIncomplete(t, s, data, ErrLimit)
	}
	// A short view does not get matched, but its expansion work is charged.
	for _, body := range []string{"", "AB"} {
		s := pipelineState(t, rules)
		before := s.remaining
		if _, err := s.rtfText([]byte(`{\rtf1 `+body+`}`), 0); err != nil || before-s.remaining != int64(len(body)) {
			t.Fatalf("short view %q: charge=%d err=%v", body, before-s.remaining, err)
		}
	}
	// Exhausted admission refuses work before producing a view. Cancellation
	// after a prefix was produced still charges that discarded prefix once.
	s = pipelineState(t, rules)
	s.files = s.maxFiles
	before := s.remaining
	if _, err := s.rtfText(data, 0); !errors.Is(err, ErrLimit) || s.remaining != before {
		t.Fatalf("pre-admission charge=%d err=%v", before-s.remaining, err)
	}
	s = pipelineState(t, rules)
	s.ctx = &rtfCancelContext{Context: context.Background()}
	before = s.remaining
	if _, err := s.rtfText([]byte(`{\rtf1 `+strings.Repeat("A", 20000)+`}`), 0); !errors.Is(err, context.Canceled) || before-s.remaining <= 0 || before-s.remaining > 4096 {
		t.Fatalf("discarded prefix charge=%d err=%v", before-s.remaining, err)
	}
}

type rtfCancelContext struct {
	context.Context
	calls int
}

func (c *rtfCancelContext) Err() error {
	c.calls++
	if c.calls >= 2 {
		return context.Canceled
	}
	return nil
}

func FuzzRTFTextBounded(f *testing.F) {
	for _, s := range []string{`{\rtf1 plain}`, `{\rtf1\uc1\u65?}`, `{\rtf1{\pict\bin3 \{}}X}`, `{\rtf1 \'ff}`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<16 {
			t.Skip()
		}
		got, _ := decodeRTFText(context.Background(), data, 4096)
		if len(got) > 4096 || cap(got) > 4096 {
			t.Fatalf("unbounded len=%d cap=%d", len(got), cap(got))
		}
	})
}
