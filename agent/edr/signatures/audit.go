package signatures

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// Audit measures, rule by rule, how much of a signature feed this compiler can
// represent. It never produces a loadable database: a feed with any rejected
// rule still fails Compile. Its only purposes are gap accounting and release
// gating, so a publisher can see exactly what an unmodified feed would lose.

// Rule-set kinds separate missing detections from missing engine metadata.
const (
	KindDetection  = "detection"
	KindPUA        = "pua-detection"
	KindAllowlist  = "allowlist"
	KindIgnore     = "ignore-list"
	KindEngineData = "engine-data"
	KindUnknown    = "unknown"
)

var formatKinds = map[string]string{
	".hdb": KindDetection, ".hsb": KindDetection, ".mdb": KindDetection, ".msb": KindDetection,
	".ndb": KindDetection, ".ldb": KindDetection, ".db": KindDetection, ".cbc": KindDetection,
	".cdb": KindDetection, ".pdb": KindDetection, ".gdb": KindDetection,
	".yar": KindDetection, ".yara": KindDetection,
	".hdu": KindPUA, ".hsu": KindPUA, ".mdu": KindPUA, ".msu": KindPUA, ".ndu": KindPUA, ".ldu": KindPUA,
	".fp": KindAllowlist, ".sfp": KindAllowlist, ".wdb": KindAllowlist, ".crb": KindAllowlist, ".cat": KindAllowlist,
	".ign": KindIgnore, ".ign2": KindIgnore,
	".cfg": KindEngineData, ".ftm": KindEngineData, ".idb": KindEngineData, ".pwdb": KindEngineData,
}

func formatKind(ext string) string {
	if k, ok := formatKinds[ext]; ok {
		return k
	}
	return KindUnknown
}

// AuditOptions describes the engine being replaced. Rules it would not load
// (outside its functionality window at ReferenceLevel, named by an ignore
// list, PUA databases unless ReferencePUA, ...) are counted as
// reference-skipped by reason, not as gaps.
type AuditOptions struct {
	ReferenceLevel int
	// ReferencePUA mirrors the replaced engine's potentially-unwanted switch.
	// When false, PUA-only databases are reference-skipped.
	ReferencePUA bool
	// Limits bound the audit itself; zero fields select compiler defaults.
	// Production-limit findings are always computed against the defaults.
	Limits   Limits
	Examples int
}

type AuditExample struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Rule string `json:"rule"`
}

type AuditReason struct {
	Format   string         `json:"format"`
	Kind     string         `json:"kind"`
	Reason   string         `json:"reason"`
	Count    int            `json:"count"`
	Features map[string]int `json:"features,omitempty"`
	Examples []AuditExample `json:"examples,omitempty"`
}

type AuditCounts struct {
	Rules                    int `json:"rules"`
	Accepted                 int `json:"accepted"`
	Rejected                 int `json:"rejected"`
	ReferenceSkipped         int `json:"reference_skipped"`
	AcceptedOutsideReference int `json:"accepted_outside_reference"`
	// Unknown counts the accepted rules loaded as unknown when reached: they
	// make an object incomplete where they could apply.
	Unknown int `json:"unknown_when_reached"`
}

func (c *AuditCounts) add(o AuditCounts) {
	c.Rules += o.Rules
	c.Accepted += o.Accepted
	c.Rejected += o.Rejected
	c.ReferenceSkipped += o.ReferenceSkipped
	c.AcceptedOutsideReference += o.AcceptedOutsideReference
	c.Unknown += o.Unknown
}

type AuditFile struct {
	Name   string `json:"name"`
	Format string `json:"format"`
	Kind   string `json:"kind"`
	Bytes  int    `json:"bytes"`
	AuditCounts
	// SkippedByReason splits ReferenceSkipped by selection reason.
	SkippedByReason map[string]int `json:"reference_skipped_by_reason,omitempty"`
	// UnknownByReason splits Unknown by the analysis the rules need.
	UnknownByReason map[string]int `json:"unknown_when_reached_by_reason,omitempty"`
	Reasons         []AuditReason  `json:"reasons,omitempty"`
}

type AuditReport struct {
	EngineLevel     int                    `json:"engine_level"`
	ReferenceLevel  int                    `json:"reference_level"`
	ReferencePUA    bool                   `json:"reference_pua"`
	InputBytes      int64                  `json:"input_bytes"`
	ExpandedBytes   int64                  `json:"expanded_bytes"`
	Containers      []ContainerInfo        `json:"containers,omitempty"`
	DynamicConfig   DynamicConfig          `json:"dynamic_config"`
	Files           []AuditFile            `json:"files"`
	Totals          AuditCounts            `json:"totals"`
	SkippedByReason map[string]int         `json:"reference_skipped_by_reason,omitempty"`
	UnknownByReason map[string]int         `json:"unknown_when_reached_by_reason,omitempty"`
	ByKind          map[string]AuditCounts `json:"by_kind"`
	Reasons         []AuditReason          `json:"reasons"`
	LimitFindings   []string               `json:"limit_findings,omitempty"`
	Errors          []Diagnostic           `json:"errors,omitempty"`
}

// Complete reports whether every rule the reference engine loads was accepted
// and every input could be read. Production-limit findings are separate.
func (r *AuditReport) Complete() bool { return r.Totals.Rejected == 0 && len(r.Errors) == 0 }

var errLevelOnly = errors.New("functionality window excludes this engine's level; syntax otherwise supported")

var errPUADisabled = errors.New("potentially unwanted application databases are not loaded by this engine")

// auditProbeRules bounds how many rules a probe database accumulates before a
// fresh one replaces it; outcomes do not depend on earlier rules.
const auditProbeRules = 20000

type auditor struct {
	opts    AuditOptions
	limits  Limits
	report  *AuditReport
	engine  *auditProbe
	ref     *auditProbe // the engine probe when the levels are equal
	global  map[string]*AuditReason
	perFile map[string]*AuditReason
}

// auditProbe compiles rules one at a time at one level with the generation's
// ignore lists and dynamic configuration, only to learn each outcome.
type auditProbe struct {
	level   int
	limits  Limits
	ignore  *ignoreList
	dconf   *dconfState
	catalog *Database
	db      *Database
	added   int
}

func (p *auditProbe) database() *Database {
	if p.db == nil || p.added >= auditProbeRules {
		p.db = newDatabase(p.limits, p.level)
		p.db.ignore, p.db.dconf = p.ignore, p.dconf.config
		p.added = 0
	}
	p.added++
	return p.db
}

// Audit compiles every rule independently with the compiler's own selection
// and records whether this engine accepts it, rejects it (and why), or
// whether the reference engine skips it anyway (and why). Unreadable
// containers and malformed selection files are reported in Errors.
func Audit(files map[string][]byte, opts AuditOptions) (*AuditReport, error) {
	if opts.ReferenceLevel <= 0 {
		return nil, errors.New("reference functionality level is required")
	}
	if opts.Examples <= 0 {
		opts.Examples = 3
	}
	limits := opts.Limits.defaults()
	a := &auditor{
		opts:   opts,
		limits: limits,
		report: &AuditReport{EngineLevel: FunctionalityLevel, ReferenceLevel: opts.ReferenceLevel, ReferencePUA: opts.ReferencePUA, ByKind: map[string]AuditCounts{}},
		global: map[string]*AuditReason{},
	}
	names := make([]string, 0, len(files))
	var total int64
	for name, b := range files {
		total += int64(len(b))
		if err := ValidateFileName(name); err != nil {
			a.report.Errors = append(a.report.Errors, Diagnostic{File: name, Reason: err.Error()})
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	if total > limits.MaxTotalBytes {
		return nil, fmt.Errorf("audit input is %d bytes, above the audit limit of %d", total, limits.MaxTotalBytes)
	}
	a.report.InputBytes = total
	budget := archiveBudget{limits: limits, total: total, files: len(files)}
	addError := func(file string, line int, err error) {
		a.report.Errors = append(a.report.Errors, Diagnostic{File: file, Line: line, Reason: err.Error()})
	}
	unpack := newDatabase(limits, FunctionalityLevel)
	sources := unpack.expand(loadOrder(names), files, &budget, addError)
	a.report.Containers = unpack.report.Containers
	a.engine = a.newProbe(sources, FunctionalityLevel, addError)
	a.ref = a.engine
	if opts.ReferenceLevel != FunctionalityLevel {
		a.ref = a.newProbe(sources, opts.ReferenceLevel, func(string, int, error) {})
	}
	a.report.DynamicConfig = a.engine.dconf.config
	for _, src := range sources {
		a.file(src)
	}
	a.report.ExpandedBytes = budget.total
	for _, r := range a.global {
		a.report.Reasons = append(a.report.Reasons, *r)
	}
	sortReasons(a.report.Reasons)
	a.limitFindings(files, budget)
	return a.report, nil
}

// newProbe reads the generation's ignore lists and dynamic configuration at
// level, as the compiler does before compiling any rule.
func (a *auditor) newProbe(sources []source, level int, addError func(string, int, error)) *auditProbe {
	d := newDatabase(a.limits, level)
	state := newDconfState()
	for _, src := range sources {
		switch src.ext {
		case ".ign", ".ign2":
			if d.ignore == nil {
				d.ignore = &ignoreList{}
			}
			if line, err := d.ignore.load(src.data, src.nested); err != nil {
				addError(src.name, line, err)
			}
		case ".cfg":
			if err := eachLine(src.data, a.limits.MaxLineBytes, func(n int, line string) { _, _, _ = state.apply(src.name, n, line, level) }); err != nil {
				addError(src.name, 0, err)
			}
		}
	}
	d.dconf = state.config
	// Catalogs are binary, signed sources. Their CRB roots may appear after
	// the catalog in an archive, so prepare the complete selected root set.
	if !d.dconfDisabled(".crb") {
		for _, src := range sources {
			if src.ext != ".crb" {
				continue
			}
			var certificateState certificateSourceState
			_ = eachLine(src.data, a.limits.MaxLineBytes, func(_ int, line string) {
				s := strings.TrimSpace(line)
				if s != "" && !strings.HasPrefix(s, "#") {
					_ = d.addCertificateSourceLine(s, &certificateState) // per-line diagnostics are recorded by a.file
				}
			})
		}
	}
	return &auditProbe{level: level, limits: a.limits, ignore: d.ignore, dconf: state, catalog: d}
}

func sortReasons(r []AuditReason) {
	sort.Slice(r, func(i, j int) bool {
		if r[i].Count != r[j].Count {
			return r[i].Count > r[j].Count
		}
		if r[i].Format != r[j].Format {
			return r[i].Format < r[j].Format
		}
		return r[i].Reason < r[j].Reason
	})
}

func (a *auditor) limitFindings(files map[string][]byte, budget archiveBudget) {
	prod := DefaultLimits()
	r := a.report
	if r.InputBytes > prod.MaxTotalBytes {
		r.LimitFindings = append(r.LimitFindings, fmt.Sprintf("input is %d bytes; the production database limit is %d", r.InputBytes, prod.MaxTotalBytes))
	}
	if r.ExpandedBytes > prod.MaxTotalBytes {
		r.LimitFindings = append(r.LimitFindings, fmt.Sprintf("input expands to %d bytes; the production database limit is %d", r.ExpandedBytes, prod.MaxTotalBytes))
	}
	if budget.files > prod.MaxFiles {
		r.LimitFindings = append(r.LimitFindings, fmt.Sprintf("input holds %d files; the production limit is %d", budget.files, prod.MaxFiles))
	}
	for _, f := range r.Files {
		if strings.Contains(f.Name, "/") && int64(f.Bytes) > prod.MaxFileBytes {
			r.LimitFindings = append(r.LimitFindings, fmt.Sprintf("%s is %d bytes; the production per-file limit is %d", f.Name, f.Bytes, prod.MaxFileBytes))
		}
	}
	for name, b := range files {
		if int64(len(b)) > prod.MaxFileBytes {
			r.LimitFindings = append(r.LimitFindings, fmt.Sprintf("%s is %d bytes; the production per-file limit is %d", name, len(b), prod.MaxFileBytes))
		}
	}
	if loaded := r.Totals.Accepted + r.Totals.Rejected; loaded > prod.MaxSignatures {
		r.LimitFindings = append(r.LimitFindings, fmt.Sprintf("closing every gap would load %d rules; the production signature limit is %d", loaded, prod.MaxSignatures))
	}
	sort.Strings(r.LimitFindings)
}

func (a *auditor) file(src source) {
	base := filepath.Base(src.name)
	if src.nested && (src.ext == ".info" || src.ext == "" && (base == "COPYING" || base == "LICENSE")) {
		return // container metadata the compiler also ignores
	}
	f := &AuditFile{Name: src.name, Format: src.ext, Kind: formatKind(src.ext), Bytes: len(src.data)}
	a.perFile = map[string]*AuditReason{}
	switch src.ext {
	case ".cbc":
		a.bytecode(f, src.data)
	case ".cat":
		a.catalog(f, src)
	case ".ign", ".ign2":
		a.ignoreList(f, src)
	case ".cfg":
		a.dynamicConfig(f, src)
	default:
		a.lines(f, src.data)
	}
	for _, r := range a.perFile {
		f.Reasons = append(f.Reasons, *r)
	}
	sortReasons(f.Reasons)
	a.report.Files = append(a.report.Files, *f)
	a.report.Totals.add(f.AuditCounts)
	for reason, n := range f.SkippedByReason {
		if a.report.SkippedByReason == nil {
			a.report.SkippedByReason = map[string]int{}
		}
		a.report.SkippedByReason[reason] += n
	}
	for reason, n := range f.UnknownByReason {
		if a.report.UnknownByReason == nil {
			a.report.UnknownByReason = map[string]int{}
		}
		a.report.UnknownByReason[reason] += n
	}
	k := a.report.ByKind[f.Kind]
	k.add(f.AuditCounts)
	a.report.ByKind[f.Kind] = k
}

func (a *auditor) catalog(f *AuditFile, src source) {
	at := time.Now().UTC()
	check := func(p *auditProbe) error {
		if p.catalog.dconfDisabled(".cat") {
			return skip(SkipDynamicConfig)
		}
		_, err := p.catalog.verifiedCatalogEntries(src.data, at)
		if err != nil {
			return skip(SkipInvalidCatalog)
		}
		return nil
	}
	engineErr := check(a.engine)
	refErr := engineErr
	if a.ref != a.engine {
		refErr = check(a.ref)
	}
	a.record(f, 0, "[signed catalog]", engineErr, refErr, "")
}

// record classifies one rule from the outcome this engine produced and the
// outcome the reference would (a nil error or an error means the reference
// loads the rule; a skip means it does not). unknown is the reason an
// accepted rule loaded as unknown when reached, or "".
func (a *auditor) record(f *AuditFile, line int, rule string, engineErr, refErr error, unknown string, extra ...string) {
	f.Rules++
	refReason, refSkips := skipReason(refErr)
	engineReason, engineSkips := skipReason(engineErr)
	switch {
	case engineErr == nil:
		f.Accepted++
		if refSkips {
			f.AcceptedOutsideReference++
		}
		if unknown != "" {
			f.Unknown++
			if f.UnknownByReason == nil {
				f.UnknownByReason = map[string]int{}
			}
			f.UnknownByReason[unknown]++
		}
		return
	case refSkips:
		f.ReferenceSkipped++
		if f.SkippedByReason == nil {
			f.SkippedByReason = map[string]int{}
		}
		f.SkippedByReason[refReason]++
		return
	case engineSkips && engineReason == SkipLevel:
		engineErr = errLevelOnly
		if refErr != nil {
			engineErr = refErr
		}
	case engineSkips && engineReason == SkipPUA:
		engineErr = errPUADisabled
	case engineSkips:
		engineErr = fmt.Errorf("this engine does not select the rule (%s) but the reference does", engineReason)
	}
	f.Rejected++
	a.reject(f, line, rule, engineErr, extra...)
}

var puaBase = map[string]string{".hdu": ".hdb", ".hsu": ".hsb", ".mdu": ".mdb", ".msu": ".msb", ".ndu": ".ndb", ".ldu": ".ldb"}

// outcome compiles one line-format rule at the probe's level. unknown is the
// reason the rule loaded as unknown when reached, or "".
func (a *auditor) outcome(p *auditProbe, ext, line, entry string, pua bool) (unknown string, err error) {
	if base, ok := puaBase[ext]; ok {
		if !pua {
			return "", skip(SkipPUA)
		}
		ext = base
	}
	d := p.database()
	if d.dconfDisabled(ext) {
		return "", skip(SkipDynamicConfig)
	}
	d.lastUnknown = ""
	switch ext {
	case ".hdb", ".hsb", ".mdb", ".msb", ".fp", ".sfp", ".ndb", ".ldb", ".db", ".pdb", ".wdb", ".gdb", ".ftm", ".idb", ".crb", ".cdb", ".pwdb":
		err = d.addLine(ext, line, entry)
	default:
		err = d.selectUnsupported(ext, line, entry)
	}
	return d.lastUnknown, err
}

func (a *auditor) lines(f *AuditFile, b []byte) {
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 4096), a.limits.MaxLineBytes)
	line := 0
	for sc.Scan() {
		line++
		raw := sc.Text()
		s := strings.TrimSpace(raw)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		entry := strings.TrimRight(raw, "\r\n")
		unknown, engineErr := a.outcome(a.engine, f.Format, s, entry, false)
		refErr := engineErr
		if a.ref != a.engine || a.opts.ReferencePUA && puaBase[f.Format] != "" {
			_, refErr = a.outcome(a.ref, f.Format, s, entry, a.opts.ReferencePUA)
		}
		a.record(f, line, s, engineErr, refErr, unknown)
	}
	if err := sc.Err(); err != nil {
		a.report.Errors = append(a.report.Errors, Diagnostic{File: f.Name, Line: line, Reason: "signature line limit or read error: " + err.Error()})
	}
}

// ignoreList counts each entry line of an ignore list as a rule. A malformed
// line stops the list, as it stops the reference's database load.
func (a *auditor) ignoreList(f *AuditFile, src source) {
	failed, err := (&ignoreList{}).load(src.data, src.nested)
	for i, line := range strings.Split(strings.TrimSuffix(string(src.data), "\n"), "\n") {
		number := i + 1
		if strings.HasPrefix(line, "#") {
			continue
		}
		var outcome error
		switch {
		case err == nil || number < failed:
		case number == failed:
			outcome = err
		default:
			outcome = errors.New("ignore list not loaded after an earlier malformed line")
		}
		a.record(f, number, line, outcome, outcome, "")
	}
}

// dynamicConfig counts each module directive of a .cfg file as a rule:
// accepted when it applies and this engine honors the resulting
// configuration, reference-skipped when it does not apply at that level.
func (a *auditor) dynamicConfig(f *AuditFile, src source) {
	unsupported := map[string]error{}
	for _, diagnostic := range a.engine.dconf.unsupported() {
		if diagnostic.File == src.name {
			unsupported[strconv.Itoa(diagnostic.Line)] = errors.New(diagnostic.Reason)
		}
	}
	directive := func(p *auditProbe, number int, line string) (bool, error) {
		state := newDconfState()
		isDirective, applied, err := state.apply(src.name, number, line, p.level)
		switch {
		case err != nil:
			return true, err
		case !applied:
			return isDirective, skip(SkipLevel)
		}
		if e, ok := unsupported[strconv.Itoa(number)]; ok && p == a.engine {
			return true, e
		}
		return true, nil
	}
	err := eachLine(src.data, a.limits.MaxLineBytes, func(number int, line string) {
		isDirective, engineErr := directive(a.engine, number, line)
		if !isDirective {
			return
		}
		_, refErr := directive(a.ref, number, line)
		a.record(f, number, line, engineErr, refErr, "")
	})
	if err != nil {
		a.report.Errors = append(a.report.Errors, Diagnostic{File: f.Name, Reason: err.Error()})
	}
}

var bytecodeKinds = map[uint64]string{
	0: "generic", 1: "startup", 256: "logical", 257: "pe-unpacker", 258: "pdf",
	259: "pe-all", 260: "preclass", 261: "elf-unpacker", 262: "macho-unpacker",
}

func (a *auditor) bytecode(f *AuditFile, b []byte) {
	result := func(p *auditProbe) (string, error) {
		d := p.database()
		if d.dconfDisabled(".cbc") {
			return "", skip(SkipDynamicConfig)
		}
		d.lastUnknown = ""
		_, err := d.addBytecode(b)
		return d.lastUnknown, err
	}
	unknown, engineErr := result(a.engine)
	refErr := engineErr
	if a.ref != a.engine {
		_, refErr = result(a.ref)
	}
	rule, extra := "", []string(nil)
	if p, err := bytecode.Parse(b); err == nil {
		kind, ok := bytecodeKinds[p.Kind]
		if !ok {
			kind = strconv.FormatUint(p.Kind, 10)
		}
		rule, extra = "hook:"+kind+" "+strings.TrimSpace(p.LogicalSignature), []string{"hook:" + kind}
	}
	a.record(f, 0, rule, engineErr, refErr, unknown, extra...)
}

var quotedReason = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)

// normalizeReason groups diagnostics that differ only in a long quoted value.
func normalizeReason(err error) string {
	return quotedReason.ReplaceAllStringFunc(err.Error(), func(q string) string {
		if len(q) <= 14 {
			return q
		}
		return q[:9] + `..."`
	})
}

func (a *auditor) reject(f *AuditFile, line int, rule string, err error, extra ...string) {
	reason := normalizeReason(err)
	key := f.Format + "\x00" + reason
	features := ruleFeatures(f.Format, rule)
	features = append(features, extra...)
	for _, m := range []map[string]*AuditReason{a.global, a.perFile} {
		r := m[key]
		if r == nil {
			r = &AuditReason{Format: f.Format, Kind: f.Kind, Reason: reason, Features: map[string]int{}}
			m[key] = r
		}
		r.Count++
		for _, feature := range features {
			r.Features[feature]++
		}
		if len(r.Examples) < a.opts.Examples {
			if len(rule) > 240 {
				rule = rule[:240] + "..."
			}
			r.Examples = append(r.Examples, AuditExample{File: f.Name, Line: line, Rule: rule})
		}
	}
}

// ruleFeatures tags the constructs a rejected rule uses, so reasons such as
// "unsupported body syntax" can be broken down into implementable work.
func ruleFeatures(ext, s string) []string {
	if b, ok := puaBase[ext]; ok {
		ext = b
	}
	var out []string
	add := func(f string) { out = append(out, f) }
	switch ext {
	case ".ndb":
		p := strings.Split(s, ":")
		if len(p) >= 4 {
			add("target:" + p[1])
			out = append(out, offsetFeatures(p[2])...)
			out = append(out, bodyFeatures(p[3])...)
		}
	case ".db":
		if _, body, ok := strings.Cut(s, "="); ok {
			out = append(out, bodyFeatures(body)...)
		}
	case ".ldb":
		p := strings.Split(s, ";")
		if len(p) < 4 {
			return out
		}
		for _, d := range strings.Split(p[1], ",") {
			k, v, _ := strings.Cut(d, ":")
			switch k {
			case "Target":
				add("target:" + v)
			case "Engine", "FileSize", "EntryPoint", "NumberOfSections":
			default:
				add("descriptor:" + k)
			}
		}
		for _, sub := range p[3:] {
			switch {
			case strings.HasPrefix(sub, "${"):
				add("subsig:macro")
			case strings.Contains(sub, "fuzzy_img#"):
				add("subsig:fuzzy-image")
			case strings.Contains(sub, "/"):
				add("subsig:pcre")
			case strings.Contains(sub, "#"):
				add("subsig:byte-compare")
			default:
				off, body := "*", sub
				if i := strings.IndexByte(sub, ':'); i >= 0 && (i+1 == len(sub) || sub[i+1] != ':') {
					off, body = sub[:i], sub[i+1:]
				}
				out = append(out, offsetFeatures(off)...)
				out = append(out, bodyFeatures(body)...)
			}
		}
	}
	return dedupe(out)
}

func offsetFeatures(off string) []string {
	var out []string
	if strings.Contains(off, ",") {
		out = append(out, "offset:max-shift")
		off, _, _ = strings.Cut(off, ",")
	}
	switch {
	case off == "*" || off == "":
	case strings.HasPrefix(off, "EP"):
		out = append(out, "offset:entry-point")
	case strings.HasPrefix(off, "EOF"):
		out = append(out, "offset:eof")
	case strings.HasPrefix(off, "VI"):
		out = append(out, "offset:version-info")
	case strings.HasPrefix(off, "SL"):
		out = append(out, "offset:last-section")
	case strings.HasPrefix(off, "SE"):
		out = append(out, "offset:entire-section")
	case strings.HasPrefix(off, "S"):
		out = append(out, "offset:section")
	case strings.HasPrefix(off, "Macro") || strings.HasPrefix(off, "$"):
		out = append(out, "offset:macro")
	}
	return out
}

func bodyFeatures(body string) []string {
	var out []string
	body, mods, hasMods := strings.Cut(body, "::")
	if hasMods {
		for _, m := range mods {
			out = append(out, "modifier:"+string(m))
		}
	}
	for marker, feature := range map[string]string{
		"(B)": "body:boundary", "(L)": "body:line-boundary", "(W)": "body:word-boundary",
		"!(": "body:negated-alternative", "[": "body:anchored-range", "{": "body:gap",
		"*": "body:wildcard", "?": "body:nibble", "(": "body:alternative",
	} {
		if strings.Contains(body, marker) {
			out = append(out, feature)
		}
	}
	return out
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
