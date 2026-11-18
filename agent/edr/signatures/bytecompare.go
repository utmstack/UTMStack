package signatures

import (
	"errors"
	"strconv"
	"strings"
)

// byteCompare is the pinned ClamAV 1.5.4 byte-comparison profile of a logical
// subsignature. The matcher interprets binary integers and ASCII hex/decimal
// numbers, including automatic hex/decimal selection.
type byteCompare struct {
	shift   int64
	width   int
	little  bool
	format  byte
	exact   bool
	noOrder bool
	checks  [2]byteComparison
	ncheck  int
}

type byteComparison struct {
	op    byte
	value int64
}

func parseBinaryByteCompare(sub string, ref int) (*byteCompare, error) {
	open := strings.IndexByte(sub, '(')
	if open < 0 || !strings.HasSuffix(sub, ")") || sub[:open] != strconv.Itoa(ref) {
		return nil, errors.New("invalid byte comparison reference")
	}
	fields := strings.Split(sub[open+1:len(sub)-1], "#")
	if len(fields) != 3 {
		return nil, errors.New("invalid byte comparison fields")
	}
	b := &byteCompare{}
	shiftText := fields[0]
	if shiftText == "" || shiftText == "0" {
		shiftText = ""
	} else if strings.HasPrefix(shiftText, ">>") || strings.HasPrefix(shiftText, "<<") {
		shiftText = shiftText[2:]
	} else {
		return nil, errors.New("unsupported byte comparison offset")
	}
	shift := int64(0)
	if shiftText != "" {
		var err error
		shift, err = strconv.ParseInt(shiftText, 0, 32)
		if err != nil {
			return nil, errors.New("unsupported byte comparison offset")
		}
	}
	if strings.HasPrefix(fields[0], "<<") {
		shift = -shift
	}
	b.shift = shift
	options := fields[1]
	i := 0
	var endian byte
	for i < len(options) && (options[i] < '0' || options[i] > '9') {
		switch options[i] {
		case 'h', 'd', 'i', 'a':
			if b.format != 0 && b.format != options[i] {
				return nil, errors.New("unsupported byte comparison encoding")
			}
			b.format = options[i]
		case 'l', 'b':
			if endian != 0 && endian != options[i] {
				return nil, errors.New("unsupported byte comparison byte order")
			}
			endian = options[i]
		case 'e':
			b.exact = true
		default:
			return nil, errors.New("unsupported byte comparison option")
		}
		i++
	}
	if b.format == 0 {
		return nil, errors.New("unsupported byte comparison encoding")
	}
	if b.format == 'd' {
		if endian == 'l' {
			return nil, errors.New("unsupported decimal byte order")
		}
		endian = 'b'
	}
	b.noOrder = endian == 0
	b.little = endian == 'l'
	width, err := strconv.ParseUint(options[i:], 0, 32)
	if err != nil || width == 0 || b.format == 'i' && width != 1 && width != 2 && width != 4 && width != 8 || b.format == 'h' && width > 18 {
		return nil, errors.New("unsupported byte comparison width")
	}
	b.width = int(width)
	parts := strings.Split(fields[2], ",")
	if len(parts) < 1 || len(parts) > 2 {
		return nil, errors.New("unsupported byte comparison count")
	}
	for i, part := range parts {
		if len(part) < 2 || !strings.ContainsRune("<>=", rune(part[0])) {
			return nil, errors.New("invalid byte comparison operator")
		}
		value, err := strconv.ParseInt(part[1:], 0, 64)
		if err != nil {
			return nil, errors.New("invalid byte comparison value")
		}
		b.checks[i] = byteComparison{op: part[0], value: value}
	}
	b.ncheck = len(parts)
	return b, nil
}

func (m *patternCounts) byteCompare(r *rule, h *placeholder) (int, string, error) {
	ref := h.ref
	if ref < 0 || ref >= len(m.c) {
		return countUnsupported, UnknownByteCompare, nil
	}
	switch n := m.c[ref]; {
	case n == 0:
		return 0, "", nil
	case n == countNeedsHeader || n == countStateLimited:
		return n, "", nil
	case n == countUnsupported:
		return n, m.why[ref], nil
	case n < 0:
		return countUnsupported, UnknownByteCompare, nil
	}
	pc := m.counts[r.patterns[ref]]
	if pc == nil || !pc.done || pc.n == 0 {
		return countUnsupported, UnknownByteCompare, nil
	}
	b := h.bcomp
	if err := m.w.step(int64(b.width) + int64(b.ncheck)); err != nil {
		return 0, "", err
	}
	at := int64(pc.last) + b.shift
	// The pinned 1.5.4 matcher rejects a byte comparison beginning at byte
	// zero; reproduce that actual boundary, despite the broader prose spec.
	if at <= 0 || b.width > len(m.data) || at > int64(len(m.data)-b.width) {
		return 0, "", nil
	}
	field := m.data[int(at) : int(at)+b.width]
	if b.noOrder {
		// The pinned matcher accepts a missing byte-order option but no
		// switch case handles it, so the comparison is always false.
		return 0, "", nil
	}
	var value int64
	if b.format == 'i' {
		var unsigned uint64
		if b.little {
			for i := len(field) - 1; i >= 0; i-- {
				unsigned = unsigned<<8 | uint64(field[i])
			}
		} else {
			for _, v := range field {
				unsigned = unsigned<<8 | uint64(v)
			}
		}
		value = int64(unsigned) // matches the reference's uint64-to-int64 cast
	} else {
		pad := 0
		for pad < len(field) && asciiSpace(field[pad]) {
			pad++
		}
		field = field[pad:]
		format := b.format
		if format == 'a' {
			// ClamAV 1.5.4 leaves AUTO unchanged for fewer than three
			// non-padding bytes, so neither numeric switch case runs.
			if len(field) < 3 {
				return 0, "", nil
			}
			if len(field) >= 2 && field[0] == '0' && (field[1] == 'x' || field[1] == 'X') {
				format = 'h'
			} else {
				format = 'd'
			}
		}
		if b.little && format == 'd' {
			// The reference's decimal little-endian switch case is inert.
			return 0, "", nil
		}
		base := uint64(10)
		if format == 'h' {
			base = 16
		}
		var consumed int
		var overflow bool
		if b.little {
			if err := m.w.step(int64(len(field))); err != nil {
				return 0, "", err
			}
			value, consumed, overflow = asciiNumberLittleHex(field)
		} else {
			value, consumed, overflow = asciiNumber(field, base)
		}
		if overflow || b.exact && (pad != 0 || consumed != len(field)) {
			return 0, "", nil
		}
	}
	for _, check := range b.checks[:b.ncheck] {
		switch check.op {
		case '<':
			if value >= check.value {
				return 0, "", nil
			}
		case '>':
			if value <= check.value {
				return 0, "", nil
			}
		case '=':
			if value != check.value {
				return 0, "", nil
			}
		}
	}
	return 1, "", nil
}

func asciiSpace(b byte) bool { return b == ' ' || b >= '\t' && b <= '\r' }

// asciiNumber returns the strtoll-style prefix value used by the reference.
// A field with no digits evaluates to zero unless exact-length mode rejects
// its zero consumed length. The input stays bounded by the scan work budget.
func asciiNumber(field []byte, base uint64) (int64, int, bool) {
	return asciiNumberFrom(len(field), func(i int) byte { return field[i] }, base)
}

func asciiNumberFrom(length int, at func(int) byte, base uint64) (int64, int, bool) {
	i, negative := 0, false
	if length > 0 && (at(0) == '-' || at(0) == '+') {
		negative = at(0) == '-'
		i++
	}
	if base == 16 && length-i >= 3 && at(i) == '0' && (at(i+1) == 'x' || at(i+1) == 'X') && hexNibble(at(i+2)) >= 0 {
		i += 2
	}
	start := i
	limit := uint64(^uint64(0) >> 1)
	if negative {
		limit++
	}
	var magnitude uint64
	for i < length {
		digit := hexNibble(at(i))
		if digit < 0 || uint64(digit) >= base {
			break
		}
		if magnitude > (limit-uint64(digit))/base {
			return 0, i, true
		}
		magnitude = magnitude*base + uint64(digit)
		i++
	}
	if i == start {
		return 0, 0, false
	}
	if negative {
		return -int64(magnitude), i, false
	}
	return int64(magnitude), i, false
}

// asciiNumberLittleHex reproduces the pinned matcher's pair reversal. Its
// normalization writes ASCII zeroes when it encounters a non-hex pair; the
// final reset erases every earlier pair. The virtual buffer avoids allocating
// memory proportional to an untrusted rule's requested width.
func asciiNumberLittleHex(field []byte) (int64, int, bool) {
	n := len(field)
	if n == 1 {
		return asciiNumber(field, 16)
	}
	normLen := n + n%2
	prefix := n >= 3 && field[0] == '0' && (field[1] == 'x' || field[1] == 'X')
	hexAt := func(i int) byte {
		if n%2 == 0 {
			if i == 0 && prefix {
				return 'x'
			}
			return field[i]
		}
		if prefix {
			if i == 0 {
				return 'x'
			}
			if i < 3 {
				return '0'
			}
			return field[i-1]
		}
		if i == 0 {
			return '0'
		}
		return field[i-1]
	}
	valid := func(b byte) bool { return hexNibble(b) >= 0 || b == 'x' || b == 'X' }
	lastReset := 0
	for i := 0; i < normLen; i += 2 {
		if !valid(hexAt(normLen-i-2)) || !valid(hexAt(normLen-i-1)) {
			lastReset = i + 2
		}
	}
	return asciiNumberFrom(normLen, func(i int) byte {
		if i < lastReset {
			return '0'
		}
		return hexAt(normLen - (i &^ 1) - 2 + i%2)
	}, 16)
}
