//go:build linux

package fanotify

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// Broker owns only its scanner pool and child lifetime. The child is the sole
// owner of the kernel group, pending requests, deadlines and kernel responses.
type Broker struct {
	cfg           Config
	scan          Scanner
	observe       func(Outcome)
	mu            sync.Mutex
	used, running bool
	failure       error
	stats         Stats
	lastStatus    time.Time
	command       func() *exec.Cmd
}

func NewPermissionBroker(cfg Config, scan Scanner, observe func(Outcome)) (*Broker, error) {
	if err := validateBrokerConfig(cfg); err != nil {
		return nil, err
	}
	if cfg.Mode == ModeNotify || scan == nil {
		return nil, errors.New("permission broker requires a permission mode and scanner")
	}
	cfg.Mounts = append([]string(nil), cfg.Mounts...)
	cfg.Roots = append([]string(nil), cfg.Roots...)
	return &Broker{cfg: cfg, scan: scan, observe: observe, command: func() *exec.Cmd {
		// The kernel resolves this to the already running executable, including
		// an unlinked image. A replaced installation pathname is not authority.
		cmd := exec.Command("/proc/self/exe", "__fanotify-broker")
		cmd.Stderr = os.Stderr
		// EOF on the private close-on-exec channel tracks the parent process.
		// Pdeathsig instead tracks its creating OS thread and can fire while
		// the Go parent is still alive.
		return cmd
	}}, nil
}

func (b *Broker) Health() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.failure != nil {
		return b.failure
	}
	if !b.running || b.lastStatus.IsZero() {
		return errors.New("fanotify broker is not ready")
	}
	if time.Since(b.lastStatus) > time.Second {
		return errors.New("fanotify broker status expired")
	}
	return nil
}
func (b *Broker) Stats() Stats { b.mu.Lock(); defer b.mu.Unlock(); return b.stats }

func (b *Broker) Run(ctx context.Context) (runErr error) {
	b.mu.Lock()
	if b.used {
		b.mu.Unlock()
		return ErrUsed
	}
	b.used = true
	b.mu.Unlock()
	defer func() {
		b.mu.Lock()
		b.running = false
		b.stats.Pending = 0
		if runErr != nil && b.failure == nil {
			b.failure = runErr
		}
		b.mu.Unlock()
	}()
	if err := ctx.Err(); err != nil {
		return err
	}
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return err
	}
	parent, child := os.NewFile(uintptr(fds[0]), "fanotify-scanner-channel"), os.NewFile(uintptr(fds[1]), "fanotify-owner-channel")
	defer parent.Close()
	defer child.Close()
	cmd := b.command()
	cmd.ExtraFiles = []*os.File{child}
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("fanotify broker start: %w", err)
	}
	child.Close()
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer func() {
		cancel()
		parent.Close()
		select {
		case e := <-exited:
			if runErr == nil && e != nil {
				runErr = fmt.Errorf("fanotify broker exit: %w", e)
			}
		case <-time.After(time.Second):
			cmd.Process.Kill()
			<-exited
			if runErr == nil {
				runErr = errors.New("fanotify broker required forced shutdown")
			}
		}
	}()
	status := func(s Stats, e error) {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.running = true
		b.stats = s
		b.lastStatus = time.Now()
		if e != nil && b.failure == nil {
			b.failure = e
		}
	}
	return RunBrokerClient(ctx, parent, b.cfg, b.scan, b.observe, status)
}
