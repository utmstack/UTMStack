package engine

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/klauspost/compress/lzw"
	"github.com/utmstack/UTMStack/agent/edr/bytecode"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// The lexical/filter rules follow Adobe PDF Reference 1.7, sections 3.1–3.4:
// https://opensource.adobe.com/dc-acrobat-sdk-docs/pdfstandards/pdfreference1.7old.pdf
// This is bounded content extraction, not a PDF renderer, JavaScript emulator,
// complete cross-reference resolver, or image/font decoder.
type pdfRef struct{ object, generation int64 }
type pdfValue struct {
	kind   byte
	number int64
	// str holds a string's bytes up to pdfMaxKeptString bytes (encryption
	// parameters and file identifiers are short).
	str   []byte
	name  string
	ref   pdfRef
	array []pdfValue
	dict  map[string]pdfValue
}
type pdfStream struct {
	object                 pdfRef
	dict                   map[string]pdfValue
	data                   []byte
	start, end, terminator int
}
type pdfScan struct {
	s               *scanState
	refs            map[pdfRef]pdfValue
	streams         []pdfStream
	hookObjects     []pdfHookObject
	hookState       *bytecode.PDFState
	hookUncertain   bool
	tokens, objects int
	// size and budgetStart bound the tokens read (step): every token takes
	// at least one byte of the file or of data decoded from it, which is
	// charged to the scan budget.
	size, budgetStart int64
	signature         string
	incomplete        error
	// ClamAV flags hex-escaped spellings of a small set of common names in
	// object dictionaries. Keep this separate from signature matching so an
	// actual malware signature takes precedence over the heuristic.
	obfuscatedName bool
	// dataStreams are objects named as attachments, scripts, metadata or
	// forms: their streams hold files or other languages, not content syntax.
	dataStreams map[pdfRef]bool
	// strings holds the decoded strings of the object being parsed, which are
	// analyzed together at depth stringDepth; stringBytes is their length.
	strings                  [][]byte
	stringBytes, stringDepth int
	// forward holds the file's integer objects, read on the first stream
	// whose length refers to an object not read yet (pdfIntegerObjects).
	forward map[pdfRef]int64
	// encrypt and fileID are the last trailer's (or cross-reference
	// stream's) /Encrypt value and first /ID string; encryptDicts keep the
	// objects that look like encryption dictionaries; crypt decrypts the
	// streams, and undecryptable marks encryption this engine cannot undo.
	encrypt       *pdfValue
	fileID        []byte
	encryptDicts  map[pdfRef]map[string]pdfValue
	crypt         *pdfCrypt
	undecryptable bool
}

type pdfHookObject struct {
	ref    pdfRef
	offset int // first byte after the indirect object's "obj" keyword
	end    int // first byte of the parsed "endobj" keyword
	kind   byte
}

// pdfMaxKeptString bounds the string bytes a value keeps.
const pdfMaxKeptString = 1024

type pdfParser struct {
	owner      *pdfScan
	data       []byte
	pos, depth int
	dictDepth  int
}

func (s *scanState) pdf(data []byte, depth int) (string, error) {
	if !bytes.HasPrefix(data, []byte("%PDF-")) || len(data) < 8 {
		return "", ErrMalformed
	}
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	ps := &pdfScan{s: s, refs: make(map[pdfRef]pdfValue), size: int64(len(data)), budgetStart: s.remaining}
	var hookPending error
	if sig, err := s.db.MatchPDFPre(s.ctx, data, signatures.MatchContext{
		BytecodeOptions: s.bytecodeEngineOptions(), Metadata: s.representationMetadata(),
		MetadataPartial: s.collectMetadata, ScanExtractedTyped: s.bytecodeChildScanner(depth),
	}); sig != "" {
		return sig, nil
	} else if err != nil {
		hookPending = err
	}
	p := &pdfParser{owner: ps, data: data, depth: depth}
	// Offsets of the "obj" keywords of objects read in sequence.
	headers := make(map[int]bool)
	for {
		p.space()
		if p.pos >= len(data) {
			break
		}
		ref, object, err := p.objectHeader()
		if err != nil {
			ps.incomplete = err
			break
		}
		if object {
			headers[p.pos-3] = true
			if err := ps.object(p, ref, true); ps.signature != "" {
				return ps.signature, nil
			} else if err != nil {
				ps.incomplete = err
				break
			}
			continue
		}
		_, err = p.value(0)
		if ps.flushStrings(); ps.signature != "" {
			return ps.signature, nil
		} else if err != nil {
			ps.incomplete = err
			break
		}
	}
	objectsBeforeEmbedded := ps.objects
	if err := ps.embeddedObjects(data, headers, depth); ps.signature != "" {
		return ps.signature, nil
	} else if err != nil {
		ps.incomplete = err
	}
	ps.hookUncertain = ps.objects != objectsBeforeEmbedded
	ps.setUpDecryption()
	ps.hookState = ps.bytecodeState(data)
	// A parse error can stop the reference before it reaches PARSED. A
	// phase-only hook must not turn that uncertain path into a detection.
	if ps.incomplete == nil {
		if sig, err := s.db.MatchPDFPhase(s.ctx, data, 1, signatures.MatchContext{
			PDF: ps.hookState, BytecodeOptions: ps.s.bytecodeEngineOptions(), Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata,
			ScanExtractedTyped: s.bytecodeChildScanner(depth),
		}); sig != "" {
			return sig, nil
		} else if err != nil {
			ps.incomplete = err
		}
	}
	// PARSED observes the outer object before object-stream extraction. Once
	// extraction starts, its member objects may change the reference's flag
	// state. The current snapshot is only proved for PARSED in that case.
	for _, stream := range ps.streams {
		if ps.objectStreamDict(stream.dict) {
			for i := range ps.hookState.Objects {
				ps.hookState.Objects[i].FlagsKnown = false
			}
			break
		}
	}
	// Object streams are analyzed first: the objects they hold may give other
	// streams' lengths or name them as attachments or scripts. Which pass a
	// stream belongs to is decided once, so that each is analyzed once.
	objectStreams := make([]bool, len(ps.streams))
	for i, stream := range ps.streams {
		objectStreams[i] = ps.objectStreamDict(stream.dict)
	}
	for _, objects := range []bool{true, false} {
		for i, stream := range ps.streams {
			if err := s.ctx.Err(); err != nil {
				return "", err
			}
			if objectStreams[i] != objects {
				continue
			}
			if sig, err := ps.scanStream(stream, depth); sig != "" {
				return sig, nil
			} else if err != nil {
				ps.incomplete = err
			}
		}
	}
	if flags, known := ps.provenEndFlags(data); known {
		ps.hookState.Flags, ps.hookState.FlagsKnown = flags, true
		if sig, err := s.db.MatchPDFPhase(s.ctx, data, 3, signatures.MatchContext{
			PDF: ps.hookState, BytecodeOptions: ps.s.bytecodeEngineOptions(), Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata,
			ScanExtractedTyped: s.bytecodeChildScanner(depth),
		}); sig != "" {
			return sig, nil
		} else if err != nil {
			ps.incomplete = err
		}
	} else if ps.incomplete == nil {
		// An END hook may be skipped when final flags are zero. Probe with
		// unknown state so a phase-independent zero return can be proved,
		// but a label or unsupported query remains undecided.
		if sig, err := s.db.MatchPDFPhase(s.ctx, data, 3, signatures.MatchContext{
			PDF: &bytecode.PDFState{}, BytecodeOptions: s.bytecodeEngineOptions(),
			Metadata: s.representationMetadata(), MetadataPartial: s.collectMetadata,
			ScanExtractedTyped: s.bytecodeChildScanner(depth),
		}); sig != "" && hookPending == nil {
			hookPending = fmt.Errorf("%w: unproved PDF END hook reachability", signatures.ErrRuleIncomplete)
		} else if err != nil && hookPending == nil {
			hookPending = err
		}
	}
	if ps.objects == 0 || !bytes.Contains(data, []byte("%%EOF")) {
		ps.incomplete = ErrMalformed
	}
	// The reference still emits this heuristic after a failed stream decoder
	// or missing EOF when it has parsed an escaped common dictionary name.
	// As with a signature match, that detection takes precedence over an
	// otherwise incomplete result.
	if ps.obfuscatedName && !s.pdfNameHeuristicOff {
		return "Heuristics.PDF.ObfuscatedNameObject", nil
	}
	if ps.incomplete == nil {
		ps.incomplete = hookPending
	}
	return "", ps.incomplete
}

// provenEndFlags identifies a PDF for which the reference's only final flag
// is BAD_PDF_TRAILER. Without xref, encryption, streams or complex object
// names, there are no later flag writers after PRE. The reference invokes
// END only because that bit is nonzero. All other END paths stay undecided.
func (ps *pdfScan) provenEndFlags(data []byte) (uint32, bool) {
	if ps.incomplete != nil || ps.hookState == nil || !ps.hookState.ObjectsKnown ||
		len(ps.hookState.Objects) == 0 || len(ps.streams) != 0 ||
		!bytes.HasPrefix(data, []byte("%PDF-1.7")) ||
		bytes.Contains(data, []byte("%%EOF")) || bytes.Contains(data, []byte("xref")) ||
		bytes.Contains(data, []byte("/Encrypt")) {
		return 0, false
	}
	for _, obj := range ps.hookState.Objects {
		if !obj.FlagsKnown {
			return 0, false
		}
	}
	return 1 << 2, true // BAD_PDF_TRAILER
}

// bytecodeState exposes only object fields whose discovery order and bounds
// this parser can prove. ClamAV records each indirect object just after its
// "obj" keyword and reports sizes from the next start minus four bytes.
// When hidden headers cause this parser to see a different object set, the
// reference-order fallback proves IDs, starts and simple flags directly from
// the raw PDF; complex flags remain unknown. Object-stream members are not
// indirect objects in the outer PARSED-phase table.
func (ps *pdfScan) bytecodeState(data []byte) *bytecode.PDFState {
	state := &bytecode.PDFState{}
	if ps.incomplete != nil || ps.objects != len(ps.hookObjects) {
		return state
	}
	if ps.hookUncertain {
		objects, ok := pdfReferenceObjects(data)
		if !ok || len(objects) == 0 {
			return state
		}
		state.Objects = make([]bytecode.PDFObject, len(objects))
		for i, obj := range objects {
			end := len(data)
			if i+1 < len(objects) {
				end = objects[i+1].start - 4
			}
			if obj.start < 0 || obj.end < obj.start || obj.end > len(data) || end < obj.start || end > len(data) {
				return &bytecode.PDFState{}
			}
			body := bytes.TrimSpace(data[obj.start:obj.end])
			kind := byte('n')
			if bytes.HasPrefix(body, []byte("<<")) {
				kind = 'd'
			}
			flags, flagsKnown := pdfSimpleObjectFlags(body, kind, bytes.Contains(body, []byte("\nstream\n")))
			state.Objects[i] = bytecode.PDFObject{
				ID: obj.id, Offset: uint32(obj.start), OffsetKnown: true,
				Size: uint32(end - obj.start), SizeKnown: true,
				Raw: data[obj.start:end], RawKnown: true,
				Flags: flags, FlagsKnown: flagsKnown,
			}
		}
		for _, obj := range state.Objects {
			if !obj.FlagsKnown {
				for i := range state.Objects {
					state.Objects[i].FlagsKnown = false
				}
				break
			}
		}
		state.ObjectsKnown = true
		return state
	}
	objects := append([]pdfHookObject(nil), ps.hookObjects...)
	streamObjects := make(map[pdfRef]bool, len(ps.streams))
	for _, stream := range ps.streams {
		streamObjects[stream.object] = true
	}
	objectCounts := make(map[pdfRef]int, len(objects))
	for _, obj := range objects {
		objectCounts[obj.ref]++
	}
	sort.Slice(objects, func(i, j int) bool { return objects[i].offset < objects[j].offset })
	state.Objects = make([]bytecode.PDFObject, 0, len(objects))
	for i, obj := range objects {
		if obj.ref.object < 0 || obj.ref.object > 0xffffff || obj.ref.generation < 0 || obj.ref.generation > 0xffff || obj.offset < 0 || obj.offset > len(data) || obj.end < obj.offset || obj.end > len(data) ||
			i > 0 && obj.offset <= objects[i-1].offset+4 {
			return &bytecode.PDFState{}
		}
		end := len(data)
		if i+1 < len(objects) {
			end = objects[i+1].offset - 4
		}
		if end < obj.offset || end > len(data) {
			return &bytecode.PDFState{}
		}
		flags, flagsKnown := pdfSimpleObjectFlags(data[obj.offset:obj.end], obj.kind, streamObjects[obj.ref])
		if streamObjects[obj.ref] && objectCounts[obj.ref] != 1 {
			flagsKnown = false
		}
		state.Objects = append(state.Objects, bytecode.PDFObject{
			ID:     uint32(obj.ref.object<<8) | uint32(obj.ref.generation&0xff),
			Offset: uint32(obj.offset), OffsetKnown: true,
			Size: uint32(end - obj.offset), SizeKnown: true,
			Raw: data[obj.offset:end], RawKnown: true,
			Flags: flags, FlagsKnown: flagsKnown,
		})
	}
	// A complex object can mark a different object while ClamAV parses an
	// indirect reference. Expose flags only when the entire object set uses
	// the independent, exact subset below.
	for _, obj := range state.Objects {
		if !obj.FlagsKnown {
			for i := range state.Objects {
				state.Objects[i].FlagsKnown = false
			}
			break
		}
	}
	state.ObjectsKnown = true
	return state
}

// object parses an indirect object's value, optional stream and endobj after
// its header. Scalar values are kept for reference resolution; replace lets a
// later sequential definition supersede an earlier one, as an incremental
// update does.
func (ps *pdfScan) object(p *pdfParser, ref pdfRef, replace bool) error {
	// Objects are not files: like the reference, which extracts streams
	// and counts those, only the data analyzed (streams, strings) counts
	// against the file limit; the parser's token budget bounds the rest.
	ps.objects++
	headerEnd := p.pos
	record := func(kind byte) {
		ps.hookObjects = append(ps.hookObjects, pdfHookObject{ref: ref, offset: headerEnd, end: p.pos - len("endobj"), kind: kind})
	}
	if p.space(); p.keyword("endobj") {
		record('n')
		return nil // an empty object, which readers take as null
	}
	value, err := p.value(0)
	if ps.flushStrings(); ps.signature != "" || err != nil {
		return err
	}
	if _, exists := ps.refs[ref]; (replace || !exists) && (value.kind == 'i' || value.kind == 'r' || value.kind == 'n') {
		ps.refs[ref] = value
	}
	if value.kind == 'd' && len(ps.encryptDicts) < 64 {
		if _, ok := value.dict["Filter"]; ok {
			if _, ok := value.dict["U"]; ok {
				if ps.encryptDicts == nil {
					ps.encryptDicts = make(map[pdfRef]map[string]pdfValue)
				}
				ps.encryptDicts[ref] = value.dict
			}
		}
	}
	p.space()
	// The reference accepts successive literal strings and then an
	// ordinary dictionary in one indirect object. Its flag walker uses
	// the first dictionary-like pair in the raw object, including one
	// inside a string. Each additional value consumes a parser token and
	// bytes from the existing document budget.
	if value.kind == 's' {
		for p.pos < len(p.data) {
			literal := p.data[p.pos] == '('
			dict := p.pos+1 < len(p.data) && p.data[p.pos] == '<' && p.data[p.pos+1] == '<'
			if !literal && !dict {
				break
			}
			if _, err := p.value(0); err != nil {
				return err
			}
			if ps.flushStrings(); ps.signature != "" {
				return nil
			}
			p.space()
			if dict {
				break
			}
		}
	}
	if value.kind == 'd' && p.keyword("stream") {
		stream, err := p.stream(value.dict)
		if err != nil {
			return err
		}
		stream.object = ref
		ps.streams = append(ps.streams, stream)
		p.space()
	}
	if !p.keyword("endobj") {
		return ErrMalformed
	}
	record(value.kind)
	return nil
}

// embeddedObjects parses "N G obj" headers that the sequential pass did not
// read, such as an object placed inside another object's stream body, string
// or comment. Cross-reference offsets and reconstructing readers resolve such
// objects, so their strings and streams are analyzed as well. A header whose
// object cannot be parsed leaves the document incomplete.
func (ps *pdfScan) embeddedObjects(data []byte, headers map[int]bool, depth int) error {
	var incomplete error
	// These objects are parsed where they sit, so their parses may overlap
	// each other and run on to the end of the data. Together they may examine
	// at most the document's own size, which keeps the pass linear; objects
	// that do not overlap always fit.
	budget := len(data)
	for from := 0; ; {
		if err := ps.s.ctx.Err(); err != nil {
			return err
		}
		i := bytes.Index(data[from:], []byte("obj"))
		if i < 0 {
			return incomplete
		}
		k := from + i
		from = k + 3
		if headers[k] || k == 0 || !pdfWhite(data[k-1]) || k+3 < len(data) && !pdfDelimiter(data[k+3]) {
			continue
		}
		start, ok := pdfHeaderStart(data, k)
		if !ok {
			continue
		}
		p := &pdfParser{owner: ps, data: data, pos: start, depth: depth}
		ref, object, err := p.objectHeader()
		if err != nil {
			return err
		}
		if !object || p.pos != k+3 {
			continue
		}
		headers[k] = true
		err = ps.object(p, ref, false)
		// The parser position is the furthest byte the parse examined.
		budget -= p.pos - start
		if ps.signature != "" {
			return nil
		} else if err == ErrLimit || ps.s.ctx.Err() != nil {
			return err
		} else if err != nil {
			incomplete = err
		}
		if budget < 0 {
			return ErrLimit
		}
	}
}

// pdfHeaderStart returns the start of the object number of an "N G obj"
// header whose keyword begins at k. The tokens may be separated by white
// space and end-of-line comments, which a lexer started at a cross-reference
// offset skips. A cross-reference offset may point into a longer digit run,
// so the object number is its last digits.
func pdfHeaderStart(data []byte, k int) (int, bool) {
	pos := k
	separator := func() bool {
		start := pos
		for pos > 0 && start-pos < 1024 && pdfWhite(data[pos-1]) {
			pos--
			if data[pos] != '\n' && data[pos] != '\r' {
				continue
			}
			// A comment on the previous line extends to this line break.
			line := pos
			for line > 0 && pos-line < 256 && data[line-1] != '\n' && data[line-1] != '\r' {
				line--
			}
			if c := bytes.IndexByte(data[line:pos], '%'); c >= 0 {
				pos = line + c
			}
		}
		return pos < start
	}
	number := func(digits int) bool {
		end := pos
		for pos > 0 && end-pos < digits && data[pos-1] >= '0' && data[pos-1] <= '9' {
			pos--
		}
		return pos < end
	}
	// The whole generation token is read, leading zeros included; the object
	// number keeps the digits a signed 64-bit value can hold.
	if !separator() || !number(64) || pos > 0 && data[pos-1] >= '0' && data[pos-1] <= '9' || !separator() || !number(18) {
		return 0, false
	}
	return pos, true
}

func pdfWhite(b byte) bool     { return b == 0 || b == 9 || b == 10 || b == 12 || b == 13 || b == 32 }
func pdfDelimiter(b byte) bool { return pdfWhite(b) || bytes.IndexByte([]byte("()<>[]{}/%"), b) >= 0 }
func (p *pdfParser) space() {
	for p.pos < len(p.data) {
		if pdfWhite(p.data[p.pos]) {
			p.pos++
			continue
		}
		if p.data[p.pos] == '%' {
			start := p.pos
			for p.pos < len(p.data) && p.data[p.pos] != '\n' && p.data[p.pos] != '\r' {
				p.pos++
			}
			if p.dictDepth > 0 {
				p.noteCommentNames(p.data[start:p.pos])
			}
			continue
		}
		break
	}
}

// ClamAV's dictionary-name pass also sees escaped /names in comments between
// dictionary tokens. The structured parser otherwise skips those comments.
func (p *pdfParser) noteCommentNames(comment []byte) {
	for len(comment) > 0 {
		i := bytes.IndexByte(comment, '/')
		if i < 0 {
			return
		}
		comment = comment[i:]
		candidate := pdfParser{owner: p.owner, data: comment, dictDepth: p.dictDepth}
		name, escaped, err := candidate.name()
		if err == nil && escaped {
			p.noteObfuscatedName(name)
		}
		comment = comment[max(candidate.pos, 1):]
	}
}
func (p *pdfParser) step() error {
	ps := p.owner
	ps.tokens++
	if int64(ps.tokens) > 1_000_000+ps.size+max(0, ps.budgetStart-ps.s.remaining) {
		return ErrLimit
	}
	return p.owner.s.ctx.Err()
}
func (p *pdfParser) word() (string, error) {
	if err := p.step(); err != nil {
		return "", err
	}
	p.space()
	start := p.pos
	for p.pos < len(p.data) && !pdfDelimiter(p.data[p.pos]) {
		p.pos++
		if p.pos-start > 256 {
			return "", ErrLimit
		}
	}
	if start == p.pos {
		return "", ErrMalformed
	}
	return string(p.data[start:p.pos]), nil
}
func (p *pdfParser) keyword(word string) bool {
	end := p.pos + len(word)
	if end > len(p.data) || string(p.data[p.pos:end]) != word || (end < len(p.data) && !pdfDelimiter(p.data[end])) {
		return false
	}
	p.pos = end
	return true
}
func (p *pdfParser) objectHeader() (pdfRef, bool, error) {
	start := p.pos
	a, e := p.word()
	if e != nil {
		p.pos = start
		if e == ErrLimit || p.owner.s.ctx.Err() != nil {
			return pdfRef{}, false, e
		}
		return pdfRef{}, false, nil
	}
	n, e := strconv.ParseInt(a, 10, 64)
	if e != nil || n <= 0 {
		p.pos = start
		return pdfRef{}, false, nil
	}
	b, e := p.word()
	if e != nil {
		p.pos = start
		if e == ErrLimit || p.owner.s.ctx.Err() != nil {
			return pdfRef{}, false, e
		}
		return pdfRef{}, false, nil
	}
	g, e := strconv.ParseInt(b, 10, 64)
	if e != nil || g < 0 || g > 65535 {
		p.pos = start
		return pdfRef{}, false, nil
	}
	p.space()
	if !p.keyword("obj") {
		p.pos = start
		return pdfRef{}, false, nil
	}
	return pdfRef{n, g}, true, nil
}
func (p *pdfParser) value(nesting int) (pdfValue, error) {
	if err := p.step(); err != nil {
		return pdfValue{}, err
	}
	if nesting > 64 {
		return pdfValue{}, ErrLimit
	}
	p.space()
	if p.pos >= len(p.data) {
		return pdfValue{}, ErrMalformed
	}
	switch p.data[p.pos] {
	case '(':
		start := p.pos + 1
		decoded, err := p.literal()
		// A string without escapes is already matched as part of the raw bytes.
		if err == nil && !bytes.Equal(decoded, p.data[start:p.pos-1]) {
			p.scanString(decoded)
		}
		return pdfStringValue(decoded), err
	case '/':
		name, escaped, err := p.name()
		if escaped && err == nil {
			p.scanString(append([]byte{'/'}, []byte(name)...))
			p.noteObfuscatedName(name)
		}
		return pdfValue{kind: 'n', name: name}, err
	case '[':
		p.pos++
		v := pdfValue{kind: 'a'}
		for {
			p.space()
			if p.pos >= len(p.data) {
				return v, ErrMalformed
			}
			if p.data[p.pos] == ']' {
				p.pos++
				return v, nil
			}
			if len(v.array) >= 4096 {
				return v, ErrLimit
			}
			x, err := p.value(nesting + 1)
			if err != nil {
				return v, err
			}
			v.array = append(v.array, x)
			if p.owner.signature != "" {
				return v, nil
			}
		}
	case '<':
		if p.pos+1 >= len(p.data) || p.data[p.pos+1] != '<' {
			p.pos++
			decoded, used, err := pdfHex(p.data[p.pos:], min(p.owner.s.maxFile, p.owner.s.remaining))
			p.pos += used
			if err == nil {
				p.scanString(decoded)
			}
			return pdfStringValue(decoded), err
		}
		p.pos += 2
		p.dictDepth++
		defer func() { p.dictDepth-- }()
		v := pdfValue{kind: 'd', dict: make(map[string]pdfValue)}
		for {
			p.space()
			if p.pos+1 < len(p.data) && p.data[p.pos] == '>' && p.data[p.pos+1] == '>' {
				p.pos += 2
				p.owner.noteEncryption(v.dict)
				return v, nil
			}
			if p.pos >= len(p.data) || p.data[p.pos] != '/' {
				return v, ErrMalformed
			}
			if len(v.dict) >= 4096 {
				return v, ErrLimit
			}
			name, escaped, err := p.name()
			if err != nil {
				return v, err
			}
			if escaped {
				p.scanString(append([]byte{'/'}, []byte(name)...))
				p.noteObfuscatedName(name)
			}
			x, err := p.value(nesting + 1)
			if err != nil {
				return v, err
			}
			// A repeated key keeps its first value, as the reference reads
			// it (checked with two /Filter entries).
			if _, exists := v.dict[name]; !exists {
				v.dict[name] = x
			}
			p.owner.noteDataStreams(name, x)
			if p.owner.signature != "" {
				return v, nil
			}
		}
	case ']', '>', ')':
		return pdfValue{}, ErrMalformed
	}
	word, err := p.word()
	if err != nil {
		return pdfValue{}, err
	}
	if n, e := strconv.ParseInt(word, 10, 64); e == nil {
		v := pdfValue{kind: 'i', number: n}
		saved := p.pos
		next, e := p.word()
		if e == ErrLimit || p.owner.s.ctx.Err() != nil {
			return v, e
		}
		if e == nil {
			generation, e := strconv.ParseInt(next, 10, 64)
			p.space()
			if e == nil && n > 0 && generation >= 0 && generation <= 65535 && p.keyword("R") {
				return pdfValue{kind: 'r', ref: pdfRef{n, generation}}, nil
			}
		}
		p.pos = saved
		return v, nil
	}
	return pdfValue{kind: 'k', name: word}, nil
}

// Only names inside PDF object dictionaries qualify. The set is the names
// marked NAMEFLAG_HEURISTIC in ClamAV 1.5.4's pdfname_actions table; ordinary
// metadata names (for example Author) and arbitrary escaped names do not.
func (p *pdfParser) noteObfuscatedName(name string) {
	if p.dictDepth == 0 {
		return
	}
	switch name {
	case "ASCIIHexDecode", "ASCII85Decode", "A85", "AHx", "EmbeddedFile",
		"FlateDecode", "Fl", "Image", "LZWDecode", "LZW", "RunLengthDecode",
		"RL", "CCITTFaxDecode", "CCF", "JBIG2Decode", "DCTDecode", "DCT",
		"JPXDecode", "Crypt", "Standard", "Sig", "V", "R", "Linearized",
		"Filter", "JavaScript", "Length", "S", "Type", "OpenAction",
		"Launch", "Page", "Contents":
		p.owner.obfuscatedName = true
	}
}

// scanString keeps a string whose decoded bytes differ from its raw bytes for
// analysis with the other strings of its object. Each analyzed object costs a
// full rule pass, and a text-heavy document holds thousands of strings.
// Strings shorter than the smallest analyzed object were never analyzed.
func (p *pdfParser) scanString(data []byte) {
	ps := p.owner
	if ps.signature != "" || len(data) < minAnalyzedObject {
		return
	}
	if len(ps.strings) > 0 && (ps.stringDepth != p.depth+1 || int64(ps.stringBytes+len(data)) > min(ps.s.maxFile, ps.s.remaining)) {
		ps.flushStrings()
	}
	ps.strings = append(ps.strings, data)
	ps.stringBytes += len(data)
	ps.stringDepth = p.depth + 1
}

// flushStrings analyzes the kept strings of one object as one representation:
// a single string as it is, several joined in order.
func (ps *pdfScan) flushStrings() {
	kept, depth := ps.strings, ps.stringDepth
	ps.strings, ps.stringBytes = nil, 0
	if ps.signature != "" || len(kept) == 0 {
		return
	}
	joined := kept[0]
	if len(kept) > 1 {
		joined = nil
		for _, str := range kept {
			joined = append(joined, str...)
		}
	}
	if sig, err := ps.s.visit(joined, depth); sig != "" {
		ps.signature = sig
	} else if err != nil {
		ps.incomplete = err
	}
}

// noteDataStreams records the objects a dictionary entry names as an
// attachment (the /EF entries of a file specification), a script, XML
// metadata or an XFA form.
func (ps *pdfScan) noteDataStreams(key string, v pdfValue) {
	var refs []pdfValue
	switch key {
	case "EF":
		for _, x := range v.dict {
			refs = append(refs, x)
		}
	case "JS", "Metadata":
		refs = []pdfValue{v}
	case "XFA":
		refs = append([]pdfValue{v}, v.array...)
	}
	for _, x := range refs {
		if x.kind != 'r' {
			continue
		}
		if ps.dataStreams == nil {
			ps.dataStreams = make(map[pdfRef]bool)
		}
		ps.dataStreams[x.ref] = true
	}
}

// dataStream reports a stream that holds a file or another language rather
// than PDF content syntax: one named by noteDataStreams, or one whose own
// dictionary marks it as an embedded file (its /Type, /Params or MIME
// /Subtype), as XML, as a function (a type 4 function is PostScript
// calculator code) or as a PostScript XObject.
func (ps *pdfScan) dataStream(stream pdfStream) bool {
	if ps.dataStreams[stream.object] {
		return true
	}
	typ, _ := ps.resolve(stream.dict["Type"])
	subtype, _ := ps.resolve(stream.dict["Subtype"])
	_, params := stream.dict["Params"]
	_, function := stream.dict["FunctionType"]
	return params || function || typ.kind == 'n' && (typ.name == "EmbeddedFile" || typ.name == "Metadata") ||
		subtype.kind == 'n' && (subtype.name == "XML" || subtype.name == "PS" || strings.Contains(subtype.name, "/"))
}
func pdfNibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}
func (p *pdfParser) name() (string, bool, error) {
	if err := p.step(); err != nil {
		return "", false, err
	}
	if p.pos >= len(p.data) || p.data[p.pos] != '/' {
		return "", false, ErrMalformed
	}
	p.pos++
	var decoded []byte
	escaped := false
	for p.pos < len(p.data) && !pdfDelimiter(p.data[p.pos]) {
		b := p.data[p.pos]
		p.pos++
		if b == '#' {
			if len(p.data)-p.pos < 2 {
				return "", false, ErrMalformed
			}
			a, ok := pdfNibble(p.data[p.pos])
			c, ok2 := pdfNibble(p.data[p.pos+1])
			if !ok || !ok2 {
				return "", false, ErrMalformed
			}
			b = a<<4 | c
			p.pos += 2
			escaped = true
		}
		if b == 0 {
			return "", false, ErrMalformed
		}
		if len(decoded) >= 256 {
			return "", false, ErrLimit
		}
		decoded = append(decoded, b)
	}
	return string(decoded), escaped, nil
}
func (p *pdfParser) literal() ([]byte, error) {
	p.pos++
	balance := 1
	var out []byte
	limit := min(p.owner.s.maxFile, p.owner.s.remaining)
	for p.pos < len(p.data) {
		if p.pos&4095 == 0 {
			if err := p.owner.s.ctx.Err(); err != nil {
				return nil, err
			}
		}
		b := p.data[p.pos]
		p.pos++
		switch b {
		case '(':
			balance++
			if balance > 64 {
				return nil, ErrLimit
			}
		case ')':
			balance--
			if balance == 0 {
				return out, nil
			}
		case '\\':
			if p.pos >= len(p.data) {
				return nil, ErrMalformed
			}
			b = p.data[p.pos]
			p.pos++
			switch b {
			case 'n':
				b = '\n'
			case 'r':
				b = '\r'
			case 't':
				b = '\t'
			case 'b':
				b = '\b'
			case 'f':
				b = '\f'
			case '\r':
				if p.pos < len(p.data) && p.data[p.pos] == '\n' {
					p.pos++
				}
				continue
			case '\n':
				continue
			default:
				if b >= '0' && b <= '7' {
					n := int(b - '0')
					for i := 0; i < 2 && p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '7'; i++ {
						n = n*8 + int(p.data[p.pos]-'0')
						p.pos++
					}
					b = byte(n)
				}
			}
		case '\r':
			if p.pos < len(p.data) && p.data[p.pos] == '\n' {
				p.pos++
			}
			b = '\n'
		}
		if int64(len(out)) >= limit {
			return nil, ErrLimit
		}
		out = append(out, b)
	}
	return nil, ErrMalformed
}
func (ps *pdfScan) resolve(v pdfValue) (pdfValue, error) {
	for i := 0; v.kind == 'r'; i++ {
		if i >= 16 {
			return v, ErrLimit
		}
		next, ok := ps.refs[v.ref]
		if !ok {
			return v, ErrUnsupported
		}
		v = next
	}
	return v, nil
}
func (p *pdfParser) stream(dict map[string]pdfValue) (pdfStream, error) {
	stream := pdfStream{dict: dict, data: p.data}
	if p.pos >= len(p.data) {
		return stream, ErrMalformed
	}
	if p.data[p.pos] == '\r' {
		p.pos++
		if p.pos < len(p.data) && p.data[p.pos] == '\n' {
			p.pos++
		}
	} else if p.data[p.pos] == '\n' {
		p.pos++
	} else {
		return stream, ErrMalformed
	}
	stream.start = p.pos
	length, err := p.owner.resolve(dict["Length"])
	if err != nil && dict["Length"].kind == 'r' {
		// A length defined later in the file, as Ghostscript writes it
		// (with the data running straight into endstream).
		if p.owner.forward == nil {
			p.owner.forward = pdfIntegerObjects(p.data)
		}
		if n, ok := p.owner.forward[dict["Length"].ref]; ok {
			length, err = pdfValue{kind: 'i', number: n}, nil
		}
	}
	if err == nil && length.kind == 'i' && length.number >= 0 && length.number <= int64(len(p.data)-p.pos) {
		end := p.pos + int(length.number)
		p.pos = pdfSkipPadding(p.data, end)
		stream.terminator = p.pos
		if p.keyword("endstream") {
			stream.end = end
			return stream, nil
		}
	}
	// Forward indirect lengths are resolved after all plain objects have been
	// read. A delimiter guess is never trusted without that later exact check.
	for start := stream.start; start < len(p.data); {
		i := bytes.Index(p.data[start:], []byte("endstream"))
		if i < 0 {
			break
		}
		i += start
		// As the reference calculates a length: the first endstream keyword,
		// whatever precedes it (writers that give an exact length need not
		// end the data with an end of line).
		if i+9 == len(p.data) || pdfDelimiter(p.data[i+9]) {
			stream.terminator = i
			stream.end = i
			if stream.end > stream.start && p.data[stream.end-1] == '\n' {
				stream.end--
				if stream.end > stream.start && p.data[stream.end-1] == '\r' {
					stream.end--
				}
			} else if stream.end > stream.start && p.data[stream.end-1] == '\r' {
				stream.end--
			}
			p.pos = i + 9
			return stream, nil
		}
		start = i + 9
	}
	// The search examined the rest of the data.
	p.pos = len(p.data)
	return stream, ErrMalformed
}

// pdfIntegerObjects finds the objects defined as a plain integer
// ("N G obj <integer> endobj") anywhere in data; a later definition replaces
// an earlier one, as an incremental update does.
func pdfIntegerObjects(data []byte) map[pdfRef]int64 {
	out := make(map[pdfRef]int64)
	digitsBefore := func(end int) int {
		i := end
		for i > 0 && i > end-18 && data[i-1] >= '0' && data[i-1] <= '9' {
			i--
		}
		return i
	}
	whiteBefore := func(end int) int {
		i := end
		for i > 0 && pdfWhite(data[i-1]) {
			i--
		}
		return i
	}
	for from := 0; ; {
		j := bytes.Index(data[from:], []byte("obj"))
		if j < 0 {
			return out
		}
		at := from + j
		from = at + 3
		// before: object number, white space, generation, white space
		genEnd := whiteBefore(at)
		genStart := digitsBefore(genEnd)
		if genEnd == at || genStart == genEnd {
			continue
		}
		numEnd := whiteBefore(genStart)
		numStart := digitsBefore(numEnd)
		if numEnd == genStart || numStart == numEnd || numStart > 0 && !pdfDelimiter(data[numStart-1]) {
			continue
		}
		// after: white space, an integer, white space, endobj
		k := at + 3
		for k < len(data) && pdfWhite(data[k]) {
			k++
		}
		if k == at+3 {
			continue
		}
		valueStart := k
		if k < len(data) && (data[k] == '-' || data[k] == '+') {
			k++
		}
		digits := k
		for k < len(data) && k < digits+18 && data[k] >= '0' && data[k] <= '9' {
			k++
		}
		if k == digits {
			continue
		}
		valueEnd := k
		for k < len(data) && pdfWhite(data[k]) {
			k++
		}
		if k == valueEnd || !bytes.HasPrefix(data[k:], []byte("endobj")) || k+6 < len(data) && !pdfDelimiter(data[k+6]) {
			continue
		}
		n, e1 := strconv.ParseInt(string(data[numStart:numEnd]), 10, 64)
		g, e2 := strconv.ParseInt(string(data[genStart:genEnd]), 10, 64)
		v, e3 := strconv.ParseInt(string(data[valueStart:valueEnd]), 10, 64)
		if e1 == nil && e2 == nil && e3 == nil && n > 0 && g <= 65535 {
			out[pdfRef{n, g}] = v
		}
	}
}

// pdfMaxStreamPadding bounds the white space accepted between a stream's
// declared end and its endstream keyword, where writers put one end-of-line
// marker. Without a bound, every stream whose length points into one long
// white-space run would skip that run again.
const pdfMaxStreamPadding = 4096

func pdfSkipPadding(data []byte, pos int) int {
	for limit := min(len(data), pos+pdfMaxStreamPadding); pos < limit && pdfWhite(data[pos]); {
		pos++
	}
	return pos
}

// scanStream analyzes a stream's decoded data. The reference extracts a
// stream only when it has no filter or one of Flate, ASCIIHex, ASCII85, DCT
// and JBIG2 (checked on streams with each filter, attachments and scripts
// among them); this engine analyzes the others too, but what it cannot
// analyze in them is not analysis the reference makes.
func (ps *pdfScan) scanStream(stream pdfStream, depth int) (string, error) {
	sig, err := ps.analyzeStream(stream, depth)
	if sig != "" || err == nil || ps.s.ctx.Err() != nil || errors.Is(err, ErrLimit) || errors.Is(err, signatures.ErrMatchLimit) {
		return sig, err
	}
	if filters, ferr := ps.filters(stream.dict["Filter"]); ferr == nil && !pdfReferenceExtracts(filters) {
		return "", nil
	}
	return "", err
}

// pdfReferenceExtracts reports a filter list whose stream the reference
// extracts.
func pdfReferenceExtracts(filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, f := range filters {
		switch f {
		case "FlateDecode", "Fl", "ASCIIHexDecode", "AHx", "ASCII85Decode", "A85", "DCTDecode", "DCT", "JBIG2Decode":
			return true
		}
	}
	return false
}

func (ps *pdfScan) analyzeStream(stream pdfStream, depth int) (string, error) {
	var incomplete error
	// The declared length is used when endstream follows it; otherwise the
	// end found at endstream stands, as the reference calculates the length
	// (it reads streams whose /Length is missing, too short or too long).
	if length, err := ps.resolve(stream.dict["Length"]); err == nil && length.kind == 'i' &&
		length.number >= 0 && length.number <= int64(len(stream.data)-stream.start) {
		if end := stream.start + int(length.number); pdfSkipPadding(stream.data, end) == stream.terminator {
			stream.end = end
		}
	}
	if v, exists := stream.dict["F"]; exists && !(v.kind == 'k' && v.name == "null") {
		return "", ErrUnsupported
	}
	filters, err := ps.filters(stream.dict["Filter"])
	if err != nil {
		return "", err
	}
	referenceExtracts := pdfReferenceExtracts(filters)
	// Decode parameters are not read: the reference does not undo
	// predictors, and decodes LZW with early change 1 whatever /EarlyChange
	// says.
	data, err := ps.decryptStream(stream, stream.data[stream.start:stream.end])
	if err != nil {
		return "", err
	}
	// As in the reference: decode up to the first filter neither engine
	// decodes, skipping names it does not know; the data decoded so far is
	// the stream's object.
	last := -1
	var decodeIncomplete error
	var prepaid int64
	for i, filter := range filters {
		if kind := pdfFilterKind(filter); kind == pdfStopFilter {
			filters = filters[:i]
			break
		} else if kind == pdfDecodedFilter {
			last = i
		}
	}
	for i, filter := range filters {
		if err := ps.s.ctx.Err(); err != nil {
			return "", err
		}
		if pdfFilterKind(filter) != pdfDecodedFilter {
			continue
		}
		decoded, err := ps.decode(filter, data)
		if err != nil {
			if !errors.Is(err, ErrMalformed) {
				// Resource, cancellation and unsupported-filter failures must
				// retain their ordinary bounds and error handling.
				ps.s.spend(len(decoded))
				return "", err
			}
			// A damaged Flate stream can emit useful bytes before failing.
			// Other failed PDF decoders discard their tentative output in
			// the reference. In that case scan the previous stage instead.
			decodeIncomplete = err
			if len(decoded) == 0 || filter != "FlateDecode" && filter != "Fl" {
				ps.s.spend(len(decoded))
				break
			}
		}
		data = decoded
		prepaid = 0
		if i < last {
			// An intermediate filter output is decoding work, charged to the
			// budget, but not an object: the reference matches only the
			// fully decoded stream.
			if int64(len(data)) > ps.s.remaining {
				return "", ErrLimit
			}
			ps.s.remaining -= int64(len(data))
			prepaid = int64(len(data))
		}
	}
	incomplete = decodeIncomplete
	if sig, err := ps.s.visitHeld(data, depth+1, prepaid); sig != "" {
		return sig, nil
	} else if err != nil {
		incomplete = err
	}
	// ClamAV invokes POSTDUMP hooks only after a nonempty extracted object
	// has passed its child scan. The direct signature matcher probes this
	// phase conservatively; the parser can now promote a reached label to a
	// detection for streams that the reference actually extracts.
	postDumpData := data
	if len(filters) == 0 && len(data) == 0 && stream.end == stream.start && stream.terminator > stream.end {
		// For /Length 0, the reference falls back to its measured stream
		// extent. A line break before endstream is then a one-byte object.
		postDumpData = stream.data[stream.end:stream.terminator]
	}
	if referenceExtracts && len(postDumpData) > 0 && incomplete == nil {
		if sig, err := ps.s.db.MatchPDFPostDump(ps.s.ctx, postDumpData, signatures.MatchContext{
			BytecodeOptions:    ps.s.bytecodeEngineOptions(),
			Metadata:           ps.s.representationMetadata(),
			MetadataPartial:    ps.s.collectMetadata,
			ScanExtractedTyped: ps.s.bytecodeChildScanner(depth + 1),
			PDF:                ps.hookState,
		}); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		}
	}
	if ps.objectStreamDict(stream.dict) {
		if sig, err := ps.objectStream(data, stream.dict, depth+1); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		}
	} else if pdfText(data) && !ps.dataStream(stream) {
		// PDF page-content streams may hide bytes in literal/hex strings, which
		// this engine also analyzes (the reference does not). Binary streams
		// still receive raw matching. Attachments, scripts and XML were
		// analyzed above as the data they are.
		p := &pdfParser{owner: ps, data: data, depth: depth + 1}
		for {
			p.space()
			if p.pos >= len(data) {
				break
			}
			// The reference does not search content streams for strings:
			// where this search cannot go on (binary image data, an inline
			// image), it stops without making the stream incomplete.
			v, err := p.value(0)
			if ps.signature != "" {
				return ps.signature, nil
			}
			if err != nil || v.kind == 'k' && (v.name == "BI" || v.name == "stream") {
				break
			}
		}
		// The strings of one content stream are analyzed together.
		if ps.flushStrings(); ps.signature != "" {
			return ps.signature, nil
		}
	}
	return "", incomplete
}

// objectStreamDict reports the dictionary of a stream of compressed objects.
func (ps *pdfScan) objectStreamDict(dict map[string]pdfValue) bool {
	typ, _ := ps.resolve(dict["Type"])
	return typ.kind == 'n' && typ.name == "ObjStm"
}
func pdfText(data []byte) bool {
	for _, b := range data[:min(len(data), 4096)] {
		if b < 32 && !pdfWhite(b) {
			return false
		}
	}
	return true
}
func (ps *pdfScan) filters(v pdfValue) ([]string, error) {
	if v.kind == 0 || (v.kind == 'k' && v.name == "null") {
		return nil, nil
	}
	v, err := ps.resolve(v)
	if err != nil {
		return nil, err
	}
	if v.kind == 'n' {
		return []string{v.name}, nil
	}
	if v.kind != 'a' {
		return nil, ErrUnsupported
	}
	if len(v.array) > 8 {
		return nil, ErrLimit
	}
	var out []string
	for _, x := range v.array {
		x, err = ps.resolve(x)
		if err != nil || x.kind != 'n' {
			return nil, ErrUnsupported
		}
		out = append(out, x.name)
	}
	return out, nil
}
func (ps *pdfScan) decode(filter string, data []byte) ([]byte, error) {
	limit := min(ps.s.maxFile, ps.s.remaining)
	if limit < 0 {
		return nil, ErrLimit
	}
	switch filter {
	case "FlateDecode", "Fl":
		return ps.decodeFlate(data, limit)
	case "LZWDecode", "LZW":
		input := bytes.NewReader(data)
		r := lzw.NewReader(input, lzw.MSB, 8).(*lzw.Reader)
		r.SetAldusCompatible(true)
		return ps.readCompressed(r, limit)
	case "ASCIIHexDecode", "AHx":
		out, used, err := pdfHex(data, limit)
		if err == nil && !pdfOnlySpace(data[used:]) {
			err = ErrMalformed
		}
		return out, err
	case "ASCII85Decode", "A85":
		return pdfASCII85(data, limit)
	case "RunLengthDecode", "RL":
		return pdfRunLength(data, limit)
	default:
		return nil, ErrUnsupported
	}
}

// decodeFlate accepts two malformed-but-recoverable starts that the PDF
// reference reader handles: a stray CR before zlib data, and a bad first
// line followed by a valid zlib stream. Retry after the first line only when
// the initial decode produced no bytes; otherwise its bounded prefix is the
// object to scan.
func (ps *pdfScan) decodeFlate(data []byte, limit int64) ([]byte, error) {
	if len(data) > 0 && data[0] == '\r' {
		data = data[1:]
	}
	decode := func(b []byte) ([]byte, error) {
		r, err := zlib.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, ErrMalformed
		}
		return ps.readCompressed(r, limit)
	}
	out, err := decode(data)
	if err == nil || len(out) > 0 || !errors.Is(err, ErrMalformed) {
		return out, err
	}
	end := bytes.IndexAny(data, "\r\n")
	if end < 0 {
		return out, err
	}
	for end < len(data) && (data[end] == '\r' || data[end] == '\n') {
		end++
	}
	if end == len(data) {
		return out, err
	}
	return decode(data[end:])
}

// readCompressed returns the bytes decompressed so far with any error, so the
// caller can charge expansion work that did not produce a usable stream.
// Bytes after the end of the compressed data are accepted (the reference
// reports them as remaining).
func (ps *pdfScan) readCompressed(r io.ReadCloser, limit int64) ([]byte, error) {
	out, err := io.ReadAll(io.LimitReader(contextReader{ps.s.ctx, r}, limit+1))
	closeErr := r.Close()
	if ps.s.ctx.Err() != nil {
		return out, ps.s.ctx.Err()
	}
	if int64(len(out)) > limit {
		return out, ErrLimit
	}
	if err != nil || closeErr != nil {
		return out, ErrMalformed
	}
	return out, nil
}
func pdfOnlySpace(data []byte) bool {
	for _, b := range data {
		if !pdfWhite(b) {
			return false
		}
	}
	return true
}
func pdfHex(data []byte, limit int64) ([]byte, int, error) {
	var out []byte
	var high byte
	pending := false
	for i, b := range data {
		if pdfWhite(b) {
			continue
		}
		if b == '>' {
			if pending {
				if int64(len(out)) >= limit {
					return out, i + 1, ErrLimit
				}
				out = append(out, high<<4)
			}
			return out, i + 1, nil
		}
		n, ok := pdfNibble(b)
		if !ok {
			return out, i + 1, ErrMalformed
		}
		if !pending {
			high = n
			pending = true
			continue
		}
		if int64(len(out)) >= limit {
			return out, i + 1, ErrLimit
		}
		out = append(out, high<<4|n)
		pending = false
	}
	return out, len(data), ErrMalformed
}
func pdfASCII85(data []byte, limit int64) ([]byte, error) {
	var out []byte
	var group uint64
	count := 0
	appendGroup := func(n int) error {
		if group > 0xffffffff {
			return ErrMalformed
		}
		if int64(len(out)+n) > limit {
			return ErrLimit
		}
		for i := 0; i < n; i++ {
			out = append(out, byte(group>>uint(24-i*8)))
		}
		return nil
	}
	for i := 0; i < len(data); i++ {
		b := data[i]
		if pdfWhite(b) {
			continue
		}
		if b == '~' {
			if i+1 >= len(data) || data[i+1] != '>' || !pdfOnlySpace(data[i+2:]) || count == 1 {
				return out, ErrMalformed
			}
			if count > 1 {
				n := count
				for count < 5 {
					group = group*85 + 84
					count++
				}
				if err := appendGroup(n - 1); err != nil {
					return out, err
				}
			}
			return out, nil
		}
		if b == 'z' {
			if count != 0 {
				return out, ErrMalformed
			}
			group = 0
			if err := appendGroup(4); err != nil {
				return out, err
			}
			continue
		}
		if b < '!' || b > 'u' {
			return out, ErrMalformed
		}
		group = group*85 + uint64(b-'!')
		count++
		if count == 5 {
			if err := appendGroup(4); err != nil {
				return out, err
			}
			count = 0
			group = 0
		}
	}
	return out, ErrMalformed
}
func pdfRunLength(data []byte, limit int64) ([]byte, error) {
	var out []byte
	for i := 0; i < len(data); {
		n := int(data[i])
		i++
		if n == 128 {
			if !pdfOnlySpace(data[i:]) {
				return out, ErrMalformed
			}
			return out, nil
		}
		if n < 128 {
			n++
			if n > len(data)-i {
				return out, ErrMalformed
			}
			if int64(len(out)+n) > limit {
				return out, ErrLimit
			}
			out = append(out, data[i:i+n]...)
			i += n
		} else {
			n = 257 - n
			if i >= len(data) {
				return out, ErrMalformed
			}
			if int64(len(out)+n) > limit {
				return out, ErrLimit
			}
			for j := 0; j < n; j++ {
				out = append(out, data[i])
			}
			i++
		}
	}
	return out, ErrMalformed
}
func (ps *pdfScan) objectStream(data []byte, dict map[string]pdfValue, depth int) (string, error) {
	n, err := ps.resolve(dict["N"])
	if err != nil {
		return "", err
	}
	first, err := ps.resolve(dict["First"])
	if err != nil {
		return "", err
	}
	if n.kind != 'i' || first.kind != 'i' || n.number < 0 || first.number < 0 || first.number > int64(len(data)) {
		return "", ErrMalformed
	}
	// Each index entry takes at least four bytes ("1 0 ").
	if depth >= ps.s.maxDepth || n.number > first.number/4+1 {
		return "", ErrLimit
	}
	p := &pdfParser{owner: ps, data: data[:first.number], depth: depth}
	ids := make([]int64, 0, int(n.number))
	offsets := make([]int, 0, int(n.number))
	seen := make(map[int64]bool)
	for i := int64(0); i < n.number; i++ {
		a, err := p.word()
		if err != nil {
			return "", err
		}
		b, err := p.word()
		if err != nil {
			return "", err
		}
		id, e1 := strconv.ParseInt(a, 10, 64)
		off, e2 := strconv.ParseInt(b, 10, 64)
		if e1 != nil || e2 != nil || id <= 0 || seen[id] || off < 0 || off > int64(len(data))-first.number || (len(offsets) > 0 && off < int64(offsets[len(offsets)-1])) {
			return "", ErrMalformed
		}
		seen[id] = true
		ids = append(ids, id)
		offsets = append(offsets, int(off))
	}
	p.space()
	if p.pos != len(p.data) {
		return "", ErrMalformed
	}
	var incomplete error
	for i, off := range offsets {
		end := len(data)
		if i+1 < len(offsets) {
			end = int(first.number) + offsets[i+1]
		}
		p = &pdfParser{owner: ps, data: data[int(first.number)+off : end], depth: depth}
		v, err := p.value(0)
		if ps.flushStrings(); ps.signature != "" {
			return ps.signature, nil
		}
		if err != nil {
			incomplete = err
			continue
		}
		p.space()
		if p.pos != len(p.data) {
			incomplete = ErrMalformed
		}
		if v.kind == 'i' || v.kind == 'r' || v.kind == 'n' {
			ps.refs[pdfRef{ids[i], 0}] = v
		}
	}
	return "", incomplete
}

func pdfStringValue(b []byte) pdfValue {
	v := pdfValue{kind: 's'}
	if len(b) <= pdfMaxKeptString {
		v.str = b
	}
	return v
}

// noteEncryption records a trailer's (or cross-reference stream's)
// encryption entry and first file identifier; a later trailer, as an
// incremental update writes it, replaces an earlier one.
func (ps *pdfScan) noteEncryption(dict map[string]pdfValue) {
	e, ok := dict["Encrypt"]
	if !ok || e.kind == 'k' && e.name == "null" {
		return
	}
	ps.encrypt = &e
	ps.fileID = nil
	if id, ok := dict["ID"]; ok && id.kind == 'a' && len(id.array) > 0 && id.array[0].kind == 's' {
		ps.fileID = id.array[0].str
	}
}

// setUpDecryption derives the file key of an encrypted file before its
// streams are analyzed. A file this engine cannot decrypt is incomplete.
func (ps *pdfScan) setUpDecryption() {
	if ps.encrypt == nil {
		return
	}
	var dict map[string]pdfValue
	switch ps.encrypt.kind {
	case 'd':
		dict = ps.encrypt.dict
	case 'r':
		dict = ps.encryptDicts[ps.encrypt.ref]
	}
	var err error
	if dict == nil {
		err = ErrUnsupported
	} else {
		ps.crypt, err = newPDFCrypt(ps, dict, ps.fileID)
	}
	if err != nil {
		ps.crypt, ps.undecryptable = nil, true
		ps.incomplete = ErrUnsupported
	}
}

// decryptStream undoes a stream's encryption. Cross-reference streams are
// never encrypted, metadata is not when the file says so, and embedded files
// use their own crypt filter.
func (ps *pdfScan) decryptStream(stream pdfStream, data []byte) ([]byte, error) {
	if ps.crypt == nil && !ps.undecryptable {
		return data, nil
	}
	typ, _ := ps.resolve(stream.dict["Type"])
	if typ.kind == 'n' && typ.name == "XRef" {
		return data, nil
	}
	if ps.undecryptable {
		return nil, ErrUnsupported
	}
	with := ps.crypt.streams
	switch {
	case typ.kind == 'n' && typ.name == "Metadata" && !ps.crypt.encryptMetadata:
		with = pdfIdentity
	case typ.kind == 'n' && typ.name == "EmbeddedFile":
		with = ps.crypt.files
	}
	out, err := ps.crypt.decrypt(stream.object, with, data)
	if err != nil {
		return nil, err
	}
	ps.s.spend(len(out))
	return out, nil
}

const (
	pdfSkippedFilter = iota
	pdfDecodedFilter
	pdfStopFilter
)

// pdfFilterKind classifies a filter name the way the reference treats it
// (checked on chains around a Flate stream): the filters both engines decode,
// the ones the reference knows and does not decode (image filters and Crypt;
// it stops there), and the names it does not know (skipped).
func pdfFilterKind(name string) int {
	switch name {
	case "FlateDecode", "Fl", "LZWDecode", "LZW", "ASCIIHexDecode", "AHx", "ASCII85Decode", "A85", "RunLengthDecode", "RL":
		return pdfDecodedFilter
	case "DCTDecode", "DCT", "JPXDecode", "JBIG2Decode", "CCITTFaxDecode", "CCF", "Crypt":
		return pdfStopFilter
	}
	return pdfSkippedFilter
}
