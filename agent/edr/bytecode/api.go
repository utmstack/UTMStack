package bytecode

import (
	"bytes"
	"context"
	"math"
	"runtime"
	"strconv"

	"github.com/utmstack/UTMStack/agent/edr/x86dis"
)

// EngineLevel is the functionality level of the pinned ClamAV 1.5.4 release
// whose rule selection this engine targets. Programs read it through APIs 66
// and 67, which upstream answers with CL_FLEVEL and CL_FLEVEL_DCONF (equal
// in that release). Other packages should reference this constant.
const EngineLevel = 234

const (
	maxAPIOutput = 16 << 20
	maxVirusName = 256
	maxAPIDebug  = 64 << 10
	maxHashsets  = 4096
	maxHashKeys  = 1 << 18
	maxPipes     = 4096
	maxMaps      = 4096
	maxInflates  = 256
	maxBzip2s    = 256
	maxLZMAs     = 256
	maxLZMADict  = 64 << 20
	maxJSNorms   = 256
	maxMapSlots  = 1 << 18
	maxMapKeys   = 32 << 20
	apiFailure   = uint64(math.MaxUint32)
)

type apiDefinition struct {
	name   string
	result int
	args   []int
}

// Keys are one-based wire IDs, not indexes in the upstream C dispatch table.
// Reject unknown declarations even when their calls appear unreachable.
var builtinAPIs = map[int]apiDefinition{
	1:  {"test1", 32, []int{32, 32}},
	2:  {"read", 32, []int{65, 32}},
	3:  {"write", 32, []int{65, 32}},
	4:  {"seek", 32, []int{32, 32}},
	5:  {"setvirusname", 32, []int{65, 32}},
	8:  {"disasm_x86", 32, []int{95, 32}},
	9:  {"trace_directory", 32, []int{65, 32}},
	10: {"trace_scope", 32, []int{65, 32}},
	11: {"trace_source", 32, []int{65, 32}},
	12: {"trace_op", 32, []int{65, 32}},
	13: {"trace_value", 32, []int{65, 32}},
	14: {"trace_ptr", 32, []int{65, 32}},
	15: {"pe_rawaddr", 32, []int{32}},
	18: {"malloc", 65, []int{32}},
	20: {"get_pe_section", 32, []int{91, 32}},
	21: {"fill_buffer", 32, []int{65, 32, 32, 32, 32}},
	22: {"extract_new", 32, []int{32}},
	23: {"read_number", 32, []int{32}},
	24: {"hashset_new", 32, nil},
	25: {"hashset_add", 32, []int{32, 32}},
	26: {"hashset_remove", 32, []int{32, 32}},
	27: {"hashset_contains", 32, []int{32, 32}},
	28: {"hashset_done", 32, []int{32}},
	29: {"hashset_empty", 32, []int{32}},
	30: {"buffer_pipe_new", 32, []int{32}},
	31: {"buffer_pipe_new_fromfile", 32, []int{32}},
	32: {"buffer_pipe_read_avail", 32, []int{32}},
	33: {"buffer_pipe_read_get", 65, []int{32, 32}},
	34: {"buffer_pipe_read_stopped", 32, []int{32, 32}},
	35: {"buffer_pipe_write_avail", 32, []int{32}},
	36: {"buffer_pipe_write_get", 65, []int{32, 32}},
	37: {"buffer_pipe_write_stopped", 32, []int{32, 32}},
	38: {"buffer_pipe_done", 32, []int{32}},
	39: {"inflate_init", 32, []int{32, 32, 32}},
	40: {"inflate_process", 32, []int{32}},
	41: {"inflate_done", 32, []int{32}},
	57: {"map_new", 32, []int{32, 32}},
	58: {"map_addkey", 32, []int{65, 32, 32}},
	59: {"map_setvalue", 32, []int{65, 32, 32}},
	60: {"map_remove", 32, []int{65, 32, 32}},
	61: {"map_find", 32, []int{65, 32, 32}},
	62: {"map_getvaluesize", 32, []int{32}},
	63: {"map_getvalue", 65, []int{32, 32}},
	64: {"map_done", 32, []int{32}},
	51: {"memstr", 32, []int{65, 32, 65, 32}},
	52: {"hex2ui", 32, []int{32, 32}},
	56: {"entropy_buffer", 32, []int{65, 32}},
	72: {"get_environment", 32, []int{86, 32}},
	73: {"disable_bytecode_if", 32, []int{65, 32, 32}},
	74: {"disable_jit_if", 32, []int{65, 32, 32}},
	75: {"version_compare", 32, []int{65, 32, 65, 32}},
	76: {"check_platform", 32, []int{32, 32, 32}},
	6:  {"debug_print_str", 32, []int{65, 32}},
	7:  {"debug_print_uint", 32, []int{32}},
	16: {"file_find", 32, []int{65, 32}},
	17: {"file_byteat", 32, []int{32}},
	19: {"test2", 32, []int{32}},
	42: {"bytecode_rt_error", 32, []int{32}},
	43: {"jsnorm_init", 32, []int{32}},
	44: {"jsnorm_process", 32, []int{32}},
	45: {"jsnorm_done", 32, []int{32}},
	46: {"ilog2", 32, []int{32, 32}},
	47: {"ipow", 32, []int{32, 32, 32}},
	48: {"iexp", 32, []int{32, 32, 32}},
	49: {"isin", 32, []int{32, 32, 32}},
	50: {"icos", 32, []int{32, 32, 32}},
	54: {"debug_print_str_start", 32, []int{65, 32}},
	55: {"debug_print_str_nonl", 32, []int{65, 32}},
	65: {"file_find_limit", 32, []int{65, 32, 32}},
	66: {"engine_functionality_level", 32, nil},
	67: {"engine_dconf_level", 32, nil},
	68: {"engine_scan_options", 32, nil},
	69: {"engine_db_options", 32, nil},
	70: {"extract_set_container", 32, []int{32}},
	71: {"input_switch", 32, []int{32}},
	// PDF hook functions. The parser supplies proven object fields per phase;
	// uncertain fields stop the run with ErrPDFObjects.
	77:  {"pdf_get_obj_num", 32, nil},
	78:  {"pdf_get_flags", 32, nil},
	79:  {"pdf_set_flags", 32, []int{32}},
	80:  {"pdf_lookupobj", 32, []int{32}},
	81:  {"pdf_getobjsize", 32, []int{32}},
	82:  {"pdf_getobj", 65, []int{32, 32}},
	83:  {"pdf_getobjid", 32, []int{32}},
	84:  {"pdf_getobjflags", 32, []int{32}},
	85:  {"pdf_setobjflags", 32, []int{32, 32}},
	86:  {"pdf_get_offset", 32, []int{32}},
	87:  {"pdf_get_phase", 32, nil},
	88:  {"pdf_get_dumpedobjid", 32, nil},
	89:  {"matchicon", 32, []int{65, 32, 65, 32}},
	90:  {"running_on_jit", 32, nil},
	91:  {"get_file_reliability", 32, nil},
	92:  {"json_is_active", 32, nil},
	93:  {"json_get_object", 32, []int{65, 32, 32}},
	94:  {"json_get_type", 32, []int{32}},
	95:  {"json_get_array_length", 32, []int{32}},
	96:  {"json_get_array_idx", 32, []int{32, 32}},
	97:  {"json_get_string_length", 32, []int{32}},
	98:  {"json_get_string", 32, []int{65, 32, 32}},
	99:  {"json_get_boolean", 32, []int{32}},
	100: {"json_get_int", 32, []int{32}},
	101: {"engine_scan_options_ex", 32, []int{65, 32}},
	102: {"lzma_init", 32, []int{32, 32}},
	103: {"lzma_process", 32, []int{32}},
	104: {"lzma_done", 32, []int{32}},
	105: {"bzip2_init", 32, []int{32, 32}},
	106: {"bzip2_process", 32, []int{32}},
	107: {"bzip2_done", 32, []int{32}},
}

type apiState struct {
	kind           uint64
	disableStatus  uint32
	input          []byte
	originalInput  []byte
	extractedInput bool
	offset         int
	output         []byte
	outputOpen     bool
	virusName      string
	virusNameSet   bool
	debug          []string
	debugBytes     int
	debugTruncated bool
	// pe is the object's PE header information for a logical run on a
	// parsed PE object, nil otherwise.
	pe *PEInfo
	// scanning marks a run on a scanned object (logical or hook), where
	// written output is extracted content to scan.
	scanning bool
	// scanExtracted is supplied only by the enclosing recursive scanner.
	// Without it, a scanning run must stop before accepting output.
	scanExtracted func([]byte, int32) (string, error)
	// scanExtractedTyped also passes ClamAV's numeric container type. An
	// untyped callback cannot decide a non-default forced-type child.
	scanExtractedTyped func([]byte, int32, uint32) (string, error)
	containerType      uint32
	childName          string
	outputID           int32
	// pdfPhase is the PDF hook phase of the run, 0 outside PDF hooks.
	pdfPhase      uint32
	pdfFlagsKnown bool
	pdfFlags      uint32
	pdf           *PDFState
	// options describes the enclosing scanner run. A generic VM run has no
	// scanner configuration; option queries there must stay undecided.
	options *EngineOptions
	// fileReliability is nil when the enclosing scanner has not proved the
	// reference's corrupted_input value for this representation.
	fileReliability *uint32
	// jsonRoot is the scanner-owned root property tree. Child scans can add
	// properties while this invocation is suspended in an extraction callback.
	// Nil disables metadata collection.
	jsonRoot    map[string]any
	jsonObjs    []any
	jsonPartial bool
	// Hash sets are private to one VM invocation. Nil slots represent sets
	// destroyed before the last slot, as in the reference's ID table.
	hashsets     []map[uint32]struct{}
	hashKeys     int
	pipes        []bufferPipe
	filePtr      uint64
	savedFilePtr uint64
	fileSizePtr  uint64
	maps         []*bytecodeMap
	mapSlots     int
	mapKeys      int
	inflates     []*inflateState
	bzip2s       []*bzip2State
	bzip2Memory  uint64
	lzmas        []*lzmaState
	lzmaDict     uint64
	lzmaModels   uint64
	jsnorms      []*jsNormalizerState
	jsInput      int
	jsOutput     int
	jsFiles      [][]byte
	jsUnknown    bool
}

// flushExtracted hands one complete output file to the same recursive scanner
// as ordinary decoded children. The callback is synchronous: the VM owns and
// may reuse its buffer only after it returns. A positive result is the
// reference API's "previous extracted file was infected" return value.
func (s *apiState) flushExtracted(id int32) (uint64, error) {
	if !s.scanning || s.scanExtracted == nil && s.scanExtractedTyped == nil {
		return 0, ErrExtraction
	}
	if len(s.output) == 0 {
		s.outputOpen = false
		return 0, nil
	}
	name, err := s.scanChild(s.output, id, s.containerType)
	if name != "" && s.childName == "" {
		s.childName = name
	}
	// The reference recycles its temporary file between members. Clear the
	// bounded buffer even on a child error; the run will stop on that error.
	s.output = s.output[:0]
	s.outputOpen = false
	if err != nil {
		return 0, err
	}
	if name != "" {
		return 1, nil
	}
	return 0, nil
}

func (s *apiState) scanChild(data []byte, id int32, containerType uint32) (string, error) {
	if s.scanExtractedTyped != nil {
		return s.scanExtractedTyped(data, id, containerType)
	}
	if containerType == 0 && s.scanExtracted != nil {
		return s.scanExtracted(data, id)
	}
	return "", ErrExtraction
}

// invoke operates only on VM memory and explicit input/output buffers. A
// successful call is not a completed scan: callers must assess extracted
// output and dispatch scanner hooks before permitting bytecode activation.
func (s *apiState) invoke(ctx context.Context, id int, args []uint64, mem *memory, charge func(uint64) error) (uint64, error) {
	def, ok := builtinAPIs[id]
	if !ok {
		return 0, ErrUnsupported
	}
	if s == nil || len(args) != len(def.args) || ctx == nil || charge == nil {
		return 0, ErrRuntime
	}
	work := func(n uint64) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return charge(n)
	}
	if err := work(1); err != nil {
		return 0, err
	}
	// Bytes moved or scanned are charged like the interpreter's own bulk
	// memory operations: one step per 64 bytes, rounded up.
	bytesWork := func(n uint64) error { return work((n + 63) / 64) }
	if len(s.input) > math.MaxInt32 || s.offset < 0 {
		return 0, ErrRuntime
	}
	// The VM validates the full declared pointer span before the host API.
	// In particular, a negative signed length becomes an invalid huge span.
	n := 0
	switch {
	case id == 20:
		// The record pointer is checked for the whole record
		// (bytecode_vm.c apisize_override).
		if mem == nil || mem.referencePointer(args[0], sectionRecordSize) != nil {
			return 0, ErrRuntime
		}
	case id == 21:
		// The interpreter validates the complete writable buffer before the
		// host function checks its filled/cursor arguments.
		n = int(uint32(args[1]))
		if mem == nil || mem.referencePointer(args[0], uint32(args[1])) != nil || mem.check(args[0], n, true) != nil {
			return 0, ErrRuntime
		}
	case id == 8:
		// The decoded-instruction record is checked whole (apisize_override).
		if mem == nil || mem.referencePointer(args[0], disasmRecordSize) != nil {
			return 0, ErrRuntime
		}
	case id == 72:
		if mem == nil || mem.referencePointer(args[0], uint32(args[1])) != nil {
			return 0, ErrRuntime
		}
	case id >= 9 && id <= 14:
		// The reference's disabled trace profile still resolves its first
		// pointer argument using the second argument as the checked span.
		if mem == nil || mem.referencePointer(args[0], uint32(args[1])) != nil {
			return 0, ErrRuntime
		}
	case id == 51 || id == 75 || id == 89:
		// memstr passes two buffers; the interpreter checks both before
		// the call (bytecode_vm.c READP).
		for _, i := range []int{0, 2} {
			if mem == nil || mem.referencePointer(args[i], uint32(args[i+1])) != nil {
				return 0, ErrRuntime
			}
		}
	case len(def.args) > 0 && def.args[0] == 65:
		if mem == nil || uint64(uint32(args[1])) > uint64(math.MaxInt) {
			return 0, ErrRuntime
		}
		n = int(uint32(args[1]))
		if err := mem.check(args[0], n, id == 2 || id == 98); err != nil {
			return 0, err
		}
	}

	switch id {
	case 9, 10, 11, 12, 13, 14:
		// ClamAV's trace callbacks return immediately when trace_level is
		// zero, the normal scanner profile. The VM pointer check above
		// still happens before the host function is called.
		return 0, nil
	case 30, 31, 32, 33, 34, 35, 36, 37, 38:
		return s.pipeAPI(id, args, mem, bytesWork)
	case 57, 58, 59, 60, 61, 62, 63, 64:
		return s.mapAPI(id, args, mem, bytesWork)
	case 39, 40, 41:
		return s.inflateAPI(id, args, mem, bytesWork)
	case 105, 106, 107:
		return s.bzip2API(id, args, mem, bytesWork)
	case 102, 103, 104:
		return s.lzmaAPI(id, args, mem, bytesWork)
	case 43, 44, 45:
		return s.jsnormAPI(id, args, mem, bytesWork)
	case 75:
		return versionCompareAPI(args, mem, bytesWork)
	case 76:
		return checkPlatformAPI(args), nil
	case 90:
		return 0, nil // the Go VM has no JIT
	case 91:
		if !s.scanning {
			return 3, nil // ClamAV returns 3 without a scanning context.
		}
		if s.fileReliability == nil {
			return 0, ErrUnsupported
		}
		return uint64(*s.fileReliability), nil
	case 73, 74:
		if s.kind != 1 {
			return apiFailure, nil
		}
		if uint32(args[2]) != 0 {
			if id == 73 {
				s.disableStatus = 2
			} else if s.disableStatus != 2 {
				s.disableStatus = 1
			}
		}
		return uint64(s.disableStatus), nil
	case 1:
		if uint32(args[0]) == 0xf00dbeef && uint32(args[1]) == 0xbeeff00d {
			return 0x12345678, nil
		}
		return 0x55, nil
	case 2:
		if int32(args[1]) < 0 {
			return apiFailure, nil
		}
		if s.offset >= len(s.input) {
			return 0, nil
		}
		actual := min(n, len(s.input)-s.offset)
		if err := bytesWork(uint64(actual)); err != nil {
			return 0, err
		}
		if err := mem.write(args[0], s.input[s.offset:s.offset+actual]); err != nil {
			return 0, err
		}
		s.offset += actual
		return uint64(actual), nil
	case 3:
		if s.scanning && s.scanExtracted == nil && s.scanExtractedTyped == nil {
			return 0, ErrExtraction
		}
		if int32(args[1]) < 0 {
			return apiFailure, nil
		}
		if n > maxAPIOutput-len(s.output) {
			return 0, ErrLimit
		}
		if err := bytesWork(uint64(n)); err != nil {
			return 0, err
		}
		data, err := mem.read(args[0], n)
		if err != nil {
			return 0, err
		}
		s.output = append(s.output, data...)
		s.outputOpen = true
		return uint64(n), nil
	case 22:
		if s.extractedInput {
			// The reference can truncate an output file while its fmap is
			// active. Its resulting reads are platform-dependent; stop the
			// scan rather than accepting a possibly incomplete child.
			return 0, ErrUnsupported
		}
		result, err := s.flushExtracted(s.outputID)
		if err == nil {
			// The argument identifies the next file, not the one just
			// scanned (bytecode_api.h's extract_new contract).
			s.outputID = int32(args[0])
		}
		return result, err
	case 70:
		// CL_TYPE_IGNORED is the largest ClamAV 1.5.4 file type.
		// Earlier numeric gaps are accepted by the reference, but the child
		// scanner will keep an unrecognized forced type incomplete.
		if uint32(args[0]) > 587 {
			return apiFailure, nil
		}
		s.containerType = uint32(args[0])
		return 0, nil
	case 71:
		return s.inputSwitch(uint32(args[0]) != 0, mem, bytesWork)
	case 21:
		// ClamAV 1.5.4 ignores fill. If the file is already exhausted it
		// returns 0 before shifting the existing buffer. Otherwise it moves
		// [cursor:filled] to the front and reads into the remaining capacity.
		filled, cursor := uint32(args[2]), uint32(args[3])
		if n == 0 || filled > uint32(n) {
			return apiFailure, nil
		}
		if s.offset >= len(s.input) {
			return 0, nil
		}
		if cursor > filled {
			return apiFailure, nil
		}
		remaining := int(filled - cursor)
		available := min(n-remaining, len(s.input)-s.offset)
		if err := bytesWork(uint64(remaining + available)); err != nil {
			return 0, err
		}
		kept, err := mem.read(args[0]+uint64(cursor), remaining)
		if err != nil {
			return 0, err
		}
		if err := mem.write(args[0], kept); err != nil {
			return 0, err
		}
		if available == 0 {
			return 0, nil // the reference's read(0) returns 0
		}
		if err := mem.write(args[0]+uint64(remaining), s.input[s.offset:s.offset+available]); err != nil {
			return 0, err
		}
		s.offset += available
		return uint64(remaining + available), nil
	case 89:
		// ClamAV 1.5.4 checks the PE entry point and group lengths. For a
		// valid call it builds a zeroed cli_exe_info but never initializes
		// ndatadirs or dirs[2]; findres then exits before walking icons.
		// Thus API 89 cannot return a positive icon result in that version.
		// The VM already checked both declared pointer spans.
		if s.pe == nil || s.pe.EntryPoint == 0 || uint32(args[1]) > 127 || uint32(args[3]) > 127 {
			return apiFailure, nil
		}
		return 0, nil
	case 4:
		target := int64(int32(args[0]))
		switch uint32(args[1]) {
		case 0:
		case 1:
			target += int64(s.offset)
		case 2:
			target += int64(len(s.input))
		default:
			return apiFailure, nil
		}
		if target < 0 || target > int64(len(s.input)) {
			return apiFailure, nil
		}
		s.offset = int(target)
		return uint64(target), nil
	case 5:
		name, _, err := apiCString(mem, args[0], maxVirusName, maxVirusName, work)
		if err != nil {
			return 0, err
		}
		s.virusName = string(name)
		s.virusNameSet = true
		return 0, nil
	case 6:
		// Upstream measures the whole string even with debugging disabled, so
		// walk all of it (bounded by its region and the work budget) and keep
		// only what the remaining diagnostic capture can hold.
		message, complete, err := apiCString(mem, args[0], maxMemoryBytes, max(0, maxAPIDebug-s.debugBytes-1), work)
		if err != nil {
			return 0, err
		}
		if !complete {
			s.debugTruncated = true
			return 0, nil
		}
		s.addDebug(message)
		return 0, nil
	case 7:
		// Preserve the signed decimal presentation of upstream's %d, using
		// debug-disabled return values instead of platform stderr byte counts.
		message := strconv.FormatInt(int64(int32(args[0])), 10)
		if err := bytesWork(uint64(len(message))); err != nil {
			return 0, err
		}
		s.addDebug([]byte(message))
		return 0, nil
	case 16, 65:
		limit := len(s.input)
		if id == 65 {
			if int32(args[2]) <= 0 {
				return apiFailure, nil
			}
			limit = min(limit, int(int32(args[2])))
		}
		if n == 0 || n > 1024 || limit <= 0 {
			return apiFailure, nil
		}
		if err := bytesWork(uint64(n)); err != nil {
			return 0, err
		}
		needle, err := mem.read(args[0], n)
		if err != nil {
			return 0, err
		}
		// Upstream deliberately reads separate 4096-byte windows. It does
		// not overlap them or advance the input cursor (see REFERENCE.md).
		for offset := s.offset; offset < limit; {
			end := offset + min(4096, limit-offset)
			if end-offset < n {
				return apiFailure, nil
			}
			if err := bytesWork(uint64(end - offset)); err != nil {
				return 0, err
			}
			if index := bytes.Index(s.input[offset:end], needle); index >= 0 {
				return uint64(offset + index), nil
			}
			offset = end
		}
		return apiFailure, nil
	case 17:
		offset := uint64(uint32(args[0]))
		if offset >= uint64(len(s.input)) {
			return apiFailure, nil
		}
		return uint64(s.input[int(offset)]), nil
	case 19:
		if uint32(args[0]) == 0xf00d {
			return 0xd00f, nil
		}
		return 0x5555, nil
	case 42:
		location := int32(args[0])
		message := "Bytecode runtime error at line " + strconv.FormatUint(uint64(uint32(location>>8)), 10) + ", col " + strconv.FormatUint(uint64(uint32(location)&255), 10)
		if err := bytesWork(uint64(len(message))); err != nil {
			return 0, err
		}
		// Despite its name, the upstream API logs a location and returns zero.
		// It does not raise a VM error or terminate the calling program.
		s.addDebug([]byte(message))
		return 0, nil
	case 46:
		a, b := uint32(args[0]), uint32(args[1])
		if b == 0 {
			return math.MaxInt32, nil
		}
		return apiRoundedInteger((1<<26)*math.Log(float64(a)/float64(b))/math.Log(2), false)
	case 47, 48, 49, 50:
		a, b, c := int32(args[0]), int32(args[1]), int32(args[2])
		var value float64
		if id == 47 {
			if a == 0 && b < 0 {
				return math.MaxInt32, nil
			}
			value = float64(c) * math.Pow(float64(a), float64(b))
		} else {
			if b == 0 {
				return math.MaxInt32, nil
			}
			ratio := float64(a) / float64(b)
			switch id {
			case 48:
				value = float64(c) * math.Exp(ratio)
			case 49:
				value = float64(c) * math.Sin(ratio)
			case 50:
				value = float64(c) * math.Cos(ratio)
			}
		}
		return apiRoundedInteger(value, id == 48)
	case 54, 55:
		if n == 0 {
			return apiFailure, nil
		}
		if err := bytesWork(uint64(n)); err != nil {
			return 0, err
		}
		data, err := mem.read(args[0], n)
		if err != nil {
			return 0, err
		}
		if id == 54 {
			if end := bytes.IndexByte(data, 0); end >= 0 {
				data = data[:end]
			}
		}
		s.addDebug(data)
		return 0, nil
	case 18:
		size := uint32(args[0])
		if size == 0 || size > maxAllocation {
			return 0, nil // the reference returns NULL
		}
		if mem == nil {
			return 0, ErrRuntime
		}
		if err := bytesWork(uint64(size)); err != nil {
			return 0, err
		}
		ptr, err := mem.alloc(int(size), false)
		if err != nil {
			return 0, err
		}
		return ptr, mem.fill(ptr, 0, int(size)) // a fresh pool: zero bytes
	case 15:
		if at, ok := s.pe.rawAddr(uint32(args[0]), len(s.input)); ok {
			return uint64(at), nil
		}
		return apiFailure, nil // PE_INVALID_RVA
	case 20:
		if s.pe == nil || uint64(uint32(args[1])) >= uint64(len(s.pe.Sections)) {
			return apiFailure, nil
		}
		return 0, mem.write(args[0], s.pe.Sections[uint32(args[1])].record())
	case 8:
		return s.disasm(mem, args[0])
	case 23:
		return s.readNumber(uint32(args[0]), bytesWork)
	case 24:
		if len(s.hashsets) >= maxHashsets {
			return 0, ErrLimit
		}
		id := len(s.hashsets)
		s.hashsets = append(s.hashsets, make(map[uint32]struct{}))
		return uint64(id), nil
	case 25, 26, 27, 28, 29:
		setID := int(int32(args[0]))
		if setID < 0 || setID >= len(s.hashsets) {
			if id == 29 {
				return 1, nil // upstream hashset_empty treats an invalid ID as empty
			}
			return apiFailure, nil
		}
		set := s.hashsets[setID]
		if set == nil {
			// The reference forbids use after hashset_done. Its C hash table
			// may dereference freed storage here; do not turn this into clean.
			return 0, ErrRuntime
		}
		switch id {
		case 25:
			key := uint32(args[1])
			if _, exists := set[key]; !exists {
				if s.hashKeys >= maxHashKeys {
					return 0, ErrLimit
				}
				set[key] = struct{}{}
				s.hashKeys++
			}
			return 0, nil
		case 26:
			key := uint32(args[1])
			if _, exists := set[key]; !exists {
				return apiFailure, nil
			}
			delete(set, key)
			s.hashKeys--
			return 0, nil
		case 27:
			if _, exists := set[uint32(args[1])]; exists {
				return 1, nil
			}
			return 0, nil
		case 28:
			s.hashKeys -= len(set)
			s.hashsets[setID] = nil
			if setID == len(s.hashsets)-1 {
				s.hashsets = s.hashsets[:setID]
			}
			return 0, nil
		case 29:
			if len(set) == 0 {
				return 1, nil
			}
			return 0, nil
		}
	case 72:
		size := uint32(args[1])
		if size > environmentRecordSize {
			return apiFailure, nil
		}
		return 0, mem.write(args[0], environmentRecord()[:size])
	case 51:
		hs, ns := int(uint32(args[1])), int(uint32(args[3]))
		if err := bytesWork(uint64(hs) + uint64(ns)); err != nil {
			return 0, err
		}
		if hs == 0 || ns == 0 || hs < ns {
			return apiFailure, nil
		}
		if args[0] == args[2] {
			return 0, nil // the reference compares the pointers first
		}
		h, err := mem.read(args[0], hs)
		if err != nil {
			return 0, err
		}
		n, err := mem.read(args[2], ns)
		if err != nil {
			return 0, err
		}
		var i int
		if len(n) > 64 {
			i = linearIndex(h, n)
		} else {
			i = bytes.Index(h, n)
		}
		if i >= 0 {
			return uint64(i), nil
		}
		return apiFailure, nil
	case 52:
		hi, ok := hexNibble(byte(args[0]))
		if !ok {
			return apiFailure, nil
		}
		lo, ok := hexNibble(byte(args[1]))
		if !ok {
			return apiFailure, nil
		}
		// The reference stores the decoded byte in a signed plain char,
		// then widens it to int32 on its x86_64 build.
		return uint64(uint32(int32(int8(hi<<4 | lo)))), nil
	case 56:
		if int32(args[1]) <= 0 {
			return apiFailure, nil
		}
		if err := bytesWork(uint64(n)); err != nil {
			return 0, err
		}
		data, err := mem.read(args[0], n)
		if err != nil {
			return 0, err
		}
		var counts [256]uint32
		for _, b := range data {
			counts[b]++
		}
		var entropy float64
		log2 := math.Log(2)
		for _, count := range counts {
			if count != 0 {
				p := float64(count) / float64(n)
				entropy += -p * math.Log(p) / log2
			}
		}
		return uint64(uint32(entropy * (1 << 26))), nil
	case 66, 67:
		// Upstream ignores its context for both and returns the same level.
		return EngineLevel, nil
	case 68:
		if s.options == nil {
			return 0, ErrUnsupported
		}
		return uint64(s.options.Scan), nil
	case 69:
		if s.options == nil {
			return 0, ErrUnsupported
		}
		return uint64(s.options.Database), nil
	case 101:
		// ClamAV 1.5.4 writes every lowercased input byte to index zero,
		// leaving the rest of its malloc buffer uninitialized, and inverts
		// the category comparisons. The pinned interpreter returns zero for
		// the documented option names in the native differential fixtures.
		// Do not invent option values that would activate rules ClamAV leaves
		// inactive. The VM has already checked the declared pointer span.
		if n > 0 {
			if err := bytesWork(uint64(n)); err != nil {
				return 0, err
			}
			if _, err := mem.read(args[0], n); err != nil {
				return 0, err
			}
		}
		return 0, nil
	case 87:
		return uint64(s.pdfPhase), nil
	case 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 88:
		return s.invokePDF(id, args, mem, bytesWork)
	case 92, 93, 94, 95, 96, 97, 98, 99, 100:
		return s.invokeJSON(id, args, mem, n, bytesWork)
	}
	return 0, ErrUnsupported
}

func hexNibble(b byte) (byte, bool) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', true
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, true
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, true
	default:
		return 0, false
	}
}

// The upstream myround helper adds/subtracts 0.5 before a truncating integer
// cast. Preserve that order, including its floating-point rounding, but reject
// nonfinite/out-of-range casts instead of inheriting C's undefined behavior.
func apiRoundedInteger(value float64, unsigned bool) (uint64, error) {
	if value < 0 {
		value -= 0.5
	} else {
		value += 0.5
	}
	value = math.Trunc(value)
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, ErrRuntime
	}
	if unsigned {
		if value < 0 || value > math.MaxUint32 {
			return 0, ErrRuntime
		}
		return uint64(uint32(value)), nil
	}
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, ErrRuntime
	}
	return uint64(uint32(int32(value))), nil
}

// apiCString accommodates C APIs whose declared len omits the terminating NUL.
// Every byte is still checked against its virtual region and initialization.
// A string longer than limit fails with ErrLimit. At most keep leading bytes
// are copied; complete reports whether the whole string was.
func apiCString(mem *memory, ptr uint64, limit, keep int, work func(uint64) error) (data []byte, complete bool, err error) {
	for i := 0; i <= limit; i++ {
		if err := work(1); err != nil {
			return nil, false, err
		}
		position, err := pointerAdd(ptr, int64(i))
		if err != nil {
			return nil, false, err
		}
		b, err := mem.read(position, 1)
		if err != nil {
			return nil, false, err
		}
		if b[0] == 0 {
			return data, i <= keep, nil
		}
		if i == limit {
			return nil, false, ErrLimit
		}
		if i < keep {
			data = append(data, b[0])
		}
	}
	return nil, false, ErrLimit
}

// addDebug captures a diagnostic while it fits. Upstream discards debug output
// when debugging is disabled, so a full capture must never change a program's
// result: later messages are dropped instead, keeping the retained entries an
// exact prefix. One separator byte is charged even for empty messages, which
// bounds both the retained payload and the number of entries.
func (s *apiState) addDebug(message []byte) {
	if s.debugTruncated || len(message) >= maxAPIDebug-s.debugBytes {
		s.debugTruncated = true
		return
	}
	s.debug = append(s.debug, string(message))
	s.debugBytes += len(message) + 1
}

// maxAllocation is the reference's allocation limit (CLI_MAX_ALLOCATION).
const maxAllocation = 1 << 30

// readNumber is the reference's read_number: from the current offset it
// scans 16-byte blocks that must be fully readable, advancing the offset by
// 16 for each block without a digit. At the first digit (hex letters count in
// radix 16) it needs 16 readable bytes, parses like strtoul (radix 16 takes a
// 0x prefix) and moves the offset past the parsed text. The value is the
// host's unsigned long cut to 32 bits: Windows saturates at 32 bits, other
// hosts at 64 bits.
func (s *apiState) readNumber(radix uint32, work func(uint64) error) (uint64, error) {
	if radix != 10 && radix != 16 {
		return apiFailure, nil
	}
	const block = 16
	digit := func(c byte) (uint64, bool) {
		switch {
		case c >= '0' && c <= '9':
			return uint64(c - '0'), true
		case radix == 16 && c >= 'a' && c <= 'f':
			return uint64(c-'a') + 10, true
		case radix == 16 && c >= 'A' && c <= 'F':
			return uint64(c-'A') + 10, true
		}
		return 0, false
	}
	for s.offset+block <= len(s.input) {
		if err := work(block); err != nil {
			return 0, err
		}
		for i := 0; i < block; i++ {
			if _, ok := digit(s.input[s.offset+i]); !ok {
				continue
			}
			start := s.offset + i
			if start+block > len(s.input) {
				return apiFailure, nil
			}
			p := start
			if radix == 16 && s.input[p] == '0' && p+2 < len(s.input) && (s.input[p+1]|0x20) == 'x' {
				if _, ok := digit(s.input[p+2]); ok {
					p += 2
				}
			}
			limit := uint64(1<<64 - 1)
			if runtime.GOOS == "windows" {
				limit = 1<<32 - 1
			}
			var value uint64
			overflow := false
			for ; p < len(s.input); p++ {
				d, ok := digit(s.input[p])
				if !ok {
					break
				}
				if value > (limit-d)/uint64(radix) {
					overflow = true
				} else {
					value = value*uint64(radix) + d
				}
			}
			if overflow {
				value = limit
			}
			if err := work(uint64(p - start)); err != nil {
				return 0, err
			}
			s.offset = p
			return uint64(uint32(value)), nil
		}
		s.offset += block
	}
	return apiFailure, nil
}

// disasmRecordSize is the decoded-instruction record (struct DISASM_RESULT).
const disasmRecordSize = 64

// disasm is the reference's disasm_x86: it decodes the instruction at the
// current offset from at most 32 bytes, without moving the offset. At the end
// of the input nothing is written. The record's trailing 29 bytes are zeroed
// before decoding, so a refused instruction leaves only them changed; a
// decoded one writes the whole record and returns the offset after it.
func (s *apiState) disasm(mem *memory, ptr uint64) (uint64, error) {
	if s.offset >= len(s.input) {
		return apiFailure, nil
	}
	record, n, ok := x86dis.Decode(s.input[s.offset:min(len(s.input), s.offset+32)])
	if !ok {
		tail, err := pointerAdd(ptr, 35)
		if err != nil {
			return 0, err
		}
		return apiFailure, mem.write(tail, make([]byte, disasmRecordSize-35))
	}
	return uint64(s.offset + n), mem.write(ptr, record[:])
}

// linearIndex is bytes.Index in time linear in both lengths
// (Knuth-Morris-Pratt), for needles long enough that a program could make
// the standard search slow on repetitive data.
func linearIndex(haystack, needle []byte) int {
	if len(needle) == 0 {
		return 0
	}
	fail := make([]int, len(needle))
	for i, k := 1, 0; i < len(needle); i++ {
		for k > 0 && needle[i] != needle[k] {
			k = fail[k-1]
		}
		if needle[i] == needle[k] {
			k++
		}
		fail[i] = k
	}
	for i, k := 0, 0; i < len(haystack); i++ {
		for k > 0 && haystack[i] != needle[k] {
			k = fail[k-1]
		}
		if haystack[i] == needle[k] {
			k++
		}
		if k == len(needle) {
			return i - k + 1
		}
	}
	return -1
}
