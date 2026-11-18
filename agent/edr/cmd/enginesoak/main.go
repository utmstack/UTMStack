// enginesoak measures repeated resident scans of a pinned safe corpus. It
// starts no service, listener or subprocess and never executes corpus contents.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

const maxManifest = 16 << 20

type fixture struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
	Label  string `json:"label"`
	Family string `json:"family"`
}
type manifest struct {
	Schema int       `json:"schema"`
	Files  []fixture `json:"files"`
}
type options struct {
	DB, Corpus, Out, Tier              string
	Workers, Rounds, Reloads, SampleMS int
	Limits                             config.ScanLimits
}
type scanEngine interface {
	ScanSnapshot(context.Context, []byte) (bool, string, string, error)
	Reload() error
	Version() string
}
type window struct {
	Start int64 `json:"start_ns"`
	End   int64 `json:"end_ns"`
}
type verdict struct {
	State           string `json:"state"`
	Signature       string `json:"signature,omitempty"`
	Reason          string `json:"reason,omitempty"`
	EngineErrorFree bool   `json:"engine_error_free"`
}
type record struct {
	Round      int    `json:"round"`
	Fixture    string `json:"fixture"`
	SHA256     string `json:"sha256"`
	Generation string `json:"generation"`
	verdict
	ReadSeconds float64 `json:"read_seconds"`
	ScanSeconds float64 `json:"scan_seconds"`
	Window      window  `json:"scan_window"`
}
type reloadRecord struct {
	Index     int    `json:"index"`
	Window    window `json:"window"`
	Before    string `json:"before"`
	After     string `json:"after"`
	Error     string `json:"error,omitempty"`
	OverlapNS int64  `json:"scan_overlap_ns"`
}
type memorySample struct {
	Phase        string `json:"phase"`
	ElapsedNS    int64  `json:"elapsed_ns"`
	HeapAlloc    uint64 `json:"heap_alloc_bytes"`
	HeapInuse    uint64 `json:"heap_inuse_bytes"`
	HeapSys      uint64 `json:"heap_sys_bytes"`
	HeapReleased uint64 `json:"heap_released_bytes"`
	TotalAlloc   uint64 `json:"total_alloc_bytes"`
	RuntimeSys   uint64 `json:"runtime_sys_bytes"`
	NumGC        uint32 `json:"num_gc"`
	PauseTotalNS uint64 `json:"pause_total_ns"`
	Goroutines   int    `json:"goroutines"`
}
type monitor struct {
	mu    sync.Mutex
	start time.Time
	phase string
	enc   *json.Encoder
	err   error
	peaks map[string]memorySample
}

func (m *monitor) sample(phase string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if phase != "" {
		m.phase = phase
	}
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	v := memorySample{m.phase, time.Since(m.start).Nanoseconds(), s.HeapAlloc, s.HeapInuse, s.HeapSys, s.HeapReleased, s.TotalAlloc, s.Sys, s.NumGC, s.PauseTotalNs, runtime.NumGoroutine()}
	if v.HeapAlloc >= m.peaks[m.phase].HeapAlloc {
		m.peaks[m.phase] = v
	}
	if m.err == nil {
		m.err = m.enc.Encode(v)
	}
}

type summary struct {
	Schema            int                     `json:"schema"`
	Scope             string                  `json:"scope"`
	Valid             bool                    `json:"valid"`
	Implementation    string                  `json:"implementation"`
	Generation        string                  `json:"generation"`
	Signatures        int                     `json:"signatures"`
	ManifestSHA256    string                  `json:"manifest_sha256"`
	Effective         engine.Tuning           `json:"effective_tuning"`
	RequestedWorkers  int                     `json:"requested_workers"`
	Rounds            int                     `json:"rounds"`
	FixtureCount      int                     `json:"fixture_count"`
	ScanCount         int                     `json:"scan_count"`
	InputBytes        int64                   `json:"input_bytes"`
	LoadSeconds       float64                 `json:"load_seconds"`
	ResidentSeconds   float64                 `json:"resident_scan_seconds"`
	ReloadWaitSeconds float64                 `json:"post_scan_reload_wait_seconds"`
	TotalSeconds      float64                 `json:"total_seconds"`
	ResidentMiBs      float64                 `json:"resident_input_mib_per_second"`
	LatencyP50        float64                 `json:"scan_latency_p50_seconds"`
	LatencyP95        float64                 `json:"scan_latency_p95_seconds"`
	Counts            map[string]int          `json:"counts"`
	Stable            bool                    `json:"verdicts_stable"`
	MaxActive         int64                   `json:"max_active_scan_calls"`
	Reloads           []reloadRecord          `json:"reloads"`
	Peaks             map[string]memorySample `json:"sampled_heap_peaks"`
	GoVersion         string                  `json:"go_version"`
	GOMAXPROCS        int                     `json:"gomaxprocs"`
	GOMEMLIMIT        string                  `json:"gomemlimit_environment"`
	ForcedGC          bool                    `json:"forced_gc"`
}

func validateOptions(o options) error {
	if o.DB == "" || o.Corpus == "" || o.Out == "" {
		return errors.New("-db, -corpus and a new -out directory are required")
	}
	if o.Workers < 1 || o.Workers > 16 || o.Rounds < 1 || o.Rounds > 100 || o.Reloads < 0 || o.Reloads > 100 {
		return errors.New("workers must be 1..16, rounds 1..100, reloads 0..100")
	}
	if o.SampleMS < 25 || o.SampleMS > 5000 {
		return errors.New("sample-ms must be 25..5000")
	}
	cfg := config.Default()
	cfg.TierOverride = o.Tier
	cfg.ScanConcurrency = o.Workers
	cfg.ScanLimits = o.Limits
	return cfg.ValidateScanResources()
}
func safeFile(root, name string) (string, error) {
	if name == "" || path.IsAbs(name) || path.Clean(name) != name || strings.Contains(name, "\\") || strings.Contains(name, ":") {
		return "", errors.New("unsafe fixture path")
	}
	for _, part := range strings.Split(name, "/") {
		if part == "." || part == ".." || part == "" {
			return "", errors.New("unsafe fixture path")
		}
	}
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return "", err
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return "", errors.New("fixture root is not a regular directory")
	}
	current := root
	for _, part := range strings.Split(name, "/") {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("symlink fixture path")
		}
	}
	info, err := os.Stat(current)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("fixture is not regular")
	}
	return current, nil
}
func readFixture(root string, f fixture) ([]byte, error) {
	p, err := safeFile(root, f.Path)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() != f.Bytes {
		return nil, errors.New("fixture size changed")
	}
	data := make([]byte, int(f.Bytes)+1)
	n, err := io.ReadFull(file, data)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	if int64(n) != f.Bytes {
		return nil, errors.New("fixture size changed while reading")
	}
	data = data[:n]
	h := sha256.Sum256(data)
	if hex.EncodeToString(h[:]) != f.SHA256 {
		return nil, errors.New("fixture digest changed")
	}
	return data, nil
}
func loadManifest(corpus string) (manifest, string, error) {
	root, err := filepath.Abs(corpus)
	if err != nil {
		return manifest{}, "", err
	}
	p, err := safeFile(root, "manifest.json")
	if err != nil {
		return manifest{}, "", err
	}
	f, err := os.Open(p)
	if err != nil {
		return manifest{}, "", err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, maxManifest+1))
	if err != nil {
		return manifest{}, "", err
	}
	if len(raw) > maxManifest {
		return manifest{}, "", errors.New("manifest exceeds 16 MiB")
	}
	var m manifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return m, "", err
	}
	if m.Schema != 1 || len(m.Files) < 1 || len(m.Files) > 10000 {
		return m, "", errors.New("manifest requires schema 1 and 1..10000 files")
	}
	seen := map[string]bool{}
	for _, entry := range m.Files {
		if seen[entry.Path] {
			return m, "", errors.New("duplicate manifest fixture")
		}
		seen[entry.Path] = true
		h, err := hex.DecodeString(entry.SHA256)
		if err != nil || len(h) != 32 || strings.ToLower(entry.SHA256) != entry.SHA256 {
			return m, "", errors.New("invalid fixture SHA256")
		}
		if entry.Bytes < 0 || entry.Bytes > int64(config.MaxFileSizeMB)<<20 {
			return m, "", errors.New("fixture size outside 0..200 MiB")
		}
		if _, err = readFixture(filepath.Join(root, "files"), entry); err != nil {
			return m, "", fmt.Errorf("fixture %s: %w", entry.Path, err)
		}
	}
	h := sha256.Sum256(raw)
	return m, hex.EncodeToString(h[:]), nil
}
func classify(clean bool, sig string, err error) verdict {
	if sig != "" {
		v := verdict{State: "detected", Signature: sig, EngineErrorFree: err == nil}
		if err != nil {
			v.Reason = engine.IncompleteReason(err)
		}
		return v
	}
	if clean && err == nil {
		return verdict{State: "clean", EngineErrorFree: true}
	}
	if err != nil {
		return verdict{State: "incomplete", Reason: engine.IncompleteReason(err)}
	}
	return verdict{State: "error", Reason: "undecided without error", EngineErrorFree: true}
}
func mergedWindows(windows []window) []window {
	if len(windows) == 0 {
		return nil
	}
	sort.Slice(windows, func(i, j int) bool { return windows[i].Start < windows[j].Start })
	out := []window{windows[0]}
	for _, w := range windows[1:] {
		last := &out[len(out)-1]
		if w.Start <= last.End {
			if w.End > last.End {
				last.End = w.End
			}
		} else {
			out = append(out, w)
		}
	}
	return out
}
func overlapNS(reload window, scans []window) int64 {
	var total int64
	for _, w := range scans {
		left, right := max(reload.Start, w.Start), min(reload.End, w.End)
		if right > left {
			total += right - left
		}
	}
	return total
}
func quantile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sort.Float64s(values)
	return values[int(float64(len(values)-1)*p)]
}

func scanRounds(ctx context.Context, e scanEngine, o options, m manifest, start time.Time, enc *json.Encoder, scansDone func()) (int, int64, map[string]int, bool, int64, []window, []reloadRecord, []float64, float64, error) {
	type job struct{ round, index int }
	jobs := make(chan job)
	results := make(chan record, o.Workers)
	firstScan := make(chan struct{})
	var first sync.Once
	var active, maxActive atomic.Int64
	var workers sync.WaitGroup
	for worker := 0; worker < o.Workers; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := range jobs {
				entry := m.Files[j.index]
				readStart := time.Now()
				data, err := readFixture(filepath.Join(o.Corpus, "files"), entry)
				r := record{Round: j.round, Fixture: entry.Path, SHA256: entry.SHA256, ReadSeconds: time.Since(readStart).Seconds()}
				if err != nil {
					r.verdict = verdict{State: "error", Reason: err.Error()}
				} else {
					n := active.Add(1)
					for old := maxActive.Load(); n > old && !maxActive.CompareAndSwap(old, n); old = maxActive.Load() {
					}
					r.Window.Start = time.Since(start).Nanoseconds()
					first.Do(func() { close(firstScan) })
					clean, sig, generation, scanErr := e.ScanSnapshot(ctx, data)
					r.Window.End = time.Since(start).Nanoseconds()
					active.Add(-1)
					r.ScanSeconds = float64(r.Window.End-r.Window.Start) / 1e9
					r.Generation = generation
					r.verdict = classify(clean, sig, scanErr)
				}
				results <- r
			}
		}()
	}
	go func() {
		defer close(jobs)
		for round := 0; round < o.Rounds; round++ {
			for index := range m.Files {
				select {
				case jobs <- job{round, index}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	go func() { workers.Wait(); close(results); first.Do(func() { close(firstScan) }) }()
	reloaded := make(chan []reloadRecord, 1)
	go func() {
		<-firstScan
		records := make([]reloadRecord, 0, o.Reloads)
		for i := 0; i < o.Reloads; i++ {
			if ctx.Err() != nil {
				break
			}
			r := reloadRecord{Index: i, Before: e.Version(), Window: window{Start: time.Since(start).Nanoseconds()}}
			if err := e.Reload(); err != nil {
				r.Error = err.Error()
			}
			r.Window.End = time.Since(start).Nanoseconds()
			r.After = e.Version()
			records = append(records, r)
		}
		reloaded <- records
	}()
	count := 0
	var bytes int64
	counts := map[string]int{}
	baseline := map[string]verdict{}
	stable := true
	sizes := map[string]int64{}
	for _, f := range m.Files {
		sizes[f.Path] = f.Bytes
	}
	windows := make([]window, 0, len(m.Files)*o.Rounds)
	latencies := make([]float64, 0, len(m.Files)*o.Rounds)
	var writeErr error
	for r := range results {
		count++
		counts[r.State]++
		bytes += sizes[r.Fixture]
		if old, ok := baseline[r.Fixture]; ok && old != r.verdict {
			stable = false
		} else if !ok {
			baseline[r.Fixture] = r.verdict
		}
		if r.Window.End > 0 {
			windows = append(windows, r.Window)
			latencies = append(latencies, r.ScanSeconds)
		}
		if writeErr == nil {
			writeErr = enc.Encode(r)
		}
	}
	scanEnd := time.Now()
	if scansDone != nil {
		scansDone()
	}
	reloads := <-reloaded
	wait := time.Since(scanEnd).Seconds()
	merged := mergedWindows(windows)
	for i := range reloads {
		reloads[i].OverlapNS = overlapNS(reloads[i].Window, merged)
	}
	return count, bytes, counts, stable, maxActive.Load(), windows, reloads, latencies, wait, writeErr
}

func execute(ctx context.Context, o options) (summary, error) {
	var report summary
	if err := validateOptions(o); err != nil {
		return report, err
	}
	var err error
	o.Corpus, err = filepath.Abs(o.Corpus)
	if err != nil {
		return report, err
	}
	m, manifestHash, err := loadManifest(o.Corpus)
	if err != nil {
		return report, err
	}
	if err = os.Mkdir(o.Out, 0700); err != nil {
		return report, err
	}
	memfile, err := os.OpenFile(filepath.Join(o.Out, "runtime.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return report, err
	}
	defer memfile.Close()
	recordfile, err := os.OpenFile(filepath.Join(o.Out, "records.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return report, err
	}
	defer recordfile.Close()
	start := time.Now()
	monitor := &monitor{start: start, enc: json.NewEncoder(memfile), peaks: map[string]memorySample{}}
	monitor.sample("start")
	done, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(time.Duration(o.SampleMS) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				monitor.sample("")
			case <-done:
				return
			}
		}
	}()
	var stopOnce sync.Once
	stopMonitor := func() { stopOnce.Do(func() { close(done); <-stopped }) }
	defer stopMonitor()
	cfg := config.Default()
	cfg.SignatureDBDir = o.DB
	cfg.ScanConcurrency = o.Workers
	cfg.TierOverride = o.Tier
	cfg.ScanLimits = o.Limits
	e := engine.New(cfg)
	defer e.Stop()
	tuning := e.Tuning()
	if tuning.MaxThreads != o.Workers {
		return report, fmt.Errorf("effective engine workers %d differ from requested %d (%s)", tuning.MaxThreads, o.Workers, tuning.Reason)
	}
	monitor.sample("load")
	loadStart := time.Now()
	if err = e.EnsureRunning(); err != nil {
		return report, err
	}
	loadSeconds := time.Since(loadStart).Seconds()
	monitor.sample("resident_scan")
	scanStart := time.Now()
	count, bytes, counts, stable, maxActive, _, reloads, latencies, wait, err := scanRounds(ctx, e, o, m, start, json.NewEncoder(recordfile), func() { monitor.sample("post_scan_reload_wait") })
	if err != nil {
		return report, err
	}
	residentSeconds := time.Since(scanStart).Seconds() - wait
	monitor.sample("complete")
	stopMonitor()
	valid := count == len(m.Files)*o.Rounds && stable && ctx.Err() == nil && counts["error"] == 0 && len(reloads) == o.Reloads
	for _, r := range reloads {
		if r.Error != "" {
			valid = false
		}
	}
	monitor.mu.Lock()
	peaks := make(map[string]memorySample, len(monitor.peaks))
	for name, v := range monitor.peaks {
		peaks[name] = v
	}
	monitorErr := monitor.err
	monitor.mu.Unlock()
	if monitorErr != nil {
		return report, monitorErr
	}
	report = summary{Schema: 1, Scope: "resident scans include bounded input read/digest verification, engine calls and requested overlapping serial reloads; no forced GC; process CPU/RSS measured separately", Valid: valid, Implementation: engine.ImplementationVersion, Generation: e.Version(), Signatures: e.SignatureCount(), ManifestSHA256: manifestHash, Effective: tuning, RequestedWorkers: o.Workers, Rounds: o.Rounds, FixtureCount: len(m.Files), ScanCount: count, InputBytes: bytes, LoadSeconds: loadSeconds, ResidentSeconds: residentSeconds, ReloadWaitSeconds: wait, TotalSeconds: time.Since(start).Seconds(), Counts: counts, Stable: stable, MaxActive: maxActive, Reloads: reloads, Peaks: peaks, GoVersion: runtime.Version(), GOMAXPROCS: runtime.GOMAXPROCS(0), GOMEMLIMIT: os.Getenv("GOMEMLIMIT")}
	if residentSeconds > 0 {
		report.ResidentMiBs = float64(bytes) / (1 << 20) / residentSeconds
	}
	report.LatencyP50 = quantile(latencies, .5)
	report.LatencyP95 = quantile(latencies, .95)
	f, err := os.OpenFile(filepath.Join(o.Out, "summary.json"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return report, err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(report)
	return report, err
}
func main() {
	var o options
	flag.StringVar(&o.DB, "db", "", "pinned database directory")
	flag.StringVar(&o.Corpus, "corpus", "", "corpus root with manifest.json and files/")
	flag.StringVar(&o.Out, "out", "", "new output directory")
	flag.StringVar(&o.Tier, "tier", "", "optional host tier override, still subject to viability/security limits")
	flag.IntVar(&o.Workers, "workers", 1, "resident scan workers, 1..16")
	flag.IntVar(&o.Rounds, "rounds", 3, "complete corpus rounds, 1..100")
	flag.IntVar(&o.Reloads, "reloads", 0, "serial reloads overlapping scans, 0..100")
	flag.IntVar(&o.SampleMS, "sample-ms", 100, "runtime memory sample interval, 25..5000 ms")
	flag.IntVar(&o.Limits.MaxFileSizeMB, "file-mib", 0, "file scan budget; zero selects host policy")
	flag.IntVar(&o.Limits.MaxScanSizeMB, "expanded-mib", 0, "aggregate decoded budget")
	flag.IntVar(&o.Limits.MaxRecursion, "depth", 0, "recursion budget")
	flag.IntVar(&o.Limits.MaxFiles, "objects", 0, "object count budget")
	flag.IntVar(&o.Limits.MaxScanTimeMS, "scan-time-ms", 0, "per-input deadline including engine slot wait")
	flag.Parse()
	if len(flag.Args()) != 0 {
		fmt.Fprintln(os.Stderr, "unexpected positional arguments")
		os.Exit(2)
	}
	report, err := execute(context.Background(), o)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err = json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if !report.Valid {
		os.Exit(2)
	}
}
