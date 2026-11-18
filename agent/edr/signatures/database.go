// Package signatures compiles a deliberately bounded, explicit subset of malware
// signatures into an immutable database. Supported logical bytecode callbacks
// run in a bounded Go sandbox. Unsupported input is an error, never a skipped
// rule; only rules the reference engine itself does not select are skipped,
// and they are counted.
package signatures

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// FunctionalityLevel selects rules exactly as the reference engine release at
// this level does (its functionality windows, ignore lists and dynamic
// configuration). It is not a claim of complete compatibility: each selected
// rule is still parsed against the subset documented in NATIVE_ENGINE.md. It is
// the level bytecode programs read through their engine-level APIs.
const FunctionalityLevel = bytecode.EngineLevel

// Limits bound database expansion and per-object matching. Zero fields select
// defaults. Exhausting a matching budget returns an error, never a clean verdict.
type Limits struct {
	MaxFileBytes  int64
	MaxTotalBytes int64
	MaxFiles      int
	MaxSignatures int
	MaxLineBytes  int
	MaxMatchSteps int64
	// MatchStepsPerByte lets the matching budget grow with the object:
	// the budget is max(MaxMatchSteps, MatchStepsPerByte*len(object)).
	// Real executables need about 2-12 steps per byte with the official feed,
	// so a fixed budget alone leaves very large objects incomplete. It
	// defaults to 16 only when neither budget field is set, so an explicit
	// MaxMatchSteps stays a hard cap.
	MatchStepsPerByte int64
}

// DefaultLimits are the bounds used when a Limits field is zero. They admit
// the complete official feed with headroom: its three containers are about
// 108 MiB, expand to about 416 MiB including the inputs, and the largest member
// is about 159 MiB, with about 3.6 million rules.
func DefaultLimits() Limits { return Limits{}.defaults() }

func (l Limits) defaults() Limits {
	if l.MaxFileBytes <= 0 {
		l.MaxFileBytes = 384 << 20
	}
	if l.MaxTotalBytes <= 0 {
		l.MaxTotalBytes = 1 << 30
	}
	if l.MaxFiles <= 0 {
		l.MaxFiles = 4096
	}
	if l.MaxSignatures <= 0 {
		l.MaxSignatures = 8000000
	}
	if l.MaxLineBytes <= 0 {
		l.MaxLineBytes = 64 << 10
	}
	if l.MaxMatchSteps <= 0 && l.MatchStepsPerByte <= 0 {
		l.MatchStepsPerByte = 16
	}
	if l.MaxMatchSteps <= 0 {
		l.MaxMatchSteps = 500000000
	}
	return l
}

// matchLimit is the work budget for matching one object of n bytes. Objects
// are already bounded by the scan's per-object size cap, so the budget stays
// bounded (16 steps per byte is about 3.2 billion steps for 200 MiB).
func (l Limits) matchLimit(n int) int64 {
	if l.MatchStepsPerByte > 0 && int64(n) > l.MaxMatchSteps/l.MatchStepsPerByte {
		return int64(n) * l.MatchStepsPerByte
	}
	return l.MaxMatchSteps
}

// ContainerInfo distinguishes corruption detection from publisher authentication.
// MD5 checksums alone are not proof of origin. Authenticated is set only when a
// .cvd header carries a valid upstream publisher signature over a digest that
// matches the payload; incrementally patched .cld files never qualify. Remote
// artifacts must still be authenticated by the pinned signed manifest.
type ContainerInfo struct {
	File              string
	Version           string
	IntegrityVerified bool
	Authenticated     bool
}
type Report struct {
	Loaded     int
	Files      []string
	Containers []ContainerInfo
	// Skipped counts, by reason (SkipLevel, SkipIgnoreList, ...), the rules the
	// reference engine does not load either. They are not part of Loaded.
	Skipped map[string]int
	// FileTypes counts the file-type database entries loaded; they type
	// objects and are not rules.
	FileTypes int
	// Unknown counts, by reason (UnknownPCRE, ...), the loaded rules and
	// allowlist entries that need analysis this engine does not perform.
	// Each makes an object incomplete only where it could apply.
	Unknown map[string]int
	// IconData counts the icon-group database entries loaded.
	IconData int
}
type Diagnostic struct {
	File   string
	Line   int
	Reason string
}
type CompileError struct{ Diagnostics []Diagnostic }

func (e *CompileError) Error() string {
	if len(e.Diagnostics) == 0 {
		return "UTMStack EDR signature compilation failed"
	}
	d := e.Diagnostics[0]
	return fmt.Sprintf("UTMStack EDR signature compilation failed at %s:%d: %s (%d diagnostics)", d.File, d.Line, d.Reason, len(e.Diagnostics))
}

var ErrMatchLimit = errors.New("UTMStack EDR signature matching work limit exceeded")

// ErrRuleIncomplete marks a Match result in which one or more rules could not
// be decided while every other rule was evaluated. It is never a clean verdict,
// and it also wraps the specific cause for errors.Is.
var ErrRuleIncomplete = errors.New("UTMStack EDR signature rule evaluation incomplete")

type rule struct {
	name     string
	target   int
	patterns []int
	expr     *expression
	*ruleCold
}

// Most official rules need none of these fields. The shared zero value keeps
// their common record compact; any compiler write first takes an owned copy.
type ruleCold struct {
	size, entry, sections *numberRange
	bytecode              *bytecode.Program
	// byteCompareRefs marks body subsignatures whose last match position a
	// supported byte comparison reads. Those patterns must be counted to the
	// end of the object, even when the logical expression needs only one hit.
	byteCompareRefs uint64
	// unknown is the reason the rule cannot be decided once its expression
	// holds: a descriptor or a bytecode program this engine does not
	// evaluate.
	unknown string
	// handler requests a whole-object scan with this forced type after the
	// expression holds. PDF, GRAPHICS, SCRIPT and ZIP are modeled.
	handler *FileType
	// caps, when set, are the counts the rule's subsignatures need, including
	// what regex triggers read; otherwise the expression decides them.
	caps []int
	// container and intermediates restrict the rule to objects extracted
	// from those types (layers.go).
	container     *FileType
	intermediates []FileType
	// icons are the rule's IconGroup1 and IconGroup2 names ("" is any);
	// dead marks an icon rule no icon-database entry can satisfy.
	icons *[2]string
	dead  bool
	// hook is the kind of a PE hook program (257 unpacker, 259 all PE) and
	// hookOrder its load order; 0 for other rules.
	hook      uint16
	hookOrder int
}

var emptyRuleCold ruleCold

func (r *rule) ownCold() *ruleCold {
	if r.ruleCold == nil || r.ruleCold == &emptyRuleCold {
		r.ruleCold = &ruleCold{}
	}
	return r.ruleCold
}

// A changed source can clone the rule table from a published generation.
// Writes to retained rules must copy their cold record before mutation.
func (r *rule) cloneCold() *ruleCold {
	if r.ruleCold == nil || r.ruleCold == &emptyRuleCold {
		return r.ownCold()
	}
	copyOf := *r.ruleCold
	r.ruleCold = &copyOf
	return r.ruleCold
}

// inLayers applies Container and Intermediates. Container must equal the
// parent's type, normalized layers skipped; ANY means a scanned file itself.
// Intermediates lists the chain outermost first and is compared from its
// last element upward; an ANY element uses no layer.
func (r *rule) inLayers(a *Ancestry) bool {
	if r.container != nil && a.TypeAt(-2) != *r.container {
		return false
	}
	if len(r.intermediates) == 0 {
		return true
	}
	if a.Depth()-1 < len(r.intermediates) {
		return false
	}
	j := -2
	for i := len(r.intermediates) - 1; i >= 0; i-- {
		if t := r.intermediates[i]; t != TypeAny {
			if a.TypeAt(j) != t {
				return false
			}
			j--
		}
	}
	return true
}

// countCap is the count subsignature i needs for this rule.
func (r *rule) countCap(i int) int {
	switch {
	case r.caps != nil:
		return r.caps[i]
	case r.expr != nil:
		return r.expr.countLimit(i)
	}
	return 1
}

type Database struct {
	// lastUnknown is the reason of the last rule loaded as unknown when
	// reached; the audit reads it after each rule.
	lastUnknown string
	// iconEntries are the icon-database entries: group names and hashes.
	iconEntries []iconEntry
	// Certificate rules are immutable after compilation. Authenticode analysis
	// compares embedded certificates with blocklist entries during PE scans;
	// trusting a signer still requires signed-digest and chain verification.
	certRules []certificateRule
	// iconNames are the distinct names of each group field, for the
	// loader's limit of 256 each.
	iconNames [2]map[string]bool
	hashes    *hashTable
	// sourceHashes retain immutable whole-file hash rules by top-level input.
	// An existing changed .hdb/.hsb can replace its source table while sharing
	// the unchanged rule and pattern matcher.
	sourceHashes map[string]*hashTable
	// memberHashes retain each archive hash-list table separately, in archive
	// order under its source table, so one changed member can be replaced.
	memberHashes  map[string]*hashTable
	sourceDigests map[string][sha256.Size]byte
	sourceStats   map[string]*sourceCompileStats
	// containerMembers and memberStats identify each expanded archive member in
	// its original order. They let a changed CVD prove which compiled segment
	// changed without trusting its header or reparsing unchanged members.
	containerMembers  map[string][]containerMemberInfo
	memberStats       map[string]*sourceCompileStats
	sourceOrder       []string
	activeSource      string       // compiler-only; cleared before publication
	activeMember      string       // compiler-only; cleared before publication
	activeHashReserve *hashReserve // compiler-only capacity hint for first member hash insertion
	sections          *hashTable
	sectionSizes      map[int64]bool // section rule sizes; -1 is any size
	allow             *hashTable
	// Archive passwords retain source order within the ZIP-specific and ANY
	// lists. Password bytes are immutable after the generation is published.
	zipPasswords [2][][]byte
	// catalog contains authenticated SHA-1/SHA-256 hashes from signed .cat
	// sources. Sentinel sizes 1 and 2 distinguish CAB from Authenticode PE.
	catalog *hashTable
	// catalogCerts are independently CRB-authenticated certificate paths
	// carried by catalogs, including those whose member signature fails.
	// A catalog may carry an intermediate needed by a later signed PE.
	catalogCerts    []catalogCertificate
	pendingCatalogs []source // compiler-only, after all CRB trust rules load
	rules           []rule
	// patterns holds each distinct compiled subsignature once; rules refer to
	// them by id, so identical subsignatures share one count per object.
	patterns []bodyPattern
	// patternKeys are the compiler's original 128-bit deduplication keys in
	// pattern-id order. Retaining this compact array lets a changed source
	// reuse unchanged compiled patterns without keeping the construction map.
	patternKeys [][16]byte
	// retiredPatterns marks slots whose compiled bodies were released after a
	// source replacement. A later generation can recompile into a retired
	// slot, but must not deduplicate against its empty body.
	retiredPatterns bitset
	// Compiler-only: reused retired slots need fresh anchors even though their
	// IDs are below the previous generation's pattern-count boundary.
	freshReusedPatterns bitset
	reusePatternCursor  int
	// patternIDs finds shared patterns while compiling; nil once frozen.
	patternIDs map[[16]byte]int
	// expressionIDs shares immutable logical trees during compilation. The
	// lookup itself is discarded before publishing the generation.
	expressionIDs map[expressionKey]*expression
	// index finds every pattern anchor in one pass over an object.
	index anchorIndex
	// anchorGroups preserve immutable literals and pattern membership in
	// construction order so a changed source can rebuild an active index
	// without reparsing every unchanged signature.
	anchorGroups []*anchorGroup
	unanchored   []int
	// ruleIndex finds the rules an object's candidate patterns can make true.
	ruleIndex ruleIndex
	// bmPrefixes records nonzero native BM load-balancing prefixes by rule ID.
	// It is rebuilt in source order with the immutable rule index.
	bmPrefixes map[int]uint32
	// fixedAC is present only for roots whose AC registrations can all use
	// the fixed-width anchor model. Other roots retain the general matcher.
	fixedAC  [maxTargets]map[int]fixedACCandidate
	nativeAC [maxTargets]*nativeACRoot
	// legacyMetadataRules selects the deprecated target-13 scan of serialized
	// root JSON. It is rebuilt with the rule index after incremental refreshes.
	legacyMetadataRules bool
	limits              Limits
	report              Report
	version             string
	phishing            phishingRules
	bytecodeBytes       uint64
	// bytecodeSources identifies each loaded CBC contribution so an updated
	// member can replace its program without retaining old global hooks.
	bytecodeSources map[string]bytecodeSourceState
	// level is the functionality level rules are selected against; zero means
	// FunctionalityLevel. Only gap-accounting probes use another value.
	level  int
	ignore *ignoreList
	dconf  DynamicConfig
	// hookCount numbers executable hook programs in load order.
	hookCount int
	// bareHooks run at their parser hook point without a logical trigger.
	bareHooks [maxTargets][]rule
	// pdfHooks are the PDF hook programs, in load order (pdfhook.go).
	pdfHooks []*bytecode.Program
	// startupHooks are preparation programs in feed order. ClamAV runs only
	// the first loaded one, before any scanner callback is activated.
	startupHooks []*bytecode.Program
	// bytecodeDisabled is the first startup program's request to leave every
	// ClamBC callback inactive for this compiled generation.
	bytecodeDisabled bool
	// preclassHooks run once after the root object's representations have been
	// scanned, in feed order. Triggered entries are flagged by root matching.
	preclassHooks   []preclassHook
	unknownPreclass bool
	ftypes          fileTypes
	// unknownHooks marks, per target, bytecode hooks this engine cannot run
	// that the reference runs on every object of that target.
	unknownHooks [maxTargets]string
	// textChunk and textFileMode describe how normalized text is matched
	// (TextNormalization).
	textChunk    int
	textFileMode bool
	// containers are the container metadata rules (cdb.go).
	containers []cdbRule
}

func newDatabase(limits Limits, level int) *Database {
	return &Database{hashes: newWholeHashIndex(), sections: newDetectionHashTable(), allow: newHashTable(), catalog: newHashTable(), limits: limits,
		level: level, patternIDs: map[[16]byte]int{}, expressionIDs: map[expressionKey]*expression{}, dconf: DefaultDynamicConfig()}
}

func (d *Database) functionality() int {
	if d.level == 0 {
		return FunctionalityLevel
	}
	return d.level
}

func (d *Database) Version() string { return d.version }
func (d *Database) Count() int      { return d.report.Loaded }

// DetectionCount excludes false-positive allow hashes: an allowlist alone is
// not a functioning malware database.
func (d *Database) DetectionCount() int {
	n := d.hashes.len() + d.sections.len() + len(d.rules) + d.phishing.count() + len(d.containers)
	for _, hooks := range d.bareHooks {
		n += len(hooks)
	}
	for _, hook := range d.preclassHooks {
		if hook.bare {
			n++
		}
	}
	return n
}

// HasWholeHashWildcard reports whether a compiled whole-file rule covers a
// digest at every object size. The feed publisher uses this when removing
// redundant bare indicators; a size-specific official rule does not cover a
// size-agnostic indicator and must not cause that indicator to be discarded.
func (d *Database) HasWholeHashWildcard(digest []byte) bool {
	if d == nil || d.hashes == nil {
		return false
	}
	return d.hashes.lookup(string(digest), -1) != ""
}

// Allowed reports an explicit whole-object false-positive exemption. The
// container walker must honor it before recursively scanning the object's
// members, otherwise a known-good outer file could still be quarantined.
func (d *Database) Allowed(ctx context.Context, data []byte) (bool, error) {
	if d.allow.len() == 0 {
		return false, ctx.Err()
	}
	digests, err := hashesFor(ctx, data)
	if err != nil {
		return false, err
	}
	for _, digest := range digests {
		if d.allow.lookup(digest, int64(len(data))) != "" {
			return true, nil
		}
	}
	return false, nil
}
func (d *Database) Report() Report {
	r := d.report
	r.Files = append([]string(nil), r.Files...)
	r.Containers = append([]ContainerInfo(nil), r.Containers...)
	if r.Skipped != nil {
		r.Skipped = maps.Clone(r.Skipped)
	}
	if r.Unknown != nil {
		r.Unknown = maps.Clone(r.Unknown)
	}
	return r
}

// DynamicConfig returns the feature masks the generation's dynamic
// configuration selected.
func (d *Database) DynamicConfig() DynamicConfig { return d.dconf }

// ForEachZIPPassword visits ZIP-specific passwords first, then passwords for
// any container. Within each group the reference's prepend-only loader tries
// later database lines first. The borrowed bytes must not be modified or
// retained. Returning true stops iteration.
func (d *Database) ForEachZIPPassword(visit func([]byte) bool) {
	if d == nil || d.dconf.Archive&dconfArchivePasswd == 0 {
		return
	}
	for _, group := range d.zipPasswords {
		for i := len(group) - 1; i >= 0; i-- {
			if visit(group[i]) {
				return
			}
		}
	}
}

func (d *Database) skipped(reason string, n int) {
	if d.report.Skipped == nil {
		d.report.Skipped = map[string]int{}
	}
	d.report.Skipped[reason] += n
	if stats := d.sourceStats[d.activeSource]; stats != nil {
		stats.skipped[reason] += n
	}
	if stats := d.memberStats[d.activeMember]; stats != nil {
		stats.skipped[reason] += n
	}
}

// Load reads signature files in one directory. Subdirectories and symlinks are
// rejected. Unrelated configuration, lock, log and database-updater state files
// are ignored; unknown signature formats fail compilation.
func Load(dir string) (*Database, error) { return LoadWithLimits(dir, Limits{}) }
func LoadWithLimits(dir string, limits Limits) (*Database, error) {
	return loadWithOptions(dir, limits, compileOptions{})
}

// LoadWithPrior validates all source bytes and may share immutable matcher
// components when only existing hash-list, top-level NDB, or NDB/LDB/HDB/HSB/CBC
// members of one validated CVD/CLD source changed.
func LoadWithPrior(dir string, limits Limits, prior *Database) (*Database, error) {
	return loadWithOptions(dir, limits, compileOptions{prior: prior})
}

func loadWithOptions(dir string, limits Limits, options compileOptions) (*Database, error) {
	limits = limits.defaults()
	if options.prior != nil {
		if refreshed, handled, err := tryIncrementalDirectoryRefresh(dir, limits, options); handled {
			return refreshed, err
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make(map[string][]byte)
	var total int64
	for _, entry := range entries {
		name := entry.Name()
		// An interrupted activation can leave an immutable generation directory
		// before CURRENT exists. It must not break the loose local fallback DB.
		if entry.IsDir() && name == "generations" {
			continue
		}
		if metadataFile(name) {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			return nil, fmt.Errorf("UTMStack EDR signature path is not a regular file: %s", name)
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("UTMStack EDR invalid signature file: %s", name)
		}
		if info.Size() > limits.MaxFileBytes || info.Size() > limits.MaxTotalBytes-total {
			return nil, fmt.Errorf("UTMStack EDR signature database size limit exceeded")
		}
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		b, readErr := readSizedFile(f, info.Size())
		closeErr := f.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if int64(len(b)) > limits.MaxFileBytes {
			return nil, fmt.Errorf("UTMStack EDR signature file size limit exceeded")
		}
		total += int64(len(b))
		if total > limits.MaxTotalBytes {
			return nil, fmt.Errorf("UTMStack EDR signature database size limit exceeded")
		}
		files[name] = b
		if len(files) > limits.MaxFiles {
			return nil, fmt.Errorf("UTMStack EDR signature file count limit exceeded")
		}
	}
	return compileWithLimits(files, limits, options)
}

// readSizedFile uses the size already checked against the database budgets.
// A changed file is rejected rather than silently truncating a newer version.
func readSizedFile(r io.Reader, size int64) ([]byte, error) {
	if size < 0 || uint64(size) >= uint64(^uint(0)>>1) {
		return nil, errors.New("invalid signature file size")
	}
	b := make([]byte, int(size))
	if _, err := io.ReadFull(r, b); err != nil {
		return nil, fmt.Errorf("signature file changed or truncated: %w", err)
	}
	var extra [1]byte
	if n, err := io.ReadFull(r, extra[:]); n != 0 || err != io.EOF {
		return nil, errors.New("signature file changed while reading")
	}
	return b, nil
}
func metadataFile(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json", ".dat", ".log", ".lock", ".tmp", ".conf":
		return true
	}
	return false
}

// Compile atomically compiles all files. The database is nil if any signature
// cannot be represented exactly by the supported formats.
func Compile(files map[string][]byte) (*Database, error) { return CompileWithLimits(files, Limits{}) }
func CompileWithLimits(files map[string][]byte, limits Limits) (*Database, error) {
	return compileWithLimits(files, limits, compileOptions{})
}

// ErrNoDetections means that a database contains no malware detection rules.
// False-positive allowlists, file types and configuration cannot make a feed
// usable by themselves.
var ErrNoDetections = errors.New("UTMStack EDR signature database contains no detection signatures")

// ContentVersion validates source names and byte limits, and identifies their
// exact bytes in canonical filename order. It does not parse rules, authenticate
// a container or construct a matcher. Callers must compile untrusted new content
// before using it; this is suitable for checking a stored copy against content
// that was already compiled and authenticated.
func ContentVersion(files map[string][]byte, limits Limits) (string, error) {
	version, _, _, err := sourceIdentity(files, limits.defaults())
	return version, err
}

// DirectoryContentVersion identifies the exact loose signature files without
// compiling them or retaining their bytes. It applies the same file selection,
// regular-file and size rules as LoadWithLimits. Callers may reuse only a
// previously compiled database with this exact content version; new bytes
// must still pass the full compiler before they can serve scans.
func DirectoryContentVersion(dir string, limits Limits) (string, error) {
	limits = limits.defaults()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	identity := sha256.New()
	var total int64
	count := 0
	buf := make([]byte, 64<<10)
	for _, entry := range entries { // os.ReadDir returns lexically sorted names.
		name := entry.Name()
		if (entry.IsDir() && name == "generations") || metadataFile(name) {
			continue
		}
		if err := ValidateFileName(name); err != nil {
			return "", err
		}
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			return "", fmt.Errorf("UTMStack EDR signature path is not a regular file: %s", name)
		}
		info, err := entry.Info()
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("UTMStack EDR invalid signature file: %s", name)
		}
		if info.Size() < 0 || info.Size() > limits.MaxFileBytes || info.Size() > limits.MaxTotalBytes-total {
			return "", errors.New("UTMStack EDR signature database size limit exceeded")
		}
		count++
		if count > limits.MaxFiles {
			return "", errors.New("UTMStack EDR signature file count limit exceeded")
		}
		total += info.Size()
		if _, err := fmt.Fprintf(identity, "%d:%s:%d:", len(name), name, info.Size()); err != nil {
			return "", err
		}
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			return "", err
		}
		n, readErr := io.CopyBuffer(identity, io.LimitReader(f, info.Size()), buf)
		if readErr == nil && n != info.Size() {
			readErr = errors.New("signature file changed or truncated while reading")
		}
		if readErr == nil {
			var extra [1]byte
			_, readErr = io.ReadFull(f, extra[:])
			if readErr == io.EOF {
				readErr = nil
			} else if readErr == nil {
				readErr = errors.New("signature file changed while reading")
			}
		}
		closeErr := f.Close()
		if readErr != nil {
			return "", readErr
		}
		if closeErr != nil {
			return "", closeErr
		}
	}
	return "go-v1-" + hex.EncodeToString(identity.Sum(nil)), nil
}

func sourceIdentity(files map[string][]byte, limits Limits) (string, []string, int64, error) {
	names, total, err := sourceNames(files, limits)
	if err != nil {
		return "", nil, 0, err
	}
	identity := sha256.New()
	for _, name := range names {
		fmt.Fprintf(identity, "%d:%s:%d:", len(name), name, len(files[name]))
		identity.Write(files[name])
	}
	return "go-v1-" + hex.EncodeToString(identity.Sum(nil)), names, total, nil
}

func sourceNames(files map[string][]byte, limits Limits) ([]string, int64, error) {
	limits = limits.defaults()
	if len(files) > limits.MaxFiles {
		return nil, 0, fmt.Errorf("UTMStack EDR signature file count limit exceeded")
	}
	names := make([]string, 0, len(files))
	var total int64
	for name, b := range files {
		if err := ValidateFileName(name); err != nil {
			return nil, 0, err
		}
		names = append(names, name)
		total += int64(len(b))
		if int64(len(b)) > limits.MaxFileBytes || total > limits.MaxTotalBytes {
			return nil, 0, fmt.Errorf("UTMStack EDR signature database size limit exceeded")
		}
	}
	sort.Strings(names)
	return names, total, nil
}

// CompileWithLocalFiles compiles base feed files and administrator files as one
// database, preserving global ignore/configuration selection and source order.
// A valid base must contain detections independently of the local files. When
// locals fail, db is the base-only matcher and localErr reports their rejection;
// a non-nil fatalErr means that the base itself cannot serve and db is nil.
// The normal path parses and constructs one matcher. A changed local selection
// is conservatively validated with a separate base-only compilation first.
func CompileWithLocalFiles(base, local map[string][]byte, limits Limits) (db *Database, localErr, fatalErr error) {
	return compileWithLocalFiles(base, local, limits, compileOptions{})
}

// CompileWithPreparedLocalFiles uses decoded native hash-list sources while
// retaining the exact durable source identity and all normal selection rules.
func CompileWithPreparedLocalFiles(base, local map[string][]byte, prepared map[string]*PreparedHashList, limits Limits) (db *Database, localErr, fatalErr error) {
	return CompileWithPreparedLocalFilesAndPrior(base, local, prepared, limits, nil)
}

// CompileWithPreparedLocalFilesAndPrior attempts validated source reuse while
// preserving the usual base/local selection and fallback policy.
func CompileWithPreparedLocalFilesAndPrior(base, local map[string][]byte, prepared map[string]*PreparedHashList, limits Limits, prior *Database) (db *Database, localErr, fatalErr error) {
	for name, p := range prepared {
		if _, exists := base[name]; !exists || p == nil {
			return nil, nil, errors.New("prepared hash list has no matching base source")
		}
	}
	return compileWithLocalFiles(base, local, limits, compileOptions{prepared: prepared, prior: prior})
}

// compileOptions binds base provenance and optional invocation-local construction
// observations used by regression tests. No package-global hook can race with reloads.
type compileOptions struct {
	baseSources map[string]bool
	prepared    map[string]*PreparedHashList
	prior       *Database
	onParse     func()
	onFinalize  func()
}

var errSeparateBaseSelection = errors.New("local files change base signature selection")

func compileWithLocalFiles(base, local map[string][]byte, limits Limits, options compileOptions) (db *Database, localErr, fatalErr error) {
	if _, _, err := sourceNames(base, limits); err != nil {
		return nil, nil, err
	}
	baseOnly := func() (*Database, error) {
		db, err := compileWithLimits(base, limits, options)
		if err == nil && db.DetectionCount() == 0 {
			err = ErrNoDetections
		}
		return db, err
	}
	if len(local) == 0 {
		db, err := baseOnly()
		if err != nil {
			return nil, nil, err
		}
		return db, nil, nil
	}
	if len(base)+len(local) > limits.defaults().MaxFiles {
		baseDB, err := baseOnly()
		if err != nil {
			return nil, nil, err
		}
		return baseDB, errors.New("UTMStack EDR signature file count limit exceeded"), nil
	}
	merged := make(map[string][]byte, len(base)+len(local))
	baseNames := make(map[string]bool, len(base))
	for name, b := range base {
		merged[name], baseNames[name] = b, true
	}
	for name, b := range local {
		if _, exists := merged[name]; exists {
			localErr = fmt.Errorf("%s has the same name as a feed generation artifact", name)
			break
		}
		merged[name] = b
	}
	if localErr == nil {
		combinedOptions := options
		combinedOptions.baseSources = baseNames
		db, localErr = compileWithLimits(merged, limits, combinedOptions)
		if localErr == nil {
			return db, nil, nil
		}
	}
	baseDB, err := baseOnly()
	if err != nil {
		return nil, nil, err
	}
	if errors.Is(localErr, errSeparateBaseSelection) {
		// Ignore/dconf changes can hide an otherwise invalid feed rule. Validate
		// the feed on its own before allowing that policy to affect the matcher.
		db, localErr = compileWithLimits(merged, limits, options)
		if localErr == nil {
			return db, nil, nil
		}
	}
	return baseDB, localErr, nil
}

func sameSelection(a, b *Database) bool {
	if a.dconf != b.dconf {
		return false
	}
	var ae, be map[string]*[md5.Size]byte
	if a.ignore != nil {
		ae = a.ignore.entries
	}
	if b.ignore != nil {
		be = b.ignore.entries
	}
	return maps.EqualFunc(ae, be, func(x, y *[md5.Size]byte) bool {
		return x == nil && y == nil || x != nil && y != nil && *x == *y
	})
}

func compileWithLimits(files map[string][]byte, limits Limits, options compileOptions) (*Database, error) {
	limits = limits.defaults()
	if options.prior != nil {
		if reused, handled, err := tryHashSourceRefresh(files, limits, options); handled {
			return reused, err
		}
		if reused, handled, err := tryRuleSourceRefresh(files, limits, options); handled {
			return reused, err
		}
		if reused, handled, err := tryContainerSourceRefresh(files, limits, options); handled {
			return reused, err
		}
	}
	version, names, total, err := sourceIdentity(files, limits)
	if err != nil {
		return nil, err
	}
	d := newDatabase(limits, 0)
	d.sourceHashes = make(map[string]*hashTable, len(names))
	d.memberHashes = make(map[string]*hashTable)
	d.sourceDigests = make(map[string][sha256.Size]byte, len(names))
	d.sourceStats = make(map[string]*sourceCompileStats, len(names))
	d.containerMembers = make(map[string][]containerMemberInfo)
	d.memberStats = make(map[string]*sourceCompileStats)
	d.sourceOrder = loadOrder(names)
	for _, name := range names {
		d.sourceDigests[name] = sha256.Sum256(files[name])
		d.sourceStats[name] = &sourceCompileStats{skipped: make(map[string]int), unknown: make(map[string]int)}
	}
	errs := &CompileError{}
	addError := func(file string, line int, err error) {
		if len(errs.Diagnostics) < 100 {
			errs.Diagnostics = append(errs.Diagnostics, Diagnostic{file, line, err.Error()})
		}
	}
	budget := archiveBudget{limits: limits, total: total, files: len(files)}
	sources := d.expandForCompilation(d.sourceOrder, files, &budget, addError)
	// Ignore lists and dynamic configuration anywhere in the generation apply
	// to every rule, so they are read before any rule is compiled.
	d.loadSelection(sources, addError)
	// Large authenticated text members give a bounded upper estimate of new
	// patterns. Reserve once before compilation so growing the million-entry
	// pattern arrays does not repeatedly overlap old and new backing storage.
	patternCapacity := 0
	patternReservationDisabled := false
	for _, members := range d.containerMembers {
		for _, member := range members {
			patternCapacity += member.patternCandidates
			if patternCapacity > 2_000_000 {
				patternReservationDisabled = true
				break
			}
		}
		if patternReservationDisabled {
			break
		}
	}
	if !patternReservationDisabled && patternCapacity >= 128_000 {
		d.patterns = make([]bodyPattern, 0, patternCapacity)
		d.patternKeys = make([][16]byte, 0, patternCapacity)
	}
	if options.baseSources != nil {
		localSelection := false
		for _, src := range sources {
			name, _, _ := strings.Cut(src.name, "/")
			if !options.baseSources[name] && (src.ext == ".ign" || src.ext == ".ign2" || src.ext == ".cfg") {
				localSelection = true
				break
			}
		}
		if localSelection {
			var baseSources []source
			for _, src := range sources {
				name, _, _ := strings.Cut(src.name, "/")
				if options.baseSources[name] {
					baseSources = append(baseSources, src)
				}
			}
			baseSelection := newDatabase(limits, 0)
			baseSelection.loadSelection(baseSources, addError)
			if !sameSelection(baseSelection, d) {
				return nil, errSeparateBaseSelection
			}
		}
	}
	if options.onParse != nil {
		options.onParse()
	}
	for i := range sources {
		if sources[i].nested {
			sources[i].data = nil // selection bytes have already been consumed
		}
	}
	baseDetects := false
	compileOne := func(src source, stream io.Reader) {
		name, _, _ := strings.Cut(src.name, "/")
		d.activeSource = name
		if src.nested {
			d.activeMember = src.name
		} else {
			d.activeMember = ""
		}
		beforeDetections, beforeLoaded, beforeRules := d.DetectionCount(), d.report.Loaded, len(d.rules)
		if p := options.prepared[src.name]; p != nil {
			d.compilePreparedHashList(src, p, addError)
		} else if stream != nil {
			d.compileTextSource(src, stream, addError)
		} else {
			d.compileSource(src, addError)
		}
		stats := d.sourceStats[name]
		if !stats.ruleSeen {
			stats.ruleStart, stats.ruleSeen = beforeRules, true
		}
		stats.ruleEnd = len(d.rules)
		stats.loaded += d.report.Loaded - beforeLoaded
		if d.DetectionCount() > beforeDetections || src.ext == ".pdb" && d.report.Loaded > beforeLoaded {
			stats.detections += max(1, d.DetectionCount()-beforeDetections)
		}
		if src.nested {
			memberStats := d.memberStats[src.name]
			memberStats.ruleStart, memberStats.ruleEnd, memberStats.ruleSeen = beforeRules, len(d.rules), true
			memberStats.loaded = d.report.Loaded - beforeLoaded
			if d.DetectionCount() > beforeDetections || src.ext == ".pdb" && d.report.Loaded > beforeLoaded {
				memberStats.detections = max(1, d.DetectionCount()-beforeDetections)
			}
		}
		if options.baseSources[name] && (d.DetectionCount() > beforeDetections ||
			// A protected-domain host can already have been inserted by a
			// local source, but is still a detection in the base by itself.
			src.ext == ".pdb" && d.report.Loaded > beforeLoaded) {
			baseDetects = true
		}
	}
	replayBudget := archiveBudget{limits: limits, total: total, files: len(files)}
	for i := 0; i < len(sources); {
		if len(errs.Diagnostics) >= 100 {
			break
		}
		src := sources[i]
		if !src.nested {
			compileOne(src, nil)
			sources[i].data = nil
			i++
			continue
		}
		name, _, _ := strings.Cut(src.name, "/")
		end := i + 1
		for end < len(sources) && sources[end].nested && strings.HasPrefix(sources[end].name, name+"/") {
			end++
		}
		if err := d.replayContainer(name, files[name], &replayBudget, sources[i:end], d.containerMembers[name], options.prepared, compileOne); err != nil {
			addError(name, 0, err)
		}
		i = end
	}
	files = nil
	for _, src := range d.pendingCatalogs {
		name, _, _ := strings.Cut(src.name, "/")
		d.activeSource, d.activeMember = name, ""
		if src.nested {
			d.activeMember = src.name
		}
		before := d.report.Loaded
		d.compileCatalog(src, addError)
		d.sourceStats[name].loaded += d.report.Loaded - before
		if src.nested {
			d.memberStats[src.name].loaded += d.report.Loaded - before
		}
	}
	d.pendingCatalogs = nil
	if len(errs.Diagnostics) > 0 {
		return nil, errs
	}
	if err := d.prepareStartup(); err != nil {
		return nil, err
	}
	d.activeSource = ""
	d.activeMember = ""
	d.activeHashReserve = nil
	if d.report.Loaded == 0 {
		return nil, errors.New("UTMStack EDR signature database contains no supported signatures")
	}
	if options.baseSources != nil && !baseDetects {
		return nil, ErrNoDetections
	}
	if options.onFinalize != nil {
		options.onFinalize()
	}
	d.useBuiltInFileTypes()
	d.resolveIconGroups()
	d.version = version
	// Force later appends to allocate without cloning these large slices during
	// finalization. Their growth slack is small on the official feed; copying
	// both here makes roughly a quarter GiB of transient allocations.
	d.patterns = d.patterns[:len(d.patterns):len(d.patterns)]
	d.rules = d.rules[:len(d.rules):len(d.rules)]
	for _, h := range []*hashTable{d.hashes, d.sections, d.allow, d.catalog} {
		if err := h.freeze(); err != nil {
			return nil, err
		}
	}
	for _, h := range d.sourceHashes {
		if err := h.freeze(); err != nil {
			return nil, err
		}
	}
	if err := d.buildAnchors(); err != nil {
		return nil, err
	}
	d.buildRuleIndex()
	d.measureTextRoot()
	return d, nil
}

// TextNormalization says how the reference matches normalized text against
// root 7 and root 0 with this database. chunk is the input length after which
// the normalizer's white-space state resets: 131072 plus the longest root-7
// pattern part. fileMode is true when root 7 holds a pattern with a relative
// offset or a rule with bytecode: the reference then matches the normalized
// text as a file (hashes checked, EOF offsets from its end); otherwise it
// matches it as a buffer.
func (d *Database) TextNormalization() (chunk int, fileMode bool) {
	return d.textChunk, d.textFileMode
}

func (d *Database) measureTextRoot() {
	longest := 0
	for i := range d.rules {
		r := &d.rules[i]
		if r.target != 7 {
			continue
		}
		if !d.bytecodeDisabled && (r.bytecode != nil || r.unknown == UnknownBytecode) {
			d.textFileMode = true
		}
		for _, id := range r.patterns {
			p := &d.patterns[id]
			if p.holder != nil {
				continue // regex subsignatures are not matcher parts
			}
			d.textFileMode = d.textFileMode || p.offset.relative()
			longest = max(longest, p.longestPart())
		}
	}
	d.textChunk = scanWindow + longest
}

// source is one signature file in load order; container members are named
// "container/member".
type source struct {
	name, ext string
	data      []byte
	nested    bool
}

type containerMemberInfo struct {
	name              string
	digest            [sha256.Size]byte
	size              int64
	reserve           hashReserve // compiler-only capacity hint; zero on refreshed members
	patternCandidates int         // bounded preflight upper estimate for pattern storage
}

// loadOrder is the reference's database-directory order: ignore lists, then
// daily.cld, daily.cvd, local.gdb, daily.cfg and .crb files, then everything
// else. Files of equal priority keep name order.
func loadOrder(sorted []string) []string {
	priority := func(name string) int {
		lower := strings.ToLower(name)
		switch {
		case strings.HasSuffix(lower, ".ign") || strings.HasSuffix(lower, ".ign2"):
			return 1
		case name == "daily.cld":
			return 2
		case name == "daily.cvd":
			return 3
		case name == "local.gdb":
			return 4
		case name == "daily.cfg":
			return 5
		case strings.HasSuffix(lower, ".crb"):
			return 6
		}
		return 7
	}
	out := append([]string(nil), sorted...)
	sort.SliceStable(out, func(i, j int) bool { return priority(out[i]) < priority(out[j]) })
	return out
}

// expand lists every file in load order, unpacking containers into their
// members in archive order.
func (d *Database) expand(order []string, files map[string][]byte, budget *archiveBudget, addError func(string, int, error)) []source {
	if d.containerMembers == nil {
		d.containerMembers = make(map[string][]containerMemberInfo)
	}
	if d.memberStats == nil {
		d.memberStats = make(map[string]*sourceCompileStats)
	}
	var out []source
	for _, name := range order {
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".cvd" && ext != ".cld" {
			out = append(out, source{name, ext, files[name], false})
			continue
		}
		members, info, err := unpackContainer(name, files[name], budget)
		if err != nil {
			addError(name, 0, err)
			continue
		}
		d.report.Containers = append(d.report.Containers, info)
		for _, m := range members {
			full := name + "/" + m.name
			d.containerMembers[name] = append(d.containerMembers[name], containerMemberInfo{
				name: m.name, digest: sha256.Sum256(m.data), size: int64(len(m.data)),
			})
			d.memberStats[full] = &sourceCompileStats{skipped: make(map[string]int), unknown: make(map[string]int)}
			mext := strings.ToLower(filepath.Ext(m.name))
			if mext == ".cvd" || mext == ".cld" {
				addError(full, 0, errors.New("nested signature containers are unsupported"))
				continue
			}
			out = append(out, source{full, mext, m.data, true})
		}
	}
	return out
}

// expandForCompilation keeps only selection members from each preflighted
// archive. The full compiler replays these immutable containers one member
// at a time after global selection has been applied.
func (d *Database) expandForCompilation(order []string, files map[string][]byte, budget *archiveBudget, addError func(string, int, error)) []source {
	if d.containerMembers == nil {
		d.containerMembers = make(map[string][]containerMemberInfo)
	}
	if d.memberStats == nil {
		d.memberStats = make(map[string]*sourceCompileStats)
	}
	var out []source
	for _, name := range order {
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".cvd" && ext != ".cld" {
			out = append(out, source{name, ext, files[name], false})
			continue
		}
		members, infos, info, err := preflightContainer(name, files[name], budget)
		if err != nil {
			addError(name, 0, err)
			continue
		}
		d.report.Containers = append(d.report.Containers, info)
		d.containerMembers[name] = infos
		for _, src := range members {
			d.memberStats[src.name] = &sourceCompileStats{skipped: make(map[string]int), unknown: make(map[string]int)}
			if src.ext == ".cvd" || src.ext == ".cld" {
				addError(src.name, 0, errors.New("nested signature containers are unsupported"))
			}
		}
		out = append(out, members...)
	}
	return out
}

// loadSelection reads every ignore list and dynamic-configuration file, in
// load order, before any rule is compiled. A configuration this engine cannot
// honor is a diagnostic.
func (d *Database) loadSelection(sources []source, addError func(string, int, error)) {
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
			err := eachLine(src.data, d.limits.MaxLineBytes, func(number int, line string) {
				if _, _, err := state.apply(src.name, number, line, d.functionality()); err != nil {
					addError(src.name, number, err)
				}
			})
			if err != nil {
				addError(src.name, 0, err)
			}
		}
	}
	for _, diagnostic := range state.unsupported() {
		addError(diagnostic.File, diagnostic.Line, errors.New(diagnostic.Reason))
	}
	d.dconf = state.config
}

// eachLine calls fn with every line number and line, line endings removed.
func eachLine(b []byte, maxLine int, fn func(int, string)) error {
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Buffer(make([]byte, 4096), maxLine)
	number := 0
	for scanner.Scan() {
		number++
		fn(number, strings.TrimRight(scanner.Text(), "\r\n"))
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("signature line limit or read error: %w", err)
	}
	return nil
}

// puaFormats hold potentially-unwanted-application rules, which the reference
// loads only when that category is enabled; this engine never enables it.
var puaFormats = map[string]bool{".hdu": true, ".hsu": true, ".mdu": true, ".msu": true, ".ndu": true, ".ldu": true}

// ruleLines counts the lines the line loop would treat as rules.
func ruleLines(b []byte) int {
	n := 0
	for _, line := range bytes.Split(b, []byte("\n")) {
		if s := bytes.TrimSpace(line); len(s) > 0 && s[0] != '#' {
			n++
		}
	}
	return n
}

// dconfDisabled reports whether the dynamic configuration makes the reference
// skip a whole database format.
func (d *Database) dconfDisabled(ext string) bool {
	switch ext {
	case ".pdb", ".wdb", ".gdb":
		return d.dconf.Phishing&dconfPhishingEngine == 0
	case ".cbc":
		return d.dconf.Bytecode&dconfBytecodeEngines == 0
	case ".crb":
		return d.dconf.PE&dconfPECerts == 0
	case ".cat":
		return d.dconf.PE&dconfPECerts == 0 || d.dconf.PE&dconfPECatalog == 0
	}
	return false
}

func (d *Database) compileSource(src source, addError func(string, int, error)) {
	switch src.ext {
	case ".ign", ".ign2", ".cfg":
		d.report.Files = append(d.report.Files, src.name) // read by loadSelection
		return
	case ".info":
		if src.nested {
			return
		}
	case "":
		if base := filepath.Base(src.name); src.nested && (base == "COPYING" || base == "LICENSE") {
			return
		}
	}
	if puaFormats[src.ext] {
		d.skipped(SkipPUA, ruleLines(src.data))
		return
	}
	if d.dconfDisabled(src.ext) {
		n := 1
		if src.ext != ".cbc" && src.ext != ".cat" {
			n = ruleLines(src.data)
		}
		d.skipped(SkipDynamicConfig, n)
		return
	}
	switch src.ext {
	case ".cbc":
		d.compileBytecode(src, addError)
		return
	case ".cat":
		d.pendingCatalogs = append(d.pendingCatalogs, src)
		return
	case ".ftm":
		d.compileFileTypes(src, addError)
		return
	case ".pwdb":
		before := d.report.Loaded
		d.compileTextSource(src, bytes.NewReader(src.data), addError)
		if d.report.Loaded == before {
			addError(src.name, 0, errors.New("empty archive password database"))
		}
		return
	case ".hdb", ".hsb", ".mdb", ".msb", ".fp", ".sfp", ".ndb", ".ldb", ".db", ".pdb", ".wdb", ".gdb", ".idb", ".crb", ".cdb":
	default:
		addError(src.name, 0, fmt.Errorf("unsupported signature database format %q", src.ext))
		return
	}
	d.compileTextSource(src, bytes.NewReader(src.data), addError)
}

// compileTextSource accepts a bounded tar member reader as well as a loose
// file. No input line bytes survive this call unless the parser copies them
// into the immutable matcher.
func (d *Database) compileTextSource(src source, r io.Reader, addError func(string, int, error)) {
	d.report.Files = append(d.report.Files, src.name)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), d.limits.MaxLineBytes)
	hashFormat := src.ext == ".hdb" || src.ext == ".hsb" || src.ext == ".mdb" || src.ext == ".msb" || src.ext == ".fp" || src.ext == ".sfp"
	var certificateState certificateSourceState
	line := 0
	for scanner.Scan() {
		line++
		if hashFormat {
			raw := scanner.Bytes()
			s := bytes.TrimSpace(raw)
			if len(s) == 0 || s[0] == '#' {
				continue
			}
			if d.report.Loaded >= d.limits.MaxSignatures {
				addError(src.name, line, errors.New("signature count limit exceeded"))
				return
			}
			err := d.addHashBytes(src.ext, s, bytes.TrimRight(raw, "\r\n"))
			if reason, ok := skipReason(err); ok {
				d.skipped(reason, 1)
			} else if err != nil {
				addError(src.name, line, err)
			} else {
				d.report.Loaded++
			}
			continue
		}
		raw := scanner.Text()
		s := strings.TrimSpace(raw)
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		if d.report.Loaded >= d.limits.MaxSignatures {
			addError(src.name, line, errors.New("signature count limit exceeded"))
			return
		}
		var err error
		if src.ext == ".crb" {
			err = d.addCertificateSourceLine(s, &certificateState)
		} else {
			err = d.addLine(src.ext, s, strings.TrimRight(raw, "\r\n"))
		}
		if reason, ok := skipReason(err); ok {
			d.skipped(reason, 1)
		} else if err != nil {
			addError(src.name, line, err)
		} else if src.ext != ".idb" { // icon data are not rules
			d.report.Loaded++
		}
	}
	if err := scanner.Err(); err != nil {
		addError(src.name, line, fmt.Errorf("signature line limit or read error: %w", err))
	}
}

// compileFileTypes loads a file-type database. Its entries type objects and
// are not counted as rules; a file without any entry line is malformed.
func (d *Database) compileFileTypes(src source, addError func(string, int, error)) {
	d.report.Files = append(d.report.Files, src.name)
	entries := 0
	err := eachLine(src.data, d.limits.MaxLineBytes, func(number int, line string) {
		s := strings.TrimSpace(line)
		if s == "" || strings.HasPrefix(s, "#") {
			return
		}
		entries++
		if err := d.addFileTypeLine(s); err != nil {
			if reason, ok := skipReason(err); ok {
				d.skipped(reason, 1)
			} else {
				addError(src.name, number, err)
			}
		}
	})
	if err != nil {
		addError(src.name, 0, err)
	} else if entries == 0 {
		addError(src.name, 0, errors.New("empty file-type database"))
	}
}

// bytecodeTrigger returns a bytecode file's logical trigger line when its
// second line is a logical signature.
func bytecodeTrigger(b []byte) (string, bool) {
	lines := bytes.SplitN(b, []byte("\n"), 3)
	if len(lines) < 2 {
		return "", false
	}
	line := strings.TrimRight(string(lines[1]), "\r")
	if !strings.Contains(line, ";") {
		return "", false
	}
	return line, true
}

// bytecodeTriggerName is the trigger's rule name, which ignore lists match.
func bytecodeTriggerName(b []byte) (string, bool) {
	line, ok := bytecodeTrigger(b)
	if !ok {
		return "", false
	}
	return ldbTokenize(line, 1)[0], true
}

func (d *Database) compileBytecode(src source, addError func(string, int, error)) {
	if d.report.Loaded >= d.limits.MaxSignatures {
		addError(src.name, 0, errors.New("signature count limit exceeded"))
		return
	}
	line, err := d.addBytecode(src.data)
	if reason, ok := skipReason(err); ok {
		d.skipped(reason, 1)
		return
	}
	if err != nil {
		addError(src.name, line, err)
		return
	}
	d.report.Files = append(d.report.Files, src.name)
	d.report.Loaded++
}

// addBytecode compiles one bytecode program with its logical trigger. The
// header window is checked before the program is validated, and an ignored
// trigger name drops the whole program. It returns the line an error refers to.
func (d *Database) addBytecode(b []byte) (int, error) {
	if err := d.bytecodeWindow(b); err != nil {
		return 1, err
	}
	if name, ok := bytecodeTriggerName(b); ok && d.ignore.ignored(name, name) {
		return 2, skip(SkipIgnoreList)
	}
	p, err := bytecode.Parse(b)
	hook := false
	if err == nil {
		if p.Kind == 1 {
			err = p.ValidateStartup(uint64(d.functionality()))
		} else if err = p.ValidateLogical(uint64(d.functionality())); errors.Is(err, bytecode.ErrUnsupported) && p.ValidateHook(uint64(d.functionality())) == nil {
			hook, err = true, nil
		}
	}
	if errors.Is(err, bytecode.ErrUnsupported) {
		if bytecodeKind(b) == 1 {
			// A startup program with unknown semantics can disable every
			// ClamBC callback. Never publish a generation that ignores it.
			return 0, err
		}
		return d.addUnknownBytecode(b)
	}
	if err != nil {
		return 0, err
	}
	budget := uint64(min(d.limits.MaxTotalBytes, 128<<20))
	if p.StorageBytes() > budget-d.bytecodeBytes {
		return 0, errors.New("compiled bytecode storage limit exceeded")
	}
	if p.Kind == 1 {
		d.startupHooks = append(d.startupHooks, p)
		d.bytecodeBytes += p.StorageBytes()
		d.recordBytecodeSource(bytecodeSourceState{storage: p.StorageBytes(), startupHook: p})
		return 0, nil
	}
	if hook && p.Kind == 258 {
		d.pdfHooks = append(d.pdfHooks, p)
		d.bytecodeBytes += p.StorageBytes()
		d.recordBytecodeSource(bytecodeSourceState{storage: p.StorageBytes(), pdfHook: p})
		return 0, nil
	}
	if hook && p.Kind == 260 && !strings.Contains(p.LogicalSignature, ";") {
		d.preclassHooks = append(d.preclassHooks, preclassHook{program: p, order: d.hookCount, bare: true})
		d.hookCount++
		d.bytecodeBytes += p.StorageBytes()
		d.recordBytecodeSource(bytecodeSourceState{storage: p.StorageBytes(), preclassHook: p})
		return 0, nil
	}
	if hook && !strings.Contains(p.LogicalSignature, ";") {
		target := hookTargets[p.Kind]
		d.bareHooks[target] = append(d.bareHooks[target], rule{ruleCold: &ruleCold{bytecode: p, hook: uint16(p.Kind), hookOrder: d.hookCount}})
		d.hookCount++
		d.bytecodeBytes += p.StorageBytes()
		d.recordBytecodeSource(bytecodeSourceState{storage: p.StorageBytes(), bareHook: p})
		return 0, nil
	}
	if err = d.addLogicalRule(p.LogicalSignature, "", true); err != nil {
		if reason, ok := skipReason(err); ok {
			// The reference rejects a program whose header is selected
			// while its own trigger is not.
			err = fmt.Errorf("bytecode trigger not selected (%s) although its header is", reason)
		}
		return 2, err
	}
	r := &d.rules[len(d.rules)-1]
	r.ownCold()
	r.bytecode = p
	if hook {
		r.hook, r.hookOrder = uint16(p.Kind), d.hookCount
		if p.Kind == 260 {
			d.preclassHooks = append(d.preclassHooks, preclassHook{program: p, order: d.hookCount})
		}
		d.hookCount++
	}
	d.bytecodeBytes += p.StorageBytes()
	d.recordBytecodeSource(bytecodeSourceState{storage: p.StorageBytes()})
	if reason := ruleUnknownReason(r, d.patterns); reason != "" {
		d.unknown(reason, 1)
	}
	return 0, nil
}

// ruleUnknownReason is the reason a rule is unknown when reached, or "".
func ruleUnknownReason(r *rule, patterns []bodyPattern) string {
	if r.unknown != "" {
		return r.unknown
	}
	for _, id := range r.patterns {
		if h := patterns[id].holder; h != nil && h.regex == nil {
			if h.bcomp != nil && r.byteCompareRefs&(uint64(1)<<uint(h.ref)) != 0 {
				continue
			}
			return h.reason
		}
	}
	return ""
}

// hookTargets are the targets whose objects the reference runs each bytecode
// hook kind on: PE unpacker and all-PE hooks, the PDF hook, and the ELF and
// Mach-O unpacker hooks.
var hookTargets = map[uint64]int{257: 1, 258: 10, 259: 1, 261: 6, 262: 9}

// addUnknownBytecode loads a program this engine cannot run. With a logical
// trigger it is unknown on the objects where the trigger holds; a hook
// without one is unknown on every object of its hook's type.
func (d *Database) addUnknownBytecode(b []byte) (int, error) {
	if trigger, ok := bytecodeTrigger(b); ok {
		if err := d.addLogicalRule(trigger, "", true); err != nil {
			if reason, ok := skipReason(err); ok {
				err = fmt.Errorf("bytecode trigger not selected (%s) although its header is", reason)
			}
			return 2, err
		}
		r := &d.rules[len(d.rules)-1]
		r.ownCold().unknown = UnknownBytecode
		if bytecodeKind(b) == 260 {
			r.hook, r.hookOrder = 260, d.hookCount
			d.hookCount++
		}
		d.unknown(UnknownBytecode, 1)
		return 0, nil
	}
	if bytecodeKind(b) == 260 {
		d.unknownPreclass = true
		d.recordBytecodeSource(bytecodeSourceState{unknownPreclass: true})
		d.unknown(UnknownBytecode, 1)
		return 0, nil
	}
	target, ok := hookTargets[bytecodeKind(b)]
	if !ok {
		return 0, errors.New("signature bytecode without a logical trigger")
	}
	d.unknownHooks[target] = UnknownBytecode
	d.recordBytecodeSource(bytecodeSourceState{unknownTarget: target + 1})
	d.unknown(UnknownBytecode, 1)
	return 0, nil
}

// addLine compiles one text-format rule. ext must be a supported line format.
// entry is the whole line with only its line ending removed, which ignore-list
// MD5 entries are compared against. A rule the reference does not select
// returns a skip.
func (d *Database) addLine(ext, s, entry string) error {
	switch ext {
	case ".hdb", ".hsb", ".mdb", ".msb", ".fp", ".sfp":
		return d.addHash(ext, s, entry)
	case ".ndb":
		return d.addExtended(s, entry)
	case ".ldb":
		return d.addLogicalRule(s, entry, false)
	case ".db":
		return d.addLegacy(s, entry)
	case ".pdb", ".wdb":
		return d.addPhishing(ext, s)
	case ".gdb":
		return d.addSafeBrowsing(s)
	case ".ftm":
		return d.addFileTypeLine(s)
	case ".idb":
		return d.addIconLine(s, entry)
	case ".cdb":
		return d.addContainerLine(s)
	case ".crb":
		return d.addCertificateLine(s)
	case ".pwdb":
		return d.addPasswordLine(s, entry)
	}
	return fmt.Errorf("unsupported signature database format %q", ext)
}
func parseSize(s string) (int64, error) {
	if s == "*" {
		return -1, nil
	}
	n, e := strconv.ParseInt(s, 10, 64)
	if e != nil || n < 0 {
		return 0, fmt.Errorf("invalid signature size %q", s)
	}
	return n, nil
}
func validName(n string) error {
	if strings.TrimSpace(n) == "" || len(n) > 256 || !utf8.ValidString(n) {
		return errors.New("invalid signature name")
	}
	for _, r := range n {
		if r < 32 || r == 127 || r == 0x2028 || r == 0x2029 {
			return errors.New("invalid control character in signature name")
		}
	}
	return nil
}

func (d *Database) addHash(ext, line, entry string) error {
	return d.addHashBytes(ext, []byte(line), []byte(entry))
}

func (d *Database) addHashBytes(ext string, line, entry []byte) error {
	// Hash lines have at most five fields. Keep the common 3–5 field parse
	// on the stack; a sixth field is sufficient to reject any longer line.
	var fields [6][]byte
	count := 0
	for count < 5 {
		before, after, ok := bytes.Cut(line, []byte(":"))
		if !ok {
			break
		}
		fields[count] = before
		count++
		line = after
	}
	fields[count] = line
	p := fields[:count+1]
	if len(p) < 3 {
		return errors.New("invalid hash signature fields")
	}
	var window [6]string
	if len(p) > 3 {
		window[3] = string(p[3])
	}
	if len(p) > 4 {
		window[4] = string(p[4])
	}
	minimum, err := d.hashWindow(window[:len(p)])
	if err != nil {
		return err
	}
	digest, size := p[0], p[1]
	table := d.hashes
	if ext == ".mdb" || ext == ".msb" {
		digest, size = p[1], p[0]
		table = d.sections
	}
	if ext == ".fp" || ext == ".sfp" {
		table = d.allow
	}
	n, err := parseHashSize(string(size), len(p), minimum)
	if err != nil {
		return err
	}
	if d.ignore.ignoredBytes(p[2], entry) {
		return skip(SkipIgnoreList)
	}
	if err := validName(string(p[2])); err != nil {
		return err
	}
	if len(digest) != 32 && len(digest) != 40 && len(digest) != 64 {
		return errors.New("hash must contain 32, 40 or 64 hexadecimal characters")
	}
	var decoded [32]byte
	if _, err := hex.Decode(decoded[:], digest); err != nil {
		return errors.New("invalid hexadecimal hash")
	}
	if table == d.sections && d.dconf.PE&dconfPEMD5Sect == 0 {
		return skip(SkipDynamicConfig) // the reference loads but never matches them
	}
	if table == d.hashes {
		d.addWholeHash(decoded[:len(digest)/2], n, p[2])
	} else {
		table.addBytes(decoded[:len(digest)/2], n, p[2])
	}
	return nil
}

// supportedTarget maps a selected rule target (below maxTargets) to its rule
// root.
func supportedTarget(target uint32) (int, error) {
	if target >= maxTargets {
		return 0, fmt.Errorf("unsupported target %q", strconv.FormatUint(uint64(target), 10))
	}
	return int(target), nil
}

// executableTarget reports the roots whose objects carry executable header
// information: PE, ELF and Mach-O. Only their rules may use header offsets
// and descriptors.
func executableTarget(target int) bool { return target == 1 || target == 6 || target == 9 }

func (d *Database) addExtended(line, entry string) error {
	p := strings.Split(line, ":")
	if len(p) < 4 || len(p) > 6 {
		return errors.New("invalid extended signature fields")
	}
	if d.ignore.ignored(p[0], entry) {
		return skip(SkipIgnoreList)
	}
	if len(p) > 4 {
		var maximum *string
		if len(p) == 6 {
			maximum = &p[5]
		}
		if err := d.fieldWindow(p[4], maximum); err != nil {
			return err
		}
	}
	// The target is read only after the window; '*' and empty mean 0.
	if p[1] != "*" && !isNumber(p[1]) {
		return fmt.Errorf("invalid signature target %q", p[1])
	}
	selected := uint32(cAtoi(p[1]))
	if selected >= maxTargets {
		return skip(SkipTarget)
	}
	if e := validName(p[0]); e != nil {
		return e
	}
	target, e := supportedTarget(selected)
	if e != nil {
		return e
	}
	ids, e := d.compilePatterns(target, []subsignature{{offset: p[2], body: p[3]}})
	if e != nil {
		return e
	}
	d.rules = append(d.rules, rule{name: strings.Clone(p[0]), target: target, patterns: ids, ruleCold: &emptyRuleCold})
	if reason := ruleUnknownReason(&d.rules[len(d.rules)-1], d.patterns); reason != "" {
		d.unknown(reason, 1)
	}
	return nil
}

// addLegacy compiles a .db "name=hex" rule: target 0, any offset. Its body is a
// plain pattern; the reference passes it on without modifier parsing.
func (d *Database) addLegacy(line, entry string) error {
	name, body, ok := strings.Cut(line, "=")
	if !ok {
		return errors.New("invalid legacy signature")
	}
	if d.ignore.ignored(name, entry) {
		return skip(SkipIgnoreList)
	}
	if strings.HasPrefix(body, "=") {
		return skip(SkipLegacyFormat)
	}
	if e := validName(name); e != nil {
		return e
	}
	if strings.Contains(body, ":") {
		return errors.New("legacy signature bodies do not take modifiers")
	}
	ids, e := d.compilePatterns(0, []subsignature{{offset: "*", body: body}})
	if e != nil {
		return e
	}
	d.rules = append(d.rules, rule{name: strings.Clone(name), patterns: ids, ruleCold: &emptyRuleCold})
	return nil
}

// subsignature is one subsignature's offset and body text, or a placeholder
// for one this engine loads without its matcher.
type subsignature struct {
	offset, body string
	holder       *placeholder
}

// compilePatterns returns the shared pattern ids of subsignatures. Identical
// text shares one compiled pattern. Nothing is registered unless every
// subsignature compiles.
func (d *Database) compilePatterns(target int, subsignatures []subsignature) ([]int, error) {
	ids := make([]int, len(subsignatures))
	var inlineFresh [4]bodyPattern
	var inlineKeys [4][16]byte
	var keyBuffer [512]byte
	var fresh []bodyPattern
	var freshKeys [][16]byte
	pending := map[[16]byte]int{}
	for i, sub := range subsignatures {
		keyInput := keyBuffer[:0]
		keyInput = append(keyInput, sub.offset...)
		keyInput = append(keyInput, 0)
		keyInput = append(keyInput, sub.body...)
		h := sha256.Sum256(keyInput)
		key := [16]byte(h[:16])
		id, shared := d.patternIDs[key]
		if !shared {
			if fresh == nil {
				// All-shared rules need no temporary pattern storage. Most new
				// rules have at most four subsignatures and fit in local arrays.
				capacity := len(subsignatures) - i
				if capacity <= len(inlineFresh) {
					fresh = inlineFresh[:0]
					freshKeys = inlineKeys[:0]
				} else {
					fresh = make([]bodyPattern, 0, capacity)
					freshKeys = make([][16]byte, 0, capacity)
				}
			}
			id, shared = pending[key]
		}
		if !shared {
			pat, err := bodyPattern{holder: sub.holder, offset: offset{kind: offsetAny}}, error(nil)
			if sub.holder == nil {
				pat, err = compileBody(sub.offset, sub.body)
			}
			if err != nil {
				return nil, err
			}
			pat.nativeAC = strings.ContainsAny(sub.body, "?([{*") || strings.Contains(sub.offset, "VI") || strings.Contains(sub.offset, "$")
			pat.nativeFloat = strings.Contains(sub.offset, ",")
			pat.nativeFixed = sub.holder == nil && fixedACSyntax(sub.body)
			if sub.holder == nil && (pat.nativeAC || !pat.simple) {
				if model, err := compileNativeAC(sub.body); err == nil {
					// The old graph remains the fallback for an unmodeled
					// root and serves file-type recognition. Its reported
					// start can include a character anchor outside the
					// native core. Keep both necessary-anchor views so
					// either consumer has conservative start windows.
					for i := 0; i < pat.compiledAnchorCount(); i++ {
						a := pat.compiledAnchor(i)
						duplicate := false
						for _, b := range model.anchors {
							if a.fold == b.fold && a.minPre == b.minPre && a.maxPre == b.maxPre && bytes.Equal(a.literal, b.literal) {
								duplicate = true
								break
							}
						}
						if !duplicate {
							model.anchors = append(model.anchors, a)
						}
					}
					if pat.graph == nil {
						pat.graph = &patternGraph{}
					}
					pat.graph.native = model
					pat.anchored = true
				} else if pat.graph != nil {
					pat.graph.nativeError = err.Error()
				}
			}
			id = len(d.patterns) + len(fresh)
			pending[key] = id
			fresh = append(fresh, pat)
			freshKeys = append(freshKeys, key)
		}
		ids[i] = id
		var pat *bodyPattern
		if id < len(d.patterns) {
			pat = &d.patterns[id]
		} else {
			pat = &fresh[id-len(d.patterns)]
		}
		if !executableTarget(target) && pat.offset.executable() {
			return nil, errors.New("executable offset requires an executable target")
		}
	}
	base := len(d.patterns)
	assigned := make([]int, len(fresh))
	for i, pat := range fresh {
		if id, ok := d.takeRetiredPatternSlot(); ok {
			d.patterns[id], d.patternKeys[id] = pat, freshKeys[i]
			assigned[i] = id
		} else {
			assigned[i] = len(d.patterns)
			d.patterns = append(d.patterns, pat)
			d.patternKeys = append(d.patternKeys, freshKeys[i])
		}
	}
	for i, id := range ids {
		if id >= base {
			ids[i] = assigned[id-base]
		}
	}
	for key, id := range pending {
		d.patternIDs[key] = assigned[id-base]
	}
	return ids, nil
}

func (d *Database) takeRetiredPatternSlot() (int, bool) {
	limit := min(len(d.patterns), len(d.retiredPatterns)<<6)
	for d.reusePatternCursor < limit {
		id := d.reusePatternCursor
		d.reusePatternCursor++
		if !d.retiredPatterns.has(id) {
			continue
		}
		d.retiredPatterns[id>>6] &^= 1 << uint(id&63)
		if d.freshReusedPatterns == nil {
			d.freshReusedPatterns = newBitset(len(d.patterns))
		}
		d.freshReusedPatterns.set(id)
		return id, true
	}
	return 0, false
}

type matchWork struct {
	ctx             context.Context
	steps, limit    int64
	bytecodeOptions *bytecode.EngineOptions
	fileReliability *uint32
	metadata        map[string]any
	metadataPartial bool
	preclassMatched func(*bytecode.Program)
	// scanExtracted is the enclosing scanner's bounded child path. A nil
	// callback keeps a reached bytecode output operation incomplete.
	scanExtracted      func([]byte, int32) (string, error)
	scanExtractedTyped func([]byte, int32, uint32) (string, error)
	// scanHandler is the enclosing engine's bounded forced-type path. The
	// digest guard is shared by nested scans of the same top-level object.
	scanHandler func([]byte, FileType) (string, error)
	handlerHash *[16]byte
	bufferMode  bool
	// searchHits counts the literal search hits the object's counts keep
	// beyond the first of each literal (see literalSearches).
	searchHits int
	// bytecodeSteps are the steps bytecode runs used; together they may use
	// at most half of the budget, so programs cannot starve the signatures
	// evaluated after them.
	bytecodeSteps int64
	// hookTrace, set only by the reference comparison test, sees each hook
	// program run: its trigger name, kind, return value and failure.
	hookTrace func(name string, kind uint16, value uint64, cause error)
	// rxLiterals remembers which required regex texts the object holds
	// (regex_sub.go).
	rxLiterals map[rxLiteralKey]bool
	// forward lets a start past a pattern's memory bound be decided by the
	// slower forward search (pattern_forward.go). It is set only while a
	// rule whose result depends on such a start is evaluated again.
	forward bool
}

func (w *matchWork) step(n int64) error {
	if n < 0 || n > w.limit-w.steps {
		return ErrMatchLimit
	}
	w.steps += n
	if n >= 1024 || w.steps&1023 < n {
		return w.ctx.Err()
	}
	return nil
}
func hashesFor(ctx context.Context, data []byte) ([]string, error) {
	hs := []hash.Hash{md5.New(), sha1.New(), sha256.New()}
	for start := 0; start < len(data); start += 1 << 20 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		end := min(start+1<<20, len(data))
		for _, h := range hs {
			h.Write(data[start:end])
		}
	}
	result := make([]string, len(hs))
	for i, h := range hs {
		result[i] = string(h.Sum(nil))
	}
	return result, nil
}

// Pattern counts use negative sentinels: not yet computed, and two causes of an
// unknown count. Unknown counts make only the rules that depend on them
// undecidable; every other rule is still evaluated.
const (
	countPending      = -1
	countNeedsHeader  = -2
	countStateLimited = -3
)

// errExecutableHeader reports an executable whose header the offset model
// cannot represent, so header-relative rules and section hashes are undecided.
var errExecutableHeader = errors.New("UTMStack EDR executable header could not be analyzed")

// unknownCause is the first cause among a rule's unknown counts; why holds
// the reasons of placeholder counts.
func unknownCause(counts []int, why []string) error {
	for i, n := range counts {
		switch n {
		case countNeedsHeader:
			return errExecutableHeader
		case countStateLimited:
			return errPatternStates
		case countUnsupported:
			return unknownBecause(why[i])
		}
	}
	return errExecutableHeader
}

// MatchMode selects how an object's bytes are matched. FileMode matches a
// whole object: hash rules run and every offset resolves. BufferMode matches a
// buffer the reference scans without offsets or hashes, such as a macro
// module's text: relative offsets never match, executable descriptors are
// false, and regular expressions apply to the whole buffer as to a file.
type MatchMode uint8

const (
	FileMode MatchMode = iota
	BufferMode
)

// MatchContext describes the object being matched.
type MatchContext struct {
	// Type selects the typed rule root matched before root 0. A raw scan of
	// TEXT_ASCII passes ANY, as the reference does.
	Type FileType
	Mode MatchMode
	// PEAlertBroken matches ClamAV's --alert-broken option. The direct matcher
	// defaults to ClamAV's off setting; the endpoint policy opts in.
	PEAlertBroken bool
	// Layers is the recursion stack for Container and Intermediates, with
	// the object itself on top; nil for an object matched on its own.
	Layers *Ancestry
	// TypeRecognition also runs the type patterns over the whole object and
	// reports what they found.
	TypeRecognition bool
	// ScanExtracted analyzes bytecode-produced output as a recursive child.
	// The callback is synchronous and must enforce the enclosing scan limits.
	ScanExtracted func([]byte, int32) (string, error)
	// ScanExtractedTyped additionally carries a ClamBC-forced container type.
	// It takes precedence over ScanExtracted when provided.
	ScanExtractedTyped func([]byte, int32, uint32) (string, error)
	// BytecodeOptions is the enclosing scanner's effective legacy ClamBC
	// option mask. Nil keeps a reached option query undecided.
	BytecodeOptions *bytecode.EngineOptions
	// FileReliability is the proved ClamAV corrupted_input value for this
	// representation. Nil makes a reached bytecode query incomplete.
	FileReliability *uint32
	// PDF is the per-scan parser snapshot for reached PDF hook phases.
	PDF *bytecode.PDFState
	// PDFParserHooks means the enclosing PDF parser will invoke reached hook
	// phases. The raw matcher must not add speculative phase uncertainty.
	PDFParserHooks bool
	// Metadata is the scan-wide root property tree for bytecode JSON queries.
	// Nil reports JSON inactive; a partial tree leaves absent queries undecided.
	Metadata        map[string]any
	MetadataPartial bool
	// PreclassMatched flags a root-layer kind-260 trigger for the later
	// metadata preclassification phase. A nil program flags an unsupported
	// kind-260 trigger. A nil callback leaves the trigger undecided.
	PreclassMatched func(*bytecode.Program)
	// ScanHandler runs the full forced-type object scan for a HandlerType
	// rule. Nil leaves such a reached rule incomplete.
	ScanHandler func([]byte, FileType) (string, error)
	HandlerHash *[16]byte
}

// MatchResult is the outcome of matching one object.
type MatchResult struct {
	Name string
	// Deferred is set for a format heuristic whose detection loses to a
	// recursive embedded-object detection in the reference's scan order.
	Deferred bool
	// PEHeaderRejected means ClamAV's PE parser stopped at a structural
	// header error before producing any executable representation.
	PEHeaderRejected bool
	// DetectedType and Embedded are what type recognition found.
	DetectedType FileType
	Embedded     []EmbeddedType
	// Unpacked is the program the reference's UPX step rebuilt from this
	// Windows program; the reference scans it as a child. UPXDecided is
	// false when this engine cannot tell what that step does.
	Unpacked   []byte
	UPXDecided bool
}

// EmbeddedType is a type recognized at an offset inside an object.
type EmbeddedType struct {
	Type   FileType
	Offset int
}

// exeStatus is how an object's executable header resolves relative offsets.
type exeStatus uint8

const (
	// exeNone: the object's target has no executable header; EOF offsets
	// resolve and executable offsets cannot occur in its roots.
	exeNone exeStatus = iota
	// exeParsed: every offset and executable descriptor resolves.
	exeParsed
	// exeFailed: the reference cannot parse a header either (or the object
	// is a buffer), so no relative offset matches, EOF included, and
	// executable descriptors are false.
	exeFailed
	// exeUnknown: a header this engine cannot model although the reference
	// may. An executable offset is unknown when its pattern occurs anywhere
	// in the object, and executable descriptors are unknown.
	exeUnknown
)

// objectInfo is what matching needs to know about an object besides its bytes.
type objectInfo struct {
	typ    FileType
	target int
	mode   MatchMode
	status exeStatus
	meta   *peMeta // header information when status is exeParsed
	layers *Ancestry
	// groupIcon caches iconsComparable's resource walk: 1 yes, -1 no.
	groupIcon int8
	// icons caches the first group icon's parsed icons when iconState is
	// 1; -1 is an uncertain group (icons.go).
	icons     []programIcon
	iconState int8
	// vinfo caches versionOffsets: the version-string key offsets, known
	// when vinfoState is 1, unknown (header not modeled) when -1.
	vinfo      []int
	vinfoState int8
}

// versionOffsets are the offsets VI offsets match at, or ok false when this
// engine cannot know them, on a header it cannot model.
func (info *objectInfo) versionOffsets(data []byte) ([]int, bool) {
	if info.vinfoState == 0 {
		info.vinfoState = -1
		if info.status == exeParsed {
			info.vinfo, info.vinfoState = info.meta.versionOffsets(data), 1
		}
	}
	return info.vinfo, info.vinfoState > 0
}

func (d *Database) objectInfo(data []byte, mc MatchContext) objectInfo {
	info := objectInfo{typ: mc.Type, target: targetOf(mc.Type), mode: mc.Mode, layers: mc.Layers}
	switch {
	case mc.Mode == BufferMode:
		info.status = exeFailed
	case info.target == 1:
		switch meta := parsePE(data); {
		case meta == nil:
			info.status = exeFailed // typed an executable without MZ at offset 0
		case meta.valid:
			info.status, info.meta = exeParsed, meta
		case meta.failed:
			info.status = exeFailed
		default:
			info.status = exeUnknown
		}
	case info.target == 6 || info.target == 9:
		// These readers follow the reference's exactly, so a header they
		// reject is one the reference rejects too.
		var meta *peMeta
		switch mc.Type {
		case TypeELF:
			meta = parseELF(data)
		case TypeMachO:
			meta = parseMachO(data)
		}
		if meta == nil {
			info.status = exeFailed
		} else {
			info.status, info.meta = exeParsed, meta
		}
	}
	return info
}

// patternCounts is one object's subsignature counting state. Rules share
// patterns, so each candidate is counted at most once per object, resuming
// when a later rule needs a higher cap.
type patternCounts struct {
	d          *Database
	data       []byte
	info       *objectInfo
	w          *matchWork
	candidates *candidateSet
	counts     map[int]*patternCount
	// occurs caches, per executable-offset pattern on an unusable header,
	// whether it matches anywhere (countNeedsHeader) or nowhere (0).
	occurs map[int]int
	// c and first are the last rule's counts and first match offsets; why
	// holds the reason of each placeholder count that is unknown.
	c     []int
	first []uint32
	why   []string
	// hooks are the hook rules whose triggers held on this object.
	hooks []*rule
	// regexes are the regex results of this object, by pattern.
	regexes map[int]regexResult
	// image is the object's fuzzy image hash, once computed.
	image   *imageHash
	native  [maxTargets]*nativeACResult
	reached bitset
}

func (d *Database) newPatternCounts(data []byte, info *objectInfo, w *matchWork) (*patternCounts, error) {
	candidates, err := d.candidates(data, w)
	if err != nil {
		return nil, err
	}
	return &patternCounts{d: d, data: data, info: info, w: w, candidates: candidates, counts: map[int]*patternCount{}, occurs: map[int]int{}}, nil
}

// rule sets c and first to the counts and first offsets of r's subsignatures,
// capped as r needs them. A pattern none of whose anchors occurs has a known
// count of 0 whatever its offset. An error is the shared budget or
// cancellation.
func (m *patternCounts) rule(r *rule) error {
	m.c, m.first, m.why = m.c[:0], m.first[:0], m.why[:0]
	for i, id := range r.patterns {
		pattern := &m.d.patterns[id]
		n, offset, why := 0, missingOffset, ""
		switch {
		case pattern.holder != nil && pattern.holder.reason == UnknownFuzzyImage && targetOf(m.info.typ) == 5:
			if n, why = m.fuzzyImage(pattern.holder); n == 1 {
				offset = 0
			}
		// In buffer mode (macro module text) the reference's regex subject
		// is the whole buffer (its debug output on 1,700 module scans) and
		// absolute and end offsets apply as in a file (checked with marker
		// rules on a module).
		case pattern.holder != nil && pattern.holder.regex != nil:
			var err error
			if n, offset, why, err = m.regex(id, pattern.holder, r.countCap(i)); err != nil {
				return err
			}
		case pattern.holder != nil && pattern.holder.bcomp != nil &&
			r.byteCompareRefs&(uint64(1)<<uint(pattern.holder.ref)) != 0:
			var err error
			if n, why, err = m.byteCompare(r, pattern.holder); err != nil {
				return err
			}
		case pattern.holder != nil:
			if n = pattern.holder.count(m.c, m.info.typ); n == countUnsupported {
				why = pattern.holder.reason
			}
		case !m.candidates.has(id):
		case acRegistration(m.d, r) && m.d.nativeAC[r.target] != nil && m.d.nativeAC[r.target].modeled(id):
			res, err := m.nativeRoot(r.target)
			if err != nil {
				return err
			}
			if pc := res.counts[id]; pc != nil {
				m.counts[id] = pc
				cap := r.countCap(i)
				if r.bytecode != nil {
					cap = int(^uint(0) >> 1)
				}
				if r.byteCompareRefs&(uint64(1)<<uint(i)) != 0 {
					cap = int(^uint(0) >> 1)
				}
				n, offset = pc.value(cap)
				why = res.why[id]
			}
		case pattern.offset.relative() && m.info.status == exeFailed:
		case pattern.offset.kind == offsetVersionInfo:
			// Only PE headers carry version information; the reference
			// matches these offsets nowhere else. A match must start at the
			// key of a version string.
			if m.info.target != 1 {
				break
			}
			if keys, ok := m.info.versionOffsets(m.data); ok {
				var err error
				if n, offset, err = pattern.countAt(m.data, keys, m.w); err != nil {
					return err
				}
				break
			}
			var err error
			if n, err = m.unknownUnlessAbsent(id, pattern); err != nil {
				return err
			}
			if n != 0 {
				n, why = countUnsupported, UnknownVersionInfo
			}
		case pattern.offset.executable() && m.info.status == exeUnknown:
			var err error
			if n, err = m.unknownUnlessAbsent(id, pattern); err != nil {
				return err
			}
		default:
			maxCount := r.countCap(i)
			if r.bytecode != nil {
				maxCount = len(m.data) + 1
			}
			if r.byteCompareRefs&(uint64(1)<<uint(i)) != 0 {
				// One more than the maximum possible number of start
				// positions ensures pc.done and a final pc.last.
				maxCount = len(m.data) + 2
			}
			pc := m.counts[id]
			if pc == nil {
				pc = m.candidates.startCount(pattern, id, m.data, m.info.meta)
				m.counts[id] = pc
			}
			if err := pattern.countMore(m.data, m.w, pc, maxCount); err != nil {
				return err
			}
			n, offset = pc.value(maxCount)
		}
		m.c, m.first, m.why = append(m.c, n), append(m.first, offset), append(m.why, why)
	}
	return nil
}

// unlimit prepares the counts of r's patterns that stopped at their memory
// bound to be resumed with the forward search, when r's result depends on
// them: its expression is undecided, or it is true and its program reads
// every count. It reports whether any count was prepared.
func (m *patternCounts) unlimit(r *rule) bool {
	limited := false
	for _, n := range m.c {
		limited = limited || n == countStateLimited
	}
	if !limited {
		return false
	}
	if state := ruleState(r, m.c); state == logicFalse || state == logicTrue && r.bytecode == nil {
		return false
	}
	any := false
	for _, id := range r.patterns {
		if pc := m.counts[id]; pc != nil && pc.limited {
			pc.limited, any = false, true
		}
	}
	return any
}

// fuzzyImage counts a fuzzy image hash subsignature on an image object: 1
// when the object's hash is the rule's, 0 when its certain bits already
// differ, otherwise unknown. The hash is computed once per object. Only a
// distance of 0 is decided: the reference in use matches no other.
func (m *patternCounts) fuzzyImage(h *placeholder) (int, string) {
	// The official .ftm classifies JPEG 2000 as GRAPHICS, but the pinned
	// ClamAV image-hash decoder cannot load JP2. Its fuzzy rule is false.
	if bytes.HasPrefix(m.data, []byte("\x00\x00\x00\x0cjP  \r\n\x87\n")) {
		return 0, ""
	}
	if m.image == nil {
		ih := fuzzyImageHash(m.data, m.w)
		m.image = &ih
	}
	ih := m.image
	switch {
	case !ih.ok:
		return countUnsupported, UnknownFuzzyImage
	case (ih.hash^h.imageHash)&^ih.uncertain != 0:
		return 0, ""
	case ih.uncertain == 0:
		return 1, ""
	}
	return countUnsupported, UnknownFuzzyImage
}

// regex counts a regex subsignature: 0 when its trigger does not hold on
// the earlier counts, otherwise the regex's count on the object, found once
// per object and pattern. A trigger that is unknown leaves a nonzero count
// unknown; so does a regex the reference's library could have given up on.
func (m *patternCounts) regex(id int, h *placeholder, maxCount int) (int, uint32, string, error) {
	trigger := h.trigger.evaluate(m.c).state
	if trigger == logicFalse {
		return 0, missingOffset, "", nil
	}
	res, ok := m.regexes[id]
	if !ok || res.err == nil && res.count >= res.cap && res.cap < maxCount {
		n, first, err := h.regex.run(m.data, maxCount, m.w)
		if err != nil && err != errRegexLimit {
			return 0, missingOffset, "", err
		}
		res = regexResult{count: n, first: first, cap: maxCount, err: err}
		if m.regexes == nil {
			m.regexes = map[int]regexResult{}
		}
		m.regexes[id] = res
	}
	switch {
	case res.err != nil:
		return countUnsupported, missingOffset, UnknownPCRELimit, nil
	case res.count == 0:
		return 0, missingOffset, "", nil
	case trigger != logicTrue:
		why := UnknownPCRE
		for _, w := range m.why {
			if w != "" {
				why = w
				break
			}
		}
		return countUnsupported, missingOffset, why, nil
	}
	return min(res.count, maxCount), res.first, "", nil
}

type regexResult struct {
	count int
	first uint32
	cap   int
	err   error
}

// unknownUnlessAbsent decides an executable-offset pattern on a header this
// engine cannot model: 0 when the pattern matches nowhere in the object,
// which holds wherever the header would put the offset; otherwise unknown.
func (m *patternCounts) unknownUnlessAbsent(id int, p *bodyPattern) (int, error) {
	if n, ok := m.occurs[id]; ok {
		return n, nil
	}
	anywhere := *p
	anywhere.offset = offset{kind: offsetAny}
	pc := m.candidates.startCount(&anywhere, id, m.data, nil)
	if err := anywhere.countMore(m.data, m.w, pc, 1); err != nil {
		return 0, err
	}
	n := 0
	if v, _ := pc.value(1); v != 0 {
		n = countNeedsHeader
	}
	m.occurs[id] = n
	return n, nil
}

// Match matches data as the raw scan of an object matched on its own: it is
// typed first, and TEXT_ASCII, HTML and IGNORED objects are matched against
// root 0 only. The scan pipeline calls MatchObject with the object's context
// instead. It is safe for concurrent callers.
func (d *Database) Match(ctx context.Context, data []byte) (string, error) {
	name, _, err := d.matchCounted(ctx, data, d.limits.matchLimit(len(data)))
	return name, err
}

// matchCounted is Match with an explicit work budget. It also reports the
// steps used, for tests and measurements.
func (d *Database) matchCounted(ctx context.Context, data []byte, limit int64) (string, int64, error) {
	w := &matchWork{ctx: ctx, limit: limit}
	res, err := d.matchObject(w, data, MatchContext{Type: d.rawScanType(data)})
	return res.Name, w.steps, err
}

// rawScanType is the type Match gives an object.
func (d *Database) rawScanType(data []byte) FileType {
	t, _ := d.DetermineType(data)
	switch t {
	case TypeTextASCII, TypeHTML, TypeIgnored:
		return TypeAny
	}
	return t
}

// MatchObject applies allow hashes (file mode), PE section hashes
// (executables in file mode), standalone bodies of the object's typed root
// followed by root 0, whole-file hashes, then logical rules and their bytecode
// callbacks in registration order. It is safe for concurrent callers.
// Decoding and normalization are the caller's responsibility.
//
// A rule that cannot be decided (a per-pattern state bound, an unusable
// header, a failed bytecode callback or an unsupported bytecode label) does
// not stop the others: a detection by any rule wins, and otherwise the
// result is an ErrRuleIncomplete error. Cancellation and the shared work
// budget stop at once.
func (d *Database) MatchObject(ctx context.Context, data []byte, mc MatchContext) (MatchResult, error) {
	return d.matchObject(&matchWork{ctx: ctx, limit: d.limits.matchLimit(len(data))}, data, mc)
}

func (d *Database) matchObject(w *matchWork, data []byte, mc MatchContext) (MatchResult, error) {
	ctx := w.ctx
	w.scanExtracted = mc.ScanExtracted
	w.scanExtractedTyped = mc.ScanExtractedTyped
	w.bytecodeOptions = mc.BytecodeOptions
	w.fileReliability = mc.FileReliability
	w.metadata = mc.Metadata
	w.metadataPartial = mc.MetadataPartial
	w.preclassMatched = mc.PreclassMatched
	w.scanHandler = mc.ScanHandler
	w.handlerHash = mc.HandlerHash
	w.bufferMode = mc.Mode == BufferMode
	var res MatchResult
	if err := ctx.Err(); err != nil {
		return res, err
	}
	var pending error
	var digests []string
	incomplete := func(err error) {
		if pending == nil {
			pending = fmt.Errorf("%w: %w", ErrRuleIncomplete, err)
		}
	}
	if mc.Mode == FileMode && (d.hashes.len() > 0 || d.allow.len() > 0) {
		var err error
		digests, err = hashesFor(ctx, data)
		if err != nil {
			return res, err
		}
		for _, digest := range digests {
			if d.allow.lookup(digest, int64(len(data))) != "" {
				return res, nil
			}
		}
	}
	info := d.objectInfo(data, mc)
	if info.target == 1 && mc.Mode == FileMode && d.sections.len() > 0 && bytes.HasPrefix(data, []byte("MZ")) {
		if name, err := d.matchSections(ctx, data, &info, incomplete); name != "" || err != nil {
			res.Name = name
			return res, err
		}
	}
	if reason := d.unknownHooks[info.target]; reason != "" && mc.Mode == FileMode {
		incomplete(unknownBecause(reason))
	}
	m, err := d.newPatternCounts(data, &info, w)
	if err != nil {
		return res, err
	}
	// Rules left out of reached cannot change the result (see ruleindex.go).
	reached, err := d.reachedRules(m.candidates, info.target, w)
	if err != nil {
		return res, err
	}
	m.reached = reached
	if mc.TypeRecognition && mc.Mode == FileMode {
		if res.DetectedType, res.Embedded, err = m.recognize(mc.Type); err != nil {
			return res, err
		}
	}
	roots := []int{0}
	if info.target != 0 {
		roots = []int{info.target, 0}
	}
	// Standalone BM and AC bodies precede hashes. Logical expressions and
	// their bytecode callbacks share the later registration-ordered pass;
	// a logical body is not a standalone AC detection. Each typed root's
	// BM/AC pass completes before the generic root begins.
	for phase := 0; phase < 2; phase++ {
		if phase == 1 {
			for _, digest := range digests {
				if name := d.hashes.lookup(digest, int64(len(data))); name != "" {
					res.Name = name
					return res, nil
				}
			}
		}
		for _, root := range roots {
			passes := 1
			if phase == 0 {
				passes = 2 // BM, then standalone AC
			}
			for pass := 0; pass < passes; pass++ {
				if phase == 0 && pass == 0 {
					name, cause, err := d.matchStandaloneBM(reached, root, m, &info, w)
					if err != nil {
						return res, err
					}
					if name != "" {
						res.Name = name
						return res, nil
					}
					if cause != nil {
						incomplete(cause)
					}
					continue
				}
				if phase == 0 && pass == 1 && d.nativeAC[root] != nil {
					name, cause, err := d.matchNativeAC(root, m, &info, w)
					if err != nil {
						return res, err
					}
					if name != "" {
						res.Name = name
						return res, nil
					}
					if cause != nil {
						incomplete(cause)
					}
					continue
				}
				if phase == 0 && pass == 1 && d.fixedAC[root] != nil {
					name, cause, err := d.matchFixedAC(reached, root, m, &info, w)
					if err != nil {
						return res, err
					}
					if name != "" {
						res.Name = name
						return res, nil
					}
					if cause != nil {
						incomplete(cause)
					}
					continue
				}
				for ri := range reached.members() {
					r := &d.rules[ri]
					logical := r.expr != nil || r.bytecode != nil || r.unknown == UnknownBytecode
					if r.target != root || logical != (phase == 1) || phase == 0 && d.standaloneAC(r) != (pass == 1) {
						continue
					}
					name, cause, err := d.evaluateRule(r, m, &info, w)
					if err != nil {
						return res, err
					}
					if name != "" {
						res.Name = name
						return res, nil
					}
					if cause != nil {
						incomplete(cause)
					}
				}
			}
		}
	}
	res.UPXDecided = true
	if info.target == 1 && mc.Mode == FileMode {
		p, err := d.peParser(w, data, d.executableHooks(info.target, m.hooks), !mc.PEAlertBroken)
		if err != nil {
			return res, err
		}
		res.PEHeaderRejected = p.headerRejected
		if p.name != "" {
			res.Name = p.name
			res.Deferred = p.deferred
			return res, nil
		}
		if p.cause != nil {
			incomplete(p.cause)
		}
		res.Unpacked, res.UPXDecided = p.unpacked, p.upxDecided
	}
	if (info.target == 6 || info.target == 9) && mc.Mode == FileMode && (len(m.hooks) > 0 || len(d.bareHooks[info.target]) > 0) {
		name, cause, err := d.runOtherExecutableHooks(w, data, d.executableHooks(info.target, m.hooks), info.target)
		if err != nil {
			return res, err
		}
		if name != "" {
			res.Name = name
			return res, nil
		}
		if cause != nil {
			incomplete(cause)
		}
	}
	if info.target == 10 && mc.Mode == FileMode && len(d.pdfHooks) > 0 && !mc.PDFParserHooks {
		name, cause, err := d.runPDFHooks(w, data)
		if err != nil {
			return res, err
		}
		if name != "" {
			res.Name = name
			return res, nil
		}
		if cause != nil {
			incomplete(cause)
		}
	}
	if err := ctx.Err(); err != nil {
		return res, err
	}
	return res, pending
}

// standaloneAC is the pinned profile's matcher family for an NDB/DB rule.
// The native root table permits BM only for generic and PE rules. PE's
// offset-mode BM path additionally excludes any and floating offsets.
func (d *Database) standaloneAC(r *rule) bool {
	if r.target != 0 && r.target != 1 || len(r.patterns) != 1 {
		return true
	}
	p := &d.patterns[r.patterns[0]]
	return p.nativeAC || !p.simple || p.nocase || p.fullword || r.target == 1 && (p.offset.kind == offsetAny || p.nativeFloat)
}

// matchSections checks PE section hashes. Like the reference, a section is
// hashed only when a rule names its size; a header the reference rejects has
// no sections to hash, and one this engine cannot model leaves them
// undecided.
func (d *Database) matchSections(ctx context.Context, data []byte, info *objectInfo, incomplete func(error)) (string, error) {
	switch info.status {
	case exeFailed:
		return "", nil
	case exeUnknown:
		incomplete(errExecutableHeader)
		return "", nil
	}
	for _, s := range info.meta.sections {
		if s.size == 0 || !d.sectionSizes[int64(s.size)] && !d.sectionSizes[-1] {
			continue
		}
		digests, err := hashesFor(ctx, data[s.raw:int64(s.raw)+int64(s.size)])
		if err != nil {
			return "", err
		}
		for _, digest := range digests {
			if name := d.sections.lookup(digest, int64(s.size)); name != "" {
				return name, nil
			}
		}
	}
	return "", nil
}

// evaluateRule decides one rule on the object: a detection name, a cause
// that leaves it undecided, or a budget, cancellation or context error.
func (d *Database) evaluateRule(r *rule, m *patternCounts, info *objectInfo, w *matchWork) (string, error, error) {
	data := m.data
	if err := w.step(1); err != nil {
		return "", nil, err
	}
	if d.bytecodeDisabled && (r.bytecode != nil || r.unknown == UnknownBytecode) {
		return "", nil, nil
	}
	if r.dead || r.size != nil && !r.size.contains(int64(len(data))) || !r.inLayers(info.layers) {
		return "", nil, nil
	}
	// Descriptors are checked after the expression, as in the reference:
	// an unusable header only matters when the expression can be true.
	descriptorsUnknown := false
	if r.entry != nil || r.sections != nil {
		switch info.status {
		case exeParsed:
			if r.entry != nil && !r.entry.contains(int64(info.meta.entry)) || r.sections != nil && !r.sections.contains(int64(len(info.meta.sections))) {
				return "", nil, nil
			}
		case exeUnknown:
			descriptorsUnknown = true
		default:
			return "", nil, nil
		}
	}
	if err := m.rule(r); err != nil {
		return "", nil, err
	}
	if m.unlimit(r) {
		w.forward = true
		err := m.rule(r)
		w.forward = false
		if err != nil {
			return "", nil, err
		}
	}
	c, first := m.c, m.first
	state := ruleState(r, c)
	if state == logicFalse {
		return "", nil, nil
	}
	if state == logicUnknown {
		return "", unknownCause(c, m.why), nil
	}
	if descriptorsUnknown {
		return "", errExecutableHeader, nil
	}
	if r.icons != nil {
		state, err := d.iconConditionWithWork(info, data, *r.icons, w)
		if err != nil {
			return "", nil, err
		}
		if state == logicFalse {
			return "", nil, nil
		}
		if state == logicUnknown {
			return "", unknownBecause(UnknownIconGroup), nil
		}
	}
	switch {
	case r.unknown == UnknownHandlerType && w.bufferMode:
		// Normalized sub-buffers have no reference whole-object hash, so
		// ClamAV alerts directly instead of duplicating them as a new layer.
		return r.name, nil, nil
	case r.hook == 260 && r.unknown == UnknownBytecode:
		if w.preclassMatched == nil {
			return "", unknownBecause(UnknownBytecode), nil
		}
		w.preclassMatched(nil)
		return "", nil, nil
	case r.unknown != "":
		return "", unknownBecause(r.unknown), nil
	case r.handler != nil:
		if w.bufferMode {
			return r.name, nil, nil
		}
		if w.scanHandler == nil || w.handlerHash == nil {
			return "", unknownBecause(UnknownHandlerType), nil
		}
		digest := md5.Sum(data)
		if digest == *w.handlerHash {
			return r.name, nil, nil
		}
		*w.handlerHash = digest
		name, err := w.scanHandler(data, *r.handler)
		return name, nil, err
	case r.bytecode == nil:
		return r.name, nil, nil
	case r.hook != 0:
		// A hook's trigger only flags the object; the program runs in the
		// hook phase after the rule pass (pehook.go).
		if r.hook == 260 {
			if w.preclassMatched == nil {
				return "", unknownBecause(UnknownBytecode), nil
			}
			w.preclassMatched(r.bytecode)
			return "", nil, nil
		}
		m.hooks = append(m.hooks, r)
		return "", nil, nil
	}
	// The reference gives a program the header of a parsed PE file (and of
	// parsed ELF and Mach-O files, which this engine leaves empty); where
	// this engine's header data would differ, a program that reads it
	// cannot decide the object.
	if r.bytecode.ReadsExecutableHeader() && (info.status == exeUnknown || info.status == exeParsed && info.target != 1) {
		return "", errExecutableHeader, nil
	}
	// A callback receives every count, so each one must be known.
	matches := bytecode.EmptyMatchContext()
	if info.status == exeParsed && info.target == 1 {
		matches.PE = info.meta.bytecodePE()
	}
	for i := range r.patterns {
		if c[i] < 0 {
			return "", unknownCause(c, m.why), nil
		}
		if c[i] > 0 {
			if uint64(c[i]) > uint64(^uint32(0)) {
				return "", ErrMatchLimit, nil
			}
			matches.Counts[i] = uint32(c[i])
			matches.Offsets[i] = first[i]
		}
	}
	name, _, cause, err := runBytecode(w, r.bytecode, data, matches)
	return name, cause, err
}

// errBytecodeShare leaves a program undecided once the object's bytecode runs
// have used their share of the budget.
var errBytecodeShare = fmt.Errorf("%w: bytecode share of the object's budget used", ErrMatchLimit)

// runBytecode runs a program on the object within the shared work budget. It
// returns the detection name the program set (checked and classified), its
// return value, a cause that leaves the object undecided, or a budget,
// cancellation or context error.
func runBytecode(w *matchWork, p *bytecode.Program, data []byte, matches bytecode.MatchContext) (string, uint64, error, error) {
	ctx := w.ctx
	remaining := w.limit - w.steps
	if remaining <= 0 {
		return "", 0, nil, ErrMatchLimit
	}
	share := w.limit/2 - w.bytecodeSteps
	if share <= 0 {
		return "", 0, errBytecodeShare, nil
	}
	allowance := min(remaining, share, bytecode.MaxRunSteps)
	matches.Options = w.bytecodeOptions
	matches.FileReliability = w.fileReliability
	matches.Metadata = w.metadata
	matches.MetadataPartial = w.metadataPartial
	var result bytecode.Result
	var err error
	if w.scanExtractedTyped != nil {
		result, err = p.RunWithMatchContextAndTypedScanner(ctx, 0, nil, data, matches, w.scanExtractedTyped, uint64(allowance))
	} else {
		result, err = p.RunWithMatchContextAndScanner(ctx, 0, nil, data, matches, w.scanExtracted, uint64(allowance))
	}
	used := int64(min(result.WorkUsed, uint64(allowance)))
	w.bytecodeSteps += used
	if result.ChildName != "" {
		// Only the enclosing recursive scanner can set this field. A child
		// detection survives a later bytecode fault or cancellation. The
		// allowance already bounds used by the remaining matcher budget.
		_ = w.step(used)
		return result.ChildName, result.Value, nil, nil
	}
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", 0, nil, ctxErr
		}
		// A failed run is charged the steps it used.
		if e := w.step(used); e != nil || w.steps >= w.limit {
			return "", 0, nil, ErrMatchLimit
		}
		if errors.Is(err, bytecode.ErrExtraction) || errors.Is(err, bytecode.ErrPDFObjects) || errors.Is(err, bytecode.ErrUnsupported) {
			// The reference would scan the extracted output, or answer
			// from its PDF parser's object model.
			return "", 0, unknownBecause(UnknownBytecode), nil
		}
		return "", 0, fmt.Errorf("UTMStack EDR bytecode analysis failed: %w", err), nil
	}
	if err = w.step(used); err != nil {
		return "", 0, nil, err
	}
	if !result.VirusNameSet {
		return "", result.Value, nil, nil
	}
	if err := validName(result.VirusName); err != nil {
		return "", 0, fmt.Errorf("UTMStack EDR bytecode detection label: %w", err), nil
	}
	for _, prefix := range []string{"PUA.", "Heuristics.", "BC.Heuristics.", "Weak."} {
		if strings.HasPrefix(result.VirusName, prefix) {
			return "", 0, fmt.Errorf("UTMStack EDR bytecode detection classification: %w", bytecode.ErrUnsupported), nil
		}
	}
	return result.VirusName, result.Value, nil, nil
}
