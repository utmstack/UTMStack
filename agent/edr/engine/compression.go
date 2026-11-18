package engine

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/nwaples/rardecode/v2"
	"github.com/ulikunitz/xz"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

const maxCodecDictionary = 64 << 20

// The XZ decoder allocates every block's declared dictionary, even for an
// empty block. Dictionary bytes beyond a block's own output are not covered by
// the expansion budget, so their total across a scanned XZ object is capped.
const maxXZDictionaryOverhead = 4 * maxCodecDictionary

func (s *scanState) xz(data []byte, depth int) (string, error) {
	// xz.ReaderConfig.DictCap is a minimum, not a maximum. Validate every
	// block's dictionary request before constructing any decoder.
	remaining := uint64(max(0, min(s.maxFile, s.remaining)))
	partial := false
	if err := preflightXZ(data, remaining); err != nil {
		if !errors.Is(err, ErrMalformed) {
			return "", err
		}
		// A missing or cut-off footer can follow a fully decodable block and
		// index. Reconstruct it only for the dictionary/expansion preflight;
		// the decoder still sees the original damaged input.
		prefix, footer, tailErr := s.xzMissingFooter(data)
		if tailErr != nil {
			if !errors.Is(tailErr, ErrMalformed) {
				return "", tailErr
			}
			return "", err
		}
		if err := preflightXZWithTail(prefix, footer, remaining); err != nil {
			return "", err
		}
		partial = true
	}
	r, err := (xz.ReaderConfig{DictCap: 4096}).NewReader(contextReader{s.ctx, bytes.NewReader(data)})
	if err != nil {
		return "", ErrMalformed
	}
	if partial {
		return s.partialChild(r, depth)
	}
	return s.child(r, depth)
}

// xzMissingFooter walks the physical LZMA2 chunk extents to locate a final
// stream's index. It does not decode or allocate dictionaries. The subsequent
// full preflight validates every block against that index before NewReader.
func (s *scanState) xzMissingFooter(data []byte) ([]byte, []byte, error) {
	pos := 0
	for pos < len(data) {
		if err := s.ctx.Err(); err != nil {
			return nil, nil, err
		}
		for pos+4 <= len(data) && bytes.Equal(data[pos:pos+4], []byte{0, 0, 0, 0}) {
			pos += 4
		}
		if len(data)-pos < 12 || !bytes.Equal(data[pos:pos+6], []byte{0xfd, '7', 'z', 'X', 'Z', 0}) {
			return nil, nil, ErrMalformed
		}
		flags := data[pos+6 : pos+8]
		checkSizes := map[byte]int{0: 0, 1: 4, 4: 8, 10: 32}
		checkSize, ok := checkSizes[flags[1]]
		if flags[0] != 0 || !ok {
			return nil, nil, ErrMalformed
		}
		pos += 12
		blocks := 0
		for {
			if err := s.ctx.Err(); err != nil {
				return nil, nil, err
			}
			if pos >= len(data) {
				return nil, nil, ErrMalformed
			}
			if data[pos] == 0 {
				break
			}
			if blocks >= 65536 {
				return nil, nil, ErrLimit
			}
			blocks++
			start := pos
			hlen := (int(data[pos]) + 1) * 4
			if hlen < 8 || hlen > len(data)-pos {
				return nil, nil, ErrMalformed
			}
			pos += hlen
			for {
				if err := s.ctx.Err(); err != nil {
					return nil, nil, err
				}
				if pos >= len(data) {
					return nil, nil, ErrMalformed
				}
				control := data[pos]
				pos++
				if control == 0 {
					break
				}
				if control < 0x80 && control != 1 && control != 2 || len(data)-pos < 2 {
					return nil, nil, ErrMalformed
				}
				u := int(binary.BigEndian.Uint16(data[pos:pos+2])) + 1
				packed := u
				pos += 2
				if control >= 0x80 {
					if len(data)-pos < 2 {
						return nil, nil, ErrMalformed
					}
					packed = int(binary.BigEndian.Uint16(data[pos:pos+2])) + 1
					pos += 2
					if control >= 0xc0 {
						pos++
					}
				}
				if packed > len(data)-pos {
					return nil, nil, ErrMalformed
				}
				pos += packed
			}
			padding := (4 - (pos-start)%4) % 4
			if padding+checkSize > len(data)-pos {
				return nil, nil, ErrMalformed
			}
			pos += padding + checkSize
		}
		indexStart := pos
		pos++ // index indicator
		count, err := xzVLI(data, &pos)
		if err != nil || count != uint64(blocks) {
			return nil, nil, ErrMalformed
		}
		for i := uint64(0); i < count; i++ {
			if _, err := xzVLI(data, &pos); err != nil {
				return nil, nil, err
			}
			if _, err := xzVLI(data, &pos); err != nil {
				return nil, nil, err
			}
		}
		for (pos-indexStart)%4 != 0 {
			if pos >= len(data) || data[pos] != 0 {
				return nil, nil, ErrMalformed
			}
			pos++
		}
		if len(data)-pos < 4 {
			return nil, nil, ErrMalformed
		}
		pos += 4 // index CRC
		indexSize := pos - indexStart
		footer := make([]byte, 12)
		binary.LittleEndian.PutUint32(footer[4:], uint32(indexSize/4-1))
		copy(footer[8:10], flags)
		binary.LittleEndian.PutUint32(footer[:4], crc32.ChecksumIEEE(footer[4:10]))
		footer[10], footer[11] = 'Y', 'Z'
		if len(data)-pos < len(footer) {
			if !bytes.Equal(data[pos:], footer[:len(data)-pos]) {
				return nil, nil, ErrMalformed
			}
			return data[:pos], footer, nil
		}
		if !bytes.Equal(data[pos:pos+len(footer)], footer) {
			return nil, nil, ErrMalformed
		}
		pos += len(footer)
	}
	return nil, nil, ErrMalformed
}

// Zstandard skippable frames (RFC 8878 section 3.1.2) may precede the first
// data frame. Decoders ignore them, so they also start a Zstandard stream; the
// decoder reports a stream that does not continue validly as malformed.
func zstdSkippableFrame(b []byte) bool {
	return len(b) >= 4 && b[0]&0xf0 == 0x50 && b[1] == 0x2a && b[2] == 0x4d && b[3] == 0x18
}

func (s *scanState) zstd(data []byte, depth int) (string, error) {
	r, err := zstd.NewReader(contextReader{s.ctx, bytes.NewReader(data)},
		zstd.WithDecoderConcurrency(1), zstd.WithDecoderLowmem(true),
		zstd.WithDecoderMaxWindow(maxCodecDictionary), zstd.WithDecoderMaxMemory(maxCodecDictionary))
	if err != nil {
		return "", ErrMalformed
	}
	defer r.Close()
	return s.child(r, depth)
}

// RAR4's PPM model has an independent, format-limited 256MiB allocation. Admit
// one RAR decoder at a time across all scans, in addition to the scan semaphore.
// Nested RARs are read completely before recursively visiting their output.
var rarDecoderSlot = make(chan struct{}, 1)

func (s *scanState) rar(data []byte, depth int) (string, error) {
	select {
	case rarDecoderSlot <- struct{}{}:
	case <-s.ctx.Done():
		return "", s.ctx.Err()
	}
	locked := true
	defer func() {
		if locked {
			<-rarDecoderSlot
		}
	}()
	r, err := rardecode.NewReader(contextReader{s.ctx, bytes.NewReader(data)}, rardecode.MaxDictionarySize(maxCodecDictionary))
	if err != nil {
		return "", rarError(err)
	}
	// Buffering the bounded extracted members lets us release the decoder before
	// visiting nested archives. Buffered members stay in memory while earlier
	// ones are analyzed, so each is charged to the scan budget when it is read.
	var members [][]byte
	var incomplete error
	count := 0
	for {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			incomplete = rarError(err)
			break
		}
		count++
		member := signatures.MemberMeta{Name: h.Name, HasName: true, CompressedSize: h.PackedSize, Size: h.UnPackedSize,
			Encrypted: h.Encrypted, Index: count}
		if name, err := s.db.MatchMember(s.ctx, s.layers.TypeAt(-1), int64(len(data)), member); name != "" || err != nil {
			return name, err
		}
		if count > s.maxFiles-s.files || depth >= s.maxDepth {
			incomplete = ErrLimit
			break
		}
		if h.Encrypted || h.HeaderEncrypted {
			incomplete = ErrUnsupported
			break
		}
		if h.IsDir {
			continue
		}
		limit := min(s.maxFile, s.remaining)
		if limit < 0 || !h.UnKnownSize && (h.UnPackedSize < 0 || h.UnPackedSize > limit) {
			incomplete = ErrLimit
			break
		}
		b, err := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, limit+1))
		if err != nil {
			s.spend(len(b)) // failed expansion work is charged
			incomplete = rarError(err)
			break
		}
		if int64(len(b)) > limit {
			s.spend(len(b))
			incomplete = ErrLimit
			break
		}
		s.remaining -= int64(len(b))
		members = append(members, b)
	}
	<-rarDecoderSlot
	locked = false
	for _, b := range members {
		sig, err := s.visitHeld(b, depth+1, int64(len(b)))
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}

func rarError(err error) error {
	if errors.Is(err, rardecode.ErrDictionaryTooLarge) {
		return ErrLimit
	}
	if errors.Is(err, rardecode.ErrArchiveEncrypted) || errors.Is(err, rardecode.ErrArchivedFileEncrypted) || errors.Is(err, rardecode.ErrUnknownVersion) {
		return ErrUnsupported
	}
	return ErrMalformed
}

// XZ VLI values have at most 63 bits and a minimal little-endian encoding.
func xzVLI(b []byte, pos *int) (uint64, error) {
	var v uint64
	for n := 0; n < 9; n++ {
		if *pos >= len(b) {
			return 0, ErrMalformed
		}
		c := b[*pos]
		*pos++
		v |= uint64(c&127) << uint(n*7)
		if c&128 == 0 {
			if n > 0 && c == 0 {
				return 0, ErrMalformed
			}
			return v, nil
		}
	}
	return 0, ErrMalformed
}

// The index gives each padded block extent, so even concatenated XZ streams can
// be checked without decompressing attacker-controlled data or allocating from
// an unchecked count. The decoder subsequently verifies all CRCs and sizes.
func preflightXZ(data []byte, remaining uint64) error {
	return preflightXZWithTail(data, nil, remaining)
}

// tail is a virtual footer for a final stream whose on-disk footer is missing.
// The input remains immutable, avoiding a second full-size XZ allocation.
func preflightXZWithTail(data, tail []byte, remaining uint64) error {
	end := len(data)
	virtual := len(tail) == 12
	if virtual {
		end += 12
	}
	blocks := 0
	var overhead uint64
	for end > 0 {
		padding := 0
		if !virtual {
			for end > 0 && data[end-1] == 0 {
				end--
				padding++
			}
		}
		if padding%4 != 0 || end < 24 {
			return ErrMalformed
		}
		f := tail
		if !virtual {
			f = data[end-12 : end]
		}
		if f[10] != 'Y' || f[11] != 'Z' {
			return ErrMalformed
		}
		n := (uint64(binary.LittleEndian.Uint32(f[4:8])) + 1) * 4
		if n > uint64(end-24) || n < 8 {
			return ErrMalformed
		}
		indexStart := end - 12 - int(n)
		index := data[indexStart : end-16]
		if index[0] != 0 {
			return ErrMalformed
		}
		p := 1
		count, err := xzVLI(index, &p)
		if err != nil {
			return err
		}
		if count > uint64(len(index)/2) || count > 65536-uint64(blocks) {
			return ErrLimit
		}
		blocks += int(count)
		checkSizes := map[byte]uint64{0: 0, 1: 4, 4: 8, 10: 32}
		checkSize, ok := checkSizes[f[9]]
		if f[8] != 0 || !ok {
			return ErrUnsupported
		}
		type block struct{ size, unpacked uint64 }
		entries := make([]block, 0, int(count))
		var compressed uint64
		for i := uint64(0); i < count; i++ {
			u, err := xzVLI(index, &p)
			if err != nil {
				return err
			}
			v, err := xzVLI(index, &p)
			if err != nil {
				return err
			}
			if u < 5 || u > uint64(indexStart) || (u+3)&^3 > uint64(indexStart)-compressed {
				return ErrMalformed
			}
			if v > remaining {
				return ErrLimit
			}
			remaining -= v
			compressed += (u + 3) &^ 3
			entries = append(entries, block{u, v})
		}
		if len(index)-p > 3 {
			return ErrMalformed
		}
		for _, b := range index[p:] {
			if b != 0 {
				return ErrMalformed
			}
		}
		start := indexStart - int(compressed) - 12
		if start < 0 || !bytes.Equal(data[start:start+6], []byte{0xfd, '7', 'z', 'X', 'Z', 0}) || !bytes.Equal(data[start+6:start+8], f[8:10]) {
			return ErrMalformed
		}
		pos := start + 12
		for _, e := range entries {
			hlen := (int(data[pos]) + 1) * 4
			if hlen < 8 || uint64(hlen) > e.size || hlen > indexStart-pos {
				return ErrMalformed
			}
			h := data[pos : pos+hlen-4]
			if h[1]&0x3f != 0 {
				return ErrUnsupported
			} // currently one LZMA2 filter
			q := 2
			if h[1]&0x40 != 0 {
				if _, err := xzVLI(h, &q); err != nil {
					return err
				}
			}
			if h[1]&0x80 != 0 {
				if _, err := xzVLI(h, &q); err != nil {
					return err
				}
			}
			id, err := xzVLI(h, &q)
			if err != nil {
				return err
			}
			props, err := xzVLI(h, &q)
			if err != nil {
				return err
			}
			if id != 0x21 || props != 1 {
				return ErrUnsupported
			}
			if q >= len(h) || h[q] > 40 {
				return ErrMalformed
			}
			dict := uint64(2|h[q]&1) << uint(h[q]/2+11)
			if dict > maxCodecDictionary {
				return ErrLimit
			}
			if dict > e.unpacked {
				if overhead += dict - e.unpacked; overhead > maxXZDictionaryOverhead {
					return ErrLimit
				}
			}
			q++
			for _, b := range h[q:] {
				if b != 0 {
					return ErrMalformed
				}
			}
			// Locate the actual end of the LZMA2 chunks, independently of the
			// index. Otherwise a forged index could conceal a later dictionary
			// header from this preflight while the decoder still reaches it.
			cpos := pos + hlen
			var unpacked uint64
			for {
				if cpos >= indexStart {
					return ErrMalformed
				}
				control := data[cpos]
				cpos++
				if control == 0 {
					break
				}
				if control < 0x80 && control != 1 && control != 2 {
					return ErrMalformed
				}
				if indexStart-cpos < 2 {
					return ErrMalformed
				}
				u := uint64(binary.BigEndian.Uint16(data[cpos:cpos+2])) + 1
				packed := u
				cpos += 2
				if control >= 0x80 {
					u += uint64(control&31) << 16
					if indexStart-cpos < 2 {
						return ErrMalformed
					}
					packed = uint64(binary.BigEndian.Uint16(data[cpos:cpos+2])) + 1
					cpos += 2
					if control >= 0xc0 {
						cpos++
					}
				}
				if cpos > indexStart || packed > uint64(indexStart-cpos) || u > e.unpacked-unpacked {
					return ErrMalformed
				}
				unpacked += u
				cpos += int(packed)
			}
			if unpacked != e.unpacked || uint64(cpos-pos)+checkSize != e.size {
				return ErrMalformed
			}
			pos += int((e.size + 3) &^ 3)
		}
		end = start
		virtual = false
	}
	return nil
}
