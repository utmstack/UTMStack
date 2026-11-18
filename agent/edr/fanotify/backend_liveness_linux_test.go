//go:build linux

package fanotify

import (
	"errors"
	"os"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestPIDFDLivenessInterruptedProbe(t *testing.T) {
	for _, exited := range []bool{false, true} {
		t.Run(map[bool]string{false: "live", true: "exited"}[exited], func(t *testing.T) {
			calls := 0
			alive, err := pidfdAlivePoll(17, func(p []unix.PollFd, timeout int) (int, error) {
				calls++
				if len(p) != 1 || p[0].Fd != 17 || timeout != 0 {
					t.Fatal("liveness probe became blocking or changed its descriptor")
				}
				if calls == 1 {
					return 0, unix.EINTR
				}
				if exited {
					p[0].Revents = unix.POLLIN
					return 1, nil
				}
				return 0, nil
			})
			if err != nil || alive == exited || calls != 2 {
				t.Fatalf("alive=%v err=%v calls=%d", alive, err, calls)
			}
		})
	}
}

func TestPIDFDLivenessFailureBounds(t *testing.T) {
	for _, tc := range []struct {
		name      string
		fd        int
		err       error
		ready     int16
		wantCalls int
		wantError error
	}{
		{name: "interruptions bounded", fd: 17, err: unix.EINTR, wantCalls: 8, wantError: unix.EINTR},
		{name: "real error retained", fd: 17, err: unix.EIO, wantCalls: 1, wantError: unix.EIO},
		{name: "invalid descriptor readiness", fd: 17, ready: unix.POLLNVAL, wantCalls: 1, wantError: unix.EBADF},
		{name: "missing kernel descriptor", fd: -1, wantCalls: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			alive, err := pidfdAlivePoll(tc.fd, func(p []unix.PollFd, _ int) (int, error) {
				calls++
				p[0].Revents = tc.ready
				if tc.ready != 0 {
					return 1, tc.err
				}
				return 0, tc.err
			})
			if alive || err == nil || calls != tc.wantCalls || tc.wantError != nil && !errors.Is(err, tc.wantError) {
				t.Fatalf("alive=%v err=%v calls=%d", alive, err, calls)
			}
		})
	}
}

func TestLinuxPIDFDAttributionOwnProcess(t *testing.T) {
	fd, err := unix.PidfdOpen(os.Getpid(), 0)
	if errors.Is(err, unix.ENOSYS) {
		t.Skip("kernel has no pidfd_open")
	}
	if err != nil {
		t.Fatal(err)
	}
	defer unix.Close(fd)
	backend := &nativeBackend{permissionGroup: true}
	var event Event
	if err := backend.Attribute(rawEvent{pid: os.Getpid(), pidfd: fd}, &event); err != nil {
		t.Fatal(err)
	}
	if event.OSIdentity <= 0 || event.Image != "" || event.Cmdline != "" {
		t.Fatalf("invalid identity-only permission attribution: %+v", event)
	}
	err = backend.Attribute(rawEvent{pid: os.Getpid(), pidfd: -1}, &Event{})
	if !errors.Is(err, ErrAttribution) || !strings.Contains(err.Error(), "pidfd not live before metadata") || !strings.Contains(err.Error(), "pid=") {
		t.Fatalf("attribution failure lost stage/process diagnostics: %v", err)
	}
}
