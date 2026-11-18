package engine

import (
	"bytes"
	"encoding/base64"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Declared output sizes are untrusted: decoders must not reserve memory for
// them before the input has produced that much output.
func TestDeclaredSizesDoNotReserveMemory(t *testing.T) {
	const declared = 64 << 20
	few := []byte{0xff, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	if out, err := szddExpand(few, declared); err == nil || cap(out) > 1<<20 {
		t.Errorf("SZDD reserved %d bytes for 9 input bytes (err=%v)", cap(out), err)
	}
	items, _ := arjItems("harmless text ")
	d := &arjDecoder{arjBits: arjBits{in: arjCompress(items, 3)}}
	if out, err := d.decode(declared); err == nil || cap(out) > 1<<20 {
		t.Errorf("ARJ method 1 reserved %d bytes (err=%v)", cap(out), err)
	}
	if out, err := arjDecodeFast(arjCompressFast(items), declared); err == nil || cap(out) > 1<<20 {
		t.Errorf("ARJ method 4 reserved %d bytes (err=%v)", cap(out), err)
	}
}

// An NSIS block's LZMA dictionary is sized from its header; a block of a few
// bytes cannot produce output that needs a large one.
func TestNSISDictionaryFollowsBlockSize(t *testing.T) {
	s := payloadState(t)
	s.maxFile, s.remaining = 200<<20, 400<<20
	block := append([]byte{0x5d, 0, 0, 0, 0x04}, make([]byte, 16)...) // 64 MiB dictionary declared
	runtime.GC()
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	if _, err := s.nsisReader(block); err != nil {
		t.Fatal(err)
	}
	runtime.ReadMemStats(&m1)
	if n := m1.TotalAlloc - m0.TotalAlloc; n > 4<<20 {
		t.Fatalf("a 21-byte block allocated %d bytes", n)
	}
}

// The self-extractor search checks each ARJ header candidate's CRC over up
// to 2600 bytes; crafted overlays full of candidates must not multiply that
// work without bound.
func TestOverlayARJCandidatesAreBounded(t *testing.T) {
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.overlay = bytes.Repeat([]byte{0x60, 0xea, 0x28, 0x0a}, 1<<20)
	// A type database of its own keeps out the built-in table's archive
	// patterns, whose own matching on such data is not what is measured.
	s := pipelineState(t, map[string]string{"a.ndb": "Payload.Test:0:*:" + hx(payloadMarker) + "\n",
		"a.ftm": "0:0:4d5a:PE:CL_TYPE_ANY:CL_TYPE_MSEXE\n"})
	start := time.Now()
	expectIncomplete(t, s, f.build(), ErrLimit)
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("candidate search took %s", d)
	}
}

// Bytes after installer data are overlay data like any other: an archive or
// an executable appended there is analyzed, and installer data cut short is
// incomplete, never clean.
func TestOverlayAfterInstallerData(t *testing.T) {
	gz := gzipFixture(t, []byte(payloadMarker))
	script := bytes.Repeat([]byte("install script "), 20)
	nsis := nsisOverlay(len(script), nsisBlock(nsisLZMA(t, script), true))
	msi := installShieldMSIOverlay(t, []string{"a.ini"}, [][]byte{[]byte("[Info] harmless")})
	record := installShieldOverlay(map[string][]byte{"setup.ini": []byte("[Startup]\r\n")}, []string{"setup.ini"})
	cat := func(parts ...[]byte) []byte { return bytes.Join(parts, nil) }
	pe := func(overlay []byte) []byte {
		f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
		f.overlay = overlay
		return f.build()
	}
	for name, overlay := range map[string][]byte{
		"archive after NSIS data":              cat(nsis, gz),
		"archive after InstallShield MSI data": cat(msi, gz),
		"archive after an unreadable record":   cat(record, []byte("\x01\x02not a record\x00"), gz),
		"archive after two installers":         cat(nsis, msi, gz),
		"archive after NSIS then a bad record": cat(nsis, record, []byte("\x01not a record\x00"), gz),
		"archive in a record after NSIS data":  cat(nsis, installShieldOverlay(map[string][]byte{"a.gz": gz}, []string{"a.gz"})),
		"NSIS after InstallShield records":     cat(record, nsisOverlay(len(script), nsisBlock(gz, false))),
		"archive after a record, then NSIS":    cat(record, []byte("\x01\x02not a record\x00"), gz, nsis),
	} {
		t.Run(name, func(t *testing.T) {
			requireHidden(t, overlay)
			expectDetected(t, payloadState(t), pe(overlay), "Payload.Test")
		})
	}
	// An executable stored after installer data is found by the embedded
	// executable search.
	inner := simplePE(append(bytes.Repeat([]byte{0x90}, 0x200), "stored executable"...)).build()
	expectDetected(t, payloadState(t, inner), pe(cat(nsis, inner)), "Payload.Hash0")
	// A signed installer: the certificate is not overlay data.
	f := simplePE(bytes.Repeat([]byte{0x90}, 0x200))
	f.overlay = cat(nsis, gz)
	f.cert = authenticodeTable(bytes.Repeat([]byte{0x30}, 64), nil)
	expectDetected(t, payloadState(t), f.build(), "Payload.Test")
	// Installer data cut short is incomplete.
	two := installShieldOverlay(map[string][]byte{"a.ini": []byte("[Startup]\r\n"), "b.bin": bytes.Repeat([]byte("x"), 400)}, []string{"a.ini", "b.bin"})
	expectIncomplete(t, payloadState(t), pe(two[:len(two)-100]), ErrMalformed)
	noise := make([]byte, 5000) // incompressible, so a cut lands in its stored data
	for i, x := 0, uint32(7); i < len(noise); i++ {
		x = x*1664525 + 1013904223
		noise[i] = byte(x >> 24)
	}
	msi2 := installShieldMSIOverlay(t, []string{"a.ini", "b.bin"}, [][]byte{[]byte("[Info] harmless"), noise})
	expectIncomplete(t, payloadState(t), pe(msi2[:len(msi2)-400]), ErrMalformed)
	expectIncomplete(t, payloadState(t), pe(msi2[:len(msi2)-installShieldMSIRecord]), ErrMalformed)
	// The number of installer segments in one overlay is bounded.
	tiny := nsisOverlay(4, nsisBlock([]byte("tiny stored block"), false))
	expectIncomplete(t, payloadState(t), pe(bytes.Repeat(tiny, maxOverlaySegments+1)), ErrLimit)
	if sig, err := payloadState(t).visit(pe(bytes.Repeat(tiny, maxOverlaySegments-1)), 0); sig != "" || err != nil {
		t.Fatalf("segments under the bound: sig=%q err=%v", sig, err)
	}
}

// Objects decoded from RTF text and messages split from a mailbox are
// charged to the budget once, not once when decoded or split and again when
// analyzed.
func TestTextContainersChargeOnce(t *testing.T) {
	gz := gzipFixture(t, []byte(payloadMarker))
	plain := make([]byte, 800000)
	for i, x := 0, uint32(1); i < len(plain); i++ {
		x = x*1664525 + 1013904223
		plain[i] = byte(x >> 24)
	}
	plain = append(plain, payloadMarker...)
	object := gzipFixture(t, plain) // incompressible: as large as its content
	rtf := rtfDocument(splitHex(object))
	s := payloadState(t)
	s.remaining = int64(len(rtf)) + int64(len(object)) + int64(len(plain)) + 4096
	expectDetected(t, s, rtf, "Payload.Test")

	body := strings.Repeat("quoted text line\n", 50000)
	attachment := base64.StdEncoding.EncodeToString(gz)
	mbox := []byte("From a@example.com  Thu Jul 31 13:49:50 2008\nFrom: a@example.com\nMIME-Version: 1.0\n" +
		"Content-Type: multipart/mixed; boundary=b\n\n--b\nContent-Type: text/plain\n\n" + body +
		"\n--b\nContent-Type: application/octet-stream\nContent-Transfer-Encoding: base64\n\n" + attachment + "\n--b--\n")
	requireHidden(t, mbox)
	s = payloadState(t)
	s.remaining = int64(len(mbox)) + int64(len(body)) + int64(len(gz))*4 + 4096
	expectDetected(t, s, mbox, "Payload.Test")
	// The budget still stops analysis that does not fit.
	s = payloadState(t)
	s.remaining = int64(len(mbox)) + 1024
	if sig, err := s.visit(mbox, 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("over budget: sig=%q err=%v", sig, err)
	}
}
