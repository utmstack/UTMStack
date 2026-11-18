package jsnorm

import (
	"strconv"
	"testing"
)

func TestUnprovedPackerArgumentTypeStaysIncomplete(t *testing.T) {
	input := `function pack(p,a,c,k,e,r){return p} pack('a;','bad',2,''.split('|'),0,{});`
	if _, unknown := NormalizeBlocks([][][]byte{{[]byte(input)}}); !unknown {
		t.Fatal("unproved argument type reported complete")
	}
}

// Native reference controls establish three distinct boundary behaviors. The
// fuzz property varies the pending identifier, exponent suffix and escaped
// string payload while keeping an explicit delimiter that releases each token.
func FuzzDecoderCarry(f *testing.F) {
	f.Add([]byte("boundary"), uint16(3))
	f.Add([]byte{0, 255, 32, 1}, uint16(0))
	f.Fuzz(func(t *testing.T, data []byte, split uint16) {
		if len(data) > 256 {
			t.Skip()
		}
		letters := make([]byte, len(data))
		digits := make([]byte, len(data))
		for i, b := range data {
			letters[i] = 'a' + b%26
			digits[i] = '0' + b%10
		}
		cut := int(split) % (len(letters) + 1)
		ev := func(s string) string { return "eval(" + strconv.Quote(s) + ")" }
		cases := []struct{ source, want string }{
			{ev("x"+string(letters[:cut])) + `;bridge();` + ev(string(letters[cut:])+";") + `;`, `<script>bridge();x` + string(letters) + `;</script>`},
			{ev("1.e") + ev(string(digits)+";") + `;`, `<script>1 e` + string(digits) + `;</script>`},
			{`eval("\"x\u005c");bridge();` + ev(string(letters)+`";`) + `;`, `<script>bridge();"x\` + string(letters) + `";</script>`},
		}
		for _, tc := range cases {
			got, unknown := NormalizeBlocks([][][]byte{{[]byte(tc.source)}})
			if unknown || string(got) != tc.want {
				t.Fatalf("source=%q got=%q unknown=%t want=%q", tc.source, got, unknown, tc.want)
			}
		}
		if len(cases[0].want) > 4096 {
			t.Fatal("unbounded carry output")
		}
	})
}
