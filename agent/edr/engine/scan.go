package engine

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"github.com/utmstack/UTMStack/agent/edr/sevenzip"
	"hash/crc32"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

var (
	ErrLimit       = errors.New("UTMStack EDR: analysis resource limit reached")
	ErrUnsupported = errors.New("UTMStack EDR: unsupported or encrypted content")
	ErrMalformed   = errors.New("UTMStack EDR: malformed content")
)

// IncompleteReason names why an analysis is incomplete, for operators:
// "Unsupported.<analysis>" when a rule or representation this engine does not
// evaluate could apply, otherwise ResourceLimit, Unsupported, Malformed,
// Timeout, Unavailable or Error.
func IncompleteReason(err error) string {
	var unknown *signatures.UnknownError
	switch {
	case errors.As(err, &unknown):
		return "Unsupported." + unknown.Reason
	case errors.Is(err, ErrLimit), errors.Is(err, signatures.ErrMatchLimit):
		return "ResourceLimit"
	case errors.Is(err, ErrUnsupported):
		return "Unsupported"
	case errors.Is(err, ErrMalformed):
		return "Malformed"
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return "Timeout"
	case errors.Is(err, ErrUnavailable):
		return "Unavailable"
	}
	return "Error"
}

// Scan never reports clean when a recognized container cannot be fully scanned.
// These errors are unknown verdicts, NOT malware detections; the existing AMSI
// fail policy decides whether incomplete analysis blocks script execution.
func (e *Engine) Scan(ctx context.Context, data []byte) (bool, string, error) {
	clean, sig, _, err := e.ScanSnapshot(ctx, data)
	return clean, sig, err
}
func (e *Engine) ScanSnapshot(ctx context.Context, data []byte) (bool, string, string, error) {
	return e.ScanSnapshotWithPath(ctx, data, "")
}

// ScanSnapshotWithPath scans the bytes read from path while making the root
// file name and path available to metadata rules. Callers must pass the path
// of the same file snapshot; empty path retains buffer-scan semantics.
func (e *Engine) ScanSnapshotWithPath(ctx context.Context, data []byte, path string) (bool, string, string, error) {
	t := e.Tuning()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(t.MaxScanTimeMS)*time.Millisecond)
	defer cancel()
	select {
	case e.slots <- struct{}{}:
		defer func() { <-e.slots }()
	case <-ctx.Done():
		return false, "", "", ctx.Err()
	}
	g := e.current.Load()
	if g == nil {
		return false, "", "", ErrUnavailable
	}
	version := e.versionFor(g)
	s := scanState{ctx: ctx, db: g.db, maxFile: int64(t.MaxFileSizeMB) << 20, remaining: int64(t.MaxScanSizeMB) << 20, maxFiles: t.MaxFiles, maxDepth: t.MaxRecursion, maxPartitions: t.MaxPartitions,
		pdfNameHeuristicOff: !e.cfg.ScanHeuristics.PDFObfuscatedNameOn(), partitionIntersectionHeuristicOn: e.cfg.ScanHeuristics.PartitionIntersectionOn(),
		pePermissive: !e.cfg.ScanHeuristics.BrokenExecutableOn(), collectMetadata: e.cfg.CollectMetadata, metadataPath: path}
	sig, err := s.visit(data, 0)
	if sig == "" && s.rootEligible && ctx.Err() == nil && e.cfg.CollectMetadata &&
		(s.preclassFlags != nil || s.db.HasPreclassHooks() || s.db.HasLegacyMetadataRules()) {
		normalReliability := uint32(0)
		metadata, metadataErr := finishRootMetadata(ctx, data, s.rootType, s.metadataRoot)
		if metadataErr != nil && err == nil {
			err = metadataErr
		}
		if metadata != nil && ctx.Err() == nil {
			s.preclassPhase = true
			s.layers.Push(signatures.Layer{Type: s.rootType, Size: int64(len(data))})
			preSig, preErr := s.db.MatchPreclass(ctx, data, s.preclassFlags, signatures.MatchContext{
				Metadata: metadata, MetadataPartial: true,
				FileReliability:    &normalReliability,
				ScanExtractedTyped: s.bytecodeChildScanner(0), BytecodeOptions: s.bytecodeEngineOptions(),
			})
			if preSig != "" {
				sig, err = preSig, nil
			} else if err == nil {
				err = preErr
			}
			if sig == "" && s.db.HasLegacyMetadataRules() {
				// The reference runs root 13 only inside an active metadata
				// phase, after BC_PRECLASS, on the JSON bytes as a new layer.
				// Only childless plain-binary and ASCII-text roots have a
				// proved complete scalar tree here.
				complete := (s.rootType == signatures.TypeBinaryData || s.rootType == signatures.TypeTextASCII) && s.files == 1 && err == nil
				var serialized []byte
				if complete {
					serialized = plainScalarMetadataJSON(metadata)
					complete = serialized != nil
				}
				var legacySig string
				var legacyErr error
				if complete {
					if done, admitErr := s.admitObject(serialized, 1, 0); done {
						legacyErr = admitErr
					}
				}
				if legacyErr == nil {
					// The nested JSON scan suspends collection, but the existing
					// root and JSON handles remain active in native ClamAV.
					s.metadataSuspended = true
					legacySig, legacyErr = s.db.MatchLegacyMetadata(ctx, serialized, complete, signatures.MatchContext{
						Metadata: metadata, MetadataPartial: true,
						Layers: &s.layers, ScanExtractedTyped: s.bytecodeChildScanner(1), BytecodeOptions: s.bytecodeEngineOptions(),
					})
					s.metadataSuspended = false
				}
				if legacySig != "" {
					sig, err = legacySig, nil
				} else if err == nil {
					err = legacyErr
				}
			}
			s.layers.Pop()
		}
	}
	return sig == "" && err == nil, sig, version, err
}

type scanState struct {
	ctx                       context.Context
	db                        *signatures.Database
	maxFile, remaining        int64
	files, maxFiles, maxDepth int
	maxPartitions             int
	// layers is the recursion stack of the objects being analyzed.
	layers            signatures.Ancestry
	rootEligible      bool
	rootType          signatures.FileType
	preclassFlags     map[*bytecode.Program]bool
	preclassPhase     bool
	collectMetadata   bool
	metadataPath      string
	metadataRoot      map[string]any
	metadataSuspended bool
	// A policy switch affects only the heuristic verdict. Signature matching
	// and format analysis still run when it is disabled.
	pdfNameHeuristicOff              bool
	partitionIntersectionHeuristicOn bool
	pePermissive                     bool
	// ClamAV suppresses a second HandlerType retype of the same bytes.
	handlerHash [16]byte
}

// objectMetadata returns the scan-wide root, even while a child is matched.
// The native JSON handle table is initialized from ctx->metadata_json, never
// ctx->this_layer_metadata_json. Missing parser and child fields stay unknown.
func (s *scanState) objectMetadata() map[string]any {
	return s.representationMetadata()
}

// Derived views and parser hooks share the scan-wide JSON activation state,
// but their exact metadata tree is not yet proved. Keep queries for missing
// fields undecided. Suspending collection for the legacy JSON pass does not
// deactivate a root that already exists.
func (s *scanState) representationMetadata() map[string]any {
	if !s.collectMetadata && s.metadataRoot == nil {
		return nil
	}
	if s.metadataRoot == nil {
		s.metadataRoot = map[string]any{"Magic": "CLAMJSONv0"}
	}
	return s.metadataRoot
}

func (s *scanState) preclassMatched(p *bytecode.Program) {
	if s.preclassPhase || s.layers.PhysicalDepth() != 1 {
		return
	}
	if s.preclassFlags == nil {
		s.preclassFlags = make(map[*bytecode.Program]bool)
	}
	s.preclassFlags[p] = true
}

func (s *scanState) visit(data []byte, depth int) (string, error) {
	return s.visitObject(data, depth, objectOptions{})
}

// visitHeld analyzes data of which the caller already charged prepaid bytes to
// the scan budget, because it held them in memory while analyzing other data.
func (s *scanState) visitHeld(data []byte, depth int, prepaid int64) (string, error) {
	return s.visitObject(data, depth, objectOptions{prepaid: prepaid})
}

// htmlMode says whether phishing link analysis runs on an object.
type htmlMode uint8

const (
	htmlDetect   htmlMode = iota // when the content is recognized as HTML (htmlDocument)
	htmlDeclared                 // the container declares it HTML, as a text/html mail part
	htmlNever                    // a normalized representation of HTML already inspected
)

// minAnalyzedObject matches the reference engine, which never analyzes objects
// of five bytes or less: no hashes, no patterns, no format parsing.
const minAnalyzedObject = 6

// spend charges the aggregate budget for decoded bytes that were discarded
// because their decoder failed or exceeded a limit. Producing them was
// expansion work; uncharged, every later object could again expand to the
// full per-object limit.
func (s *scanState) spend(n int) {
	s.remaining -= min(int64(n), max(0, s.remaining))
}

func (s *scanState) visitRepresentation(data []byte, depth int, mode htmlMode, prepaid int64) (string, error) {
	return s.visitObject(data, depth, objectOptions{mode: mode, prepaid: prepaid})
}

// analyze runs format-specific decoding and recursion after raw matching.
// parsedMail is set when the object was already parsed as a message.
func (s *scanState) analyze(data []byte, depth int, mode htmlMode, parsedMail bool, parser parserFacts) (string, error) {
	// Detection runs first, allowing known hashes to identify encrypted or
	// unsupported objects without pretending to have decoded their contents.
	if kind := cpioFormat(data); kind != cpioNone {
		// The binary cpio magic is two bytes; a first header that does not
		// parse leaves the object to the analysis below.
		if sig, recognized, err := s.cpio(data, kind, depth); recognized {
			return sig, err
		}
	}
	switch {
	case isoHeader(data):
		return s.iso9660(data, depth)
	case lhaHeader(data):
		return s.lha(data, depth)
	case alzHeader(data):
		return s.alz(data, depth)
	case eggHeader(data):
		return s.egg(data, depth)
	case bytes.HasPrefix(data, arMagic), bytes.HasPrefix(data, arThinMagic):
		return s.arArchive(data, depth)
	case bytes.HasPrefix(data, []byte("PK\x03\x04")), bytes.HasPrefix(data, []byte("PK\x05\x06")), bytes.HasPrefix(data, []byte("PK\x07\x08")):
		return s.zip(data, depth)
	case bytes.HasPrefix(data, []byte{0x1f, 0x8b}):
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return "", ErrMalformed
		}
		defer r.Close()
		return s.partialChild(r, depth)
	case bytes.HasPrefix(data, []byte("BZh")):
		return s.bzip2Member(data, depth)
	case bytes.HasPrefix(data, []byte{0xfd, '7', 'z', 'X', 'Z', 0}):
		return s.xz(data, depth)
	case bytes.HasPrefix(data, []byte{0x28, 0xb5, 0x2f, 0xfd}), zstdSkippableFrame(data):
		return s.zstd(data, depth)
	case bytes.HasPrefix(data, []byte("Rar!\x1a\x07")):
		return s.rar(data, depth)
	case bytes.HasPrefix(data, []byte("MSCF")):
		return s.cab(data, depth)
	case bytes.HasPrefix(data, []byte("xar!")):
		return s.xar(data, depth)
	case bytes.HasPrefix(data, []byte{0x78, 0x9f, 0x3e, 0x22}):
		return s.tnef(data, depth)
	case len(data) > 262 && string(data[257:262]) == "ustar":
		return s.tar(data, depth)
	case bytes.HasPrefix(data, szddMagic):
		return s.szdd(data, depth)
	case bytes.HasPrefix(data, kwajMagic):
		return s.kwaj(data, depth)
	case bytes.HasPrefix(data, []byte("MZ")):
		return s.pe(data, depth, parser)
	case bytes.HasPrefix(data, []byte{0x7f, 'E', 'L', 'F'}):
		return "", validateELF(data)
	case bytes.HasPrefix(data, oleMagic):
		return s.ole(data, depth)
	case bytes.HasPrefix(data, []byte("%PDF-")):
		return s.pdf(data, depth)
	case bytes.HasPrefix(data, []byte("FWS")), bytes.HasPrefix(data, []byte("CWS")), bytes.HasPrefix(data, []byte("ZWS")):
		return s.swf(data, depth)
	case bytes.HasPrefix(data, []byte(`{\rt`)):
		return s.rtf(data, depth)
	case arjHeader(data):
		return s.arj(data, depth)
	case bytes.HasPrefix(data, sevenzip.Signature):
		return s.sevenZip(data, depth)
	case chmHeader(data):
		return s.chm(data, depth)
	case sisHeader(data):
		return s.sis(data, depth)
	case unsupportedMagic(data):
		return "", ErrUnsupported
	}
	if !parsedMail {
		if isMail, err := looksLikeMail(data); err != nil {
			return "", err
		} else if isMail {
			return s.mail(data, depth)
		}
	}
	// Text formats below may each hold content; an incomplete one does not
	// stop the others from being analyzed.
	var incomplete error
	result := func(sig string, err error) (string, error) {
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			return "", err
		}
		return "", incomplete
	}
	if mboxPostmark(data[:min(len(data), 1024)]) {
		sig, _, err := s.mbox(data, depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	if start := binhexStart(data); start >= 0 {
		sig, err := s.binhex(data, start, depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	if bytes.Contains(data, []byte("begin ")) || bytes.Contains(data, []byte("begin-base64 ")) {
		if sig, err := s.uuencoded(data, depth); sig != "" || err != nil {
			return sig, err
		}
	}
	// HTML fragments need no <html> declaration. The bounded tokenizer can
	// inspect links after arbitrary prose, and does nothing without phishing
	// rules. Link analysis runs on HTML only: other text and binary data that
	// merely hold a '<' are not documents whose links a reader follows.
	if mode != htmlNever && bytes.IndexByte(data, '<') >= 0 {
		if mode == htmlDeclared || htmlDocument(data) {
			if sig, err := s.db.MatchHTML(s.ctx, data); sig != "" {
				return sig, nil
			} else if err != nil {
				// Uncertain link analysis must not hide independently decodable
				// data URI content. Retain the error unless another view detects.
				incomplete = err
			}
		}
		if sig, err := s.dataURIs(data, depth); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		}
	}
	return result("", nil)
}

// unsupportedMagic recognizes formats whose contents are not decoded here,
// with enough of their headers that ordinary data is not mistaken for them.
func unsupportedMagic(b []byte) bool {
	switch {
	case bytes.HasPrefix(b, []byte("ISc(")):
		return true // InstallShield cabinet
	}
	return false
}

// sisHeader recognizes Symbian installation packages: the EPOC release 3-5
// or 6 UIDs, or the Symbian OS 9 SISX container UID.
func sisHeader(b []byte) bool {
	if len(b) < 16 {
		return false
	}
	uid1 := binary.LittleEndian.Uint32(b)
	uid2 := binary.LittleEndian.Uint32(b[4:])
	uid3 := binary.LittleEndian.Uint32(b[8:])
	return uid3 == 0x10000419 && (uid2 == 0x1000006d || uid2 == sisRelease6) || uid1 == sisX
}

// arjHeader recognizes an ARJ archive by its first header's CRC-32.
func arjHeader(b []byte) bool {
	if !arjHeaderCandidate(b) {
		return false
	}
	size := int(binary.LittleEndian.Uint16(b[2:]))
	return crc32.ChecksumIEEE(b[4:4+size]) == binary.LittleEndian.Uint32(b[4+size:])
}

// arjHeaderCandidate checks an ARJ header's magic and size, everything but
// its CRC-32.
func arjHeaderCandidate(b []byte) bool {
	if len(b) < 8 || b[0] != 0x60 || b[1] != 0xea {
		return false
	}
	size := int(binary.LittleEndian.Uint16(b[2:]))
	return size >= 30 && size <= 2600 && len(b) >= 4+size+4
}
func (s *scanState) child(r io.Reader, depth int) (string, error) {
	data, err := s.readChild(r, depth)
	if err != nil {
		return "", err
	}
	return s.visit(data, depth+1)
}

// partialChild scans bytes a decoder produced even when a late checksum or
// footer error follows them. Discarding those bytes could miss a signature
// on a damaged but decodable stream. An undetected damaged stream remains
// incomplete. The pinned reference does this for gzip; XZ gains coverage.
func (s *scanState) partialChild(r io.Reader, depth int) (string, error) {
	max, err := s.childBudget(depth)
	if err != nil {
		return "", err
	}
	data, readErr := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, max+1))
	if err := s.ctx.Err(); err != nil {
		s.spend(len(data))
		return "", err
	}
	if int64(len(data)) > max {
		s.spend(len(data))
		return "", ErrLimit
	}
	if readErr != nil && len(data) < minAnalyzedObject {
		s.spend(len(data))
		return "", ErrMalformed
	}
	name, childErr := s.visit(data, depth+1)
	if name != "" {
		return name, nil
	}
	if childErr != nil {
		return "", childErr
	}
	if readErr != nil {
		return "", ErrMalformed
	}
	return "", nil
}

// childBytes visits an already-decoded, immutable view of its parent's bytes.
// Stored ZIP members need no copy. Keep readChild's preflight and failed-read
// charges, including its max+1 probe, before the ordinary object admission.
func (s *scanState) childBytes(data []byte, depth int) (string, error) {
	max, err := s.childBudget(depth)
	if err != nil {
		return "", err
	}
	if err := s.ctx.Err(); err != nil {
		return "", err
	}
	if int64(len(data)) > max {
		s.spend(int(max + 1)) // safe: max < len(data)
		return "", ErrLimit
	}
	return s.visit(data, depth+1)
}

// bytecodeChildScanner keeps ClamBC output on the same generation, ancestry,
// deadline and aggregate child limits as every other decoded representation.
func (s *scanState) bytecodeChildScanner(depth int) func([]byte, int32, uint32) (string, error) {
	return func(data []byte, _ int32, containerType uint32) (string, error) {
		if containerType == 0 {
			return s.childBytes(data, depth)
		}
		max, err := s.childBudget(depth)
		if err != nil {
			return "", err
		}
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if int64(len(data)) > max {
			s.spend(int(max + 1))
			return "", ErrLimit
		}
		// The reference accepts numeric gaps below CL_TYPE_TEXT_ASCII but
		// cannot scan them as a known type. PART_ANY needs its distinct type
		// recognition mode; ERROR represents a failed read.
		if containerType < uint32(signatures.TypeTextASCII) || containerType > uint32(signatures.TypeIgnored) ||
			containerType == uint32(signatures.TypePartAny) || containerType == uint32(signatures.TypeError) {
			return "", ErrUnsupported
		}
		forced := signatures.FileType(containerType)
		if forced != signatures.TypeIgnored && len(data) >= minAnalyzedObject {
			observed, proven := s.db.DetermineType(data)
			// The forced type can make ClamAV run a parser even when the
			// bytes do not advertise that format. Until that parser path is
			// modeled, a mismatched type cannot be accepted as clean.
			if !proven || forced != observed && !(forced == signatures.TypeOther && textOrBinary(observed)) {
				return "", ErrUnsupported
			}
		}
		return s.visitObject(data, depth+1, objectOptions{forcedType: &forced})
	}
}

func (s *scanState) childBudget(depth int) (int64, error) {
	if depth >= s.maxDepth || s.files >= s.maxFiles {
		return 0, ErrLimit
	}
	max := min(s.maxFile, s.remaining)
	if max < 0 {
		return 0, ErrLimit
	}
	return max, nil
}

// readChild reads one decoded child object within the per-object limit and
// the remaining aggregate budget. The bytes of a failed or oversized
// expansion are charged, since decoding them was expansion work.
func (s *scanState) readChild(r io.Reader, depth int) ([]byte, error) {
	max, err := s.childBudget(depth)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(contextReader{s.ctx, r}, max+1))
	if err != nil {
		s.spend(len(data))
		if s.ctx.Err() != nil {
			return nil, s.ctx.Err()
		}
		return nil, ErrMalformed
	}
	if int64(len(data)) > max {
		s.spend(len(data))
		return nil, ErrLimit
	}
	return data, nil
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (r contextReader) Read(b []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.r.Read(b)
}

const zipBzip2 = 12

// zip analyzes an archive's members as the reference reads them
// (zipMembers): stored data, data stored with fewer bytes than its declared
// size (which the reference inflates), deflate data as far as it decodes
// (inflatedMember), bzip2 data, Deflate64 and implode data (zipMethod), and
// Zstandard data. Other compression methods remain explicitly incomplete.
func (s *scanState) zip(data []byte, depth int) (string, error) {
	members, err := zipMembers(data, s.maxFiles-s.files)
	if err != nil {
		return "", err
	}
	// Directory entries may share or overlap one local record. Exact duplicates
	// produce the same bytes and are decoded once; other entries together may
	// consume at most the archive's own size of compressed input, which bounds
	// decompression work that produces little or no charged output.
	type zipEntry struct {
		offset        int64
		compressed    uint64
		method, flags uint16
	}
	seen := make(map[zipEntry]struct{}, len(members))
	input := uint64(len(data))
	var incomplete error
	archive := s.layers.TypeAt(-1)
	for _, m := range members {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		// Container metadata rules see each member before it is extracted,
		// as its central record describes it and as its local header does.
		if name, err := s.db.MatchMember(s.ctx, archive, int64(len(data)), m.central); name != "" || err != nil {
			return name, err
		}
		if m.hasLocal && m.local != m.central {
			if name, err := s.db.MatchMember(s.ctx, archive, int64(len(data)), m.local); name != "" || err != nil {
				return name, err
			}
		}
		if m.issue != nil {
			incomplete = m.issue
		}
		if m.data < 0 || m.compressed == 0 {
			continue
		}
		entry := zipEntry{m.data, m.compressed, m.method, m.flags}
		if _, ok := seen[entry]; ok {
			continue
		}
		seen[entry] = struct{}{}
		if uint64(m.compressed) > input {
			incomplete = ErrLimit
			continue
		}
		input -= uint64(m.compressed)
		packed := data[m.data : m.data+int64(m.compressed)]
		if m.flags&zipFlagEncrypted != 0 {
			packed, err = s.zipDecrypt(packed, m.version, m.flags, m.mtime, m.crc)
			if err != nil {
				incomplete = err
				continue
			}
		}
		var sig string
		switch {
		case m.method == zip0Stored && uint64(len(packed)) >= m.size:
			sig, err = s.childBytes(packed, depth)
		case m.method == zip0Stored, m.method == zip8Deflate:
			sig, err = s.inflatedMember(packed, depth)
		case m.method == zipBzip2:
			sig, err = s.child(bzip2.NewReader(bytes.NewReader(packed)), depth)
		case m.method == zip6Implode, m.method == zip9Deflate64:
			sig, err = s.zipMethod(m.method, m.flags, packed, depth)
		case m.method == zip93Zstd, m.method == zip20ZstdLegacy:
			sig, err = s.zstd(packed, depth)
		case m.method == zip14LZMA:
			sig, err = s.zipLZMA(packed, m.flags, m.size, depth)
		default:
			// A reference-engine skip does not establish that an otherwise
			// valid member was analyzed. Report missing coverage explicitly.
			sig, err = "", ErrUnsupported
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
func (s *scanState) tar(data []byte, depth int) (string, error) {
	r := tar.NewReader(bytes.NewReader(data))
	var incomplete error
	count := 0
	for {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		h, err := r.Next()
		if err == io.EOF {
			return "", incomplete
		}
		if err != nil {
			return "", ErrMalformed
		}
		count++
		if count > s.maxFiles {
			return "", ErrLimit
		}
		switch h.Typeflag {
		case tar.TypeLink, tar.TypeSymlink, tar.TypeChar, tar.TypeBlock, tar.TypeDir, tar.TypeFifo, tar.TypeXGlobalHeader:
			// archive/tar gives these entries no member data. It consumes PAX
			// and GNU long-name records itself and returns a global PAX header
			// only after reading its records. When such an entry still declares
			// a size, archive/tar reads those bytes as the following headers
			// while other readers may skip them as the entry's data.
			if h.Size > 0 {
				incomplete = ErrMalformed
			}
			continue
		}
		// Every other typeflag carries member data that extractors write out,
		// including contiguous ('7'), old GNU sparse ('S') and unknown types.
		sig, err := s.child(r, depth)
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		} else if h.Typeflag == tarMultiVolume {
			incomplete = ErrUnsupported
		}
	}
}

// A GNU multi-volume continuation holds only the rest of a file whose start is
// in another volume, so its data is scanned but the file remains incomplete.
const tarMultiVolume = 'M'

// A mail header block is validated within these bounds before net/mail builds
// its header map, whose memory grows per field.
const (
	mailMaxHeaderBytes = 256 << 10
	mailMaxHeaderLines = 8192
	mailMaxFieldName   = 998 // RFC 5322 line length limit
)

// mailHeader describes an RFC 5322 header block at the start of data.
type mailHeader struct {
	valid     bool // the first line is a header field
	broken    bool // a later line before the blank separator is neither a field nor a continuation
	mime      bool // a MIME-Version, Content-Type or Content-Transfer-Encoding field
	mailFirst bool // the first field is a common message, trace or delivery field
	truncated bool // the block had not ended when the bounds were reached
}

// Fields that begin stored or delivered messages, including trace and
// delivery fields that mail transfer agents prepend.
var mailFirstFields = []string{
	"Received", "Return-Path", "Delivered-To", "Delivery-Date", "Envelope-To",
	"X-Original-To", "X-Envelope-From", "X-Apparently-To", "X-Received",
	"ARC-Seal", "ARC-Message-Signature", "ARC-Authentication-Results",
	"Authentication-Results", "DKIM-Signature", "Received-SPF",
	"From", "Sender", "To", "Cc", "Reply-To", "Subject", "Date", "Message-ID",
	"MIME-Version", "Content-Type",
}

// scanMailHeader walks the header block without copying it. A field name is
// RFC 5322 ftext (printable ASCII except colon), optionally followed by white
// space before its colon; a continuation line starts with white space.
func scanMailHeader(data []byte) mailHeader {
	var h mailHeader
	limit := min(len(data), mailMaxHeaderBytes)
	for pos, lines := 0, 0; pos < len(data); lines++ {
		if pos >= limit || lines >= mailMaxHeaderLines {
			h.truncated = true
			return h
		}
		if data[pos] == '\n' || data[pos] == '\r' && pos+1 < len(data) && data[pos+1] == '\n' {
			return h // the blank line ends the header block
		}
		if data[pos] == ' ' || data[pos] == '\t' {
			if lines == 0 {
				return h
			}
		} else {
			start := pos
			for pos < limit && pos-start < mailMaxFieldName && data[pos] > ' ' && data[pos] < 0x7f && data[pos] != ':' {
				pos++
			}
			name := data[start:pos]
			for pos < limit && (data[pos] == ' ' || data[pos] == '\t') {
				pos++
			}
			if pos >= limit && limit < len(data) {
				h.truncated = h.valid
				return h
			}
			switch {
			case len(name) != 0 && pos < limit && data[pos] == ':':
				for _, field := range []string{"MIME-Version", "Content-Type", "Content-Transfer-Encoding"} {
					h.mime = h.mime || bytes.EqualFold(name, []byte(field))
				}
				if lines == 0 {
					h.valid = true
					for _, field := range mailFirstFields {
						h.mailFirst = h.mailFirst || bytes.EqualFold(name, []byte(field))
					}
				}
			case lines == 0:
				return h
			default:
				h.broken = true
			}
		}
		end := bytes.IndexByte(data[pos:limit], '\n')
		if end < 0 {
			h.truncated = limit < len(data)
			return h // otherwise the header block ends with the data
		}
		pos += end + 1
	}
	return h
}

// looksLikeMail recognizes a header block that declares MIME content, from
// whichever field it starts with. A mail-like block that has not ended at its
// bounds is a resource limit; one with MIME fields after a line that is not a
// field cannot be parsed as the message a lenient mail reader may still see.
func looksLikeMail(data []byte) (bool, error) {
	h := scanMailHeader(data)
	switch {
	case !h.valid:
		return false, nil
	case h.truncated:
		if h.mime || h.mailFirst {
			return true, ErrLimit
		}
		return false, nil
	case h.broken:
		if h.mime {
			return true, ErrMalformed
		}
		return false, nil
	}
	return h.mime, nil
}
func (s *scanState) mail(data []byte, depth int) (string, error) {
	// An encapsulated message may have an empty header block.
	empty := bytes.HasPrefix(data, []byte("\n")) || bytes.HasPrefix(data, []byte("\r\n"))
	switch h := scanMailHeader(data); {
	case h.truncated:
		return "", ErrLimit
	case h.broken, !h.valid && !empty:
		return "", ErrMalformed
	}
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return "", ErrMalformed
	}
	contentType, encoding, err := mimeFields(msg.Header)
	if err != nil {
		return "", err
	}
	return s.mimeBody(contentType, encoding, msg.Body, depth)
}

// mimeFields returns the first Content-Type and Content-Transfer-Encoding
// values, also when a field is spelled with white space before its colon:
// RFC 5322 obsolete syntax that receivers must accept, which Go's readers keep
// in the key. Different values under different spellings are ambiguous.
func mimeFields(h map[string][]string) (contentType, encoding string, err error) {
	var typeFound, encodingFound bool
	for key, values := range h {
		if len(values) == 0 {
			continue
		}
		var value *string
		var found *bool
		switch name := strings.TrimRight(key, " \t"); {
		case strings.EqualFold(name, "Content-Type"):
			value, found = &contentType, &typeFound
		case strings.EqualFold(name, "Content-Transfer-Encoding"):
			value, found = &encoding, &encodingFound
		default:
			continue
		}
		if *found && *value != values[0] {
			return "", "", ErrMalformed
		}
		*value, *found = values[0], true
	}
	return contentType, encoding, nil
}

// visitMessage analyzes a part declared as an encapsulated message. It is
// typed, charged and raw-matched like any object, then parsed as a message
// whatever field it starts with, because its declared type already
// identifies it.
func (s *scanState) visitMessage(data []byte, depth int, prepaid int64) (string, error) {
	return s.visitObject(data, depth, objectOptions{prepaid: prepaid, message: true})
}
func (s *scanState) mimeBody(contentType, encoding string, body io.Reader, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	if contentType == "" {
		contentType = "text/plain"
	}
	mt, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", ErrMalformed
	}
	if strings.HasPrefix(mt, "multipart/") {
		if params["boundary"] == "" {
			return "", ErrMalformed
		}
		r := multipart.NewReader(body, params["boundary"])
		var incomplete error
		for {
			if s.files >= s.maxFiles {
				return "", ErrLimit
			}
			p, err := r.NextRawPart()
			if err == io.EOF {
				return "", incomplete
			}
			if err != nil {
				return "", ErrMalformed
			}
			s.files++
			var sig string
			partType, partEncoding, err := mimeFields(p.Header)
			if err == nil {
				sig, err = s.mimeBody(partType, partEncoding, p, depth+1)
			}
			if sig != "" {
				return sig, nil
			}
			if err != nil {
				incomplete = err
			}
			if err := p.Close(); err != nil {
				return "", ErrMalformed
			}
		}
	}
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		body = base64.NewDecoder(base64.StdEncoding, mimeBase64WhitespaceReader{s.ctx, body})
	case "quoted-printable":
		body = quotedprintable.NewReader(body)
	case "", "7bit", "8bit", "binary":
	default:
		return "", ErrUnsupported
	}
	switch mt {
	case "message/rfc822", "message/global":
		data, err := s.readChild(body, depth)
		if err != nil {
			return "", err
		}
		return s.visitMessage(data, depth+1, 0)
	case "message/partial":
		// RFC 2046 fragments are reassembled from several messages; one
		// fragment alone is not the complete encapsulated message.
		if sig, err := s.child(body, depth); sig != "" || err != nil {
			return sig, err
		}
		return "", ErrUnsupported
	case "text/html":
		// The part is HTML by its declared type, whatever markers it holds.
		data, err := s.readChild(body, depth)
		if err != nil {
			return "", err
		}
		return s.visitRepresentation(data, depth+1, htmlDeclared, 0)
	}
	return s.child(body, depth)
}

// MIME transfer bodies may fold base64 with spaces and horizontal tabs as
// well as CR/LF. encoding/base64 ignores only CR/LF. Filter MIME whitespace
// without making arbitrary corrupt bytes silently acceptable or buffering
// another copy of the encoded attachment.
type mimeBase64WhitespaceReader struct {
	ctx context.Context
	r   io.Reader
}

func (r mimeBase64WhitespaceReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	for {
		if err := r.ctx.Err(); err != nil {
			return 0, err
		}
		n, err := r.r.Read(p)
		out := 0
		for _, b := range p[:n] {
			if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
				continue
			}
			p[out] = b
			out++
		}
		if out != 0 || err != nil {
			return out, err
		}
	}
}

// htmlDocument recognizes HTML the way the reference engine types an object
// as HTML before parsing it, restricted to text. The object is text when its
// first 1,024 bytes hold no byte of the reference's never-text set (0x00-0x06,
// 0x0B, 0x0E-0x1A, 0x1C-0x1F and 0x7F; bytes from 0x80 are text), and it is
// HTML when one of the reference's HTML file-type markers occurs anywhere in
// it, as the reference searches the whole of a text object. The markers are
// case-sensitive as the reference writes them: "<a " or "<A " followed later
// by [hH][rR]EF or [hH][rR]ef; <HEAD>, <Head>, <head>, <HTML>, <Html> and
// <html>; <IFRAME, <iframe, <IMG, <Img, <img, <OBJECT, <Object, <object,
// <SCRIPT, <Script, <script, <TABLE and <table followed by '>', '/' or white
// space; and "<%@ LANGUAGE =" at the start. This is narrower than the
// reference, which also types binary data holding a marker in its first 1,024
// bytes as HTML. UTF-16 text is recognized in its decoded representation.
func htmlDocument(data []byte) bool {
	for _, c := range data[:min(len(data), 1024)] {
		if c <= 0x06 || c == 0x0b || c >= 0x0e && c <= 0x1a || c >= 0x1c && c <= 0x1f || c == 0x7f {
			return false
		}
	}
	if bytes.HasPrefix(data, []byte("<%@ LANGUAGE =")) {
		return true
	}
	anchor := -1
	for rest := data; ; {
		i := bytes.IndexByte(rest, '<')
		if i < 0 {
			break
		}
		rest = rest[i+1:]
		if anchor < 0 && len(rest) >= 2 && (rest[0] == 'a' || rest[0] == 'A') && rest[1] == ' ' {
			anchor = len(data) - len(rest) + 2
		}
		for _, marker := range htmlClosedMarkers {
			if bytes.HasPrefix(rest, marker) {
				return true
			}
		}
		for _, marker := range htmlOpenMarkers {
			if len(rest) > len(marker) && bytes.HasPrefix(rest, marker) && strings.IndexByte("> \t\n\v\f\r/", rest[len(marker)]) >= 0 {
				return true
			}
		}
	}
	if anchor < 0 {
		return false
	}
	for rest := data[anchor:]; len(rest) >= 3; {
		i := bytes.IndexAny(rest[1:], "rR")
		if i < 0 || i+4 > len(rest) {
			return false
		}
		if (rest[i] == 'h' || rest[i] == 'H') && (string(rest[i+2:i+4]) == "EF" || string(rest[i+2:i+4]) == "ef") {
			return true
		}
		rest = rest[i+1:]
	}
	return false
}

var (
	htmlClosedMarkers = [][]byte{[]byte("HEAD>"), []byte("Head>"), []byte("head>"), []byte("HTML>"), []byte("Html>"), []byte("html>")}
	htmlOpenMarkers   = [][]byte{[]byte("IFRAME"), []byte("iframe"), []byte("IMG"), []byte("Img"), []byte("img"), []byte("OBJECT"),
		[]byte("Object"), []byte("object"), []byte("SCRIPT"), []byte("Script"), []byte("script"), []byte("TABLE"), []byte("table")}
)
