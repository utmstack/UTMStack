package procwatch

import (
	"testing"
	"time"
)

func TestOriginalWMICreationDateMatchesNativeIdentity(t *testing.T) {
	created := time.Date(2026, 10, 20, 14, 33, 12, 123456000, time.UTC)
	identity := (created.UnixMicro()+11644473600000000)*10 + 7
	for _, s := range []string{"20261020143312.123456+000", "20261020163312.123456+120", "20261020093312.123456-300"} {
		if !matchesCreationDate(s, identity) {
			t.Errorf("same creation rejected %q", s)
		}
	}
	for _, s := range []string{"", "20261020143312.123455+000", "20261020143312.123456+***", "202610201433**.123456+000", "20261320143312.123456+000", "20261020143312.123456+000more"} {
		if matchesCreationDate(s, identity) {
			t.Errorf("ambiguous/different creation accepted %q", s)
		}
	}
	if matchesCreationDate("20261020143312.123456+000", identity+10) || matchesCreationDate("20261020143312.123456+000", 0) {
		t.Fatal("reused or unknown identity accepted")
	}
}
