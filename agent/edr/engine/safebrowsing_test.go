package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestGDBDetectionReachesEngineVerdict(t *testing.T) {
	digest := sha256.Sum256([]byte("malware.example/"))
	db, err := signatures.Compile(map[string][]byte{
		"local.gdb": []byte("S:F:" + hex.EncodeToString(digest[:])),
	})
	if err != nil {
		t.Fatal(err)
	}
	s := scanState{ctx: context.Background(), db: db, maxFile: 1 << 20, remaining: 2 << 20, maxFiles: 100, maxDepth: 8}
	sig, err := s.visit([]byte(`<html><a href="https://malware.example/">Open</a></html>`), 0)
	if err != nil || sig != "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net" {
		t.Fatalf("engine GDB verdict = %q, %v", sig, err)
	}
}

func TestGDBMIMEEmailReachesEngineVerdict(t *testing.T) {
	digest := sha256.Sum256([]byte("evil.example/visit"))
	db, err := signatures.Compile(map[string][]byte{
		"local.gdb": []byte("S:F:" + hex.EncodeToString(digest[:])),
	})
	if err != nil {
		t.Fatal(err)
	}
	mail := []byte("From: sender@example.org\nTo: receiver@example.org\nSubject: inert\nMIME-Version: 1.0\nContent-Type: text/html; charset=UTF-8\n\n<html><a href=\"https://evil.example/visit\">Click here</a></html>\n")
	s := scanState{ctx: context.Background(), db: db, maxFile: 1 << 20, remaining: 2 << 20, maxFiles: 100, maxDepth: 8}
	sig, err := s.visit(mail, 0)
	if err != nil || sig != "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net" {
		t.Fatalf("engine MIME GDB verdict = %q, %v", sig, err)
	}
}

func TestPDBBareUnknownTLDReachesMIMEEngineVerdict(t *testing.T) {
	db, err := signatures.Compile(map[string][]byte{
		"local.pdb": []byte("H:amazon.com\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	mail := []byte("From: sender@example.org\nTo: receiver@example.org\nSubject: inert\nMIME-Version: 1.0\nContent-Type: text/html; charset=UTF-8\n\n<html><a href=\"https://evil.test/\">amazon.com.evil.test</a></html>\n")
	s := scanState{ctx: context.Background(), db: db, maxFile: 1 << 20, remaining: 2 << 20, maxFiles: 100, maxDepth: 8}
	sig, err := s.visit(mail, 0)
	if err != nil || sig != "Heuristics.Phishing.Email.SpoofedDomain" {
		t.Fatalf("engine MIME PDB verdict = %q, %v", sig, err)
	}
}
