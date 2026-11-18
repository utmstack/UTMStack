package sevenzip

import "encoding/binary"

// PPMd variant H as 7-Zip stores it (with the 7-Zip range coder), after
// Dmitry Shkarin's public-domain PPMd var.H and the public-domain LZMA SDK
// (Ppmd7.c, Ppmd7Dec.c). A decoder only decodes what the encoder's model
// predicted, so the model is reproduced exactly: the 12-byte unit
// sub-allocator with its free lists, gluing and restart when memory runs
// out, the context tree with raw and record successors, the SEE contexts and
// the binary contexts. The model's memory is addressed by offsets, as the
// reference code does on 64-bit hosts.

const (
	ppmdMinOrder   = 2
	ppmdMaxOrder   = 64
	ppmdMinMem     = 1 << 11
	ppmdMaxMem     = 0xffffffff - 12*3
	ppmdNumIndexes = 38
	ppmdUnitSize   = 12
	ppmdMaxFreq    = 124
	ppmdIntBits    = 7
	ppmdPeriodBits = 7
	ppmdBinScale   = 1 << (ppmdIntBits + ppmdPeriodBits)
	ppmdTop        = 1 << 24
	// ppmdMaxModel bounds the model memory one decoding allocates (7-Zip's
	// largest preset is 192 MiB).
	ppmdMaxModel = 256 << 20
)

var (
	ppmdExpEscape  = [16]byte{25, 14, 9, 7, 5, 5, 4, 4, 4, 3, 3, 3, 2, 2, 2, 2}
	ppmdInitBinEsc = [8]uint16{0x3cdd, 0x1f3f, 0x59bf, 0x48f3, 0x64a1, 0x5abc, 0x6632, 0x6051}
)

// ppmdMem is the model's memory.
type ppmdMem []byte

func (m ppmdMem) u8(off uint32) byte         { return m[off] }
func (m ppmdMem) set8(off uint32, v byte)    { m[off] = v }
func (m ppmdMem) u16(off uint32) uint16      { return binary.LittleEndian.Uint16(m[off:]) }
func (m ppmdMem) set16(off uint32, v uint16) { binary.LittleEndian.PutUint16(m[off:], v) }
func (m ppmdMem) u32(off uint32) uint32      { return binary.LittleEndian.Uint32(m[off:]) }
func (m ppmdMem) set32(off uint32, v uint32) { binary.LittleEndian.PutUint32(m[off:], v) }
func (m ppmdMem) copyUnits(dst, src, n uint32) {
	copy(m[dst:dst+n*ppmdUnitSize], m[src:src+n*ppmdUnitSize])
}

type ppmdSee struct {
	summ  uint16
	shift byte
	count byte
}

func (s *ppmdSee) update() {
	if s.shift < ppmdPeriodBits {
		if s.count--; s.count == 0 {
			s.summ <<= 1
			s.count = byte(3 << s.shift)
			s.shift++
		}
	}
}

// ppmd7 is the model and the range decoder. A context (12 bytes) holds
// NumStats (16 bits), SummFreq (16 bits) or, with one symbol, the state
// itself from byte 2, the offset of its states and its suffix context. A
// state (6 bytes) holds the symbol, its frequency and its successor.
type ppmd7 struct {
	m                                  ppmdMem
	minContext, maxContext, foundState uint32
	orderFall, initEsc, prevSuccess    uint32
	maxOrder, hiBitsFlag               uint32
	runLength, initRL                  int32
	size, glueCount, alignOffset       uint32
	loUnit, hiUnit, text, unitsStart   uint32
	indx2Units                         [ppmdNumIndexes]byte
	units2Indx                         [128]byte
	freeList                           [ppmdNumIndexes]uint32
	ns2BSIndx, ns2Indx                 [256]byte
	dummySee                           ppmdSee
	see                                [25][16]ppmdSee
	binSumm                            [128][64]uint16

	// mask holds escCount for the symbols an escape excluded.
	mask     [256]byte
	escCount byte

	src       []byte
	pos       int
	rng, code uint32
	extra     bool
}

func (p *ppmd7) numStats(c uint32) uint32  { return uint32(p.m.u16(c)) }
func (p *ppmd7) setNumStats(c, v uint32)   { p.m.set16(c, uint16(v)) }
func (p *ppmd7) summFreq(c uint32) uint32  { return uint32(p.m.u16(c + 2)) }
func (p *ppmd7) setSummFreq(c, v uint32)   { p.m.set16(c+2, uint16(v)) }
func (p *ppmd7) stats(c uint32) uint32     { return p.m.u32(c + 4) }
func (p *ppmd7) setStats(c, v uint32)      { p.m.set32(c+4, v) }
func (p *ppmd7) suffix(c uint32) uint32    { return p.m.u32(c + 8) }
func (p *ppmd7) setSuffix(c, v uint32)     { p.m.set32(c+8, v) }
func oneState(c uint32) uint32             { return c + 2 }
func (p *ppmd7) sym(s uint32) uint32       { return uint32(p.m.u8(s)) }
func (p *ppmd7) setSym(s, v uint32)        { p.m.set8(s, byte(v)) }
func (p *ppmd7) freq(s uint32) uint32      { return uint32(p.m.u8(s + 1)) }
func (p *ppmd7) setFreq(s, v uint32)       { p.m.set8(s+1, byte(v)) }
func (p *ppmd7) successor(s uint32) uint32 { return p.m.u32(s + 2) }
func (p *ppmd7) setSuccessor(s, v uint32)  { p.m.set32(s+2, v) }
func (p *ppmd7) i2u(indx uint32) uint32    { return uint32(p.indx2Units[indx]) }
func (p *ppmd7) u2i(nu uint32) uint32      { return uint32(p.units2Indx[nu-1]) }
func (p *ppmd7) copyState(dst, src uint32) {
	p.m.set16(dst, p.m.u16(src))
	p.m.set32(dst+2, p.m.u32(src+2))
}
func hiBitsFlag3(sym uint32) uint32 { return (sym + 0xc0) >> 5 & 8 }
func hiBitsFlag4(sym uint32) uint32 { return (sym + 0xc0) >> 4 & 16 }
func (p *ppmd7) swapStates(s uint32) {
	a0, a1 := p.m.u16(s), p.m.u32(s+2)
	p.copyState(s, s-6)
	p.m.set16(s-6, a0)
	p.m.set32(s-4, a1)
}

func newPPMd7(order, size uint32) *ppmd7 {
	p := &ppmd7{maxOrder: order, size: size, alignOffset: (4 - size) & 3}
	k := 0
	for i := 0; i < ppmdNumIndexes; i++ {
		step := 4
		if i < 12 {
			step = i>>2 + 1
		}
		for ; step > 0; step-- {
			p.units2Indx[k] = byte(i)
			k++
		}
		p.indx2Units[i] = byte(k)
	}
	p.ns2BSIndx[1] = 2
	for i := 2; i < 11; i++ {
		p.ns2BSIndx[i] = 4
	}
	for i := 11; i < 256; i++ {
		p.ns2BSIndx[i] = 6
	}
	for i := 0; i < 3; i++ {
		p.ns2Indx[i] = byte(i)
	}
	m, n := 3, 1
	for i := 3; i < 256; i++ {
		p.ns2Indx[i] = byte(m)
		if n--; n == 0 {
			m++
			n = m - 2
		}
	}
	p.m = make(ppmdMem, uint64(p.alignOffset)+uint64(size))
	p.restart()
	return p
}

func (p *ppmd7) insertNode(node, indx uint32) {
	p.m.set32(node, p.freeList[indx])
	p.freeList[indx] = node
}

func (p *ppmd7) removeNode(indx uint32) uint32 {
	node := p.freeList[indx]
	p.freeList[indx] = p.m.u32(node)
	return node
}

func (p *ppmd7) splitBlock(ptr, oldIndx, newIndx uint32) {
	nu := p.i2u(oldIndx) - p.i2u(newIndx)
	ptr += p.i2u(newIndx) * ppmdUnitSize
	i := p.u2i(nu)
	if p.i2u(i) != nu {
		i--
		k := p.i2u(i)
		p.insertNode(ptr+k*ppmdUnitSize, nu-k-1)
	}
	p.insertNode(ptr, i)
}

// glueFreeBlocks joins adjacent free blocks. A free node holds a zero stamp
// (where a context or state holds a nonzero count or frequency), its units
// and the next node.
func (p *ppmd7) glueFreeBlocks() {
	p.glueCount = 255
	if p.loUnit != p.hiUnit {
		p.m.set16(p.loUnit, 1) // the guard
	}
	n := uint32(0)
	for i := 0; i < ppmdNumIndexes; i++ {
		nu := uint16(p.indx2Units[i])
		next := p.freeList[i]
		p.freeList[i] = 0
		for next != 0 {
			at := next
			next = p.m.u32(at)
			p.m.set16(at, 0)
			p.m.set16(at+2, nu)
			p.m.set32(at+4, n)
			n = at
		}
	}
	head := n
	prev := uint32(0) // the node whose next field links on; 0 is the head
	for n != 0 {
		node := n
		nu := uint32(p.m.u16(node + 2))
		n = p.m.u32(node + 4)
		if nu == 0 {
			if prev == 0 {
				head = n
			} else {
				p.m.set32(prev+4, n)
			}
			continue
		}
		prev = node
		for {
			node2 := node + nu*ppmdUnitSize
			nu += uint32(p.m.u16(node2 + 2))
			if p.m.u16(node2) != 0 || nu >= 0x10000 {
				break
			}
			p.m.set16(node+2, uint16(nu))
			p.m.set16(node2+2, 0)
		}
	}
	for n = head; n != 0; {
		node := n
		nu := uint32(p.m.u16(node + 2))
		n = p.m.u32(node + 4)
		if nu == 0 {
			continue
		}
		for ; nu > 128; nu, node = nu-128, node+128*ppmdUnitSize {
			p.insertNode(node, ppmdNumIndexes-1)
		}
		i := p.u2i(nu)
		if p.i2u(i) != nu {
			i--
			k := p.i2u(i)
			p.insertNode(node+k*ppmdUnitSize, nu-k-1)
		}
		p.insertNode(node, i)
	}
}

func (p *ppmd7) allocUnitsRare(indx uint32) uint32 {
	if p.glueCount == 0 {
		p.glueFreeBlocks()
		if p.freeList[indx] != 0 {
			return p.removeNode(indx)
		}
	}
	i := indx
	for {
		if i++; i == ppmdNumIndexes {
			numBytes := p.i2u(indx) * ppmdUnitSize
			us := p.unitsStart
			p.glueCount--
			if us-p.text > numBytes {
				p.unitsStart = us - numBytes
				return p.unitsStart
			}
			return 0
		}
		if p.freeList[i] != 0 {
			break
		}
	}
	block := p.removeNode(i)
	p.splitBlock(block, i, indx)
	return block
}

func (p *ppmd7) allocUnits(indx uint32) uint32 {
	if p.freeList[indx] != 0 {
		return p.removeNode(indx)
	}
	numBytes := p.i2u(indx) * ppmdUnitSize
	if lo := p.loUnit; p.hiUnit-lo >= numBytes {
		p.loUnit = lo + numBytes
		return lo
	}
	return p.allocUnitsRare(indx)
}

// restart sets the model to its start: the order-0 context with all 256
// symbols at frequency 1, the memory free.
func (p *ppmd7) restart() {
	p.freeList = [ppmdNumIndexes]uint32{}
	p.text = p.alignOffset
	p.hiUnit = p.text + p.size
	p.loUnit = p.hiUnit - p.size/8/ppmdUnitSize*7*ppmdUnitSize
	p.unitsStart = p.loUnit
	p.glueCount = 0
	p.orderFall = p.maxOrder
	p.initRL = -int32(min(p.maxOrder, 12)) - 1
	p.runLength = p.initRL
	p.prevSuccess = 0

	p.hiUnit -= ppmdUnitSize
	mc, s := p.hiUnit, p.loUnit
	p.loUnit += 128 * ppmdUnitSize
	p.maxContext, p.minContext, p.foundState = mc, mc, s
	p.setNumStats(mc, 256)
	p.setSummFreq(mc, 257)
	p.setStats(mc, s)
	p.setSuffix(mc, 0)
	for i := uint32(0); i < 256; i++ {
		p.setSym(s+i*6, i)
		p.setFreq(s+i*6, 1)
		p.setSuccessor(s+i*6, 0)
	}
	for i := range p.binSumm {
		for k := 0; k < 8; k++ {
			v := uint16(ppmdBinScale - uint32(ppmdInitBinEsc[k])/uint32(i+2))
			for m := 0; m < 64; m += 8 {
				p.binSumm[i][k+m] = v
			}
		}
	}
	for i := range p.see {
		for k := range p.see[i] {
			p.see[i][k] = ppmdSee{summ: uint16((5*i + 10) << (ppmdPeriodBits - 4)), shift: ppmdPeriodBits - 4, count: 4}
		}
	}
	p.dummySee = ppmdSee{shift: ppmdPeriodBits, count: 64}
}

// createSuccessors turns the raw successor of the found state into context
// records; 0 when memory runs out.
func (p *ppmd7) createSuccessors() uint32 {
	c := p.minContext
	upBranch := p.successor(p.foundState)
	var ps [ppmdMaxOrder]uint32
	numPs := 0
	if p.orderFall != 0 {
		ps[numPs] = p.foundState
		numPs++
	}
	for p.suffix(c) != 0 {
		c = p.suffix(c)
		var s uint32
		if p.numStats(c) != 1 {
			sym := p.sym(p.foundState)
			for s = p.stats(c); p.sym(s) != sym; s += 6 {
			}
		} else {
			s = oneState(c)
		}
		if succ := p.successor(s); succ != upBranch {
			c = succ
			if numPs == 0 {
				return c
			}
			break
		}
		ps[numPs] = s
		numPs++
	}
	newSym := uint32(p.m.u8(upBranch))
	upBranch++
	var newFreq uint32
	if p.numStats(c) == 1 {
		newFreq = p.freq(oneState(c))
	} else {
		s := p.stats(c)
		for p.sym(s) != newSym {
			s += 6
		}
		cf := p.freq(s) - 1
		s0 := p.summFreq(c) - p.numStats(c) - cf
		if 2*cf <= s0 {
			newFreq = 1
			if 5*cf > s0 {
				newFreq = 2
			}
		} else {
			newFreq = 1 + (2*cf+s0-1)/(2*s0) + 1
		}
		newFreq &= 0xff
	}
	for {
		var c1 uint32
		switch {
		case p.hiUnit != p.loUnit:
			p.hiUnit -= ppmdUnitSize
			c1 = p.hiUnit
		case p.freeList[0] != 0:
			c1 = p.removeNode(0)
		default:
			if c1 = p.allocUnitsRare(0); c1 == 0 {
				return 0
			}
		}
		p.setNumStats(c1, 1)
		p.setSym(oneState(c1), newSym)
		p.setFreq(oneState(c1), newFreq)
		p.setSuccessor(oneState(c1), upBranch)
		p.setSuffix(c1, c)
		numPs--
		p.setSuccessor(ps[numPs], c1)
		c = c1
		if numPs == 0 {
			return c
		}
	}
}

func (p *ppmd7) updateModel() {
	fs := p.foundState
	if p.freq(fs) < ppmdMaxFreq/4 && p.suffix(p.minContext) != 0 {
		c := p.suffix(p.minContext)
		if p.numStats(c) == 1 {
			if s := oneState(c); p.freq(s) < 32 {
				p.setFreq(s, p.freq(s)+1)
			}
		} else {
			s := p.stats(c)
			sym := p.sym(fs)
			if p.sym(s) != sym {
				for {
					s += 6
					if p.sym(s) == sym {
						break
					}
				}
				if p.freq(s) >= p.freq(s-6) {
					p.swapStates(s)
					s -= 6
				}
			}
			if p.freq(s) < ppmdMaxFreq-9 {
				p.setFreq(s, p.freq(s)+2)
				p.setSummFreq(c, p.summFreq(c)+2)
			}
		}
	}

	if p.orderFall == 0 {
		p.minContext = p.createSuccessors()
		p.maxContext = p.minContext
		if p.minContext == 0 {
			p.restart()
			return
		}
		p.setSuccessor(p.foundState, p.minContext)
		return
	}

	p.m.set8(p.text, byte(p.sym(fs)))
	p.text++
	if p.text >= p.unitsStart {
		p.restart()
		return
	}
	maxSuccessor := p.text
	minSuccessor := p.successor(fs)
	if minSuccessor != 0 {
		if minSuccessor <= maxSuccessor {
			cs := p.createSuccessors()
			if cs == 0 {
				p.restart()
				return
			}
			minSuccessor = cs
		}
		if p.orderFall--; p.orderFall == 0 {
			maxSuccessor = minSuccessor
			if p.maxContext != p.minContext {
				p.text--
			}
		}
	} else {
		p.setSuccessor(fs, maxSuccessor)
		minSuccessor = p.minContext
	}

	mc, c := p.minContext, p.maxContext
	p.minContext, p.maxContext = minSuccessor, minSuccessor
	if c == mc {
		return
	}
	ns := p.numStats(mc)
	s0 := p.summFreq(mc) - ns - (p.freq(fs) - 1)
	for {
		ns1 := p.numStats(c)
		var sum uint32
		if ns1 != 1 {
			if ns1&1 == 0 {
				oldNU := ns1 >> 1
				i := p.u2i(oldNU)
				if i != p.u2i(oldNU+1) {
					ptr := p.allocUnits(i + 1)
					if ptr == 0 {
						p.restart()
						return
					}
					old := p.stats(c)
					p.m.copyUnits(ptr, old, oldNU)
					p.insertNode(old, i)
					p.setStats(c, ptr)
				}
			}
			sum = p.summFreq(c)
			var a, b uint32
			if 2*ns1 < ns {
				a = 1
			}
			if 4*ns1 <= ns && sum <= 8*ns1 {
				b = 2
			}
			sum += a + b
		} else {
			s := p.allocUnits(0)
			if s == 0 {
				p.restart()
				return
			}
			freq := p.freq(oneState(c))
			p.copyState(s, oneState(c))
			p.setStats(c, s)
			if freq < ppmdMaxFreq/4-1 {
				freq <<= 1
			} else {
				freq = ppmdMaxFreq - 4
			}
			p.setFreq(s, freq)
			sum = freq + p.initEsc
			if ns > 3 {
				sum++
			}
		}
		s := p.stats(c) + ns1*6
		cf := 2 * (sum + 6) * p.freq(fs)
		sf := s0 + sum
		p.setSym(s, p.sym(fs))
		p.setNumStats(c, ns1+1)
		p.setSuccessor(s, maxSuccessor)
		if cf < 6*sf {
			n := uint32(1)
			if cf > sf {
				n++
			}
			if cf >= 4*sf {
				n++
			}
			cf = n
			sum += 3
		} else {
			n := uint32(4)
			for _, k := range []uint32{9, 12, 15} {
				if cf >= k*sf {
					n++
				}
			}
			cf = n
			sum += cf
		}
		p.setSummFreq(c, sum)
		p.setFreq(s, cf)
		if c = p.suffix(c); c == mc {
			return
		}
	}
}

func (p *ppmd7) rescale() {
	mc := p.minContext
	stats := p.stats(mc)
	s := p.foundState
	if s != stats {
		a0, a1 := p.m.u16(s), p.m.u32(s+2)
		for ; s != stats; s -= 6 {
			p.copyState(s, s-6)
		}
		p.m.set16(s, a0)
		p.m.set32(s+2, a1)
	}
	sumFreq := p.freq(s)
	escFreq := p.summFreq(mc) - sumFreq
	adder := uint32(0)
	if p.orderFall != 0 {
		adder = 1
	}
	sumFreq = (sumFreq + 4 + adder) >> 1
	i := p.numStats(mc) - 1
	p.setFreq(s, sumFreq)
	for {
		s += 6
		freq := p.freq(s)
		escFreq -= freq
		freq = (freq + adder) >> 1
		sumFreq += freq
		p.setFreq(s, freq)
		if freq > p.freq(s-6) {
			a0, a1 := p.m.u16(s), p.m.u32(s+2)
			s1 := s
			for {
				p.copyState(s1, s1-6)
				s1 -= 6
				if s1 == stats || freq <= p.freq(s1-6) {
					break
				}
			}
			p.m.set16(s1, a0)
			p.m.set32(s1+2, a1)
		}
		if i--; i == 0 {
			break
		}
	}
	if p.freq(s) == 0 {
		i = 0
		for {
			i++
			s -= 6
			if p.freq(s) != 0 {
				break
			}
		}
		escFreq += i
		numStats := p.numStats(mc)
		numStatsNew := numStats - i
		p.setNumStats(mc, numStatsNew)
		n0 := (numStats + 1) >> 1
		if numStatsNew == 1 {
			freq := p.freq(stats)
			for {
				escFreq >>= 1
				freq = (freq + 1) >> 1
				if escFreq <= 1 {
					break
				}
			}
			one := oneState(mc)
			p.copyState(one, stats)
			p.setFreq(one, freq)
			p.foundState = one
			p.insertNode(stats, p.u2i(n0))
			return
		}
		n1 := (numStatsNew + 1) >> 1
		if n0 != n1 {
			i0, i1 := p.u2i(n0), p.u2i(n1)
			if i0 != i1 {
				if p.freeList[i1] != 0 {
					ptr := p.removeNode(i1)
					p.setStats(mc, ptr)
					p.m.copyUnits(ptr, stats, n1)
					p.insertNode(stats, i0)
				} else {
					p.splitBlock(stats, i0, i1)
				}
			}
		}
	}
	p.setSummFreq(mc, sumFreq+escFreq-escFreq>>1)
	p.foundState = p.stats(mc)
}

func (p *ppmd7) makeEscFreq(numMasked uint32) (*ppmdSee, uint32) {
	mc := p.minContext
	numStats := p.numStats(mc)
	if numStats == 256 {
		return &p.dummySee, 1
	}
	nonMasked := numStats - numMasked
	k := p.hiBitsFlag
	if nonMasked < p.numStats(p.suffix(mc))-numStats {
		k++
	}
	if p.summFreq(mc) < 11*numStats {
		k += 2
	}
	if numMasked > nonMasked {
		k += 4
	}
	see := &p.see[p.ns2Indx[nonMasked-1]][k]
	r := uint32(see.summ >> see.shift)
	see.summ -= uint16(r)
	if r == 0 {
		r = 1
	}
	return see, r
}

func (p *ppmd7) nextContext() {
	c := p.successor(p.foundState)
	if p.orderFall == 0 && c > p.text {
		p.minContext, p.maxContext = c, c
	} else {
		p.updateModel()
	}
}

func (p *ppmd7) update1() {
	s := p.foundState
	freq := p.freq(s) + 4
	p.setSummFreq(p.minContext, p.summFreq(p.minContext)+4)
	p.setFreq(s, freq)
	if freq > p.freq(s-6) {
		p.swapStates(s)
		p.foundState = s - 6
		if freq > ppmdMaxFreq {
			p.rescale()
		}
	}
	p.nextContext()
}

func (p *ppmd7) update1First() {
	s, mc := p.foundState, p.minContext
	freq, summ := p.freq(s), p.summFreq(mc)
	p.prevSuccess = 0
	if 2*freq > summ {
		p.prevSuccess = 1
	}
	p.runLength += int32(p.prevSuccess)
	p.setSummFreq(mc, summ+4)
	freq += 4
	p.setFreq(s, freq)
	if freq > ppmdMaxFreq {
		p.rescale()
	}
	p.nextContext()
}

func (p *ppmd7) update2() {
	s := p.foundState
	freq := p.freq(s) + 4
	p.runLength = p.initRL
	p.setSummFreq(p.minContext, p.summFreq(p.minContext)+4)
	p.setFreq(s, freq)
	if freq > ppmdMaxFreq {
		p.rescale()
	}
	p.updateModel()
}

// newMask starts an empty set of excluded symbols.
func (p *ppmd7) newMask() {
	if p.escCount++; p.escCount == 0 {
		p.mask = [256]byte{}
		p.escCount = 1
	}
}

func (p *ppmd7) next() uint32 {
	if p.pos >= len(p.src) {
		p.extra = true
		return 0
	}
	b := p.src[p.pos]
	p.pos++
	return uint32(b)
}

func (p *ppmd7) rcInit(src []byte) bool {
	p.src, p.pos, p.code, p.rng = src, 0, 0, 0xffffffff
	if p.next() != 0 {
		return false
	}
	for i := 0; i < 4; i++ {
		p.code = p.code<<8 | p.next()
	}
	return p.code < 0xffffffff && !p.extra
}

func (p *ppmd7) normalize() {
	for i := 0; i < 2 && p.rng < ppmdTop; i++ {
		p.code = p.code<<8 | p.next()
		p.rng <<= 8
	}
}

func (p *ppmd7) rcDecode(start, size uint32) {
	p.code -= start * p.rng
	p.rng *= size
}

func (p *ppmd7) threshold(total uint32) uint32 {
	p.rng /= total
	return p.code / p.rng
}

// decodeSymbol returns the next byte, -1 at the end marker and -2 on a data
// error.
func (p *ppmd7) decodeSymbol() int {
	if p.numStats(p.minContext) != 1 {
		s := p.stats(p.minContext)
		summFreq := p.summFreq(p.minContext)
		count := p.threshold(summFreq)
		hiCnt := count
		if count -= p.freq(s); int32(count) < 0 {
			p.rcDecode(0, p.freq(s))
			p.normalize()
			p.foundState = s
			sym := p.sym(s)
			p.update1First()
			return int(sym)
		}
		p.prevSuccess = 0
		for i := p.numStats(p.minContext) - 1; i > 0; i-- {
			s += 6
			if count -= p.freq(s); int32(count) < 0 {
				p.rcDecode(hiCnt-count-p.freq(s), p.freq(s))
				p.normalize()
				p.foundState = s
				sym := p.sym(s)
				p.update1()
				return int(sym)
			}
		}
		if hiCnt >= summFreq {
			return -2
		}
		hiCnt -= count
		p.rcDecode(hiCnt, summFreq-hiCnt)
		p.hiBitsFlag = hiBitsFlag3(p.sym(p.foundState))
		p.newMask()
		for s2 := p.stats(p.minContext); s2 <= s; s2 += 6 {
			p.mask[p.sym(s2)] = p.escCount
		}
	} else {
		s := oneState(p.minContext)
		p.hiBitsFlag = hiBitsFlag3(p.sym(p.foundState))
		row := p.freq(s) - 1
		col := p.prevSuccess + uint32(p.runLength>>26)&0x20 +
			uint32(p.ns2BSIndx[p.numStats(p.suffix(p.minContext))-1]) + hiBitsFlag4(p.sym(s)) + p.hiBitsFlag
		prob := &p.binSumm[row][col]
		pr := uint32(*prob)
		size0 := (p.rng >> 14) * pr
		pr -= (pr + 1<<(ppmdPeriodBits-2)) >> ppmdPeriodBits
		if p.code < size0 {
			*prob = uint16(pr + 1<<ppmdIntBits)
			p.rng = size0
			p.normalize()
			freq := p.freq(s)
			c := p.successor(s)
			sym := p.sym(s)
			p.foundState = s
			p.prevSuccess = 1
			p.runLength++
			if freq < 128 {
				freq++
			}
			p.setFreq(s, freq)
			if p.orderFall == 0 && c > p.text {
				p.minContext, p.maxContext = c, c
			} else {
				p.updateModel()
			}
			return int(sym)
		}
		*prob = uint16(pr)
		p.initEsc = uint32(ppmdExpEscape[pr>>10])
		p.code -= size0
		p.rng -= size0
		p.newMask()
		p.mask[p.sym(s)] = p.escCount
		p.prevSuccess = 0
	}
	for {
		p.normalize()
		mc := p.minContext
		numMasked := p.numStats(mc)
		for {
			p.orderFall++
			if p.suffix(mc) == 0 {
				return -1
			}
			mc = p.suffix(mc)
			if p.numStats(mc) != numMasked {
				break
			}
		}
		p.minContext = mc
		stats := p.stats(mc)
		end := stats + p.numStats(mc)*6
		hiCnt := uint32(0)
		for s := stats; s != end; s += 6 {
			if st := p.m.u16(s); p.mask[byte(st)] != p.escCount {
				hiCnt += uint32(st >> 8)
			}
		}
		see, freqSum := p.makeEscFreq(numMasked)
		freqSum += hiCnt
		count := p.threshold(freqSum)
		if count < hiCnt {
			hiCnt = count
			s := stats
			for {
				if st := p.m.u16(s); p.mask[byte(st)] != p.escCount {
					count -= uint32(st >> 8)
				}
				s += 6
				if int32(count) < 0 {
					break
				}
			}
			s -= 6
			p.rcDecode(hiCnt-count-p.freq(s), p.freq(s))
			p.normalize()
			see.update()
			p.foundState = s
			sym := p.sym(s)
			p.update2()
			return int(sym)
		}
		if count >= freqSum {
			return -2
		}
		p.rcDecode(hiCnt, freqSum-hiCnt)
		see.summ += uint16(freqSum)
		for s := stats; s != end; s += 6 {
			p.mask[p.sym(s)] = p.escCount
		}
	}
}

// decodePPMd decodes a 7z PPMd stream (order, then the model size) to size
// bytes; the stream must end exactly with the range coder's final state.
func decodePPMd(props, src []byte, size uint64, done <-chan struct{}) (out []byte, err error) {
	if len(props) != 5 {
		return nil, ErrData
	}
	order, mem := uint32(props[0]), binary.LittleEndian.Uint32(props[1:])
	if order < ppmdMinOrder || order > ppmdMaxOrder || mem < ppmdMinMem || mem > ppmdMaxMem {
		return nil, ErrData
	}
	if mem > ppmdMaxModel {
		return nil, ErrLimit
	}
	defer func() {
		if r := recover(); r != nil {
			out, err = nil, ErrData // the model's invariants do not hold
		}
	}()
	p := newPPMd7(order, mem)
	if !p.rcInit(src) {
		return nil, ErrData
	}
	out = make([]byte, 0, size)
	for uint64(len(out)) < size {
		if len(out)&0xfff == 0 && stopped(done) {
			return nil, ErrAborted
		}
		sym := p.decodeSymbol()
		if p.extra || sym < 0 {
			return nil, ErrData
		}
		out = append(out, byte(sym))
	}
	if p.pos != len(src) || p.code != 0 {
		return nil, ErrData
	}
	return out, nil
}
