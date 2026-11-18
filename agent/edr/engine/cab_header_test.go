package engine

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func cabHeaderFixture(t *testing.T, codec string) ([]byte, []byte) {
	t.Helper()
	child := []byte("Independent harmless CAB child for salvage compatibility.\n")
	var b []byte
	switch codec {
	case "stored":
		b = cabMetadataFixture(t, [][]byte{child}, []uint16{0}, []cabMetadataFile{{"child.txt", 0, 0, uint32(len(child))}}, 0)
	case "MSZIP":
		b = cabFixture(t, 1, [][]byte{child}, []cabFixtureFile{{"child.txt", 0, uint32(len(child))}}, false)
	case "LZX":
		b = cabLZXFixture(t, child, 18, false)
		copy(b[60:], "child.txt")
	case "Quantum":
		child = []byte("If you can read this, the Quantum decompressor is working!\n")
		packed, err := hex.DecodeString("d606690bcb47f02c2a3a8f2cabbb3cb933018bd8584b7b01ba6f6d516e3ac367424beb023643d66656ca9e72cc300000")
		if err != nil {
			t.Fatal(err)
		}
		b = cabFixture(t, 18<<8|2, [][]byte{packed}, []cabFixtureFile{{"child.txt", 0, uint32(len(child))}}, false)
	default:
		t.Fatal(codec)
	}
	return b, child
}

// Insert cabinet-set names before a single CFFOLDER and update its physical
// offsets. The member's folder marker is independent of the header flags.
func cabHeaderSet(b []byte, flags, member uint16) []byte {
	var names []byte
	if flags&1 != 0 {
		names = append(names, []byte("previous.cab\x00\x00")...)
	}
	if flags&2 != 0 {
		names = append(names, []byte("next.cab\x00\x00")...)
	}
	le := binary.LittleEndian
	fileAt, dataAt := int(le.Uint32(b[16:])), int(le.Uint32(b[36:]))
	b = append(append(append([]byte(nil), b[:36]...), names...), b[36:]...)
	le.PutUint16(b[30:], flags)
	le.PutUint32(b[16:], uint32(fileAt+len(names)))
	le.PutUint32(b[36+len(names):], uint32(dataAt+len(names)))
	le.PutUint16(b[fileAt+len(names)+8:], member)
	le.PutUint32(b[8:], uint32(len(b)))
	return b
}

func TestCABHeaderConfirmedMisses(t *testing.T) {
	for _, name := range []string{"reserved-field", "alternate-version", "oversize-cabinet", "continued-to-next"} {
		t.Run(name, func(t *testing.T) {
			b, child := cabHeaderFixture(t, "stored")
			switch name {
			case "reserved-field":
				b[12] = 1
			case "alternate-version":
				b[24] = 9
			case "oversize-cabinet":
				binary.LittleEndian.PutUint32(b[8:], uint32(len(b)+100))
			case "continued-to-next":
				b = cabHeaderSet(b, 2, 0xfffe)
			}
			referenceCABSalvage(t, b, child, true)
			if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
				t.Fatalf("header hid complete child: signature=%q error=%v", sig, err)
			}
		})
	}
}

func TestCABHeaderFieldAndLengthMatrix(t *testing.T) {
	for _, codec := range []string{"stored", "MSZIP", "LZX", "Quantum"} {
		for _, mutation := range []string{"reserved-one", "reserved-two", "reserved-three", "minor", "major", "both-versions", "zero-size", "small-size", "size-in-payload", "oversize", "maximum-size", "index", "unknown-flag"} {
			t.Run(codec+"/"+mutation, func(t *testing.T) {
				b, child := cabHeaderFixture(t, codec)
				switch mutation {
				case "reserved-one":
					b[4] = 0x81
				case "reserved-two":
					b[12] = 0x81
				case "reserved-three":
					b[20] = 0x81
				case "minor":
					b[24] = 9
				case "major":
					b[25] = 9
				case "both-versions":
					b[24], b[25] = 0xff, 0xff
				case "zero-size":
					binary.LittleEndian.PutUint32(b[8:], 0)
				case "small-size":
					binary.LittleEndian.PutUint32(b[8:], 35)
				case "size-in-payload":
					binary.LittleEndian.PutUint32(b[8:], uint32(len(b)-1))
				case "oversize":
					binary.LittleEndian.PutUint32(b[8:], uint32(len(b)+100))
				case "maximum-size":
					binary.LittleEndian.PutUint32(b[8:], 0xffffffff)
				case "index":
					binary.LittleEndian.PutUint16(b[34:], 7)
				case "unknown-flag":
					binary.LittleEndian.PutUint16(b[30:], 0x8000)
				}
				if mutation == "reserved-one" {
					// The reference's built-in type signature requires the
					// first reserved word to be zero, unlike the other two.
					referenceCABSalvage(t, b, child, false)
					if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "" || !errors.Is(err, ErrMalformed) {
						t.Fatalf("non-CAB magic extracted a child: %q %v", sig, err)
					}
					return
				}
				referenceCABSalvage(t, b, child, true)
				if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
					t.Fatalf("%s child: signature=%q error=%v", mutation, sig, err)
				}
				noHit := map[string]string{"meta.cdb": "Absent:CL_TYPE_MSCAB:*:^never$:*:*:*:*:*:\n"}
				s := pipelineState(t, noHit)
				s.maxDepth = 2
				if sig, err := s.visit(b, 0); sig != "" || err != nil {
					t.Fatalf("complete physical archive recursed or failed: signature=%q error=%v", sig, err)
				}
			})
		}
	}
}

func TestCABHeaderMissingDataRemainsIncomplete(t *testing.T) {
	for _, codec := range []string{"stored", "MSZIP", "LZX", "Quantum"} {
		for _, marker := range []uint16{0xfffd, 0xffff} {
			t.Run(fmt.Sprintf("%s/from-%x", codec, marker), func(t *testing.T) {
				b, child := cabHeaderFixture(t, codec)
				b = cabHeaderSet(b, 3, marker)
				referenceCABSalvage(t, b, child, false)
				if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
					t.Fatalf("missing predecessor became a child: %q %v", sig, err)
				}
			})
		}
		t.Run(codec+"/split-CFDATA", func(t *testing.T) {
			b, child := cabHeaderFixture(t, codec)
			b = cabHeaderSet(b, 2, 0xfffe)
			at := int(binary.LittleEndian.Uint32(b[46:])) // header + next-set labels
			binary.LittleEndian.PutUint16(b[at+6:], 0)
			referenceCABSalvage(t, b, child, false)
			if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "" || !(errors.Is(err, ErrUnsupported) || errors.Is(err, ErrMalformed)) {
				t.Fatalf("split data invented a complete member: %q %v", sig, err)
			}
		})
	}
	b, child := cabHeaderFixture(t, "stored")
	binary.LittleEndian.PutUint32(b[8:], 0xffffffff)
	b = b[:len(b)-1]
	referenceCABSalvage(t, b, child, false)
	if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("physical truncation accepted: %q %v", sig, err)
	}
}

func TestCABHeaderLocalSetData(t *testing.T) {
	for _, codec := range []string{"stored", "MSZIP", "LZX", "Quantum"} {
		for _, flags := range []uint16{0, 1, 2, 3} {
			for _, member := range []uint16{0, 0xfffe} {
				t.Run(fmt.Sprintf("%s/flags-%d/member-%x", codec, flags, member), func(t *testing.T) {
					b, child := cabHeaderFixture(t, codec)
					b = cabHeaderSet(b, flags, member)
					binary.LittleEndian.PutUint16(b[34:], 17)
					referenceCABSalvage(t, b, child, true)
					if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
						t.Fatalf("local set child: %q %v", sig, err)
					}
					if sig, err := pipelineState(t, map[string]string{"no.cdb": "Absent:CL_TYPE_MSCAB:*:^never$:*:*:*:*:*:\n"}).visit(b, 0); sig != "" || err != nil {
						t.Fatalf("complete local set classified incomplete: %q %v", sig, err)
					}
				})
			}
		}
	}
	t.Run("next-maps-last-folder", func(t *testing.T) {
		child := []byte("Last-folder local continuation control.")
		b := cabMetadataFixture(t, [][]byte{[]byte("unused"), child}, []uint16{15, 0}, []cabMetadataFile{{"last", 0xfffe, 0, uint32(len(child))}}, 2)
		referenceCABSalvage(t, b, child, true)
		if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
			t.Fatalf("continued member used the wrong folder: %q %v", sig, err)
		}
	})
	t.Run("unavailable-first-local-second-folder", func(t *testing.T) {
		child := []byte("Independent local second-folder control.")
		b := cabMetadataFixture(t, [][]byte{[]byte("missing"), child}, []uint16{0, 0}, []cabMetadataFile{{"first", 0xfffd, 0, 7}, {"second", 1, 0, uint32(len(child))}}, 1)
		referenceCABSalvage(t, b, child, true)
		if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "Salvage.Child" || err != nil {
			t.Fatalf("missing first folder hid a local second folder: %q %v", sig, err)
		}
	})
	t.Run("invalid-name-retains-predecessor-state", func(t *testing.T) {
		child := []byte("A skipped name must not erase required history.")
		b := cabMetadataFixture(t, [][]byte{child}, []uint16{0}, []cabMetadataFile{{strings.Repeat("x", 256), 0xfffd, 0, 1}, {"second", 0, 0, uint32(len(child))}}, 0)
		// The bounded failed string read consumes exactly 256 bytes; remove
		// the extra terminator so the next fixed record remains aligned.
		at := 36 + 8 + 16 + 256
		b = append(b[:at], b[at+1:]...)
		binary.LittleEndian.PutUint32(b[36:], binary.LittleEndian.Uint32(b[36:])-1)
		binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
		referenceCABSalvage(t, b, child, false)
		if sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
			t.Fatalf("skipped name erased required predecessor: %q %v", sig, err)
		}
	})
}

func TestCABHeaderEmbeddedAndExtent(t *testing.T) {
	for _, codec := range []string{"stored", "MSZIP", "LZX", "Quantum"} {
		for _, field := range []string{"reserved", "version", "set"} {
			t.Run(codec+"/"+field, func(t *testing.T) {
				b, child := cabHeaderFixture(t, codec)
				switch field {
				case "reserved":
					b[12], b[20] = 0xa5, 0xff
				case "version":
					b[24], b[25] = 0xff, 0xff
				case "set":
					b = cabHeaderSet(b, 3, 0xfffe)
				}
				parent := append(peValidationFixture(), b...)
				referenceCABSalvage(t, parent, child, true)
				if sig, err := pipelineState(t, cabSalvageRule(child)).visit(parent, 0); sig != "Salvage.Child" || err != nil {
					t.Fatalf("embedded advisory header hid child: %q %v", sig, err)
				}
			})
		}
	}
	for _, size := range []uint32{0, 35, 129, 235, 0xffffffff} {
		t.Run(fmt.Sprintf("embedded-declared-%d", size), func(t *testing.T) {
			b, child := cabHeaderFixture(t, "stored")
			binary.LittleEndian.PutUint32(b[8:], size)
			parent := append(peValidationFixture(), b...)
			found := size == 0 || size >= uint32(len(b))
			referenceCABSalvage(t, parent, child, found)
			sig, _ := pipelineState(t, cabSalvageRule(child)).visit(parent, 0)
			if (sig == "Salvage.Child") != found {
				t.Fatalf("embedded declared-size child: %q want hit=%v", sig, found)
			}
		})
	}
	for _, embedded := range []bool{false, true} {
		t.Run(fmt.Sprintf("metadata-embedded-%t", embedded), func(t *testing.T) {
			b, child := cabHeaderFixture(t, "stored")
			size := len(b)
			b = append(b, []byte("trailing harmless bytes")...)
			if embedded {
				b = append(peValidationFixture(), b...)
			} else {
				size = len(b)
			}
			rules := map[string]string{"extent.cdb": fmt.Sprintf("Header.Extent:CL_TYPE_MSCAB:%d:^child[.]txt$:0:%d:0:0:*:\n", size, len(child))}
			referenceCABMetadata(t, b, rules, "Header.Extent")
			if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Header.Extent" || err != nil {
				t.Fatalf("root/embedded physical extent: %q %v", sig, err)
			}
		})
	}
}

func TestCABHeaderRawEmbeddingAndRootMetadata(t *testing.T) {
	for _, size := range []uint32{0, 35, 129, 235, 0xffffffff} {
		t.Run(fmt.Sprintf("raw-embedded-declared-%d", size), func(t *testing.T) {
			b, child := cabHeaderFixture(t, "stored")
			binary.LittleEndian.PutUint32(b[8:], size)
			parent := append(bytes.Repeat([]byte{0xa5}, 128), b...)
			found := size == 0 || size >= uint32(len(b))
			referenceCABSalvage(t, parent, child, found)
			sig, _ := pipelineState(t, cabSalvageRule(child)).visit(parent, 0)
			if (sig == "Salvage.Child") != found {
				t.Fatalf("raw embedded view: %q want hit=%v", sig, found)
			}
		})
	}
	for _, codec := range []string{"stored", "MSZIP", "LZX", "Quantum"} {
		for _, size := range []uint32{0, 0xffffffff} {
			t.Run(fmt.Sprintf("root-metadata-%s/%d", codec, size), func(t *testing.T) {
				b, child := cabHeaderFixture(t, codec)
				binary.LittleEndian.PutUint32(b[8:], size)
				b[12], b[20], b[24], b[25] = 0xff, 0x80, 7, 9
				rules := map[string]string{"meta.cdb": fmt.Sprintf("Header.Meta:CL_TYPE_MSCAB:%d:^child[.]txt$:0:%d:0:0:*:\n", len(b), len(child))}
				referenceCABMetadata(t, b, rules, "Header.Meta")
				if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Header.Meta" || err != nil {
					t.Fatalf("advisory header altered callback fields: %q %v", sig, err)
				}
				rules["meta.cdb"] = fmt.Sprintf("Header.Meta:CL_TYPE_MSCAB:%d:^child[.]txt$:0:%d:0:0:*:\n", len(b)+1, len(child))
				referenceCABMetadata(t, b, rules, "")
				if sig, err := pipelineState(t, rules).visit(b, 0); sig != "" || err != nil {
					t.Fatalf("wrong archive-size callback matched or failed: %q %v", sig, err)
				}
			})
		}
	}
}

func TestCABHeaderPhysicalBoundsAndBudgets(t *testing.T) {
	b, child := cabHeaderFixture(t, "stored")
	binary.LittleEndian.PutUint32(b[8:], 0xffffffff)
	noHit := map[string]string{"no.cdb": "Absent:CL_TYPE_MSCAB:*:^never$:*:*:*:*:*:\n"}
	s := pipelineState(t, noHit)
	var sig string
	var err error
	allocated := measureAllocation(func() { sig, err = s.visit(b, 0) })
	if sig != "" || err != nil || allocated > 16<<20 {
		t.Fatalf("declared length allocated or rejected physical data: %q %v %d bytes", sig, err, allocated)
	}
	for _, limit := range []string{"output", "objects", "depth", "cancellation"} {
		t.Run(limit, func(t *testing.T) {
			s := pipelineState(t, cabSalvageRule(child))
			want := ErrLimit
			switch limit {
			case "output":
				s.remaining = int64(len(b))
			case "objects":
				s.maxFiles = 1
			case "depth":
				s.maxDepth = 0
			case "cancellation":
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				s.ctx = ctx
				want = context.Canceled
			}
			if sig, err := s.visit(b, 0); sig != "" || !errors.Is(err, want) {
				t.Fatalf("advisory size bypassed %s: %q %v", limit, sig, err)
			}
		})
	}
	// Short advisory lengths must still leave a genuine physical suffix to
	// scan. The complete member above ends before this independent object.
	binary.LittleEndian.PutUint32(b[8:], 0)
	trailer := []byte("A separate harmless physical CAB trailer.\n")
	b = append(b, trailer...)
	if sig, err := pipelineState(t, cabSalvageRule(trailer)).visit(b, 0); sig != "Salvage.Child" || err != nil {
		t.Fatalf("short declaration swallowed real trailer: %q %v", sig, err)
	}
}
