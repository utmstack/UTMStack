package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func activateFiles(t *testing.T, files map[string][]byte) *Engine {
	t.Helper()
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	if _, err := e.ActivateGeneration(files, 10, manifestHash("followup-generation")); err != nil {
		t.Fatalf("activate: %v", err)
	}
	return e
}

// An undecidable rule on a container must not hide a detection inside it:
// the incomplete result only stands when nothing is detected.
func TestUndecidedOuterRuleStillScansContainerMembers(t *testing.T) {
	e := activateFiles(t, map[string][]byte{
		"logical.cbc": engineCBCProgram(256, "abort", false),
		"plain.ndb":   []byte("Native.Test:0:*:" + hex.EncodeToString([]byte(testMarker)) + "\n"),
	})
	build := func(content []byte) []byte {
		var b bytes.Buffer
		w := zip.NewWriter(&b)
		// The member name carries the bytecode trigger in the archive's raw
		// bytes; the deflated member body hides the plain marker from them.
		f, err := w.CreateHeader(&zip.FileHeader{Name: engineCBCMarker, Method: zip.Deflate})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Write(content); err != nil {
			t.Fatal(err)
		}
		if err = w.Close(); err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(b.Bytes(), []byte(testMarker)) {
			t.Fatal("fixture exposes the member marker in raw bytes")
		}
		return b.Bytes()
	}
	clean, sig, err := e.Scan(context.Background(), build([]byte("padding "+testMarker+" padding")))
	if clean || sig != "Native.Test" || err != nil {
		t.Fatalf("member detection hidden by undecided outer rule: clean=%v sig=%q err=%v", clean, sig, err)
	}
	clean, sig, err = e.Scan(context.Background(), build([]byte("harmless member content")))
	if clean || sig != "" || !errors.Is(err, signatures.ErrRuleIncomplete) {
		t.Fatalf("undecided rule without any detection must stay incomplete: clean=%v sig=%q err=%v", clean, sig, err)
	}
}

// The reference engine never analyzes objects of five bytes or less, so hash
// rules for tiny or empty objects cannot flag them, at any depth.
func TestObjectsOfFiveBytesOrLessAreNotAnalyzed(t *testing.T) {
	five, six := []byte("ABCDE"), []byte("ABCDEF")
	hashRule := func(b []byte, name string) string {
		return fmt.Sprintf("%x:%d:%s\n", md5.Sum(b), len(b), name)
	}
	empty := md5.Sum(nil)
	e := activateFiles(t, map[string][]byte{
		"tiny.hdb": []byte(hashRule(five, "Tiny.Five") + hashRule(six, "Small.Six") +
			hex.EncodeToString(empty[:]) + ":*:Any.Empty:73\n"),
		"tiny.ndb": []byte("Tiny.Body:0:*:414243\n"),
	})
	for name, input := range map[string][]byte{
		"five": five, "empty": {}, "nested-five": zipFixture(t, five),
	} {
		if clean, sig, err := e.Scan(context.Background(), input); !clean || sig != "" || err != nil {
			t.Fatalf("%s: tiny object analyzed: clean=%v sig=%q err=%v", name, clean, sig, err)
		}
	}
	if clean, sig, err := e.Scan(context.Background(), six); clean || sig != "Tiny.Body" || err != nil {
		t.Fatalf("six-byte object: clean=%v sig=%q err=%v", clean, sig, err)
	}
}
