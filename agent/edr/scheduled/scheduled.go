// Package scheduled enumerates configured trees into the shared scan queue.
// It never executes content. Discovered symlinks and configured symlink ancestors
// are skipped; monitoring paths do not provide an adversarial openat-style boundary.
package scheduled

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

var ErrBudget = errors.New("scheduled scan enumeration budget exceeded")
var ErrRunning = errors.New("scheduled scan sweep already running")

type Status struct {
	Enabled  bool      `json:"enabled"`
	Running  bool      `json:"running"`
	Started  time.Time `json:"started"`
	Finished time.Time `json:"finished"`
	Entries  int       `json:"entries"`
	Enqueued int       `json:"enqueued"`
	Skipped  int       `json:"skipped"`
	Error    string    `json:"error,omitempty"`
}

type Scheduler struct {
	cfg      config.ScheduledScan
	enqueue  func(context.Context, string) error
	excluded func(string) bool
	mu       sync.Mutex
	status   Status
}

func New(cfg config.ScheduledScan, enqueue func(context.Context, string) error, excluded func(string) bool) (*Scheduler, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if enqueue == nil {
		return nil, errors.New("scheduled scan needs a queue")
	}
	return &Scheduler{cfg: cfg.Effective(), enqueue: enqueue, excluded: excluded, status: Status{Enabled: cfg.Enabled}}, nil
}

func (s *Scheduler) Status() Status { s.mu.Lock(); defer s.mu.Unlock(); return s.status }

// Run waits between sweeps, so slow enumeration cannot overlap or accumulate
// catch-up work. Per-sweep deadlines also bound backpressure from a full queue.
func (s *Scheduler) Run(ctx context.Context) {
	if !s.cfg.Enabled {
		return
	}
	delay := time.Duration(s.cfg.IntervalMinutes) * time.Minute
	if s.cfg.RunOnStart {
		delay = 0
	}
	for {
		t := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
		_ = s.Sweep(ctx)
		delay = time.Duration(s.cfg.IntervalMinutes) * time.Minute
	}
}

func (s *Scheduler) Sweep(parent context.Context) (err error) {
	s.mu.Lock()
	if s.status.Running {
		s.mu.Unlock()
		return ErrRunning
	}
	s.status = Status{Enabled: s.cfg.Enabled, Running: true, Started: time.Now().UTC()}
	s.mu.Unlock()
	st := s.Status()
	defer func() {
		st.Running = false
		st.Finished = time.Now().UTC()
		if err != nil {
			st.Error = err.Error()
		}
		s.mu.Lock()
		s.status = st
		s.mu.Unlock()
	}()
	ctx, cancel := context.WithTimeout(parent, time.Duration(s.cfg.MaxDurationMinutes)*time.Minute)
	defer cancel()
	var firstErr error
	var walk func(string, int) error
	walk = func(path string, depth int) error {
		if e := ctx.Err(); e != nil {
			return e
		}
		if st.Entries >= s.cfg.MaxEntries {
			return ErrBudget
		}
		st.Entries++
		if skipSystemPath(path) || (s.excluded != nil && s.excluded(path)) {
			st.Skipped++
			return nil
		}
		info, e := os.Lstat(path)
		if e != nil {
			if firstErr == nil {
				firstErr = e
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			st.Skipped++
			return nil
		}
		if !info.IsDir() {
			if !info.Mode().IsRegular() {
				st.Skipped++
				return nil
			}
			if st.Enqueued >= s.cfg.MaxFiles {
				return ErrBudget
			}
			if e = s.enqueue(ctx, path); e != nil {
				return e
			}
			st.Enqueued++
			return nil
		}
		if depth >= s.cfg.MaxDepth {
			if firstErr == nil {
				firstErr = fmt.Errorf("%w: depth at %s", ErrBudget, path)
			}
			st.Skipped++
			return nil
		}
		f, e := os.Open(path)
		if e != nil {
			if firstErr == nil {
				firstErr = e
			}
			return nil
		}
		defer f.Close()
		opened, statErr := f.Stat()
		if statErr != nil || !os.SameFile(info, opened) {
			if firstErr == nil {
				firstErr = fmt.Errorf("scheduled directory changed while opening: %s", path)
			}
			return nil
		}
		// Read bounded batches instead of allocating every directory entry.
		for {
			entries, readErr := f.ReadDir(128)
			for _, entry := range entries {
				if e = walk(filepath.Join(path, entry.Name()), depth+1); e != nil {
					return e
				}
			}
			if readErr == io.EOF {
				return nil
			}
			if readErr != nil {
				if firstErr == nil {
					firstErr = readErr
				}
				return nil
			}
		}
	}
	for _, p := range s.cfg.Paths {
		if e := checkAncestors(filepath.Clean(p)); e != nil {
			if firstErr == nil {
				firstErr = e
			}
			continue
		}
		if err = walk(filepath.Clean(p), 0); err != nil {
			return err
		}
	}
	return firstErr
}

func checkAncestors(path string) error {
	for parent := filepath.Dir(path); ; parent = filepath.Dir(parent) {
		info, err := os.Lstat(parent)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("scheduled root has a symbolic-link ancestor: %s", parent)
		}
		if filepath.Dir(parent) == parent {
			return nil
		}
	}
}
