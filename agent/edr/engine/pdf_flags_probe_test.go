package engine

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// The pinned ClamAV hook probes lexical edge cases before the simple-PDF
// object-flag profile is extended. Every result remains observable without
// treating an unproved shape as clean.
func TestPDFLiteralDictionaryFlagReferenceProbe(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 distribution")
	}
	clamscan, err := os.ReadFile(filepath.Join(ref, "bin", "clamscan"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(clamscan)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("clamscan is not the pinned ClamAV 1.5.4 executable: %v", err)
	}
	for _, tc := range []struct{ body, file, want string }{
		{"(<< /Page >> << /Catalog >>)", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"(<< /Catalog >> << /Page >>)", "pdf-flags-catalog.cbc", "UTM.PDF.FlagsCatalog"},
		{"(alpha)(<< /Page >>)", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"(alpha) (<< /Page >>)", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"(<< /Page >>) (<< /Catalog >>)", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
		{"(<< /Page >>) << /Type /Catalog >>", "pdf-flags-page.cbc", "UTM.PDF.FlagsPage"},
	} {
		t.Run(fmt.Sprintf("%q", tc.body), func(t *testing.T) {
			pdf := []byte("%PDF-1.7\n1 0 obj\n" + tc.body + "\nendobj\n%%EOF\n")
			hook, err := os.ReadFile("testdata/" + tc.file)
			if err != nil {
				t.Fatal(err)
			}
			if name, _ := referencePDFHook(t, hook, pdf); name != tc.want {
				t.Fatalf("reference hook: %q, want %q", name, tc.want)
			}
			s := pipelineState(t, map[string]string{tc.file: string(hook)})
			if name, err := s.visit(pdf, 0); name != tc.want || err != nil {
				t.Fatalf("Go hook: %q, %v; want %q", name, err, tc.want)
			}
		})
	}
}
