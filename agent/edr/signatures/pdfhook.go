package signatures

import (
	"bytes"
	"context"
	"fmt"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// MatchPDFPhase runs a reached parser hook phase on its actual mapped input.
// For PARSED that is the PDF; for POSTDUMP it is the extracted object. The
// caller supplies the per-scan object snapshot when its facts are proved.
func (d *Database) MatchPDFPhase(ctx context.Context, input []byte, phase uint32, mc MatchContext) (string, error) {
	if len(d.pdfHooks) == 0 || phase == 0 {
		return "", nil
	}
	w := &matchWork{
		ctx: ctx, limit: d.limits.matchLimit(len(input)),
		bytecodeOptions: mc.BytecodeOptions, metadata: mc.Metadata,
		metadataPartial: mc.MetadataPartial, scanExtracted: mc.ScanExtracted, scanExtractedTyped: mc.ScanExtractedTyped,
	}
	var pending error
	for _, p := range d.pdfHooks {
		name, value, cause, err := runBytecode(w, p, input, bytecode.MatchContext{Hook: true, PDFPhase: phase, PDF: mc.PDF})
		if err != nil {
			return "", err
		}
		if name != "" {
			return name, nil
		}
		if cause != nil && pending == nil {
			pending = fmt.Errorf("%w: %w", ErrRuleIncomplete, cause)
		}
		if value == 0xcea5e && pending == nil {
			pending = fmt.Errorf("%w: %w", ErrRuleIncomplete, unknownBecause(UnknownBytecode))
		}
	}
	if mc.PDF != nil && (mc.PDF.MutatedFlags || mc.PDF.MutatedObjectFlags) && pending == nil {
		pending = fmt.Errorf("%w: %w", ErrRuleIncomplete, unknownBecause(UnknownBytecode))
	}
	return "", pending
}

// MatchPDFPostDump runs hooks after a nonempty, reference-extractable object
// was scanned without a fatal extraction error.
func (d *Database) MatchPDFPostDump(ctx context.Context, extracted []byte, mc MatchContext) (string, error) {
	if len(extracted) == 0 {
		return "", nil
	}
	return d.MatchPDFPhase(ctx, extracted, 2, mc)
}

// MatchPDFPre runs the reached PRE hook with the reference's pre-parse flag
// facts. A PDF parser calls this before discovering indirect objects.
func (d *Database) MatchPDFPre(ctx context.Context, data []byte, mc MatchContext) (string, error) {
	flags, known, present := pdfPreFlags(data)
	if !present {
		return "", nil
	}
	return d.MatchPDFPhase(ctx, data, 4, MatchContext{
		PDF:             &bytecode.PDFState{Flags: flags, FlagsKnown: known},
		BytecodeOptions: mc.BytecodeOptions, Metadata: mc.Metadata,
		MetadataPartial: mc.MetadataPartial, ScanExtracted: mc.ScanExtracted,
		ScanExtractedTyped: mc.ScanExtractedTyped,
	})
}

// pdfPreFlags returns the exact flags before object discovery when the
// reference cannot find an EOF marker. Header and trailer checks are the only
// flag writers before PRE in that path. A present EOF enters xref validation,
// so its flags remain unknown until that parser is modeled.
func pdfPreFlags(data []byte) (uint32, bool, bool) {
	window := data[:min(len(data), 1032)]
	start := bytes.Index(window, []byte("%PDF-"))
	if start < 0 || len(window)-start < 8 {
		return 0, false, false
	}
	var flags uint32
	version := window[start+5 : start+8]
	if version[0] != '1' || version[1] != '.' || version[2] < '1' || version[2] > '9' {
		flags |= 1 << 0 // BAD_PDF_VERSION
	}
	if start != 0 {
		flags |= 1 << 1 // BAD_PDF_HEADERPOS
	}
	tail := data[max(0, len(data)-2048):]
	if len(tail) < 5 || bytes.LastIndex(tail, []byte("%%EOF")) <= 0 {
		return flags | 1<<2, true, true // BAD_PDF_TRAILER
	}
	return flags, false, true
}

// The reference runs its PDF hook programs (kind 258) while it parses a PDF
// file, in four phases: before parsing (4), after parsing the objects (1),
// after each object it extracts (2), and at the end when flags are set (3). Each program
// asks for the phase with pdf_get_phase, and for the parser's state (flags,
// objects, offsets) with other functions. At
// functionality level 216 the official programs (BC.Pdf.Trojan.Agent-1350178
// and -1350184) compare the engine's dynamic configuration level with 55 and
// 56 and return 0 in every phase, as the reference's debug log shows; they
// change nothing.
//
// This engine runs every PDF hook program in reference phase order, on every PDF
// object it matches as a file. A run that returns 0 changes nothing; a
// PRE detection label detects; a run that needs unavailable parser state, fails, or
// returns the break value 0xcea5e (which ends the reference's PDF parsing at
// that point, a consequence not modeled here) leaves the object incomplete.
// The direct matcher cannot prove PARSED was reached, a POSTDUMP extraction
// happened, or END flags were set. It runs those phases for rule discovery,
// but their labels stay incomplete until the parser proves reachability.
func (d *Database) runPDFHooks(w *matchWork, data []byte) (string, error, error) {
	preFlags, preKnown, present := pdfPreFlags(data)
	if !present {
		return "", nil, nil
	}
	state := &bytecode.PDFState{Flags: preFlags, FlagsKnown: preKnown}
	for _, phase := range [...]uint32{4, 1, 2, 3} {
		if phase == 1 {
			// Object discovery and parsing can set additional flags.
			state.FlagsKnown = false
		}
		for _, p := range d.pdfHooks {
			name, value, cause, err := runBytecode(w, p, data, bytecode.MatchContext{
				Hook: true, PDFPhase: phase, PDF: state,
			})
			switch {
			case err != nil:
				return "", nil, err
			case name != "":
				if phase != 4 {
					return "", unknownBecause(UnknownBytecode), nil
				}
				return name, nil, nil
			case cause != nil:
				return "", cause, nil
			case value == 0xcea5e:
				return "", unknownBecause(UnknownBytecode), nil
			}
		}
	}
	if state.MutatedFlags || state.MutatedObjectFlags {
		return "", unknownBecause(UnknownBytecode), nil
	}
	return "", nil, nil
}
