// Package fanotify owns bounded, explicitly opted-in Linux mount notifications
// and permission requests. No monitor is activated merely by constructing it.
package fanotify

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	ModeNotify                = "notify"
	ModeExec                  = "permission_exec"
	ModeOpen                  = "permission_open"
	Modify             uint64 = 0x2
	CloseWrite         uint64 = 0x8
	OpenExecPermission uint64 = 0x40000
	OpenPermission     uint64 = 0x10000
	QueueOverflow      uint64 = 0x4000
)

var (
	errNotRegular    = errors.New("fanotify object is not a regular file")
	errAmbiguousName = errors.New("fanotify object has no stable scoped path: ambiguous deletion suffix")
	ErrDisabled      = errors.New("fanotify is not explicitly enabled")
	ErrUnsupported   = errors.New("fanotify platform, kernel or capability unavailable")
	ErrUsed          = errors.New("fanotify monitor is single-use")
	ErrOverflow      = errors.New("fanotify coverage lost to bounded admission or queue overflow")
	ErrIncomplete    = errors.New("fanotify scanner did not produce a complete verdict")
	ErrFileLimit     = errors.New("fanotify input exceeds configured byte limit")
	ErrAttribution   = errors.New("fanotify process identity unavailable or stale")
)

type Config struct {
	Enabled             bool
	Mounts, Roots       []string
	Mode                string
	FailClosed          bool
	Deadline            time.Duration
	Workers, MaxPending int
	MaxFileBytes        int64
}
type Event struct {
	Path           string
	Size           int64
	Device, Inode  uint64 // identity from the pinned event FD, not its pathname
	version        string
	PID, PPID      int
	OSIdentity     int64
	Image, Cmdline string
	Mask           uint64
	Permission     bool
}
type Decision struct {
	Clean, Detected       bool
	Signature, Generation string
}
type Scanner func(context.Context, *os.File, Event) (Decision, error)

type Runner interface {
	Run(context.Context) error
	Health() error
	Stats() Stats
}
type Outcome struct {
	Event    Event
	Decision Decision
	Err      error
	Allowed  bool
	Reason   string
}
type Stats struct {
	Received, Resolved, Denied, Timeouts, Overflows, AttributionLost, ObserverLost uint64
	Pending                                                                        int
}

type rawEvent struct {
	fd, pidfd, pid int
	mask           uint64
	pidfds         []int
}
type backend interface {
	Read(context.Context) ([]rawEvent, error)
	Inspect(rawEvent) (Event, error)
	Attribute(rawEvent, *Event) error
	Unchanged(rawEvent, Event) error
	Duplicate(rawEvent) (*os.File, error)
	Respond(rawEvent, bool) error
	Release(rawEvent)
	Close() error
}
type Monitor struct {
	cfg           Config
	scan          Scanner
	observe       func(Outcome)
	mu            sync.Mutex
	used, running bool
	failure       error
	stats         Stats
	open          func(Config) (backend, error)
}

func New(cfg Config, scan Scanner, observe func(Outcome)) (*Monitor, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	if cfg.Mode != ModeNotify && scan == nil {
		return nil, errors.New("fanotify permission mode requires a pinned-descriptor scanner")
	}
	cfg.Mounts = append([]string(nil), cfg.Mounts...)
	cfg.Roots = append([]string(nil), cfg.Roots...)
	return &Monitor{cfg: cfg, scan: scan, observe: observe, open: openBackend}, nil
}
func Validate(cfg Config) error {
	if !cfg.Enabled {
		return ErrDisabled
	}
	if cfg.Mode != ModeNotify && cfg.Mode != ModeExec && cfg.Mode != ModeOpen {
		return errors.New("fanotify mode must be notify, permission_exec or permission_open")
	}
	if cfg.Deadline < 50*time.Millisecond || cfg.Deadline > 5*time.Second || cfg.Workers < 1 || cfg.Workers > 8 || cfg.MaxPending < cfg.Workers || cfg.MaxPending > 64 || cfg.MaxFileBytes < 1 || cfg.MaxFileBytes > 512<<20 {
		return errors.New("fanotify resource limits outside bounds")
	}
	if len(cfg.Mounts) < 1 || len(cfg.Mounts) > 16 || len(cfg.Roots) < 1 || len(cfg.Roots) > 64 {
		return errors.New("fanotify requires 1..16 explicit mounts and 1..64 roots")
	}
	for _, p := range append(append([]string{}, cfg.Mounts...), cfg.Roots...) {
		if !path.IsAbs(p) || path.Clean(p) != p || len(p) > 4096 || strings.IndexByte(p, 0) >= 0 {
			return errors.New("fanotify paths must be clean absolute paths")
		}
		for _, pseudo := range []string{"/proc", "/sys", "/dev", "/run"} {
			if within(p, pseudo) {
				return errors.New("fanotify pseudo-filesystem scopes are unavailable")
			}
		}
	}
	for _, root := range cfg.Roots {
		found := false
		for _, mount := range cfg.Mounts {
			if within(root, mount) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("fanotify root %q has no declared mount", root)
		}
	}
	return nil
}
func within(path, root string) bool {
	return path == root || root == "/" && strings.HasPrefix(path, "/") || strings.HasPrefix(path, root+"/")
}
func (m *Monitor) Health() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failure != nil {
		return m.failure
	}
	if !m.running {
		return errors.New("fanotify is not running")
	}
	return nil
}
func (m *Monitor) Stats() Stats { m.mu.Lock(); defer m.mu.Unlock(); return m.stats }
func (m *Monitor) degrade(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failure == nil {
		m.failure = err
	}
}
func (m *Monitor) count(fn func(*Stats)) { m.mu.Lock(); defer m.mu.Unlock(); fn(&m.stats) }
func (m *Monitor) inScope(path string) bool {
	for _, root := range m.cfg.Roots {
		if within(path, root) {
			return true
		}
	}
	return false
}

// Linux appends its deletion marker to the final path component. A literal
// name with that suffix is indistinguishable, but both have the same parent
// directory. Native setup verifies directory roots; a bounded canonical path
// outside all of them can be excluded without interpreting its uncertain name.
func (m *Monitor) excludesAmbiguousName(name string, err error) bool {
	return errors.Is(err, errAmbiguousName) &&
		strings.HasPrefix(name, "/") && len(name) <= 4096 &&
		path.Clean(name) == name && strings.HasSuffix(name, " (deleted)") &&
		!m.inScope(name)
}

type job struct {
	id    uint64
	ctx   context.Context
	file  *os.File
	event Event
}
type result struct {
	event    Event
	id       uint64
	decision Decision
	err      error
}
type pending struct {
	raw      rawEvent
	event    Event
	deadline time.Time
	cancel   context.CancelFunc
}

// Run is single-use. Even a scanner/observer that ignores cancellation can retain
// only this monitor's fixed workers/queue; restarting this object cannot create
// another unbounded pool. Kernel permission deadlines are independent of scans.
func (m *Monitor) Run(ctx context.Context) (runErr error) {
	m.mu.Lock()
	if m.used {
		m.mu.Unlock()
		return ErrUsed
	}
	m.used = true
	m.mu.Unlock()
	source, err := m.open(m.cfg)
	if err != nil {
		m.degrade(err)
		return err
	}
	active := map[uint64]pending{}
	jobs := make(chan job, m.cfg.MaxPending)
	results := make(chan result, m.cfg.MaxPending)
	observations := make(chan Outcome, m.cfg.MaxPending)
	workerCtx, stop := context.WithCancel(ctx)
	defer func() {
		stop()
		for id, p := range active {
			p.cancel()
			if p.event.Permission {
				_ = source.Respond(p.raw, !m.cfg.FailClosed)
			}
			source.Release(p.raw)
			delete(active, id)
		}
		// Closing the kernel group releases unread requests as allowed. FailClosed
		// applies only while a healthy group can deliver a permission response.
		_ = source.Close()
		close(jobs)
		for j := range jobs {
			j.file.Close()
		}
		close(observations)
		m.mu.Lock()
		m.running = false
		m.stats.Pending = 0
		if runErr != nil && m.failure == nil {
			m.failure = runErr
		}
		m.mu.Unlock()
	}()
	deliver := func(o Outcome) {
		if m.observe == nil {
			return
		}
		select {
		case observations <- o:
		default:
			m.count(func(s *Stats) { s.ObserverLost++ })
			m.degrade(ErrOverflow)
		}
	}
	if m.observe != nil {
		go func() {
			for o := range observations {
				if workerCtx.Err() != nil {
					return
				}
				func() {
					defer func() {
						if p := recover(); p != nil {
							m.degrade(fmt.Errorf("fanotify observer panic: %v", p))
						}
					}()
					m.observe(o)
				}()
			}
		}()
	}
	for i := 0; i < m.cfg.Workers; i++ {
		go func() {
			for j := range jobs {
				r := result{id: j.id, event: j.event}
				func() {
					defer j.file.Close()
					defer func() {
						if p := recover(); p != nil {
							r.err = fmt.Errorf("fanotify scanner panic: %v", p)
						}
					}()
					if err := j.ctx.Err(); err != nil {
						r.err = err
						return
					}
					r.decision, r.err = m.scan(j.ctx, j.file, j.event)
				}()
				select {
				case results <- r:
				case <-workerCtx.Done():
					return
				}
			}
		}()
	}
	m.mu.Lock()
	m.running = true
	m.mu.Unlock()
	resolve := func(id uint64, p pending, d Decision, e error, reason string) error {
		delete(active, id)
		p.cancel()
		if !d.Detected && e == nil {
			if err := source.Unchanged(p.raw, p.event); err != nil {
				e = err
				d.Clean = false
				m.degrade(err)
			}
		}
		if !d.Clean && !d.Detected && e == nil {
			e = ErrIncomplete
			m.degrade(e)
		}
		allowed := !d.Detected
		if !d.Detected && (e != nil || !d.Clean) {
			allowed = !m.cfg.FailClosed
		}
		if d.Clean && d.Detected {
			e = errors.New("inconsistent fanotify scanner decision")
			allowed = false
		}
		if !p.event.Permission {
			allowed = true
		}
		if p.event.Permission {
			if err := source.Respond(p.raw, allowed); err != nil {
				source.Release(p.raw)
				m.degrade(err)
				return err
			}
			m.count(func(s *Stats) {
				s.Resolved++
				if !allowed {
					s.Denied++
				}
			})
		}
		source.Release(p.raw)
		deliver(Outcome{Event: p.event, Decision: d, Err: e, Allowed: allowed, Reason: reason})
		return nil
	}
	var serial uint64
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		for len(results) > 0 {
			r := <-results
			if p, ok := active[r.id]; ok {
				if time.Now().After(p.deadline) {
					if r.decision.Detected {
						deliver(Outcome{Event: r.event, Decision: r.decision, Err: context.DeadlineExceeded, Allowed: !m.cfg.FailClosed || !r.event.Permission, Reason: "late_scan"})
					}
					r.err = context.DeadlineExceeded
					r.decision = Decision{}
					m.count(func(s *Stats) { s.Timeouts++ })
					m.degrade(r.err)
				}
				if err := resolve(r.id, p, r.decision, r.err, "scan"); err != nil {
					return err
				}
			} else if r.decision.Detected {
				deliver(Outcome{Event: r.event, Decision: r.decision, Err: context.DeadlineExceeded, Allowed: !m.cfg.FailClosed || !r.event.Permission, Reason: "late_scan"})
			}
		}
		now := time.Now()
		for id, p := range active {
			if !now.Before(p.deadline) {
				m.count(func(s *Stats) { s.Timeouts++ })
				m.degrade(context.DeadlineExceeded)
				if err := resolve(id, p, Decision{}, context.DeadlineExceeded, "deadline"); err != nil {
					return err
				}
			}
		}
		m.count(func(s *Stats) { s.Pending = len(active) })
		readCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
		events, readErr := source.Read(readCtx)
		cancel()
		if readErr != nil && !errors.Is(readErr, context.DeadlineExceeded) && !errors.Is(readErr, context.Canceled) {
			for _, raw := range events {
				source.Release(raw)
			}
			return readErr
		}
		for index, raw := range events {
			// Release every already-read descriptor on any fatal batch error.
			abort := func(err error) error {
				for _, rest := range events[index+1:] {
					source.Release(rest)
				}
				return err
			}
			m.count(func(s *Stats) { s.Received++ })
			if raw.mask&QueueOverflow != 0 {
				source.Release(raw)
				m.count(func(s *Stats) { s.Overflows++ })
				m.degrade(ErrOverflow)
				return abort(ErrOverflow)
			}
			received := time.Now()
			ev, inspectErr := source.Inspect(raw)
			ev.Permission = raw.mask&(OpenPermission|OpenExecPermission) != 0
			if errors.Is(inspectErr, errNotRegular) {
				if ev.Permission {
					if err := source.Respond(raw, true); err != nil {
						source.Release(raw)
						return abort(err)
					}
				}
				source.Release(raw)
				continue
			}
			excluded := inspectErr == nil && (ev.PID == os.Getpid() || !m.inScope(ev.Path))
			// Inside a root, an uncertain name remains incomplete/fail-policy.
			if m.excludesAmbiguousName(ev.Path, inspectErr) {
				excluded = true
			}
			if excluded {
				if ev.Permission {
					if err := source.Respond(raw, true); err != nil {
						source.Release(raw)
						return abort(err)
					}
				}
				source.Release(raw)
				continue
			}
			if inspectErr == nil && (m.scan != nil || raw.mask&Modify != 0) {
				if err := source.Attribute(raw, &ev); err != nil {
					m.count(func(s *Stats) { s.AttributionLost++ })
					m.degrade(err)
				}
			}
			serial++
			id := serial
			deadline := received.Add(m.cfg.Deadline)
			jobCtx, jobCancel := context.WithDeadline(ctx, deadline)
			p := pending{raw: raw, event: ev, deadline: deadline, cancel: jobCancel}
			if inspectErr != nil {
				m.degrade(inspectErr)
				if err := resolve(id, p, Decision{}, inspectErr, "metadata"); err != nil {
					return abort(err)
				}
				continue
			}
			if m.scan == nil {
				source.Release(raw)
				jobCancel()
				deliver(Outcome{Event: ev, Allowed: true, Reason: "notification"})
				continue
			}
			if ev.Size < 0 || ev.Size > m.cfg.MaxFileBytes {
				m.degrade(ErrFileLimit)
				if err := resolve(id, p, Decision{}, ErrFileLimit, "file_limit"); err != nil {
					return abort(err)
				}
				continue
			}
			if len(active) >= m.cfg.MaxPending {
				m.count(func(s *Stats) { s.Overflows++ })
				m.degrade(ErrOverflow)
				if err := resolve(id, p, Decision{}, ErrOverflow, "admission"); err != nil {
					return abort(err)
				}
				continue
			}
			file, err := source.Duplicate(raw)
			if err != nil {
				m.degrade(err)
				if err = resolve(id, p, Decision{}, err, "descriptor"); err != nil {
					return abort(err)
				}
				continue
			}
			select {
			case jobs <- job{id: id, ctx: jobCtx, file: file, event: ev}:
				active[id] = p
			default:
				file.Close()
				m.count(func(s *Stats) { s.Overflows++ })
				m.degrade(ErrOverflow)
				if err := resolve(id, p, Decision{}, ErrOverflow, "admission"); err != nil {
					return abort(err)
				}
			}
		}
	}
}
