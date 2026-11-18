package engine

import (
	"bytes"
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// The per-object pipeline follows the reference engine's object scan: the
// object is typed from its content, an ignored type is not analyzed at all,
// the raw bytes are matched against the object's typed rule root and root 0
// (never for HTML, whose normalized forms are matched instead), the raw scan
// may retype text it recognizes as HTML or mail, and the object's
// representations are then produced: members of containers, the parts of a
// message, and normalized text. Every object is a layer of the scan's
// recursion stack, which Container and Intermediates rules read.

// Size limits of the reference's representations.
const (
	maxScriptNormalize = 20 << 20
	maxHTMLNormalize   = 40 << 20
)

// objectOptions describe how an object reached the scan.
type objectOptions struct {
	mode       htmlMode
	prepaid    int64
	forcedType *signatures.FileType
	// message marks a part declared as an encapsulated message: it is
	// parsed as mail whatever its type.
	message bool
	// htmlData identifies an internal data-URI child. Its compatibility
	// bytes still receive the complete mail-object rule/hash pipeline.
	htmlData *htmlDataPart
	// parser is what the reference's Windows-program parser did in the
	// object's raw scan.
	parser parserFacts
}

// parserFacts is what the raw scan's Windows-program parser step produced:
// the program the UPX step rebuilt, which is scanned as a child, and
// whether that step's outcome is known. headerRejected means the reference
// parser stopped before forming an executable representation.
type parserFacts struct {
	unpacked       []byte
	upxDecided     bool
	headerRejected bool
}

// unknownRepresentation is an incomplete result because the reference
// matches a representation of the object that this engine does not produce.
func unknownRepresentation(reason string) error {
	return fmt.Errorf("%w: %w", ErrUnsupported, &signatures.UnknownError{Reason: reason})
}

// unsupportedType is an incomplete result for an object or embedded object of
// a type this engine cannot parse.
func unsupportedType(t signatures.FileType) error {
	return unknownRepresentation(formatReason(t))
}

func formatReason(t signatures.FileType) string { return "format-" + strings.ToLower(t.String()) }

// unparsedFormats are formats found inside objects for which this engine does
// not yet produce every reference representation (see embeddedObjects);
// referenceOnlyFormats applies the same conservative gate to object types.
var unparsedFormats = [...]signatures.FileType{
	signatures.TypeXMLWord, signatures.TypeXMLXL, signatures.TypeXMLHWP,
	signatures.TypeDMG, signatures.TypeISO9660, signatures.TypeUDF,
	signatures.TypeEGGSFX, signatures.TypeAutoIt,
}

// RepresentationsNotProduced lists the representations the reference
// matches that this engine does not produce yet. An object that needs one is
// reported incomplete, never clean.
func RepresentationsNotProduced() []string {
	out := []string{signatures.UnknownHTML, signatures.UnknownVBA, signatures.UnknownWordMacros, signatures.UnknownWorkbookPictures, signatures.UnknownCompoundFile}
	types := map[signatures.FileType]bool{}
	for _, t := range unparsedFormats {
		types[t] = true
	}
	for t := range referenceOnlyFormats {
		types[t] = true
	}
	sorted := make([]signatures.FileType, 0, len(types))
	for t := range types {
		sorted = append(sorted, t)
	}
	slices.Sort(sorted)
	for _, t := range sorted {
		out = append(out, formatReason(t))
	}
	return out
}

func isText(t signatures.FileType) bool {
	switch t {
	case signatures.TypeTextASCII, signatures.TypeTextUTF8, signatures.TypeTextUTF16LE, signatures.TypeTextUTF16BE, signatures.TypeScript:
		return true
	}
	return false
}

func textOrBinary(t signatures.FileType) bool {
	return t >= signatures.TypeTextASCII && t <= signatures.TypeBinaryData
}

// typeRecognition reports whether the raw scan of an object of this type
// also runs the type patterns, as the reference does for all but archives
// and tar files.
func typeRecognition(t signatures.FileType) bool {
	switch t {
	case signatures.TypeGZ, signatures.TypeCPIOOld, signatures.TypeZIP, signatures.TypeOldTar, signatures.TypePosixTar,
		signatures.TypeZIPSFX, signatures.TypeRARSFX, signatures.Type7ZSFX, signatures.TypeCABSFX, signatures.TypeARJSFX, signatures.TypeEGGSFX:
		return false
	}
	return true
}

func (s *scanState) visitObject(data []byte, depth int, o objectOptions) (string, error) {
	if done, err := s.admitObject(data, depth, o.prepaid); done {
		return "", err
	}
	typ, typed := s.db.DetermineType(data)
	if isoHeader(data) && (typ == signatures.TypeBinaryData || typ == signatures.TypeAny) {
		typ, typed = signatures.TypeISO9660, true
	}
	if lhaHeader(data) && (typ == signatures.TypeBinaryData || typ == signatures.TypeAny) {
		typ, typed = signatures.TypeLHALZH, true
	}
	if alzHeader(data) && (typ == signatures.TypeBinaryData || typ == signatures.TypeAny) {
		typ, typed = signatures.TypeALZ, true
	}
	if eggHeader(data) && (typ == signatures.TypeBinaryData || typ == signatures.TypeAny) {
		typ, typed = signatures.TypeEGG, true
	}
	if o.forcedType != nil {
		typ, typed = *o.forcedType, true
	}
	if o.htmlData != nil {
		typ, typed = signatures.TypeMail, true
	}
	if typ == signatures.TypeIgnored {
		return "", nil
	}
	if depth == 0 {
		s.rootEligible, s.rootType = true, typ
	}
	var metadataErr error
	if depth == 0 && s.collectMetadata {
		s.metadataRoot, metadataErr = initialRootMetadata(s.ctx, data, typ, s.metadataPath)
		if metadataErr != nil {
			// Invalid paths must not suppress an independent content detection.
			// Preserve the proved root fields and keep absent queries unknown.
			s.metadataRoot, _ = initialRootMetadata(s.ctx, data, typ, "")
		}
	}
	s.layers.Push(signatures.Layer{Type: typ, Size: int64(len(data))})
	defer s.layers.Pop()
	undecided := metadataErr
	if !typed {
		undecided = unknownRepresentation("text-encoding")
	}
	// Authenticode trust must be decided before the raw rule pass, which it
	// can exempt only after authenticating the signer, chain and PE digest.
	// Check every embedded certificate for a CRB blocklist hit first.
	if bytes.HasPrefix(data, []byte("MZ")) {
		name, trusted, err := s.authenticode(data)
		if name != "" {
			return name, nil
		}
		if trusted {
			if depth == 0 {
				s.rootEligible = false
			}
			return "", nil
		}
		if err != nil && undecided == nil {
			undecided = err
		}
	}
	if s.db.HasCatalogHashes() {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		switch {
		case bytes.HasPrefix(data, []byte("MZ")):
			for _, algorithm := range []crypto.Hash{crypto.SHA1, crypto.SHA256} {
				digest, err := peCatalogDigest(data, algorithm)
				if err == nil && s.db.CatalogAllows(2, digest) {
					if depth == 0 {
						s.rootEligible = false
					}
					return "", nil
				}
				if err != nil && undecided == nil {
					undecided = err
				}
			}
		case bytes.HasPrefix(data, []byte("MSCF")):
			sha1Digest, sha256Digest := sha1.Sum(data), sha256.Sum256(data)
			if s.db.CatalogAllows(1, sha1Digest[:]) || s.db.CatalogAllows(1, sha256Digest[:]) {
				if depth == 0 {
					s.rootEligible = false
				}
				return "", nil
			}
		}
	}
	if allowed, err := s.db.Allowed(s.ctx, data); err != nil || allowed {
		if allowed && depth == 0 {
			s.rootEligible = false
		}
		return "", err
	}
	// ClamAV runs cli_scanscript before the raw rule pass for a SCRIPT
	// object, including a HandlerType-forced SCRIPT duplicate.
	if typ == signatures.TypeScript && len(data) <= maxScriptNormalize {
		if sig, err := s.normalizedText(data, depth); sig != "" {
			return sig, nil
		} else if err != nil && undecided == nil {
			undecided = err
		}
	}
	// ZIP extraction precedes raw matching in cli_magic_scan. This also
	// matters for a HandlerType-forced ZIP duplicate: a member alert can
	// take precedence over the generic rule that requested the retype.
	if typ == signatures.TypeZIP {
		if sig, err := s.zip(data, depth); sig != "" {
			return sig, nil
		} else if err != nil && undecided == nil {
			undecided = err
		}
	}
	res, err := s.rawScan(data, typ, depth)
	switch {
	case res.Name != "" && !res.Deferred:
		return res.Name, nil
	case err != nil && errors.Is(err, signatures.ErrRuleIncomplete) && s.ctx.Err() == nil:
		// An undecided rule does not stop the object's analysis; it stands
		// only if nothing inside the object is detected.
		if undecided == nil {
			undecided = err
		}
	case err != nil:
		return "", err
	}
	original := typ
	switch {
	case res.DetectedType == signatures.TypeHTML && (typ == signatures.TypeTextASCII || typ == signatures.TypeGIF) &&
		s.layers.TypeAt(-2) != signatures.TypeAutoIt:
		typ = signatures.TypeHTML
	case res.DetectedType == signatures.TypeMail && typ == signatures.TypeTextASCII:
		typ = signatures.TypeMail
	}
	if typ != original {
		s.layers.Retype(typ)
		if depth == 0 {
			s.rootType = typ
			if s.metadataRoot != nil {
				s.metadataRoot["FileType"] = "CL_TYPE_" + typ.String()
			}
		}
	}
	o.parser = parserFacts{unpacked: res.Unpacked, upxDecided: res.UPXDecided, headerRejected: res.PEHeaderRejected}
	sig, err := s.analyzeObject(data, depth, original, typ, res.Embedded, o)
	if sig != "" {
		return sig, nil
	}
	if res.Deferred && res.Name != "" && s.ctx.Err() == nil && err == nil {
		return res.Name, nil
	}
	if undecided != nil {
		return "", undecided
	}
	return "", err
}

// admitObject applies the object limits and charges the object to the scan
// budget. done means nothing more is analyzed: the object is too small to
// analyze, over a limit, or the scan was cancelled.
func (s *scanState) admitObject(data []byte, depth int, prepaid int64) (bool, error) {
	if err := s.ctx.Err(); err != nil {
		return true, err
	}
	if len(data) < minAnalyzedObject {
		return true, nil
	}
	charge := max(0, int64(len(data))-prepaid)
	if depth > s.maxDepth || s.files >= s.maxFiles || int64(len(data)) > s.maxFile || charge > s.remaining {
		return true, ErrLimit
	}
	s.files++
	s.remaining -= charge
	return false, nil
}

// rawScan matches the object's own bytes. As in the reference, an object
// typed HTML is not matched raw, not even by its file hashes (checked against
// the reference): its normalized views are (htmlViews). UTF-16 HTML gets the
// full raw scan, as there.
func (s *scanState) rawScan(data []byte, typ signatures.FileType, depth int) (signatures.MatchResult, error) {
	if typ == signatures.TypeHTML {
		return signatures.MatchResult{}, nil
	}
	var fileReliability *uint32
	if depth == 0 {
		// A top-level input has not been reconstructed by an unpacker or
		// carved from another executable.
		normal := uint32(0)
		fileReliability = &normal
	}
	scan := typ
	if scan == signatures.TypeTextASCII {
		scan = signatures.TypeAny // plain text meets root 0 only; root 7 is for normalized text
	}
	return s.db.MatchObject(s.ctx, data, signatures.MatchContext{
		Type: scan, Layers: &s.layers, TypeRecognition: typeRecognition(typ), PEAlertBroken: !s.pePermissive,
		PDFParserHooks: typ == signatures.TypePDF,
		Metadata:       s.objectMetadata(), MetadataPartial: s.metadataRoot != nil,
		ScanExtractedTyped: s.bytecodeChildScanner(depth), BytecodeOptions: s.bytecodeEngineOptions(),
		FileReliability: fileReliability,
		PreclassMatched: s.preclassMatched,
		ScanHandler: func(data []byte, forced signatures.FileType) (string, error) {
			return s.visitObject(data, depth+1, objectOptions{forcedType: &forced})
		}, HandlerHash: &s.handlerHash,
	})
}

// analyzeObject produces the object's representations after its raw scan.
// original is the type before the raw scan retyped it to typ.
func (s *scanState) analyzeObject(data []byte, depth int, original, typ signatures.FileType, embedded []signatures.EmbeddedType, o objectOptions) (string, error) {
	if o.htmlData != nil {
		p := o.htmlData
		return s.mimeBody(p.contentType, p.encoding, bytes.NewReader(p.body), depth)
	}
	var incomplete error
	note := func(err error) {
		if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	if sig, err := s.embeddedObjects(data, depth, typ, embedded); sig != "" {
		return sig, nil
	} else {
		note(err)
	}
	// ClamAV 1.5.4 makes a second, virus-only pass through target 14 for
	// these three types after the ordinary raw scan. The current layer keeps
	// its original file type; OTHER selects the matcher root for this pass.
	if typ == signatures.TypeBinaryData || typ == signatures.TypePythonComp || typ == signatures.TypeAIModel {
		if res, err := s.rawScan(data, signatures.TypeOther, depth); res.Name != "" {
			return res.Name, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeMachOUnibin {
		if sig, err := s.machoSlices(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	mail := typ == signatures.TypeMail || o.message
	// A mailbox's From postmark is outside its RFC 5322 message headers.
	// Let analyze dispatch the mailbox once; parsing the entire mailbox as
	// one message would retain a spurious malformed result after its actual
	// messages were successfully analyzed. Explicit message/rfc822 parts
	// still receive their declared-message validation.
	if mail && (o.message || !mboxPostmark(data[:min(len(data), 1024)])) {
		if sig, err := s.mail(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeMHTML {
		if sig, err := s.mhtml(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
		mail = true
	}
	if typ == signatures.TypeScrEnc {
		if sig, err := s.screnc(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeXDP {
		if sig, err := s.xdp(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeXMLWord || typ == signatures.TypeXMLXL {
		if sig, err := s.msxml(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeXMLHWP {
		if sig, err := s.hwpML(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeHWPOLE2 {
		// The four-byte prefix describes the uncompressed OLE size. The
		// reference warns when it differs from the available bytes but still
		// scans the nested document after that prefix.
		if len(data) < 4 {
			note(ErrMalformed)
		} else if sig, err := s.visit(data[4:], depth+1); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeOneNote {
		if !oneNoteWebEnvelope(data) {
			if sig, err := s.oneNoteAttachments(data, depth); sig != "" {
				return sig, nil
			} else {
				note(err)
			}
		}
		if sig, err := s.oneNoteBlobs(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if referenceOnlyFormats[typ] {
		note(unsupportedType(typ))
	}
	if typ == signatures.TypeISO9660 && !isoHeader(data) {
		note(unsupportedType(typ))
	}
	if typ == signatures.TypeLHALZH && !lhaHeader(data) {
		note(unsupportedType(typ))
	}
	if typ == signatures.TypeMBR {
		if sig, err := s.mbr(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeGPT {
		if sig, err := s.gpt(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeAPM {
		if sig, err := s.apm(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ == signatures.TypeCryptFF {
		if sig, err := s.cryptFF(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if typ != signatures.TypeZIP && typ != signatures.TypeMBR && typ != signatures.TypeGPT && typ != signatures.TypeAPM && typ != signatures.TypeCryptFF {
		if sig, err := s.analyze(data, depth, o.mode, mail, o.parser); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if (typ == signatures.TypeHTML || typ == signatures.TypeHTMLUTF16) && len(data) <= maxHTMLNormalize {
		if sig, err := s.htmlViews(data, depth, typ); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if isText(original) && typ != signatures.TypeHTML && typ != signatures.TypeScript && len(data) <= maxScriptNormalize {
		if sig, err := s.normalizedText(data, depth); sig != "" {
			return sig, nil
		} else {
			note(err)
		}
	}
	if isText(original) && typ == signatures.TypeMail {
		// Text recognized as a message is also matched raw as mail.
		res, err := s.db.MatchObject(s.ctx, data, signatures.MatchContext{Type: signatures.TypeMail, Layers: &s.layers, Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata, ScanExtractedTyped: s.bytecodeChildScanner(depth), BytecodeOptions: s.bytecodeEngineOptions(), PreclassMatched: s.preclassMatched})
		if res.Name != "" {
			return res.Name, nil
		}
		note(err)
	}
	return "", incomplete
}

// normalizedText matches the object's normalized text, a layer of its own,
// against root 7 and root 0: as a file (with hashes and EOF offsets) or as a
// buffer, as the database's text root requires.
func (s *scanState) normalizedText(data []byte, depth int) (string, error) {
	if depth+1 > s.maxDepth {
		return "", ErrLimit
	}
	chunk, fileMode := s.db.TextNormalization()
	mode := signatures.FileMode
	if !fileMode {
		mode, chunk = signatures.BufferMode, 0
	}
	normalized := normalizeText(data, chunk)
	if len(normalized) == 0 {
		return "", nil
	}
	s.layers.Push(signatures.Layer{Type: signatures.TypeTextASCII, Size: int64(len(normalized)), Normalized: true})
	defer s.layers.Pop()
	res, err := s.db.MatchObject(s.ctx, normalized, signatures.MatchContext{Type: signatures.TypeTextASCII, Mode: mode, Layers: &s.layers, Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata, ScanExtractedTyped: s.bytecodeChildScanner(depth + 1), BytecodeOptions: s.bytecodeEngineOptions(), PreclassMatched: s.preclassMatched})
	return res.Name, err
}

// embeddedObjects handles what the raw scan's type recognition found inside
// the object, as the reference does: a message found in text or binary is
// parsed as mail, archives found after the start are extracted on a layer of
// their archive type, and the first valid executable inside an executable is
// analyzed. In an executable, the overlay is left to the program analysis
// (installer.go), which searches it for archives and installers itself.
// Formats this engine cannot parse make the object incomplete.
func (s *scanState) embeddedObjects(data []byte, depth int, typ signatures.FileType, records []signatures.EmbeddedType) (string, error) {
	var incomplete error
	note := func(sig string, err error) (string, bool) {
		if sig != "" {
			return sig, true
		}
		if err != nil {
			incomplete = err
		}
		return "", false
	}
	exe := typ == signatures.TypeMSEXE
	image := len(data)
	if exe {
		if img, err := parsePEImage(data); err == nil {
			image = img.overlay
		}
	}
	message := false
	for _, r := range records {
		if r.Offset <= 0 || r.Offset >= len(data) {
			continue
		}
		switch r.Type {
		case signatures.TypeMHTML:
			if textOrBinary(typ) && !message {
				message = true
				s.layers.Retype(signatures.TypeMHTML)
				if sig, done := note(s.mhtml(data, depth)); done {
					return sig, nil
				}
			}
		case signatures.TypeXDP:
			if textOrBinary(typ) || typ == signatures.TypePDF {
				if sig, done := note(s.xdp(data, depth)); done {
					return sig, nil
				}
			}
		case signatures.TypeXMLWord, signatures.TypeXMLXL:
			if textOrBinary(typ) {
				if sig, done := note(s.msxml(data, depth)); done {
					return sig, nil
				}
				incomplete = unsupportedType(r.Type)
			}
		case signatures.TypeXMLHWP:
			if textOrBinary(typ) {
				if sig, done := note(s.hwpML(data, depth)); done {
					return sig, nil
				}
				incomplete = unsupportedType(r.Type)
			}
		case signatures.TypeISO9660:
			// The volume-descriptor magic is an interior match even for a
			// top-level image that the ISO reader already traverses.
			layout, recognized := isoLayoutOf(data)
			pvd := 16*layout.sectorSize + layout.dataOffset
			if typ != signatures.TypeISO9660 || !recognized ||
				r.Offset < pvd || r.Offset > pvd+6 {
				incomplete = unsupportedType(r.Type)
			}
		case signatures.TypeMBR:
			if typ != signatures.TypeMBR && typ != signatures.TypeGPT && r.Offset == 510 &&
				len(data) >= 2*mbrSectorSize {
				if sig, done := note(s.mbr(data, depth)); done {
					return sig, nil
				}
			} else if typ != signatures.TypeMBR && typ != signatures.TypeGPT && len(data) >= 2*mbrSectorSize {
				incomplete = unsupportedType(r.Type)
			}
		case signatures.TypeGPT:
			if typ != signatures.TypeGPT && typ != signatures.TypeMBR {
				if sig, done := note(s.gpt(data, depth)); done {
					return sig, nil
				}
			}
		case signatures.TypeDMG, signatures.TypeUDF:
			incomplete = unsupportedType(r.Type)
		case signatures.TypeEGGSFX:
			if typ == signatures.TypeEGG || r.Offset >= image {
				continue
			}
			if sig, done := note(s.embeddedArchive(data[r.Offset:], signatures.TypeEGG, depth)); done {
				return sig, nil
			}
		case signatures.TypeAutoIt:
			if exe {
				incomplete = unsupportedType(r.Type)
			}
		case signatures.TypeZIPSFX:
			if typ == signatures.TypeZIP || r.Offset >= image {
				continue
			}
			if sig, done := note(s.zipSingle(data, r.Offset, depth)); done {
				return sig, nil
			}
		case signatures.TypeRARSFX, signatures.TypeCABSFX, signatures.TypeARJSFX, signatures.Type7ZSFX:
			if typ == sfxArchive(r.Type) || r.Offset >= image {
				continue
			}
			if sig, done := note(s.embeddedArchive(data[r.Offset:], sfxArchive(r.Type), depth)); done {
				return sig, nil
			}
		case signatures.TypeMSEXE:
			if exe || len(data)-r.Offset > maxEmbeddedPE {
				continue
			}
			if _, err := parsePEImage(data[r.Offset:]); err != nil {
				continue
			}
			exe = true // the reference analyzes the first valid one only
			if sig, done := note(s.visit(data[r.Offset:], depth+1)); done {
				return sig, nil
			}
		}
	}
	return "", incomplete
}

// embeddedArchive analyzes an archive found inside another object on a layer
// of the archive's type, as the reference pushes it, with the archive's
// parser. The reference ignores an archive header that does not parse there,
// so that describes nothing to analyze.
func (s *scanState) embeddedArchive(tail []byte, archive signatures.FileType, depth int) (string, error) {
	if depth+1 > s.maxDepth {
		return "", ErrLimit
	}
	if archive == signatures.TypeMSCAB {
		tail = cabEmbeddedView(tail)
	}
	s.layers.Push(signatures.Layer{Type: archive, Size: int64(len(tail))})
	defer s.layers.Pop()
	var sig string
	var err error
	switch archive {
	case signatures.TypeRAR:
		sig, err = s.rar(tail, depth+1)
	case signatures.TypeMSCAB:
		sig, err = s.cab(tail, depth+1)
	case signatures.TypeARJ:
		sig, err = s.arj(tail, depth+1)
	case signatures.Type7Z:
		sig, err = s.sevenZip(tail, depth+1)
	case signatures.TypeEGG:
		sig, err = s.egg(tail, depth+1)
	}
	if errors.Is(err, ErrMalformed) {
		err = nil
	}
	return sig, err
}

// mhtml parses an object holding a MIME message (MHTML) as mail, as the
// reference does. Its mail parser reads whatever it can of any object; where
// this engine's stricter parser finds no message, the reference finds no
// parts either, so that is not reported.
func (s *scanState) mhtml(data []byte, depth int) (string, error) {
	sig, err := s.mail(data, depth)
	if errors.Is(err, ErrMalformed) {
		err = nil
	}
	return sig, err
}

// referenceOnlyFormats are types for which the reference produces at least
// one representation this engine does not yet produce. An undetected object
// remains incomplete after any supported partial scan.
var referenceOnlyFormats = map[signatures.FileType]bool{
	signatures.TypeOneNote: true,
	signatures.TypeHWP3:    true,
	signatures.TypeXMLWord: true, signatures.TypeXMLXL: true, signatures.TypeXMLHWP: true,
	signatures.TypeAutoIt: true, signatures.TypeDMG: true,
	signatures.TypeUDF:         true,
	signatures.TypePartHFSPlus: true, signatures.TypeNullsoft: true, signatures.TypeIShieldMSI: true,
}

// sfxArchive is the archive type a self-extractor type is found as.
func sfxArchive(t signatures.FileType) signatures.FileType {
	switch t {
	case signatures.TypeZIPSFX:
		return signatures.TypeZIP
	case signatures.TypeRARSFX:
		return signatures.TypeRAR
	case signatures.TypeCABSFX:
		return signatures.TypeMSCAB
	case signatures.TypeARJSFX:
		return signatures.TypeARJ
	case signatures.Type7ZSFX:
		return signatures.Type7Z
	}
	return 0
}

// machoSlices analyzes each architecture of a universal Mach-O binary as an
// object of its own, a window of the universal file, as the reference does.
func (s *scanState) machoSlices(data []byte, depth int) (string, error) {
	if len(data) < 8 {
		return "", ErrMalformed
	}
	count := binary.BigEndian.Uint32(data[4:])
	if count > 32 || uint64(len(data)) < 8+20*uint64(count) {
		return "", ErrMalformed
	}
	var incomplete error
	for i := uint32(0); i < count; i++ {
		arch := data[8+20*i:]
		offset, size := uint64(binary.BigEndian.Uint32(arch[8:])), uint64(binary.BigEndian.Uint32(arch[12:]))
		if offset > uint64(len(data)) || size > uint64(len(data))-offset {
			incomplete = ErrMalformed
			continue
		}
		sig, err := s.visitObject(data[offset:offset+size], depth+1, objectOptions{})
		if sig != "" {
			return sig, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}
