// Package engine implements the in-process UTMStack EDR scanning engine.
package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/buildinfo"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
	"github.com/utmstack/UTMStack/shared/logger"
)

// ImplementationVersion is part of every verdict's identity. v7 selects rules
// at functionality level 216 with ignore lists and dynamic configuration, so
// the same database files can now load different rules. v8 extracts more
// formats (cpio, SZDD, BinHex, mbox, RTF objects, HTML data URIs, bzip2 ZIP
// members, ARJ, Symbian SIS, NSIS and InstallShield installers, embedded
// executables) and reports packed executables as incomplete. v9 types every
// object like the reference, matches typed rule roots, normalized text and
// layered container rules, and loads rules it cannot evaluate as unknown when
// reached, so earlier clean verdicts may no longer hold. v10 gives logical
// bytecode programs the PE header record, sections and the runtime functions
// they were missing (malloc, memstr, read_number, the PE functions,
// get_environment, disasm_x86), so programs that were unknown when reached
// now run and programs that ran see real header data. v11 runs the PE hook
// programs (all-PE and unpacker hooks) on the reference's parser path, reads
// __clambc_kind as 0, charges runtime-call bytes per 64 and lets a run use
// 50 million steps (half of the object's budget for all runs together), so
// undecided objects may now be decided. v12 evaluates regex subsignatures
// (they were unknown when reached) and decides long pattern gaps by a
// forward search. v13 matches HTML through the reference's normalized views
// (page without comments, text, scripts, data: URI files) instead of
// leaving it incomplete, and no longer matches raw HTML with the generic
// rules. v14 decides icon-group rules on a program's first group icon where
// the measures it computes are enough (they were unknown when reached), and
// rejects icon databases the reference rejects. v15 unpacks UPX programs as
// the reference does (NRV and LZMA attempts, the rebuilt program scanned as
// a child) and lets the reference's parser decide packed programs: only a
// built-in unpacker this engine lacks leaves one incomplete, which can turn
// a clean verdict incomplete. v16 reads PDF files as the reference does
// (PDF hook programs run, streams decrypted, predictors not undone, lengths
// measured to endstream), which can change a clean verdict. v17 matches
// macro projects as the reference does (its project dump as a file, module
// source as buffers, both through its own decompressor) instead of leaving
// every document with macros incomplete, ignores VBA storages in files
// without a _VBA_PROJECT stream, and scans the pictures of workbook drawing
// groups. v27 extracts 7-Zip archives (LZMA, LZMA2, PPMd and the branch
// filters, as the reference decodes them) instead of leaving them
// incomplete, which can turn an incomplete verdict into a detection. v28
// unpacks PESpin programs as the reference does (the last built-in unpacker
// this engine lacked) instead of leaving them incomplete. v29 decodes
// Deflate64 and imploded ZIP members as the reference does. v30 extracts
// Compiled HTML Help files (LZX-compressed and stored entries) instead of
// leaving them incomplete.
// v31 continues supported HTML views when script or link analysis is incomplete
// and decodes single-cabinet LZX folders with cancellation and expansion limits.
// v32 adds bounded ZIP64/Zstandard/LZMA and normalized HTML extraction, propagates
// PDF resource limits, and includes the effective scan deadline in cache identity.
// v33 decodes folded MIME base64, HTML entity-escaped data URLs and Unix ar
// members, extending the decoded views and therefore invalidating prior caches.
// v34 streams native child metadata/CSS content and charges discarded malformed
// CSS decode bytes to the work budget, so previous clean cache entries expire.
// v35 additionally matches bounded RTF visible-text views, including escaped
// bytes and Unicode controls, so previous raw-only clean verdicts expire.
// v36 evaluates binary-integer logical byte comparisons and can run the
// selected extract_new-declaring PE bytecode programs until a reached
// extraction call stops as incomplete. Both change which objects are decided.
// v37 scans bytecode-produced output as bounded recursive children, including
// explicit extract_new members and the final implicit flush. A child detection
// can change a former incomplete verdict, so prior clean caches must expire.
// v38 also scans pending output after a VM fault. v39 adds ClamBC matchicon
// argument validation. v40 reproduces ClamAV 1.5.4's source-proven API 89
// no-match behavior, so previously incomplete candidates may be decided.
// v41 decodes and recursively scans FSG 2.0 and 1.33 PE children; an earlier
// incomplete PE may now be detected or decided.
// v42 decodes and recursively scans NsPack PE children; an earlier incomplete
// PE may now be detected or decided.
// v43 scans bounded Flate output from damaged PDF streams and recovers two
// malformed stream starts; these can change incomplete verdicts to detections.
// v44 lets ClamBC programs query the native scanner's effective scan and
// database option masks. A reached option query can now change an incomplete
// result into a detection, so previous clean cache entries must expire.
// v45 matches the reference's X-type phishing allowlist regexes against the
// cleaned full URL pair, including scheme and explicit anchors, and removes
// its trailing optional-path forms. A previously incomplete or detected
// phishing link may now be allowed by an official rule.
// v46 preserves the reference's literal terminal %xx and root-dot host forms
// for static phishing allow rules, and keeps the root dot distinct for domain
// spoof checks. It also resolves static backslash patterns as unreachable on
// the reference's mandatory normalized-link path.
// v47 preserves displayed URL userinfo for X allow rules. The final official
// WDB rule can now match the pair that ClamAV accepts, so prior phishing and
// incomplete verdicts may change.
// v48-v51 tighten HandlerType selection and force the bounded rescan after
// object views, including SCRIPT normalization and ZIP children. v52 rejects
// unsupported nonzero fuzzy-image distances at database load. v53 decodes
// classic TIFF images for fuzzy hash rules. v54 types and decodes BMP images
// for the same rules. v55 uses the reference's f32 resize arithmetic for
// fuzzy images and narrows the rounding guard on small lossless images; it
// also bounds image dimensions to prevent oversized temporary buffers. v56
// prevents multiply-add fusion in image luma conversion so the reference's
// half-integer rounding is preserved. v57 also preserves separate f32
// products in Lanczos3 resizing; those products affect its rounded pixels.
// Its 5,000-image independent lossless oracle permits removing the small-image
// rounding guard, allowing exact-match decisions on the verified path.
// Each changes verdicts or the
// interpretation of a signed generation, so previous caches must expire.
// v58 extends exact lossless BMP/TIFF/PNG image decisions through 512 pixels
// on each axis, after a 1,000-image larger-format differential corpus.
// v59 decides JPEG 2000 fuzzy-image rules false where the pinned reference
// cannot decode the image despite its GRAPHICS type.
// v60 computes icon edge measures and the reference's exact grayscale and
// colored confidence formula. Previously incomplete icon-group outcomes can
// now become detected or clean, so prior verdict caches must expire.
// v61 uses the project-owned JPEG decoder and reference color path. v62
// admits tested restart-marker JPEGs. v63 expands the guarded exact-decision
// profile to tested large JFIF and Adobe CMYK images; each change expires
// prior verdict caches.
// v64 decodes ScriptAsylum dF() JavaScript bodies and scans their normalized
// script view. Previously incomplete HTML may now detect or become clean.
// v65 decodes standalone Microsoft Script Encoder files and scans their
// generated HTML child; encoded scripts inside HTML can now yield a detection
// from the decoded body while unresolved page normalization remains incomplete.
// v66 decodes bounded six-argument Dean-Edwards JavaScript packer calls in
// the normalized HTML script view. Previously incomplete packer pages can
// now be detected or decided clean for the validated call shapes.
// v67 reproduces the reference's Script Encoder transitions inside HTML and
// ASP, including the second ASP pass over a chunk. Supported pages now scan
// their actual normalized views; truncated or unmodeled boundaries remain
// explicitly incomplete.
// v68 traverses standard ISO 9660 and Joliet directory trees, applying member
// metadata and recursive file scanning. It can turn an incomplete disc image
// into a detection or a supported clean verdict, so earlier caches expire.
// v69 also scans individual ISO multi-extent sections as the reference does.
// A section detection can now replace an earlier incomplete outcome.
// v80 checks Authenticode certificates against CRB blocklists. v81 may
// exempt a PE only after PKCS#7 signer, code-signing chain and PE digest
// verification; blocked certificates retain precedence.
// v82 validates legacy PKCS#9 countersignatures with a trusted time-signing
// chain before using signing time to exempt a PE whose code signer has expired.
// v83 loads CRB-verified signed catalog member hashes and exempts matching PE
// image or CAB file bytes; changed catalog trust can alter a prior verdict.
// v84 verifies signer chains against a trusted CRB RSA key even when the
// corresponding root certificate is absent from SignedData.
// v85 carries verified catalog signer-chain certificates into subsequent PE
// Authenticode checks, including PE files that omit an intermediate signer.
// v86 hashes 16-bit BMP icons in PE resource groups with the reference's
// color-word conversion and AND-mask handling.
// v87 detects hex-escaped common PDF dictionary names. v88 adds the
// default-on policy toggle to the cache identity. v89 decodes 32-bit PE icons
// whose alpha bytes are all zero using their AND mask or opaque fallback.
// v90 scans all PE icon resource languages under each icon ID. v91 restores
// transparent GIF colors from the first image's local palette for fuzzy
// hashing. v92 scans MBR boot sectors and conventional primary/extended
// partition windows, so a previous incomplete verdict can become a match.
// v93 validates GPT header/table checksums, uses the secondary table when
// needed, and scans supported GPT partition windows as children.
// v94 follows the reference's alternate EBR entry orderings, so a logical
// partition previously left undecidable can now detect.
// v95 adds the opt-in MBR/GPT partition-overlap alert. v96 makes the
// partition-count policy configurable and part of the cache identity.
// v97 ignores an invalid HTML base URL when no relative link depends on it;
// a previously incomplete benign page may now be clean.
// v98 suppresses VERSION_INFO matching when the PE resource directory size
// is zero, preventing a detection the reference does not make.
// v99 follows the reference's 16-bit resource-directory entry-count sum,
// preventing false version-info matches when malformed counts wrap to zero.
// v100 handles the full 16-bit PE section count in signature matching and
// recognizes reference-rejected early headers without incomplete offsets.
// v101 carries that bound through PE image and overlay analysis, so large
// section tables do not prevent embedded installer/archive inspection.
// v110 scans bounded XZ output when the final stream footer is missing or
// partial, after a virtual-footer preflight enforces dictionary limits.
// v111 matches the reference's fuzzy image hash on PNGs whose only decoding
// defect is a bad CRC on the terminal IEND chunk.
// v112 also matches truncated IEND CRCs and zero-length non-IDAT terminal
// chunks that the reference accepts as the end of a PNG image.
// v113 verifies MD5 PE image digests and accepts the reference's SHA/MD5
// Authenticode hash-OID aliases. v114 scans XAR TOCs and bounded heap extents.
// v115 extracts PDF-bearing base64 chunks from XDP XML. v116 scans decoded
// binData children in Microsoft Word/Workbook XML while retaining an
// incomplete result for unimplemented document metadata representations.
// v117 repairs unterminated numeric entities before Office XML parsing.
// v118 adds bounded ClamBC JSON query APIs and Office XML property parsing.
// v119 matches ClamAV's default metadata-off scan: JSON bytecode sees an
// inactive tree and returns -1 for queries. This removes false matches from
// programs that branch on json_is_active. The partial Office XML tree remains
// available only to explicit metadata callers until full collection exists.
// v120 orders PDF bytecode hooks as PRE, PARSED, POSTDUMP, END, answers
// pre-parse empty-object queries and bounded header/trailer flags, and keeps
// labels from unproven POSTDUMP/END reachability incomplete. Prior verdicts
// for custom PDF hooks may change, so signed generation and cache identity do.
// v121 runs POSTDUMP hooks on a reached, nonempty reference-extractable PDF
// stream after its child scan. A formerly incomplete custom-hook result can
// become a detection, so previous verdict caches must expire.
// v122 accepts the remaining PDF object API declarations, answers verified
// ordinary indirect-object fields at PARSED/POSTDUMP, and preserves PDF hook
// mutations as incomplete until their parser effects are fully represented.
// Custom-hook verdicts can change, so previous verdict caches must expire.
// v123 answers pdf_getobjflags for a reference-validated subset of complete
// independent PDF objects. Custom-hook verdicts can change, so prior caches
// must expire. Complex object state remains incomplete.
// v124 dispatches PDF END hooks when a simple PDF's only final parser flag
// is the reference-verified missing-trailer bit. A formerly incomplete
// custom-hook verdict can become detected, so cached verdicts must expire.
// v125 supplies proved flags for ordinary PDF stream objects. v126 decodes
// FSG 1.31 in the permissive PE profile and reports a proved strict PE
// layout error as Heuristics.Broken.Executable. Either path can change a
// prior verdict, so cached verdicts must expire. v127 follows the wider
// reference PE-header error boundary, including deferred broken-executable
// detection after embedded members and clean permissive rejection. v128
// decides icon-group rules after complete entries in a partial group and
// follows the reference's 100-entry default limit; verdicts can change.
// v139 decodes LArc -lz4-, -lzs- and -lz5- child members with bounded
// output, so archives that previously ended incomplete may now detect.
// v140 reads bounded LHA level-3 headers and four-byte extension chains,
// allowing previously incomplete members to reach the recursive scanner.
// v141 decodes LH1 adaptive-Huffman members and the exact OSK level-2
// terminator quirk, changing formerly incomplete archive verdicts.
// v142 scans bounded ALZ stored, bzip2 and deflate children, with explicit
// incomplete outcomes for encrypted and unsupported members.
// v143 scans ordinary non-solid EGG members. v144 scans top-level EGG archive
// comments, including valid UTF-8 multibyte comments.
// v145 validates EGG block CRCs and declared lengths while still scanning
// decoded child bytes from a damaged archive.
// v146 packs compiled signature offsets into bounded fields, reducing
// pattern storage without changing the parsed offset range or match logic.
// v147 moves matching graphs off literal-only body patterns. This changes
// matcher storage and requires a fresh release identity and corpus validation.
// v148 derives literal-only anchors while building the immutable index and
// keeps complex anchors only on their graphs until index publication.
// v151 stores uncommon rule descriptors and hooks outside the common rule
// record, preserving per-generation ownership during incremental refresh.
// v152 packs bounded hash-signature sizes into four bytes per index entry.
// v153 stores hash-rule names in a length-prefixed arena with four-byte
// sorted references, while preserving duplicate precedence and bounds.
// v154 parses bounded hash-signature fields without a per-line split array.
// v155 bounds temporary pattern and key storage to local arrays for common
// logical rules, and hashes offset/body keys without concatenating strings.
// v157 preflights authenticated container members and compiles text rules
// directly from one bounded member reader at a time.
// v158 uses bounded authenticated hash-member counts to reserve packed
// digest and name arrays before their large streaming compilation passes.
// v159 reserves compiled-pattern arrays from authenticated member counts;
// v160 reserves contiguous exact-hex literals while parsing each body.
// v162 deduplicates anchor literals through compact hash slots and exact
// byte comparison on collisions. The compiler layout and generation identity
// change, so prior cached verdicts must expire.
// v163 decides an unreadable icon-group data entry as a clean icon miss,
// matching the reference's group scanner. Formerly incomplete PE verdicts
// can change, so previous clean caches must expire.
// v167 and v168 decide chained GDB dot segments and deeply escaped URLs;
// prior clean verdicts must expire when these newly covered forms are scanned.
// v169 removes temporary anchor-key spelling allocations during index build.
// v170 aligns file-type numeric values with ClamAV 1.5.4 (including AI_MODEL),
// scans BINARY_DATA, PYTHON_COMPILED and AI_MODEL through target 14, and
// scans complete childless binary root metadata through legacy target 13.
// Both can change detection outcomes, so prior clean caches must expire.
// v171 aligns the built-in HTML type markers with the reference fallback
// table, preventing prefix-only matches such as <header> and recognizing
// additional image, object, table, link and JSP markers. It also types the
// built-in CLAMJSONv0 metadata magic as INTERNAL for local-only target-13
// feeds, and compiled Python headers for local-only target-14 feeds. These
// changes can alter prior verdicts.
// v172 passes the root file name and path to active metadata rules and
// prevents path-dependent verdicts from sharing hash-only scanner cache
// entries. Both changes can affect detection outcomes.
// v173 extends the proved scalar metadata serialization to childless plain
// ASCII text roots, and aligns file-name separators and raw Unicode string
// escaping with json-c. These can change target-13 outcomes and expire caches.
// v174 selects rules and reports engine/dconf level at the pinned 1.5.4
// functionality level 234. A new generation may activate different rules,
// and bytecode level queries can branch differently, so caches must expire.
// v175 aligns the bytecode environment's pinned macOS compiler/platform
// fields and Unix host strings. Hooks reading API 72 can change verdicts.
// v176 follows the pinned PE icon-resource walk across duplicate IDs while
// skipping non-directory types/names and nested language directories. Formerly
// incomplete icon-rule verdicts can now be detected or decided clean.
// v177 scans all languages of the first PE icon-group ID, carries the shared
// 100-entry budget across them, and sign-extends icon IDs before resource
// lookup. These can reveal detections or change formerly incomplete results.
// v178 evaluates static body rules before whole-file hashes and logical
// bytecode callbacks, changing some detection labels but not verdict states.
// v179 loads .pwdb and decrypts traditional PKWARE ZIP members, so a former
// incomplete result can now detect or be decided clean.
// v180 executes ClamBC hash-set, hex conversion and entropy APIs. A reached
// call can change an incomplete bytecode verdict, so old caches expire.
// v181 translates the bundled ClamAV ERE escape rules in container names and
// keeps directional word boundaries conditional. This can change archive
// member detections and previously clean verdicts, so old caches expire.
// v182 executes ClamBC fill_buffer for bounded input-window refills.
// v183 executes the ClamBC fixed and file-backed buffer-pipe APIs with
// bounded per-run storage and lifetime-checked virtual pointers. An
// unpacker program that previously stopped incomplete can now expose a child
// detection or a clean result, so old caches expire.
// v184 executes the ClamBC map APIs, including fixed and variable values,
// and corrects zero-size buffer-pipe allocation. Map-backed callbacks that
// previously stopped incomplete can now affect verdicts, so caches expire.
// v185-v187 add the ClamBC inflater, buffer-pipe JavaScript normalizer and
// shared HTML tokenizer. v188 adds version, platform and JIT host queries;
// v189 adds startup bytecode/JIT disable controls with generation-scoped
// suppression of ClamBC callbacks. These changes can alter verdicts, so
// previously compiled scanner caches expire.
// v200 makes metadata collection an explicit default-off scan mode. Root
// PRECLASS and target-13 activation, JSON query availability and the legacy
// bytecode option mask can change verdicts, so caches must expire.
// v201 accepts ClamBC wire API 101 with the pinned 1.5.4 implementation's
// observed zero result for documented option names. Previously unsupported
// bytecode programs can now be evaluated, changing verdicts and cache identity.
// v202 skips icon bitmaps whose non-wrapping declared header extends beyond
// the mapped file, matching the reference's failed pixel read. This turns a
// conditional unknown into a clean result on that input, so caches expire.
// v203 extends exact fuzzy-image decisions to GIFs up to 512 pixels
// on each axis. GIFs in the newly covered range can change verdicts,
// so scanner verdict caches expire.
// v204 preserves bytes after a completed ClamBC bzip2 or LZMA stream in
// the source pipe. A bytecode program can now act on the remaining input,
// so scanner verdict caches expire.
// v205 exposes the proved outer object table of a PDF with an object
// stream to PARSED bytecode hooks. Those hooks can now decide counts,
// lookups, offsets and sizes, so scanner verdict caches expire.
// v210 follows the reference's raw slash-name walk through bounded PDF
// dictionary strings, comments, arrays and nested dictionaries. ClamBC
// PDF flag queries can now decide those objects, so verdict caches expire.
// v237 hashes bounded BI_RGB ICOs with V2/V3/V4/V5 DIB headers. A graphics
// fuzzy rule on such an icon can change from incomplete to detected or clean,
// so previous verdict caches expire.
// v238 hashes bounded 16-bit BI_RGB and 40-byte BI_BITFIELDS ICOs with
// canonical masks. Newly decided graphics fuzzy rules expire old caches.
// v255 verifies bounded historical RSA certificate paths, resolves legacy
// countersigner collisions cryptographically, and retains independently CRB-
// authenticated catalog certificates. Prior trust-dependent verdicts expire.
// v256 owns the Authenticode SignedData wire parser and exact signed-byte
// verification for embedded signatures, catalogs and countersignatures.
// Metadata and signer-selection outcomes change, so cached verdicts expire.
// v257 validates image/catalog digest contexts and unsigned attribute uniqueness.
// v258 binds quarantine admission to verified snapshots and durable recovery.
// v262 retains CRB duplicate precedence and verifies local issuer time policy.
// v275 retains bzip2 decoder state across bytecode pipe windows. Newly
// reachable extracted children can change verdicts, so cached results expire.
// v276 adds streaming LZMA pipe decoding and native error cleanup. Newly
// reachable decoded children likewise invalidate earlier cached outcomes.
// v279 reproduces carried decoder lexer state and native buffer boundaries;
// earlier clean caches cannot represent the corrected normalized views.
// v278 uses one ordered decoder walk for eval, dF and named packers;
// this changes normalized views and requires invalidating earlier clean caches.
// v277 normalizes repeated named packer calls and decoded function scopes;
// changed normalized signature views invalidate previous cached results.
// v281 shares the scan-wide root JSON context across children and legacy
// metadata scans. Names can affect ordinary bytecode, so metadata collection
// requires path-aware scanning and bypasses content-only verdict caching.
const ImplementationVersion = buildinfo.ImplementationVersion

var ErrUnavailable = errors.New("UTMStack EDR: no compatible signature generation loaded")

// ErrLocalSignatures reports administrator files in the database directory
// that could not be loaded together with the active feed generation. The
// generation itself keeps serving; the local files are not partially loaded.
var ErrLocalSignatures = errors.New("UTMStack EDR: local signature files are not loaded with the feed generation")

// databaseLimits bound reading and compiling one database. They equal the
// compiler defaults that signatures.Load applies, so a feed generation
// combined with local files is held to the same bounds as one directory.
var databaseLimits = signatures.DefaultLimits()

const maxMetadataBytes = 64 << 10

// generation is one immutable database. meta always describes the signed feed
// generation (or the loose directory when none is active); db may also hold
// the administrator's local files compiled together with it.
type generation struct {
	db   *signatures.Database
	meta generationMetadata
}
type generationMetadata struct {
	Sequence       uint64 `json:"sequence"`
	ManifestHash   string `json:"manifest_hash"`
	ContentVersion string `json:"content_version"`
}

// Engine owns immutable snapshots. A recursive scan and its cache stamp always
// use the same generation. It never launches helpers or opens a scan socket.
type Engine struct {
	cfg        config.EDRConfig
	mu         sync.Mutex
	current    atomic.Pointer[generation]
	tuningOnce sync.Once
	tuning     Tuning
	slots      chan struct{}
}

func New(cfg config.EDRConfig) *Engine { return &Engine{cfg: cfg} }
func (e *Engine) Tuning() Tuning {
	e.tuningOnce.Do(func() {
		e.tuning = PlanTuning(e.cfg)
		// Keep the derivation reason, so a forced or reduced tier stays
		// visible in status.json.
		e.tuning.Reason = "native Go engine; bounded scans and atomic signature generations; " + e.tuning.Reason
		e.slots = make(chan struct{}, e.tuning.MaxThreads)
	})
	return e.tuning
}
func (e *Engine) DatabaseDir() string {
	if e.cfg.SignatureDBDir != "" {
		return e.cfg.SignatureDBDir
	}
	return filepath.Join(config.EngineDir, "database")
}
func (e *Engine) EnsureRunning() error {
	e.Tuning()
	if e.current.Load() != nil {
		return nil
	}
	return e.Reload()
}

// Reload retains the last working snapshot on failure. Loose files are trusted
// administrator input; remote generations use CURRENT and immutable directories.
// Loose files in the same directory are compiled together with the active
// generation. When they cannot be, the generation alone is loaded and an
// ErrLocalSignatures error reports them.
func (e *Engine) Reload() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.reloadLocked()
}

func (e *Engine) reloadLocked() error {
	root := e.DatabaseDir()
	meta, persisted, err := e.durableStateLocked()
	if err != nil {
		return err
	}
	if !persisted {
		// A loose database has no signed generation pointer. Verify every
		// selected file's current bytes before reusing its already-compiled
		// matcher; a content change still takes the full compile path.
		if old := e.current.Load(); old != nil && old.meta.Sequence == 0 && old.meta.ManifestHash == "" {
			version, err := signatures.DirectoryContentVersion(root, databaseLimits)
			if err != nil {
				return fmt.Errorf("loading native signatures: %w", err)
			}
			if version == old.db.Version() {
				return nil
			}
		}
		var prior *signatures.Database
		if old := e.current.Load(); old != nil && old.meta.Sequence == 0 && old.meta.ManifestHash == "" {
			prior = old.db
		}
		db, err := signatures.LoadWithPrior(root, databaseLimits, prior)
		if err != nil {
			return fmt.Errorf("loading native signatures: %w", err)
		}
		if db.DetectionCount() == 0 {
			return ErrUnavailable
		}
		meta.ContentVersion = db.Version()
		return e.publishLocked(&generation{db: db, meta: meta}, nil)
	}
	storedDir := filepath.Join(root, "generations", generationID(meta))
	// The serving signed matcher may be reused only after re-reading the
	// authenticated generation's exact signature bytes. Stream their content
	// identity so a routine reload does not retain another full source copy.
	if old := e.current.Load(); old != nil && old.meta == meta && old.db.Version() == meta.ContentVersion {
		if empty, err := noLocalSignatureEntries(root); err != nil {
			return err
		} else if empty {
			version, err := signatures.DirectoryContentVersion(filepath.Join(storedDir, "signatures"), databaseLimits)
			if err != nil {
				return fmt.Errorf("loading native signatures: %w", err)
			}
			if version != meta.ContentVersion {
				return errors.New("signature generation content mismatch")
			}
			return nil
		}
	}
	files, err := loadGeneration(storedDir)
	if err != nil {
		return fmt.Errorf("loading native signatures: %w", err)
	}
	version, err := signatures.ContentVersion(files, databaseLimits)
	if err != nil {
		return err
	}
	if version != meta.ContentVersion {
		return errors.New("signature generation content mismatch")
	}
	var prior *signatures.Database
	if old := e.current.Load(); old != nil {
		prior = old.db
	}
	db, localErr, err := withLocalFiles(root, files, nil, prior)
	if err != nil {
		return err
	}
	return e.publishLocked(&generation{db: db, meta: meta}, localErr)
}

// noLocalSignatureEntries checks names only. Any candidate, including a
// non-regular entry that compilation must reject, disables the fast path.
func noLocalSignatureEntries(root string) (bool, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !generationStoreEntry(entry) && !signatureStateFile(entry.Name()) {
			return false, nil
		}
	}
	return true, nil
}

// publishLocked makes g the scanning snapshot unless it would roll back the
// one already serving. localErr is returned after a successful publish.
func (e *Engine) publishLocked(g *generation, localErr error) error {
	if old := e.current.Load(); old != nil && (g.meta.Sequence < old.meta.Sequence ||
		(g.meta.Sequence == old.meta.Sequence && g.meta.ManifestHash != old.meta.ManifestHash)) {
		return errors.New("persisted signature generation rollback rejected")
	}
	e.current.Store(g)
	return localErr
}

// durableStateLocked reads the rollback state that CURRENT and its metadata
// record, without loading signature content. A stored generation whose
// content no longer loads (damage, or a stricter compiler after an upgrade)
// still binds anti-rollback, and a newer signed generation can replace it.
// No CURRENT is an empty state; a damaged pointer or metadata fails closed.
func (e *Engine) durableStateLocked() (generationMetadata, bool, error) {
	var meta generationMetadata
	root := e.DatabaseDir()
	id, err := readBoundedFile(filepath.Join(root, "CURRENT"), maxMetadataBytes)
	if errors.Is(err, os.ErrNotExist) {
		return meta, false, nil
	} else if err != nil {
		return meta, false, err
	}
	name := strings.TrimSpace(string(id))
	if !validDigest(name) {
		return meta, false, errors.New("invalid signature generation pointer")
	}
	b, err := readBoundedFile(filepath.Join(root, "generations", name, "metadata.json"), maxMetadataBytes)
	if err != nil {
		return meta, false, err
	}
	if err := json.Unmarshal(b, &meta); err != nil {
		return meta, false, err
	}
	if generationID(meta) != name {
		return meta, false, errors.New("signature generation metadata mismatch")
	}
	return meta, true, nil
}

// loadGeneration reads the bounded signature files of one stored generation.
// Content identity and compilation are separate: an integrity check must not
// construct another complete matcher for bytes already compiled on activation.
func loadGeneration(dir string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	budget := databaseLimits.MaxTotalBytes
	if err := readSignatureDir(filepath.Join(dir, "signatures"), nil, &budget, files); err != nil {
		return nil, err
	}
	return files, nil
}

// storedGenerationMatches reports whether dir holds exactly the generation
// that meta describes.
func storedGenerationMatches(dir string, meta generationMetadata) bool {
	b, err := readBoundedFile(filepath.Join(dir, "metadata.json"), maxMetadataBytes)
	var prior generationMetadata
	if err != nil || json.Unmarshal(b, &prior) != nil || prior != meta {
		return false
	}
	files, err := loadGeneration(dir)
	if err != nil {
		return false
	}
	version, err := signatures.ContentVersion(files, databaseLimits)
	return err == nil && version == meta.ContentVersion
}

// withLocalFiles compiles administrator files placed directly in the database
// directory together with the feed generation, as one database, which is what
// the documented local workflow expects. When the local files cannot be
// combined (unsupported rules, a name shared with a feed artifact, limits),
// none of them load: the generation alone is returned with an
// ErrLocalSignatures error, so they are neither silently ignored nor able to
// block the signed feed.
func withLocalFiles(root string, gen map[string][]byte, prepared map[string]*signatures.PreparedHashList, prior *signatures.Database) (*signatures.Database, error, error) {
	budget := databaseLimits.MaxTotalBytes
	for _, b := range gen {
		budget -= int64(len(b))
	}
	local := make(map[string][]byte)
	readErr := readSignatureDir(root, generationStoreEntry, &budget, local)
	if readErr != nil {
		local = nil // a partial local directory must never be loaded
	}
	db, localErr, err := signatures.CompileWithPreparedLocalFilesAndPrior(gen, local, prepared, databaseLimits, prior)
	if errors.Is(err, signatures.ErrNoDetections) {
		return nil, nil, ErrUnavailable
	}
	if err != nil {
		return nil, nil, fmt.Errorf("compiling native signatures: %w", err)
	}
	if readErr != nil {
		localErr = readErr
	}
	if localErr != nil {
		localErr = fmt.Errorf("%w: %v", ErrLocalSignatures, localErr)
	}
	return db, localErr, nil
}

// generationStoreEntry names the database directory entries that belong to the
// generation store rather than to the administrator's loose files.
func generationStoreEntry(entry os.DirEntry) bool {
	switch entry.Name() {
	case "CURRENT":
		return !entry.IsDir()
	case "generations":
		return entry.IsDir()
	}
	return false
}

// readSignatureDir adds the signature files of one directory to files, under
// the rules of signatures.Load: dot files and updater, lock and configuration
// state are skipped; links, other subdirectories and oversized content are
// rejected. budget is shared across calls, so a combined database is bounded
// as one.
func readSignatureDir(dir string, skip func(os.DirEntry) bool, budget *int64, files map[string][]byte) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if (skip != nil && skip(entry)) || signatureStateFile(name) {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() {
			return fmt.Errorf("UTMStack EDR signature path is not a regular file: %s", name)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("UTMStack EDR invalid signature file: %s", name)
		}
		limit := min(databaseLimits.MaxFileBytes, *budget)
		if info.Size() > limit {
			return errors.New("UTMStack EDR signature database size limit exceeded")
		}
		b, err := readBoundedFile(filepath.Join(dir, name), limit)
		if err != nil {
			return err
		}
		*budget -= int64(len(b))
		files[name] = b
		if len(files) > databaseLimits.MaxFiles {
			return errors.New("UTMStack EDR signature file count limit exceeded")
		}
	}
	return nil
}

// signatureStateFile mirrors the names signatures.Load skips.
func signatureStateFile(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json", ".dat", ".log", ".lock", ".tmp", ".conf":
		return true
	}
	return false
}

func readBoundedFile(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > limit || uint64(info.Size()) >= uint64(^uint(0)>>1) {
		return nil, errors.New("UTMStack EDR signature database size limit exceeded")
	}
	// Use the validated size once, and reject a concurrent size change. This
	// avoids ReadAll's repeated large backing arrays for full feed generations.
	b := make([]byte, int(info.Size()))
	if _, err := io.ReadFull(f, b); err != nil {
		return nil, err
	}
	var extra [1]byte
	if n, err := io.ReadFull(f, extra[:]); n != 0 || err != io.EOF {
		return nil, errors.New("UTMStack EDR signature file changed while reading")
	}
	return b, nil
}
func validDigest(v string) bool {
	b, err := hex.DecodeString(v)
	return err == nil && len(b) == sha256.Size && strings.ToLower(v) == v
}
func generationID(meta generationMetadata) string {
	b, _ := json.Marshal(meta)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func (e *Engine) Version() string {
	if g := e.current.Load(); g != nil {
		return e.versionFor(g)
	}
	return ""
}

// MetadataSensitive reports whether the active scan can depend on its path.
// Ordinary logical bytecode and parser hooks can read root names as well as
// PRECLASS/target-13 rules. Conservatively bypass hash-only verdict caching
// whenever metadata collection is enabled, including after a generation reload.
func (e *Engine) MetadataSensitive() bool {
	if g := e.current.Load(); g != nil {
		return e.cfg.CollectMetadata
	}
	return false
}

func (e *Engine) versionFor(g *generation) string {
	t := e.Tuning()
	return fmt.Sprintf("%s:%s:%d:%d:%d:%d:%d:%d:pdfname=%t:partitionintersection=%t:brokenexecutable=%t:collectmetadata=%t", ImplementationVersion, g.db.Version(),
		t.MaxFileSizeMB, t.MaxScanSizeMB, t.MaxRecursion, t.MaxFiles, t.MaxPartitions, t.MaxScanTimeMS,
		e.cfg.ScanHeuristics.PDFObfuscatedNameOn(), e.cfg.ScanHeuristics.PartitionIntersectionOn(), e.cfg.ScanHeuristics.BrokenExecutableOn(), e.cfg.CollectMetadata)
}
func (e *Engine) SignatureCount() int {
	if g := e.current.Load(); g != nil {
		return g.db.Count()
	}
	return 0
}

// LoadReport describes how the serving generation loaded: rules skipped as
// the reference skips them, file-type entries, and rules that are unknown
// when reached. It is empty when no generation is loaded.
func (e *Engine) LoadReport() signatures.Report {
	if g := e.current.Load(); g != nil {
		return g.db.Report()
	}
	return signatures.Report{}
}

// UnknownRules counts, by reason, the serving generation's rules that need
// analysis this engine does not perform. Each makes an object incomplete only
// where it could apply.
func (e *Engine) UnknownRules() map[string]int { return e.LoadReport().Unknown }
func (e *Engine) FeedState() (uint64, string) {
	if g := e.current.Load(); g != nil {
		return g.meta.Sequence, g.meta.ManifestHash
	}
	return 0, ""
}

// ActivateGeneration commits all content and rollback metadata before publishing
// CURRENT. Callers must authenticate remote content before invoking it.
func (e *Engine) ActivateGeneration(files map[string][]byte, sequence uint64, manifestHash string) (string, error) {
	return e.activateGeneration(files, nil, sequence, manifestHash)
}

// ActivatePreparedGeneration consumes already-decoded hash-list sources after
// the feed has verified its signed manifest and artifact checksums. Prepared
// inputs are bound to the same source bytes that are committed durably.
func (e *Engine) ActivatePreparedGeneration(files map[string][]byte, prepared map[string]*signatures.PreparedHashList, sequence uint64, manifestHash string) (string, error) {
	return e.activateGeneration(files, prepared, sequence, manifestHash)
}

func (e *Engine) activateGeneration(files map[string][]byte, prepared map[string]*signatures.PreparedHashList, sequence uint64, manifestHash string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if sequence == 0 || !validDigest(manifestHash) {
		return "", errors.New("invalid signature generation metadata")
	}
	// Serialize writers across processes and refresh durable replay state, even
	// if this Engine instance initialized before another writer's activation.
	root := e.DatabaseDir()
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	unlock, err := lockUpdate(filepath.Join(root, ".update-lock"))
	if err != nil {
		return "", fmt.Errorf("signature update lock unavailable: %w", err)
	}
	defer unlock()
	// The persisted sequence and manifest bind even when the stored content no
	// longer loads, so a damaged generation cannot block a newer signed one.
	durable, persisted, err := e.durableStateLocked()
	if err != nil {
		return "", fmt.Errorf("reading signature rollback state: %w", err)
	}
	var priors []generationMetadata
	if persisted {
		priors = append(priors, durable)
		if old := e.current.Load(); old == nil || generationID(old.meta) != generationID(durable) {
			// Best effort: serve the persisted generation if it still loads.
			_ = e.reloadLocked()
		}
	}
	if old := e.current.Load(); old != nil {
		priors = append(priors, old.meta)
	}
	for _, prior := range priors {
		if sequence < prior.Sequence || (sequence == prior.Sequence && manifestHash != prior.ManifestHash) {
			return "", errors.New("signature generation rollback or equivocation rejected")
		}
	}
	for name := range files {
		if name == "" || name == "." || strings.ContainsAny(name, "/\\:") || filepath.Base(name) != name {
			return "", errors.New("invalid signature artifact name")
		}
	}
	contentVersion, err := signatures.ContentVersion(files, databaseLimits)
	if err != nil {
		return "", fmt.Errorf("identifying native signatures: %w", err)
	}
	for _, prior := range priors {
		if sequence == prior.Sequence && contentVersion != prior.ContentVersion {
			return "", errors.New("signature generation content changed without advancing sequence")
		}
	}
	meta := generationMetadata{Sequence: sequence, ManifestHash: manifestHash, ContentVersion: contentVersion}
	var prior *signatures.Database
	if old := e.current.Load(); old != nil {
		prior = old.db
	}
	db, localErr, err := withLocalFiles(root, files, prepared, prior)
	if err != nil {
		return "", err
	}
	g := &generation{db: db, meta: meta}
	parent := filepath.Join(root, "generations")
	if err := os.MkdirAll(parent, 0700); err != nil {
		return "", err
	}
	stage, err := os.MkdirTemp(parent, ".staging-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stage)
	if err := os.Mkdir(filepath.Join(stage, "signatures"), 0700); err != nil {
		return "", err
	}
	for name, data := range files {
		if err := durableWrite(filepath.Join(stage, "signatures", name), data); err != nil {
			return "", err
		}
	}
	b, _ := json.Marshal(meta)
	if err := durableWrite(filepath.Join(stage, "metadata.json"), b); err != nil {
		return "", err
	}
	if err := syncDirectory(filepath.Join(stage, "signatures")); err != nil {
		return "", err
	}
	if err := syncDirectory(stage); err != nil {
		return "", err
	}
	id := generationID(meta)
	final := filepath.Join(parent, id)
	if _, err := os.Stat(final); errors.Is(err, os.ErrNotExist) {
		if err := os.Rename(stage, final); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else if !storedGenerationMatches(final, meta) {
		// A damaged copy of this exact signed generation (same sequence,
		// manifest and content digest) is replaced by the verified staged copy.
		// It is moved aside for investigation, never deleted.
		aside := filepath.Join(parent, fmt.Sprintf(".damaged-%s-%d", id[:16], time.Now().UnixNano()))
		if err := renameDirectory(final, aside); err != nil {
			return "", fmt.Errorf("replacing damaged stored generation: %w", err)
		}
		if err := renameDirectory(stage, final); err != nil {
			// Put the damaged copy back, so CURRENT still names readable
			// rollback state and a later activation can retry the repair.
			if restoreErr := renameDirectory(aside, final); restoreErr != nil {
				return "", fmt.Errorf("replacing damaged stored generation: %w (restoring it: %v)", err, restoreErr)
			}
			return "", fmt.Errorf("replacing damaged stored generation: %w", err)
		}
	}
	if err := syncDirectory(parent); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(root, ".current-")
	if err != nil {
		return "", err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.WriteString(id + "\n"); err != nil {
		f.Close()
		return "", err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = replaceFile(name, filepath.Join(root, "CURRENT")); err != nil {
		return "", err
	}
	e.current.Store(g)
	if localErr != nil {
		logger.Error("%v", localErr)
	}
	if err = syncDirectory(root); err != nil {
		return "", err
	}
	return e.Version(), nil
}

// renameDirectory moves generation directories; tests replace it to simulate
// a rename that Windows refuses while another process holds a file.
var renameDirectory = os.Rename

func durableWrite(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
func (e *Engine) Supervise(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if e.current.Load() == nil {
				_ = e.EnsureRunning()
			}
		}
	}
}
func (e *Engine) Stop() error { e.current.Store(nil); return nil }
