package utils

import (
	"github.com/attreios/holmes"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/cloudfoundry/gosigar"
	"github.com/shirou/gopsutil/v3/host"
)

func CheckMem(size uint64) {
	h := holmes.New("debug", "UTMStack")
	m := sigar.Mem{}
	m.Get()
	total := m.Total/1024/1024
	if total < size {
		h.FatalError("Your system does not have the minimal memory (%v MB) required: %v MB", total, size)
	}
}

func CheckCPU(cores int) {
	h := holmes.New("debug", "UTMStack")
	c, _ := cpu.Counts(true)
	if c < cores {
		h.FatalError("Your system does not have the minimal CPU (%v cores) required: %v cores", c, cores)
	}
}

func CheckDistro(distro string){
	h := holmes.New("debug", "UTMStack")
	info, _ := host.Info()
	if info.Platform != distro {
		h.FatalError("Your Linux distribution (%s) is not %s", info.Platform, distro)
	}
}