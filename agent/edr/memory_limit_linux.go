package main

import (
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
)

const (
	goMemoryMiB     = uint64(1 << 20)
	goMemoryGiB     = uint64(1 << 30)
	maxAutoGoMemory = uint64(1000) * goMemoryMiB
)

// configureGoMemoryLimit leaves an explicit operator setting alone. The
// official full feed has been qualified under 1, 1.5 and 2 GiB Linux cgroups
// with a Go soft target of 75% of the cap, up to 1000 MiB. Set that target
// before signature loading, and leave other host sizes to their normal Go
// runtime or operator policy until separately qualified.
func configureGoMemoryLimit() int64 {
	target := desiredGoMemoryLimit(os.Getenv("GOMEMLIMIT"), readCgroupMemoryLimit("/sys/fs/cgroup"))
	if target > 0 {
		debug.SetMemoryLimit(target)
	}
	return target
}

func desiredGoMemoryLimit(explicit string, cgroupLimit uint64) int64 {
	if explicit != "" {
		return 0
	}
	return cgroupGoMemoryTarget(cgroupLimit)
}

func cgroupGoMemoryTarget(limit uint64) int64 {
	if limit < goMemoryGiB || limit > 2*goMemoryGiB {
		return 0
	}
	target := limit * 3 / 4
	if target > maxAutoGoMemory {
		target = maxAutoGoMemory
	}
	return int64(target)
}

func readCgroupMemoryLimit(root string) uint64 {
	return readCgroupMemoryHierarchy(root, readCgroupProc("/proc/self/cgroup"), readCgroupProc("/proc/self/mountinfo"))
}

// These kernel metadata files can grow with the host's mount/controller count.
// An incomplete table cannot safely identify a process's mounted hierarchy.
func readCgroupProc(name string) string {
	f, err := os.Open(name)
	if err != nil {
		return ""
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return ""
	}
	return string(data)
}

func readCgroupMemoryHierarchy(root, memberships, mounts string) uint64 {
	root = filepath.Clean(root)
	if !filepath.IsAbs(root) {
		return 0
	}
	var limit uint64
	readLimit := func(name string) {
		data, err := os.ReadFile(name)
		if err != nil {
			return
		}
		value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		if err != nil || value == 0 {
			return
		}
		if limit == 0 || value < limit {
			limit = value
		}
	}
	// Preserve mount-root detection for private cgroup namespaces and for hosts
	// whose process metadata is unavailable. A systemd service normally needs
	// its own directory and every visible ancestor as well.
	readLimit(filepath.Join(root, "memory.max"))
	readLimit(filepath.Join(root, "memory", "memory.limit_in_bytes"))
	for _, line := range strings.Split(memberships, "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) != 3 || !validCgroupPath(parts[2]) {
			continue
		}
		base, filename, filesystem := root, "memory.max", "cgroup2"
		if parts[0] != "0" || parts[1] != "" {
			memory := false
			for _, controller := range strings.Split(parts[1], ",") {
				memory = memory || controller == "memory"
			}
			if !memory {
				continue
			}
			base, filename, filesystem = filepath.Join(root, "memory"), "memory.limit_in_bytes", "cgroup"
		}
		mountRoot := cgroupMountRoot(base, filesystem, mounts)
		path := parts[2]
		if mountRoot != "/" {
			if path == mountRoot {
				path = "/"
			} else if strings.HasPrefix(path, mountRoot+"/") {
				path = strings.TrimPrefix(path, mountRoot)
			} else {
				continue
			}
		}
		dir := filepath.Join(base, strings.TrimPrefix(path, "/"))
		for {
			readLimit(filepath.Join(dir, filename))
			if dir == base {
				break
			}
			dir = filepath.Dir(dir)
		}
	}
	return limit
}

func validCgroupPath(path string) bool {
	return len(path) <= 4096 && filepath.IsAbs(path) && filepath.Clean(path) == path && !strings.ContainsAny(path, "\x00\r\n")
}

func cgroupMountRoot(base, filesystem, mounts string) string {
	for _, line := range strings.Split(mounts, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		separator := -1
		for i := 6; i < len(fields); i++ {
			if fields[i] == "-" {
				separator = i
				break
			}
		}
		if separator < 0 || separator+3 >= len(fields) || fields[separator+1] != filesystem {
			continue
		}
		point, ok := decodeCgroupMountPath(fields[4])
		if !ok || point != base {
			continue
		}
		if root, ok := decodeCgroupMountPath(fields[3]); ok && validCgroupPath(root) {
			return root
		}
	}
	return "/"
}

func decodeCgroupMountPath(encoded string) (string, bool) {
	var out strings.Builder
	for i := 0; i < len(encoded); i++ {
		if encoded[i] != '\\' {
			out.WriteByte(encoded[i])
			continue
		}
		if i+3 >= len(encoded) {
			return "", false
		}
		escape := encoded[i+1 : i+4]
		switch escape {
		case "040":
			out.WriteByte(' ')
		case "011":
			out.WriteByte('\t')
		case "012":
			out.WriteByte('\n')
		case "134":
			out.WriteByte('\\')
		default:
			return "", false
		}
		i += 3
	}
	return out.String(), true
}
