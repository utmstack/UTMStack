package engine

import (
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"strconv"

	"github.com/ulikunitz/xz/lzma"

	"github.com/utmstack/UTMStack/agent/edr/engine/internal/nsisbzip2"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Executables carry installers and payloads in three places the scanner
// inspects without running the stub: installer data in the overlay (NSIS,
// InstallShield, AutoIt, self-extracting archives), and complete executables
// stored at any offset inside the file, which the reference engine also
// extracts and scans ("embedded PE").
func (s *scanState) pe(data []byte, depth int, parser parserFacts) (string, error) {
	if notPE(data) {
		return "", nil
	}
	img, err := parsePEScanImage(data)
	if err != nil {
		if parser.headerRejected && errors.Is(err, ErrMalformed) {
			// The reference already rejected this PE header. Raw signatures
			// and embedded objects were scanned before this format pass.
			return "", nil
		}
		return "", err
	}
	var incomplete error
	// Where this engine follows the reference's parser to its built-in
	// unpackers, the parser decides (an unpacker this engine lacks makes the
	// match incomplete); a program whose header the parser model cannot read
	// keeps the packer evidence: code and data may need an unpacker.
	if img.packed && !parser.upxDecided {
		incomplete = ErrUnsupported
	}
	// The program the UPX step rebuilt is scanned as a child, as the
	// reference scans it inside its parser.
	if parser.unpacked != nil {
		sig, err := s.visit(parser.unpacked, depth+1)
		if sig != "" {
			return sig, nil
		}
		if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	// Installer data and archives analyze their own members; executables
	// stored as they are in the image or in overlay data nothing recognized
	// are analyzed separately. The Authenticode signature appended by signing
	// tools is not overlay data; data before and after it is. Executables are
	// recognized by a loadable header rather than a short magic, and the
	// reference engine finds them anywhere in a file, so the signature is
	// searched for them too.
	ranges := [][2]int{{1, img.overlay}}
	for i, part := range [][2]int{{img.overlay, img.certificate}, {img.certificateEnd, len(data)}} {
		if i == 1 && img.certificate < img.certificateEnd {
			ranges = append(ranges, [2]int{img.certificate, img.certificateEnd})
		}
		sig, unrecognized, err := s.peOverlay(data[part[0]:part[1]], depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
		for _, r := range unrecognized {
			ranges = append(ranges, [2]int{part[0] + r[0], part[0] + r[1]})
		}
	}
	sig, err := s.embeddedPE(data, ranges, depth)
	if sig != "" {
		return sig, nil
	}
	if err != nil {
		incomplete = err
	}
	return "", incomplete
}

// maxEmbeddedPE bounds how much of a file is analyzed as one embedded
// executable, as the reference engine's default MaxEmbeddedPE does.
// maxEmbeddedCandidates bounds the PE headers parsed while searching.
const (
	maxEmbeddedPE         = 40 << 20
	maxEmbeddedCandidates = 1024
)

// embeddedPE finds executables with a loadable PE header at offsets inside
// the given ranges of data and analyzes each from its start to the next such
// executable or the end of the file. Executables nested inside one found here
// are found again when it is analyzed, so each is analyzed once and the total
// charged is at most the size of data.
func (s *scanState) embeddedPE(data []byte, ranges [][2]int, depth int) (string, error) {
	var incomplete error
	type candidate struct{ start, extent int }
	var found []candidate
	parsed, checked, pos := 0, 0, 1
search:
	for _, r := range ranges {
		for pos = max(pos, r[0]); pos < r[1]; checked++ {
			if checked&1023 == 0 {
				if err := s.ctx.Err(); err != nil {
					return "", err
				}
			}
			i := bytes.Index(data[pos:r[1]], []byte("MZ"))
			if i < 0 {
				break
			}
			start := pos + i
			pos = start + 1
			if !embeddedPEHeader(data[start:]) {
				continue
			}
			// Each header parse may walk a long section table; crafted files
			// with thousands of candidate headers must not multiply that work.
			if parsed++; parsed > maxEmbeddedCandidates {
				incomplete = ErrLimit
				break search
			}
			img, err := parsePEImage(data[start:])
			if err != nil {
				continue
			}
			if len(found) >= s.maxFiles-s.files {
				incomplete = ErrLimit
				break search
			}
			found = append(found, candidate{start, img.overlay})
			pos = max(pos, start+img.overlay) // nested executables are found when this one is analyzed
		}
	}
	for i, c := range found {
		end := len(data)
		if i+1 < len(found) {
			end = found[i+1].start
		}
		if end-c.start > maxEmbeddedPE {
			if c.extent > maxEmbeddedPE {
				incomplete = ErrLimit
				continue
			}
			end = c.start + maxEmbeddedPE
		}
		sig, err := s.visit(data[c.start:end], depth+1)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}

// embeddedPEHeader cheaply checks for a PE signature before full parsing.
func embeddedPEHeader(b []byte) bool {
	if len(b) < 64 {
		return false
	}
	offset := uint64(binary.LittleEndian.Uint32(b[0x3c:0x40]))
	return offset >= 4 && offset <= 4096 && fileSpan(b, offset, 24) && bytes.Equal(b[offset:offset+4], []byte{'P', 'E', 0, 0})
}

// maxOverlaySegments bounds how many installer data blocks one overlay is
// split into. Each segment searches the rest of the overlay for installer
// markers again; setup programs have one or two.
const maxOverlaySegments = 16

// peOverlay analyzes installer data and archives stored after the image.
// Installer data can end before the overlay does, and what follows it is
// analyzed like the start of an overlay; data before it may hold a
// self-extractor's archive. It returns the overlay ranges that nothing
// recognized: they have only been matched as raw bytes of the executable.
func (s *scanState) peOverlay(overlay []byte, depth int) (sig string, unrecognized [][2]int, err error) {
	var incomplete error
	for pos, segment := 0, 0; pos < len(overlay); segment++ {
		if segment == maxOverlaySegments {
			return "", unrecognized, ErrLimit
		}
		rest := overlay[pos:]
		start, end, sig, err := s.installerData(rest, depth)
		if sig != "" {
			return sig, nil, nil
		}
		if err != nil {
			incomplete = err
		}
		if start < 0 {
			start, end = len(rest), len(rest)
		}
		gap := start
		if start > 0 {
			i, err := s.archiveSearch(rest[:start])
			if err != nil {
				incomplete = err
			}
			if i >= 0 {
				gap = i
				sig, err := s.visit(cabEmbeddedView(rest[i:start]), depth+1)
				if sig != "" {
					return sig, nil, nil
				}
				if err != nil {
					incomplete = err
				}
			}
		}
		if gap > 0 {
			unrecognized = append(unrecognized, [2]int{pos, pos + gap})
		}
		pos += max(end, start+1)
	}
	return "", unrecognized, incomplete
}

// installerData analyzes the first installer data in rest. start is where it
// begins, or -1 when there is none; end is where it ends, len(rest) when that
// is not known.
func (s *scanState) installerData(rest []byte, depth int) (start, end int, sig string, err error) {
	if i := bytes.Index(rest, nsisMagic); i >= 4 {
		sig, n, err := s.nsis(rest[i-4:], depth)
		return i - 4, i - 4 + n, sig, err
	}
	if i := bytes.Index(rest, autoItMagic); i >= 0 {
		// Compiled AutoIt scripts are encrypted and compressed; the script is
		// not decoded here.
		return i, len(rest), "", ErrUnsupported
	}
	if i := bytes.Index(rest, []byte("InstallShield\x00")); i >= 0 {
		s.layers.Push(signatures.Layer{Type: signatures.TypeIShieldMSI, Size: int64(len(rest) - i)})
		sig, n, ok, err := s.installShieldMSI(rest[i+14:], depth)
		s.layers.Pop()
		if ok {
			return i, i + 14 + n, sig, err
		}
	}
	if sig, n, ok, err := s.installShield(rest, depth); ok {
		return 0, n, sig, err
	}
	return -1, 0, "", nil
}

// archiveSearch returns the offset of the first self-extractor archive
// header in b, or -1. Checking an ARJ header candidate computes a CRC over up
// to 2600 bytes, and a gzip or bzip2 header is checked by decoding the start
// of its stream, so the number of candidates checked is bounded. Decoding
// probes together read at most len(b) input bytes, which keeps the search
// linear.
func (s *scanState) archiveSearch(b []byte) (int, error) {
	arjCandidates, probes, probed := 0, 0, 0
	for i := 0; i < len(b); i++ {
		if i&0xffff == 0 {
			if err := s.ctx.Err(); err != nil {
				return -1, err
			}
		}
		if b[i] == 0x60 {
			if !arjHeaderCandidate(b[i:]) {
				continue
			}
			if arjCandidates++; arjCandidates > maxEmbeddedCandidates {
				return -1, ErrLimit
			}
		}
		if !archiveStart(b[i:]) {
			continue
		}
		if b[i] != 0x1f && b[i] != 'B' {
			return i, nil
		}
		if probes++; probes > maxEmbeddedCandidates {
			return -1, ErrLimit
		}
		archive, n, err := s.archiveDecodes(b[i:], len(b)-probed)
		probed += n
		if err != nil {
			return -1, err
		}
		if archive {
			return i, nil
		}
	}
	return -1, nil
}

// archiveProbe is how much output a gzip or bzip2 header must start before
// the data after it is taken for an archive, unless the stream ends first.
const archiveProbe = 4 << 10

// archiveDecodes decodes the start of the gzip or bzip2 stream that b begins
// with. The header checks leave a gzip header about 27 bits long, which
// Authenticode data and symbol tables hold by chance: data that fails to
// decode right after its header is not an archive. Data that ends before the
// probe does is a truncated archive, which the caller analyzes and reports as
// incomplete. The probe reads at most budget input bytes and returns how many
// it read; needing more is a resource limit.
func (s *scanState) archiveDecodes(b []byte, budget int) (bool, int, error) {
	input := bytes.NewReader(b[:min(len(b), budget)])
	var decoded int64
	var err error
	if b[0] == 'B' {
		decoded, err = io.CopyN(io.Discard, contextReader{s.ctx, bzip2.NewReader(input)}, archiveProbe)
	} else if z, e := gzip.NewReader(input); e != nil {
		err = e
	} else {
		z.Multistream(false) // a complete member may be followed by other data
		decoded, err = io.CopyN(io.Discard, contextReader{s.ctx, z}, archiveProbe)
	}
	read := int(input.Size()) - input.Len()
	switch {
	case s.ctx.Err() != nil:
		return false, read, s.ctx.Err()
	case err == nil, err == io.EOF:
		return true, read, nil // the probe decoded, or the whole stream did
	case err == io.ErrUnexpectedEOF && budget < len(b):
		return false, read, ErrLimit
	case err == io.ErrUnexpectedEOF:
		return true, read, nil // a truncated archive
	case b[0] == 'B' && decoded > 0:
		// bzip2 emits nothing before a block's structure has decoded; a
		// complete stream followed by other data fails only after it.
		return true, read, nil
	}
	return false, read, nil
}

// archiveStart recognizes a self-extracting payload's archive header with
// more than its magic bytes, so that a two- or three-byte magic occurring by
// chance inside other installer data is not mistaken for an archive.
func archiveStart(b []byte) bool {
	if len(b) < 8 {
		return false
	}
	switch b[0] {
	case 'P':
		// ZIP local header: version needed, known method, bounded name.
		return len(b) >= 30 && bytes.HasPrefix(b, []byte("PK\x03\x04")) && b[5] == 0 &&
			binary.LittleEndian.Uint16(b[8:10]) <= 99 && binary.LittleEndian.Uint16(b[26:28]) <= 1024
	case 0x1f:
		return b[1] == 0x8b && b[2] == 8 && b[3]&0xe0 == 0
	case 'B':
		return len(b) >= 10 && bytes.HasPrefix(b, []byte("BZh")) && b[3] >= '1' && b[3] <= '9' &&
			(bytes.Equal(b[4:10], []byte{0x31, 0x41, 0x59, 0x26, 0x53, 0x59}) || bytes.Equal(b[4:10], []byte{0x17, 0x72, 0x45, 0x38, 0x50, 0x90}))
	case 'R':
		return len(b) >= 7 && bytes.HasPrefix(b, []byte("Rar!\x1a\x07")) && (b[6] == 0 || b[6] == 1)
	case '7':
		return bytes.HasPrefix(b, []byte("7z\xbc\xaf\x27\x1c")) && b[6] == 0
	case 'M':
		return len(b) >= 36 && bytes.HasPrefix(b, []byte("MSCF\x00\x00\x00\x00")) &&
			binary.LittleEndian.Uint16(b[26:28]) != 0 && binary.LittleEndian.Uint16(b[28:30]) != 0
	case 0x60:
		return arjHeader(b)
	}
	return false
}

var (
	nsisMagic   = []byte("\xef\xbe\xad\xdeNullsoftInst")
	autoItMagic = []byte("\xa3\x48\x4b\xbe\x98\x6c\x4a\xa9\x99\x4c\x53\x0a\x86\xd6\x48\x7d\x41\x55\x33\x21\x45\x41\x30")
)

// NSIS stores its script header and each file as a separate block, or all of
// them in one "solid" stream, compressed with LZMA, deflate or its own bzip2
// variant. The first header is flags, a signature, the magic, the unpacked
// script-header length and the length of all data including a trailing CRC
// (NSIS's own fileform.h describes the layout; the reference engine reads it
// the same way).
const (
	nsisFirstHeader = 28
	nsisNoCRC       = 4
)

// nsis analyzes the NSIS data at the start of data and returns its length,
// or len(data) when the header gives no usable length.
func (s *scanState) nsis(data []byte, depth int) (string, int, error) {
	if len(data) < nsisFirstHeader+4 {
		return "", len(data), ErrMalformed
	}
	total := uint64(binary.LittleEndian.Uint32(data[24:]))
	if total < nsisFirstHeader+4 || total > uint64(len(data)) {
		return "", len(data), ErrMalformed
	}
	// Like the reference, the installer data is a layer of its own, so its
	// files have the installer as their container.
	s.layers.Push(signatures.Layer{Type: signatures.TypeNullsoft, Size: int64(total)})
	sig, err := s.nsisData(data[:total], depth)
	s.layers.Pop()
	return sig, int(total), err
}

func (s *scanState) nsisData(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	if len(data) < nsisFirstHeader+4 {
		return "", ErrMalformed
	}
	flags := binary.LittleEndian.Uint32(data)
	headerSize := binary.LittleEndian.Uint32(data[20:])
	total := uint64(binary.LittleEndian.Uint32(data[24:]))
	end := total
	if flags&nsisNoCRC == 0 {
		end -= 4
	}
	if headerSize == 0 || total < nsisFirstHeader+4 || total > uint64(len(data)) || end < nsisFirstHeader {
		return "", ErrMalformed
	}
	body := data[nsisFirstHeader:end]
	// Blocks that exactly tile the data are separate; otherwise one solid
	// stream starts at the first byte.
	var blocks [][]byte
	var compressed []bool
	for pos := 0; pos < len(body); {
		if len(body)-pos < 4 {
			blocks = nil
			break
		}
		n := binary.LittleEndian.Uint32(body[pos:])
		size := uint64(n & 0x7fffffff)
		if size > uint64(len(body)-pos-4) {
			blocks = nil
			break
		}
		blocks = append(blocks, body[pos+4:pos+4+int(size)])
		compressed = append(compressed, n&0x80000000 != 0)
		pos += 4 + int(size)
		if len(blocks) > s.maxFiles-s.files {
			return "", ErrLimit
		}
	}
	if blocks == nil {
		return s.nsisSolid(body, depth)
	}
	var incomplete error
	for i, block := range blocks {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		var sig string
		var err error
		if compressed[i] {
			var out []byte
			out, err = s.nsisDecode(block)
			if err == nil {
				sig, err = s.visitHeld(out, depth+1, int64(len(out)))
			}
		} else {
			sig, err = s.visit(block, depth+1)
		}
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}

// nsisSolid decodes one stream holding the script header and every file,
// each preceded by its 32-bit length, and analyzes the entries in order.
func (s *scanState) nsisSolid(body []byte, depth int) (string, error) {
	r, err := s.nsisReader(body)
	if err != nil {
		return "", err
	}
	var incomplete error
	var length [4]byte
	for count := 0; ; count++ {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if _, err := io.ReadFull(r, length[:]); err != nil {
			if err == io.EOF {
				return "", incomplete // the stream ends after its last entry
			}
			if ctxErr := s.ctx.Err(); ctxErr != nil {
				return "", ctxErr
			}
			return "", ErrMalformed
		}
		if count >= s.maxFiles-s.files {
			return "", ErrLimit
		}
		size := int64(binary.LittleEndian.Uint32(length[:]) & 0x7fffffff)
		if size > min(s.maxFile, s.remaining) {
			return "", ErrLimit
		}
		entry, err := s.readChild(io.LimitReader(r, size), depth)
		if err == nil && int64(len(entry)) != size {
			s.spend(len(entry))
			err = ErrMalformed
		}
		if err != nil {
			return "", err
		}
		sig, err := s.visit(entry, depth+1)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
}

// nsisDecode decodes one separately compressed block within the scan budget.
// Decoded bytes are charged whether or not decoding succeeds.
func (s *scanState) nsisDecode(block []byte) ([]byte, error) {
	r, err := s.nsisReader(block)
	if err != nil {
		return nil, err
	}
	limit := min(s.maxFile, s.remaining)
	if limit < 0 {
		return nil, ErrLimit
	}
	out, err := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, limit+1))
	s.spend(len(out)) // held while analyzed, or discarded expansion work
	if ctxErr := s.ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}
	switch {
	case int64(len(out)) > limit:
		return nil, ErrLimit
	case err != nil:
		// NSIS ends LZMA data with the end marker; a block without one is
		// truncated.
		return nil, ErrMalformed
	}
	return out, nil
}

// lzmaMaxExpansion bounds LZMA output per input byte. The range coder spends
// at least 0.02 bits on a binary decision (11-bit probabilities that never
// pass 2017/2048), and the cheapest symbol, a repeated 273-byte match, takes
// 14 decisions: at most about 7,300 output bytes per input byte.
const lzmaMaxExpansion = 1 << 14

// nsisReader identifies the compression of an NSIS stream the way the NSIS
// decompressor does: LZMA properties, the NSIS bzip2 variant's block marker,
// or otherwise raw deflate.
func (s *scanState) nsisReader(b []byte) (io.Reader, error) {
	switch {
	case len(b) < 5:
		return nil, ErrMalformed
	case b[0] == 0x5d:
		dict := binary.LittleEndian.Uint32(b[1:5])
		if dict > maxCodecDictionary {
			return nil, ErrLimit
		}
		// The dictionary never needs to exceed the output the budget allows, or
		// the output this input can produce.
		limit := min(s.maxFile, s.remaining, int64(len(b))*lzmaMaxExpansion)
		dict = uint32(max(lzma.MinDictCap, min(int64(dict), limit)))
		header := make([]byte, 0, lzma.HeaderLen)
		header = append(header, b[0])
		header = binary.LittleEndian.AppendUint32(header, dict)
		header = binary.LittleEndian.AppendUint64(header, ^uint64(0))
		r, err := lzma.ReaderConfig{DictCap: maxCodecDictionary}.NewReader(io.MultiReader(bytes.NewReader(header), bytes.NewReader(b[5:])))
		if err != nil {
			return nil, ErrMalformed
		}
		return contextReader{s.ctx, r}, nil
	case b[0] == '1':
		return contextReader{s.ctx, nsisbzip2.NewReader(bytes.NewReader(b))}, nil
	}
	return contextReader{s.ctx, flate.NewReader(bytes.NewReader(b))}, nil
}

// installShield reads the overlay of older InstallShield setup programs: for
// each file a NUL-terminated name, source path and version, its size as
// decimal text, then its bytes. The overlay is recognized only when the first
// record is well formed. n is where the records end: the reference stops at
// the first record it cannot read, and the caller analyzes what follows as
// other overlay data. A record whose size runs past the data was cut short.
func (s *scanState) installShield(data []byte, depth int) (sig string, n int, recognized bool, err error) {
	var incomplete error
	pos := 0
	for count := 0; pos < len(data); count++ {
		if err := s.ctx.Err(); err != nil {
			return "", len(data), count > 0, err
		}
		var fields [4][]byte
		next := pos
		ok := true
		for i := range fields {
			n := bytes.IndexByte(data[next:min(len(data), next+2048)], 0)
			if n < 0 {
				ok = false
				break
			}
			fields[i] = data[next : next+n]
			next += n + 1
		}
		size, err := strconv.ParseUint(string(fields[3]), 10, 63)
		if !ok || err != nil || len(fields[0]) == 0 || !installShieldRecord(fields[0], fields[1], fields[2]) {
			if count == 0 {
				return "", 0, false, nil
			}
			break
		}
		if size > uint64(len(data)-next) {
			if count == 0 {
				return "", 0, false, nil
			}
			return "", len(data), true, ErrMalformed
		}
		if count >= s.maxFiles-s.files {
			return "", len(data), true, ErrLimit
		}
		sig, err := s.visit(data[next:next+int(size)], depth+1)
		if sig != "" {
			return sig, 0, true, nil
		}
		if err != nil {
			incomplete = err
		}
		pos = next + int(size)
	}
	return "", pos, true, incomplete
}

// installShieldRecord requires a printable file name without a directory, a
// printable source path and a dotted numeric version, as setup programs write.
func installShieldRecord(name, path, version []byte) bool {
	if !printable(name) || !printable(path) || bytes.ContainsAny(name, "/\\:") || len(version) == 0 {
		return false
	}
	for _, c := range version {
		if (c < '0' || c > '9') && c != '.' {
			return false
		}
	}
	return true
}

func printable(b []byte) bool {
	for _, c := range b {
		if c < 0x20 || c == 0x7f {
			return false
		}
	}
	return true
}

// InstallShield MSI setup programs follow the "InstallShield" marker with a
// file count, zero padding and per-file records: a 260-byte name, fields, a
// 64-bit stored size, and zlib data obfuscated by swapping each byte's nibbles
// and XORing it with the file name, itself XORed with a fixed key.
const installShieldMSIRecord = 0x138

var installShieldMSIKey = [4]byte{0xec, 0xca, 0x79, 0xf8}

func (s *scanState) installShieldMSI(data []byte, depth int) (sig string, n int, recognized bool, err error) {
	if len(data) < 0x20+installShieldMSIRecord || !allZeroOLE(data[8:0x20]) {
		return "", 0, false, nil
	}
	count := uint64(binary.LittleEndian.Uint32(data))
	first := data[0x20 : 0x20+installShieldMSIRecord]
	name := bytes.IndexByte(first[:0x104], 0)
	if count == 0 || name <= 0 || !printable(first[:name]) {
		return "", 0, false, nil
	}
	// A count beyond what the data can hold means the data was cut short;
	// the records present are analyzed and the rest is malformed below.
	if count > uint64(len(data)-0x20)/installShieldMSIRecord {
		count = uint64(len(data)-0x20)/installShieldMSIRecord + 1
	}
	if count > uint64(s.maxFiles-s.files) {
		return "", len(data), true, ErrLimit
	}
	var incomplete error
	pos := 0x20
	for i := uint64(0); i < count; i++ {
		if err := s.ctx.Err(); err != nil {
			return "", len(data), true, err
		}
		if len(data)-pos < installShieldMSIRecord {
			return "", len(data), true, ErrMalformed
		}
		record := data[pos : pos+installShieldMSIRecord]
		pos += installShieldMSIRecord
		nameLen := bytes.IndexByte(record[:0x104], 0)
		size := binary.LittleEndian.Uint64(record[0x10c:])
		if nameLen <= 0 || size > uint64(len(data)-pos) {
			return "", len(data), true, ErrMalformed
		}
		key := make([]byte, nameLen)
		for j := range key {
			key[j] = record[j] ^ installShieldMSIKey[j&3]
		}
		stored := data[pos : pos+int(size)]
		pos += int(size)
		sig, err := s.installShieldMSIFile(stored, key, depth)
		if sig != "" {
			return sig, 0, true, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", pos, true, incomplete
}

func (s *scanState) installShieldMSIFile(stored, key []byte, depth int) (string, error) {
	if int64(len(stored)) > min(s.maxFile, s.remaining) {
		return "", ErrLimit
	}
	plain := make([]byte, len(stored))
	for i, c := range stored {
		plain[i] = (c>>4 | c<<4) ^ key[(i&0x3ff)%len(key)]
	}
	s.spend(len(plain)) // the deobfuscated copy is held while it is decoded
	r, err := zlib.NewReader(bytes.NewReader(plain))
	if err != nil {
		return "", ErrMalformed
	}
	defer r.Close()
	return s.child(r, depth)
}
