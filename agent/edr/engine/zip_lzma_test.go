package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/ulikunitz/xz/lzma"
)

func zipLZMAPacked(t testing.TB, payload []byte, eos bool) []byte {
	t.Helper()
	var b bytes.Buffer
	w, err := (lzma.WriterConfig{DictCap: 1 << 16, SizeInHeader: true, Size: int64(len(payload)), EOSMarker: eos}).NewWriter(&b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = w.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	classic := b.Bytes()
	packed := append([]byte{9, 4, 5, 0}, classic[:5]...)
	return append(packed, classic[13:]...)
}

func zipLZMAFixture(t testing.TB, packed, payload []byte, eos bool) []byte {
	b := formatRawZipFixture(t, zip14LZMA, packed, payload)
	if eos {
		binary.LittleEndian.PutUint16(b[6:], 2)
		central := bytes.Index(b, []byte("PK\x01\x02"))
		binary.LittleEndian.PutUint16(b[central+8:], 2)
	}
	return b
}

// Produced by Python's independent standard-library ZIP_LZMA writer, using a
// fixed ZipInfo date and two inert text strings. No sample is executed.
func TestZIPLZMAIndependentPythonFixtures(t *testing.T) {
	rules := map[string]string{"format.ndb": "Format.Coverage:0:*:" + hx(formatCoverageMarker) + "\n"}
	for i, encoded := range []string{
		"UEsDBD8AAgAOAAAAIVxIVFjDOgAAACgAAAANAAAAaW5lcnQuZml4dHVyZQkEBQBdAACAAAAqlQW9tzGj7bnrWlICDgIOQ9xW9rVfGQnHvm5TQ947LGqlf/TH5jRp1qn//7JvAABQSwECPwM/AAIADgAAACFcSFRYwzoAAAAoAAAADQAAAAAAAAAAAAAAgAEAAAAAaW5lcnQuZml4dHVyZVBLBQYAAAAAAQABADsAAABlAAAAAAA=",
		"UEsDBD8AAgAOAAAAIVyCOqyYMQAAAB4AAAANAAAAaW5lcnQuZml4dHVyZQkEBQBdAACAAAA3nIisLztz4F5W0V88VZb8odhvjCkkJEySBZ4QjBDCqNCNf/jHAABQSwECPwM/AAIADgAAACFcgjqsmDEAAAAeAAAADQAAAAAAAAAAAAAAgAEAAAAAaW5lcnQuZml4dHVyZVBLBQYAAAAAAQABADsAAABcAAAAAAA=",
	} {
		data, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte(formatCoverageMarker)) {
			t.Fatal("fixture exposes marker")
		}
		want := ""
		if i == 0 {
			want = "Format.Coverage"
		}
		expectScan(t, pipelineState(t, rules), data, want, "")
	}
}

func TestZIPLZMAEOSAndKnownSize(t *testing.T) {
	rules := map[string]string{"format.ndb": "Format.Coverage:0:*:" + hx(formatCoverageMarker) + "\n"}
	for _, eos := range []bool{false, true} {
		payload := gzipFixture(t, []byte(formatCoverageMarker))
		packed := zipLZMAPacked(t, payload, eos)
		data := zipLZMAFixture(t, packed, payload, eos)
		expectScan(t, pipelineState(t, rules), data, "Format.Coverage", "")
		flags := uint16(0)
		if eos {
			flags = 2
		}
		local := localHeader(flags, zip14LZMA, uint32(len(packed)), uint32(len(payload)), "inert.fixture", packed)
		sfxRules := map[string]string{"format.ndb": rules["format.ndb"], "format.ftm": sfxTypes}
		expectScan(t, pipelineState(t, sfxRules), append(append([]byte(nil), pngPrefix...), local...), "Format.Coverage", "")
	}
}

func TestZIPLZMALimitsAndMalformed(t *testing.T) {
	rules := map[string]string{"format.ndb": "Never:0:*:" + hx(formatCoverageMarker) + "\n"}
	plain := []byte("ordinary harmless LZMA control")
	for _, which := range []string{"short-header", "unknown-properties", "invalid-properties", "dictionary-limit", "truncated-stream", "file-expansion", "scan-expansion", "known-size-limit", "depth", "count"} {
		t.Run(which, func(t *testing.T) {
			s := pipelineState(t, rules)
			payload := plain
			if which == "file-expansion" || which == "scan-expansion" || which == "known-size-limit" {
				payload = bytes.Repeat(plain, 4096)
			}
			eos := which != "known-size-limit"
			packed := zipLZMAPacked(t, payload, eos)
			want := ErrMalformed
			switch which {
			case "short-header":
				packed = packed[:8]
			case "unknown-properties":
				binary.LittleEndian.PutUint16(packed[2:], 6)
				want = ErrUnsupported
			case "invalid-properties":
				packed[4] = 255
			case "dictionary-limit":
				binary.LittleEndian.PutUint32(packed[5:], maxCodecDictionary+1)
				want = ErrLimit
			case "truncated-stream":
				packed = packed[:10]
			case "file-expansion", "known-size-limit":
				s.maxFile = 2048
				want = ErrLimit
			case "scan-expansion":
				s.remaining = 2048
				want = ErrLimit
			case "depth":
				s.maxDepth = 0
				want = ErrLimit
			case "count":
				s.maxFiles = 1
				want = ErrLimit
			}
			if sig, err := s.visit(zipLZMAFixture(t, packed, payload, eos), 0); sig != "" || !errors.Is(err, want) {
				t.Fatalf("scan: %q %v; want %v", sig, err, want)
			}
		})
	}
}
