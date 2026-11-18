package x86dis

// The opcode maps below are written by hand in the processor manuals' operand
// notation. Each entry carries the operation number the reference decoder
// reports for that encoding; the numbers (and every quirk noted in a comment)
// were learned from the reference library's outputs on chosen inputs, never
// from its source.

// arg is an operand kind in opcode-map notation.
type arg uint8

const (
	__  arg = iota // no operand
	Eb             // ModRM r/m, byte
	Ew             // ModRM r/m, word
	Ev             // ModRM r/m, word or dword by operand size
	Ep             // ModRM r/m; memory is a far pointer (six bytes, a dword under 66)
	Gb             // ModRM reg, byte register
	Gw             // ModRM reg, word register
	Gv             // ModRM reg, word or dword register by operand size
	Rd             // ModRM r/m as a dword register whatever the mod field says
	Cd             // ModRM reg as a control register
	Dd             // ModRM reg as a debug register
	Sw             // ModRM reg as a segment register
	M              // memory only, word or dword by operand size
	Mb             // memory only, byte
	Ms             // memory only, six bytes (descriptor table register)
	Mq             // memory only, qword
	Ib             // byte immediate
	Iw             // word immediate
	Iz             // word or dword immediate by operand size
	Ap             // far pointer immediate
	Jb             // byte relative offset
	Jz             // word or dword relative offset by operand size
	Ob             // byte memory at an absolute offset (no ModRM)
	Ov             // word or dword memory at an absolute offset
	One            // the implicit count 1 of the D0-D3 shifts
	AL
	CL
	DX
	EAX // AX or EAX by operand size
	Zb  // byte register named by the opcode's low three bits
	Zv  // word or dword register named by the opcode's low three bits
	Zd  // dword register named by the opcode's low three bits
	ES
	CS
	SS
	DS
	FS
	GS
	ST0
	STi // x87 stack register named by the ModRM r/m field
)

// Size codes of memory, immediate and relative operands in the record.
const (
	sizeByte  = 0
	sizeWord  = 1
	sizeDword = 2
	sizeSix   = 3
	sizeQword = 4
	sizeTen   = 5
	sizePtr   = 6
)

// Register numbers in the record.
const (
	regEAX  = 0  // EAX..EDI are 0-7
	regAX   = 8  // AX..DI are 8-15
	regAH   = 16 // AH, CH, DH, BH are 16-19
	regAL   = 20 // AL, CL, DL, BL are 20-23
	regES   = 24 // ES, CS, SS, DS, FS, GS are 24-29
	regCR0  = 30 // CR0-CR7 are 30-37
	regDR0  = 38 // DR0-DR7 are 38-45
	regST0  = 46 // ST0-ST7 are 46-53
	regNone = 54
)

type kind uint8

const (
	invalid kind = iota
	plain
	group  // the ModRM reg field selects the form from groups[g]
	x87    // D8-DF: ModRM decides between memory and register tables
	escape // 0F: the next byte indexes the two-byte map
	prefix
)

type form struct {
	kind kind
	op   uint16 // operation number
	op16 uint16 // operation number under the operand-size prefix (0: same)
	g    uint8  // group index for kind group
	args [3]arg
}

func op(n uint16, a ...arg) form {
	f := form{kind: plain, op: n}
	copy(f.args[:], a)
	return f
}

// op2 is an instruction whose operation number changes under the operand-size prefix.
func op2(n, n16 uint16, a ...arg) form {
	f := op(n, a...)
	f.op16 = n16
	return f
}

func grp(g uint8) form { return form{kind: group, g: g} }

var (
	bad = form{}
	pfx = form{kind: prefix}
	fpu = form{kind: x87}
	esc = form{kind: escape}
)

// oneByte is the one-byte opcode map.
var oneByte = [256]form{
	0x00: op(5, Eb, Gb), 0x01: op(5, Ev, Gv), 0x02: op(5, Gb, Eb), 0x03: op(5, Gv, Ev), // add
	0x04: op(5, AL, Ib), 0x05: op(5, EAX, Iz),
	0x06: op(124, ES), 0x07: op(129, ES), // push, pop
	0x08: op(119, Eb, Gb), 0x09: op(119, Ev, Gv), 0x0a: op(119, Gb, Eb), 0x0b: op(119, Gv, Ev), // or
	0x0c: op(119, AL, Ib), 0x0d: op(119, EAX, Iz),
	0x0e: op(124, CS), 0x0f: esc,
	0x10: op(6, Eb, Gb), 0x11: op(6, Ev, Gv), 0x12: op(6, Gb, Eb), 0x13: op(6, Gv, Ev), // adc
	0x14: op(6, AL, Ib), 0x15: op(6, EAX, Iz),
	0x16: op(124, SS), 0x17: op(129, SS),
	0x18: op(147, Eb, Gb), 0x19: op(147, Ev, Gv), 0x1a: op(147, Gb, Eb), 0x1b: op(147, Gv, Ev), // sbb
	0x1c: op(147, AL, Ib), 0x1d: op(147, EAX, Iz),
	0x1e: op(124, DS), 0x1f: op(129, DS),
	0x20: op(7, Eb, Gb), 0x21: op(7, Ev, Gv), 0x22: op(7, Gb, Eb), 0x23: op(7, Gv, Ev), // and
	0x24: op(7, AL, Ib), 0x25: op(7, EAX, Iz),
	0x26: pfx, 0x27: op(50), // daa
	0x28: op(181, Eb, Gb), 0x29: op(181, Ev, Gv), 0x2a: op(181, Gb, Eb), 0x2b: op(181, Gv, Ev), // sub
	0x2c: op(181, AL, Ib), 0x2d: op(181, EAX, Iz),
	0x2e: pfx, 0x2f: op(51), // das
	0x30: op(195, Eb, Gb), 0x31: op(195, Ev, Gv), 0x32: op(195, Gb, Eb), 0x33: op(195, Gv, Ev), // xor
	0x34: op(195, AL, Ib), 0x35: op(195, EAX, Iz),
	0x36: pfx, 0x37: op(1), // aaa
	0x38: op(43, Eb, Gb), 0x39: op(43, Ev, Gv), 0x3a: op(43, Gb, Eb), 0x3b: op(43, Gv, Ev), // cmp
	0x3c: op(43, AL, Ib), 0x3d: op(43, EAX, Iz),
	0x3e: pfx, 0x3f: op(4), // aas

	0x40: op(59, Zv), 0x41: op(59, Zv), 0x42: op(59, Zv), 0x43: op(59, Zv), // inc
	0x44: op(59, Zv), 0x45: op(59, Zv), 0x46: op(59, Zv), 0x47: op(59, Zv),
	0x48: op(52, Zv), 0x49: op(52, Zv), 0x4a: op(52, Zv), 0x4b: op(52, Zv), // dec
	0x4c: op(52, Zv), 0x4d: op(52, Zv), 0x4e: op(52, Zv), 0x4f: op(52, Zv),
	0x50: op(124, Zv), 0x51: op(124, Zv), 0x52: op(124, Zv), 0x53: op(124, Zv), // push
	0x54: op(124, Zv), 0x55: op(124, Zv), 0x56: op(124, Zv), 0x57: op(124, Zv),
	0x58: op(129, Zv), 0x59: op(129, Zv), 0x5a: op(129, Zv), 0x5b: op(129, Zv), // pop
	0x5c: op(129, Zv), 0x5d: op(129, Zv), 0x5e: op(129, Zv), 0x5f: op(129, Zv),

	0x60: op2(125, 126), // pushad
	0x61: op2(130, 131), // popad
	0x62: op(9, Gv, M),  // bound
	0x63: op(8, Ew, Gw), // arpl
	0x64: pfx, 0x65: pfx, 0x66: pfx, 0x67: pfx,
	0x68: op(124, Iz),        // push
	0x69: op(58, Gv, Ev, Iz), // imul
	0x6a: op(124, Ib),
	0x6b: op(58, Gv, Ev, Ib),
	0x6c: op(63),        // insb
	0x6d: op2(61, 62),   // insd
	0x6e: op(123),       // outsb
	0x6f: op2(121, 122), // outsd

	0x70: op(70, Jb), 0x71: op(71, Jb), 0x72: op(72, Jb), 0x73: op(73, Jb), // jo..jae
	0x74: op(74, Jb), 0x75: op(75, Jb), 0x76: op(76, Jb), 0x77: op(77, Jb), // je..ja
	0x78: op(78, Jb), 0x79: op(79, Jb), 0x7a: op(80, Jb), 0x7b: op(81, Jb), // js..jnp
	0x7c: op(82, Jb), 0x7d: op(83, Jb), 0x7e: op(84, Jb), 0x7f: op(85, Jb), // jl..jg

	0x80: grp(gImmB), 0x81: grp(gImmZ), 0x82: grp(gImmB), 0x83: grp(gImmSB),
	0x84: op(186, Eb, Gb), 0x85: op(186, Ev, Gv), // test
	0x86: op(193, Eb, Gb), 0x87: op(193, Ev, Gv), // xchg
	0x88: op(109, Eb, Gb), 0x89: op(109, Ev, Gv), 0x8a: op(109, Gb, Eb), 0x8b: op(109, Gv, Ev), // mov
	0x8c: op(109, Ew, Sw),
	0x8d: op(93, Gv, M), // lea
	0x8e: op(109, Sw, Ew),
	0x8f: grp(gPop),

	0x90: op(117),                                                          // nop
	0x91: op(193, Zv, EAX), 0x92: op(193, Zv, EAX), 0x93: op(193, Zv, EAX), // xchg
	0x94: op(193, Zv, EAX), 0x95: op(193, Zv, EAX), 0x96: op(193, Zv, EAX), 0x97: op(193, Zv, EAX),
	0x98: op2(20, 21),   // cwde
	0x99: op2(18, 19),   // cdq
	0x9a: op(17, Ap),    // call far
	0x9b: op(55),        // wait
	0x9c: op2(127, 128), // pushfd
	0x9d: op2(131, 132), // popfd
	0x9e: op(145),       // sahf
	0x9f: op(87),        // lahf

	0xa0: op(109, AL, Ob), 0xa1: op(109, EAX, Ov), 0xa2: op(109, Ob, AL), 0xa3: op(109, Ov, EAX),
	0xa4: op(112),                                 // movsb
	0xa5: op2(110, 111),                           // movsd
	0xa6: op(46),                                  // cmpsb
	0xa7: op2(44, 45),                             // cmpsd
	0xa8: op(186, AL, Ib), 0xa9: op(186, EAX, Iz), // test
	0xaa: op(176),       // stosb
	0xab: op2(174, 175), // stosd
	0xac: op(101),       // lodsb
	0xad: op2(99, 100),  // lodsd
	0xae: op(150),       // scasb
	0xaf: op2(148, 149), // scasd

	0xb0: op(109, Zb, Ib), 0xb1: op(109, Zb, Ib), 0xb2: op(109, Zb, Ib), 0xb3: op(109, Zb, Ib), // mov
	0xb4: op(109, Zb, Ib), 0xb5: op(109, Zb, Ib), 0xb6: op(109, Zb, Ib), 0xb7: op(109, Zb, Ib),
	0xb8: op(109, Zv, Iz), 0xb9: op(109, Zv, Iz), 0xba: op(109, Zv, Iz), 0xbb: op(109, Zv, Iz),
	0xbc: op(109, Zv, Iz), 0xbd: op(109, Zv, Iz), 0xbe: op(109, Zv, Iz), 0xbf: op(109, Zv, Iz),

	0xc0: grp(gShiftIb), 0xc1: grp(gShiftVIb),
	0xc2: op(141, Iw), 0xc3: op(141), // ret
	0xc4: op(90, Gv, M), // les
	0xc5: op(89, Gv, M), // lds
	0xc6: grp(gMovB), 0xc7: grp(gMovZ),
	0xc8: op(54, Iw, Ib),             // enter
	0xc9: op(94),                     // leave
	0xca: op(140, Iw), 0xcb: op(140), // retf
	0xcc: op(65),     // int3
	0xcd: op(64, Ib), // int
	0xce: op(66),     // into
	0xcf: op(69),     // iret

	0xd0: grp(gShift1B), 0xd1: grp(gShift1V), 0xd2: grp(gShiftClB), 0xd3: grp(gShiftClV),
	0xd4: op(3, Ib), // aam
	0xd5: op(2, Ib), // aad
	0xd6: bad,
	0xd7: op(194), // xlat
	0xd8: fpu, 0xd9: fpu, 0xda: fpu, 0xdb: fpu, 0xdc: fpu, 0xdd: fpu, 0xde: fpu, 0xdf: fpu,

	0xe0: op(104, Jb),                           // loopne
	0xe1: op(103, Jb),                           // loope
	0xe2: op(102, Jb),                           // loop
	0xe3: op(105, Jb),                           // jecxz
	0xe4: op(60, AL, Ib), 0xe5: op(60, EAX, Ib), // in
	0xe6: op(120, Ib, AL), 0xe7: op(120, Ib, EAX), // out
	0xe8: op(17, Jz),                                     // call
	0xe9: op(86, Jz), 0xea: op(86, Ap), 0xeb: op(86, Jb), // jmp
	0xec: op(60, AL, DX), 0xed: op(60, EAX, DX),
	0xee: op(120, DX, AL), 0xef: op(120, DX, EAX),

	// lock, repne and rep come back as instructions of their own.
	0xf0: op(98), 0xf1: bad, 0xf2: op(139), 0xf3: op(138),
	0xf4: op(56), // hlt
	0xf5: op(26), // cmc
	0xf6: grp(gUnaryB), 0xf7: grp(gUnaryV),
	0xf8: op(22),  // clc
	0xf9: op(178), // stc
	0xfa: op(24),  // cli
	0xfb: op(180), // sti
	0xfc: op(23),  // cld
	0xfd: op(179), // std
	0xfe: grp(gIncB), 0xff: grp(gIncV),
}

// twoByte is the map of opcodes after 0F.
var twoByte = [256]form{
	0x00: grp(gSys0), 0x01: grp(gSys1),
	0x02: op(88, Gv, Ev),  // lar
	0x03: op(106, Gv, Ev), // lsl
	0x05: op(182),         // syscall
	0x06: op(25),          // clts
	0x07: op(185),         // sysret
	0x08: op(67),          // invd
	0x09: op(190),         // wbinvd
	0x0b: op(187),         // ud2

	0x20: op(109, Rd, Cd), 0x21: op(109, Rd, Dd), 0x22: op(109, Cd, Rd), 0x23: op(109, Dd, Rd), // mov

	0x30: op(191), // wrmsr
	0x31: op(137), // rdtsc
	0x32: op(135), // rdmsr
	0x33: op(136), // rdpmc
	0x34: op(183), // sysenter
	0x35: op(184), // sysexit

	0x40: op(27, Gv, Ev), 0x41: op(28, Gv, Ev), 0x42: op(29, Gv, Ev), 0x43: op(30, Gv, Ev), // cmovo..cmovae
	0x44: op(31, Gv, Ev), 0x45: op(32, Gv, Ev), 0x46: op(33, Gv, Ev), 0x47: op(34, Gv, Ev), // cmove..cmova
	0x48: op(35, Gv, Ev), 0x49: op(36, Gv, Ev), 0x4a: op(37, Gv, Ev), 0x4b: op(38, Gv, Ev), // cmovs..cmovnp
	0x4c: op(39, Gv, Ev), 0x4d: op(40, Gv, Ev), 0x4e: op(41, Gv, Ev), 0x4f: op(42, Gv, Ev), // cmovl..cmovg

	0x80: op(70, Jz), 0x81: op(71, Jz), 0x82: op(72, Jz), 0x83: op(73, Jz), // jo..jae
	0x84: op(74, Jz), 0x85: op(75, Jz), 0x86: op(76, Jz), 0x87: op(77, Jz), // je..ja
	0x88: op(78, Jz), 0x89: op(79, Jz), 0x8a: op(80, Jz), 0x8b: op(81, Jz), // js..jnp
	0x8c: op(82, Jz), 0x8d: op(83, Jz), 0x8e: op(84, Jz), 0x8f: op(85, Jz), // jl..jg

	0x90: op(151, Eb), 0x91: op(152, Eb), 0x92: op(153, Eb), 0x93: op(154, Eb), // seto..setae
	// setne reports the same operation number as setno.
	0x94: op(155, Eb), 0x95: op(152, Eb), 0x96: op(157, Eb), 0x97: op(158, Eb), // sete..seta
	0x98: op(159, Eb), 0x99: op(160, Eb), 0x9a: op(161, Eb), 0x9b: op(162, Eb), // sets..setnp
	0x9c: op(163, Eb), 0x9d: op(164, Eb), 0x9e: op(165, Eb), 0x9f: op(166, Eb), // setl..setg

	0xa0: op(124, FS), 0xa1: op(129, FS), // push, pop
	0xa2: op(49),                                         // cpuid
	0xa3: op(13, Ev, Gv),                                 // bt
	0xa4: op(170, Ev, Gv, Ib), 0xa5: op(170, Ev, Gv, CL), // shld
	0xa8: op(124, GS), 0xa9: op(129, GS),
	0xaa: op(144),                                        // rsm
	0xab: op(16, Ev, Gv),                                 // bts
	0xac: op(172, Ev, Gv, Ib), 0xad: op(172, Ev, Gv, CL), // shrd
	0xaf: op(58, Gv, Ev), // imul

	0xb0: op(47, Eb, Gb), 0xb1: op(47, Ev, Gv), // cmpxchg
	0xb2: op(107, Gv, M),                         // lss
	0xb3: op(15, Ev, Gv),                         // btr
	0xb4: op(91, Gv, M),                          // lfs
	0xb5: op(92, Gv, M),                          // lgs
	0xb6: op(114, Gv, Eb), 0xb7: op(114, Gv, Ew), // movzx
	0xba: grp(gBit),
	0xbb: op(14, Ev, Gv),                         // btc
	0xbc: op(10, Gv, Ev),                         // bsf
	0xbd: op(11, Gv, Ev),                         // bsr
	0xbe: op(113, Gv, Eb), 0xbf: op(113, Gv, Ew), // movsx

	0xc0: op(192, Eb, Gb), 0xc1: op(192, Ev, Gv), // xadd
	0xc7: grp(gCmpxchg8b),
	0xc8: op(12, Zd), 0xc9: op(12, Zd), 0xca: op(12, Zd), 0xcb: op(12, Zd), // bswap
	0xcc: op(12, Zd), 0xcd: op(12, Zd), 0xce: op(12, Zd), 0xcf: op(12, Zd),
}

// ModRM groups, indexed by the reg field.
const (
	gImmB = iota
	gImmZ
	gImmSB
	gPop
	gShiftIb
	gShiftVIb
	gShift1B
	gShift1V
	gShiftClB
	gShiftClV
	gMovB
	gMovZ
	gUnaryB
	gUnaryV
	gIncB
	gIncV
	gSys0
	gSys1
	gBit
	gCmpxchg8b
)

var groups = [...][8]form{
	// add, or, adc, sbb, and, sub, xor, cmp
	gImmB:  {op(5, Eb, Ib), op(119, Eb, Ib), op(6, Eb, Ib), op(147, Eb, Ib), op(7, Eb, Ib), op(181, Eb, Ib), op(195, Eb, Ib), op(43, Eb, Ib)},
	gImmZ:  {op(5, Ev, Iz), op(119, Ev, Iz), op(6, Ev, Iz), op(147, Ev, Iz), op(7, Ev, Iz), op(181, Ev, Iz), op(195, Ev, Iz), op(43, Ev, Iz)},
	gImmSB: {op(5, Ev, Ib), op(119, Ev, Ib), op(6, Ev, Ib), op(147, Ev, Ib), op(7, Ev, Ib), op(181, Ev, Ib), op(195, Ev, Ib), op(43, Ev, Ib)},
	gPop:   {op(129, Ev)},
	// rol, ror, rcl, rcr, shl, shr, (invalid), sar
	gShiftIb:  {op(142, Eb, Ib), op(143, Eb, Ib), op(133, Eb, Ib), op(134, Eb, Ib), op(169, Eb, Ib), op(171, Eb, Ib), bad, op(146, Eb, Ib)},
	gShiftVIb: {op(142, Ev, Ib), op(143, Ev, Ib), op(133, Ev, Ib), op(134, Ev, Ib), op(169, Ev, Ib), op(171, Ev, Ib), bad, op(146, Ev, Ib)},
	gShift1B:  {op(142, Eb, One), op(143, Eb, One), op(133, Eb, One), op(134, Eb, One), op(169, Eb, One), op(171, Eb, One), bad, op(146, Eb, One)},
	gShift1V:  {op(142, Ev, One), op(143, Ev, One), op(133, Ev, One), op(134, Ev, One), op(169, Ev, One), op(171, Ev, One), bad, op(146, Ev, One)},
	gShiftClB: {op(142, Eb, CL), op(143, Eb, CL), op(133, Eb, CL), op(134, Eb, CL), op(169, Eb, CL), op(171, Eb, CL), bad, op(146, Eb, CL)},
	gShiftClV: {op(142, Ev, CL), op(143, Ev, CL), op(133, Ev, CL), op(134, Ev, CL), op(169, Ev, CL), op(171, Ev, CL), bad, op(146, Ev, CL)},
	gMovB:     {op(109, Eb, Ib)},
	gMovZ:     {op(109, Ev, Iz)},
	// test, (invalid), not, neg, mul, imul, div, idiv
	gUnaryB: {op(186, Eb, Ib), bad, op(118, Eb), op(116, Eb), op(115, Eb), op(58, Eb), op(53, Eb), op(57, Eb)},
	gUnaryV: {op(186, Ev, Iz), bad, op(118, Ev), op(116, Ev), op(115, Ev), op(58, Ev), op(53, Ev), op(57, Ev)},
	// inc, dec
	gIncB: {op(59, Eb), op(52, Eb)},
	// inc, dec, call, call far, jmp, jmp far, push
	gIncV: {op(59, Ev), op(52, Ev), op(17, Ev), op(17, Ep), op(86, Ev), op(86, Ep), op(124, Ev)},
	// sldt, str, lldt, ltr, verr, verw
	gSys0: {op(173, Ew), op(177, Ew), op(97, Ew), op(108, Ew), op(188, Ew), op(189, Ew)},
	// sgdt, sidt, lgdt, lidt, (smsw, invalid, lmsw: not decoded), invlpg
	gSys1: {op(167, Ms), op(168, Ms), op(95, Ms), op(96, Ms), bad, bad, bad, op(68, Mb)},
	// bt, bts, btr, btc
	gBit:       {4: op(13, Ev, Ib), 5: op(16, Ev, Ib), 6: op(15, Ev, Ib), 7: op(14, Ev, Ib)},
	gCmpxchg8b: {1: op(48, Mq)},
}

// x87 forms: memory operands by the ModRM reg field, register forms by the
// whole ModRM byte (row = reg field, column = r/m field).
type fpuMem struct {
	op   uint16 // 0: invalid
	size uint8
}

type fpuReg struct {
	op   uint16 // 0: invalid
	args [2]arg
}

func fpuRow(n uint16, a ...arg) [8]fpuReg {
	var row [8]fpuReg
	for i := range row {
		row[i].op = n
		copy(row[i].args[:], a)
	}
	return row
}

// fpuOps is a row whose instructions have no operands, one per r/m value.
func fpuOps(ns ...uint16) [8]fpuReg {
	var row [8]fpuReg
	for i, n := range ns {
		row[i].op = n
	}
	return row
}

var fpuMemory = [8][8]fpuMem{
	// fadd, fmul, fcom, fcomp, fsub, fsubr, fdiv, fdivr (m32real)
	{{203, sizeDword}, {253, sizeDword}, {217, sizeDword}, {220, sizeDword}, {272, sizeDword}, {274, sizeDword}, {224, sizeDword}, {226, sizeDword}},
	// fld, -, fst, fstp (m32real), fldenv, fldcw, fnstenv, fnstcw
	{{243, sizeDword}, {}, {267, sizeDword}, {270, sizeDword}, {246, sizePtr}, {245, sizeWord}, {269, sizePtr}, {268, sizeWord}},
	// fiadd, fimul, ficom, ficomp, fisub, fisubr, fidiv, fidivr (m32int)
	{{229, sizeDword}, {235, sizeDword}, {230, sizeDword}, {231, sizeDword}, {241, sizeDword}, {242, sizeDword}, {232, sizeDword}, {233, sizeDword}},
	// fild, fisttp, fist, fistp (m32int), -, fld (m80), -, fstp (m80)
	{{234, sizeDword}, {240, sizeDword}, {238, sizeDword}, {239, sizeDword}, {}, {243, sizeTen}, {}, {270, sizeTen}},
	// the D8 operations on m64real
	{{203, sizeQword}, {253, sizeQword}, {217, sizeQword}, {220, sizeQword}, {272, sizeQword}, {274, sizeQword}, {224, sizeQword}, {226, sizeQword}},
	// fld, fisttp, fst, fstp (m64), frstor, -, fnsave, fnstsw
	{{243, sizeQword}, {240, sizeQword}, {267, sizeQword}, {270, sizeQword}, {261, sizePtr}, {}, {266, sizePtr}, {271, sizeWord}},
	// the DA operations on m16int
	{{229, sizeWord}, {235, sizeWord}, {230, sizeWord}, {231, sizeWord}, {241, sizeWord}, {242, sizeWord}, {232, sizeWord}, {233, sizeWord}},
	// fild, fisttp, fist, fistp (m16int), fbld, fild (m64int), fbstp, fistp (m64int)
	{{234, sizeWord}, {240, sizeWord}, {238, sizeWord}, {239, sizeWord}, {205, sizeTen}, {234, sizeQword}, {206, sizeTen}, {239, sizeQword}},
}

var fpuRegister = [8][8][8]fpuReg{
	// D8: fadd, fmul, fcom, fcomp, fsub, fsubr, fdiv, fdivr st0, st(i)
	{fpuRow(203, ST0, STi), fpuRow(253, ST0, STi), fpuRow(217, ST0, STi), fpuRow(220, ST0, STi),
		fpuRow(272, ST0, STi), fpuRow(274, ST0, STi), fpuRow(224, ST0, STi), fpuRow(226, ST0, STi)},
	// D9: fld, fxch; fnop; -; fchs fabs - - ftst fxam; fld1 fldl2t fldl2e fldpi
	// fldlg2 fldln2 fldz; f2xm1 fyl2x fptan fpatan fxtract fprem1 fdecstp
	// fincstp; fprem fyl2xp1 fsqrt fsincos frndint fscale fsin fcos.
	// fldz reports the operands st0, st(6).
	{fpuRow(243, ST0, STi), fpuRow(283, ST0, STi), fpuOps(255), {},
		fpuOps(207, 202, 0, 0, 276, 282),
		{{op: 244}, {op: 248}, {op: 247}, {op: 251}, {op: 249}, {op: 250}, {op: 252, args: [2]arg{ST0, STi}}, {}},
		fpuOps(201, 285, 259, 256, 284, 258, 223, 236),
		fpuOps(257, 286, 265, 264, 260, 262, 263, 222)},
	// DA: fcmovb, fcmove, fcmovbe, fcmovu; -; fucompp
	{fpuRow(209, ST0, STi), fpuRow(211, ST0, STi), fpuRow(210, ST0, STi), fpuRow(216, ST0, STi),
		{}, fpuOps(0, 281), {}, {}},
	// DB: fcmovnb, fcmovne, fcmovnbe, fcmovnu; fnclex fninit; fucomi, fcomi
	{fpuRow(212, ST0, STi), fpuRow(214, ST0, STi), fpuRow(213, ST0, STi), fpuRow(215, ST0, STi),
		fpuOps(0, 0, 208, 237), fpuRow(278, ST0, STi), fpuRow(218, ST0, STi), {}},
	// DC: fadd, fmul st(i), st0; -, -; fsubr, fsub, fdivr, fdiv st(i), st0
	{fpuRow(203, STi, ST0), fpuRow(253, STi, ST0), {}, {},
		fpuRow(274, STi, ST0), fpuRow(272, STi, ST0), fpuRow(226, STi, ST0), fpuRow(224, STi, ST0)},
	// DD: ffree; -; fst, fstp st(i); fucom, fucomp st(i), st0
	{fpuRow(228, STi), {}, fpuRow(267, STi), fpuRow(270, STi),
		fpuRow(277, STi, ST0), fpuRow(280, STi, ST0), {}, {}},
	// DE: faddp, fmulp; -; fcompp; fsubrp, fsubp, fdivrp, fdivp st(i), st0
	{fpuRow(204, STi, ST0), fpuRow(254, STi, ST0), {}, fpuOps(0, 221),
		fpuRow(275, STi, ST0), fpuRow(273, STi, ST0), fpuRow(227, STi, ST0), fpuRow(225, STi, ST0)},
	// DF: DF E0 (fnstsw ax) is refused and DF E1 reports fnstsw's operation;
	// fucomip, fcomip st0, st(i)
	{{}, {}, {}, {}, fpuOps(0, 271), fpuRow(279, ST0, STi), fpuRow(219, ST0, STi), {}},
}
