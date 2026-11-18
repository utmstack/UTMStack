package engine

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	sysinfo "github.com/elastic/go-sysinfo"
)

// The cgroup limit is the usable RAM for an agent in a container. The host
// total alone can select scan budgets that are too large for the process.
// Database compilation and overlapping generations still require separate
// memory qualification; this function does not set a Go heap limit.
func profileHost() (ramBytes uint64, cores int, err error) {
	cores = runtime.NumCPU()
	host, err := sysinfo.Host()
	if err != nil {
		return 0, cores, err
	}
	mem, err := host.Memory()
	if err != nil {
		return 0, cores, err
	}
	return effectiveLinuxRAM(mem.Total, "/sys/fs/cgroup"), cores, nil
}

func effectiveLinuxRAM(hostTotal uint64, cgroupRoot string) uint64 {
	for _, name := range []string{"memory.max", filepath.Join("memory", "memory.limit_in_bytes")} {
		data, err := os.ReadFile(filepath.Join(cgroupRoot, name))
		if err != nil {
			continue
		}
		limit, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
		if err == nil && limit > 0 && limit < hostTotal {
			hostTotal = limit
		}
	}
	return hostTotal
}
