package engine

import (
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"encoding/binary"
	"errors"
	"io"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
	"github.com/utmstack/UTMStack/agent/edr/zipcodec"
)

// A ZIP local file header found inside another object (a self-extractor
// record of the raw scan) is read the way the reference reads it: that one
// header and the member it describes, without a central directory. A header
// that does not hold up (its name, extra field or data would run past the
// object, its data is masked, or the member is empty) describes nothing to
// analyze. Every local header is its own record, so each member is analyzed
// once.

const (
	zipLocalHeaderSize = 30
	zipFlagEncrypted   = 0x0001
	zipFlagDescriptor  = 0x0008 // sizes follow the data
	zipFlagMasked      = 0x2000
)

// zipSingle analyzes the member whose local header starts at off.
func (s *scanState) zipSingle(data []byte, off, depth int) (string, error) {
	rest := len(data) - off
	if rest < zipLocalHeaderSize {
		return "", nil
	}
	h := data[off:]
	flags := binary.LittleEndian.Uint16(h[6:])
	method := binary.LittleEndian.Uint16(h[8:])
	csize := uint64(binary.LittleEndian.Uint32(h[18:]))
	usize := uint64(binary.LittleEndian.Uint32(h[22:]))
	nameLen := int(binary.LittleEndian.Uint16(h[26:]))
	extraLen := int(binary.LittleEndian.Uint16(h[28:]))
	room := rest - zipLocalHeaderSize
	if room <= nameLen {
		return "", nil
	}
	name := string(h[zipLocalHeaderSize : zipLocalHeaderSize+nameLen])
	if csize == 0xffffffff || usize == 0xffffffff {
		if extraLen > room-nameLen {
			return "", ErrMalformed
		}
		var err error
		start := zipLocalHeaderSize + nameLen
		usize, csize, _, err = zip64Extra(h[start:start+extraLen], uint32(usize), uint32(csize), 0)
		if err != nil {
			return "", err
		}
	}
	if depth+1 > s.maxDepth {
		return "", ErrLimit
	}
	// The reference pushes the found archive as a ZIP layer: container rules
	// and container metadata rules see the member inside a ZIP.
	s.layers.Push(signatures.Layer{Type: signatures.TypeZIP, Size: int64(rest)})
	defer s.layers.Pop()
	depth++
	meta := signatures.MemberMeta{Name: name, HasName: true, CompressedSize: zipMetaSize(csize), Size: zipMetaSize(usize),
		Encrypted: flags&zipFlagEncrypted != 0, Index: 1}
	if sig, err := s.db.MatchMember(s.ctx, signatures.TypeZIP, int64(rest), meta); sig != "" || err != nil {
		return sig, err
	}
	room -= nameLen
	if flags&zipFlagMasked != 0 || room <= extraLen {
		return "", nil
	}
	room -= extraLen
	start := off + zipLocalHeaderSize + nameLen + extraLen
	if flags&zipFlagDescriptor != 0 {
		return s.zipStreamed(data[start:], method, flags, depth)
	}
	if csize == 0 || uint64(room) < csize {
		return "", nil
	}
	packed := data[start : start+int(csize)]
	if flags&zipFlagEncrypted != 0 {
		var err error
		packed, err = s.zipDecrypt(packed, binary.LittleEndian.Uint16(h[4:]), flags,
			binary.LittleEndian.Uint16(h[10:]), binary.LittleEndian.Uint32(h[14:]))
		if err != nil {
			return "", err
		}
	}
	switch method {
	case zip0Stored:
		if uint64(len(packed)) < usize {
			return s.inflatedMember(packed, depth)
		}
		return s.childBytes(packed, depth)
	case zip8Deflate:
		return s.inflatedMember(packed, depth)
	case zipBzip2:
		return s.child(bzip2.NewReader(bytes.NewReader(packed)), depth)
	case zip6Implode, zip9Deflate64:
		return s.zipMethod(method, flags, packed, depth)
	case zip93Zstd, zip20ZstdLegacy:
		return s.zstd(packed, depth)
	case zip14LZMA:
		return s.zipLZMA(packed, flags, usize, depth)
	}
	return "", ErrUnsupported
}

// inflatedMember analyzes a deflate member as the reference does (checked
// with damaged members, docs/2026-10-14-cfb-oracle/ziplab.py): data that ends
// inside the stream yields what decoded up to there, and a stream that fails
// to decode yields nothing, even past its first block. Its sizes and CRC are
// not checked.
func (s *scanState) inflatedMember(packed []byte, depth int) (string, error) {
	if depth >= s.maxDepth || s.files >= s.maxFiles {
		return "", ErrLimit
	}
	limit := min(s.maxFile, s.remaining)
	if limit < 0 {
		return "", ErrLimit
	}
	data, err := io.ReadAll(io.LimitReader(contextReader{s.ctx, flate.NewReader(bytes.NewReader(packed))}, limit+1))
	switch {
	case s.ctx.Err() != nil:
		return "", s.ctx.Err()
	case int64(len(data)) > limit:
		s.spend(len(data))
		return "", ErrLimit
	case errors.Is(err, io.ErrUnexpectedEOF):
	case err != nil:
		s.spend(len(data))
		return "", nil
	}
	return s.visit(data, depth+1)
}

// zipMethod analyzes a Deflate64 or implode member as the reference does
// (docs/edr/native-engine/parity/zip.md section 4): Deflate64 data yields
// what decoded before it ends early and nothing when it fails, as deflate
// does; imploded data is decoded until it can hold no further token, whatever
// the declared size, and yields nothing when it fails. Sizes and CRC are not
// checked.
func (s *scanState) zipMethod(method, flags uint16, packed []byte, depth int) (string, error) {
	if depth >= s.maxDepth || s.files >= s.maxFiles {
		return "", ErrLimit
	}
	limit := min(s.maxFile, s.remaining, 1<<30)
	if limit < 0 {
		return "", ErrLimit
	}
	var data []byte
	var err error
	if method == zip9Deflate64 {
		data, err = zipcodec.Inflate64(packed, int(limit))
	} else {
		data, err = zipcodec.Explode(packed, flags, int(limit))
	}
	switch {
	case s.ctx.Err() != nil:
		return "", s.ctx.Err()
	case errors.Is(err, zipcodec.ErrLimit):
		s.spend(len(data)) // failed expansion work is charged
		return "", ErrLimit
	case err != nil && !errors.Is(err, zipcodec.ErrTruncated), len(data) == 0:
		s.spend(len(data)) // a failure (written out but not scanned by the reference)
		return "", nil
	}
	return s.visit(data, depth+1)
}

// ZIP compression methods (APPNOTE 4.4.5) besides bzip2 (zipBzip2).
const (
	zip0Stored    = 0
	zip6Implode   = 6
	zip8Deflate   = 8
	zip9Deflate64 = 9
	zip14LZMA     = 14
	// APPNOTE 6.3.8 moved Zstandard from method 20 to 93. Read the
	// deprecated identifier too, as current readers do for compatibility:
	// https://pkware.cachefly.net/webdocs/APPNOTE/APPNOTE-6.3.9.TXT
	zip20ZstdLegacy = 20
	zip93Zstd       = 93
)

// zipStreamed analyzes a member whose sizes follow its data. The reference
// does not analyze such a member without a central directory; a deflate or
// bzip2 stream ends by itself, so it is decoded here when it can be. Nothing
// is lost when it cannot, so a failure is not reported.
func (s *scanState) zipStreamed(rest []byte, method, flags uint16, depth int) (string, error) {
	if flags&zipFlagEncrypted != 0 {
		return "", nil
	}
	var r io.Reader
	switch method {
	case zip8Deflate:
		r = flate.NewReader(bytes.NewReader(rest))
	case zipBzip2:
		r = bzip2.NewReader(bytes.NewReader(rest))
	default:
		return "", nil
	}
	if sig, _ := s.child(r, depth); sig != "" {
		return sig, nil
	}
	return "", s.ctx.Err()
}
