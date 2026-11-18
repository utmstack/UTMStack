//go:build linux

package netblock

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
)

func TestNFTNamespaceForeignTableIsNeverAdoptedOrDeleted(t *testing.T) {
	nftNamespaceGate(t)
	b, err := newLinuxBlocker(fmt.Sprintf("OwnedForeignRefusal-%s", t.Name()), filepath.Join(t.TempDir(), "runtime"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if !b.closed {
			_ = b.Close()
		}
	})
	nsRun(t, "nft", fmt.Sprintf(`create table inet %s { comment "OWNED_TEST_FOREIGN_TABLE"; }`, b.c.table))
	t.Cleanup(func() { nsRun(t, "nft", "delete", "table", "inet", b.c.table) })
	before := nsRun(t, "nft", "-j", "list", "table", "inet", b.c.table)
	if err := b.SetSafetyPolicy(testSafety()); !errors.Is(err, ErrNetworkOwnership) {
		t.Fatalf("foreign table accepted as managed safety: %v", err)
	}
	if err := b.Reset(); !errors.Is(err, ErrNetworkOwnership) {
		t.Fatalf("foreign reset accepted: %v", err)
	}
	if err := b.Close(); !errors.Is(err, ErrNetworkOwnership) {
		t.Fatalf("foreign close accepted: %v", err)
	}
	after := nsRun(t, "nft", "-j", "list", "table", "inet", b.c.table)
	if !bytes.Equal(before, after) {
		t.Fatalf("foreign table changed: before=%s after=%s", before, after)
	}
	t.Logf("NATIVE_NFT_FOREIGN_UNCHANGED %s", after)
}
