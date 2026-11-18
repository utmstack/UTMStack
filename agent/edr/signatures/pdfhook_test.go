package signatures

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// pdfHook is a PDF hook program (kind 258, no trigger) with the given types,
// API declarations, locals and one block.
func pdfHook(types, apis []string, apiMax uint64, locals []int, block ...string) callbackFixture {
	return callbackFixture{kind: 258, logical: "BC.Pdf.Test-1", result: 32, types: types, apis: apis, apiMax: apiMax, locals: locals, blocks: [][]string{block}}
}

// breakInPhase is a PDF hook that returns the break value 0xcea5e in the
// given phase (pdf_get_phase) and 0 otherwise.
func breakInPhase(phase uint64) callbackFixture {
	return pdfHook([]string{callbackType(1, 1, 32)}, []string{callbackAPI(87, 69, "pdf_get_phase")}, 87, []int{32, 1, 32},
		callbackCall(0, 87),
		callbackInst(1, 1, 21, callbackNum(32), callbackNum(0), callbackConstant(phase, 32)),
		callbackInst(32, 2, 31, callbackNum(1), callbackConstant(0xcea5e, 32), callbackConstant(0, 32)),
		"T"+callbackFixed(19, 2)+callbackNum(32)+callbackNum(2),
	)
}

func compilePDFHook(t *testing.T, f callbackFixture) *Database {
	t.Helper()
	d, err := Compile(map[string][]byte{"hook.cbc": f.encode(), "a.ndb": []byte("Other:0:*:57575757\n")})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func phaseLabelHook(label string, phase uint64) callbackFixture {
	f := callbackWithLabel([]byte(label), true)
	f.kind, f.logical = 258, ""
	f.types = append(f.types, callbackType(1, 1, 32))
	f.apiMax = 87
	f.apis = append(f.apis, callbackAPI(87, 71, "pdf_get_phase"))
	f.locals = []int{32, 1}
	f.blocks = [][]string{{
		callbackCall(0, 87),
		callbackInst(1, 1, 21, callbackNum(32), callbackNum(0), callbackConstant(phase, 32)),
		"T" + callbackFixed(17, 2) + callbackNum(1) + callbackNum(1) + callbackNum(2),
	}, f.blocks[0], {callbackReturn(32, 0)}}
	return f
}

// pdfAPIEqualsHook labels only when the reached phase and one scalar PDF API
// result agree with a reference fixture.
func pdfAPIEqualsHook(label string, phase, apiID, want uint64, apiName string, apiArg ...uint64) callbackFixture {
	f := callbackWithLabel([]byte(label), true)
	f.kind, f.logical = 258, ""
	f.types = append(f.types, callbackType(1, 1, 32))
	f.apiMax = max(87, apiID)
	apiType := 71
	var callArgs []string
	if len(apiArg) == 1 {
		f.types = append(f.types, callbackType(1, 2, 32, 32))
		apiType = 72
		callArgs = []string{callbackConstant(apiArg[0], 32)}
	}
	f.apis = append(f.apis, callbackAPI(87, 71, "pdf_get_phase"), callbackAPI(int(apiID), apiType, apiName))
	f.locals = []int{32, 32, 1}
	branch := func(yes, no uint64) string {
		return "T" + callbackFixed(17, 2) + callbackNum(2) + callbackNum(yes) + callbackNum(no)
	}
	f.blocks = [][]string{
		{callbackCall(0, 87), callbackInst(1, 2, 21, callbackNum(32), callbackNum(0), callbackConstant(phase, 32)), branch(1, 3)},
		{callbackCall(1, int(apiID), callArgs...), callbackInst(1, 2, 21, callbackNum(32), callbackNum(1), callbackConstant(want, 32)), branch(2, 3)},
		f.blocks[0],
		{callbackReturn(32, 0)},
	}
	return f
}

func TestPDFParserHookFixturesMatchSources(t *testing.T) {
	for name, f := range map[string]callbackFixture{
		"pdf-pre-phase.cbc":            phaseLabelHook("UTM.PDF.Pre", 4),
		"pdf-end-bad-trailer.cbc":      pdfAPIEqualsHook("UTM.PDF.EndBadTrailer", 3, 78, 4, "pdf_get_flags"),
		"pdf-flags-catalog.cbc":        pdfAPIEqualsHook("UTM.PDF.FlagsCatalog", 1, 84, 2, "pdf_getobjflags", 0),
		"pdf-flags-page.cbc":           pdfAPIEqualsHook("UTM.PDF.FlagsPage", 1, 84, 2|1<<23, "pdf_getobjflags", 0),
		"pdf-flags-scalar.cbc":         pdfAPIEqualsHook("UTM.PDF.FlagsScalar", 1, 84, 0, "pdf_getobjflags", 0),
		"pdf-flags-image.cbc":          pdfAPIEqualsHook("UTM.PDF.FlagsImage", 1, 84, 2|1<<18, "pdf_getobjflags", 0),
		"pdf-flags-embedded.cbc":       pdfAPIEqualsHook("UTM.PDF.FlagsEmbedded", 1, 84, 2|1<<2, "pdf_getobjflags", 0),
		"pdf-flags-signed.cbc":         pdfAPIEqualsHook("UTM.PDF.FlagsSigned", 1, 84, 2|1<<17, "pdf_getobjflags", 0),
		"pdf-flags-stream.cbc":         pdfAPIEqualsHook("UTM.PDF.FlagsStream", 1, 84, 3, "pdf_getobjflags", 0),
		"pdf-flags-hex.cbc":            pdfAPIEqualsHook("UTM.PDF.FlagsHex", 1, 84, 3|1<<16|1<<3, "pdf_getobjflags", 0),
		"pdf-flags-flate.cbc":          pdfAPIEqualsHook("UTM.PDF.FlagsFlate", 1, 84, 3|1<<16|1<<5, "pdf_getobjflags", 0),
		"pdf-parsed-phase.cbc":         phaseLabelHook("UTM.PDF.Parsed", 1),
		"pdf-parsed-count.cbc":         pdfAPIEqualsHook("UTM.PDF.ObjectCount", 1, 77, 2, "pdf_get_obj_num"),
		"pdf-dumped-index.cbc":         pdfAPIEqualsHook("UTM.PDF.DumpedIndex", 2, 88, 0, "pdf_get_dumpedobjid"),
		"pdf-parsed-lookup.cbc":        pdfAPIEqualsHook("UTM.PDF.Lookup", 1, 80, 1, "pdf_lookupobj", 0x200),
		"pdf-parsed-size.cbc":          pdfAPIEqualsHook("UTM.PDF.Size", 1, 81, 32, "pdf_getobjsize", 0),
		"pdf-parsed-size-hidden.cbc":   pdfAPIEqualsHook("UTM.PDF.HiddenSize", 1, 81, 34, "pdf_getobjsize", 0),
		"pdf-parsed-id.cbc":            pdfAPIEqualsHook("UTM.PDF.ID", 1, 83, 0x200, "pdf_getobjid", 1),
		"pdf-parsed-offset.cbc":        pdfAPIEqualsHook("UTM.PDF.Offset", 1, 86, 52, "pdf_get_offset", 1),
		"pdf-parsed-offset-hidden.cbc": pdfAPIEqualsHook("UTM.PDF.HiddenOffset", 1, 86, 54, "pdf_get_offset", 1),
	} {
		actual, err := os.ReadFile("../engine/testdata/" + name)
		if err != nil || !bytes.Equal(actual, f.encode()) {
			t.Fatalf("%s drift: %v", name, err)
		}
	}
}

func TestPDFHookPrecedesParsedPhase(t *testing.T) {
	pre := phaseLabelHook("UTM.PDF.Pre", 4)
	parsed := phaseLabelHook("UTM.PDF.Parsed", 1)
	d, err := Compile(map[string][]byte{"pre.cbc": pre.encode(), "parsed.cbc": parsed.encode()})
	if err != nil {
		t.Fatal(err)
	}
	pdf := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")
	result, err := d.MatchObject(context.Background(), pdf, MatchContext{Type: TypePDF})
	if err != nil || result.Name != "UTM.PDF.Pre" {
		t.Fatalf("Go PDF hook order: %+v, %v", result, err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, data := range map[string][]byte{"pre.cbc": pre.encode(), "parsed.cbc": parsed.encode()} {
			if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(dir, "sample.pdf")
		if err := os.WriteFile(path, pdf, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir,
			"--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "UTM.PDF.Pre") {
			t.Fatalf("reference PDF hook order: %v: %s", err, out)
		}
	}
}

func TestPDFParsedHookNeedsParserReachability(t *testing.T) {
	d := compilePDFHook(t, phaseLabelHook("UTM.PDF.Parsed", 1))
	for _, pdf := range [][]byte{
		[]byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n"),
		[]byte("%PDF-1.7\ntruncated dictionary << /Type"),
	} {
		result, err := d.MatchObject(context.Background(), pdf, MatchContext{Type: TypePDF})
		if result.Name != "" || unknownReason(err) != UnknownBytecode {
			t.Fatalf("direct matcher promoted unproved PARSED hook: %+v, %v", result, err)
		}
	}
}

func TestPDFPreLookupHasNoObjects(t *testing.T) {
	f := callbackWithLabel([]byte("UTM.PDF.EmptyPre"), true)
	f.kind, f.logical = 258, ""
	f.types = append(f.types, callbackType(1, 2, 32, 32))
	f.apiMax = 80
	f.apis = append(f.apis, callbackAPI(80, 71, "pdf_lookupobj"))
	f.locals = []int{32, 1}
	f.blocks = [][]string{{
		callbackCall(0, 80, callbackConstant(0x100, 32)),
		callbackInst(1, 1, 21, callbackNum(32), callbackNum(0), callbackConstant(0xffffffff, 32)),
		"T" + callbackFixed(17, 2) + callbackNum(1) + callbackNum(1) + callbackNum(2),
	}, f.blocks[0], {callbackReturn(32, 0)}}
	d := compilePDFHook(t, f)
	pdf := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")
	result, err := d.MatchObject(context.Background(), pdf, MatchContext{Type: TypePDF})
	if err != nil || result.Name != "UTM.PDF.EmptyPre" {
		t.Fatalf("Go PRE object lookup: %+v, %v", result, err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "pre.cbc"), f.encode(), 0o600); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "sample.pdf")
		if err := os.WriteFile(path, pdf, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir,
			"--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "UTM.PDF.EmptyPre") {
			t.Fatalf("reference PRE object lookup: %v: %s", err, out)
		}
	}
}

func TestPDFPreHeaderAndTrailerFlagsReference(t *testing.T) {
	f := callbackWithLabel([]byte("UTM.PDF.PreFlags"), true)
	f.kind, f.logical = 258, ""
	f.types = append(f.types, callbackType(1, 1, 32))
	f.apiMax = 78
	f.apis = append(f.apis, callbackAPI(78, 71, "pdf_get_flags"))
	f.locals = []int{32, 1}
	f.blocks = [][]string{{
		callbackCall(0, 78),
		callbackInst(1, 1, 21, callbackNum(32), callbackNum(0), callbackConstant(5, 32)),
		"T" + callbackFixed(17, 2) + callbackNum(1) + callbackNum(1) + callbackNum(2),
	}, f.blocks[0], {callbackReturn(32, 0)}}
	pdf := []byte("%PDF-2.0\nno trailer\n") // BAD_VERSION | BAD_TRAILER
	if flags, known, present := pdfPreFlags(pdf); !known || !present || flags != 5 {
		t.Fatalf("pre flags = %#x, %t, %t", flags, known, present)
	}
	d := compilePDFHook(t, f)
	result, err := d.MatchObject(context.Background(), pdf, MatchContext{Type: TypePDF})
	if err != nil || result.Name != "UTM.PDF.PreFlags" {
		t.Fatalf("Go PRE flags: %+v, %v", result, err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "flags.cbc"), f.encode(), 0o600); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "sample.pdf")
		if err := os.WriteFile(path, pdf, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir,
			"--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "UTM.PDF.PreFlags") {
			t.Fatalf("reference PRE flags: %v: %s", err, out)
		}
	}
}

func TestPDFEndDetectionNeedsParserFlags(t *testing.T) {
	f := phaseLabelHook("UTM.PDF.End", 3)
	// A canonical one-object PDF with a valid xref leaves ClamAV's PDF flag
	// mask at zero, so it never invokes the END (3) hook.
	header := "%PDF-1.4\n"
	object := "1 0 obj\n<< /Type /Catalog >>\nendobj\n"
	pdf := []byte(header + object + fmt.Sprintf(
		"xref\n0 2\n0000000000 65535 f \n%010d 00000 n \ntrailer\n<< /Size 2 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(header), len(header)+len(object)))
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "end.cbc"), f.encode(), 0o600); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, "sample.pdf")
		if err := os.WriteFile(path, pdf, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir,
			"--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if err != nil || strings.Contains(string(out), " FOUND") {
			t.Fatalf("reference zero-flag PDF should not run END hook: %v: %s", err, out)
		}
	}
	d := compilePDFHook(t, f)
	result, err := d.MatchObject(context.Background(), pdf, MatchContext{Type: TypePDF})
	if result.Name != "" || unknownReason(err) != UnknownBytecode {
		t.Fatalf("Go cannot prove END hook reachability: %+v, %v", result, err)
	}
}

func TestPDFPostDumpRequiresReachedExtractedObject(t *testing.T) {
	f := phaseLabelHook("UTM.PDF.Extracted", 2)
	fixture, err := os.ReadFile("../engine/testdata/pdf-postdump.cbc")
	if err != nil || !bytes.Equal(fixture, f.encode()) {
		t.Fatalf("engine hook fixture drift: %v", err)
	}
	d := compilePDFHook(t, f)
	if name, err := d.MatchPDFPostDump(context.Background(), nil, MatchContext{}); name != "" || err != nil {
		t.Fatalf("empty extraction: %q, %v", name, err)
	}
	if name, err := d.MatchPDFPostDump(context.Background(), []byte("decoded stream"), MatchContext{}); name != "UTM.PDF.Extracted" || err != nil {
		t.Fatalf("reached extraction: %q, %v", name, err)
	}
}

// PDF hook programs (kind 258) run on PDF objects in each of the
// reference's four PDF phases (4 before parsing, 1 after parsing, 2 after
// an object is extracted, 3 at the end), with pdf_get_phase answering the
// phase. A run that returns 0 changes nothing; a detection label detects; the
// break value, whose effect on the reference's PDF parsing is not modeled,
// and the PDF object functions this engine does not answer leave the object
// incomplete.
func TestPDFHooksRunInEachPhase(t *testing.T) {
	pdf := "%PDF-1.4 document"
	for _, phase := range []uint64{1, 2, 3, 4} {
		d := compilePDFHook(t, breakInPhase(phase))
		expectOutcome(t, d, pdf, MatchContext{Type: TypePDF}, "", UnknownBytecode)
		expectOutcome(t, d, "plain text", anyType, "", "")
	}
	// No phase 5: the program returns 0 in every phase.
	expectOutcome(t, compilePDFHook(t, breakInPhase(5)), pdf, MatchContext{Type: TypePDF}, "", "")
	// The object functions are not answered.
	flags := pdfHook([]string{callbackType(1, 1, 32)}, []string{callbackAPI(78, 69, "pdf_get_flags")}, 78, []int{32},
		callbackCall(0, 78), callbackReturn(32, 0))
	expectOutcome(t, compilePDFHook(t, flags), pdf, MatchContext{Type: TypePDF}, "", UnknownBytecode)
	// Buffers are not PDF files to the reference's parser.
	expectOutcome(t, compilePDFHook(t, breakInPhase(1)), pdf, MatchContext{Type: TypePDF, Mode: BufferMode}, "", "")
}

func TestPDFHookDetects(t *testing.T) {
	f := callbackWithLabel([]byte("UTMStack.PDF.Hook"), true)
	f.kind, f.logical = 258, ""
	d := compilePDFHook(t, f)
	expectOutcome(t, d, "%PDF-1.4 document", MatchContext{Type: TypePDF}, "UTMStack.PDF.Hook", "")
	expectOutcome(t, d, "%PDF-1.4 document", MatchContext{Type: TypePDF, Mode: BufferMode}, "", "")
	expectOutcome(t, d, "plain text", anyType, "", "")
}

// With the official feed, the PDF hook programs load and change nothing on
// a PDF file (they return 0 in every phase at level 216), so a PDF with
// nothing to detect is clean rather than incomplete.
func TestOfficialPDFHooksChangeNothing(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR")
	if dir == "" {
		t.Skip("EDR_OFFICIAL_SIGNATURE_DIR not set")
	}
	d, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.pdfHooks) != 2 {
		t.Fatalf("%d PDF hook programs loaded, want the official 2", len(d.pdfHooks))
	}
	pdf := "%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [] /Count 0 >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF\n"
	expectOutcome(t, d, pdf, MatchContext{Type: TypePDF}, "", "")
}
