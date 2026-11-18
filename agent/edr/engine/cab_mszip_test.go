package engine

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/mszip/msziptest"
)

func cabMSZIPStreamFixture(t *testing.T, pieces [][]byte, size int) []byte {
	t.Helper()
	b := cabFixture(t, 0, pieces, []cabFixtureFile{{"child.txt", 0, uint32(size)}}, false)
	binary.LittleEndian.PutUint16(b[42:], 1)
	return b
}

func TestCABMSZIPReferenceRepairAndStreaming(t *testing.T) {
	text := []byte("If you can read this, the Quantum decompressor is working!\n")
	var original bytes.Buffer
	w, err := flate.NewWriterDict(&original, flate.BestCompression, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = w.Write(text); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	goFrame := append([]byte("CK"), original.Bytes()...)
	first := bytes.Repeat([]byte("independent history frame "), 1400)[:32768]
	second := append(bytes.Clone(first[len(first)-1024:]), text...)
	window := bytes.Clone(first)
	frames := append(msziptest.Encode(first, nil), msziptest.Encode(second, window)...)
	joined := append(bytes.Clone(first), second...)
	prefix := []byte("available prefix")
	partial := append([]byte{'C', 'K', 0, byte(len(prefix)), 0, byte(^uint16(len(prefix))), 0xff}, prefix...)
	partial = append(partial, 7) // next DEFLATE block has reserved type 3
	partialWant := append(bytes.Clone(prefix), make([]byte, len(text)-len(prefix))...)
	truncated := append([]byte{'C', 'K', 1, 100, 0, 0x9b, 0xff}, prefix...)
	resync := append([]byte("CK\x07"), msziptest.Encode(text, make([]byte, 32768))...)
	dynamicResync := append(bytes.Clone(goFrame), msziptest.Encode(text, make([]byte, 32768))...)
	resyncWant := append(make([]byte, 32768), text...)
	overflow := msziptest.Encode(append(bytes.Repeat([]byte{'A'}, 32768), 'B'), nil)
	overflowWant := append([]byte{'B'}, bytes.Repeat([]byte{'A'}, len(text)-1)...)
	packed := msziptest.Encode(text, nil)
	alphabet := make([]byte, 256)
	for i := range alphabet {
		alphabet[i] = byte(i)
	}
	cases := []struct {
		name   string
		pieces [][]byte
		child  []byte
		found  bool
	}{
		{"incomplete-distance-repaired-child", [][]byte{goFrame}, make([]byte, len(text)), true},
		{"incomplete-distance-original-child", [][]byte{goFrame}, text, false},
		{"reserved-block-type", [][]byte{[]byte("CK\x07")}, make([]byte, len(text)), true},
		{"partial-frame-repair", [][]byte{partial}, partialWant, true},
		{"truncated-stored-repair", [][]byte{truncated}, partialWant, true},
		{"junk-before-CK", [][]byte{append([]byte("junk C C!"), packed...)}, text, true},
		{"CK-crosses-CFDATA", [][]byte{packed[:1], packed[1:]}, text, true},
		{"DEFLATE-crosses-CFDATA", [][]byte{packed[:7], packed[7:]}, text, true},
		{"two-frames-one-CFDATA", [][]byte{frames}, joined, true},
		{"dictionary-across-CFDATA", [][]byte{msziptest.Encode(first, nil), msziptest.Encode(second, window)}, joined, true},
		{"initial-zero-dictionary", [][]byte{msziptest.InitialMatch(258, 1)}, make([]byte, 258), true},
		{"resync-after-repaired-frame", [][]byte{resync}, resyncWant, true},
		{"resync-after-dynamic-repair", [][]byte{dynamicResync}, resyncWant, true},
		{"overlong-frame-window-wrap", [][]byte{overflow}, overflowWant, true},
		{"fixed-literal-alphabet", [][]byte{msziptest.Encode(alphabet, nil)}, alphabet, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parent := cabMSZIPStreamFixture(t, tc.pieces, len(tc.child))
			referenceCABSalvage(t, parent, tc.child, tc.found)
			sig, err := pipelineState(t, cabSalvageRule(tc.child)).visit(parent, 0)
			if tc.found {
				if sig != "Salvage.Child" || err != nil {
					t.Fatalf("MSZIP output: signature=%q error=%v", sig, err)
				}
			} else if sig != "" || err != nil {
				t.Fatalf("wrong repaired child: signature=%q error=%v", sig, err)
			}
		})
	}
}

func TestCABMSZIPReadAheadFailureAndLimits(t *testing.T) {
	child := []byte("harmless complete prefix before a broken input record\n")
	parent := cabMSZIPStreamFixture(t, [][]byte{msziptest.Encode(child, nil)}, len(child))
	binary.LittleEndian.PutUint16(parent[40:], 2)
	parent = append(parent, 0, 0, 0, 0, 100, 0, 100, 0, 1)
	binary.LittleEndian.PutUint32(parent[8:], uint32(len(parent)))
	// The reference's first 4096-byte input call fails as a whole. Bytes
	// copied before that failure never reach the inflater or become a child.
	referenceCABSalvage(t, parent, child, false)
	if sig, err := pipelineState(t, cabSalvageRule(child)).visit(parent, 0); sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("failed read manufactured an available child: %q %v", sig, err)
	}
	full := bytes.Repeat([]byte{'P'}, 32768)
	payload := append([]byte{'C', 'K', 1, 0, 0x80, 0xff, 0x7f}, full...)
	partial := cabMSZIPStreamFixture(t, [][]byte{payload}, len(full))
	binary.LittleEndian.PutUint16(partial[40:], 2)
	partial = append(partial, 0, 0, 0, 0, 100, 0, 100, 0, 1)
	binary.LittleEndian.PutUint32(partial[8:], uint32(len(partial)))
	want := append(bytes.Repeat([]byte{'P'}, 32761), make([]byte, 7)...)
	referenceCABSalvage(t, partial, want, true)
	if sig, err := pipelineState(t, cabSalvageRule(want)).visit(partial, 0); sig != "Salvage.Child" || err != nil {
		t.Fatalf("failed buffered read lost repaired partial frame: %q %v", sig, err)
	}
	parent = cabMSZIPStreamFixture(t, [][]byte{[]byte("CK\x07")}, len(child))
	s := pipelineState(t, cabSalvageRule(make([]byte, len(child))))
	s.remaining = int64(len(parent)) // no output budget; do not synthesize a zero child
	if sig, err := s.visit(parent, 0); sig != "" || !errors.Is(err, ErrLimit) {
		t.Fatalf("repair bypassed output budget: %q %v", sig, err)
	}
	s = pipelineState(t, cabSalvageRule(child))
	s.remaining = int64(len(parent) + len(child))
	input := 1
	out, err := s.cabFolder(parent, cabFolder{start: int(binary.LittleEndian.Uint32(parent[36:])), blocks: 1, compression: 1}, 0, &input, uint64(len(child)), nil)
	if len(out) != 0 || !errors.Is(err, ErrLimit) {
		t.Fatalf("input limit became repaired output: %d %v", len(out), err)
	}
}

func TestCABMSZIPLengthAndDistanceSlots(t *testing.T) {
	lengthBase := []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258}
	lengthExtra := []uint{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0}
	distanceBase := []int{1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513, 769, 1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577}
	distanceExtra := []uint{0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13}
	for _, group := range []struct {
		name  string
		base  []int
		extra []uint
	}{{"length", lengthBase, lengthExtra}, {"distance", distanceBase, distanceExtra}} {
		seen := make(map[int]bool)
		for i, base := range group.base {
			for _, value := range []int{base, base + (1 << group.extra[i]) - 1} {
				if value > 32768 || group.name == "length" && value > 258 || seen[value] {
					continue
				}
				seen[value] = true
				t.Run(fmt.Sprintf("%s-%d", group.name, value), func(t *testing.T) {
					n, d := 258, value
					if group.name == "length" {
						n, d = value, 32768
					}
					suffix := []byte("tail!")
					child := append(make([]byte, n), suffix...)
					parent := cabMSZIPStreamFixture(t, [][]byte{msziptest.InitialMatchWithSuffix(n, d, suffix)}, len(child))
					referenceCABSalvage(t, parent, child, true)
					if sig, err := pipelineState(t, cabSalvageRule(child)).visit(parent, 0); sig != "Salvage.Child" || err != nil {
						t.Fatalf("match slot: %q %v", sig, err)
					}
				})
			}
		}
	}
}

func TestCABMSZIPIndependentMember(t *testing.T) {
	parent := cabQuantumControl(t)
	child := []byte("If you can read this, the MSZIP decompressor is working!\n")
	referenceCABSalvage(t, parent, child, true)
	if sig, err := pipelineState(t, cabSalvageRule(child)).visit(parent, 0); sig != "Salvage.Child" || err != nil {
		t.Fatalf("independent MSZIP child: %q %v", sig, err)
	}
}

func TestCABMSZIPFailedSeekDoesNotInventChild(t *testing.T) {
	prefix := []byte("available prefix")
	packed := append([]byte{'C', 'K', 1, 100, 0, 0x9b, 0xff}, prefix...)
	frame := append(bytes.Clone(prefix), make([]byte, 100)...)
	for _, offset := range []int{0, 1, 8, 16} {
		t.Run(fmt.Sprintf("offset-%d", offset), func(t *testing.T) {
			child := frame[offset : offset+59]
			b := cabMSZIPStreamFixture(t, [][]byte{packed}, len(child))
			binary.LittleEndian.PutUint32(b[48:], uint32(offset))
			referenceCABSalvage(t, b, child, offset == 0)
			sig, err := pipelineState(t, cabSalvageRule(child)).visit(b, 0)
			if offset == 0 {
				if sig != "Salvage.Child" || err != nil {
					t.Fatalf("fatal repaired request: %q %v", sig, err)
				}
			} else if sig != "" || !errors.Is(err, ErrMalformed) {
				t.Fatalf("failed seek manufactured a member: %q %v", sig, err)
			}
		})
	}
}

func TestCABMSZIPZlibGoldenOracle(t *testing.T) {
	data, err := os.ReadFile("testdata/mszip/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		File     string `json:"file"`
		SHA      string `json:"input_sha256"`
		ChildSHA string `json:"child_sha256"`
		Size     int    `json:"child_bytes"`
	}
	if err = json.Unmarshal(data, &cases); err != nil || len(cases) != 44 {
		t.Fatalf("oracle manifest: %d cases, %v", len(cases), err)
	}
	for _, tc := range cases {
		t.Run(tc.File, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("testdata/mszip", tc.File))
			if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != tc.SHA {
				t.Fatalf("changed oracle: %v", err)
			}
			rules := map[string]string{"child.hsb": fmt.Sprintf("%s:%d:MSZIP.Oracle\n", tc.ChildSHA, tc.Size)}
			referenceCABMetadata(t, b, rules, "MSZIP.Oracle")
			if sig, err := pipelineState(t, rules).visit(b, 0); sig != "MSZIP.Oracle" || err != nil {
				t.Fatalf("native golden child: signature=%q error=%v", sig, err)
			}
		})
	}
}
