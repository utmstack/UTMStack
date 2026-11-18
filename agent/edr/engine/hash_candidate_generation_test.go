package engine

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestHashCandidateOrderingFailureRetainsGeneration(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	input := []byte("harmless retained signature generation control")
	initial := []byte("Hash.Stable:0:*:" + hex.EncodeToString(input) + "\n")
	if _, err := e.ActivateGeneration(map[string][]byte{"stable.ndb": initial}, 1, manifestHash("hash-order-stable")); err != nil {
		t.Fatal(err)
	}
	version := e.Version()
	// This valid but adversarial hash order forces repeated extreme pivots.
	// A classification-sensitive duplicate requires a real candidate decision;
	// it cannot be silently compacted using the old first-name shortcut.
	var body strings.Builder
	var digest [16]byte
	for i := 1; i <= 10000; i++ {
		binary.LittleEndian.PutUint32(digest[:], uint32(i))
		fmt.Fprintf(&body, "%x:37:Candidate.C%05d\n", digest, i)
	}
	binary.LittleEndian.PutUint32(digest[:], 5000)
	fmt.Fprintf(&body, "%x:37:Weak.Candidate\n", digest)
	if _, err := e.ActivateGeneration(map[string][]byte{"ordered.hdb": []byte(body.String())}, 2, manifestHash("hash-order-limited")); !errors.Is(err, signatures.ErrMatchLimit) {
		t.Fatalf("ordering must fail closed: %v", err)
	}
	if e.Version() != version {
		t.Fatal("failed candidate replaced the active generation")
	}
	if clean, name, err := e.Scan(context.Background(), input); clean || name != "Hash.Stable" || err != nil {
		t.Fatalf("old generation unavailable: clean=%v name=%q error=%v", clean, name, err)
	}
}
