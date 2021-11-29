package utils

import (
	"fmt"

	sigar "github.com/cloudfoundry/gosigar"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

func CheckMem(size uint64) error {
	m := sigar.Mem{}
	m.Get()
	total := m.Total / 1024 / 1024 / 1024
	if total < size {
		return fmt.Errorf("your system does not have the minimal memory (%v GB) required: %v GB", total, size+1)
	}
	return nil
}

func CheckCPU(cores int) error {
	c, _ := cpu.Counts(true)
	if c < cores {
		return fmt.Errorf("your system does not have the minimal CPU (%v cores) required: %v cores", c, cores)
	}
	return nil
}

func CheckDistro(distro string) error {
	info, _ := host.Info()
	if info.Platform != distro {
		return fmt.Errorf("your Linux distribution (%s) is not %s", info.Platform, distro)
	}
	return nil
}
