package bytecode

import (
	"bytes"
	"encoding/binary"
	"runtime"
)

// Scanner globals beyond the fixed-size arrays (clambc.h): the detection-name
// table, which the reference never registers (bytecode_vm.c registers only
// the five globals of cli_globals), and the PE header record __clambc_pedata.
const (
	globalVirusNames = 0x8002
	globalPEData     = 0x8003
	// peRecordSize is struct cli_pe_hook_data (pe.h): the reference sizes the
	// registered global from its declared field types, 648 bytes.
	peRecordSize = 648
)

// Section is one PE section as the reference's header reader records it
// (struct cli_exe_section): aligned address, virtual size, raw offset and raw
// size, the characteristics, and the four values as the header states them.
type Section struct {
	RVA, VSize, Raw, RSize, Characteristics, URVA, UVSize, URaw, URSize uint32
}

// PEInfo is what a program sees of a parsed PE object: the entry point as a
// file offset, the aligned header size and the sections. Record is the full
// header record a hook program sees (648 bytes); logical programs get none.
type PEInfo struct {
	EntryPoint, HeaderSize uint32
	Sections               []Section
	Record                 []byte
}

// peRecord is the header record a logical program reads through
// __clambc_pedata. For a logical run the reference fills only the header
// offset (0 for the scanned object), the entry point, the section count and
// the header size, and leaves every other byte zero (cli_bytecode_runlsig);
// an object that is not a parsed PE gets an all-zero record. A hook program
// gets the complete record cli_scanpe fills.
func peRecord(pe *PEInfo) []byte {
	if pe != nil && len(pe.Record) == peRecordSize {
		return bytes.Clone(pe.Record)
	}
	r := make([]byte, peRecordSize)
	if pe != nil {
		binary.LittleEndian.PutUint32(r[4:], pe.EntryPoint)
		binary.LittleEndian.PutUint16(r[8:], uint16(len(pe.Sections)))
		binary.LittleEndian.PutUint32(r[644:], pe.HeaderSize)
	}
	return r
}

// sectionRecordSize is struct cli_exe_section: nine uint32.
const sectionRecordSize = 36

// record is the section as struct cli_exe_section lays it out.
func (s Section) record() []byte {
	r := make([]byte, 0, sectionRecordSize)
	for _, v := range []uint32{s.RVA, s.VSize, s.Raw, s.RSize, s.Characteristics, s.URVA, s.UVSize, s.URaw, s.URSize} {
		r = binary.LittleEndian.AppendUint32(r, v)
	}
	return r
}

// rawAddr maps an address to a file offset as the reference's cli_rawaddr
// does for the bytecode API: below the header size an address is its own
// offset if inside the object; otherwise the last section with raw data
// containing it maps it. Without PE information nothing maps.
func (pe *PEInfo) rawAddr(rva uint32, size int) (uint32, bool) {
	var headers uint32
	var sections []Section
	if pe != nil {
		headers, sections = pe.HeaderSize, pe.Sections
	}
	if rva < headers {
		return rva, uint64(rva) < uint64(size)
	}
	for i := len(sections) - 1; i >= 0; i-- {
		s := sections[i]
		if s.RSize != 0 && s.RVA <= rva && s.RSize > rva-s.RVA {
			return rva - s.RVA + s.Raw, true
		}
	}
	return 0, false
}

// environmentRecordSize is struct cli_environment (bytecode_detect.h).
const environmentRecordSize = 492

// environmentRecord is what get_environment reports: the reference's
// cli_detect_environment on this host's operating system and architecture,
// for a 1.5.4 build without a JIT. On Darwin, the compiler fields match the
// pinned portable 1.5.4 reference. Linux/Windows compiler-version fields
// remain conservative zeros pending a native build-specific oracle.
func environmentRecord() []byte {
	r := make([]byte, environmentRecordSize)
	le := binary.LittleEndian
	var arch, sizeofPtr byte = 0, 8 // arch_unknown for hosts the reference does not name
	switch runtime.GOARCH {
	case "amd64":
		arch = 2 // arch_x86_64
	case "386":
		arch, sizeofPtr = 1, 4 // arch_i386
	}
	category, compiler, sysname := byte(17), byte(1), runtime.GOOS // os_generic, compiler_gnuc
	var compilerVersion uint32
	switch runtime.GOOS {
	case "windows":
		category, compiler, sysname = 16, 5, "Microsoft Windows" // os_win64, compiler_msc
	case "linux":
		category, sysname = 10, "Linux"
	case "darwin":
		category, compiler, sysname = 4, 3, "Darwin" // os_darwin, compiler_clang
		compilerVersion = 0x00040201                 // pinned portable 1.5.4 build
	}
	if host, release, version, machine := environmentUname(); host != "" {
		sysname = host
		copy(r[288:352], release)
		copy(r[353:417], version)
		copy(r[418:482], machine)
	}
	le.PutUint32(r[20:], EngineLevel)     // functionality_level
	le.PutUint32(r[24:], EngineLevel)     // dconf_level
	le.PutUint32(r[8:], compilerVersion)  // platform_id_c, no OS-feature bits in pinned macOS profile
	le.PutUint32(r[12:], compilerVersion) // c_version
	copy(r[28:92], "1.5.4")               // engine_version
	copy(r[223:287], sysname)             // sysname
	r[484], r[485], r[486], r[488] = sizeofPtr, arch, category, compiler
	le.PutUint32(r[0:], uint32(category)<<24|uint32(arch)<<20|uint32(compiler)<<16|EngineLevel<<8|EngineLevel) // platform_id_a
	le.PutUint32(r[4:], uint32(sizeofPtr)<<24)                                                                 // platform_id_b
	return r
}

// ReadsExecutableHeader reports whether the program can observe the object's
// executable header: it declares a pointer to the PE header record or the
// header functions pe_rawaddr (15) and get_pe_section (20). A declaration
// counts even if unused.
func (p *Program) ReadsExecutableHeader() bool {
	if p == nil {
		return false
	}
	if _, ok := p.apis[15]; ok {
		return true
	}
	if _, ok := p.apis[20]; ok {
		return true
	}
	for _, g := range p.globals {
		for _, ref := range g.pointers {
			if ref.target == globalPEData {
				return true
			}
		}
	}
	return false
}
