# ClamBC sandboxed execution core

The official GandCrab program now has six inert extracted-child controls:
modified RC4, LCG and TEA with keys in `.data` or `.rdata`. Go and pinned
ClamAV 1.5.4 detect the exact same 65,800-byte child; corrupt-resource and
absent/disabled-hook controls remain clean. Five write-producing programs
activatable at level 234 now have selected positive output checks. The
v339 ELF writer is restricted to levels 110–119 and is excluded at 234.
This does not qualify every path or a future/custom feed. See the
[v241 audit](../../../docs/edr/native-engine/audits/2026-11-07-v241-quantum/REPORT.md).

The authenticated MPRESS and LizardNest programs now have positive inert
extracted-child differentials against pinned ClamAV 1.5.4. Go and ClamAV
detect the same 5,120-byte MPRESS child and 4,096-byte LizardNest child by
exact child-only HDB rules. Disabled-hook and broken/short-input controls
stay clean. MPRESS's LZMA decoder executes inside the official bytecode;
the host LZMA API is qualified separately in the v276 section below. See the
[v240 output evidence](../../../docs/edr/native-engine/audits/2026-11-07-v240-word6-macros/REPORT.md).

The authenticated hta2exe program also has a positive inert-resource
child-output differential against pinned ClamAV 1.5.4. The 256-byte child
is the bitwise inverse of the input resource; broken and disabled-hook
controls stay clean. See the [hta2exe output audit](../../../docs/edr/native-engine/audits/2026-11-01-official-hta2exe-output/REPORT.md).

An authenticated script2exe unpacker now has a positive inert-resource
output control: both Go and pinned ClamAV 1.5.4 scan its same 64-byte
child, while a broken resource path remains clean. A separate synthetic
program checks two distinct outputs in one run. See the
[official-output audit](../../../docs/edr/native-engine/audits/2026-11-01-official-script2exe-output/REPORT.md).
Other functionality windows, future/custom output-producing programs and
untested branches still require positive extracted-child validation.

Wire API 101 (`engine_scan_options_ex`) now passes declaration and bounded
pointer validation. The pinned ClamAV 1.5.4 source writes each lowercased
option-name byte to element zero and uses inverted category comparisons;
native harmless probes return zero for enabled documented names. The Go VM
returns the same observed result, without claiming the intended option query
works ([v201 audit](../../../docs/edr/native-engine/audits/2026-11-01-v201-clambc-options-ex/REPORT.md)).

Metadata collection is controlled by the EDR's default-off
`collect_metadata` setting, matching ClamAV 1.5.4 without/with `--gen-json`.
Bytecode API 68 reports `CL_SCAN_FILE_PROPERTIES` only in the enabled mode;
JSON API 92 follows that mode in raw, normalized and parser-hook callbacks.
The JSON tree is partial and queries for unproved fields stop incomplete
([v200 audit](../../../docs/edr/native-engine/audits/2026-11-01-v200-metadata-mode/REPORT.md)).

Wire APIs 102–104 use a Go LZMA-alone streaming decoder across bounded input
and output windows. Initialization still requires and consumes the 13-byte
header. Range models, dictionary history, pending matches and the small input
lookahead survive continuation. Following bytes remain available after stream
end. Errors publish the observed cursors and close the handle, as in the native
bytecode adapter. Dictionary and probability storage each share a 64 MiB ceiling
per VM, alongside existing pipe, work, output and handle-count bounds.
The [v276 reference](../../../docs/edr/native-engine/audits/2026-11-17-v276-lzma-stream/REPORT.md)
compares 2,842,278 decisions across 857 inputs and six window configurations;
24 full-scanner controls agree with ClamAV, including ten formerly incomplete
positive controls. A native EOS marker can end a stream before its declared
size is exhausted; the declared size remains an output ceiling. The dated
[v194](../../../docs/edr/native-engine/audits/2026-11-01-v194-clambc-lzma/REPORT.md)
and [v204](../../../docs/edr/native-engine/audits/2026-11-01-v204-bytecode-stream-tail/REPORT.md)
reports preserve earlier complete-window scope. See [LZMA-NOTICE.md](LZMA-NOTICE.md).

Wire APIs 105–107 use an independent Go bzip2 state machine across bounded
input/output pipe windows. Bit, Huffman, move-to-front, BWT, RLE, legacy
randomization and CRC state survive continuation. The first stream stops
without consuming trailing input. Block reservations are bounded to 64 MiB
per VM, in addition to the existing pipe, output, work and stream-count limits.
The retained native adapter corpus compares 1,809,696 call decisions across
265 inputs and seven window configurations. Full-engine inert child controls
also agree with pinned ClamAV 1.5.4. Re-entry after a data error remains
explicitly incomplete; this does not qualify every corrupt-stream recovery
path or broader bytecode/unpacker parity. The original complete-stream
[v193](../../../docs/edr/native-engine/audits/2026-11-01-v193-clambc-bzip2/REPORT.md)
and trailing-input [v204](../../../docs/edr/native-engine/audits/2026-11-01-v204-bytecode-stream-tail/REPORT.md)
reports retain their dated scope. The legacy table attribution is in
[BZIP2-NOTICE.md](BZIP2-NOTICE.md).

The root file scanner also uses this decoder through `Bzip2MemberReader`.
It publishes fixed 8192-byte packets, discards the current unflushed packet on
error, and charges discarded output and read-ahead to the expansion budget.
Consecutive members are scanned separately, preserving whole-child hash
boundaries and later-member detections; native ClamAV stops at the first member.
Cancellation, close and terminal errors release decoder storage. This scope
covers root `BZh` dispatch; ZIP, ALZ, EGG and installer-specific readers retain
their existing behavior. See the [v282 member-boundary audit](../../../docs/edr/native-engine/audits/2026-11-17-v282-bzip2-members/REPORT.md)
for the native comparisons, regression replay and remaining limits.

Wire API 91 (`get_file_reliability`) returns the proved file reliability
value for a normal root scan or root preclassification, and ClamAV's `3`
when no scanner context exists. A nested or reconstructed representation
without a proved value stops incomplete. Four harmless ClamAV/Go scan cases
agree in the [v192 oracle](../../../docs/edr/native-engine/audits/2026-11-01-v192-file-reliability/REPORT.md).

Wire API 70 (`extract_set_container`) now carries a requested file type into
the recursive scan of bytecode output. The Go engine uses its forced-type
child path, so the requested type can select the child's typed signature
root. A harmless `CL_TYPE_OTHER` child rule and its absent/invalid-type
controls match ClamAV 1.5.4 in the [v191 oracle](../../../docs/edr/native-engine/audits/2026-11-01-v191-container-type/REPORT.md).
Unrecognized numeric types, `PART_ANY`, `ERROR`, and forced types whose
content needs an unmodeled parser path remain incomplete rather than clean.

Wire API 71 (`input_switch`) now lets bytecode read a snapshot of its
extracted output and switch back to the original object. The VM updates the
read-only file-size global, preserves the file cursor, makes existing
file-backed pipes follow the active input, and retires pointers into the
extracted mapping when it is released. Switching without output returns
`-1`; `extract_new` while the extracted mapping is active remains incomplete
because truncating a mapped temporary file has platform-dependent behavior.
A harmless serialized ClamBC rule agrees with the pinned ClamAV 1.5.4
interpreter on a detection and its clean control ([v190 audit](../../../docs/edr/native-engine/audits/2026-11-01-v190-input-switch/REPORT.md)).

Wire APIs 43–45 normalize JavaScript from file-backed or memory buffer pipes.
The tokenizer is shared with HTML script normalization and carries lexical
state across process calls. `jsnorm_done` queues the generated `<script>`
file, which enters recursive child scanning after the VM run. Unmodeled
transformations and missing child scanners stop incomplete. Normalizer count
and aggregate input/output bytes are bounded per run.

The v279 continuation models carried identifier/number/string/comment state,
native exponent and invalid-backslash lookahead, string normalization per input
buffer, trailing-plus publication and line-comment end behavior. The original
2,247-page oracle now has no incomplete pages or differences; 2,846 new native
byte controls and 102 actual full-scanner controls agree. Six prior false-clean
controls now detect. Calls without six collected packer arguments follow the
native no-op; collected but unproved argument shapes and budget refusals still
stop incomplete. See the
[v279 audit](../../../docs/edr/native-engine/audits/2026-11-17-v279-js-carried-state/REPORT.md).
Earlier carried-state limitations below describe the dated v278/v277 checkpoints.

The shared normalizer now processes packer, dF and eval calls in one forward
token walk, including nested calls introduced by decoded text and the native
function-end markers. Decoded declarations are processed immediately. A 20 MiB
aggregate decoded-byte ceiling, 4,096 transform calls and 8,388,608 charged token
work units bound the walk; packer function/call limits also apply. Unproved
carried lexical states and malformed/over-budget transforms remain incomplete.
The [v278 audit](../../../docs/edr/native-engine/audits/2026-11-17-v278-js-decoder-order/REPORT.md)
retains 363 new native byte controls, 136 actual complete-scanner controls and
unchanged older oracle identities. It corrects six prior false-clean controls.
The earlier v277 mixed-decoder limitation below is historical.

Repeated calls to one named Dean-Edwards six-argument unpacker follow the
pinned native concatenation and replacement range, including its adjacent-call
skip behavior. Decoded functions have their own declaration scopes. The shared
normalizer bounds decoded payloads to 20 MiB and recognized calls to 128;
unproved carried-state paths remain incomplete. The
[v277 audit](../../../docs/edr/native-engine/audits/2026-11-17-v277-js-named-calls/REPORT.md)
retains 204 new native byte controls, unchanged earlier oracle results and
35 actual HTML/ClamBC full-scanner comparisons. This qualifies those controls,
not all JavaScript normalization semantics or native endpoint delivery.

Wire APIs 39–41 now decode raw DEFLATE, zlib and gzip streams through the
per-run buffer pipes. The Go decoder preserves bit, Huffman, history,
checksum and wrapper state across input and output windows. Automatic
zlib/gzip detection, optional gzip headers and tiny pipe windows have
synthetic tests. A serialized program's decoded output reaches the recursive
child scanner. Corrupt-stream `inflateSync` recovery searches for full-flush
markers across pipe windows and returns zlib-compatible per-call statuses.
The test-only C zlib adapter checks cursor, status and output parity on valid
and selected corrupt streams. A preset-dictionary stream returns zlib's
`Z_NEED_DICT` status and exact cursor, but ClamBC has no dictionary setter
to resume it. The number of inflaters is
bounded to 256 per run. Native differential ClamBC execution is still needed
before claiming full unpacker parity.

Wire APIs 9–14 accept ClamAV's private trace calls under its normal
trace-disabled scanner profile. The VM still validates the first pointer
against the second argument before each no-op call, as the reference does.

Wire APIs 57–64 now provide per-run maps with fixed or variable-length
values, source-order `last_insert` and `last_find` state, bounded key storage
and VM-owned value pointers. Fixed values occupy one contiguous virtual
block; growing it retires prior pointers, as a moving `realloc` can. Removed
fixed values are zeroed. Dynamic replacement and `map_done` retire pointers.
Invalid map IDs follow the reference return codes; use of a destroyed
non-last slot faults. This closes another host-API gap but does not complete
ClamBC decompression or all JavaScript normalization semantics.

Wire APIs 30–38 now execute fixed and file-backed buffer pipes. Pipe storage
and cursors are private to a VM run; memory-backed pointers retire at `done`,
while file-backed reads borrow a read-only view of the immutable input
snapshot. Pipe count is bounded to 4,096 and fixed storage counts against the
32 MiB live virtual-data cap. Reached use after `done` faults rather than
reinterpreting a freed buffer as a file pipe. Inflate, bounded bzip2 and
bounded LZMA consume these pipes. The bzip2 and LZMA continuation controls above
do not qualify every codec profile or complete ClamBC unpacker parity.

The JSON query APIs (wire IDs 92–100) have bounded per-run object handles,
typed values, and VM memory checks. When collection is enabled, every raw
object, extracted child, normalized view and parser hook reads the same
scan-wide root tree, as in the native API. The scanner records root identity,
path, type and SHA-256 before analysis; a root type promotion updates that
shared tree. `RootFileType` is added after root analysis, before PRECLASS.
Kind-260 hooks run only for root-layer triggers or unconditional hooks.
An Office metadata-parser error preserves proved root fields and still permits
independent hooks to detect. A no-hit hook retains the incomplete verdict;
unproved parser subtrees are omitted. The [v284 error-ordering audit](../../../docs/edr/native-engine/audits/2026-11-17-v284-preclass-parser-error/REPORT.md)
retains the native positive and negative controls.
PRECLASS output starts at the first child depth and shares the root ancestry.
Metadata-enabled file scans always supply the path and bypass content-only
verdict caching, including for ordinary logical bytecode that reads a name.

The legacy target-13 JSON pass suspends new metadata collection and clears
its FILE_PROPERTIES option bit, while its JSON APIs continue reading the
existing root. That nested pass retains the original root as its parent.
The proved complete serialization remains limited to childless binary and
plain-ASCII scalar roots. PRECLASS supplies proved HWPML, Word XML and
Workbook XML property subtrees, but other parser fields, child/embedded trees,
indicators and optional hashes remain partial. A reached lookup for an absent
field in that partial tree stays undecided. With collection disabled and no
root tree, `json_is_active` returns zero and the other JSON APIs return -1.
This is partial metadata coverage; it does not establish complete PRECLASS
or parser-bearing target-13 parity.

The signature compiler accepts BC_STARTUP (kind 1) programs without a
logical trigger. It runs the first loaded startup program through the bounded
Go interpreter before publishing a generation, including an incremental CVD
refresh. Wire APIs 73 and 74 can request full bytecode disablement or JIT
disablement. A full-disable request suppresses all bytecode callbacks in that
generation, including unsupported callbacks; static signatures continue to
scan. A JIT-disable request leaves the Go interpreter active. A failed or
unsupported run rejects that generation; a non-magic return is accepted,
matching ClamAV's normal-mode warning behavior. The
builtin ClamAV startup program is not used because this engine does not use
ClamAV's JIT or interpreter. The official v339 bytecode feed has no startup
program, so this path is validated with synthetic CBC fixtures.

The authenticated `bytecode.cvd` v339 snapshot (SHA-256
`6d4aa01f219e988060fc419f495d07f27e0cdf1a2cccc065971da922c76f7ffb`)
contains 85 ClamBC programs: 74 logical callbacks (kind 256), six PE
unpacker hooks (257), two PDF hooks (258), two all-PE hooks (259) and one
ELF unpacker hook (261). Decoding each serialized header found no startup,
preclass or Mach-O hook in this snapshot. Their absence from this feed does
not remove the need to implement them for compatible custom or future feeds.

This package implements a bounded Go parser and interpreter for a defined subset of ClamBC formats 6 and 7. The signature database can activate supported logical callbacks (kind 256), PE hooks (kinds 257 and 259), PDF phase hooks (kind 258), root preclass hooks (kind 260), and ELF/Mach-O unpacker hooks (kinds 261 and 262), subject to their trigger and entrypoint checks. Executable hooks may run with or without a logical trigger. Unsupported declarations reject standalone parsing; the signature database retains unsupported feed programs as conditional unknowns. A reached output-producing operation needs the enclosing scanner's child callback.

Implemented operations include fixed-width arithmetic, comparisons, division, bit operations, typed integer casts, branches, returns, direct calls, copy, byte swaps, endianness queries, abort, load/store, GEP1/GEPZ, memory fill/copy/move/compare, pointer differences and pointer-to-integer conversion. Types include arrays, both struct kinds, pointers and function prototypes. Global initializers follow the pinned interpreter: integer and struct globals stay zero, arrays with 1, 2, 4 or 8-byte elements receive their first raw initializer components (so pointer arrays hold raw numbers), and top-level pointer globals are composed, including forward/self references and read-only file-size, kind, match-count, first-match-offset and PE metadata values. The interpreter turns a zero pointer global into a pointer to its shared global block, which has no native equivalent; reading that value is a runtime fault. The legacy virus-name global remains unsupported; ClamAV 1.5.4's interpreter does not register it either.

The supported host APIs provide input read/seek/search/byte access, bounded `fill_buffer` input-window refills (ID 21), output writes, copied virus labels, bounded diagnostics, the basic API test, bounded per-run hash sets (IDs 24–29), two-digit hex conversion, buffer entropy, integer-scaled logarithm/power/exponential/trigonometric operations, and the engine functionality and dconf level queries, which both report `EngineLevel` (234, the level of pinned reference release 1.5.4). APIs 68 and 69 report the enclosing scanner's ClamBC scan/database option masks from an explicit `EngineOptions` profile. A generic VM run without that profile stops at a reached query with `ErrUnsupported`. APIs 75, 76 and 90 compare version strings, test the platform record with per-field wildcards, and report that this Go interpreter has no JIT. The VM recognizes `extract_new` and scans bytecode output through the enclosing scanner's bounded recursive-child callback, both on explicit member transitions and the final implicit flush. API 89 (`matchicon`) validates its pointer and PE preconditions, then reproduces ClamAV 1.5.4's source-proven no-match result for valid calls; that release constructs the API's resource record without the data-directory count its resource walker requires. Without the child callback a reached output operation stops with `ErrExtraction`; no output is silently discarded. Declarations must match the wire ID, name and complete prototype. Every unsupported declaration fails even if unused; unsupported instructions fail even if unreachable. `RunWithInput` uses private memory and an input snapshot, and returns owned output/diagnostic snapshots only on successful completion. It has no host filesystem, network, process, native-pointer, JIT or external-runtime access.

Virtual pointers contain a region slot, the slot's generation and a byte offset, never a Go address. Reads check bounds and per-byte initialization. Writes cannot change the file-size hook; stack pointers become invalid when their frame returns. A retired slot is reused only with a new generation, so a stale pointer never resolves again, and a slot whose generations are exhausted is not reused. Overlapping `memcpy`, invalid shifts/division and nonfinite/out-of-range numeric casts fault explicitly. Bulk operations and host API work consume the execution budget. A VM failure discards partial labels and output; callers cannot mistake them for a successful result.

`RunWithMatchContext` requires real match counts and first offsets; missing offsets are `0xfffffffe`. `RunWithInput` rejects programs that need that metadata. The scanner evaluates the logical trigger and target constraints before dispatch, charges successful `Result.WorkUsed` to the shared object matching budget, and treats runtime failures as incomplete analysis. `VirusNameSet` distinguishes an explicit empty label (invalid for detection) from no label call. The scanner accepts valid strong labels; PUA/heuristic/weak prefixes return unsupported classification instead of being promoted to malware. The braced catalog name is not a label prefixing instruction.

The scanner evaluates static body rules and whole-file hashes before logical bytecode callbacks. Callback triggers still receive their own match counts and offsets. This precedence matches the pinned reference on a harmless overlapping body/hash case and on official-feed EICAR with and without a local body rule; it does not establish every other signature-class ordering.

Limits include 4 MiB serialized programs, 1 MiB lines, 1,024 functions, 4,096 blocks per function, 8,192 values per function, 100,000 instructions, 64 call frames, 16 MiB individual types/frames and total globals, 32 MiB live virtual data, fewer than 65,536 live regions, 4,096 hash sets and 262,144 live hash keys per execution. The number of calls is not limited separately: every frame charges its bytes to the work budget. Initialization bits and Go allocation overhead are additional. Input snapshots are capped at 64 MiB; output at 16 MiB; labels at 256 bytes. Diagnostic capture keeps at most 64 KiB; later diagnostics are dropped and `Result.DiagnosticsTruncated` is set, which never changes the program result. The default work budget is 1,000,000 and maximum 50,000,000. Parser type/reference/depth and initializer-work limits apply separately.

The source reference is pinned to ClamAV commit `72cd48c9faed4fa4afc22bc4ed0b9b19f8d3f8f7`; see [REFERENCE.md](REFERENCE.md) for source hashes, API contracts and intentional compatibility restrictions. In particular, debug API return values follow a deterministic debug-disabled profile, signed i1 arithmetic follows the interpreter's 0/1 behavior, and SEXT i1 preserves its separate sign-extension semantics. Literal storage is truncated before an operation reads the zero-initialized constant slot at its required width. Direct casts of untyped constants remain unsupported. Nonstruct GEPZ offsets are bytes, matching the interpreter; struct GEPZ uses packed field offsets. COPY, LOAD, STORE, SELECT, RET, comparisons, GEP1 results and function results accept integer types and the `i8*` type 65 only: the interpreter runs other pointer types with its one-bit handlers, so they are unsupported. Direct-call arguments must be constants or plain values; the interpreter's argument copy always fails for global operands and addressed locals, so those are unsupported.

On 2026-09-23, native runs of pinned upstream fixtures returned the results asserted by upstream's unit tests: `retmagic_7.cbc` → `0x1234f00d`, `arith_7.cbc` → `0xd55555dd`, `apicalls_7.cbc` → `0xf00d`, and `api_files_7.cbc` with its benign input → `0xf00d`. `inf_7.cbc` exhausted the bounded work budget. This checks published fixture expectations; it is **not a local differential run of the C engine**, nor full feed/corpus coverage. No upstream fixture is installed as an endpoint signature.

Repository tests use synthetic programs. Optional offline fixture checks require the named, SHA-256-verified files described in REFERENCE.md:

```sh
CLAMBC_REFERENCE_DIR=/path/to/pinned-fixtures go test -race ./edr/bytecode
```

Remaining work includes scanner APIs beyond the supported table, complete preclass metadata and parser/child-object target-13 serialization. The engine currently proves only childless binary and plain-ASCII scalar metadata trees; other active target-13 trees remain incomplete. A newer ClamAV release that repairs API 89's resource record will require a versioned semantic profile and a positive oracle. Direct VM/matcher runs without a child scanner remain incomplete on reached `write` or `extract_new`; the engine-provided callback scans those outputs recursively, subject to its 16 MiB per-output buffer cap and the shared scan limits. The official unpackers' output paths still need differential validation. GEPN and debug metadata are unsupported; the [ClamAV 1.5.4 interpreter](https://github.com/Cisco-Talos/clamav/blob/clamav-1.5.4/libclamav/bytecode_vm.c#L1312) also leaves GEPN unimplemented, so it is not a current reference-compatibility gap. Encoded source trailers are inert. A standalone `Result.VirusName` still needs scanner validation; `Result.Value` is never a malware verdict by itself. Full official bytecode feed activation and differential corpus validation remain open.
