//go:build linux

package netblock

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

const nftRuntimeDir = "/run/utmstack-edr"

type nftExec struct{ path string }
type boundedNFTBuffer struct {
	bytes.Buffer
	limit    int
	exceeded bool
	cancel   context.CancelFunc
}

func (b *boundedNFTBuffer) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.limit - b.Len()
	if remaining > 0 {
		if remaining > len(p) {
			remaining = len(p)
		}
		_, _ = b.Buffer.Write(p[:remaining])
	}
	if n > remaining {
		b.exceeded = true
		b.cancel()
	}
	return n, nil
}

func (x nftExec) Run(parent context.Context, input []byte, args ...string) ([]byte, error) {
	if len(input) > nftMaxPayload {
		return nil, ErrNetworkBudget
	}
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, x.path, args...)
	cmd.Env = []string{"PATH=/usr/sbin:/usr/bin:/sbin:/bin", "LANG=C", "LC_ALL=C"}
	cmd.Stdin = bytes.NewReader(input)
	cmd.WaitDelay = 250 * time.Millisecond
	out := &boundedNFTBuffer{limit: nftMaxOutput, cancel: cancel}
	stderr := &boundedNFTBuffer{limit: 64 << 10, cancel: cancel}
	cmd.Stdout = out
	cmd.Stderr = stderr
	err := cmd.Run()
	if out.exceeded || stderr.exceeded {
		return nil, ErrNetworkBudget
	}
	if ctx.Err() != nil {
		return nil, fmt.Errorf("nftables subprocess deadline: %w", ctx.Err())
	}
	if err != nil {
		return nil, fmt.Errorf("nftables command failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if len(args) == 6 && args[0] == "-j" && args[1] == "-n" && args[2] == "list" && args[3] == "table" && args[4] == "inet" {
		return normalizeNFTReadback(out.Bytes())
	}
	return out.Bytes(), nil
}

func trustedNFTPath() (string, error) {
	for _, candidate := range []string{"/usr/sbin/nft", "/sbin/nft", "/usr/bin/nft", "/bin/nft"} {
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			continue
		}
		info, err := os.Stat(resolved)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			continue
		}
		trusted := true
		for path := resolved; ; path = filepath.Dir(path) {
			var st unix.Stat_t
			if unix.Lstat(path, &st) != nil || st.Uid != 0 || st.Mode&0022 != 0 {
				trusted = false
				break
			}
			if path == "/" {
				break
			}
		}
		if trusted {
			return resolved, nil
		}
	}
	return "", errors.New("a root-owned, non-writable nft executable is required")
}

func secureNFTRuntime(runtimeDir string) error {
	if os.Geteuid() != 0 {
		return errors.New("nftables backend requires root in its network namespace")
	}
	if err := os.Mkdir(runtimeDir, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	var st unix.Stat_t
	if err := unix.Lstat(runtimeDir, &st); err != nil {
		return err
	}
	if st.Uid != 0 || st.Mode&unix.S_IFMT != unix.S_IFDIR || st.Mode&0077 != 0 {
		return errors.New("unsafe nftables runtime directory")
	}
	return nil
}

func nftLock(runtimeDir, table string) func(context.Context) (func(), error) {
	return func(ctx context.Context) (func(), error) {
		fd, err := unix.Open(filepath.Join(runtimeDir, table+".lock"), unix.O_RDWR|unix.O_CREAT|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
		if err != nil {
			return nil, err
		}
		var st unix.Stat_t
		if err = unix.Fstat(fd, &st); err != nil || st.Uid != 0 || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Mode&0077 != 0 || st.Nlink != 1 {
			_ = unix.Close(fd)
			return nil, errors.New("unsafe nftables lock file")
		}
		for {
			err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
			if err == nil {
				var once sync.Once
				return func() { once.Do(func() { _ = unix.Flock(fd, unix.LOCK_UN); _ = unix.Close(fd) }) }, nil
			}
			if err != unix.EWOULDBLOCK && err != unix.EAGAIN {
				_ = unix.Close(fd)
				return nil, err
			}
			timer := time.NewTimer(25 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				_ = unix.Close(fd)
				return nil, fmt.Errorf("nftables ownership lock: %w", ctx.Err())
			case <-timer.C:
			}
		}
	}
}

// A random root-only token prevents collision with another program's same-name
// table. It is runtime-scoped; reboot also clears the kernel network namespace.
func nftOwner(runtimeDir, table string) (string, error) {
	name := filepath.Join(runtimeDir, table+".owner")
	fd, err := unix.Open(name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if errors.Is(err, unix.ENOENT) {
		raw := make([]byte, 16)
		if _, err = rand.Read(raw); err != nil {
			return "", err
		}
		value := []byte(hex.EncodeToString(raw))
		fd, err = unix.Open(name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
		if err != nil {
			return "", err
		}
		f := os.NewFile(uintptr(fd), name)
		_, writeErr := f.Write(value)
		syncErr := f.Sync()
		closeErr := f.Close()
		if err = errors.Join(writeErr, syncErr, closeErr); err != nil {
			return "", err
		}
		return "utmstack-edr-v1-" + string(value), nil
	}
	if err != nil {
		return "", err
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	var st unix.Stat_t
	if err = unix.Fstat(fd, &st); err != nil || st.Uid != 0 || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Mode&0077 != 0 || st.Nlink != 1 {
		return "", errors.New("unsafe nftables ownership token")
	}
	b, err := io.ReadAll(io.LimitReader(f, 33))
	if err != nil || len(b) != 32 {
		return "", errors.New("invalid nftables ownership token")
	}
	if _, err = hex.DecodeString(string(b)); err != nil {
		return "", err
	}
	return "utmstack-edr-v1-" + string(b), nil
}

func linuxController(table, runtimeDir string) (*nftController, error) {
	path, err := trustedNFTPath()
	if err != nil {
		return nil, err
	}
	if err = secureNFTRuntime(runtimeDir); err != nil {
		return nil, err
	}
	return &nftController{ops: nftExec{path}, table: table, now: time.Now}, nil
}

func NewOSBlocker(name string) (Blocker, error) {
	return newLinuxBlocker(name, nftRuntimeDir)
}

func newLinuxBlocker(name, runtimeDir string) (*nftBlocker, error) {
	h := sha256.Sum256([]byte(name))
	table := "utmstack_edr_bl_" + hex.EncodeToString(h[:4])
	c, err := linuxController(table, runtimeDir)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock, err := nftLock(runtimeDir, table)(ctx)
	if err != nil {
		return nil, err
	}
	c.owner, err = nftOwner(runtimeDir, table)
	if err != nil {
		unlock()
		return nil, err
	}
	return newNFTBlocker(c, unlock), nil
}

func NewHostIsolator() (*HostIsolator, error) {
	return newLinuxHostIsolator(nftRuntimeDir)
}

func newLinuxHostIsolator(runtimeDir string) (*HostIsolator, error) {
	c, err := linuxController("utmstack_edr_isolation", runtimeDir)
	if err != nil {
		return nil, err
	}
	c.lock = nftLock(runtimeDir, c.table)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock, err := c.lock(ctx)
	if err != nil {
		return nil, err
	}
	defer unlock()
	c.owner, err = nftOwner(runtimeDir, c.table)
	if err != nil {
		return nil, err
	}
	return &HostIsolator{controller: c}, nil
}
