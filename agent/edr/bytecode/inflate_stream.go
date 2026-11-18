package bytecode

import (
	"errors"
	"hash/crc32"
)

// This decoder keeps DEFLATE's bit, block, Huffman and 32 KiB history state
// across calls. It consumes only bytes actually needed by the current step,
// which lets the bytecode pipe cursors advance incrementally. RFC 1951 and
// RFC 1950 and RFC 1952 define its raw, zlib and gzip modes. The adapter
// searches for full-flush markers after data errors, as zlib's inflateSync
// does, while preserving the pipe cursors at each call boundary.
var errInflateData = errors.New("invalid deflate stream")

const (
	inflateOK        = 0
	inflateStreamEnd = 1
	inflateNeedDict  = 2
	inflateBufError  = -5
)

type inflatePhase uint8

const (
	phaseAutoHeader inflatePhase = iota
	phaseZlibHeader
	phaseZlibDictID
	phaseNeedDict
	phaseGzipHeader
	phaseGzipExtraLen
	phaseGzipExtra
	phaseGzipName
	phaseGzipComment
	phaseGzipHeaderCRC
	phaseBlockHeader
	phaseStoredHeader
	phaseStoredBytes
	phaseDynamicCounts
	phaseDynamicCodes
	phaseDynamicLengths
	phaseDynamicRepeat
	phaseCompressed
	phaseLiteral
	phaseLengthExtra
	phaseDistanceSymbol
	phaseDistanceExtra
	phaseCopy
	phaseZlibTrailer
	phaseGzipTrailerCRC
	phaseGzipTrailerSize
	phaseComplete
)

type inflateWrap uint8

const (
	wrapRaw inflateWrap = iota
	wrapZlib
	wrapGzip
	wrapAuto
)

type huffmanTree struct {
	codes  [16]map[uint16]uint16
	maxLen uint8
}

func (h *huffmanTree) build(lengths []uint8, allowEmpty, allowSingle bool) error {
	*h = huffmanTree{}
	var count [16]int
	for _, length := range lengths {
		if length > 15 {
			return errInflateData
		}
		if length != 0 {
			count[length]++
			if length > h.maxLen {
				h.maxLen = length
			}
		}
	}
	if h.maxLen == 0 {
		if allowEmpty {
			return nil
		}
		return errInflateData
	}
	left := 1
	var next [16]int
	for n := 1; n <= 15; n++ {
		left = left*2 - count[n]
		if left < 0 {
			return errInflateData
		}
		if n > 1 {
			next[n] = (next[n-1] + count[n-1]) << 1
		}
	}
	// RFC 1951 allows a single-symbol literal or distance alphabet, but
	// otherwise an incomplete code tree is invalid. The code-length alphabet
	// must be complete, as required by zlib's inflate_table(CODES).
	if left != 0 && !(allowSingle && h.maxLen == 1 && count[1] == 1) {
		return errInflateData
	}
	for symbol, length := range lengths {
		if length == 0 {
			continue
		}
		code := uint16(next[length])
		next[length]++
		var reversed uint16
		for i := uint8(0); i < length; i++ {
			reversed = reversed<<1 | code&1
			code >>= 1
		}
		if h.codes[length] == nil {
			h.codes[length] = make(map[uint16]uint16)
		}
		h.codes[length][reversed] = uint16(symbol)
	}
	return nil
}

type inflateIO struct {
	in  []byte
	out []byte
	i   int
	o   int
}

type streamInflater struct {
	phase      inflatePhase
	wrap       inflateWrap
	maxWindow  int
	windowSize int
	headerSeen bool
	needSync   bool
	syncing    bool
	syncHave   uint8
	noCheck    bool
	bits       uint64
	nbits      uint
	final      bool
	window     [32768]byte
	winPos     int
	produced   uint64
	adlerA     uint32
	adlerB     uint32
	crc        uint32
	headerCRC  uint32
	gzipHeader [10]byte
	headerAt   int
	gzipFlags  byte
	extraLeft  int
	lit        huffmanTree
	dist       huffmanTree
	code       huffmanTree
	storedLeft int
	copyLeft   int
	copyDist   int
	length     int
	lengthX    uint8
	literal    byte
	distance   int
	distanceX  uint8
	dynLit     int
	dynDist    int
	dynCode    int
	dynIndex   int
	dynRepeat  int
	dynLens    [320]uint8
	codeLens   [19]uint8
}

func newStreamInflater(windowBits int32) (*streamInflater, bool) {
	d := &streamInflater{adlerA: 1}
	switch {
	case windowBits >= -15 && windowBits <= -8:
		d.wrap = wrapRaw
		d.headerSeen = true
		d.maxWindow = 1 << uint(-windowBits)
		d.windowSize = d.maxWindow
		d.phase = phaseBlockHeader
	case windowBits == 0:
		d.wrap = wrapZlib
		d.maxWindow = 32768
		d.phase = phaseZlibHeader
	case windowBits >= 8 && windowBits <= 15:
		d.wrap = wrapZlib
		d.maxWindow = 1 << uint(windowBits)
		d.phase = phaseZlibHeader
	case windowBits >= 24 && windowBits <= 31:
		d.wrap = wrapGzip
		d.maxWindow = 1 << uint(windowBits-16)
		d.windowSize = d.maxWindow
		d.phase = phaseGzipHeader
	case windowBits >= 40 && windowBits <= 47:
		d.wrap = wrapAuto
		d.maxWindow = 1 << uint(windowBits-32)
		d.windowSize = d.maxWindow
		d.phase = phaseAutoHeader
	default:
		return nil, false
	}
	return d, true
}

func (d *streamInflater) need(io *inflateIO, count uint) bool {
	for d.nbits < count {
		if io.i >= len(io.in) {
			return false
		}
		d.bits |= uint64(io.in[io.i]) << d.nbits
		d.nbits += 8
		io.i++
	}
	return true
}

func (d *streamInflater) take(count uint) uint32 {
	value := uint32(d.bits & (uint64(1)<<count - 1))
	d.bits >>= count
	d.nbits -= count
	return value
}

func (d *streamInflater) symbol(io *inflateIO, tree *huffmanTree) (int, bool, error) {
	for n := uint(1); n <= uint(tree.maxLen); n++ {
		if !d.need(io, n) {
			return 0, false, nil
		}
		code := uint16(d.bits & (uint64(1)<<n - 1))
		if symbol, ok := tree.codes[n][code]; ok {
			d.take(n)
			return int(symbol), true, nil
		}
	}
	return 0, true, errInflateData
}

func (d *streamInflater) emit(io *inflateIO, b byte) bool {
	if io.o == len(io.out) {
		return false
	}
	io.out[io.o] = b
	io.o++
	d.window[d.winPos] = b
	d.winPos++
	if d.winPos == d.windowSize {
		d.winPos = 0
	}
	d.produced++
	if d.wrap == wrapZlib {
		d.adlerA += uint32(b)
		if d.adlerA >= 65521 {
			d.adlerA -= 65521
		}
		d.adlerB += d.adlerA
		if d.adlerB >= 65521 {
			d.adlerB -= 65521
		}
	} else if d.wrap == wrapGzip {
		var one = [1]byte{b}
		d.crc = crc32.Update(d.crc, crc32.IEEETable, one[:])
	}
	return true
}

func (d *streamInflater) afterDeflate() {
	switch d.wrap {
	case wrapRaw:
		d.phase = phaseComplete
	case wrapZlib:
		d.phase = phaseZlibTrailer
	case wrapGzip:
		d.phase = phaseGzipTrailerCRC
	}
}

func (d *streamInflater) gzipHeaderByte(b byte) {
	d.headerCRC = crc32.Update(d.headerCRC, crc32.IEEETable, []byte{b})
}

func (d *streamInflater) afterGzipExtra() {
	if d.gzipFlags&8 != 0 {
		d.phase = phaseGzipName
	} else {
		d.afterGzipName()
	}
}

func (d *streamInflater) afterGzipName() {
	if d.gzipFlags&16 != 0 {
		d.phase = phaseGzipComment
	} else {
		d.afterGzipComment()
	}
}

func (d *streamInflater) afterGzipComment() {
	if d.gzipFlags&2 != 0 {
		d.phase = phaseGzipHeaderCRC
	} else {
		d.phase = phaseBlockHeader
		d.headerSeen = true
	}
}

var lengthBase = [...]int{3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258}
var lengthExtra = [...]uint8{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0}
var distanceBase = [...]int{1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513, 769, 1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577}
var distanceExtra = [...]uint8{0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13}
var codeOrder = [...]int{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15}

func (d *streamInflater) fixedTrees() error {
	var lit [288]uint8
	for i := 0; i < 144; i++ {
		lit[i] = 8
	}
	for i := 144; i < 256; i++ {
		lit[i] = 9
	}
	for i := 256; i < 280; i++ {
		lit[i] = 7
	}
	for i := 280; i < 288; i++ {
		lit[i] = 8
	}
	var dist [32]uint8
	for i := range dist {
		dist[i] = 5
	}
	if err := d.lit.build(lit[:], false, true); err != nil {
		return err
	}
	return d.dist.build(dist[:], false, true)
}

func (d *streamInflater) dynamicTrees() error {
	if d.dynLens[256] == 0 {
		return errInflateData
	}
	if err := d.lit.build(d.dynLens[:d.dynLit], false, true); err != nil {
		return err
	}
	return d.dist.build(d.dynLens[d.dynLit:d.dynLit+d.dynDist], true, true)
}

// process returns the exact consumed and emitted byte counts. A status of
// inflateOK means the current input or output window was exhausted; callers
// may provide another window. A malformed stream returns errInflateData.
func (d *streamInflater) process(input, output []byte) (int, int, int, error) {
	io := &inflateIO{in: input, out: output}
	finish := func(status int, err error) (int, int, int, error) {
		if status == inflateOK && io.i == 0 && io.o == 0 {
			status = inflateBufError
		}
		return io.i, io.o, status, err
	}
	for {
		switch d.phase {
		case phaseAutoHeader:
			if !d.need(io, 16) {
				return finish(inflateOK, nil)
			}
			if d.bits&0xffff == 0x8b1f {
				d.wrap = wrapGzip
				d.phase = phaseGzipHeader
			} else {
				d.wrap = wrapZlib
				d.phase = phaseZlibHeader
			}
		case phaseZlibHeader:
			if !d.need(io, 16) {
				return finish(inflateOK, nil)
			}
			header := uint32(d.bits & 0xffff)
			cmf, flg := byte(header), byte(header>>8)
			windowBits := int(cmf>>4) + 8
			if cmf&15 != 8 || windowBits > 15 || windowBits < 8 || (uint32(cmf)<<8|uint32(flg))%31 != 0 || windowBits > bitLenWindow(d.maxWindow) {
				return finish(inflateOK, errInflateData)
			}
			d.take(16)
			d.windowSize = 1 << uint(windowBits)
			d.headerSeen = true
			if flg&0x20 != 0 {
				d.phase = phaseZlibDictID
			} else {
				d.phase = phaseBlockHeader
			}
		case phaseZlibDictID:
			if !d.need(io, 32) {
				return finish(inflateOK, nil)
			}
			d.take(32) // dictionary Adler ID; bytecode has no dictionary setter
			d.phase = phaseNeedDict
		case phaseNeedDict:
			return finish(inflateNeedDict, nil)
		case phaseGzipHeader:
			if d.headerAt == 0 {
				if !d.need(io, 16) {
					return finish(inflateOK, nil)
				}
				if d.bits&0xffff != 0x8b1f {
					return finish(inflateOK, errInflateData)
				}
			}
			if d.headerAt == len(d.gzipHeader) {
				if d.gzipHeader[0] != 0x1f || d.gzipHeader[1] != 0x8b || d.gzipHeader[2] != 8 || d.gzipHeader[3]&0xe0 != 0 {
					return finish(inflateOK, errInflateData)
				}
				d.gzipFlags = d.gzipHeader[3]
				if d.gzipFlags&4 != 0 {
					d.phase = phaseGzipExtraLen
				} else {
					d.afterGzipExtra()
				}
				continue
			}
			if !d.need(io, 8) {
				return finish(inflateOK, nil)
			}
			b := byte(d.take(8))
			d.gzipHeader[d.headerAt] = b
			d.headerAt++
			d.gzipHeaderByte(b)
		case phaseGzipExtraLen:
			if !d.need(io, 16) {
				return finish(inflateOK, nil)
			}
			n := d.take(16)
			d.gzipHeaderByte(byte(n))
			d.gzipHeaderByte(byte(n >> 8))
			d.extraLeft = int(n)
			d.phase = phaseGzipExtra
		case phaseGzipExtra:
			if d.extraLeft == 0 {
				d.afterGzipExtra()
				continue
			}
			if !d.need(io, 8) {
				return finish(inflateOK, nil)
			}
			d.gzipHeaderByte(byte(d.take(8)))
			d.extraLeft--
		case phaseGzipName:
			if !d.need(io, 8) {
				return finish(inflateOK, nil)
			}
			b := byte(d.take(8))
			d.gzipHeaderByte(b)
			if b == 0 {
				d.afterGzipName()
			}
		case phaseGzipComment:
			if !d.need(io, 8) {
				return finish(inflateOK, nil)
			}
			b := byte(d.take(8))
			d.gzipHeaderByte(b)
			if b == 0 {
				d.afterGzipComment()
			}
		case phaseGzipHeaderCRC:
			if !d.need(io, 16) {
				return finish(inflateOK, nil)
			}
			if d.take(16) != d.headerCRC&0xffff {
				return finish(inflateOK, errInflateData)
			}
			d.phase = phaseBlockHeader
			d.headerSeen = true
		case phaseBlockHeader:
			if !d.need(io, 3) {
				return finish(inflateOK, nil)
			}
			header := d.take(3)
			d.final = header&1 != 0
			switch header >> 1 {
			case 0:
				d.phase = phaseStoredHeader
			case 1:
				d.lit, d.dist = huffmanTree{}, huffmanTree{}
				if err := d.fixedTrees(); err != nil {
					return finish(inflateOK, err)
				}
				d.phase = phaseCompressed
			case 2:
				d.phase = phaseDynamicCounts
			default:
				return finish(inflateOK, errInflateData)
			}
		case phaseStoredHeader:
			// Stored blocks start at the next byte boundary.
			d.take(d.nbits % 8)
			if !d.need(io, 32) {
				return finish(inflateOK, nil)
			}
			header := d.take(32)
			if uint16(header)^uint16(header>>16) != 0xffff {
				return finish(inflateOK, errInflateData)
			}
			d.storedLeft = int(uint16(header))
			d.phase = phaseStoredBytes
		case phaseStoredBytes:
			if d.storedLeft == 0 {
				if d.final {
					d.afterDeflate()
				} else {
					d.phase = phaseBlockHeader
				}
				continue
			}
			if io.o == len(io.out) || !d.need(io, 8) {
				return finish(inflateOK, nil)
			}
			b := byte(d.take(8))
			d.emit(io, b)
			d.storedLeft--
		case phaseDynamicCounts:
			if !d.need(io, 14) {
				return finish(inflateOK, nil)
			}
			header := d.take(14)
			d.dynLit = int(header&31) + 257
			d.dynDist = int((header>>5)&31) + 1
			d.dynCode = int((header>>10)&15) + 4
			if d.dynLit > 286 || d.dynDist > 30 {
				return finish(inflateOK, errInflateData)
			}
			d.dynIndex = 0
			d.code, d.lit, d.dist = huffmanTree{}, huffmanTree{}, huffmanTree{}
			d.codeLens, d.dynLens = [19]uint8{}, [320]uint8{}
			d.phase = phaseDynamicCodes
		case phaseDynamicCodes:
			if d.dynIndex == d.dynCode {
				if err := d.code.build(d.codeLens[:], false, false); err != nil {
					return finish(inflateOK, err)
				}
				d.dynIndex = 0
				d.phase = phaseDynamicLengths
				continue
			}
			if !d.need(io, 3) {
				return finish(inflateOK, nil)
			}
			d.codeLens[codeOrder[d.dynIndex]] = uint8(d.take(3))
			d.dynIndex++
		case phaseDynamicLengths:
			if d.dynIndex == d.dynLit+d.dynDist {
				if err := d.dynamicTrees(); err != nil {
					return finish(inflateOK, err)
				}
				d.phase = phaseCompressed
				continue
			}
			symbol, ok, err := d.symbol(io, &d.code)
			if err != nil {
				return finish(inflateOK, err)
			}
			if !ok {
				return finish(inflateOK, nil)
			}
			if symbol <= 15 {
				d.dynLens[d.dynIndex] = uint8(symbol)
				d.dynIndex++
				continue
			}
			d.dynRepeat = symbol
			d.phase = phaseDynamicRepeat
		case phaseDynamicRepeat:
			var count uint
			var times int
			var value uint8
			switch d.dynRepeat {
			case 16:
				if d.dynIndex == 0 {
					return finish(inflateOK, errInflateData)
				}
				count, times, value = 2, 3, d.dynLens[d.dynIndex-1]
			case 17:
				count, times = 3, 3
			case 18:
				count, times = 7, 11
			default:
				return finish(inflateOK, errInflateData)
			}
			if !d.need(io, count) {
				return finish(inflateOK, nil)
			}
			times += int(d.take(count))
			if d.dynIndex+times > d.dynLit+d.dynDist {
				return finish(inflateOK, errInflateData)
			}
			for range times {
				d.dynLens[d.dynIndex] = value
				d.dynIndex++
			}
			d.phase = phaseDynamicLengths
		case phaseCompressed:
			symbol, ok, err := d.symbol(io, &d.lit)
			if err != nil {
				return finish(inflateOK, err)
			}
			if !ok {
				return finish(inflateOK, nil)
			}
			if symbol < 256 {
				d.literal = byte(symbol)
				d.phase = phaseLiteral
				continue
			}
			if symbol == 256 {
				if d.final {
					d.afterDeflate()
				} else {
					d.phase = phaseBlockHeader
				}
				continue
			}
			if symbol < 257 || symbol > 285 {
				return finish(inflateOK, errInflateData)
			}
			i := symbol - 257
			d.length, d.lengthX = lengthBase[i], lengthExtra[i]
			d.phase = phaseLengthExtra
		case phaseLiteral:
			if !d.emit(io, d.literal) {
				return finish(inflateOK, nil)
			}
			d.phase = phaseCompressed
		case phaseLengthExtra:
			if !d.need(io, uint(d.lengthX)) {
				return finish(inflateOK, nil)
			}
			d.length += int(d.take(uint(d.lengthX)))
			d.phase = phaseDistanceSymbol
		case phaseDistanceSymbol:
			symbol, ok, err := d.symbol(io, &d.dist)
			if err != nil {
				return finish(inflateOK, err)
			}
			if !ok {
				return finish(inflateOK, nil)
			}
			if symbol >= 30 {
				return finish(inflateOK, errInflateData)
			}
			d.distance, d.distanceX = distanceBase[symbol], distanceExtra[symbol]
			d.phase = phaseDistanceExtra
		case phaseDistanceExtra:
			if !d.need(io, uint(d.distanceX)) {
				return finish(inflateOK, nil)
			}
			d.distance += int(d.take(uint(d.distanceX)))
			if d.distance > d.windowSize || uint64(d.distance) > d.produced {
				return finish(inflateOK, errInflateData)
			}
			d.copyLeft, d.copyDist = d.length, d.distance
			d.phase = phaseCopy
		case phaseCopy:
			if d.copyLeft == 0 {
				d.phase = phaseCompressed
				continue
			}
			if io.o == len(io.out) {
				return finish(inflateOK, nil)
			}
			at := d.winPos - d.copyDist
			if at < 0 {
				at += d.windowSize
			}
			d.emit(io, d.window[at])
			d.copyLeft--
		case phaseZlibTrailer:
			d.take(d.nbits % 8)
			if !d.need(io, 32) {
				return finish(inflateOK, nil)
			}
			trailer := uint32(d.bits & 0xffffffff)
			want := trailer>>24 | (trailer>>8)&0xff00 | (trailer<<8)&0xff0000 | trailer<<24
			if !d.noCheck && want != d.adlerB<<16|d.adlerA {
				return finish(inflateOK, errInflateData)
			}
			d.take(32)
			d.phase = phaseComplete
		case phaseGzipTrailerCRC:
			d.take(d.nbits % 8)
			if !d.need(io, 32) {
				return finish(inflateOK, nil)
			}
			if uint32(d.bits&0xffffffff) != d.crc && !d.noCheck {
				return finish(inflateOK, errInflateData)
			}
			d.take(32)
			d.phase = phaseGzipTrailerSize
		case phaseGzipTrailerSize:
			if !d.need(io, 32) {
				return finish(inflateOK, nil)
			}
			if uint32(d.bits&0xffffffff) != uint32(d.produced) && !d.noCheck {
				return finish(inflateOK, errInflateData)
			}
			d.take(32)
			d.phase = phaseComplete
		case phaseComplete:
			return finish(inflateStreamEnd, nil)
		default:
			return finish(inflateOK, errInflateData)
		}
	}
}

func (d *streamInflater) syncByte(b byte) {
	if b == 0 && d.syncHave < 2 || b == 0xff && d.syncHave >= 2 {
		d.syncHave++
	} else if b != 0 {
		d.syncHave = 0
	} else {
		d.syncHave = 4 - d.syncHave
	}
}

// sync searches for zlib's full-flush marker, including bytes already held
// by the bit reader. Its match state survives input-window boundaries.
func (d *streamInflater) sync(input []byte) (int, int) {
	if len(input) == 0 && d.nbits < 8 {
		return 0, inflateBufError
	}
	if !d.syncing {
		d.syncing = true
		d.syncHave = 0
		d.take(d.nbits % 8)
		for d.nbits >= 8 && d.syncHave < 4 {
			d.syncByte(byte(d.take(8)))
		}
	}
	used := 0
	for used < len(input) && d.syncHave < 4 {
		d.syncByte(input[used])
		used++
	}
	if d.syncHave != 4 {
		return used, -3 // Z_DATA_ERROR: more input may still recover
	}
	// inflateSync calls inflateReset, retaining wrapper selection but disabling
	// the final checksum. A corrupt header instead recovers as a raw stream.
	d.needSync, d.syncing, d.syncHave = false, false, 0
	d.bits, d.nbits, d.final = 0, 0, false
	d.phase, d.produced, d.winPos = phaseBlockHeader, 0, 0
	d.noCheck = true
	if !d.headerSeen {
		d.wrap = wrapRaw
		d.windowSize = d.maxWindow
		d.headerSeen = true
	}
	return used, inflateOK
}

// processWithRecovery matches ClamAV's inflate_process loop: after a data
// error it repeatedly calls inflateSync, then resumes inflate at the next
// full-flush block. A failed search returns zlib's status and keeps the
// search state for the next pipe window.
func (d *streamInflater) processWithRecovery(input, output []byte) (int, int, int, error) {
	read, wrote := 0, 0
	for attempts := 0; attempts <= len(input)+8; attempts++ {
		if !d.needSync {
			n, m, status, err := d.process(input[read:], output[wrote:])
			read, wrote = read+n, wrote+m
			if err == nil || err != errInflateData {
				return read, wrote, status, err
			}
			d.needSync = true
		}
		n, status := d.sync(input[read:])
		read += n
		if status != inflateOK {
			return read, wrote, status, nil
		}
	}
	return read, wrote, inflateBufError, ErrLimit
}

func bitLenWindow(window int) int {
	bits := 0
	for window > 1 {
		window >>= 1
		bits++
	}
	return bits
}
