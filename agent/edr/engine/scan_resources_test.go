package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// Share immutable rules with a control engine so each case exercises the same
// matching semantics under different configured analysis budgets.
func engineWithScanLimits(t *testing.T, control *Engine, limits config.ScanLimits) *Engine {
	t.Helper()
	cfg := control.cfg
	cfg.ScanConcurrency = 1
	cfg.ScanLimits = limits
	e := New(cfg)
	e.current.Store(control.current.Load())
	return e
}

func TestConfiguredResourcesBoundRealAnalysis(t *testing.T) {
	control := nativeTestEngine(t)
	benign := bytes.Repeat([]byte("A"), 600<<10)
	var multi bytes.Buffer
	zw := zip.NewWriter(&multi)
	for i := 0; i < 2; i++ {
		w, err := zw.Create(fmt.Sprintf("%d.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(benign); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		limits config.ScanLimits
		data   []byte
	}{
		{"per-file", config.ScanLimits{MaxFileSizeMB: 1}, bytes.Repeat([]byte("A"), (1<<20)+1)},
		{"zip-expanded-bytes", config.ScanLimits{MaxScanSizeMB: 1}, multi.Bytes()},
		{"pdf-expanded-bytes", config.ScanLimits{MaxScanSizeMB: 1}, pdfFixture("/Filter /FlateDecode", zlibFixture(t, bytes.Repeat([]byte("A"), (1<<20)+1)))},
		{"zip-object-count", config.ScanLimits{MaxFiles: 2}, multi.Bytes()},
		{"nested-zip-depth", config.ScanLimits{MaxRecursion: 1}, zipFixture(t, zipFixture(t, []byte("benign content")))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if clean, sig, err := control.Scan(context.Background(), tc.data); !clean || sig != "" || err != nil {
				t.Fatalf("control fixture: clean=%v sig=%q err=%v", clean, sig, err)
			}
			e := engineWithScanLimits(t, control, tc.limits)
			if clean, sig, err := e.Scan(context.Background(), tc.data); clean || sig != "" || !errors.Is(err, ErrLimit) {
				t.Fatalf("limit yielded clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
}

func TestConfiguredResourceBudgetPreservesDetectionsWithinLimits(t *testing.T) {
	control := nativeTestEngine(t)
	e := engineWithScanLimits(t, control, config.ScanLimits{MaxFileSizeMB: 1, MaxScanSizeMB: 2, MaxFiles: 20, MaxRecursion: 4, MaxScanTimeMS: 5000})
	for name, data := range map[string][]byte{
		"raw":  []byte(testMarker),
		"zip":  zipFixture(t, []byte(testMarker)),
		"gzip": gzipFixture(t, []byte(testMarker)),
		"pdf":  pdfFixture("/Filter /FlateDecode", zlibFixture(t, []byte(testMarker))),
	} {
		if clean, sig, err := e.Scan(context.Background(), data); clean || sig != "Native.Test" || err != nil {
			t.Errorf("%s: clean=%v sig=%q err=%v", name, clean, sig, err)
		}
	}
}

func TestConfiguredScanDeadlineIncludesWorkerWait(t *testing.T) {
	e := engineWithScanLimits(t, nativeTestEngine(t), config.ScanLimits{MaxScanTimeMS: 20})
	if tn := e.Tuning(); tn.MaxThreads != 1 {
		t.Fatalf("concurrency did not reach engine: %+v", tn)
	}
	e.slots <- struct{}{}
	defer func() { <-e.slots }()
	// The caller deadline is deliberately longer than the configured limit.
	// An ignored config deadline would otherwise block for a full second.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	start := time.Now()
	clean, sig, err := e.Scan(ctx, []byte("benign content"))
	if clean || sig != "" || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("busy worker: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("configured 20 ms worker wait took %v", elapsed)
	}
}

func TestEffectiveBudgetsInvalidateCachedScanIdentity(t *testing.T) {
	control := nativeTestEngine(t)
	base := control.Version()
	seen := map[string]bool{base: true}
	for _, limits := range []config.ScanLimits{
		{MaxFileSizeMB: 1}, {MaxScanSizeMB: 2}, {MaxRecursion: 1}, {MaxFiles: 2}, {MaxScanTimeMS: 1000},
	} {
		e := engineWithScanLimits(t, control, limits)
		v := e.Version()
		if seen[v] {
			t.Errorf("changed budgets %+v reused scan identity %q", limits, v)
		}
		seen[v] = true
	}
}
