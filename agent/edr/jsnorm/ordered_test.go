package jsnorm

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func TestDecoderRangeWorkRemainsIncomplete(t *testing.T) {
	source := []byte(strings.Repeat(`eval("a;");`, 4097))
	j := newJSNormalizer()
	j.feed(source)
	j.finish()
	if !j.decoderIncomplete || j.decoderTokenWork > maxDecoderTokenWork {
		t.Fatalf("work=%d incomplete=%t", j.decoderTokenWork, j.decoderIncomplete)
	}
}

// Native controls establish that dF appends and skips the following token,
// while an earlier eval is replaced. Varying the encoded identifier checks
// that dictionary-free text cannot change that ordering or leak encoded bytes.
func FuzzDecoderOrder(f *testing.F) {
	f.Add([]byte("sample"))
	f.Add([]byte{0, 255, 32, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			t.Skip()
		}
		name := make([]byte, len(data)+1)
		name[0] = 'x'
		for i, b := range data {
			name[i+1] = 'a' + b%26
		}
		plain := append(name, ';')
		encoded := make([]byte, 0, 3*len(plain))
		for _, b := range plain {
			encoded = append(encoded, '%')
			var pair [2]byte
			hex.Encode(pair[:], []byte{b})
			encoded = append(encoded, pair[:]...)
		}
		call := `dF("` + string(encoded) + `0");`
		for _, tc := range []struct{ source, want string }{
			{`eval("left;");` + call + ` `, `<script>left;` + string(plain) + `</script>`},
			{call + `eval("right;"); `, `<script>eval("right;");` + string(plain) + `</script>`},
		} {
			got, unknown := NormalizeBlocks([][][]byte{{[]byte(tc.source)}})
			if unknown || !bytes.Equal(got, []byte(tc.want)) {
				t.Fatalf("source=%q got=%q unknown=%t want=%q", tc.source, got, unknown, tc.want)
			}
		}
	})
}
