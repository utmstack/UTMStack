package bytecode

// This decoder follows the LZMA wire format and the bounded streaming contract
// of Igor Pavlov's public-domain LZMA SDK. ClamAV's pinned SDK is an independent
// test oracle; neither its C code nor a native library is linked here.
import "encoding/binary"

const (
	lzMatch     = 0
	lzRep       = 192
	lzRepG0     = 204
	lzRepG1     = 216
	lzRepG2     = 228
	lzRepLong   = 240
	lzPosSlot   = 432
	lzSpecPos   = 688
	lzAlign     = 802
	lzLength    = 818
	lzRepLength = 1332
	lzLiteral   = 1846
	lzFinished  = 274
)

type lzmaRegisters struct {
	rangeValue, code   uint32
	processed, checked uint32
	state              uint32
	reps               [4]uint32 // distances, including the extra one
	pos                int
}

type streamLZMA struct {
	lzmaRegisters
	lc, lp, pb           uint
	dict                 []byte
	probs                []uint16
	remaining            uint64
	matchLeft            uint32
	temp                 [20]byte
	tempSize             int
	needSeed, needModels bool
	charge               func(uint64) error
}

func lzmaAllocation(header []byte) (dict, probabilities uint64, valid bool) {
	if len(header) != 13 || header[0] >= 225 {
		return 0, 0, false
	}
	lc, lp := uint(header[0]%9), uint(header[0]/9%5)
	return uint64(max(binary.LittleEndian.Uint32(header[1:5]), 4096)),
		2 * (lzLiteral + (768 << (lc + lp))), true
}

func newStreamLZMA(header []byte) (*streamLZMA, error) {
	dict, probabilities, valid := lzmaAllocation(header)
	if !valid {
		return nil, ErrUnsupported
	}
	if dict > maxLZMADict || probabilities > maxLZMADict {
		return nil, ErrLimit
	}
	d := &streamLZMA{lc: uint(header[0] % 9), lp: uint(header[0] / 9 % 5), pb: uint(header[0] / 45),
		dict: make([]byte, int(dict)), probs: make([]uint16, int(probabilities/2)),
		remaining: binary.LittleEndian.Uint64(header[5:13]), needSeed: true, needModels: true}
	return d, nil
}

type lzmaRange struct {
	input            []byte
	pos              int
	rangeValue, code uint32
	probs            []uint16
	update           bool
	short            bool
	bits             uint64
}

func (r *lzmaRange) normalize() {
	if r.rangeValue < 1<<24 && !r.short {
		if r.pos == len(r.input) {
			r.short = true
			return
		}
		r.rangeValue <<= 8
		r.code = r.code<<8 | uint32(r.input[r.pos])
		r.pos++
	}
}

func (r *lzmaRange) bit(index int) uint32 {
	r.normalize()
	if r.short {
		return 0
	}
	r.bits++
	p := uint32(r.probs[index])
	bound := (r.rangeValue >> 11) * p
	if r.code < bound {
		r.rangeValue = bound
		if r.update {
			r.probs[index] = uint16(p + ((2048 - p) >> 5))
		}
		return 0
	}
	r.rangeValue -= bound
	r.code -= bound
	if r.update {
		r.probs[index] = uint16(p - (p >> 5))
	}
	return 1
}

func (r *lzmaRange) tree(base, bits int) uint32 {
	n := uint32(1)
	for i := 0; i < bits; i++ {
		n = n<<1 | r.bit(base+int(n))
	}
	return n - (1 << bits)
}

func (r *lzmaRange) reverse(base, bits int) uint32 {
	n, result := uint32(1), uint32(0)
	for i := 0; i < bits; i++ {
		b := r.bit(base + int(n))
		n = n<<1 | b
		result |= b << i
	}
	return result
}

func (r *lzmaRange) length(base int, position uint32) uint32 {
	if r.bit(base) == 0 {
		return r.tree(base+2+int(position)*8, 3)
	}
	if r.bit(base+1) == 0 {
		return 8 + r.tree(base+130+int(position)*8, 3)
	}
	return 16 + r.tree(base+258, 8)
}

// token decodes a literal, ordinary match or repetition. Probe mode leaves
// probabilities unchanged and checks input sufficiency without committing data.
func (d *streamLZMA) token(c *lzmaRegisters, r *lzmaRange) (kind int, literal byte, distance, length uint32, valid bool) {
	valid = true
	position := c.processed & ((1 << d.pb) - 1)
	if r.bit(lzMatch+int(c.state)*16+int(position)) == 0 {
		kind, length = 1, 1
		context := uint32(0)
		if c.checked != 0 || c.processed != 0 {
			prev := c.pos
			if prev == 0 {
				prev = len(d.dict)
			}
			context = ((c.processed & ((1 << d.lp) - 1)) << d.lc) + uint32(d.dict[prev-1])>>(8-d.lc)
		}
		base := lzLiteral + 768*int(context)
		n := uint32(1)
		if c.state < 7 {
			for n < 256 {
				n = n<<1 | r.bit(base+int(n))
			}
		} else {
			p := c.pos - int(c.reps[0])
			if p < 0 {
				p += len(d.dict)
			}
			if p < 0 || p >= len(d.dict) {
				return kind, 0, 0, 0, false
			}
			match, mask := uint32(d.dict[p]), uint32(256)
			for n < 256 {
				match <<= 1
				b := match & mask
				bit := r.bit(base + int(mask+b+n))
				n = n<<1 | bit
				if bit == 0 {
					mask &= ^b
				} else {
					mask &= b
				}
			}
		}
		if c.state < 4 {
			c.state = 0
		} else if c.state < 10 {
			c.state -= 3
		} else {
			c.state -= 6
		}
		return kind, byte(n), 0, 1, true
	}
	oldState := c.state
	if r.bit(lzRep+int(oldState)) == 0 {
		kind = 2
		length = r.length(lzLength, position)
		slot := r.tree(lzPosSlot+64*int(min(length, 3)), 6)
		offset := slot
		if slot >= 4 {
			bits := int(slot>>1) - 1
			offset = 2 | (slot & 1)
			if slot < 14 {
				offset <<= bits
				offset |= r.reverse(lzSpecPos+int(offset)-int(slot)-1, bits)
			} else {
				for i := 0; i < bits-4; i++ {
					r.normalize()
					r.bits++
					r.rangeValue >>= 1
					r.code -= r.rangeValue
					t := uint32(0) - (r.code >> 31)
					offset = (offset << 1) + (t + 1)
					r.code += r.rangeValue & t
				}
				offset = offset<<4 | r.reverse(lzAlign, 4)
				if offset == ^uint32(0) {
					return kind, 0, 0, length + lzFinished, true
				}
			}
		}
		distance = offset + 1
		if r.update && (c.checked == 0 && offset >= c.processed || c.checked != 0 && offset >= c.checked) {
			return kind, 0, 0, 0, false
		}
		c.reps[3], c.reps[2], c.reps[1], c.reps[0] = c.reps[2], c.reps[1], c.reps[0], distance
		if oldState < 7 {
			c.state = 7
		} else {
			c.state = 10
		}
		return kind, 0, distance, length + 2, true
	}
	kind = 3
	if r.update && c.checked == 0 && c.processed == 0 {
		return kind, 0, 0, 0, false
	}
	distance = c.reps[0]
	if r.bit(lzRepG0+int(oldState)) == 0 {
		if r.bit(lzRepLong+int(oldState)*16+int(position)) == 0 {
			if oldState < 7 {
				c.state = 9
			} else {
				c.state = 11
			}
			return kind, 0, distance, 1, true
		}
	} else {
		if r.bit(lzRepG1+int(oldState)) == 0 {
			distance = c.reps[1]
		} else {
			if r.bit(lzRepG2+int(oldState)) == 0 {
				distance = c.reps[2]
			} else {
				distance = c.reps[3]
				c.reps[3] = c.reps[2]
			}
			c.reps[2] = c.reps[1]
		}
		c.reps[1], c.reps[0] = c.reps[0], distance
	}
	length = r.length(lzRepLength, position) + 2
	if oldState < 7 {
		c.state = 8
	} else {
		c.state = 11
	}
	return kind, 0, distance, length, true
}

func (d *streamLZMA) probe(input []byte) (int, error) {
	c := d.lzmaRegisters
	r := lzmaRange{input: input, rangeValue: c.rangeValue, code: c.code, probs: d.probs}
	kind, _, _, _, valid := d.token(&c, &r)
	r.normalize()
	if err := d.charge(r.bits); err != nil {
		return 0, err
	}
	if r.short || !valid {
		return 0, nil
	}
	return kind, nil
}

func (d *streamLZMA) copyMatch(c *lzmaRegisters, length uint32, limit int) (uint32, bool) {
	n := min(int(length), limit-c.pos)
	if c.reps[0] == 0 || int(c.reps[0]) > len(d.dict) || n < 0 {
		return 0, false
	}
	for i := 0; i < n; i++ {
		p := c.pos - int(c.reps[0])
		if p < 0 {
			p += len(d.dict)
		}
		d.dict[c.pos] = d.dict[p]
		c.pos++
		c.processed++
	}
	return length - uint32(n), true
}

func (d *streamLZMA) writeRemaining(limit int) error {
	if d.matchLeft == 0 || d.matchLeft >= lzFinished {
		return nil
	}
	n := min(int(d.matchLeft), limit-d.pos)
	if err := d.charge(uint64(n)); err != nil {
		return err
	}
	if d.checked == 0 && uint32(len(d.dict))-d.processed <= uint32(n) {
		d.checked = uint32(len(d.dict))
	}
	var valid bool
	d.matchLeft, valid = d.copyMatch(&d.lzmaRegisters, d.matchLeft, limit)
	if !valid {
		return ErrRuntime
	}
	return nil
}

// real commits registers and the input cursor only after a whole safe decoding
// group succeeds. Damaged data must not publish a speculative output cursor.
func (d *streamLZMA) real(input []byte, limit, inputLimit int) (int, bool, error) {
	c := d.lzmaRegisters
	r := lzmaRange{input: input, rangeValue: c.rangeValue, code: c.code, probs: d.probs, update: true}
	left := uint32(0)
	for {
		before := r.bits
		kind, b, distance, length, valid := d.token(&c, &r)
		if err := d.charge(r.bits - before + uint64(min(length, 273))); err != nil {
			return 0, false, err
		}
		if !valid || r.short {
			return 0, false, nil
		}
		if kind == 1 {
			if c.pos >= limit {
				return 0, false, nil
			}
			d.dict[c.pos] = b
			c.pos++
			c.processed++
		} else {
			if length >= lzFinished {
				left = length
				break
			}
			if c.pos == limit {
				return 0, false, nil
			}
			c.reps[0] = distance
			left, valid = d.copyMatch(&c, length, limit)
			if !valid {
				return 0, false, nil
			}
		}
		if c.pos >= limit || r.pos >= inputLimit {
			break
		}
	}
	r.normalize()
	if r.short {
		return 0, false, nil
	}
	c.rangeValue, c.code = r.rangeValue, r.code
	d.lzmaRegisters = c
	d.matchLeft = left
	return r.pos, true, nil
}

func (d *streamLZMA) realGroups(input []byte, limit, inputLimit int) (consumed int, valid bool, err error) {
	for {
		bound := limit
		if d.checked == 0 {
			bound = min(bound, d.pos+len(d.dict)-int(d.processed))
		}
		n, ok, e := d.real(input[consumed:], bound, inputLimit-consumed)
		if e != nil || !ok {
			return 0, ok, e
		}
		consumed += n
		if d.processed >= uint32(len(d.dict)) {
			d.checked = uint32(len(d.dict))
		}
		if err = d.writeRemaining(limit); err != nil {
			return 0, false, err
		}
		if d.pos >= limit || consumed >= inputLimit || d.matchLeft >= lzFinished {
			break
		}
	}
	if d.matchLeft > lzFinished {
		d.matchLeft = lzFinished
	}
	return consumed, true, nil
}

func (d *streamLZMA) decodeDictionary(input []byte, limit int, finish bool) (consumed, status int, valid bool, err error) {
	valid = true
	if err = d.writeRemaining(limit); err != nil {
		return
	}
	for d.matchLeft != lzFinished {
		if d.needSeed {
			for consumed < len(input) && d.tempSize < 5 {
				d.temp[d.tempSize] = input[consumed]
				d.tempSize++
				consumed++
			}
			if d.tempSize < 5 {
				return consumed, 3, true, nil
			}
			if d.temp[0] != 0 {
				return consumed, 0, false, nil
			}
			d.code = binary.BigEndian.Uint32(d.temp[1:5])
			d.rangeValue = ^uint32(0)
			d.tempSize = 0
			d.needSeed = false
		}
		checkEnd := false
		if d.pos >= limit {
			if d.matchLeft == 0 && d.code == 0 {
				return consumed, 4, true, nil
			}
			if !finish {
				return consumed, 2, true, nil
			}
			if d.matchLeft != 0 {
				return consumed, 2, false, nil
			}
			checkEnd = true
		}
		if d.needModels {
			for i := range d.probs {
				d.probs[i] = 1024
			}
			d.reps = [4]uint32{1, 1, 1, 1}
			d.state = 0
			d.needModels = false
		}
		if d.tempSize == 0 {
			inputLimit := len(input) - consumed - 20
			if inputLimit < 0 || checkEnd {
				var kind int
				kind, err = d.probe(input[consumed:])
				if err != nil {
					return
				}
				if kind == 0 {
					d.tempSize = copy(d.temp[:], input[consumed:])
					consumed += d.tempSize
					return consumed, 3, true, nil
				}
				if checkEnd && kind != 2 {
					return consumed, 2, false, nil
				}
				inputLimit = 0
			}
			var n int
			n, valid, err = d.realGroups(input[consumed:], limit, inputLimit)
			if err != nil || !valid {
				return
			}
			consumed += n
		} else {
			old := d.tempSize
			look := copy(d.temp[old:], input[consumed:])
			d.tempSize += look
			if d.tempSize < 20 || checkEnd {
				var kind int
				kind, err = d.probe(d.temp[:d.tempSize])
				if err != nil {
					return
				}
				if kind == 0 {
					return consumed + look, 3, true, nil
				}
				if checkEnd && kind != 2 {
					return consumed, 2, false, nil
				}
			}
			n, ok, e := d.realGroups(d.temp[:d.tempSize], limit, 0)
			if e != nil || !ok {
				return consumed, 0, ok, e
			}
			if n < old || n > d.tempSize {
				return consumed, 0, false, ErrRuntime
			}
			consumed += n - old
			d.tempSize = 0
		}
	}
	return consumed, 1, d.code == 0, nil
}

func (d *streamLZMA) process(input, output []byte, charge func(uint64) error) (consumed, written, status int, err error) {
	if charge == nil {
		charge = func(uint64) error { return nil }
	}
	d.charge = charge
	finish := d.remaining != ^uint64(0) && uint64(len(output)) > d.remaining
	limit := len(output)
	if finish {
		limit = int(d.remaining)
	}
	valid := true
	for {
		if d.pos == len(d.dict) {
			d.pos = 0
		}
		start := d.pos
		end := min(len(d.dict), start+limit-written)
		endFinish := finish && start+limit-written <= len(d.dict)
		n, s, ok, e := d.decodeDictionary(input[consumed:], end, endFinish)
		consumed += n
		got := d.pos - start
		copy(output[written:], d.dict[start:d.pos])
		written += got
		status, valid, err = s, ok, e
		if err != nil || !valid || got == 0 || written == limit {
			break
		}
	}
	if d.remaining != ^uint64(0) {
		d.remaining -= uint64(written)
	}
	if err != nil {
		return
	}
	if !valid {
		return consumed, written, 1, nil
	}
	if d.remaining == 0 || status == 1 {
		return consumed, written, 2, nil
	}
	return consumed, written, 0, nil
}
