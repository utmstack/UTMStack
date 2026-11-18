package engine

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/utmstack/UTMStack/agent/edr/lzx"
	"github.com/utmstack/UTMStack/agent/edr/quantum"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// CAB is defined by Microsoft's MS-CAB and MS-MCI specifications. ClamAV uses
// libmspack for its wider CAB support; this decoder intentionally admits only
// single-cabinet stored, MSZIP, Quantum and LZX folders, salvages available
// members and reports unfinished analysis instead of treating it as clean.
// https://download.microsoft.com/download/4/D/A/4DA14F27-B4EF-4170-A6E6-5B1EF85B1BAA/%5BMS-CAB%5D.pdf
// https://learn.microsoft.com/en-us/openspecs/exchange_server_protocols/ms-mci/3048c471-f739-4d30-bcb3-a962cfe84a69
// https://github.com/Cisco-Talos/clamav/blob/main/libclamav/libmspack.c
type cabFolder struct {
	start, blocks int
	compression   uint16
}
type cabFile struct {
	start, size uint32
	folder      int
	name        string
}

type cabDecodedFolder struct {
	data          []byte
	held          int64
	remainingUses int
	loaded        bool
	needsPrevious bool
	requestedEnd  uint64
	err           error
}

// cabString follows libmspack's bounded, byte-preserving CAB string reader.
// Header disk labels may be empty; cabinet and member names may not.
func cabString(data []byte, pos *int, allowEmpty bool) (string, bool) {
	if *pos < 0 || *pos >= len(data) {
		return "", false
	}
	end := min(*pos+256, len(data))
	n := bytes.IndexByte(data[*pos:end], 0)
	start := *pos
	*pos = end // a failed reference read still consumes its bounded buffer
	if n < 0 || n == 0 && !allowEmpty {
		return "", false
	}
	name := string(data[start : start+n])
	*pos = start + n + 1
	return name, true
}

func (s *scanState) cab(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth || len(data) < 36 {
		if depth >= s.maxDepth {
			return "", ErrLimit
		}
		return "", ErrMalformed
	}
	if !bytes.Equal(data[:8], []byte("MSCF\x00\x00\x00\x00")) {
		return "", ErrMalformed
	}
	// The reference's root reader ignores the other reserved fields, version
	// and declared cabinet length. Only physical input bounds govern reads;
	// an attacker-controlled length must neither hide a child nor allocate.
	length := uint64(binary.LittleEndian.Uint32(data[8:12]))
	archiveSize := int64(len(data)) // includes bytes following cbCabinet, as the reference layer does
	countFolders := int(binary.LittleEndian.Uint16(data[26:28]))
	countFiles := int(binary.LittleEndian.Uint16(data[28:30]))
	if countFolders == 0 || countFiles == 0 {
		return "", ErrMalformed
	}
	// Each decoded folder is an object held alongside the files it contains.
	if countFolders+countFiles > s.maxFiles-s.files {
		return "", ErrLimit
	}
	flags := binary.LittleEndian.Uint16(data[30:32])
	var incomplete error
	pos, folderReserve, blockReserve := 36, 0, 0
	if flags&4 != 0 {
		if len(data)-pos < 4 {
			return "", ErrMalformed
		}
		headerReserve := int(binary.LittleEndian.Uint16(data[pos : pos+2]))
		folderReserve, blockReserve = int(data[pos+2]), int(data[pos+3])
		pos += 4
		if headerReserve > len(data)-pos {
			return "", ErrMalformed
		}
		pos += headerReserve
	}
	for _, flag := range []uint16{1, 2} {
		if flags&flag != 0 {
			if _, ok := cabString(data, &pos, false); !ok {
				return "", ErrMalformed
			}
			if _, ok := cabString(data, &pos, true); !ok {
				return "", ErrMalformed
			}
		}
	}
	if countFolders > (len(data)-pos)/(8+folderReserve) {
		return "", ErrMalformed
	}
	folders := make([]cabFolder, countFolders)
	for i := range folders {
		folders[i] = cabFolder{
			start:       int(binary.LittleEndian.Uint32(data[pos : pos+4])),
			blocks:      int(binary.LittleEndian.Uint16(data[pos+4 : pos+6])),
			compression: binary.LittleEndian.Uint16(data[pos+6 : pos+8]),
		}
		pos += 8 + folderReserve
	}
	// ClamAV 1.5.4's bundled reader takes CFFILE records immediately after
	// the folders, even when the header's coffFiles field disagrees.
	filePos := pos
	files := make([]cabFile, 0, countFiles)
	cache := make([]cabDecodedFolder, countFolders)
	for i := 0; i < countFiles; i++ {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if len(data)-filePos < 16 {
			return "", ErrMalformed // a short fixed record rejects the entire table
		}
		file := cabFile{
			size:   binary.LittleEndian.Uint32(data[filePos : filePos+4]),
			start:  binary.LittleEndian.Uint32(data[filePos+4 : filePos+8]),
			folder: int(binary.LittleEndian.Uint16(data[filePos+8 : filePos+10])),
		}
		filePos += 16
		if file.folder >= 0xfffd {
			fromPrevious := file.folder != 0xfffe
			file.folder = countFolders - 1
			if fromPrevious {
				file.folder = 0
				cache[0].needsPrevious = true
			}
		}
		// Merge state precedes name validation in the reference: even a
		// skipped unnamed entry can require unavailable predecessor history.
		var ok bool
		file.name, ok = cabString(data, &filePos, false)
		if !ok {
			incomplete = ErrMalformed
			continue // invalid names are skipped after the bounded read
		}
		if file.folder >= countFolders {
			incomplete = ErrMalformed
			continue // invalid folder entries are skipped in reference salvage mode
		}
		files = append(files, file)
		cache[file.folder].remainingUses++
		cache[file.folder].requestedEnd = max(cache[file.folder].requestedEnd, uint64(file.start)+uint64(file.size))
	}
	if len(files) == 0 {
		return "", ErrMalformed
	}
	// Decode referenced folders lazily, once per folder, while visiting files
	// in CFFILE order. The shared compressed-byte allowance still bounds
	// repeated work when different folders alias the same CFDATA region.
	input := len(data)
	extent := filePos
	for index, file := range files {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		entry := &cache[file.folder]
		sig, err := func() (string, error) {
			defer func() {
				entry.remainingUses--
				if entry.remainingUses == 0 {
					entry.data = nil // release a folder after its last member
				}
			}()
			meta := signatures.MemberMeta{Name: file.name, HasName: true, Size: int64(file.size), Index: index}
			if name, err := s.db.MatchMember(s.ctx, s.layers.TypeAt(-1), archiveSize, meta); name != "" {
				return name, nil
			} else if err != nil {
				if ctxErr := s.ctx.Err(); ctxErr != nil {
					return "", ctxErr
				}
				incomplete = err // an undecided name predicate must not hide a child
			}
			if entry.needsPrevious {
				return "", ErrUnsupported
			}
			if file.size == 0 {
				return "", nil
			}
			if !entry.loaded {
				entry.loaded = true
				folder := folders[file.folder]
				switch {
				case folder.start < filePos || folder.start > len(data):
					entry.err = ErrMalformed
				case s.files >= s.maxFiles:
					entry.err = ErrLimit
				default:
					entry.data, entry.err = s.cabFolder(data, folder, blockReserve, &input, entry.requestedEnd, &extent)
					if entry.err == nil || len(entry.data) != 0 {
						s.files++
						entry.held = int64(len(entry.data))
						s.remaining -= entry.held
					}
				}
			}
			if err := s.ctx.Err(); err != nil {
				return "", err
			}
			if stopped, ok := entry.err.(*cabMSZIPInputError); ok && uint64(file.start) > stopped.frameStart {
				return "", entry.err // the reference fails while seeking, before writing this member
			}
			end := uint64(file.start) + uint64(file.size)
			memberErr := entry.err
			if end > uint64(len(entry.data)) {
				end = uint64(len(entry.data))
				if memberErr == nil {
					memberErr = ErrMalformed
				}
				if folders[file.folder].compression&0xf == 2 {
					// A failed Quantum extraction only writes completed window
					// segments. Bytes buffered in its unfinished request are not
					// an extracted child (unlike a completed shorter member).
					window := uint64(1) << (folders[file.folder].compression >> 8 & 0x1f)
					end = end / window * window
				}
			}
			if uint64(file.start) >= end {
				if memberErr == nil {
					memberErr = ErrMalformed
				}
				return "", memberErr
			}
			prepaid := min(entry.held, int64(end-uint64(file.start)))
			entry.held -= prepaid
			if name, err := s.visitHeld(entry.data[int(file.start):int(end)], depth+1, prepaid); name != "" || err != nil {
				return name, err
			}
			return "", memberErr
		}()
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			if ctxErr := s.ctx.Err(); ctxErr != nil {
				return "", ctxErr
			}
			incomplete = err
		}
	}
	// A short or zero advisory length must not rescan header/data bytes as
	// another archive. Track bounded CFDATA spans even when decoding aborts;
	// a real suffix beyond both extents still receives ordinary scan limits.
	trailing := data[max(int(min(length, uint64(len(data)))), extent):]
	if len(trailing) != 0 {
		if sig, err := s.visit(trailing, depth+1); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		} else if incomplete == nil {
			incomplete = ErrUnsupported
		}
	}
	return "", incomplete
}

// cabFolder decodes one folder into at most min(maxFile, remaining) bytes. Each
// block's header and compressed bytes are deducted from input, the cabinet's
// shared allowance of block bytes that may still be decoded. A valid prefix
// survives a later failure, paired with that error; the caller charges and
// scans it. Failed codec output that cannot be trusted is charged here.
func (s *scanState) cabFolder(data []byte, folder cabFolder, reserve int, input *int, requestedEnd uint64, extent *int) (_ []byte, err error) {
	method := folder.compression & 0xf
	window := int(folder.compression >> 8 & 0x1f)
	if method == 3 {
		// FCI's TCOMPfromLZXWindow stores a 15..21 bit window in bits 8..12.
		if window < 15 || window > 21 || folder.compression != uint16(window<<8)|3 {
			return nil, ErrUnsupported
		}
	} else if method == 2 {
		// The Quantum level bits do not affect decoding; libmspack only
		// consults the 10..21 bit window in this method's parameters.
		if window < 10 || window > 21 {
			return nil, ErrUnsupported
		}
	} else if folder.compression != 0 && folder.compression != 1 {
		return nil, ErrUnsupported
	}
	maxOutput := min(s.maxFile, s.remaining)
	if maxOutput < 0 {
		return nil, ErrLimit
	}
	if method == 1 {
		return s.cabMSZIP(data, folder, reserve, input, requestedEnd, maxOutput, extent)
	}
	pos := folder.start
	var out, packed []byte
	var quantumBlocks [][]byte
	var totalOutput int64
	blockErr := func() error {
		for i := 0; i < folder.blocks; i++ {
			if err := s.ctx.Err(); err != nil {
				return err
			}
			if len(data)-pos < 8+reserve {
				return ErrMalformed
			}
			header := data[pos : pos+8]
			compressed := int(binary.LittleEndian.Uint16(header[4:6]))
			uncompressed := int(binary.LittleEndian.Uint16(header[6:8]))
			pos += 8 + reserve
			recordCABExtent(extent, pos)
			// ClamAV enables salvage: both CFDATA lengths may use their full
			// 16-bit range, and checksum failures do not suppress scanning.
			if compressed > len(data)-pos {
				return ErrMalformed
			}
			recordCABExtent(extent, pos+compressed)
			if uncompressed == 0 {
				return ErrUnsupported // block continues into another cabinet
			}
			blockOutput := uncompressed
			if method == 0 {
				blockOutput = compressed // stored output ignores the declared size
			} else if method == 2 {
				blockOutput = 0 // Quantum ignores the declared size
			}
			if int64(blockOutput) > maxOutput-totalOutput || 8+reserve+compressed > *input {
				return ErrLimit
			}
			totalOutput += int64(blockOutput)
			*input -= 8 + reserve + compressed
			block := data[pos : pos+compressed]
			pos += compressed
			if method == 3 {
				// One folder is one continuous LZX stream. CFDATA headers are
				// framing only; resetting the decoder here loses its dictionary.
				packed = append(packed, block...)
				continue
			}
			if method == 2 {
				quantumBlocks = append(quantumBlocks, block)
				continue
			}
			if folder.compression == 0 {
				out = append(out, block...)
				continue
			}
		}
		return nil
	}()
	if ctxErr := s.ctx.Err(); ctxErr != nil {
		return out, ctxErr
	}
	if method == 3 {
		if blockErr != nil {
			// Without the last CFDATA block the reference does not know a
			// short final LZX frame's size. Only whole frames can be emitted.
			totalOutput = totalOutput / 32768 * 32768
		}
		var produced int
		out, produced, err = lzx.DecodeValidatedContext(s.ctx, packed, window, int(totalOutput), 0)
		s.spend(min(produced-len(out), int(maxOutput)-len(out)))
		if err != nil {
			if ctxErr := s.ctx.Err(); ctxErr != nil {
				return out, ctxErr
			}
			return out, fmt.Errorf("CAB LZX: %v: %w", err, ErrMalformed)
		}
	}
	if method == 2 {
		// Quantum's caller requests file bytes; CFDATA output lengths do not
		// terminate the arithmetic stream. Padding can encode more symbols.
		target := min(requestedEnd, uint64(maxOutput))
		if requestedEnd > uint64(maxOutput) && blockErr == nil {
			blockErr = ErrLimit
		}
		var produced int
		out, produced, err = quantum.DecodeValidatedContext(s.ctx, quantumBlocks, window, int(target))
		s.spend(min(produced-len(out), int(maxOutput)-len(out)))
		if err != nil {
			if ctxErr := s.ctx.Err(); ctxErr != nil {
				return out, ctxErr
			}
			return out, fmt.Errorf("CAB Quantum: %v: %w", err, ErrMalformed)
		}
	}
	return out, blockErr
}

func recordCABExtent(extent *int, end int) {
	if extent != nil {
		*extent = max(*extent, end)
	}
}

// Embedded CABs differ from root CABs: ClamAV opens the header in the
// containing object, then scans a nested view bounded by cbCabinet. A zero
// length means the rest of that object, and oversized lengths are clamped
// to available bytes. Keep that view identical in raw and PE-overlay paths.
func cabEmbeddedView(data []byte) []byte {
	if len(data) >= 36 && bytes.HasPrefix(data, []byte("MSCF\x00\x00\x00\x00")) {
		size := uint64(binary.LittleEndian.Uint32(data[8:12]))
		if size != 0 && size < uint64(len(data)) {
			return data[:int(size)]
		}
	}
	return data
}
