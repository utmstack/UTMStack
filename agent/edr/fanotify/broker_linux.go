//go:build linux

package fanotify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"golang.org/x/sys/unix"
)

// The broker protocol is private to a same-version pair of processes. It never
// transmits the fanotify group or accepts a pathname in place of an event FD.
const brokerVersion = 1
const brokerPacketLimit = 16 << 10

func validateBrokerConfig(cfg Config) error {
	if err := Validate(cfg); err != nil {
		return err
	}
	for _, paths := range [][]string{cfg.Mounts, cfg.Roots} {
		for _, path := range paths {
			if !utf8.ValidString(path) {
				return errors.New("fanotify broker configured scope requires UTF-8 paths")
			}
		}
	}
	data, err := json.Marshal(brokerMessage{Version: brokerVersion, Kind: "configure", Config: &cfg})
	if err != nil {
		return err
	}
	if len(data) > brokerPacketLimit {
		return errors.New("fanotify broker configuration exceeds 16 KiB protocol bound")
	}
	return nil
}

type brokerMessage struct {
	Version  int      `json:"version"`
	Kind     string   `json:"kind"`
	ID       uint64   `json:"id,omitempty"`
	Config   *Config  `json:"config,omitempty"`
	Event    Event    `json:"event"`
	Path     []byte   `json:"path,omitempty"` // Unix names are bytes, not JSON text.
	Deadline int64    `json:"deadline,omitempty"`
	Decision Decision `json:"decision"`
	Error    string   `json:"error,omitempty"`
	Allowed  bool     `json:"allowed,omitempty"`
	Reason   string   `json:"reason,omitempty"`
	Stats    Stats    `json:"stats"`
}

type brokerSocket struct {
	fd       int
	sendGate chan struct{}
}

func (s *brokerSocket) close() {
	s.sendGate <- struct{}{}
	defer func() { <-s.sendGate }()
	unix.Close(s.fd)
}

func brokerSocketFromFile(f *os.File) (*brokerSocket, error) {
	if f == nil {
		return nil, errors.New("fanotify broker requires an inherited socket")
	}
	fd, err := unix.FcntlInt(f.Fd(), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	fail := func(e error) (*brokerSocket, error) { unix.Close(fd); return nil, e }
	kind, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_TYPE)
	if err != nil || kind != unix.SOCK_SEQPACKET {
		return fail(errors.New("fanotify broker requires a Unix seqpacket socket"))
	}
	addr, err := unix.Getsockname(fd)
	if err != nil {
		return fail(err)
	}
	if _, ok := addr.(*unix.SockaddrUnix); !ok {
		return fail(errors.New("fanotify broker requires a Unix socket"))
	}
	return &brokerSocket{fd: fd, sendGate: make(chan struct{}, 1)}, nil
}

// All data operations use MSG_DONTWAIT. A stopped scanner cannot block the
// broker's response owner, even when its socket receive buffer is full.
func (s *brokerSocket) send(ctx context.Context, m brokerMessage, file *os.File) error {
	m.Version = brokerVersion
	m.Path, m.Event.Path = []byte(m.Event.Path), ""
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if len(data) > brokerPacketLimit {
		return errors.New("fanotify broker packet too large")
	}
	var rights []byte
	if file != nil {
		rights = unix.UnixRights(int(file.Fd()))
	}
	select {
	case s.sendGate <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-s.sendGate }()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := unix.SendmsgN(s.fd, data, rights, nil, unix.MSG_DONTWAIT|unix.MSG_NOSIGNAL)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if errors.Is(err, unix.EAGAIN) {
			if err = brokerPoll(ctx, s.fd, unix.POLLOUT); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if n != len(data) {
			return io.ErrShortWrite
		}
		return nil
	}
}

func brokerPoll(ctx context.Context, fd int, events int16) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		p := []unix.PollFd{{Fd: int32(fd), Events: events}}
		n, err := unix.Poll(p, 10)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		if p[0].Revents&events != 0 {
			return nil
		}
		return io.EOF
	}
}

func (s *brokerSocket) receive(ctx context.Context) (brokerMessage, *os.File, error) {
	var m brokerMessage
	var data [brokerPacketLimit]byte
	var control [256]byte
	for {
		if err := ctx.Err(); err != nil {
			return m, nil, err
		}
		n, cn, flags, _, err := unix.Recvmsg(s.fd, data[:], control[:], unix.MSG_DONTWAIT|unix.MSG_CMSG_CLOEXEC)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if errors.Is(err, unix.EAGAIN) {
			if err = brokerPoll(ctx, s.fd, unix.POLLIN); err != nil {
				return m, nil, err
			}
			continue
		}
		if err != nil {
			return m, nil, err
		}
		var fds []int
		messages, parseErr := unix.ParseSocketControlMessage(control[:cn])
		for _, c := range messages {
			if c.Header.Level != unix.SOL_SOCKET || c.Header.Type != unix.SCM_RIGHTS {
				parseErr = errors.New("unexpected broker control message")
				continue
			}
			rights, e := unix.ParseUnixRights(&c)
			if e != nil {
				parseErr = e
			}
			fds = append(fds, rights...)
		}
		// Close every received FD on all invalid/truncated paths, including rights
		// received alongside otherwise malformed JSON.
		fail := func(e error) (brokerMessage, *os.File, error) {
			for _, fd := range fds {
				unix.Close(fd)
			}
			return brokerMessage{}, nil, e
		}
		if n == 0 {
			return fail(io.EOF)
		}
		if flags&(unix.MSG_TRUNC|unix.MSG_CTRUNC) != 0 || parseErr != nil || len(fds) > 1 {
			return fail(errors.New("invalid or truncated broker packet"))
		}
		d := json.NewDecoder(bytes.NewReader(data[:n]))
		d.DisallowUnknownFields()
		if err = d.Decode(&m); err != nil {
			return fail(err)
		}
		var trailing any
		if err = d.Decode(&trailing); err != io.EOF {
			return fail(errors.New("trailing broker data"))
		}
		if m.Version != brokerVersion || m.Event.Path != "" || len(m.Path) > 4096 {
			return fail(errors.New("incompatible broker packet"))
		}
		m.Event.Path = string(m.Path)
		if len(fds) == 1 {
			return m, os.NewFile(uintptr(fds[0]), "fanotify-broker-pinned-input"), nil
		}
		return m, nil, nil
	}
}

type brokerReply struct {
	decision Decision
	err      error
}

// ServeBroker runs only in an independent process. The inherited connected
// socket is the sole authority: there is no listener, pathname RPC or group FD
// transfer. Only this verified independent owner can mark executable mounts.
// The service selects this owner for permission modes; in-process Monitor
// activation retains its executable-mount refusal.
func ServeBroker(ctx context.Context, channel *os.File) (runErr error) {
	s, err := brokerSocketFromFile(channel)
	if err != nil {
		return err
	}
	defer s.close()
	defer func() {
		if runErr == nil || errors.Is(runErr, context.Canceled) || errors.Is(runErr, io.EOF) || errors.Is(runErr, unix.ECONNRESET) {
			return
		}
		m := brokerMessage{Kind: "failure", Error: runErr.Error()}
		if len(m.Error) > 4096 {
			m.Error = m.Error[:4096]
		}
		if errors.Is(runErr, ErrUnsupported) {
			m.Reason = "unsupported"
		}
		write, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		_ = s.send(write, m, nil)
	}()
	peer, err := unix.GetsockoptUcred(s.fd, unix.SOL_SOCKET, unix.SO_PEERCRED)
	if err != nil || int(peer.Pid) != os.Getppid() || peer.Uid != uint32(os.Geteuid()) {
		return errors.New("fanotify broker channel must be inherited from its same-user parent")
	}
	return serveBroker(ctx, s, func(cfg Config) (backend, error) { return openNativeBackend(cfg, true) })
}

func serveBroker(ctx context.Context, s *brokerSocket, open func(Config) (backend, error)) error {
	setup, cancelSetup := context.WithTimeout(ctx, 5*time.Second)
	m, f, err := s.receive(setup)
	cancelSetup()
	if f != nil {
		f.Close()
	}
	if err != nil {
		return err
	}
	if f != nil || m.Kind != "configure" || m.Config == nil || m.ID != 0 {
		return errors.New("fanotify broker requires configuration first")
	}
	cfg := *m.Config
	if err = validateBrokerConfig(cfg); err != nil {
		return err
	}
	if cfg.Mode == ModeNotify {
		return errors.New("fanotify broker is for permission modes")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var serial atomic.Uint64
	var mu sync.Mutex
	waiting := map[uint64]chan brokerReply{}
	var monitor *Monitor
	scan := func(scanCtx context.Context, f *os.File, ev Event) (Decision, error) {
		id := serial.Add(1)
		replies := make(chan brokerReply, 1)
		mu.Lock()
		waiting[id] = replies
		mu.Unlock()
		defer func() { mu.Lock(); delete(waiting, id); mu.Unlock() }()
		deadline, _ := scanCtx.Deadline()
		if err := s.send(scanCtx, brokerMessage{Kind: "scan", ID: id, Event: ev, Deadline: deadline.UnixNano()}, f); err != nil {
			return Decision{}, err
		}
		// Monitor owns the kernel deadline and late-result reporting. Retain
		// this worker's reply slot until completion or transport shutdown, so a
		// decision crossing the deadline cannot disappear between processes.
		// There are at most cfg.Workers slots, even for a stopped scanner.
		select {
		case r := <-replies:
			if r.err != nil {
				monitor.degrade(r.err)
			}
			return r.decision, r.err
		case <-ctx.Done():
			return Decision{}, ctx.Err()
		}
	}
	// Observer traffic must never wait behind a scanner's full receive buffer.
	observe := func(o Outcome) {
		c, done := context.WithTimeout(ctx, 100*time.Millisecond)
		defer done()
		m := brokerMessage{Kind: "outcome", Event: o.Event, Decision: o.Decision, Allowed: o.Allowed, Reason: o.Reason}
		if o.Err != nil {
			m.Error = o.Err.Error()
		}
		if err := s.send(c, m, nil); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				monitor.count(func(s *Stats) { s.ObserverLost++ })
				monitor.degrade(ErrOverflow)
			} else {
				cancel()
			}
		}
	}
	monitor, err = New(cfg, scan, observe)
	if err != nil {
		return err
	}
	monitor.open = open
	readDone := make(chan error, 1)
	go func() {
		for {
			m, f, e := s.receive(ctx)
			if f != nil {
				f.Close()
			}
			if e != nil {
				readDone <- e
				cancel()
				return
			}
			if f != nil || m.Kind != "decision" || m.ID == 0 || len(m.Decision.Signature) > 4096 || len(m.Decision.Generation) > 4096 || len(m.Error) > 4096 {
				readDone <- errors.New("invalid broker decision")
				cancel()
				return
			}
			mu.Lock()
			reply := waiting[m.ID]
			mu.Unlock()
			if reply != nil {
				var e error
				if m.Error != "" {
					e = errors.New(m.Error)
				}
				select {
				case reply <- brokerReply{m.Decision, e}:
				default:
				}
			}
		}
	}()
	statusDone := make(chan struct{})
	go func() {
		defer close(statusDone)
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			monitor.mu.Lock()
			starting := !monitor.running && monitor.failure == nil
			monitor.mu.Unlock()
			if starting {
				continue
			}
			m := brokerMessage{Kind: "status", Stats: monitor.Stats()}
			if e := monitor.Health(); e != nil {
				m.Error = e.Error()
			}
			c, done := context.WithTimeout(ctx, 50*time.Millisecond)
			e := s.send(c, m, nil)
			done()
			// A delayed/dropped heartbeat is not loss of kernel authority. Keep
			// resolving permissions; the parent independently expires status.
			if e != nil && !errors.Is(e, context.DeadlineExceeded) {
				cancel()
				return
			}
		}
	}()
	err = monitor.Run(ctx)
	cancel()
	<-statusDone
	readErr := <-readDone
	if errors.Is(err, context.Canceled) && !errors.Is(readErr, context.Canceled) {
		return readErr
	}
	return err
}

// RunBrokerClient serves one inherited socket with a fixed scanner pool. The
// caller owns process startup/reaping and must close its copy of the child's
// socket. Cancellation closes the channel after at most the bounded poll wait.
// Broker status (including missing heartbeats) is reported separately from scan
// outcomes; an application must not infer healthy enforcement from this goroutine.
func RunBrokerClient(ctx context.Context, channel *os.File, cfg Config, scan Scanner, observe func(Outcome), status func(Stats, error)) error {
	if err := validateBrokerConfig(cfg); err != nil {
		return err
	}
	if cfg.Mode == ModeNotify || scan == nil {
		return errors.New("fanotify broker client requires a permission scanner")
	}
	s, err := brokerSocketFromFile(channel)
	if err != nil {
		return err
	}
	defer s.close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	setup, done := context.WithTimeout(ctx, 5*time.Second)
	err = s.send(setup, brokerMessage{Kind: "configure", Config: &cfg}, nil)
	done()
	if err != nil {
		return err
	}
	type request struct {
		message brokerMessage
		file    *os.File
	}
	jobs := make(chan request, cfg.MaxPending)
	callbacks := make(chan brokerMessage, cfg.MaxPending)
	// Workers may finish after RunBrokerClient returns. Never close their
	// callback queue; cancellation bounds retained work and stops its consumer.
	defer func() {
		close(jobs)
		for j := range jobs {
			j.file.Close()
		}
	}()
	// A scanner ignoring cancellation can retain only Workers objects. We do
	// not wait for such a scanner during shutdown, as with Monitor.Run.
	for i := 0; i < cfg.Workers; i++ {
		go func() {
			for j := range jobs {
				func() {
					defer j.file.Close()
					c, done := context.WithDeadline(ctx, time.Unix(0, j.message.Deadline))
					defer done()
					var d Decision
					var e error
					func() {
						defer func() {
							if p := recover(); p != nil {
								e = fmt.Errorf("broker scanner panic: %v", p)
							}
						}()
						if e = c.Err(); e == nil {
							d, e = scan(c, j.file, j.message.Event)
						}
					}()
					m := brokerMessage{Kind: "decision", ID: j.message.ID, Decision: d}
					if e != nil {
						m.Error = e.Error()
					}
					// Completion must outlive the scan deadline, including requests
					// that expired in the queue. Only the owner knows whether the
					// kernel response has already been sent. A failed delivery makes
					// the transport unhealthy instead of silently losing the result.
					write, stop := context.WithTimeout(ctx, 100*time.Millisecond)
					defer stop()
					if err := s.send(write, m, nil); err != nil {
						cancel()
					}
				}()
			}
		}()
	}
	// User observers are decoupled from the transport just like local Monitor
	// observers. A blocked observer must not stop receiving scan requests.
	go func() {
		for {
			var m brokerMessage
			select {
			case <-ctx.Done():
				return
			case m = <-callbacks:
			}
			func() {
				defer func() {
					if recover() != nil {
						cancel()
					}
				}()
				var e error
				if m.Error != "" {
					e = errors.New(m.Error)
				}
				if m.Kind == "status" {
					if status != nil {
						status(m.Stats, e)
					}
				} else if observe != nil {
					observe(Outcome{Event: m.Event, Decision: m.Decision, Err: e, Allowed: m.Allowed, Reason: m.Reason})
				}
			}()
		}
	}()
	statusReceived := false
	for {
		readBudget := time.Second
		if !statusReceived {
			readBudget = 5 * time.Second
		}
		read, stop := context.WithTimeout(ctx, readBudget)
		m, f, e := s.receive(read)
		stop()
		if e != nil {
			if f != nil {
				f.Close()
			}
			return e
		}
		switch m.Kind {
		case "failure":
			if f != nil {
				f.Close()
				return errors.New("unexpected broker descriptor")
			}
			if m.Reason == "unsupported" {
				return fmt.Errorf("%w: %s", ErrUnsupported, m.Error)
			}
			return fmt.Errorf("fanotify broker: %s", m.Error)
		case "scan":
			if f == nil || m.ID == 0 || !m.Event.Permission || m.Deadline <= 0 || m.Deadline > time.Now().Add(cfg.Deadline+time.Second).UnixNano() {
				if f != nil {
					f.Close()
				}
				return errors.New("invalid broker scan request")
			}
			select {
			case jobs <- request{m, f}:
			default:
				f.Close()
				return ErrOverflow
			}
		case "status", "outcome":
			if f != nil {
				f.Close()
				return errors.New("unexpected broker descriptor")
			}
			if m.Kind == "status" {
				statusReceived = true
			}
			select {
			case callbacks <- m:
			default:
				return ErrOverflow
			}
		default:
			if f != nil {
				f.Close()
			}
			return errors.New("unexpected broker message")
		}
	}
}
