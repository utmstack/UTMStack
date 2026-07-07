package engine

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"testing"
)

func fakeClamd(t *testing.T, reply string) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		// read "zINSTREAM\0"
		_, _ = r.ReadString(0)
		// drain chunks until zero-length terminator
		for {
			var n uint32
			if err := binary.Read(r, binary.BigEndian, &n); err != nil {
				break
			}
			if n == 0 {
				break
			}
			if _, err := io.CopyN(io.Discard, r, int64(n)); err != nil {
				break
			}
		}
		_, _ = conn.Write([]byte(reply))
	}()
	return ln.Addr().String(), func() { ln.Close() }
}

func TestScanBytesParsesFound(t *testing.T) {
	addr, stop := fakeClamd(t, "stream: Win.Test.EICAR_HDB-1 FOUND\x00")
	defer stop()
	clean, sig, err := ScanBytes(addr, []byte("dummy"))
	if err != nil {
		t.Fatalf("ScanBytes: %v", err)
	}
	if clean {
		t.Fatal("expected not clean")
	}
	if !strings.Contains(sig, "EICAR") {
		t.Fatalf("signature = %q", sig)
	}
}

func TestScanBytesParsesOK(t *testing.T) {
	addr, stop := fakeClamd(t, "stream: OK\x00")
	defer stop()
	clean, _, err := ScanBytes(addr, []byte("dummy"))
	if err != nil {
		t.Fatalf("ScanBytes: %v", err)
	}
	if !clean {
		t.Fatal("expected clean")
	}
}
