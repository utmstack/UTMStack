package engine

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

// Write independent ZIP64 records around a small safe payload. Sentinel sizes
// force the scanner to use 64-bit extras even though no large allocation is
// needed to exercise the format.
func formatForceZIP64(t testing.TB, packed, payload []byte, central64, end64 bool) []byte {
	t.Helper()
	b := formatRawZipFixture(t, zip.Store, packed, payload)
	localName := int(binary.LittleEndian.Uint16(b[26:]))
	at := 30 + localName
	extra := binary.LittleEndian.AppendUint16(nil, 1)
	extra = binary.LittleEndian.AppendUint16(extra, 16)
	extra = binary.LittleEndian.AppendUint64(extra, uint64(len(payload)))
	extra = binary.LittleEndian.AppendUint64(extra, uint64(len(packed)))
	b = append(append(append([]byte(nil), b[:at]...), extra...), b[at:]...)
	binary.LittleEndian.PutUint16(b[28:], 20)
	binary.LittleEndian.PutUint32(b[18:], 0xffffffff)
	binary.LittleEndian.PutUint32(b[22:], 0xffffffff)
	end := bytes.LastIndex(b, []byte("PK\x05\x06"))
	directory := int(binary.LittleEndian.Uint32(b[end+16:])) + 20
	binary.LittleEndian.PutUint32(b[end+16:], uint32(directory))
	if central64 {
		names := int(binary.LittleEndian.Uint16(b[directory+28:]))
		at := directory + 46 + names
		extra := binary.LittleEndian.AppendUint16(nil, 1)
		extra = binary.LittleEndian.AppendUint16(extra, 24)
		extra = binary.LittleEndian.AppendUint64(extra, uint64(len(payload)))
		extra = binary.LittleEndian.AppendUint64(extra, uint64(len(packed)))
		extra = binary.LittleEndian.AppendUint64(extra, 0)
		b = append(append(append([]byte(nil), b[:at]...), extra...), b[at:]...)
		binary.LittleEndian.PutUint16(b[directory+30:], uint16(len(extra)))
		for _, off := range []int{20, 24, 42} {
			binary.LittleEndian.PutUint32(b[directory+off:], 0xffffffff)
		}
		end += len(extra)
		binary.LittleEndian.PutUint32(b[end+12:], uint32(end-directory))
	}
	if end64 {
		endRecord := append([]byte(nil), b[end:]...)
		z64 := make([]byte, 56)
		copy(z64, "PK\x06\x06")
		binary.LittleEndian.PutUint64(z64[4:], 44)
		binary.LittleEndian.PutUint16(z64[12:], 45)
		binary.LittleEndian.PutUint16(z64[14:], 45)
		binary.LittleEndian.PutUint64(z64[24:], 1)
		binary.LittleEndian.PutUint64(z64[32:], 1)
		binary.LittleEndian.PutUint64(z64[40:], uint64(end-directory))
		binary.LittleEndian.PutUint64(z64[48:], uint64(directory))
		locator := make([]byte, 20)
		copy(locator, "PK\x06\x07")
		binary.LittleEndian.PutUint64(locator[8:], uint64(end))
		binary.LittleEndian.PutUint32(locator[16:], 1)
		binary.LittleEndian.PutUint32(endRecord[16:], 0xffffffff)
		binary.LittleEndian.PutUint16(endRecord[8:], 0xffff)
		binary.LittleEndian.PutUint16(endRecord[10:], 0xffff)
		b = append(append(append(b[:end], z64...), locator...), endRecord...)
	}
	return b
}

func FuzzZIP64Bounds(f *testing.F) {
	payload := []byte("ordinary harmless ZIP64 fuzz control")
	f.Add(formatForceZIP64(f, payload, payload, true, true))
	f.Add(formatForceZIP64(f, payload, payload, false, false))
	f.Add([]byte("PK\x05\x06"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		// Parser output is bounded by the member count, regardless of the
		// values of the sizes/offsets being mutated.
		members, _ := zipMembers(data, 100)
		if len(members) > 100 {
			t.Fatal("member limit exceeded")
		}
		for _, m := range members {
			if m.data >= 0 && (uint64(m.data) > uint64(len(data)) || m.compressed > uint64(len(data))-uint64(m.data)) {
				t.Fatal("member outside source")
			}
		}
	})
}

func TestZIP64BoundedExtraction(t *testing.T) {
	rules := map[string]string{"format.ndb": "Format.Coverage:0:*:" + hx(formatCoverageMarker) + "\n"}
	for _, central64 := range []bool{false, true} {
		for _, end64 := range []bool{false, true} {
			for _, positive := range []bool{false, true} {
				payload, want := []byte("ordinary harmless ZIP64 control"), ""
				if positive {
					payload, want = []byte(formatCoverageMarker), "Format.Coverage"
				}
				gz := gzipFixture(t, payload)
				data := formatForceZIP64(t, gz, gz, central64, end64)
				if bytes.Contains(data, []byte(formatCoverageMarker)) {
					t.Fatal("fixture exposes marker")
				}
				expectScan(t, pipelineState(t, rules), data, want, "")
				// The standard library provides a second independent format
				// validity check, without extracting to the filesystem.
				zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
				if err != nil || len(zr.File) != 1 {
					t.Fatalf("ZIP64 fixture: %v", err)
				}
			}
		}
	}
	// Local-header ZIP64 records are also decoded inside a different format.
	gz := gzipFixture(t, []byte(formatCoverageMarker))
	data := formatForceZIP64(t, gz, gz, false, false)
	central := bytes.Index(data, []byte("PK\x01\x02"))
	sfxRules := map[string]string{"format.ndb": rules["format.ndb"], "format.ftm": sfxTypes}
	expectScan(t, pipelineState(t, sfxRules), append(append([]byte(nil), pngPrefix...), data[:central]...), "Format.Coverage", "")
}

func TestZIP64MalformedDoesNotOverflow(t *testing.T) {
	gz := gzipFixture(t, []byte("ordinary harmless ZIP64 control"))
	rules := map[string]string{"format.ndb": "Never:0:*:" + hx(formatCoverageMarker) + "\n"}
	for _, which := range []string{"missing-extra", "truncated-extra", "compressed-overflow", "locator-past-end", "record-size-overflow", "directory-overflow", "multidisk"} {
		t.Run(which, func(t *testing.T) {
			data := formatForceZIP64(t, gz, gz, true, true)
			extra := 30 + int(binary.LittleEndian.Uint16(data[26:]))
			end := bytes.LastIndex(data, []byte("PK\x05\x06"))
			z64 := int(binary.LittleEndian.Uint64(data[end-12:]))
			want := ErrMalformed
			switch which {
			case "missing-extra":
				binary.LittleEndian.PutUint16(data[extra:], 2)
			case "truncated-extra":
				binary.LittleEndian.PutUint16(data[extra+2:], 8)
			case "compressed-overflow":
				binary.LittleEndian.PutUint64(data[extra+12:], ^uint64(0))
			case "locator-past-end":
				binary.LittleEndian.PutUint64(data[end-12:], ^uint64(0))
			case "record-size-overflow":
				binary.LittleEndian.PutUint64(data[z64+4:], ^uint64(0))
			case "directory-overflow":
				binary.LittleEndian.PutUint64(data[z64+48:], ^uint64(0))
			case "multidisk":
				binary.LittleEndian.PutUint32(data[end-4:], 2)
				want = ErrUnsupported
			}
			if sig, err := pipelineState(t, rules).visit(data, 0); sig != "" || !errors.Is(err, want) {
				t.Fatalf("scan: %q %v; want %v", sig, err, want)
			}
		})
	}
}
