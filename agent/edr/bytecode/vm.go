package bytecode

import (
	"bytes"
	"context"
	"encoding/binary"
	"math/bits"
)

type execution struct {
	ctx       context.Context
	program   *Program
	remaining uint64
	memory    memory
	globals   []uint64
	api       apiState
}

// Run executes a validated function without file input. A return value is a
// program result, never itself a malware verdict.
func (p *Program) Run(ctx context.Context, functionID int, args []uint64, maxSteps uint64) (uint64, error) {
	r, err := p.RunWithInput(ctx, functionID, args, nil, maxSteps)
	return r.Value, err
}

// Result is a snapshot of sandbox output, not an activated signature verdict.
// VirusName is an untrusted label; scanner hooks must validate and map it to an
// authorized signature before treating it as a detection.
type Result struct {
	Value  uint64
	Output []byte
	// BytecodeDisableStatus is the startup program's requested runtime mode:
	// zero keeps the current mode, one disables JIT, two disables all ClamBC.
	BytecodeDisableStatus uint32
	// ChildName is a detection returned by the trusted recursive scanner for
	// bytecode-produced output. It remains valid if later VM instructions fail.
	ChildName    string
	VirusName    string
	VirusNameSet bool
	Diagnostics  []string
	// WorkUsed is the steps charged, also when the run fails.
	WorkUsed uint64
	// DiagnosticsTruncated reports that later diagnostics were dropped after
	// the bounded capture filled. Capture never changes execution results.
	DiagnosticsTruncated bool
}

const MissingMatchOffset uint32 = 0xfffffffe

// MaxRunSteps is the most steps one run may be given. Official programs run
// for millions of instructions on large executables (the reference limits a
// run by time, not steps).
const MaxRunSteps = 50_000_000

// EngineOptions are the legacy ClamBC option masks returned by APIs 68 and
// 69. The caller supplies values for the actual scanner run; nil means the
// VM cannot answer those queries safely. API 101 uses the pinned release's
// separate name-query behavior (see api.go).
type EngineOptions struct {
	Scan, Database uint32
}

// MatchContext supplies actual logical subsignature results, in rule order.
// Empty slots have count zero and MissingMatchOffset, not zero or 0xffffffff.
type MatchContext struct {
	Counts  [64]uint32
	Offsets [64]uint32
	// PE is the object's header information when it is a parsed PE file.
	PE *PEInfo
	// Hook marks a hook program's context: the reference gives hook programs
	// all-zero counts and offsets (not the missing marker).
	Hook bool
	// PDFPhase is the phase a PDF hook program runs in (pdf_get_phase).
	PDFPhase uint32
	// PDFFlagsKnown marks an exact snapshot of the parser's flags for this
	// phase. PRE can supply it before any objects are discovered.
	PDFFlagsKnown bool
	PDFFlags      uint32
	// PDF carries the per-scan parser object table and any hook mutations.
	// Nil leaves object-dependent calls incomplete outside the empty PRE phase.
	PDF *PDFState
	// Options is the enclosing scanner's effective configuration.
	Options *EngineOptions
	// FileReliability is ClamAV's corrupted_input value when proved for this
	// representation. Nil leaves a reached API 91 query incomplete.
	FileReliability *uint32
	// Metadata is the enclosing scan's root JSON property tree, including when
	// this program scans a child or normalized view. Nil disables JSON queries.
	Metadata map[string]any
	// MetadataPartial makes an absent JSON object lookup undecided. Use it
	// when the scanner can supply only a verified subset of the reference tree.
	MetadataPartial bool
}

func EmptyMatchContext() MatchContext {
	var c MatchContext
	for i := range c.Offsets {
		c.Offsets[i] = MissingMatchOffset
	}
	return c
}

// RunWithInput executes against an immutable snapshot of bytes. The interpreter
// has no filesystem, network, process, native-pointer or JIT access. Instruction
// and bulk-byte budgets are shared by calls; memory is private to this execution.
func (p *Program) RunWithInput(ctx context.Context, functionID int, args []uint64, input []byte, maxSteps uint64) (Result, error) {
	return p.run(ctx, functionID, args, input, nil, nil, nil, maxSteps)
}

// RunWithMatchContext executes using real scanner match metadata. It does not
// evaluate the logical trigger; callers must do that before dispatching it.
func (p *Program) RunWithMatchContext(ctx context.Context, functionID int, args []uint64, input []byte, matches MatchContext, maxSteps uint64) (Result, error) {
	return p.run(ctx, functionID, args, input, &matches, nil, nil, maxSteps)
}

// RunWithMatchContextAndScanner lets the enclosing scanner analyze output
// files synchronously as recursive children. A nil callback has the same
// fail-closed behavior as RunWithMatchContext.
func (p *Program) RunWithMatchContextAndScanner(ctx context.Context, functionID int, args []uint64, input []byte, matches MatchContext, scan func([]byte, int32) (string, error), maxSteps uint64) (Result, error) {
	return p.run(ctx, functionID, args, input, &matches, scan, nil, maxSteps)
}

// RunWithMatchContextAndTypedScanner additionally preserves a reached
// extract_set_container request. The callback must treat any file type it
// cannot faithfully scan as incomplete, never as clean.
func (p *Program) RunWithMatchContextAndTypedScanner(ctx context.Context, functionID int, args []uint64, input []byte, matches MatchContext, scan func([]byte, int32, uint32) (string, error), maxSteps uint64) (Result, error) {
	return p.run(ctx, functionID, args, input, &matches, nil, scan, maxSteps)
}

func (p *Program) run(ctx context.Context, functionID int, args []uint64, input []byte, matches *MatchContext, scan func([]byte, int32) (string, error), typedScan func([]byte, int32, uint32) (string, error), maxSteps uint64) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if len(input) > 64<<20 {
		return Result{}, ErrLimit
	}
	if maxSteps == 0 {
		maxSteps = 1_000_000
	}
	if maxSteps > MaxRunSteps {
		return Result{}, ErrLimit
	}
	if p.NeedsMatchContext() && matches == nil {
		return Result{}, ErrUnsupported
	}
	if matches != nil {
		for i, count := range matches.Counts {
			if matches.Hook {
				if count != 0 || matches.Offsets[i] != 0 {
					return Result{}, ErrRuntime
				}
				continue
			}
			if count == 0 && matches.Offsets[i] != MissingMatchOffset || count > 0 && (uint64(count) > uint64(len(input))+1 || uint64(matches.Offsets[i]) > uint64(len(input))) {
				return Result{}, ErrRuntime
			}
		}
	}
	e := &execution{ctx: ctx, program: p, remaining: maxSteps}
	e.api.kind = p.Kind
	e.api.outputID = -1
	if err := e.chargeBytes(len(input)); err != nil {
		return Result{WorkUsed: maxSteps - e.remaining}, err
	}
	e.api.input = bytes.Clone(input)
	if matches != nil {
		e.api.pe = matches.PE
		e.api.scanning = true
		e.api.pdfPhase = matches.PDFPhase
		e.api.pdfFlagsKnown = matches.PDFFlagsKnown
		e.api.pdfFlags = matches.PDFFlags
		e.api.pdf = matches.PDF
		e.api.options = matches.Options
		e.api.fileReliability = matches.FileReliability
		e.api.jsonRoot = matches.Metadata
		e.api.jsonPartial = matches.MetadataPartial
		e.api.scanExtracted = scan
		e.api.scanExtractedTyped = typedScan
	}
	special := make(map[int]uint64)
	for _, g := range p.globals {
		for _, fix := range g.pointers {
			if fix.target < 0x8000 || special[fix.target] != 0 {
				continue
			}
			if fix.target == globalVirusNames {
				special[fix.target] = unregisteredRegion
				continue
			}
			size, ok := specialGlobalSize(fix.target)
			if !ok {
				return Result{WorkUsed: maxSteps - e.remaining}, ErrUnsupported
			}
			data := make([]byte, size)
			switch fix.target {
			case globalFileSize:
				binary.LittleEndian.PutUint32(data, uint32(len(input)))
			case globalKind:
				// The reference never sets the kind pointer, so every
				// program reads 0 (bytecode.c context_safe).
			case globalPEData:
				var pe *PEInfo
				if matches != nil {
					pe = matches.PE
				}
				data = peRecord(pe)
			case globalMatchCounts, globalMatchOffsets:
				if matches == nil {
					return Result{WorkUsed: maxSteps - e.remaining}, ErrUnsupported
				}
				values := matches.Counts
				if fix.target == globalMatchOffsets {
					values = matches.Offsets
				}
				for i, v := range values {
					binary.LittleEndian.PutUint32(data[i*4:], v)
				}
			}
			ptr, err := e.memory.alloc(size, false)
			if err != nil {
				return Result{WorkUsed: maxSteps - e.remaining}, err
			}
			if err = e.memory.write(ptr, data); err != nil {
				return Result{WorkUsed: maxSteps - e.remaining}, err
			}
			e.memory.protect(ptr)
			special[fix.target] = ptr
			if fix.target == globalFileSize {
				e.api.fileSizePtr = ptr
			}
		}
	}
	e.globals = make([]uint64, len(p.globals))
	for i, g := range p.globals {
		if i == 0 {
			continue
		}
		if err := e.chargeBytes(len(g.data)); err != nil {
			return Result{WorkUsed: maxSteps - e.remaining}, err
		}
		ptr, err := e.memory.alloc(len(g.data), false)
		if err != nil {
			return Result{WorkUsed: maxSteps - e.remaining}, err
		}
		e.globals[i] = ptr
		if g.unresolved {
			continue
		}
		if err = e.memory.write(ptr, g.data); err != nil {
			return Result{WorkUsed: maxSteps - e.remaining}, err
		}
	}
	for i, g := range p.globals {
		for _, fix := range g.pointers {
			target := special[fix.target]
			if fix.target < 0x8000 {
				target = e.globals[fix.target]
			}
			offset := int64(fix.targetOffset)
			if fix.target == globalVirusNames {
				offset = int64(uint32(fix.targetOffset)) // composed as a 32-bit offset
			}
			ptr, err := pointerAdd(target, offset)
			if err != nil {
				return Result{WorkUsed: maxSteps - e.remaining}, err
			}
			if err = e.memory.store(e.globals[i]+uint64(fix.offset), 8, ptr); err != nil {
				return Result{WorkUsed: maxSteps - e.remaining}, err
			}
		}
	}
	v, err := e.call(functionID, args, 0)
	if e.api.scanning && len(e.api.output) > 0 {
		// The reference flushes a pending extracted file after the VM stops,
		// including after a VM fault. A child may contain a detection even when
		// the instruction that followed write failed. Preserve the VM error
		// unless the flush is the first failure; the matcher separately gives
		// a trusted child detection precedence over either error.
		if _, flushErr := e.api.flushExtracted(e.api.outputID); flushErr != nil && err == nil {
			err = flushErr
		}
	}
	// The reference scans completed jsnorm temp files when bytecode execution
	// ends, including after an instruction failure. A child detection retains
	// precedence over the VM error in the enclosing matcher.
	if len(e.api.jsFiles) > 0 {
		if scanErr := e.api.scanJSFiles(); scanErr != nil && err == nil {
			err = scanErr
		}
	}
	if err != nil {
		return Result{WorkUsed: maxSteps - e.remaining, ChildName: e.api.childName}, err
	} // partial labels/output cannot masquerade as successful execution
	if err := ctx.Err(); err != nil {
		return Result{WorkUsed: maxSteps - e.remaining, ChildName: e.api.childName}, err
	}
	return Result{Value: v, Output: bytes.Clone(e.api.output), BytecodeDisableStatus: e.api.disableStatus, ChildName: e.api.childName, VirusName: e.api.virusName, VirusNameSet: e.api.virusNameSet, Diagnostics: append([]string(nil), e.api.debug...), WorkUsed: maxSteps - e.remaining, DiagnosticsTruncated: e.api.debugTruncated}, nil
}
func (e *execution) charge(n uint64) error {
	if err := e.ctx.Err(); err != nil {
		return err
	}
	if n > e.remaining {
		return ErrLimit
	}
	e.remaining -= n
	return nil
}
func (e *execution) chargeBytes(n int) error {
	if n < 0 {
		return ErrRuntime
	}
	return e.charge(1 + uint64(n)/64)
}
func bitMask(width int) uint64 {
	if width >= 64 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(width) - 1
}
func signed(v uint64, width int) int64 { return int64(v<<uint(64-width)) >> uint(64-width) }
func signedArithmetic(v uint64, width int) int64 {
	if width == 1 {
		return int64(v & 1)
	} // upstream interpreter promotes i1 arithmetic to signed byte 0/1
	return signed(v, width)
}
func (e *execution) call(id int, args []uint64, depth int) (uint64, error) {
	if err := e.ctx.Err(); err != nil {
		return 0, err
	}
	if depth >= 64 {
		return 0, ErrLimit
	}
	if id < 0 || id >= len(e.program.functions) {
		return 0, ErrRuntime
	}
	f := &e.program.functions[id]
	if len(args) != f.args {
		return 0, ErrRuntime
	}
	if err := e.chargeBytes(f.frameBytes); err != nil {
		return 0, err
	}
	frame, err := e.memory.alloc(f.frameBytes, false)
	if err != nil {
		return 0, err
	}
	defer e.memory.release(frame)
	writeValue := func(id int, v uint64) error {
		typeID := f.widths[id] & 0x7fff
		size, _ := e.program.types.size(typeID)
		return e.memory.store(frame+uint64(f.offsets[id]), size, v&bitMask(typeID))
	}
	for i, a := range args {
		if err := writeValue(i, a); err != nil {
			return 0, err
		}
	}
	bb := 0
	for {
		if bb < 0 || bb >= len(f.blocks) {
			return 0, ErrRuntime
		}
		for _, inst := range f.blocks[bb] {
			if err := e.charge(1); err != nil {
				return 0, err
			}
			var a [15]uint64
			for i, op := range inst.args {
				if inst.op == 34 && i == 1 {
					continue
				}
				if op.constant {
					a[i] = op.value & bitMask(op.width)
				} else {
					ptr := uint64(0)
					if op.global {
						if op.value == 0 {
							continue
						}
						ptr = e.globals[op.value]
					} else {
						ptr = frame + uint64(f.offsets[op.value])
					}
					if op.address {
						a[i] = ptr
					} else {
						size, _ := e.program.types.size(op.width)
						v, err := e.memory.load(ptr, size)
						if err != nil {
							return 0, err
						}
						a[i] = v & bitMask(op.width)
					}
				}
			}
			var result uint64
			boolResult := func(v bool) {
				if v {
					result = 1
				}
			}
			switch inst.op {
			case 1:
				result = a[0] + a[1]
			case 2:
				result = a[0] - a[1]
			case 3:
				result = a[0] * a[1]
			case 4:
				if a[1] == 0 {
					return 0, ErrRuntime
				}
				result = a[0] / a[1]
			case 5, 7:
				x, y := signedArithmetic(a[0], inst.width), signedArithmetic(a[1], inst.width)
				if y == 0 || inst.width > 1 && y == -1 && x == -1<<uint(inst.width-1) {
					return 0, ErrRuntime
				}
				if inst.op == 5 {
					result = uint64(x / y)
				} else {
					result = uint64(x % y)
				}
			case 6:
				if a[1] == 0 {
					return 0, ErrRuntime
				}
				result = a[0] % a[1]
			case 8, 9, 10:
				if a[1] >= uint64(inst.width) {
					return 0, ErrRuntime
				}
				if inst.op == 8 {
					result = a[0] << a[1]
				} else if inst.op == 9 {
					result = a[0] >> a[1]
				} else {
					result = uint64(signedArithmetic(a[0], inst.width) >> a[1])
				}
			case 11:
				result = a[0] & a[1]
			case 12:
				result = a[0] | a[1]
			case 13:
				result = a[0] ^ a[1]
			case 14, 16:
				result = a[0]
			case 15:
				result = uint64(signed(a[0], inst.args[0].width))
			case 17:
				if a[0] != 0 {
					bb = inst.branches[0]
				} else {
					bb = inst.branches[1]
				}
				continue
			case 18:
				bb = inst.branches[0]
				continue
			case 19:
				return a[0] & bitMask(inst.width), nil
			case 20:
				return 0, nil
			case 21:
				boolResult(a[0] == a[1])
			case 22:
				boolResult(a[0] != a[1])
			case 23:
				boolResult(a[0] > a[1])
			case 24:
				boolResult(a[0] >= a[1])
			case 25:
				boolResult(a[0] < a[1])
			case 26:
				boolResult(a[0] <= a[1])
			case 27:
				boolResult(signedArithmetic(a[0], inst.width) > signedArithmetic(a[1], inst.width))
			case 28:
				boolResult(signedArithmetic(a[0], inst.width) >= signedArithmetic(a[1], inst.width))
			case 29:
				boolResult(signedArithmetic(a[0], inst.width) <= signedArithmetic(a[1], inst.width))
			case 30:
				boolResult(signedArithmetic(a[0], inst.width) < signedArithmetic(a[1], inst.width))
			case 31:
				if a[0] != 0 {
					result = a[1]
				} else {
					result = a[2]
				}
			case 32:
				var err error
				result, err = e.call(inst.function, a[:len(inst.args)], depth+1)
				if err != nil {
					return 0, err
				}
				if inst.width == 0 {
					continue
				}
			case 33:
				var err error
				result, err = e.api.invoke(e.ctx, inst.function, a[:len(inst.args)], &e.memory, e.charge)
				if err != nil {
					return 0, err
				}
				if inst.width == 0 {
					continue
				}
			case 34:
				if err := writeValue(int(inst.args[1].value), a[0]); err != nil {
					return 0, err
				}
				continue
			case 35, 36:
				delta := int64(int32(a[1]))
				if inst.op == 35 {
					_, offset, err := e.program.types.element(inst.typeID, int(delta))
					if err != nil {
						return 0, err
					}
					delta = int64(offset)
				} else if inst.gepOffset >= 0 {
					delta = int64(inst.gepOffset)
				}
				var err error
				result, err = pointerAdd(a[0], delta)
				if err != nil {
					return 0, err
				}
			case 38:
				size, _ := e.program.types.size(inst.width)
				if err := e.memory.store(a[1], size, a[0]); err != nil {
					return 0, err
				}
				continue
			case 39:
				size, _ := e.program.types.size(inst.width)
				var err error
				result, err = e.memory.load(a[0], size)
				if err != nil {
					return 0, err
				}
			case 40, 41, 42, 43:
				if a[2] > maxMemoryBytes {
					return 0, ErrLimit
				}
				n := int(a[2])
				if err := e.chargeBytes(n); err != nil {
					return 0, err
				}
				if inst.op == 40 {
					if err := e.memory.fill(a[0], byte(a[1]), n); err != nil {
						return 0, err
					}
					continue
				}
				src, err := e.memory.read(a[1], n)
				if err != nil {
					return 0, err
				}
				if inst.op == 43 {
					dst, err := e.memory.read(a[0], n)
					if err != nil {
						return 0, err
					}
					result = uint64(uint32(int32(bytes.Compare(dst, src))))
				} else {
					if inst.op == 41 && n > 0 && a[0]>>32 == a[1]>>32 && int64(uint32(a[0])) < int64(uint32(a[1]))+int64(n) && int64(uint32(a[1])) < int64(uint32(a[0]))+int64(n) {
						return 0, ErrRuntime
					}
					if err := e.memory.write(a[0], src); err != nil {
						return 0, err
					}
					continue
				}
			case 44:
				result = 0
			case 45:
				return 0, ErrRuntime
			case 46:
				result = uint64(bits.ReverseBytes16(uint16(a[0])))
			case 47:
				result = uint64(bits.ReverseBytes32(uint32(a[0])))
			case 48:
				result = bits.ReverseBytes64(a[0])
			case 49:
				var err error
				result, err = pointerDifference(a[0], a[1])
				if err != nil {
					return 0, err
				}
			case 50:
				result = a[0]
			default:
				return 0, ErrUnsupported
			}
			if err := writeValue(inst.dest, result); err != nil {
				return 0, err
			}
		}
	}
}
