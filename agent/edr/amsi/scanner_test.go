package amsi

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

func newTestScanner(t *testing.T, cfg config.EDRConfig) (*Scanner, string) {
	dir := t.TempDir()
	spoolPath := filepath.Join(dir, "e.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	s := NewScanner(cfg, sp)
	return s, spoolPath
}

func TestScanBufferMaliciousBlocksAndEmits(t *testing.T) {
	s, spoolPath := newTestScanner(t, config.Default())
	s.scanBytes = func(addr string, data []byte) (bool, string, error) { return false, "PUA.Script.Test", nil }

	if block := s.ScanBuffer("PowerShell", []byte("bad")); !block {
		t.Fatal("expected block=true for malicious content")
	}
	b, _ := os.ReadFile(spoolPath)
	line := strings.TrimSpace(string(b))
	if !strings.Contains(line, `"source":"amsi"`) || !strings.Contains(line, `"action":"blocked"`) {
		t.Fatalf("bad amsi event: %s", line)
	}
	if strings.Contains(strings.ToLower(line), "clam") {
		t.Fatalf("leaked engine name: %s", line)
	}
}

func TestScanBufferCleanAllows(t *testing.T) {
	s, _ := newTestScanner(t, config.Default())
	s.scanBytes = func(addr string, data []byte) (bool, string, error) { return true, "", nil }
	if block := s.ScanBuffer("PowerShell", []byte("ok")); block {
		t.Fatal("expected block=false for clean content")
	}
}

func TestScanBufferFailOpenVsClosed(t *testing.T) {
	open := config.Default() // FailMode=open
	s, _ := newTestScanner(t, open)
	s.scanBytes = func(addr string, data []byte) (bool, string, error) { return false, "", errors.New("engine down") }
	if block := s.ScanBuffer("PowerShell", []byte("x")); block {
		t.Fatal("fail-open must allow on engine error")
	}

	closed := config.Default()
	closed.FailMode = "closed"
	s2, _ := newTestScanner(t, closed)
	s2.scanBytes = func(addr string, data []byte) (bool, string, error) { return false, "", errors.New("engine down") }
	if block := s2.ScanBuffer("PowerShell", []byte("x")); !block {
		t.Fatal("fail-closed must block on engine error")
	}
}
