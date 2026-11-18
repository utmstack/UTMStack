package signatures

import (
	"errors"
	"fmt"
	"strings"
)

// DynamicConfig holds the per-module feature masks a feed's .cfg files select
// (bit layout of the reference engine). The compiler honors the load-time
// bits: PHISHING.ENGINE for .pdb/.wdb, the BYTECODE engine bits for .cbc,
// PE.MD5SECT for section hashes and PE.CERTS/PE.CATALOG for certificate
// databases. A directive that switches off any other default feature fails
// compilation, because the scanner cannot switch it off yet.
type DynamicConfig struct {
	PE, ELF, MachO, Archive, Document, Mail, Other, Phishing, Bytecode, Stats, PCRE uint32
}

const (
	dconfPEMD5Sect           = 0x10
	dconfPEUPX               = 0x20 // the built-in unpackers, tried before the PE unpacker hooks
	dconfPEFSG               = 0x40
	dconfPEPetite            = 0x100
	dconfPEPESpin            = 0x200
	dconfPEYC                = 0x400
	dconfPEWWPack            = 0x800
	dconfPENsPack            = 0x1000
	dconfPEMEW               = 0x2000
	dconfPEUpack             = 0x4000
	dconfPEASPack            = 0x8000
	dconfPECatalog           = 0x10000
	dconfPECerts             = 0x20000
	dconfPhishingEngine      = 0x1
	dconfPhishingEntconv     = 0x2 // the encoding-conversion HTML typing pass
	dconfBytecodeInterpreter = 0x1
	dconfBytecodeEngines     = 0xf
	dconfPCREGlobal          = 0x4
	dconfArchivePasswd       = 0x400000
)

// DefaultDynamicConfig is the configuration before any directive applies:
// every feature on except the PE SWIZZOR heuristic and the ARM bytecode JIT;
// the STATS "disabled" bits are off.
func DefaultDynamicConfig() DynamicConfig {
	return DynamicConfig{
		PE: 0xfffff &^ 0x80, ELF: 0x1, MachO: 0x1, Archive: 0x3fffffff, Document: 0x7ff,
		Mail: 0x3, Other: 0x7fff, Phishing: 0x3, Bytecode: 0x7, Stats: 0, PCRE: 0x7,
	}
}

type dconfModule struct {
	prefix string
	mask   func(*DynamicConfig) *uint32
	// switchable are default-on bits this engine honors when switched off.
	switchable uint32
}

var dconfModules = []dconfModule{
	{"PE:", func(c *DynamicConfig) *uint32 { return &c.PE }, dconfPEMD5Sect | dconfPECatalog | dconfPECerts},
	{"ELF:", func(c *DynamicConfig) *uint32 { return &c.ELF }, 0},
	{"MACHO:", func(c *DynamicConfig) *uint32 { return &c.MachO }, 0},
	{"ARCHIVE:", func(c *DynamicConfig) *uint32 { return &c.Archive }, dconfArchivePasswd},
	{"DOCUMENT:", func(c *DynamicConfig) *uint32 { return &c.Document }, 0},
	{"MAIL:", func(c *DynamicConfig) *uint32 { return &c.Mail }, 0},
	{"OTHER:", func(c *DynamicConfig) *uint32 { return &c.Other }, 0},
	{"PHISHING:", func(c *DynamicConfig) *uint32 { return &c.Phishing }, dconfPhishingEngine | dconfPhishingEntconv},
	// Bytecode either runs in this engine's interpreter or not at all; the
	// statistics switches do not affect detection.
	{"BYTECODE:", func(c *DynamicConfig) *uint32 { return &c.Bytecode }, ^uint32(0)},
	{"STATS:", func(c *DynamicConfig) *uint32 { return &c.Stats }, ^uint32(0)},
	{"PCRE:", func(c *DynamicConfig) *uint32 { return &c.PCRE }, 0},
}

// dconfOrigin records where a module's effective value came from.
type dconfOrigin struct {
	file string
	line int
}

type dconfState struct {
	config DynamicConfig
	origin [11]dconfOrigin
}

func newDconfState() *dconfState { return &dconfState{config: DefaultDynamicConfig()} }

// dconfApplies is the reference level check of a directive: the third
// ':'-separated field (runs of ':' count as one separator) is the minimum and
// the fourth the maximum. A missing minimum applies at every level; a field
// that does not start with a digit never applies.
func dconfApplies(line string, level int) bool {
	lo, ok := cStrtok(line, 2, ':')
	if !ok {
		return true
	}
	if lo[0] < '0' || lo[0] > '9' || uint32(cAtoi(lo)) > uint32(level) {
		return false
	}
	if hi, ok := cStrtok(line, 3, ':'); ok && (hi[0] < '0' || hi[0] > '9' || uint32(cAtoi(hi)) < uint32(level)) {
		return false
	}
	return true
}

// scanDconfValue reads the reference's "0x%x" conversion: a literal "0x", then
// optional whitespace, sign and a second "0x" prefix before at least one
// hexadecimal digit. Values beyond 32 bits are rejected.
func scanDconfValue(s string) (uint32, bool) {
	if !strings.HasPrefix(s, "0x") {
		return 0, false
	}
	s = strings.TrimLeft(s[2:], " \t\n\v\f\r")
	negative := false
	if s != "" && (s[0] == '+' || s[0] == '-') {
		negative, s = s[0] == '-', s[1:]
	}
	if len(s) > 2 && (s[:2] == "0x" || s[:2] == "0X") && hexNibble(s[2]) >= 0 {
		s = s[2:]
	}
	var v uint64
	digits := 0
	for ; digits < len(s) && hexNibble(s[digits]) >= 0; digits++ {
		if v = v<<4 | uint64(hexNibble(s[digits])); v > 0xffffffff {
			return 0, false
		}
	}
	if digits == 0 {
		return 0, false
	}
	if negative {
		v = -v
	}
	return uint32(v), true
}

// apply processes one chomped .cfg line: a directive for a known module that
// applies at level replaces that module's whole mask; the last one wins.
// Other lines, including comments, are ignored. It reports whether the line is
// a module directive and whether it applied.
func (s *dconfState) apply(file string, number int, line string, level int) (directive, applied bool, err error) {
	for i, m := range dconfModules {
		if !strings.HasPrefix(line, m.prefix) {
			continue
		}
		if !dconfApplies(line, level) {
			return true, false, nil
		}
		v, ok := scanDconfValue(line[len(m.prefix):])
		if !ok {
			return true, true, fmt.Errorf("malformed dynamic configuration value for %s", strings.TrimSuffix(m.prefix, ":"))
		}
		*m.mask(&s.config) = v
		s.origin[i] = dconfOrigin{file, number}
		return true, true, nil
	}
	return false, false, nil
}

// unsupported reports, per module, a final configuration this engine cannot
// honor: a default feature switched off that it cannot switch off, or bytecode
// left enabled without the interpreter.
func (s *dconfState) unsupported() []Diagnostic {
	defaults := DefaultDynamicConfig()
	var out []Diagnostic
	for i, m := range dconfModules {
		value := *m.mask(&s.config)
		off := *m.mask(&defaults) &^ value &^ m.switchable
		var err error
		switch {
		case off != 0:
			err = fmt.Errorf("dynamic configuration switches off %s features 0x%x that this engine cannot disable", strings.TrimSuffix(m.prefix, ":"), off)
		case m.prefix == "BYTECODE:" && value&dconfBytecodeEngines != 0 && value&dconfBytecodeInterpreter == 0:
			err = errors.New("dynamic configuration enables bytecode without the interpreter this engine uses")
		default:
			continue
		}
		out = append(out, Diagnostic{File: s.origin[i].file, Line: s.origin[i].line, Reason: err.Error()})
	}
	return out
}
