package engine

import (
	"context"
	"fmt"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func pdfNameFixture(object string, eof bool) []byte {
	end := ""
	if eof {
		end = "%%EOF\n"
	}
	return []byte(fmt.Sprintf("%%PDF-1.7\n1 0 obj\n%s\nendobj\n%s", object, end))
}

func TestPDFObfuscatedNamePolicyAndCacheIdentity(t *testing.T) {
	enabled := nativeTestEngine(t)
	cfg := config.Default()
	cfg.SignatureDBDir = enabled.cfg.SignatureDBDir
	off := false
	cfg.ScanHeuristics.PDFObfuscatedName = &off
	disabled := New(cfg)
	if err := disabled.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	if enabled.Version() == disabled.Version() {
		t.Fatal("different heuristic policies share a verdict cache identity")
	}
	pdf := pdfNameFixture("<< /T#79pe /Catalog >>", true)
	if clean, sig, err := enabled.Scan(context.Background(), pdf); clean || sig != "Heuristics.PDF.ObfuscatedNameObject" || err != nil {
		t.Fatal("enabled:", clean, sig, err)
	}
	if clean, sig, err := disabled.Scan(context.Background(), pdf); !clean || sig != "" || err != nil {
		t.Fatal("disabled:", clean, sig, err)
	}
	if clean, sig, err := disabled.Scan(context.Background(), pdfNameFixture("<< /T#79pe /Catalog >>", false)); clean || sig != "" || err == nil {
		t.Fatal("disabled incomplete PDF:", clean, sig, err)
	}
	withSignature := pdfNameFixture("<< /T#79pe /Catalog /Note ("+testMarker+") >>", true)
	if clean, sig, err := disabled.Scan(context.Background(), withSignature); clean || sig != "Native.Test" || err != nil {
		t.Fatal("disabled heuristic must not disable signatures:", clean, sig, err)
	}
}

func TestPDFObfuscatedCommonNameHeuristic(t *testing.T) {
	e := nativeTestEngine(t)
	for _, tc := range []struct {
		name, object string
		eof          bool
		want         string
		incomplete   bool
	}{
		{"escaped-filter-value", "<< /Filter /Flate#44ecode >>", true, "Heuristics.PDF.ObfuscatedNameObject", false},
		{"escaped-key", "<< /T#79pe /Catalog >>", true, "Heuristics.PDF.ObfuscatedNameObject", false},
		{"escaped-nested", "<< /Names [ << /Open#41ction 2 0 R >> ] >>", true, "Heuristics.PDF.ObfuscatedNameObject", false},
		{"ordinary-name", "<< /Type /Catalog >>", true, "", false},
		{"metadata-name", "<< /Aut#68or (a) >>", true, "", false},
		{"arbitrary-name", "<< /Arb#69trary /Catalog >>", true, "", false},
		{"literal-string", "(/Flate#44ecode)", true, "", false},
		{"literal-in-dictionary", "<< /Title (/Flate#44ecode) >>", true, "", false},
		{"comment-in-dictionary", "<< /Title (a) % /Flate#44ecode\n >>", true, "Heuristics.PDF.ObfuscatedNameObject", false},
		{"top-level-name", "<< /Type /Catalog >>\nendobj\n/T#79pe\n2 0 obj\nnull", true, "", false},
		{"incomplete-document", "<< /T#79pe /Catalog >>", false, "Heuristics.PDF.ObfuscatedNameObject", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clean, got, err := e.Scan(context.Background(), pdfNameFixture(tc.object, tc.eof))
			if got != tc.want || clean != (tc.want == "" && !tc.incomplete) || (err != nil) != tc.incomplete {
				t.Fatalf("clean=%v signature=%q error=%v, want signature=%q incomplete=%v", clean, got, err, tc.want, tc.incomplete)
			}
		})
	}
	t.Run("invalid-flate-still-flags", func(t *testing.T) {
		clean, got, err := e.Scan(context.Background(), pdfFixture("/Filter /Flate#44ecode", []byte("broken")))
		if clean || got != "Heuristics.PDF.ObfuscatedNameObject" || err != nil {
			t.Fatal(clean, got, err)
		}
	})
}
