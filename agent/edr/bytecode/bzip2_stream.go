package bytecode

import "math/bits"

// The bzip2 wire reader retains bit, Huffman, move-to-front, BWT and RLE
// state. Each call exposes the bytes consumed and produced by that call;
// no complete-stream buffering or speculative cursor advancement is used.
// The reference ABI is ClamAV 1.5.4's libbz2-backed pipe adapter. This Go
// implementation is independent of that adapter and has no native dependency.
const (
	bzipOK            = 0
	bzipEnd           = 4
	bzipDataError     = -4
	bzipMagicError    = -5
	bzipSequenceError = -1
	bzipMaxSelectors  = 18002
)

type bzipPhase uint8

const (
	bzHeader bzipPhase = iota
	bzMarker
	bzBlockCRC
	bzRandom
	bzOrigin
	bzRangeMap
	bzSymbolMap
	bzGroups
	bzSelectorCount
	bzSelectors
	bzLengthStart
	bzLengthFlag
	bzLengthDelta
	bzSymbolStart
	bzSymbolBits
	bzSymbolExtra
	bzOutput
	bzFinalCRC
	bzEnded
)

type bzipTable struct {
	min         int
	limit, base [23]int
	perm        [258]uint16
}

func (t *bzipTable) build(lengths []uint8) {
	*t = bzipTable{min: 20}
	maxLen := 0
	for _, n := range lengths {
		t.min = min(t.min, int(n))
		maxLen = max(maxLen, int(n))
		t.base[int(n)+1]++
	}
	k := 0
	for n := t.min; n <= maxLen; n++ {
		for sym, length := range lengths {
			if int(length) == n {
				t.perm[k] = uint16(sym)
				k++
			}
		}
	}
	for i := 1; i < len(t.base); i++ {
		t.base[i] += t.base[i-1]
	}
	code := 0
	for n := t.min; n <= maxLen; n++ {
		code += t.base[n+1] - t.base[n]
		t.limit[n] = code - 1
		code <<= 1
	}
	for n := t.min + 1; n <= maxLen; n++ {
		t.base[n] = ((t.limit[n-1] + 1) << 1) - t.base[n]
	}
}

type streamBzip2 struct {
	phase bzipPhase

	bitValue    uint64
	bitCount    uint
	field       int
	word        uint32
	finalMarker bool

	blockSize, origin int
	tt                []uint32
	counts            [256]int

	rangeMap                                    uint32
	mapIndex, alphabet                          int
	symbols, mtf                                [256]byte
	groups, selectorCount, selectorIndex, unary int
	selectors                                   [bzipMaxSelectors]byte
	selectorMTF                                 [6]byte
	lengths                                     [6][258]uint8
	tables                                      [6]bzipTable

	tree, lengthIndex, length                         int
	groupIndex, groupLeft, activeTree, codeBits, code int
	nblock, run, runPower                             int
	tPos                                              uint32
	used, last, repeat                                int
	outByte                                           byte
	blockCRC, expectedBlockCRC, fileCRC               uint32

	randomized              bool
	randomLeft, randomIndex int
	randomSeen              bool
	failed                  bool
	reserve                 func(int) error
}

func (d *streamBzip2) readBits(input []byte, pos *int, n uint) (uint32, bool) {
	for d.bitCount < n {
		if *pos == len(input) {
			return 0, false
		}
		d.bitValue = d.bitValue<<8 | uint64(input[*pos])
		d.bitCount += 8
		*pos += 1
	}
	d.bitCount -= n
	return uint32(d.bitValue>>d.bitCount) & uint32((uint64(1)<<n)-1), true
}

var bzipCRCTable = func() (table [256]uint32) {
	for i := range table {
		c := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if c&0x80000000 != 0 {
				c = c<<1 ^ 0x04c11db7
			} else {
				c <<= 1
			}
		}
		table[i] = c
	}
	return
}()

// process stops on input starvation, output pressure, stream end or a data
// error. The caller must retain unconsumed input. A malformed stream's first
// error is exposed; re-entry after an error remains unsupported until that
// reference state has independent qualification.
func (d *streamBzip2) process(input, output []byte, charge func(uint64) error) (read, written, status int, err error) {
	if d.failed {
		return 0, 0, 0, ErrUnsupported
	}
	if d.phase == bzEnded {
		return 0, 0, bzipSequenceError, nil
	}
	if charge == nil {
		charge = func(uint64) error { return nil }
	}
	outputMark := 0
	fail := func(code int) (int, int, int, error) { d.failed = true; return read, written, code, nil }
	for {
		switch d.phase {
		case bzHeader:
			v, ok := d.readBits(input, &read, 8)
			if !ok {
				return
			}
			if d.field < 3 {
				if v != uint32("BZh"[d.field]) {
					return fail(bzipMagicError)
				}
				d.field++
			} else {
				if v < '1' || v > '9' {
					return fail(bzipMagicError)
				}
				d.blockSize = int(v-'0') * 100000
				if d.reserve != nil {
					if err = d.reserve(d.blockSize * 4); err != nil {
						return
					}
				}
				d.tt = make([]uint32, d.blockSize)
				d.field, d.phase = 0, bzMarker
			}
		case bzMarker:
			v, ok := d.readBits(input, &read, 8)
			if !ok {
				return
			}
			if d.field == 0 {
				d.finalMarker = v == 0x17
			}
			magic := "\x31\x41\x59\x26\x53\x59"
			if d.finalMarker {
				magic = "\x17\x72\x45\x38\x50\x90"
			}
			if v != uint32(magic[d.field]) {
				return fail(bzipDataError)
			}
			d.field++
			if d.field == 6 {
				d.field, d.word = 0, 0
				d.phase = bzBlockCRC
				if d.finalMarker {
					d.phase = bzFinalCRC
				}
			}
		case bzBlockCRC, bzFinalCRC, bzOrigin:
			v, ok := d.readBits(input, &read, 8)
			if !ok {
				return
			}
			d.word = d.word<<8 | v
			d.field++
			want := 4
			if d.phase == bzOrigin {
				want = 3
			}
			if d.field != want {
				continue
			}
			d.field = 0
			switch d.phase {
			case bzFinalCRC:
				d.phase = bzEnded
				if d.word != d.fileCRC {
					return fail(bzipDataError)
				}
				return read, written, bzipEnd, nil
			case bzBlockCRC:
				d.expectedBlockCRC, d.blockCRC = d.word, ^uint32(0)
				d.phase = bzRandom
			case bzOrigin:
				d.origin = int(d.word)
				if d.origin > d.blockSize+10 {
					return fail(bzipDataError)
				}
				d.phase = bzRangeMap
			}
			d.word = 0
		case bzRandom:
			v, ok := d.readBits(input, &read, 1)
			if !ok {
				return
			}
			d.randomized = v != 0
			d.randomLeft, d.randomIndex = 0, 0
			d.randomSeen = d.randomSeen || d.randomized
			d.phase = bzOrigin
		case bzRangeMap:
			v, ok := d.readBits(input, &read, 16)
			if !ok {
				return
			}
			d.rangeMap = v
			d.mapIndex, d.alphabet = 0, 0
			d.phase = bzSymbolMap
		case bzSymbolMap:
			for d.mapIndex < 256 && d.rangeMap&(1<<uint(15-d.mapIndex/16)) == 0 {
				d.mapIndex = (d.mapIndex/16 + 1) * 16
			}
			if d.mapIndex == 256 {
				if d.alphabet == 0 {
					return fail(bzipDataError)
				}
				copy(d.mtf[:], d.symbols[:d.alphabet])
				d.phase = bzGroups
				continue
			}
			v, ok := d.readBits(input, &read, 1)
			if !ok {
				return
			}
			if v != 0 {
				d.symbols[d.alphabet] = byte(d.mapIndex)
				d.alphabet++
			}
			d.mapIndex++
		case bzGroups:
			v, ok := d.readBits(input, &read, 3)
			if !ok {
				return
			}
			if v < 2 || v > 6 {
				return fail(bzipDataError)
			}
			d.groups = int(v)
			d.phase = bzSelectorCount
		case bzSelectorCount:
			v, ok := d.readBits(input, &read, 15)
			if !ok {
				return
			}
			if v == 0 {
				return fail(bzipDataError)
			}
			d.selectorCount, d.selectorIndex, d.unary = int(v), 0, 0
			for i := range d.selectorMTF {
				d.selectorMTF[i] = byte(i)
			}
			d.phase = bzSelectors
		case bzSelectors:
			v, ok := d.readBits(input, &read, 1)
			if !ok {
				return
			}
			if v != 0 {
				d.unary++
				if d.unary >= d.groups {
					return fail(bzipDataError)
				}
				continue
			}
			if d.selectorIndex < len(d.selectors) {
				chosen := d.selectorMTF[d.unary]
				copy(d.selectorMTF[1:], d.selectorMTF[:d.unary])
				d.selectorMTF[0] = chosen
				d.selectors[d.selectorIndex] = chosen
			}
			d.unary = 0
			d.selectorIndex++
			if d.selectorIndex == d.selectorCount {
				d.selectorCount = min(d.selectorCount, len(d.selectors))
				d.tree, d.lengthIndex, d.phase = 0, 0, bzLengthStart
			}
		case bzLengthStart:
			v, ok := d.readBits(input, &read, 5)
			if !ok {
				return
			}
			d.length = int(v)
			d.phase = bzLengthFlag
		case bzLengthFlag:
			if d.length < 1 || d.length > 20 {
				return fail(bzipDataError)
			}
			v, ok := d.readBits(input, &read, 1)
			if !ok {
				return
			}
			if v != 0 {
				d.phase = bzLengthDelta
				continue
			}
			d.lengths[d.tree][d.lengthIndex] = uint8(d.length)
			d.lengthIndex++
			if d.lengthIndex == d.alphabet+2 {
				d.tree++
				d.lengthIndex = 0
				d.phase = bzLengthStart
				if d.tree == d.groups {
					for i := 0; i < d.groups; i++ {
						d.tables[i].build(d.lengths[i][:d.alphabet+2])
					}
					clear(d.counts[:])
					d.nblock, d.run, d.runPower = 0, 0, 1
					d.groupIndex, d.groupLeft = 0, 0
					d.phase = bzSymbolStart
				}
			}
		case bzLengthDelta:
			v, ok := d.readBits(input, &read, 1)
			if !ok {
				return
			}
			if v == 0 {
				d.length++
			} else {
				d.length--
			}
			d.phase = bzLengthFlag
		case bzSymbolStart:
			if d.groupLeft == 0 {
				if d.groupIndex >= d.selectorCount {
					return fail(bzipDataError)
				}
				d.activeTree = int(d.selectors[d.groupIndex])
				d.groupIndex++
				d.groupLeft = 50
			}
			d.groupLeft--
			d.codeBits = d.tables[d.activeTree].min
			d.phase = bzSymbolBits
		case bzSymbolBits:
			v, ok := d.readBits(input, &read, uint(d.codeBits))
			if !ok {
				return
			}
			d.code = int(v)
			d.phase = bzSymbolExtra
		case bzSymbolExtra:
			table := &d.tables[d.activeTree]
			if d.code > table.limit[d.codeBits] {
				if d.codeBits >= 20 {
					return fail(bzipDataError)
				}
				v, ok := d.readBits(input, &read, 1)
				if !ok {
					return
				}
				d.code = d.code<<1 | int(v)
				d.codeBits++
				continue
			}
			index := d.code - table.base[d.codeBits]
			if index < 0 || index >= d.alphabet+2 {
				return fail(bzipDataError)
			}
			sym := int(table.perm[index])
			d.phase = bzSymbolStart
			if sym < 2 {
				if d.runPower >= 2*1024*1024 {
					return fail(bzipDataError)
				}
				d.run += (sym + 1) * d.runPower
				d.runPower *= 2
				continue
			}
			if d.run != 0 {
				if d.run > d.blockSize-d.nblock {
					return fail(bzipDataError)
				}
				if err = charge(uint64(d.run)); err != nil {
					return
				}
				b := d.mtf[0]
				for i := 0; i < d.run; i++ {
					d.tt[d.nblock+i] = uint32(b)
				}
				d.counts[b] += d.run
				d.nblock += d.run
				d.run, d.runPower = 0, 1
			}
			if sym == d.alphabet+1 {
				if d.origin >= d.nblock {
					return fail(bzipDataError)
				}
				if err = charge(uint64(d.nblock) * 4); err != nil {
					return
				}
				sum := 0
				for i, count := range d.counts {
					d.counts[i] = sum
					sum += count
				}
				for i := 0; i < d.nblock; i++ {
					b := byte(d.tt[i])
					d.tt[d.counts[b]] |= uint32(i) << 8
					d.counts[b]++
				}
				d.tPos = d.tt[d.origin] >> 8
				d.used, d.last, d.repeat = 0, -1, 0
				d.phase = bzOutput
				outputMark = written
				continue
			}
			if sym < 2 || sym > d.alphabet || d.nblock == d.blockSize {
				return fail(bzipDataError)
			}
			b := d.mtf[sym-1]
			copy(d.mtf[1:], d.mtf[:sym-1])
			d.mtf[0] = b
			d.tt[d.nblock] = uint32(b)
			d.nblock++
			d.counts[b]++
		case bzOutput:
			if d.repeat != 0 {
				if written == len(output) {
					return
				}
				b := d.outByte
				output[written] = b
				written++
				d.blockCRC = d.blockCRC<<8 ^ bzipCRCTable[byte(d.blockCRC>>24)^b]
				d.repeat--
				continue
			}
			if d.used > d.nblock+1 {
				// The reference's fast output routine does not publish this
				// call's output cursor when it discovers an invalid RLE tail.
				// Its legacy randomized path instead commits bytes directly.
				if d.randomized {
					if written == len(output) {
						return
					}
				} else {
					written = outputMark
				}
				return fail(bzipDataError)
			}
			if d.used >= d.nblock && d.last == -1 {
				crc := ^d.blockCRC
				if crc != d.expectedBlockCRC {
					return fail(bzipDataError)
				}
				d.fileCRC = bits.RotateLeft32(d.fileCRC, 1) ^ crc
				d.field, d.phase = 0, bzMarker
				continue
			}
			if written == len(output) {
				return
			}
			next := func(required bool) (int, bool) {
				if d.used >= d.nblock && !required {
					return -1, true
				}
				if int(d.tPos) >= d.nblock {
					return -1, false
				}
				v := d.tt[d.tPos]
				d.tPos = v >> 8
				d.used++
				b := byte(v)
				if d.randomized {
					if d.randomLeft == 0 {
						d.randomLeft = bzipRandomNumbers[d.randomIndex]
						d.randomIndex = (d.randomIndex + 1) % len(bzipRandomNumbers)
					}
					d.randomLeft--
					if d.randomLeft == 1 {
						b ^= 1
					}
				}
				return int(b), true
			}
			if d.last == -1 {
				v, ok := next(false)
				if !ok {
					return fail(bzipDataError)
				}
				d.last = v
			}
			d.outByte, d.repeat = byte(d.last), 1
			for d.repeat < 4 {
				v, ok := next(false)
				if !ok {
					return fail(bzipDataError)
				}
				if v != int(d.outByte) {
					d.last = v
					break
				}
				d.repeat++
			}
			if d.repeat == 4 {
				v, ok := next(true)
				if !ok || v < 0 {
					return fail(bzipDataError)
				}
				d.repeat += v
				v, ok = next(true)
				if !ok {
					return fail(bzipDataError)
				}
				d.last = v
				if d.used == d.nblock+1 {
					d.last = -1
				}
			}
		default:
			return 0, 0, 0, ErrRuntime
		}
	}
}
