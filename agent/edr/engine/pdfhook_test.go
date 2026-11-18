package engine

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func referencePDFHook(t *testing.T, hook, pdf []byte) (string, bool) {
	t.Helper()
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return "", false
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hook.cbc"), hook, 0o600); err != nil {
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
	exit := 0
	if err != nil {
		var failure *exec.ExitError
		if !errors.As(err, &failure) {
			t.Fatalf("reference execution: %v: %s", err, out)
		}
		exit = failure.ExitCode()
	}
	if exit == 0 && strings.Contains(string(out), ": OK") {
		return "", true
	}
	if exit == 1 {
		_, after, ok := strings.Cut(strings.TrimSpace(string(out)), ": ")
		if ok && strings.HasSuffix(after, " FOUND") {
			return strings.TrimSuffix(after, " FOUND"), true
		}
	}
	t.Fatalf("unexpected reference result: exit %d: %s", exit, out)
	return "", false
}

func TestPDFPostDumpHookRunsOnlyAfterReferenceExtraction(t *testing.T) {
	hook, err := os.ReadFile("testdata/pdf-postdump.cbc")
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		pdf  []byte
		want string
	}{
		"plain-stream":              {pdfFixture("", []byte("decoded stream")), "UTM.PDF.Extracted"},
		"hex-stream":                {pdfFixture("/Filter /ASCIIHexDecode", []byte("6465636f6465642073747265616d>")), "UTM.PDF.Extracted"},
		"empty-stream-with-padding": {pdfFixture("", nil), "UTM.PDF.Extracted"},
		"no-stream":                 {[]byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n"), ""},
		// Go also decodes RunLength, but the pinned reference does not
		// extract this filter into a POSTDUMP object.
		"non-reference-filter": {pdfFixture("/Filter /RunLengthDecode", []byte{13, 'd', 'e', 'c', 'o', 'd', 'e', 'd', ' ', 's', 't', 'r', 'e', 'a', 'm', 128}), ""},
	} {
		t.Run(name, func(t *testing.T) {
			if got, checked := referencePDFHook(t, hook, tc.pdf); checked && got != tc.want {
				t.Fatalf("reference signature = %q, want %q", got, tc.want)
			}
			s := pipelineState(t, map[string]string{"postdump.cbc": string(hook)})
			sig, _ := s.visit(tc.pdf, 0)
			if sig != tc.want {
				t.Fatalf("signature = %q, want %q", sig, tc.want)
			}
		})
	}
}

func TestPDFObjectTableHooksMatchReference(t *testing.T) {
	pdf := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n2 0 obj\n<< /Length 14 >>\nstream\ndecoded stream\nendstream\nendobj\n%%EOF\n")
	for _, tc := range []struct{ file, want string }{
		{"pdf-parsed-count.cbc", "UTM.PDF.ObjectCount"},
		{"pdf-dumped-index.cbc", "UTM.PDF.DumpedIndex"},
		{"pdf-parsed-lookup.cbc", "UTM.PDF.Lookup"},
		{"pdf-parsed-size.cbc", "UTM.PDF.Size"},
		{"pdf-parsed-id.cbc", "UTM.PDF.ID"},
		{"pdf-parsed-offset.cbc", "UTM.PDF.Offset"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			hook, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if got, checked := referencePDFHook(t, hook, pdf); checked && got != tc.want {
				t.Fatalf("reference signature = %q, want %q", got, tc.want)
			}
			s := pipelineState(t, map[string]string{tc.file: string(hook)})
			if got, err := s.visit(pdf, 0); got != tc.want {
				t.Fatalf("Go signature = %q, error %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestPDFObjectStreamOuterTableMatchesReference(t *testing.T) {
	for _, tc := range []struct {
		name  string
		count int
		flate bool
	}{
		{"plain-one", 1, false},
		{"plain-two", 2, false},
		{"flate-one", 1, true},
		{"flate-two", 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first := []byte("<< /Type /Page >>")
			header := []byte("3 0 ")
			content := first
			if tc.count == 2 {
				header = []byte("3 0 4 " + strconv.Itoa(len(first)+1) + " ")
				content = append(append([]byte{}, first...), []byte(" << /Type /Pages >>")...)
			}
			stream := append(append([]byte{}, header...), content...)
			filter := ""
			if tc.flate {
				stream = zlibFixture(t, stream)
				filter = " /Filter /FlateDecode"
			}
			pdf := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n2 0 obj\n<< /Type /ObjStm /N " +
				strconv.Itoa(tc.count) + " /First " + strconv.Itoa(len(header)) + " /Length " + strconv.Itoa(len(stream)) + filter + " >>\nstream\n")
			pdf = append(pdf, stream...)
			pdf = append(pdf, []byte("\nendstream\nendobj\n%%EOF\n")...)
			for _, check := range []struct{ file, want string }{
				{"pdf-parsed-count.cbc", "UTM.PDF.ObjectCount"},
				{"pdf-parsed-lookup.cbc", "UTM.PDF.Lookup"},
				{"pdf-parsed-size.cbc", "UTM.PDF.Size"},
				{"pdf-parsed-id.cbc", "UTM.PDF.ID"},
				{"pdf-parsed-offset.cbc", "UTM.PDF.Offset"},
			} {
				t.Run(check.file, func(t *testing.T) {
					hook, err := os.ReadFile("testdata/" + check.file)
					if err != nil {
						t.Fatal(err)
					}
					if got, checked := referencePDFHook(t, hook, pdf); checked && got != check.want {
						t.Fatalf("reference signature = %q, want %q", got, check.want)
					}
					s := pipelineState(t, map[string]string{check.file: string(hook)})
					if got, err := s.visit(pdf, 0); got != check.want {
						t.Fatalf("Go signature = %q, error %v; want %q", got, err, check.want)
					}
				})
			}
		})
	}
}

func TestPDFParsedPhaseRequiresCompletedParser(t *testing.T) {
	hook, err := os.ReadFile("testdata/pdf-parsed-phase.cbc")
	if err != nil {
		t.Fatal(err)
	}
	valid := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")
	if got, checked := referencePDFHook(t, hook, valid); checked && got != "UTM.PDF.Parsed" {
		t.Fatalf("reference signature = %q", got)
	}
	for _, tc := range []struct {
		name string
		pdf  []byte
		want string
	}{
		{"valid", valid, "UTM.PDF.Parsed"},
		{"unterminated-string", []byte("%PDF-1.7\n1 0 obj\n(unterminated"), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := pipelineState(t, map[string]string{"phase.cbc": string(hook)})
			got, err := s.visit(tc.pdf, 0)
			if got != tc.want || tc.want == "" && err == nil {
				t.Fatalf("signature = %q, error %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestPDFPreHookRunsBeforeObjectParsing(t *testing.T) {
	hook, err := os.ReadFile("testdata/pdf-pre-phase.cbc")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		pdf  []byte
	}{
		{"complete", []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")},
		{"unterminated-object", []byte("%PDF-1.7\n1 0 obj\n(unterminated")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, checked := referencePDFHook(t, hook, tc.pdf); checked && got != "UTM.PDF.Pre" {
				t.Fatalf("reference signature = %q", got)
			}
			s := pipelineState(t, map[string]string{"pre.cbc": string(hook)})
			if got, err := s.visit(tc.pdf, 0); got != "UTM.PDF.Pre" {
				t.Fatalf("Go signature = %q, error %v", got, err)
			}
		})
	}
}

func TestPDFSimpleObjectFlagsMatchReference(t *testing.T) {
	for _, tc := range []struct {
		file, body, want string
	}{
		{"pdf-flags-catalog.cbc", "<< /Type /Catalog >>", "UTM.PDF.FlagsCatalog"},
		{"pdf-flags-page.cbc", "<< /Type /Page >>", "UTM.PDF.FlagsPage"},
		{"pdf-flags-scalar.cbc", "123", "UTM.PDF.FlagsScalar"},
		{"pdf-flags-image.cbc", "<< /Subtype /Image >>", "UTM.PDF.FlagsImage"},
		{"pdf-flags-embedded.cbc", "<< /Type /EmbeddedFile >>", "UTM.PDF.FlagsEmbedded"},
		{"pdf-flags-signed.cbc", "<< /Type /Sig >>", "UTM.PDF.FlagsSigned"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			hook, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			pdf := []byte("%PDF-1.7\n1 0 obj\n" + tc.body + "\nendobj\n%%EOF\n")
			if got, checked := referencePDFHook(t, hook, pdf); checked && got != tc.want {
				t.Fatalf("reference signature = %q, want %q", got, tc.want)
			}
			s := pipelineState(t, map[string]string{tc.file: string(hook)})
			if got, err := s.visit(pdf, 0); got != tc.want {
				t.Fatalf("Go signature = %q, error %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestPDFDictionaryNamesInsideComplexValuesMatchReference(t *testing.T) {
	for _, tc := range []struct {
		name, body, want string
	}{
		{"dictionary-inside-string-object", "(<< /Page >>)", "UTM.PDF.FlagsPage"},
		{"dictionary-wrapped-in-string-object", "(abc << /Page >> xyz)", "UTM.PDF.FlagsPage"},
		{"string-inside-dictionary", "<< /Note (hello /Page world) /Type /Catalog >>", "UTM.PDF.FlagsPage"},
		{"comment-inside-dictionary", "<< /Type /Catalog % /Page\n >>", "UTM.PDF.FlagsPage"},
		{"nested-dictionary", "<< /Meta << /Type /Page >> /Type /Catalog >>", "UTM.PDF.FlagsPage"},
		{"array-name", "<< /Types [/Page /Catalog] >>", "UTM.PDF.FlagsPage"},
		{"hex-encoded-name", "<< /Note <2f50616765> /Type /Catalog >>", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hook, err := os.ReadFile("testdata/pdf-flags-page.cbc")
			if err != nil {
				t.Fatal(err)
			}
			pdf := []byte("%PDF-1.7\n1 0 obj\n" + tc.body + "\nendobj\n%%EOF\n")
			if got, checked := referencePDFHook(t, hook, pdf); checked && got != tc.want {
				t.Fatalf("reference signature = %q, want %q", got, tc.want)
			}
			s := pipelineState(t, map[string]string{"page.cbc": string(hook)})
			if got, err := s.visit(pdf, 0); got != tc.want || err != nil {
				t.Fatalf("Go signature = %q, error %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestPDFLeadingStringDictionaryMatchesReference(t *testing.T) {
	for _, tc := range []struct {
		name, body, file, want string
	}{
		{"string dictionary wins page", "(<< /Page >>) << /Type /Catalog >>", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"later dictionary does not add catalog", "(<< /Page >>) << /Type /Catalog >>", "pdf-flags-catalog.cbc", ""},
		{"string dictionary suppresses later page", "(<< /Catalog >>) << /Type /Page >>", "pdf-flags-page.cbc", ""},
		{"string dictionary wins catalog", "(<< /Catalog >>) << /Type /Page >>", "pdf-flags-catalog.cbc", "UTM.PDF.FlagsCatalog"},
		{"later dictionary supplies page", "(plain) << /Type /Page >>", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"later dictionary not catalog", "(plain) << /Type /Page >>", "pdf-flags-catalog.cbc", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hook, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			pdf := []byte("%PDF-1.7\n1 0 obj\n" + tc.body + "\nendobj\n%%EOF\n")
			if got, checked := referencePDFHook(t, hook, pdf); checked && got != tc.want {
				t.Fatalf("ClamAV signature = %q, want %q", got, tc.want)
			}
			s := pipelineState(t, map[string]string{tc.file: string(hook)})
			if got, err := s.visit(pdf, 0); got != tc.want || err != nil {
				t.Fatalf("Go signature = %q, error %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestPDFStreamObjectFlagsMatchReference(t *testing.T) {
	hook, err := os.ReadFile("testdata/pdf-flags-stream.cbc")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, data string
	}{
		{"ordinary", "hello"},
		{"name-inside-stream", "/Page\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pdf := []byte("%PDF-1.7\n1 0 obj\n<< /Length " + strconv.Itoa(len(tc.data)) + " >>\nstream\n" + tc.data + "\nendstream\nendobj\n%%EOF\n")
			if got, checked := referencePDFHook(t, hook, pdf); checked && got != "UTM.PDF.FlagsStream" {
				t.Fatalf("reference signature = %q", got)
			}
			s := pipelineState(t, map[string]string{"stream.cbc": string(hook)})
			if got, err := s.visit(pdf, 0); got != "UTM.PDF.FlagsStream" {
				t.Fatalf("Go signature = %q, error %v", got, err)
			}
		})
	}
}

func TestPDFObjectStreamFlagsMatchReference(t *testing.T) {
	for _, tc := range []struct {
		name, file, want string
		flate            bool
	}{
		{"plain", "pdf-flags-stream.cbc", "UTM.PDF.FlagsStream", false},
		{"flate", "pdf-flags-flate.cbc", "UTM.PDF.FlagsFlate", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hook, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			stream := []byte("3 0 << /Type /Page >>")
			filter := ""
			if tc.flate {
				stream = zlibFixture(t, stream)
				filter = " /Filter /FlateDecode"
			}
			pdf := []byte("%PDF-1.7\n1 0 obj\n<< /Type /ObjStm /N 1 /First 4 /Length " + strconv.Itoa(len(stream)) + filter + " >>\nstream\n")
			pdf = append(pdf, stream...)
			pdf = append(pdf, []byte("\nendstream\nendobj\n%%EOF\n")...)
			if got, checked := referencePDFHook(t, hook, pdf); checked && got != tc.want {
				t.Fatalf("reference signature = %q", got)
			}
			s := pipelineState(t, map[string]string{tc.file: string(hook)})
			if got, err := s.visit(pdf, 0); got != tc.want {
				t.Fatalf("Go signature = %q, error %v", got, err)
			}
			// The member's /Page name is inside the stream. At PARSED the
			// reference has not promoted it to the outer object's flags.
			pageHook, err := os.ReadFile("testdata/pdf-flags-page.cbc")
			if err != nil {
				t.Fatal(err)
			}
			if got, checked := referencePDFHook(t, pageHook, pdf); checked && got != "" {
				t.Fatalf("reference outer page signature = %q", got)
			}
			s = pipelineState(t, map[string]string{"page.cbc": string(pageHook)})
			if got, err := s.visit(pdf, 0); got != "" || err != nil {
				t.Fatalf("Go outer page signature = %q, error %v", got, err)
			}
		})
	}
}

func TestPDFHiddenObjectIDsMatchReference(t *testing.T) {
	for _, tc := range []struct {
		name string
		pdf  string
		want map[string]string
	}{
		{"comment-before", "%PDF-1.7\n% 2 0 obj << /Type /Page >> endobj\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n",
			map[string]string{"pdf-parsed-count.cbc": "UTM.PDF.ObjectCount", "pdf-flags-page.cbc": "UTM.PDF.FlagsPage"}},
		{"comment-after", "%PDF-1.7\n1 0 obj\n<< /Type /Catalog >>\nendobj\n% 2 0 obj << /Type /Page >> endobj\n%%EOF\n",
			map[string]string{"pdf-parsed-count.cbc": "UTM.PDF.ObjectCount", "pdf-parsed-lookup.cbc": "UTM.PDF.Lookup", "pdf-parsed-id.cbc": "UTM.PDF.ID", "pdf-parsed-size-hidden.cbc": "UTM.PDF.HiddenSize", "pdf-parsed-offset-hidden.cbc": "UTM.PDF.HiddenOffset", "pdf-flags-catalog.cbc": "UTM.PDF.FlagsCatalog"}},
		{"string", "%PDF-1.7\n1 0 obj\n(2 0 obj << /Type /Page >> endobj)\nendobj\n%%EOF\n", nil},
		{"stream", "%PDF-1.7\n1 0 obj\n<< /Length 39 >>\nstream\n2 0 obj << /Type /Page >> endobj\n\nendstream\nendobj\n%%EOF\n", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := []string{"pdf-parsed-count.cbc", "pdf-parsed-lookup.cbc", "pdf-parsed-id.cbc"}
			if tc.name == "comment-after" {
				files = append(files, "pdf-parsed-size-hidden.cbc", "pdf-parsed-offset-hidden.cbc")
			}
			if strings.HasPrefix(tc.name, "comment") {
				files = append(files, "pdf-flags-catalog.cbc", "pdf-flags-page.cbc")
			}
			for _, file := range files {
				t.Run(file, func(t *testing.T) {
					hook, err := os.ReadFile("testdata/" + file)
					if err != nil {
						t.Fatal(err)
					}
					want := tc.want[file]
					if got, checked := referencePDFHook(t, hook, []byte(tc.pdf)); checked && got != want {
						t.Fatalf("reference signature = %q, want %q", got, want)
					}
					s := pipelineState(t, map[string]string{file: string(hook)})
					if got, err := s.visit([]byte(tc.pdf), 0); got != want || err != nil {
						t.Fatalf("Go signature = %q, error %v; want %q", got, err, want)
					}
				})
			}
		})
	}
}

func TestPDFFilteredStreamFlagsMatchReference(t *testing.T) {
	for _, tc := range []struct {
		file, want, filter string
		data               []byte
	}{
		{"pdf-flags-hex.cbc", "UTM.PDF.FlagsHex", "/Filter /ASCIIHexDecode", []byte("68656c6c6f>")},
		{"pdf-flags-flate.cbc", "UTM.PDF.FlagsFlate", "/Filter /FlateDecode", zlibFixture(t, []byte("hello"))},
	} {
		t.Run(tc.file, func(t *testing.T) {
			hook, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			pdf := pdfFixture(tc.filter, tc.data)
			if got, checked := referencePDFHook(t, hook, pdf); checked && got != tc.want {
				t.Fatalf("reference signature = %q, want %q", got, tc.want)
			}
			s := pipelineState(t, map[string]string{tc.file: string(hook)})
			if got, err := s.visit(pdf, 0); got != tc.want {
				t.Fatalf("Go signature = %q, error %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestPDFEndHookOnReachedBadTrailer(t *testing.T) {
	hook, err := os.ReadFile("testdata/pdf-end-bad-trailer.cbc")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body, want string
	}{
		{"catalog", "<< /Type /Catalog >>", "UTM.PDF.EndBadTrailer"},
		{"scalar", "123", "UTM.PDF.EndBadTrailer"},
		{"complex-openaction", "<< /OpenAction 2 0 R >>", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pdf := []byte("%PDF-1.7\n1 0 obj\n" + tc.body + "\nendobj\n")
			if got, checked := referencePDFHook(t, hook, pdf); checked && got != tc.want {
				t.Fatalf("reference signature = %q, want %q", got, tc.want)
			}
			s := pipelineState(t, map[string]string{"end.cbc": string(hook)})
			if got, err := s.visit(pdf, 0); got != tc.want || tc.want == "" && err == nil {
				t.Fatalf("Go signature = %q, error %v; want %q", got, err, tc.want)
			}
		})
	}
}
