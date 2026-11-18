package engine

import (
	_ "embed"
	"encoding/binary"
	"sync"
	"unicode/utf8"
)

// Codepage conversion of macro project text to UTF-8 as the Windows
// reference does it. The reference converts with its platform's conversion:
// on Windows a byte or pair without a mapping does not fail, it takes the
// platform's best-fit character ('?', U+30FB, private-use characters); on
// Debian such a byte fails the whole conversion and other tables differ in
// places (strict Shift_JIS for 932, for example). This engine runs on
// Windows, so it follows the Windows reference.
//
// vbacodepages.bin holds the Windows reference's tables, learned from its
// dumps of generated documents (one module per byte value for 38 single-byte
// codepages, one per lead byte for 932, 936, 949 and 950, and the whole
// 0x80-0xFF run; docs/2026-10-08-macro-oracle/build_cptables.py in the
// workspace). Text of the US-ASCII and UTF-8 codepages is copied unchanged
// on both platforms. Codepages 0 and 1 mean the machine's own ANSI and OEM
// codepages, so only ASCII text is decided there; a lead byte not followed by
// a trail byte (0x40-0xFE), and every other codepage, are not decided here.

//go:embed vbacodepages.bin
var vbaCodepageData []byte

type vbaCodepage struct {
	high  [128]uint16 // bytes 0x80-0xFF; 0xffff: a lead byte or unknown
	lead  [256]bool
	pairs map[byte][]uint16 // per lead byte, trails 0x40-0xFE
}

var (
	vbaCodepagesOnce sync.Once
	vbaCodepages     map[uint16]*vbaCodepage
)

func loadVBACodepages() map[uint16]*vbaCodepage {
	vbaCodepagesOnce.Do(func() {
		d := vbaCodepageData
		if len(d) < 6 || string(d[:4]) != "VBCP" {
			panic("engine: invalid codepage table")
		}
		n := int(binary.LittleEndian.Uint16(d[4:]))
		d = d[6:]
		vbaCodepages = make(map[uint16]*vbaCodepage, n)
		u16s := func(count int) []uint16 {
			out := make([]uint16, count)
			for i := range out {
				out[i] = binary.LittleEndian.Uint16(d[i*2:])
			}
			d = d[count*2:]
			return out
		}
		for i := 0; i < n; i++ {
			cp, kind := binary.LittleEndian.Uint16(d), d[2]
			d = d[3:]
			t := &vbaCodepage{}
			copy(t.high[:], u16s(128))
			if kind == 2 {
				bitmap := d[:32]
				d = d[32:]
				t.pairs = make(map[byte][]uint16)
				for b := 0; b < 256; b++ {
					if bitmap[b>>3]&(1<<(b&7)) != 0 {
						t.lead[b] = true
						t.pairs[byte(b)] = u16s(191)
					}
				}
			}
			vbaCodepages[cp] = t
		}
	})
	return vbaCodepages
}

// vbaConvert converts b from codepage cp to UTF-8. ok is false when the
// reference's conversion fails; known is false when this engine cannot tell
// what the reference produces.
func vbaConvert(b []byte, cp uint16) (out []byte, ok, known bool) {
	if len(b) == 0 {
		return nil, true, true
	}
	switch cp {
	case 20127, 65001:
		return b, true, true
	case 0, 1:
		for _, c := range b {
			if c >= 0x80 {
				return nil, false, false
			}
		}
		return b, true, true
	}
	t := loadVBACodepages()[cp]
	if t == nil {
		return nil, false, false
	}
	out = make([]byte, 0, len(b)+len(b)/2)
	for i := 0; i < len(b); i++ {
		c := b[i]
		switch {
		case c < 0x80:
			out = append(out, c)
		case t.lead[c]:
			if i+1 == len(b) || b[i+1] < 0x40 || b[i+1] > 0xfe {
				return nil, false, false
			}
			out = utf8.AppendRune(out, rune(t.pairs[c][b[i+1]-0x40]))
			i++
		case t.high[c-0x80] == 0xffff:
			return nil, false, false
		default:
			out = utf8.AppendRune(out, rune(t.high[c-0x80]))
		}
	}
	return out, true, true
}
