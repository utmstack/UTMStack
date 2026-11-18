// engineprofile is a local development diagnostic. It never starts a service,
// opens a profile listener, or executes/quarantines any input file.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

type memory struct {
	Phase        string `json:"phase"`
	ElapsedMS    int64  `json:"elapsed_ms"`
	HeapAlloc    uint64 `json:"heap_alloc"`
	HeapInuse    uint64 `json:"heap_inuse"`
	HeapSys      uint64 `json:"heap_sys"`
	HeapReleased uint64 `json:"heap_released"`
	TotalAlloc   uint64 `json:"total_alloc"`
	Sys          uint64 `json:"runtime_sys"`
	NumGC        uint32 `json:"num_gc"`
}

type monitor struct {
	mu    sync.Mutex
	start time.Time
	phase string
	peaks map[string]memory
}

func (m *monitor) read(phase string) memory {
	var s runtime.MemStats
	runtime.ReadMemStats(&s)
	return memory{phase, time.Since(m.start).Milliseconds(), s.HeapAlloc, s.HeapInuse, s.HeapSys, s.HeapReleased, s.TotalAlloc, s.Sys, s.NumGC}
}

func (m *monitor) mark(phase string) memory {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.phase = phase
	return m.read(phase)
}

func (m *monitor) sample() {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.read(m.phase)
	if s.HeapAlloc > m.peaks[m.phase].HeapAlloc {
		m.peaks[m.phase] = s
	}
}

func writeProfile(path, name string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return pprof.Lookup(name).WriteTo(f, 0)
}

func run() error {
	db := flag.String("db", "", "pinned signature database directory")
	out := flag.String("out", "", "new output directory (must not exist)")
	reload := flag.Bool("reload", false, "profile reload of the same stored generation (may reuse its verified matcher; this does not measure a changed-generation update)")
	softLimit := flag.Int64("memory-limit-mib", 0, "optional Go runtime soft limit for this diagnostic process only")
	flag.Parse()
	if *db == "" || *out == "" || *softLimit < 0 {
		return fmt.Errorf("require -db and -out; memory limit must be nonnegative")
	}
	if err := os.Mkdir(*out, 0700); err != nil {
		return err
	}
	if *softLimit > 0 {
		debug.SetMemoryLimit(*softLimit << 20)
	}
	log, err := os.OpenFile(filepath.Join(*out, "phases.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer log.Close()
	enc := json.NewEncoder(log)
	m := &monitor{start: time.Now(), peaks: map[string]memory{}}
	done, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		t := time.NewTicker(5 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				m.sample()
			case <-done:
				return
			}
		}
	}()
	defer func() { close(done); <-stopped }()
	if err := enc.Encode(m.mark("start")); err != nil {
		return err
	}
	cfg := config.Default()
	cfg.SignatureDBDir = *db
	e := engine.New(cfg)
	defer e.Stop()
	m.mark("load")
	if err := e.EnsureRunning(); err != nil {
		return err
	}
	if err := enc.Encode(m.mark("loaded_before_gc")); err != nil {
		return err
	}
	runtime.GC()
	if err := enc.Encode(m.mark("loaded_after_gc")); err != nil {
		return err
	}
	if err := writeProfile(filepath.Join(*out, "loaded.heap.pprof"), "heap"); err != nil {
		return err
	}
	for _, path := range flag.Args() {
		m.mark("scan")
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		limit := int64(e.Tuning().MaxFileSizeMB) << 20
		b, err := io.ReadAll(io.LimitReader(f, limit+1))
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if int64(len(b)) > limit {
			return fmt.Errorf("diagnostic file exceeds scan limit")
		}
		clean, sig, err := e.Scan(context.Background(), b)
		var scanError string
		if err != nil {
			scanError = err.Error()
		}
		if err := enc.Encode(struct {
			Path      string `json:"path"`
			Clean     bool   `json:"clean"`
			Signature string `json:"signature,omitempty"`
			Error     string `json:"error,omitempty"`
		}{path, clean, sig, scanError}); err != nil {
			return err
		}
	}
	if *reload {
		m.mark("reload")
		if err := e.Reload(); err != nil {
			return err
		}
		if err := enc.Encode(m.mark("reloaded_before_gc")); err != nil {
			return err
		}
		runtime.GC()
		if err := enc.Encode(m.mark("reloaded_after_gc")); err != nil {
			return err
		}
		if err := writeProfile(filepath.Join(*out, "reloaded.heap.pprof"), "heap"); err != nil {
			return err
		}
	}
	if err := enc.Encode(struct {
		Version    string `json:"version"`
		Signatures int    `json:"signatures"`
	}{e.Version(), e.SignatureCount()}); err != nil {
		return err
	}
	if err := writeProfile(filepath.Join(*out, "allocations.pprof"), "allocs"); err != nil {
		return err
	}
	m.sample()
	m.mu.Lock()
	err = enc.Encode(struct {
		Peaks map[string]memory `json:"sampled_heap_peaks"`
	}{m.peaks})
	m.mu.Unlock()
	if err != nil {
		return err
	}
	e.Stop()
	debug.FreeOSMemory()
	return enc.Encode(m.mark("stopped_and_released"))
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
