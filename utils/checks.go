package utils

import (
	"github.com/attreios/holmes"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/host"
)

func CheckMem(size uint64) {
	h := holmes.New("debug", "UTMStack")
	v, _ := mem.VirtualMemory()
	if v.Total < size {
		h.FatalError("Your system does not have the minimal memory required: %v MB", v.Total)
	}
}

func CheckCPU(cores int) {
	h := holmes.New("debug", "UTMStack")
	c, _ := cpu.Counts(true)
	if c < cores {
		h.FatalError("Your system does not have the minimal CPU required: %v Cores", c)
	}
}

func CheckDistro(){
	h := holmes.New("debug", "UTMStack")
	info, _ := host.Info()
	if info.Platform != "ubuntu" {
		h.FatalError("Your Linux distribution (%v) is not Ubuntu", info.Platform)
	}
}