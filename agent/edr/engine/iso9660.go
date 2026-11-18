package engine

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

const isoSectorSize = 2048

type isoLayout struct {
	sectorSize int
	dataOffset int
}

var isoLayouts = [...]isoLayout{
	{2048, 0},
	{2336, 0}, {2336, 8}, {2336, 16}, {2336, 24},
	{2352, 0}, {2352, 8}, {2352, 16}, {2352, 24},
	{2448, 0}, {2448, 8}, {2448, 16}, {2448, 24},
}

func (l isoLayout) sector(data []byte, n int) ([]byte, bool) {
	off := n*l.sectorSize + l.dataOffset
	if off < 0 || off+isoSectorSize > len(data) {
		return nil, false
	}
	return data[off : off+isoSectorSize], true
}

// isoLayoutOf recognizes the primary descriptor in supported cooked and raw
// sector layouts. It does not guess a layout from an arbitrary CD001 string.
func isoLayoutOf(data []byte) (isoLayout, bool) {
	for _, l := range isoLayouts {
		sector, ok := l.sector(data, 16)
		if ok && sector[0] == 1 && bytes.Equal(sector[1:6], []byte("CD001")) && sector[6] == 1 {
			return l, true
		}
	}
	return isoLayout{}, false
}

func isoHeader(data []byte) bool { _, ok := isoLayoutOf(data); return ok }

type isoVolume struct {
	data      []byte
	layout    isoLayout
	blockSize int
	joliet    bool
	seen      map[isoDirKey]bool
}

type isoDirKey struct {
	start uint64
	size  uint32
}

type isoRecord struct {
	extent   uint32
	size     uint32
	extended byte
	flags    byte
	unit     byte
	gap      byte
	name     string
}

func isoBoth16(b []byte) (int, bool) {
	if len(b) < 4 || binary.LittleEndian.Uint16(b) != binary.BigEndian.Uint16(b[2:]) {
		return 0, false
	}
	return int(binary.LittleEndian.Uint16(b)), true
}

func isoBoth32(b []byte) (uint32, bool) {
	if len(b) < 8 || binary.LittleEndian.Uint32(b) != binary.BigEndian.Uint32(b[4:]) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b), true
}

func isoName(b []byte, joliet bool) (string, bool) {
	if !joliet {
		return string(bytes.SplitN(b, []byte{';'}, 2)[0]), true
	}
	if len(b)&1 != 0 {
		return "", false
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		u[i] = binary.BigEndian.Uint16(b[2*i:])
	}
	for i, v := range u {
		if v >= 0xd800 && v <= 0xdbff {
			if i+1 >= len(u) || u[i+1] < 0xdc00 || u[i+1] > 0xdfff {
				return "", false
			}
		} else if v >= 0xdc00 && v <= 0xdfff && (i == 0 || u[i-1] < 0xd800 || u[i-1] > 0xdbff) {
			return "", false
		}
	}
	name := string(utf16.Decode(u))
	if !utf8.ValidString(name) {
		return "", false
	}
	if p := bytes.IndexByte([]byte(name), ';'); p >= 0 {
		name = name[:p]
	}
	return name, true
}

func (v *isoVolume) record(b []byte) (isoRecord, bool) {
	var r isoRecord
	if len(b) < 34 || b[0] < 34 || int(b[0]) > len(b) || int(b[32])+33 > int(b[0]) {
		return r, false
	}
	var ok bool
	r.extent, ok = isoBoth32(b[2:10])
	if !ok {
		return r, false
	}
	r.size, ok = isoBoth32(b[10:18])
	if !ok {
		return r, false
	}
	r.extended, r.flags, r.unit, r.gap = b[1], b[25], b[26], b[27]
	if b[32] == 1 && (b[33] == 0 || b[33] == 1) {
		return r, true
	}
	name, ok := isoName(b[33:33+int(b[32])], v.joliet)
	if !ok {
		return r, false
	}
	r.name = name
	return r, true
}

func (v *isoVolume) extent(r isoRecord, maxBytes int64) ([]byte, error) {
	// The extended-attribute length precedes file data in logical blocks.
	if int64(r.size) > maxBytes {
		return nil, ErrLimit
	}
	startBlock := uint64(r.extent) + uint64(r.extended)
	if v.layout.sectorSize == isoSectorSize {
		start := startBlock * uint64(v.blockSize)
		end := start + uint64(r.size)
		if start > uint64(len(v.data)) || end > uint64(len(v.data)) {
			return nil, ErrMalformed
		}
		return v.data[int(start):int(end)], nil
	}
	// Raw sectors have a header and trailer around each 2048-byte data
	// portion, so a file spanning sectors must be gathered block by block.
	// Validate the entire mapped extent before allocating a reconstruction;
	// a damaged image must not trigger repeated large failed allocations.
	for left, block := int(r.size), startBlock; left > 0; block++ {
		if _, ok := v.block(block); !ok {
			return nil, ErrMalformed
		}
		left -= min(v.blockSize, left)
	}
	out := make([]byte, int(r.size))
	for written, block := 0, startBlock; written < len(out); block++ {
		part, _ := v.block(block) // validated above against immutable input
		n := min(v.blockSize, len(out)-written)
		copy(out[written:written+n], part[:n])
		written += n
	}
	return out, nil
}

func (v *isoVolume) block(number uint64) ([]byte, bool) {
	logical := number * uint64(v.blockSize)
	sector := logical / isoSectorSize
	within := logical % isoSectorSize
	off := sector*uint64(v.layout.sectorSize) + uint64(v.layout.dataOffset) + within
	if off > uint64(len(v.data)) || uint64(v.blockSize) > uint64(len(v.data))-off {
		return nil, false
	}
	return v.data[int(off) : int(off)+v.blockSize], true
}

func (s *scanState) iso9660(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	var incomplete error
	layout, ok := isoLayoutOf(data)
	if !ok {
		return "", ErrUnsupported
	}
	for sector := 16; sector < 32; sector++ {
		desc, ok := layout.sector(data, sector)
		if !ok {
			return "", ErrMalformed
		}
		if !bytes.Equal(desc[1:6], []byte("CD001")) || desc[6] != 1 {
			return "", ErrMalformed
		}
		if desc[0] == 255 {
			return "", incomplete
		}
		joliet := desc[0] == 2 && desc[88] == '%' && desc[89] == '/' &&
			(desc[90] == '@' || desc[90] == 'C' || desc[90] == 'E')
		if desc[0] != 1 && !joliet {
			continue
		}
		blockSize, ok := isoBoth16(desc[128:132])
		if !ok || (blockSize != 512 && blockSize != 1024 && blockSize != 2048) {
			incomplete = ErrUnsupported
			continue
		}
		v := isoVolume{data: data, layout: layout, blockSize: blockSize, joliet: joliet, seen: make(map[isoDirKey]bool)}
		root, ok := v.record(desc[156:])
		if !ok {
			incomplete = ErrMalformed
			continue
		}
		if root.unit != 0 || root.gap != 0 || root.flags&2 == 0 {
			incomplete = ErrUnsupported
			continue
		}
		if sig, err := s.isoDirectory(&v, root, depth, 0); sig != "" {
			return sig, nil
		} else if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	if incomplete == nil {
		incomplete = ErrUnsupported // descriptor chain exceeded the bounded window
	}
	return "", incomplete
}

func (s *scanState) isoDirectory(v *isoVolume, dir isoRecord, depth, dirs int) (string, error) {
	if dirs >= s.maxDepth || len(v.seen) >= 1024 {
		return "", ErrLimit
	}
	key := isoDirKey{start: uint64(dir.extent) + uint64(dir.extended), size: dir.size}
	if v.seen[key] {
		return "", nil
	}
	v.seen[key] = true
	if int64(dir.size) > s.maxFile {
		return "", ErrLimit
	}
	var incomplete error
	for scanned := 0; scanned < int(dir.size); {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		block, ok := v.block(uint64(dir.extent) + uint64(dir.extended) + uint64(scanned/v.blockSize))
		if !ok {
			return "", ErrMalformed
		}
		limit := min(v.blockSize, int(dir.size)-scanned)
		for pos := 0; pos < limit; {
			if block[pos] == 0 {
				break
			}
			n := int(block[pos])
			if n < 34 || pos+n > limit {
				return "", ErrMalformed
			}
			raw := block[pos : pos+n]
			pos += n
			if raw[32] == 1 && (raw[33] == 0 || raw[33] == 1) {
				continue
			}
			r, ok := v.record(raw)
			if !ok {
				incomplete = ErrMalformed
				continue
			}
			// The reference supplies the stored length for both size fields and
			// position zero for every ISO directory record.
			meta := signatures.MemberMeta{Name: r.name, HasName: true, CompressedSize: int64(r.size), Size: int64(r.size), Index: 0}
			if sig, err := s.db.MatchMember(s.ctx, signatures.TypeISO9660, int64(len(v.data)), meta); sig != "" || err != nil {
				return sig, err
			}
			if r.unit != 0 || r.gap != 0 {
				incomplete = ErrUnsupported
				continue
			}
			// ClamAV scans each multi-extent record as an individual section.
			// The complete logical file is not reconstructed here, so a
			// non-detection must remain incomplete rather than become clean.
			if r.flags&0x80 != 0 {
				incomplete = ErrUnsupported
			}
			if r.flags&2 != 0 {
				if sig, err := s.isoDirectory(v, r, depth, dirs+1); sig != "" {
					return sig, nil
				} else if err != nil && incomplete == nil {
					incomplete = err
				}
				continue
			}
			b, err := v.extent(r, min(s.maxFile, s.remaining))
			if err != nil {
				incomplete = err
				continue
			}
			if sig, err := s.visit(b, depth+1); sig != "" {
				return sig, nil
			} else if err != nil && incomplete == nil {
				incomplete = err
			}
		}
		scanned += limit
	}
	return "", incomplete
}
