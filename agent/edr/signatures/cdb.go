package signatures

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Container metadata rules (.cdb) name archive members by their metadata,
// checked before a member is extracted: the archive's type and size, the
// member's name, sizes, encryption flag and position. The line format is
// Name:ContainerType:ContainerSize:NameRegex:CompressedSize:Size:Encrypted:
// Position:Res1:Res2[:MinFL[:MaxFL]]. Res1 (a CRC) is never read, as in the
// reference.

// MemberMeta describes one archive member as its archive's parser reports it.
type MemberMeta struct {
	Name           string
	HasName        bool
	CompressedSize int64
	Size           int64
	Encrypted      bool
	Index          int // parser-supplied position; CAB and CHM use zero-based positions
}

type metaRange struct {
	any      bool
	min, max uint64
}

// contains is the reference's range test: an exact value when both ends are
// equal, otherwise 0 means no bound on either side; an inverted range never
// holds.
func (r metaRange) contains(v int64) bool {
	if r.any {
		return true
	}
	u := uint64(v)
	if r.min == r.max {
		return u == r.min
	}
	return !(r.min != 0 && r.min > u || r.max != 0 && r.max < u)
}

type cdbRule struct {
	name                             string
	container                        *FileType
	archive, compressed, size, index metaRange
	encrypted                        int8 // -1: either
	regex                            *regexp.Regexp
	regexUnknown                     bool
}

func parseMetaRange(s string) (metaRange, error) {
	if s == "*" {
		return metaRange{any: true}, nil
	}
	if isNumber(s) && s != "" {
		v := uint64(uint32(cAtoi(s)))
		return metaRange{min: v, max: v}, nil
	}
	lo, hi, ok := strings.Cut(s, "-")
	if !ok || lo == "" || hi == "" || !isNumber(lo) || !isNumber(hi) {
		return metaRange{}, errors.New("invalid container metadata range")
	}
	return metaRange{min: uint64(uint32(cAtoi(lo))), max: uint64(uint32(cAtoi(hi)))}, nil
}

func (d *Database) addContainerLine(line string) error {
	p := cStrtokenize(line, ':', 13)
	if len(p) < 10 || len(p) > 12 {
		return errors.New("invalid container signature fields")
	}
	if len(p) > 10 {
		var maximum *string
		if len(p) == 12 {
			maximum = &p[11]
		}
		if err := d.fieldWindow(p[10], maximum); err != nil {
			return err
		}
	}
	// The reference compares only the name with ignore lists here.
	if d.ignore.ignored(p[0], p[0]) {
		return skip(SkipIgnoreList)
	}
	if err := validName(p[0]); err != nil {
		return err
	}
	r := cdbRule{name: strings.Clone(p[0]), encrypted: -1}
	if p[1] != "*" {
		t, ok := ParseFileType(p[1])
		if !ok {
			return errors.New("unknown container type in container signature")
		}
		r.container = &t
	}
	var err error
	for i, target := range []*metaRange{&r.archive, &r.compressed, &r.size, &r.index} {
		field := []int{2, 4, 5, 7}[i]
		if *target, err = parseMetaRange(p[field]); err != nil {
			return err
		}
	}
	switch p[6] {
	case "*":
	case "0", "1":
		r.encrypted = int8(p[6][0] - '0')
	default:
		return errors.New("invalid container signature encryption flag")
	}
	if p[3] != "*" {
		// POSIX extended regex, searched without anchoring; a leading (?i)
		// makes it case-insensitive; '.' matches a newline too.
		expr, flags := p[3], "(?s)"
		if rest, ok := strings.CutPrefix(expr, "(?i)"); ok {
			expr, flags = rest, "(?is)"
		}
		var translated string
		var uncertain bool
		translated, uncertain, err = containerERE(expr)
		if err != nil {
			return errors.New("invalid container signature file name expression")
		}
		r.regex, err = regexp.Compile(flags + translated)
		if err != nil {
			return errors.New("invalid container signature file name expression")
		}
		if uncertain {
			r.regex = nil
			r.regexUnknown = true
			d.unknown(UnknownContainerRegex, 1)
		}
	}
	d.containers = append(d.containers, r)
	return nil
}

// ClamAV's bundled ERE parser treats backslash followed by an ordinary byte
// (including a digit, b, n or q) as that literal byte. Go interprets some of
// these as classes or rejects them. Its \\< and \\> are directional word
// boundaries, which Go's \\b cannot represent exactly; those remain
// conditional unknowns. In a bracket each slash is an ordinary literal byte.
func containerERE(expr string) (string, bool, error) {
	var normalized strings.Builder
	uncertain, inClass := false, false
	for i := 0; i < len(expr); {
		switch {
		case expr[i] == '\\' && i+1 == len(expr):
			return "", false, errors.New("trailing regex escape")
		case expr[i] == '\\' && inClass:
			// The bundled parser reads each backslash inside a bracket as an
			// ordinary member, while Go treats it as an escape. Double it.
			normalized.WriteString(`\\`)
			i++
		case expr[i] == '\\' && (expr[i+1] == '<' || expr[i+1] == '>'):
			normalized.WriteByte('x')
			uncertain = true
			i += 2
		case expr[i] == '\\':
			normalized.WriteString(regexp.QuoteMeta(expr[i+1 : i+2]))
			i += 2
		case expr[i] == '[':
			inClass = true
			normalized.WriteByte(expr[i])
			i++
		case expr[i] == ']':
			inClass = false
			normalized.WriteByte(expr[i])
			i++
		default:
			normalized.WriteByte(expr[i])
			i++
		}
	}
	return normalized.String(), uncertain, nil
}

// MatchMember checks a member against the container metadata rules, the
// last loaded first, as the reference keeps them. archive is the archive's
// own type and archiveSize its length.
func (d *Database) MatchMember(ctx context.Context, archive FileType, archiveSize int64, m MemberMeta) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	name := m.Name
	if i := strings.IndexByte(name, 0); i >= 0 {
		name = name[:i]
	}
	if len(name) > 255 {
		name = name[:255]
	}
	var pending error
	for i := len(d.containers) - 1; i >= 0; i-- {
		r := &d.containers[i]
		switch {
		case r.container != nil && *r.container != archive,
			r.encrypted >= 0 && (r.encrypted == 1) != m.Encrypted,
			!r.archive.contains(archiveSize), !r.compressed.contains(m.CompressedSize),
			!r.size.contains(m.Size), !r.index.contains(int64(m.Index)),
			r.regex != nil && (!m.HasName || !r.regex.MatchString(name)):
			continue
		}
		if r.regexUnknown {
			if m.HasName && pending == nil {
				pending = fmt.Errorf("%w: %w", ErrRuleIncomplete, unknownBecause(UnknownContainerRegex))
			}
			continue
		}
		return r.name, nil
	}
	return "", pending
}
