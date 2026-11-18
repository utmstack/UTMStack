package signatures

import (
	"errors"
	"strconv"
	"strings"
)

// Some rules the reference loads use analysis this engine does not perform
// yet. They are loaded as "unknown when reached": each keeps a condition
// under which the reference could evaluate the missing part on an object,
// and only when that condition holds does the object's result become
// incomplete (never clean, never a detection). Everywhere else the rule is
// decided exactly. Report.Unknown counts these rules by reason, so nothing is
// dropped silently.
const (
	UnknownPCRE = "pcre"
	// A regex whose result depends on where the reference's library gives
	// up (its backtracking limits) on this object.
	UnknownPCRELimit         = "pcre-limit"
	UnknownContainerRegex    = "container-name-regex"
	UnknownFuzzyImage        = "fuzzy-image-hash"
	UnknownByteCompare       = "byte-compare"
	UnknownIconGroup         = "icon-group"
	UnknownHandlerType       = "handler-type"
	UnknownVersionInfo       = "version-info-offset"
	UnknownBytecode          = "bytecode"
	UnknownMetadataJSON      = "metadata-json"
	UnknownCertificateTrust  = "certificate-trust"
	UnknownPhishingAllowlist = "phishing-allowlist"
	// A PE unpacker hook whose outcome depends on a built-in unpacker this
	// engine does not implement.
	UnknownPEUnpacker = "pe-unpacker"
	// Representations the scan pipeline cannot produce yet.
	UnknownHTML = "html-normalization"
	// A macro project whose dump or module text this engine cannot reproduce.
	UnknownVBA = "vba-module-text"
	// Retained for consumers of historical diagnostic records. Word 6 macros
	// are now decoded by the engine; malformed inputs report their own error.
	UnknownWordMacros = "word6-macros"
	// Workbook drawing-group pictures in a form whose extraction by the
	// reference was not observed.
	UnknownWorkbookPictures = "workbook-pictures"
	// A compound file whose reading by the reference this engine cannot
	// tell: sectors smaller than 512 bytes, or a name longer than 64 bytes
	// in a workbook, which the reference keeps or drops depending on what
	// its workbook scan finds.
	UnknownCompoundFile = "compound-file"
)

// UnknownError is the cause of an incomplete result when a rule this engine
// cannot evaluate could apply to the object. errors.As finds it inside
// ErrRuleIncomplete.
type UnknownError struct{ Reason string }

func (e *UnknownError) Error() string {
	return "UTMStack EDR signature needs analysis this engine does not perform: " + e.Reason
}

func unknownBecause(reason string) error { return &UnknownError{Reason: reason} }

// countUnsupported is the count of a placeholder subsignature that the
// reference could evaluate on the object.
const countUnsupported = -4

// placeholder is a subsignature compiled without its matcher.
type placeholder struct {
	reason string
	// trigger decides a regex subsignature: the reference runs the regex only
	// when this expression over the rule's earlier subsignatures holds. A nil
	// trigger is one this engine cannot parse; the count is then unknown.
	trigger *expression
	// ref is the subsignature a byte comparison reads from; it runs only
	// when that subsignature matched.
	ref int
	// bcomp is the verified binary-integer subset of byte comparisons.
	// Other byte-comparison forms remain conditional unknowns.
	bcomp *byteCompare
	// regex is the compiled regex subsignature; nil when it uses a construct
	// this engine does not evaluate (regexErr says which), and the count is
	// then unknown whenever the trigger does not rule it out.
	regex    *regexSub
	regexErr string
	// imageHash is a fuzzy image hash subsignature's exact-match hash.
	imageHash uint64
}

// count is the placeholder's count given the counts of the rule's earlier
// subsignatures and the object's type.
func (p *placeholder) count(earlier []int, typ FileType) int {
	switch p.reason {
	case UnknownPCRE:
		if p.trigger != nil && p.trigger.evaluate(earlier).state == logicFalse {
			return 0
		}
	case UnknownByteCompare:
		if earlier[p.ref] == 0 {
			return 0
		}
	case UnknownFuzzyImage:
		// The reference hashes images only on objects typed as images.
		if targetOf(typ) != 5 {
			return 0
		}
	}
	return countUnsupported
}

// parsePlaceholder recognizes the subsignature kinds this engine loads as
// placeholders: regular expressions, fuzzy image hashes and byte
// comparisons. ok is false for any other subsignature. index is the
// subsignature's own index; references must be to earlier ones.
func parsePlaceholder(sub string, index int) (p *placeholder, ok bool, err error) {
	switch {
	case strings.HasPrefix(sub, "fuzzy_img#"):
		fields := strings.Split(sub[len("fuzzy_img#"):], "#")
		if len(fields) > 2 || len(fields[0]) != 16 || !isHex(fields[0]) || len(fields) == 2 && (fields[1] == "" || !isNumber(fields[1])) {
			return nil, true, errors.New("invalid fuzzy image hash subsignature")
		}
		if len(fields) == 2 && strings.Trim(fields[1], "0") != "" {
			return nil, true, errors.New("nonzero fuzzy image Hamming distance is unsupported by the reference")
		}
		p := &placeholder{reason: UnknownFuzzyImage}
		p.imageHash, _ = strconv.ParseUint(fields[0], 16, 64)
		return p, true, nil
	case !strings.HasPrefix(sub, "$") && strings.Contains(sub, "/"):
		return parseRegexPlaceholder(sub, index)
	}
	if ref, ok := byteCompareRef(sub); ok {
		if ref >= index {
			return nil, true, errors.New("byte comparison must follow the subsignature it reads")
		}
		bcomp, _ := parseBinaryByteCompare(sub, ref)
		return &placeholder{reason: UnknownByteCompare, ref: ref, bcomp: bcomp}, true, nil
	}
	return nil, false, nil
}

// byteCompareRef recognizes N(<<|>>offset#options#comparisons).
func byteCompareRef(sub string) (int, bool) {
	open := strings.IndexByte(sub, '(')
	if open <= 0 || !strings.HasSuffix(sub, ")") || !isNumber(sub[:open]) || strings.Count(sub, "#") < 2 {
		return 0, false
	}
	inner := sub[open+1 : len(sub)-1]
	if !strings.HasPrefix(inner, ">>") && !strings.HasPrefix(inner, "<<") && (inner == "" || inner[0] != '#' && (inner[0] < '0' || inner[0] > '9')) {
		return 0, false
	}
	return int(cAtoi(sub[:open])), true
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		if hexNibble(s[i]) < 0 {
			return false
		}
	}
	return s != ""
}

// parseRegexPlaceholder splits a regex subsignature the reference's way: an
// optional offset field before a ':' outside the regex, then
// trigger/pattern/flags between the first and last '/'. The trigger must be
// a logical expression over earlier subsignatures; the flags must be ones the
// reference knows.
func parseRegexPlaceholder(sub string, index int) (*placeholder, bool, error) {
	fields := regexTokenize(sub, ':', 5)
	body := fields[0]
	if len(fields) == 2 {
		body = fields[1]
	}
	first, last := strings.IndexByte(body, '/'), strings.LastIndexByte(body, '/')
	if first < 0 || last <= first {
		return nil, true, errors.New("regex subsignature needs /pattern/")
	}
	trigger, pattern, flags := body[:first], body[first+1:last], body[last+1:]
	if trigger == "" || pattern == "" {
		return nil, true, errors.New("regex subsignature needs a trigger and a pattern")
	}
	for i := 0; i < len(flags); i++ {
		if !strings.ContainsRune("ismxAEUgre", rune(flags[i])) {
			return nil, true, errors.New("unknown regex subsignature flag")
		}
	}
	highest, err := lsigHighestIndex(trigger)
	if err != nil {
		return nil, true, err
	}
	if highest >= index {
		return nil, true, errors.New("regex trigger must name earlier subsignatures")
	}
	p := &placeholder{reason: UnknownPCRE}
	if expr, err := parseExpression(trigger, index); err == nil {
		p.trigger = expr
	}
	offset := "*"
	if len(fields) == 2 {
		offset = fields[0]
	}
	// The PCRE dynamic configuration cannot be switched off (dconf.go), so
	// its global bit is always on.
	if re, err := compileRegexSub(offset, pattern, flags, DefaultDynamicConfig().PCRE); err == nil && p.trigger != nil {
		p.regex = re
	} else if err != nil {
		p.regexErr = err.Error()
	} else {
		p.regexErr = "trigger expression not evaluated"
	}
	return p, true, nil
}

// regexTokenize splits s at delim, except inside a /regex/: an unescaped '/'
// toggles the regex state, as the reference's tokenizer does. At most limit
// fields are returned.
func regexTokenize(s string, delim byte, limit int) []string {
	var fields []string
	within := false
	start := 0
	for i := 0; i < len(s) && len(fields) < limit-1; i++ {
		switch {
		case s[i] == '/' && (i == 0 || s[i-1] != '\\'):
			within = !within
		case s[i] == delim && !within:
			fields = append(fields, s[start:i])
			start = i + 1
		}
	}
	return append(fields, s[start:])
}

func (d *Database) unknown(reason string, n int) {
	if d.report.Unknown == nil {
		d.report.Unknown = map[string]int{}
	}
	d.report.Unknown[reason] += n
	if stats := d.sourceStats[d.activeSource]; stats != nil {
		stats.unknown[reason] += n
	}
	if stats := d.memberStats[d.activeMember]; stats != nil {
		stats.unknown[reason] += n
	}
	d.lastUnknown = reason
}

// everywhere reports a pattern that is a candidate on every object: one
// without anchors, other than regex and byte-comparison placeholders, whose
// counts follow from the rule's other subsignatures.
func (p *bodyPattern) everywhere() bool {
	return !p.anchored && (p.holder == nil || p.holder.reason == UnknownFuzzyImage)
}

// maxIconGroupNames is how many distinct names each group field may have.
const maxIconGroupNames = 256

// addIconLine loads one icon-group database entry: a name, two group names
// and the icon's hash (iconhash.go).
func (d *Database) addIconLine(line, entry string) error {
	p := cStrtokenize(line, ':', 5)
	if len(p) != 4 || len(p[3]) != 124 || !isHex(p[3]) {
		return errors.New("invalid icon signature fields")
	}
	if d.ignore.ignored(p[0], entry) {
		return skip(SkipIgnoreList)
	}
	h, err := parseIconHash(p[3])
	if err != nil {
		return err
	}
	for i, name := range p[1:3] {
		if d.iconNames[i] == nil {
			d.iconNames[i] = map[string]bool{}
		}
		if !d.iconNames[i][name] {
			if len(d.iconNames[i]) == maxIconGroupNames {
				return errors.New("too many icon groups")
			}
			d.iconNames[i][strings.Clone(name)] = true
		}
	}
	d.report.IconData++
	d.iconEntries = append(d.iconEntries, iconEntry{[2]string{strings.Clone(p[1]), strings.Clone(p[2])}, h})
	return nil
}
