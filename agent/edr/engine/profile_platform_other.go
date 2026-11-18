//go:build !darwin && !linux

package engine

import (
	"runtime"

	sysinfo "github.com/elastic/go-sysinfo"
)

// The native engine plans from RAM and cores. Available disk does not bound
// its in-memory analysis; generation/reload memory is a separate resource.
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
	return mem.Total, cores, nil
}
