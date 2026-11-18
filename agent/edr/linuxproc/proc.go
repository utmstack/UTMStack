// Package linuxproc reads bounded Linux procfs metadata without executing an
// input. It is portable so parsers and PID-generation logic can be tested with
// an isolated synthetic proc tree on development hosts.
package linuxproc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	MaxStatBytes            = 16 << 10
	MaxCommandBytes         = 64 << 10
	DefaultMaxProcesses     = 32768
	DefaultMaxSnapshotBytes = 16 << 20
)

var ErrLimit = errors.New("Linux process metadata limit exceeded")

type Process struct {
	PID, PPID      int
	StartTicks     int64
	State          byte
	Image, Cmdline string
}

// ParseStat uses the final ')' because comm can contain spaces and parentheses.
// starttime is field 22: a monotonic process creation tick within this boot.
func ParseStat(data []byte) (Process, error) {
	if len(data) > MaxStatBytes {
		return Process{}, ErrLimit
	}
	s := string(data)
	open := strings.IndexByte(s, '(')
	close := strings.LastIndexByte(s, ')')
	if open < 1 || close <= open || close+2 >= len(s) || s[close+1] != ' ' {
		return Process{}, errors.New("malformed process stat")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(s[:open]))
	if err != nil || pid <= 0 {
		return Process{}, errors.New("invalid process PID")
	}
	f := strings.Fields(s[close+2:])
	if len(f) < 20 || len(f[0]) != 1 {
		return Process{}, errors.New("short process stat")
	}
	parent, err := strconv.Atoi(f[1])
	if err != nil || parent < 0 {
		return Process{}, errors.New("invalid parent PID")
	}
	start, err := strconv.ParseInt(f[19], 10, 64)
	// Early kernel processes may start at tick zero. Metadata must remain
	// visible; response code separately refuses zero identity and PID 1.
	if err != nil || start < 0 {
		return Process{}, errors.New("invalid process creation tick")
	}
	return Process{PID: pid, PPID: parent, StartTicks: start, State: f[0][0]}, nil
}

type Reader struct {
	Root             string
	MaxProcesses     int
	MaxSnapshotBytes int
}

func (r Reader) root() string {
	if r.Root != "" {
		return r.Root
	}
	return "/proc"
}
func bounded(path string, limit int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil {
		return nil, err
	}
	if len(b) > limit {
		return nil, ErrLimit
	}
	return b, nil
}
func (r Reader) Read(pid int, details bool) (Process, error) {
	if pid <= 0 {
		return Process{}, errors.New("invalid process PID")
	}
	dir := filepath.Join(r.root(), strconv.Itoa(pid))
	readStat := func() (Process, error) {
		b, e := bounded(filepath.Join(dir, "stat"), MaxStatBytes)
		if e != nil {
			return Process{}, e
		}
		p, e := ParseStat(b)
		if e == nil && p.PID != pid {
			e = errors.New("process PID changed")
		}
		return p, e
	}
	p, err := readStat()
	if err != nil {
		return Process{}, err
	}
	if !details {
		return p, nil
	}
	image, err := os.Readlink(filepath.Join(dir, "exe"))
	if err != nil {
		return Process{}, err
	}
	if len(image) > 4096 || !filepath.IsAbs(image) {
		return Process{}, errors.New("invalid process image")
	}
	cmd, err := bounded(filepath.Join(dir, "cmdline"), MaxCommandBytes)
	if err != nil {
		return Process{}, err
	}
	after, err := readStat()
	if err != nil {
		return Process{}, err
	}
	if after.StartTicks != p.StartTicks {
		return Process{}, errors.New("process identity changed during lookup")
	}
	p.Image = image
	p.Cmdline = strings.TrimSpace(strings.ReplaceAll(string(cmd), "\x00", " "))
	return p, nil
}

// Snapshot refuses an unbounded enumeration and returns partial metadata only
// alongside an error. Disappearing/kernel processes are skipped; permission
// errors for user process metadata are returned so coverage cannot look healthy.
func (r Reader) Snapshot(ctx context.Context, details bool) ([]Process, error) {
	maxProcesses := r.MaxProcesses
	if maxProcesses <= 0 {
		maxProcesses = DefaultMaxProcesses
	}
	if maxProcesses > DefaultMaxProcesses {
		return nil, ErrLimit
	}
	maxBytes := r.MaxSnapshotBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxSnapshotBytes
	}
	if maxBytes > DefaultMaxSnapshotBytes {
		return nil, ErrLimit
	}
	retained := 0
	f, err := os.Open(r.root())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Process
	count := 0
	for {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		entries, readErr := f.ReadDir(128)
		for _, entry := range entries {
			pid, err := strconv.Atoi(entry.Name())
			if err != nil || pid <= 0 || !entry.IsDir() {
				continue
			}
			count++
			if count > maxProcesses {
				return out, ErrLimit
			}
			p, err := r.Read(pid, details)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return out, fmt.Errorf("read process %d: %w", pid, err)
			}
			retained += len(p.Image) + len(p.Cmdline) + 128
			if retained > maxBytes {
				return out, ErrLimit
			}
			out = append(out, p)
		}
		if readErr == io.EOF {
			return out, nil
		}
		if readErr != nil {
			return out, readErr
		}
	}
}
