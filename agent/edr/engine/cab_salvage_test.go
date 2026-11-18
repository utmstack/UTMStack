package engine

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
)

func cabSalvageRule(child []byte) map[string]string {
	return map[string]string{"child.hdb": fmt.Sprintf("%x:%d:Salvage.Child\n", md5.Sum(child), len(child))}
}

// Emit a raw DEFLATE stored block directly from RFC 1951 field widths. This
// avoids relying on the production flate reader or a compressor's tree policy.
func cabSalvageMSZIP(t *testing.T, child []byte) []byte {
	t.Helper()
	if len(child) > 32768 {
		t.Fatal("MSZIP fixture exceeds one frame")
	}
	packed := append([]byte{'C', 'K', 1, byte(len(child)), byte(len(child) >> 8), byte(^uint16(len(child))), byte(^uint16(len(child)) >> 8)}, child...)
	b := cabFixture(t, 1, [][]byte{child}, []cabFixtureFile{{"child", 0, uint32(len(child))}}, false)
	at := int(binary.LittleEndian.Uint32(b[36:]))
	b = append(b[:at+8], packed...)
	binary.LittleEndian.PutUint32(b[at:], 0)
	binary.LittleEndian.PutUint16(b[at+4:], uint16(len(packed)))
	binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
	return b
}

func referenceCABSalvage(t *testing.T, parent, child []byte, found bool) {
	t.Helper()
	want := ""
	if found {
		want = "Salvage.Child"
	}
	out := referenceCABMetadata(t, parent, cabSalvageRule(child), want)
	if found && out != nil && !bytes.Contains(out, []byte(fmt.Sprintf("%x", sha256.Sum256(child)))) {
		t.Fatalf("reference did not identify the exact decoded child: %s", out)
	}
}

func TestCABSalvageConfirmedMisses(t *testing.T) {
	child := []byte("Independent harmless CAB child for salvage compatibility.\n")
	for _, mutate := range []string{"valid", "wrong-checksum", "bad-later-block"} {
		t.Run(mutate, func(t *testing.T) {
			b := cabMetadataFixture(t, [][]byte{child}, []uint16{0}, []cabMetadataFile{{"child.txt", 0, 0, uint32(len(child))}}, 0)
			at := int(binary.LittleEndian.Uint32(b[36:]))
			switch mutate {
			case "wrong-checksum":
				b[at] = 1
			case "bad-later-block":
				binary.LittleEndian.PutUint16(b[40:], 2)
				b = append(b, 0, 0, 0, 0, 10, 0, 10, 0, 1)
				binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
			}
			referenceCABSalvage(t, b, child, true)
			if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
				t.Fatalf("valid child hidden by %s: signature=%q error=%v", mutate, sig, err)
			}
		})
	}
}

func TestCABSalvageChecksumsAllCodecs(t *testing.T) {
	child := []byte("If you can read this, the Quantum decompressor is working!\n")
	for _, method := range []string{"stored", "MSZIP", "LZX", "Quantum"} {
		t.Run(method, func(t *testing.T) {
			var b []byte
			switch method {
			case "stored":
				b = cabFixture(t, 0, [][]byte{child}, []cabFixtureFile{{"child", 0, uint32(len(child))}}, false)
			case "MSZIP":
				b = cabSalvageMSZIP(t, child)
			case "LZX":
				b = cabLZXFixture(t, child, 18, false)
			case "Quantum":
				b = cabQuantumControl(t)
			}
			folder := 36
			if method == "Quantum" {
				folder += 2 * 8
			}
			at := int(binary.LittleEndian.Uint32(b[folder:]))
			binary.LittleEndian.PutUint32(b[at:], 0x12345678)
			referenceCABSalvage(t, b, child, true)
			if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
				t.Fatalf("checksum hid %s: signature=%q error=%v", method, sig, err)
			}
		})
	}
}

func TestCABSalvageCompletePrefixAndPartialMember(t *testing.T) {
	// A full 32 KiB first frame is independent of a subsequent damaged block.
	// Hash-only rules ensure neither a raw archive marker nor its declared
	// length can pass as evidence of extracting the exact available child.
	child := bytes.Repeat([]byte("independent harmless CAB frame! "), 1024)
	if len(child) != 32768 {
		t.Fatal(len(child))
	}
	for _, method := range []string{"stored", "MSZIP", "LZX", "Quantum", "Quantum-small-window"} {
		for _, partial := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/partial-%v", method, partial), func(t *testing.T) {
				var b []byte
				want := child
				switch method {
				case "stored":
					b = cabFixture(t, 0, [][]byte{child}, []cabFixtureFile{{"child", 0, uint32(len(child))}}, false)
				case "MSZIP":
					b = cabSalvageMSZIP(t, child)
				case "LZX":
					b = cabLZXFixture(t, child, 18, false)
				case "Quantum", "Quantum-small-window":
					var err error
					file, hash := "literals-18-32768.cab", "54013a1cc070f3fbe4a2b6595af5f64c960ebc029f5343cb071135ee42b8f00d"
					if method == "Quantum-small-window" {
						file, hash = "literals-10-32768.cab", "efbdd6cebf9c1613f7b7ef47c61141f7d765e79d145ee4639e9321afb7a4fa48"
					}
					b, err = os.ReadFile("testdata/quantum/" + file)
					if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
						t.Fatalf("Quantum input: %v", err)
					}
					want = bytes.Repeat([]byte("OpenEDR harmless arithmetic model control. "), 782)[:32768]
					copy(want[len(want)-len("QUANTUM-HIDDEN-CHILD"):], "QUANTUM-HIDDEN-CHILD")
				}
				if method != "stored" {
					// Keep the reference decoder's 4096-byte read-ahead within
					// this valid block. Otherwise the bad following block aborts
					// that read before any complete frame is handed to it.
					at := int(binary.LittleEndian.Uint32(b[36:]))
					packed := int(binary.LittleEndian.Uint16(b[at+4:]))
					if method == "Quantum" || method == "Quantum-small-window" {
						b = append(b, 0xff) // explicit frame trailer before read-ahead padding
						packed++
					}
					padding := (packed+2+4095)/4096*4096 - packed
					b = append(b, make([]byte, padding)...)
					binary.LittleEndian.PutUint16(b[at+4:], uint16(packed+padding))
				}
				binary.LittleEndian.PutUint16(b[40:], 2)
				if partial {
					binary.LittleEndian.PutUint32(b[44:], uint32(len(want)+100))
				}
				b = append(b, 0, 0, 0, 0, 20, 0, 100, 0, 1)
				binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
				found := (method != "Quantum" && method != "Quantum-small-window") || !partial
				referenceCABSalvage(t, b, want, found)
				sig, err := pipelineState(t, cabSalvageRule(want)).visit(b, 0)
				if found && (sig != "Salvage.Child" || err != nil) {
					t.Fatalf("valid prefix lost: signature=%q error=%v", sig, err)
				} else if !found && (sig != "" || !errors.Is(err, ErrMalformed)) {
					t.Fatalf("Quantum request was incorrectly truncated to CFDATA size: signature=%q error=%v", sig, err)
				}
				noHit := map[string]string{"no.hdb": fmt.Sprintf("%x:%d:Absent\n", md5.Sum([]byte("unrelated")), len(want))}
				sig, err = pipelineState(t, noHit).visit(b, 0)
				if method == "MSZIP" && !partial {
					if sig != "" || err != nil {
						t.Fatalf("unread later record prevented a complete requested file: signature=%q error=%v", sig, err)
					}
				} else if sig != "" || !errors.Is(err, ErrMalformed) {
					t.Fatalf("failed extraction became complete: signature=%q error=%v", sig, err)
				}
			})
		}
	}
}

func TestCABSalvagePrefixAccountingAndCancellation(t *testing.T) {
	child := bytes.Repeat([]byte("prefix data "), 32)
	b := cabFixture(t, 0, [][]byte{child}, []cabFixtureFile{{"first", 0, uint32(len(child))}, {"second", 0, uint32(len(child))}}, false)
	binary.LittleEndian.PutUint16(b[40:], 2)
	b = append(b, 1)
	binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
	rules := map[string]string{"meta.cdb": "Absent:CL_TYPE_MSCAB:*:^never$:*:*:*:*:*:\n"}
	s := pipelineState(t, rules)
	s.remaining = int64(len(b) + len(child))
	if sig, err := s.visit(b, 0); sig != "" || !errors.Is(err, ErrLimit) || s.remaining != 0 || s.files != 3 {
		t.Fatalf("repeated partial output escaped budget: signature=%q error=%v remaining=%d files=%d", sig, err, s.remaining, s.files)
	}
	s = pipelineState(t, cabSalvageRule(child))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.ctx = ctx
	if sig, err := s.visit(b, 0); sig != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled salvage scanned output: signature=%q error=%v", sig, err)
	}
}

func TestCABSalvageDeclaredLengthsAndFullWidthInput(t *testing.T) {
	child := []byte("independent harmless declared-length control\n")
	for _, method := range []uint16{0, 1} {
		for _, declared := range []uint16{1, uint16(len(child)), 65535} {
			t.Run(fmt.Sprintf("method-%d/declared-%d", method, declared), func(t *testing.T) {
				b := cabFixture(t, 0, [][]byte{child}, []cabFixtureFile{{"child", 0, uint32(len(child))}}, false)
				if method == 1 {
					b = cabSalvageMSZIP(t, child)
				}
				at := int(binary.LittleEndian.Uint32(b[36:]))
				binary.LittleEndian.PutUint16(b[at+6:], declared)
				referenceCABSalvage(t, b, child, true)
				s := pipelineState(t, cabSalvageRule(child))
				s.remaining = int64(len(b) + len(child))
				if sig, err := s.visit(b, 0); sig != "Salvage.Child" || err != nil || s.remaining != 0 {
					t.Fatalf("declared length changed stored/DEFLATE output: signature=%q error=%v remaining=%d", sig, err, s.remaining)
				}
			})
		}
	}
	t.Run("65535-byte-stored-block", func(t *testing.T) {
		child := bytes.Repeat([]byte{0x53}, 65535)
		b := cabFixture(t, 0, [][]byte{child}, []cabFixtureFile{{"child", 0, uint32(len(child))}}, false)
		referenceCABSalvage(t, b, child, true)
		if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
			t.Fatalf("full-width stored block: signature=%q error=%v", sig, err)
		}
	})
	t.Run("65535-byte-MSZIP-input", func(t *testing.T) {
		b := cabSalvageMSZIP(t, child)
		at := int(binary.LittleEndian.Uint32(b[36:]))
		packed := int(binary.LittleEndian.Uint16(b[at+4:]))
		b = append(b, make([]byte, 65535-packed)...)
		binary.LittleEndian.PutUint16(b[at+4:], 65535)
		binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
		referenceCABSalvage(t, b, child, true)
		if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
			t.Fatalf("full-width MSZIP input: signature=%q error=%v", sig, err)
		}
	})
}

func TestCABSalvageEarlierChildWinsLaterLimitAndMetadata(t *testing.T) {
	child := []byte("independent earlier child before bounded expansion\n")
	b := cabFixture(t, 0, [][]byte{child, bytes.Repeat([]byte{'x'}, 1000)}, []cabFixtureFile{
		{"earlier", 0, uint32(len(child))}, {"later", uint32(len(child)), 1000},
	}, false)
	rules := cabSalvageRule(child)
	rules["meta.cdb"] = "Later.Meta:CL_TYPE_MSCAB:*:^later$:*:*:*:*:*:\n"
	referenceCABSalvage(t, b, child, true)
	s := pipelineState(t, rules)
	s.remaining = int64(len(b) + len(child))
	if sig, err := s.visit(b, 0); sig != "Salvage.Child" || err != nil || s.remaining != 0 {
		t.Fatalf("later expansion/metadata hid earlier child: signature=%q error=%v remaining=%d", sig, err, s.remaining)
	}
}

func TestCABSalvageEarlierCompressedChildBeforeCodecFailure(t *testing.T) {
	t.Run("Quantum-earlier-member", func(t *testing.T) {
		packed, err := hex.DecodeString("d606690bcb47f02c2a3a8f2cabbb3cb933018bd8584b7b01ba6f6d516e3ac367424beb023643d66656ca9e72cc300000")
		if err != nil {
			t.Fatal(err)
		}
		child := []byte("If you can read this, the Quantum decompressor is working!\n")
		b := cabFixture(t, 18<<8|2, [][]byte{packed}, []cabFixtureFile{{"first", 0, uint32(len(child))}, {"later", uint32(len(child)), 10000}}, false)
		referenceCABSalvage(t, b, child, true)
		if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
			t.Fatalf("later Quantum request hid completed earlier member: signature=%q error=%v", sig, err)
		}
	})
	t.Run("LZX-completed-frame", func(t *testing.T) {
		child := bytes.Repeat([]byte("independent harmless CAB frame! "), 1024)
		b := cabLZXFixture(t, append(bytes.Clone(child), child...), 18, false)
		first := int(binary.LittleEndian.Uint32(b[36:]))
		second := first + 8 + int(binary.LittleEndian.Uint16(b[first+4:]))
		b = b[:second+8+100] // complete CFDATA record; its LZX frame is incomplete
		binary.LittleEndian.PutUint16(b[second+4:], 100)
		binary.LittleEndian.PutUint32(b[44:], uint32(len(child)))
		binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
		referenceCABSalvage(t, b, child, true)
		if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
			t.Fatalf("later LZX failure hid a completed frame: signature=%q error=%v", sig, err)
		}
	})
}
