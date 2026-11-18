//go:build linux

package netblock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNFTBoundedCommandBufferCancelsAndRetainsOnlyLimit(t *testing.T) {
	cancelled := false
	b := boundedNFTBuffer{limit: 4, cancel: func() { cancelled = true }}
	if n, err := b.Write([]byte("0123456789")); err != nil || n != 10 || !b.exceeded || !cancelled || b.String() != "0123" {
		t.Fatalf("bounded output %+v %d %v", b, n, err)
	}
}

func TestNFTSubprocessDeadlineHelper(t *testing.T) {
	if len(os.Args) == 0 || os.Args[len(os.Args)-1] != "nft-exec-slow-helper" {
		t.Skip("controlled deadline subprocess only")
	}
	time.Sleep(10 * time.Second)
}

func TestNFTSubprocessHasFiniteDeadlineWithoutFirewall(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := (nftExec{path: os.Args[0]}).Run(ctx, nil, "-test.run=^TestNFTSubprocessDeadlineHelper$", "--", "nft-exec-slow-helper")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
		t.Fatalf("subprocess deadline failed %v elapsed %s", err, time.Since(start))
	}
}

func TestNFTRuntimeTokensAndLocksRejectUnsafeFiles(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root ownership in disposable Linux test environment; no firewall operations")
	}
	dir := filepath.Join(t.TempDir(), "runtime")
	if err := secureNFTRuntime(dir); err != nil {
		t.Fatal(err)
	}
	owner, err := nftOwner(dir, "test")
	if err != nil {
		t.Fatal(err)
	}
	same, err := nftOwner(dir, "test")
	if err != nil || same != owner {
		t.Fatal("runtime ownership identity changed", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock, err := nftLock(dir, "test")(ctx)
	if err != nil {
		t.Fatal(err)
	}
	blocked, cancelBlocked := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelBlocked()
	if _, err = nftLock(dir, "test")(blocked); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("concurrent controller obtained ownership: %v", err)
	}
	unlock()
	if err = os.Symlink(filepath.Join(dir, "test.owner"), filepath.Join(dir, "other.owner")); err != nil {
		t.Fatal(err)
	}
	if _, err = nftOwner(dir, "other"); err == nil {
		t.Fatal("symlink ownership token accepted")
	}
	if err = os.Chmod(dir, 0777); err != nil {
		t.Fatal(err)
	}
	if err = secureNFTRuntime(dir); err == nil {
		t.Fatal("world-writable runtime directory accepted")
	}
}
