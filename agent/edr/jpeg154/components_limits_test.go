package jpeg154

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func componentSeeds(t testing.TB) [][]byte {
	t.Helper()
	root := filepath.Join("testdata", "components")
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Name, Colorspace string
		Accepted         bool
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	var result [][]byte
	for _, r := range rows {
		key := r.Colorspace
		if !r.Accepted {
			key = "rejected"
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		data, err := os.ReadFile(filepath.Join(root, r.Name))
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, data)
	}
	if len(result) < 4 {
		t.Fatalf("component seed classes=%d", len(result))
	}
	for _, name := range []string{"2a90f7439064de0ae1038f24858c38ecb27f5dc42686c4c49fa9ef7028138e18.jpg", "a55b78a299452923455dd7947486fd25365d2ce93b401039c5dc9b0e24b89df0.jpg", "45722323e01296d4af7e31fac6b592102d4c2f4602a61fa4f1936c683358131f.jpg"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, data)
	}
	for _, name := range []string{"planes-0-p1-r0-s0.jpg", "planes-6-p0-r2-s0.jpg", "planes-7-p1-r0-s0.jpg", "planes-9-p0-r0-s0.jpg", "planes-11-p0-r2-s1.jpg", "planes-12-p1-r0-s0.jpg", "planes-4-p0-r0-s0.jpg", "planes-17-p1-r0-s0.jpg"} {
		data, err := os.ReadFile(filepath.Join("testdata", "planes", name))
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, data)
	}
	return result
}

func TestTIFFComponentsLimitsAndCallback(t *testing.T) {
	sentinel := errors.New("component work stopped")
	for _, data := range componentSeeds(t) {
		calls := 0
		_, err := DecodeTIFFComponents(bytes.NewReader(data), func(n int64) error { calls++; return sentinel })
		if calls > 0 && !errors.Is(err, sentinel) {
			t.Fatalf("header work error lost: %v", err)
		}
		calls = 0
		_, err = DecodeTIFFComponents(bytes.NewReader(data), func(n int64) error {
			calls++
			if calls == 3 {
				return sentinel
			}
			return nil
		})
		if calls >= 3 && !errors.Is(err, sentinel) {
			t.Fatalf("scan work error lost: %v", err)
		}
	}
	var data []byte
	at := -1
	for _, seed := range componentSeeds(t) {
		at = bytes.Index(seed, []byte{0xff, 0xc0})
		if at < 0 {
			at = bytes.Index(seed, []byte{0xff, 0xc2})
		}
		if at >= 0 {
			data = append([]byte(nil), seed...)
			break
		}
	}
	if at < 0 {
		t.Fatal("missing SOF seed")
	}
	binary.BigEndian.PutUint16(data[at+5:], 0xffff)
	binary.BigEndian.PutUint16(data[at+7:], 0xffff)
	if _, err := DecodeTIFFComponents(bytes.NewReader(data), nil); err == nil {
		t.Fatal("oversized SOF accepted")
	}
	if _, err := DecodeTIFFConfig(bytes.NewReader(data)); err == nil {
		t.Fatal("oversized SOF config accepted")
	}
	for _, malformed := range [][]byte{{0xff, 0xd8, 0xff, 0xd9}, {0xff, 0xd8, 0xff, 0xda, 0, 2}} {
		if _, err := DecodeTIFFConfig(bytes.NewReader(malformed)); err == nil {
			t.Fatal("missing frame config accepted")
		}
		if _, err := DecodeTIFFComponents(bytes.NewReader(malformed), nil); err == nil {
			t.Fatal("missing frame accepted")
		}
	}
}

func FuzzTIFFComponents(f *testing.F) {
	for _, data := range componentSeeds(f) {
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		budget := int64(1 << 16)
		_, _ = DecodeTIFFComponents(bytes.NewReader(data), func(n int64) error {
			if n < 0 || n > budget {
				return errors.New("work limit")
			}
			budget -= n
			return nil
		})
		if budget < 0 {
			t.Fatal("budget exceeded")
		}
	})
}
