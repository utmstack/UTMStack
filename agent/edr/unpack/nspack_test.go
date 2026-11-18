package unpack

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestNsPackReferenceRebuild(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-26-nspack", "oracle")
	for _, tc := range []struct{ source, reference string }{
		{"nspack-zero.exe", "reference-rebuilt.exe"},
		{"nspack-mutation-32.exe", "reference-mutation-32.exe"},
		{"nspack-mutation-64.exe", "reference-mutation-64.exe"},
		{"nspack-mutation-128.exe", "reference-mutation-128.exe"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			f, err := os.ReadFile(filepath.Join(root, tc.source))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(filepath.Join(root, tc.reference))
			if err != nil {
				t.Fatal(err)
			}
			got, child := NsPack(NsPackProgram{Packed: f[0x800 : 0x800+0x1ff], ExpectedSize: 0x1000,
				SectionRVA: 0x1000, ImageBase: 0x400000, OriginalEntry: 0x1000})
			if got != NsPackUnpacked {
				t.Fatalf("outcome = %d", got)
			}
			if !bytes.Equal(child, want) {
				t.Fatalf("rebuilt child differs from reference: %d vs %d bytes", len(child), len(want))
			}
		})
	}
}

func TestNsPackMalformedStreamDoesNotPanic(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-26-nspack", "oracle")
	base, err := os.ReadFile(filepath.Join(root, "nspack-zero.exe"))
	if err != nil {
		t.Fatal(err)
	}
	packed := base[0x800 : 0x800+0x1ff]
	program := NsPackProgram{ExpectedSize: 0x1000, SectionRVA: 0x1000, ImageBase: 0x400000, OriginalEntry: 0x1000}
	for _, tc := range []struct {
		name   string
		change func(*NsPackProgram)
		want   NsPackOutcome
	}{
		{"short header", func(p *NsPackProgram) { p.Packed = p.Packed[:12] }, NsPackFailed},
		{"short record", func(p *NsPackProgram) { p.Packed = p.Packed[:20] }, NsPackFailed},
		{"wrong output size", func(p *NsPackProgram) { p.ExpectedSize = 0x2000 }, NsPackFailed},
		{"output limit", func(p *NsPackProgram) { p.ExpectedSize = 65 << 20 }, NsPackUnknown},
		{"bad configuration", func(p *NsPackProgram) { p.Packed = bytes.Clone(p.Packed); p.Packed[0] = 0xe1 }, NsPackFailed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := program
			p.Packed = packed
			tc.change(&p)
			got, child := NsPack(p)
			if got != tc.want || child != nil {
				t.Fatalf("outcome %d, child %d bytes; want %d and no child", got, len(child), tc.want)
			}
		})
	}
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 200; i++ {
		p := program
		p.Packed = bytes.Clone(packed)
		for j := 0; j < 3; j++ {
			p.Packed[13+rng.Intn(243)] = byte(rng.Intn(256))
		}
		NsPack(p)
	}
}

func TestNsPackWorkLimit(t *testing.T) {
	r := nsRange{steps: nsMaxCodeSteps, table: []uint16{0x400}, bitmap: ^uint32(0)}
	if r.bit(0) != 0 || !r.bad || !r.limited {
		t.Fatalf("work limit not reported: %+v", r)
	}
}

func TestNsPackMutationOracle(t *testing.T) {
	root := filepath.Join("..", "..", "..", "docs", "edr", "native-engine", "audits", "2026-10-26-nspack", "oracle")
	base, err := os.ReadFile(filepath.Join(root, "nspack-zero.exe"))
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(root, "mutations.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 121 {
		t.Fatalf("oracle rows = %d, want 121 including header", len(rows))
	}
	for _, row := range rows[1:] {
		offset, err := strconv.Atoi(row[0])
		if err != nil {
			t.Fatal(err)
		}
		value, err := strconv.Atoi(row[1])
		if err != nil {
			t.Fatal(err)
		}
		input := bytes.Clone(base)
		input[0x80d+offset] = byte(value)
		outcome, child := NsPack(NsPackProgram{Packed: input[0x800 : 0x800+0x1ff], ExpectedSize: 0x1000,
			SectionRVA: 0x1000, ImageBase: 0x400000, OriginalEntry: 0x1000})
		if row[2] == "-" {
			if outcome == NsPackUnpacked {
				t.Errorf("offset %d, value %d: Go rebuilt, reference failed", offset, value)
			}
			continue
		}
		sum := sha256.Sum256(child)
		if outcome != NsPackUnpacked || hex.EncodeToString(sum[:]) != row[2] {
			t.Errorf("offset %d, value %d: outcome %d, rebuilt SHA256 %x, want %s", offset, value, outcome, sum, row[2])
		}
	}
}
