package engine

import (
	"context"
	"encoding/hex"
	"errors"
	"os"
	"testing"
)

// testdata/zipcrypto-pkware.zip was made with the system zip utility using
// `zip -P secret`; pinned ClamAV 1.5.4 detects Child.Marker in it with the
// matching .pwdb line. It does not contain the marker in its raw ZIP bytes.
func TestPKWAREPasswordDatabaseScansEncryptedZIP(t *testing.T) {
	archive, err := os.ReadFile("testdata/zipcrypto-pkware.zip")
	if err != nil {
		t.Fatal(err)
	}
	marker := "UTMSTACK-HARMLESS-PASSWORD-CHILD-MARKER-20261031"
	rule := []byte("Child.Marker:0:*:" + hex.EncodeToString([]byte(marker)) + "\n")
	for _, tc := range []struct {
		name, password, cfg string
		wantDetection       bool
	}{
		{"cleartext", "Known.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;secret\n", "", true},
		{"hex", "Known.Password;Engine:81-255,Container:CL_TYPE_ZIP;1;736563726574\n", "", true},
		{"generic", "Known.Password;Engine:81-255,Container:CL_TYPE_ANY;0;secret\n", "", true},
		{"wrong", "Known.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;wrong\n", "", false},
		{"disabled", "Known.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;secret\n", "ARCHIVE:0x3fbfffff\n", false},
		{"no password", "", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string][]byte{"child.ndb": rule}
			if tc.password != "" {
				files["known.pwdb"] = []byte(tc.password)
			}
			if tc.cfg != "" {
				files["daily.cfg"] = []byte(tc.cfg)
			}
			e := activateFiles(t, files)
			clean, sig, err := e.Scan(context.Background(), archive)
			if tc.wantDetection {
				if clean || sig != "Child.Marker" || err != nil {
					t.Fatalf("encrypted child: clean=%v sig=%q err=%v", clean, sig, err)
				}
			} else if clean || sig != "" || !errors.Is(err, ErrUnsupported) {
				t.Fatalf("unreadable child must stay incomplete: clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
}

func TestPKWAREDeflateAndNewerVersionHeader(t *testing.T) {
	archive, err := os.ReadFile("testdata/zipcrypto-deflate.zip")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		version byte
	}{
		{"version 20", 20}, {"version 21", 21},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := append([]byte(nil), archive...)
			b[4] = tc.version // pinned ClamAV also detects this v21 header variant
			e := activateFiles(t, map[string][]byte{
				"child.ndb":  []byte("Child.Marker:0:*:" + hex.EncodeToString([]byte("UTMSTACK-HARMLESS-PASSWORD-CHILD-MARKER-20261031")) + "\n"),
				"known.pwdb": []byte("Known.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;secret\n"),
			})
			clean, sig, err := e.Scan(context.Background(), b)
			if clean || sig != "Child.Marker" || err != nil {
				t.Fatalf("encrypted deflate: clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
}

func TestPKWAREEncryptedLocalRecordWithoutDescriptor(t *testing.T) {
	archive, err := os.ReadFile("testdata/zipcrypto-cloaked.zip")
	if err != nil {
		t.Fatal(err)
	}
	// zipcloak kept a local header without the data-descriptor flag. Pinned
	// ClamAV 1.5.4 detects this member alone and after an MZ-like stub.
	stub := append([]byte{'M', 'Z'}, make([]byte, 126)...)
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"archive", archive},
		{"embedded", append(stub, archive...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := activateFiles(t, map[string][]byte{
				"child.ndb":  []byte("Child.Marker:0:*:" + hex.EncodeToString([]byte("UTMSTACK-HARMLESS-PASSWORD-CHILD-MARKER-20261031")) + "\n"),
				"known.pwdb": []byte("Known.Password;Engine:81-255,Container:CL_TYPE_ZIP;0;secret\n"),
			})
			clean, sig, err := e.Scan(context.Background(), tc.data)
			if clean || sig != "Child.Marker" || err != nil {
				t.Fatalf("encrypted local record: clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
}
