package utils

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

var AssignedMemory float32

func TotalMemory() uint64 {
	v, _ := mem.VirtualMemory()
	return v.Total / 1024 / 1024
}

func FreeMemory() uint64 {
	v, _ := mem.VirtualMemory()
	return v.Free / 1024 / 1024
}

func UsedByEngineMemory() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024 / 1024
}

func Status() {
	for {
		usedByEngine := UsedByEngineMemory()
		h.Info("Memory used by engine: %v MB", usedByEngine)
		h.Info("Free memory: %v MB", FreeMemory())
		h.Info("Physical memory: %v MB", TotalMemory())
		AssignedMemory = float32(usedByEngine) / float32(TotalMemory()/3) * 100
		h.Info("Assigned memory used: %v %%", AssignedMemory)
		time.Sleep(60 * time.Second)
	}
}
