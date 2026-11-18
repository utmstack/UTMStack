package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

// "scan --dry-run --json" analyzes files and only reports: nothing is
// quarantined and neither the verdict cache nor the event spool is written.
// It is for measuring the engine and checking a folder without acting on it.

// dryRunLine is the result for one file.
type dryRunLine struct {
	Path             string  `json:"path"`
	Size             int64   `json:"size"`
	Clean            bool    `json:"clean"`
	Signature        string  `json:"signature,omitempty"`
	IncompleteReason string  `json:"incomplete_reason,omitempty"`
	Error            string  `json:"error,omitempty"`
	Ms               float64 `json:"ms"`
}

// dryRunSummary is the last line, as {"summary": {...}}.
type dryRunSummary struct {
	Implementation     string         `json:"implementation"`
	Generation         string         `json:"generation"`
	Signatures         int            `json:"signatures"`
	LoadMs             int64          `json:"load_ms"`
	Files              int            `json:"files"`
	Bytes              int64          `json:"bytes"`
	Detected           int            `json:"detected"`
	Clean              int            `json:"clean"`
	Incomplete         int            `json:"incomplete"`
	Errors             int            `json:"errors"`
	IncompleteByReason map[string]int `json:"incomplete_by_reason,omitempty"`
	Ms                 int64          `json:"ms"`
}

const dryRunUsage = "usage: utmstack_edr scan --dry-run --json [--db <dir>] [--workers N] <path>..."

// runDryRun runs the command with the arguments after "scan". The exit code
// is 0 only when every file was analyzed and clean, 1 when anything was
// detected (even if other files were incomplete), and 2 for incomplete
// analysis, file errors, or a command that could not run.
func runDryRun(args []string, out io.Writer) int {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	dryRun := flags.Bool("dry-run", false, "")
	asJSON := flags.Bool("json", false, "")
	db := flags.String("db", "", "")
	workers := flags.Int("workers", 1, "")
	if err := flags.Parse(args); err != nil || !*dryRun || !*asJSON || *workers < 1 || flags.NArg() == 0 {
		fmt.Fprintln(os.Stderr, dryRunUsage)
		return 2
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "UTMStack EDR: invalid configuration:", err)
		return 2
	}
	if *db != "" {
		cfg.SignatureDBDir = *db
	}
	started := time.Now()
	eng := engine.New(cfg)
	if err := eng.EnsureRunning(); err != nil && (!errors.Is(err, engine.ErrLocalSignatures) || eng.Version() == "") {
		fmt.Fprintln(os.Stderr, "UTMStack EDR: signature database unavailable:", err)
		return 2
	}
	defer eng.Stop()
	summary := dryRunSummary{Implementation: engine.ImplementationVersion, Generation: eng.Version(),
		Signatures: eng.SignatureCount(), LoadMs: time.Since(started).Milliseconds()}
	started = time.Now()
	dryRunScan(eng, flags.Args(), *workers, int64(eng.Tuning().MaxFileSizeMB)<<20, out, &summary)
	summary.Ms = time.Since(started).Milliseconds()
	_ = json.NewEncoder(out).Encode(struct {
		Summary dryRunSummary `json:"summary"`
	}{summary})
	if summary.Detected > 0 {
		return 1
	}
	if summary.Incomplete > 0 || summary.Errors > 0 {
		return 2
	}
	return 0
}

// dryRunScan analyzes every regular file under paths with workers
// goroutines, writing one line per file and counting into summary. Symbolic
// links and other special files are not followed.
func dryRunScan(eng *engine.Engine, paths []string, workers int, maxFile int64, out io.Writer, summary *dryRunSummary) {
	var mu sync.Mutex
	enc := json.NewEncoder(out)
	emit := func(l dryRunLine) {
		mu.Lock()
		defer mu.Unlock()
		summary.Files++
		summary.Bytes += l.Size
		switch {
		case l.Signature != "":
			summary.Detected++
		case l.IncompleteReason != "":
			summary.Incomplete++
			if summary.IncompleteByReason == nil {
				summary.IncompleteByReason = map[string]int{}
			}
			summary.IncompleteByReason[l.IncompleteReason]++
		case l.Error != "":
			summary.Errors++
		default:
			summary.Clean++
		}
		_ = enc.Encode(l)
	}
	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				emit(dryRunFile(eng, path, maxFile))
			}
		}()
	}
	for _, root := range paths {
		info, err := os.Stat(root)
		if err != nil {
			emit(dryRunLine{Path: root, Error: err.Error()})
			continue
		}
		if !info.IsDir() {
			jobs <- root
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d iofs.DirEntry, err error) error {
			switch {
			case err != nil:
				emit(dryRunLine{Path: path, Error: err.Error()})
			case d.Type().IsRegular():
				jobs <- path
			}
			return nil
		})
	}
	close(jobs)
	wg.Wait()
}

// dryRunFile analyzes one file.
func dryRunFile(eng *engine.Engine, path string, maxFile int64) (l dryRunLine) {
	l.Path = path
	info, err := os.Stat(path)
	if err != nil {
		l.Error = err.Error()
		return l
	}
	l.Size = info.Size()
	if !info.Mode().IsRegular() {
		l.Error = "not a regular file"
		return l
	}
	started := time.Now()
	defer func() { l.Ms = float64(time.Since(started).Microseconds()) / 1000 }()
	if l.Size > maxFile {
		l.IncompleteReason, l.Error = engine.IncompleteReason(engine.ErrLimit), fmt.Sprintf("file exceeds %d bytes", maxFile)
		return l
	}
	data, err := readAtMost(path, maxFile)
	if err != nil {
		l.Error = err.Error()
		if errors.Is(err, engine.ErrLimit) {
			l.IncompleteReason = engine.IncompleteReason(err)
		}
		return l
	}
	l.Size = int64(len(data))
	clean, sig, err := eng.Scan(context.Background(), data)
	switch {
	case err != nil:
		l.IncompleteReason, l.Error = engine.IncompleteReason(err), err.Error()
	case clean:
		l.Clean = true
	default:
		l.Signature = event.SanitizeName(sig)
	}
	return l
}

// readAtMost reads the file, failing with ErrLimit when it holds more than
// limit bytes (it may have grown since it was listed).
func readAtMost(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%w: file exceeds %d bytes", engine.ErrLimit, limit)
	}
	return data, nil
}
