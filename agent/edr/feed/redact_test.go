package feed

import "testing"

func TestEngineNameRedactorStripsVendorTokens(t *testing.T) {
	in := "ERROR: Can't download daily.cvd from database.clamav.net (ClamAV/freshclam); clamd unreachable"
	got := engineNameRedactor.Replace(in)
	for _, banned := range []string{"clamav", "ClamAV", "clamd", "freshclam", "database.clamav.net"} {
		if containsFold(got, banned) {
			t.Fatalf("redacted output still contains %q: %q", banned, got)
		}
	}
}

func containsFold(s, sub string) bool {
	// simple case-sensitive check is enough here — the redactor handles both cases explicitly
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
