//go:build linux

package procwatch

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/linuxproc"
	"golang.org/x/sys/unix"
)

type connectorSocket struct {
	fd           int
	port, cookie uint32
	pending      []connectorEvent
}

func openConnector(ctx context.Context) (*connectorSocket, error) {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC|unix.SOCK_NONBLOCK, unix.NETLINK_CONNECTOR)
	if err != nil {
		return nil, err
	}
	s := &connectorSocket{fd: fd}
	ok := false
	defer func() {
		if !ok {
			unix.Close(fd)
		}
	}()
	if err = unix.Bind(fd, &unix.SockaddrNetlink{Family: unix.AF_NETLINK, Groups: 1}); err != nil {
		return nil, err
	}
	address, err := unix.Getsockname(fd)
	if err != nil {
		return nil, err
	}
	s.port = address.(*unix.SockaddrNetlink).Pid
	var nonce [4]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	s.cookie = binary.NativeEndian.Uint32(nonce[:])
	if s.cookie == ^uint32(0) {
		s.cookie--
	}
	if err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, 1<<20); err != nil {
		return nil, err
	}
	if err = unix.Sendto(fd, connectorSubscription(s.port, s.cookie, 1), 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		events, err := s.receive(ctx)
		if err != nil {
			return nil, err
		}
		acknowledged := false
		for _, e := range events {
			if e.Kind == connectorAck && e.Ack == s.cookie+1 {
				if e.Error != 0 {
					return nil, fmt.Errorf("process connector subscription rejected: %v", unix.Errno(e.Error))
				}
				acknowledged = true
			}
			if len(s.pending) >= 1024 {
				return nil, errors.New("process connector startup queue limit")
			}
			s.pending = append(s.pending, e)
		}
		if acknowledged {
			ok = true
			return s, nil
		}
	}
	return nil, errors.New("process connector acknowledgement timed out; initial namespaces and CONFIG_PROC_EVENTS required")
}
func (s *connectorSocket) close() {
	_ = unix.Sendto(s.fd, connectorSubscription(s.port, s.cookie, 2), unix.MSG_DONTWAIT, &unix.SockaddrNetlink{Family: unix.AF_NETLINK})
	_ = unix.Close(s.fd)
}
func (s *connectorSocket) receive(ctx context.Context) ([]connectorEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fds := []unix.PollFd{{Fd: int32(s.fd), Events: unix.POLLIN}}
	n, err := unix.Poll(fds, 100)
	if err == unix.EINTR {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	if fds[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
		return nil, errors.New("process connector queue/socket failure")
	}
	b := make([]byte, maxConnectorDatagram)
	n, _, flags, from, err := unix.Recvmsg(s.fd, b, nil, unix.MSG_DONTWAIT)
	if err == unix.EAGAIN || err == unix.EINTR {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if flags&unix.MSG_TRUNC != 0 {
		return nil, errors.New("process connector datagram truncated")
	}
	peer, ok := from.(*unix.SockaddrNetlink)
	if !ok || peer.Pid != 0 {
		return nil, errors.New("process connector rejected userspace sender")
	}
	return parseConnector(b[:n])
}

type connectorKnown struct {
	process   linuxproc.Process
	timestamp uint64
}

func (w *Watcher) runConnector(ctx context.Context) error {
	hz, err := connectorClockTicks()
	if err != nil {
		return err
	}
	_, subscribedBoot, subscribedAfter, err := connectorClockBracket()
	if err != nil {
		return err
	}
	var subscriptionOffsetLower uint64
	if subscribedBoot > subscribedAfter {
		subscriptionOffsetLower = subscribedBoot - subscribedAfter
	}
	s, err := openConnector(ctx)
	if err != nil {
		return err
	}
	defer s.close()
	known := map[int]connectorKnown{}
	retained := 0
	// Register pre-existing processes as current inventory, without inventing
	// a launch or triggering startup response. Identity-bound parent context is
	// then available to future command observations. Kernel subscription is
	// already active; queued events retain sequence-gap checks below.
	initial, err := w.reader.Snapshot(ctx, true)
	if err != nil {
		return err
	}
	for _, p := range initial {
		if p.PID <= 1 {
			continue
		}
		current, err := w.readConnectorProcess(p.PID)
		if err != nil {
			w.recordMiss(err)
			continue
		}
		_, _, observedAt, err := connectorClockBracket()
		if err != nil {
			return err
		}
		retained += len(current.Image) + len(current.Cmdline) + 128
		if retained > linuxproc.DefaultMaxSnapshotBytes {
			return errors.New("process inventory metadata byte limit exceeded")
		}
		known[current.PID] = connectorKnown{current, observedAt}
		w.h.OnProcStart(ProcStart{PID: current.PID, PPID: current.PPID, Image: current.Image, Cmdline: current.Cmdline, OSIdentity: current.StartTicks, IdentityRequired: true, EventKind: "inventory"})
	}
	w.setMode("connector")
	w.setHealth(nil)
	sequences := map[uint32]uint32{}
	for {
		events := s.pending
		s.pending = nil
		if len(events) == 0 {
			events, err = s.receive(ctx)
			if err != nil {
				return err
			}
		}
		for _, e := range events {
			if err := ctx.Err(); err != nil {
				return err
			}
			if e.CPU >= 8192 {
				return errors.New("process connector CPU index exceeds supported bound")
			}
			if last, ok := sequences[e.CPU]; ok && e.Sequence != last+1 {
				return errors.New("process connector sequence gap; process coverage lost")
			}
			sequences[e.CPU] = e.Sequence
			if e.PID <= 1 {
				continue
			}
			switch e.Kind {
			case connectorFork, connectorExec:
				p, err := w.readConnectorProcess(e.PID)
				if err != nil {
					// A kernel event still reached us, but its process may already be gone.
					// Never invent an executable or response identity for it.
					w.recordMiss(err)
					continue
				}
				before, boot, after, clockErr := connectorClockBracket()
				kind, observedAt := "snapshot", after
				if clockErr == nil && e.TimestampNS >= subscribedAfter && connectorBirthPredates(e.TimestampNS, p.StartTicks, hz, before, boot, after, subscriptionOffsetLower) {
					observedAt = e.TimestampNS
					kind = "fork"
					if e.Kind == connectorExec {
						kind = "exec"
					}
				} else {
					w.recordMiss(errors.New("connector event identity ambiguous; current image snapshot only"))
				}
				if clockErr != nil {
					return clockErr
				}
				old, exists := known[p.PID]
				if exists && old.timestamp > observedAt {
					kind, observedAt = "snapshot", after
					w.recordMiss(errors.New("queued event predates inventory; current snapshot only"))
				}
				if len(known) >= linuxproc.DefaultMaxProcesses && !exists {
					return errors.New("process connector metadata capacity exceeded")
				}
				nextBytes := len(p.Image) + len(p.Cmdline) + 128
				if exists {
					retained -= len(old.process.Image) + len(old.process.Cmdline) + 128
				}
				if retained+nextBytes > linuxproc.DefaultMaxSnapshotBytes {
					return errors.New("process connector metadata byte limit exceeded")
				}
				retained += nextBytes
				known[p.PID] = connectorKnown{p, observedAt}
				if (kind == "snapshot" || e.Kind != connectorExec) && exists && old.process.StartTicks == p.StartTicks && old.process.Image == p.Image && old.process.Cmdline == p.Cmdline {
					continue
				}
				w.h.OnProcStart(ProcStart{PID: p.PID, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline, EventKind: kind, OSIdentity: p.StartTicks, IdentityRequired: true})
			case connectorExit:
				if old, ok := known[e.PID]; ok && e.TimestampNS >= old.timestamp {
					if exits, ok := w.h.(interface{ OnProcExit(int, int64) }); ok {
						exits.OnProcExit(e.PID, old.process.StartTicks)
					}
					delete(known, e.PID)
					retained -= len(old.process.Image) + len(old.process.Cmdline) + 128
				}
			}
		}
	}
}

func connectorClockTicks() (uint64, error) {
	f, err := os.Open("/proc/self/auxv")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 65537))
	word := strconv.IntSize / 8
	if err != nil || len(b) > 65536 || len(b)%(2*word) != 0 {
		return 0, errors.New("invalid bounded auxiliary vector")
	}
	read := func(b []byte) uint64 {
		if word == 8 {
			return binary.NativeEndian.Uint64(b)
		}
		return uint64(binary.NativeEndian.Uint32(b))
	}
	for i := 0; i+2*word <= len(b); i += 2 * word {
		if read(b[i:]) == 17 {
			hz := read(b[i+word:])
			if hz > 0 && hz <= 1_000_000 {
				return hz, nil
			}
			break
		}
	}
	return 0, errors.New("process clock tick frequency unavailable")
}

func connectorClockBracket() (uint64, uint64, uint64, error) {
	var a, b, c unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &a); err != nil {
		return 0, 0, 0, err
	}
	if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &b); err != nil {
		return 0, 0, 0, err
	}
	if err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &c); err != nil {
		return 0, 0, 0, err
	}
	if a.Nano() < 0 || b.Nano() < 0 || c.Nano() < 0 {
		return 0, 0, 0, errors.New("invalid process clock bracket")
	}
	return uint64(a.Nano()), uint64(b.Nano()), uint64(c.Nano()), nil
}

func (w *Watcher) readConnectorProcess(pid int) (linuxproc.Process, error) {
	fd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		return linuxproc.Process{}, err
	}
	defer unix.Close(fd)
	p, err := w.reader.Read(pid, true)
	if err != nil {
		return linuxproc.Process{}, err
	}
	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	n, err := unix.Poll(fds, 0)
	if err != nil {
		return linuxproc.Process{}, err
	}
	if n != 0 {
		return linuxproc.Process{}, os.ErrNotExist
	}
	if p.StartTicks <= 0 {
		return linuxproc.Process{}, errors.New("process connector could not capture response identity")
	}
	return p, nil
}
